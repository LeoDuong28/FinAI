package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

// ChatService handles AI advisor chat conversations.
type ChatService struct {
	chatRepo  domain.ChatRepository
	aiService *AIService
	insights  *InsightsService
}

// NewChatService creates a new chat service.
func NewChatService(chatRepo domain.ChatRepository, aiService *AIService, insights *InsightsService) *ChatService {
	return &ChatService{
		chatRepo:  chatRepo,
		aiService: aiService,
		insights:  insights,
	}
}

// SendMessage sends a user message, gets an AI response, stores both, and returns the response.
func (s *ChatService) SendMessage(ctx context.Context, userID, sessionID uuid.UUID, message string) (string, error) {
	if message == "" {
		return "", apperr.NewValidationError("message is required")
	}
	if len(message) > 4096 {
		return "", apperr.NewValidationError("message is too long (max 4096 characters)")
	}

	// Store user message
	userMsg := &domain.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "user",
		Content:   message,
	}
	if err := s.chatRepo.Create(ctx, userMsg); err != nil {
		return "", apperr.NewInternalError("failed to save message")
	}

	// Build financial context for the AI
	chatCtx := s.buildContext(ctx, userID)

	// Get chat history for context
	history, err := s.chatRepo.ListBySession(ctx, userID, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to load chat history")
		history = nil
	}

	// Convert to AI service format (exclude the message we just stored — it's passed separately)
	var aiHistory []ChatMessage
	for _, msg := range history {
		if msg.ID == userMsg.ID {
			continue
		}
		aiHistory = append(aiHistory, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Call AI service
	response, err := s.aiService.ChatSync(ctx, message, aiHistory, chatCtx)
	if err != nil {
		log.Error().Err(err).Msg("AI chat failed")
		response = "I'm sorry, I'm having trouble processing your request right now. Please try again in a moment."
	}

	// Store assistant response
	assistantMsg := &domain.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   response,
	}
	if err := s.chatRepo.Create(ctx, assistantMsg); err != nil {
		log.Error().Err(err).Msg("failed to save assistant message")
		return "", apperr.NewInternalError("failed to save response")
	}

	return response, nil
}

// GetHistory returns chat messages for a session.
func (s *ChatService) GetHistory(ctx context.Context, userID, sessionID uuid.UUID) ([]domain.ChatMessage, error) {
	messages, err := s.chatRepo.ListBySession(ctx, userID, sessionID)
	if err != nil {
		return nil, apperr.NewInternalError("failed to load chat history")
	}
	return messages, nil
}

// NewSession generates a new session UUID.
func (s *ChatService) NewSession() uuid.UUID {
	return uuid.New()
}

func (s *ChatService) buildContext(ctx context.Context, userID uuid.UUID) *ChatContext {
	chatCtx := &ChatContext{}

	if spending, err := s.insights.MonthlySpending(ctx, userID); err == nil {
		f := DecimalToFloat(spending)
		chatCtx.MonthlySpending = &f
	}
	if income, err := s.insights.MonthlyIncome(ctx, userID); err == nil {
		f := DecimalToFloat(income)
		chatCtx.MonthlyIncome = &f
	}

	from, to := CurrentMonthBounds()
	if categories, err := s.insights.SpendingByCategory(ctx, userID, from, to); err == nil && len(categories) > 0 {
		topCats := make(map[string]float64)
		for i, cat := range categories {
			if i >= 5 {
				break
			}
			topCats[cat.CategoryName] = DecimalToFloat(cat.Total)
		}
		chatCtx.TopCategories = topCats
	}

	return chatCtx
}
