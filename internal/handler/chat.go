package handler

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: svc}
}

// ChatPage renders the full chat page.
// GET /app/chat
func (h *ChatHandler) ChatPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Chat(csrf)).ServeHTTP(w, r)
}

// SendMessage handles sending a message and getting an AI response.
// POST /api/v1/chat
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	if req.Message == "" {
		respondError(w, apperr.NewValidationError("message is required"))
		return
	}

	if req.SessionID == "" {
		respondError(w, apperr.NewValidationError("session_id is required"))
		return
	}
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid session_id format"))
		return
	}

	response, err := h.chatService.SendMessage(r.Context(), userID, sessionID, req.Message)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"response": response})
}

// GetHistory returns chat history for a session.
// GET /api/v1/chat/history/{sessionId}
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid session ID"))
		return
	}

	messages, err := h.chatService.GetHistory(r.Context(), userID, sessionID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

// NewSession creates a new chat session.
// POST /api/v1/chat/session
func (h *ChatHandler) NewSession(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	sessionID := h.chatService.NewSession()
	respondJSON(w, http.StatusCreated, map[string]string{"session_id": sessionID.String()})
}
