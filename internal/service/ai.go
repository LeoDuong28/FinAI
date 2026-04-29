package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/config"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

// AIService is the client for the Python AI microservice.
// It is safe for concurrent use.
type AIService struct {
	baseURL    string
	apiKey     string
	hmacSecret string
	httpClient *http.Client
	cb         *CircuitBreaker
}

// NewAIService creates a new AI service client.
func NewAIService(cfg *config.AIConfig) *AIService {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &AIService{
		baseURL:    cfg.ServiceURL,
		apiKey:     cfg.ServiceKey,
		hmacSecret: cfg.HMACSecret,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		cb: NewCircuitBreaker(5, 30*time.Second, 60*time.Second),
	}
}

// --- Request/Response types ---

// CategorizeTransactionInput is a single transaction to categorize.
type CategorizeTransactionInput struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MerchantName *string `json:"merchant_name,omitempty"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
}

// CategoryResult is the categorization result for a single transaction.
type CategoryResult struct {
	TransactionID string  `json:"transaction_id"`
	Category      string  `json:"category"`
	Confidence    float64 `json:"confidence"`
}

type categorizeRequest struct {
	Transactions []CategorizeTransactionInput `json:"transactions"`
}

type categorizeResponse struct {
	Categories []CategoryResult `json:"categories"`
}

// DetectedSubscription represents a detected recurring subscription.
type DetectedSubscription struct {
	MerchantName     string  `json:"merchant_name"`
	Amount           float64 `json:"amount"`
	Frequency        string  `json:"frequency"`
	Confidence       float64 `json:"confidence"`
	NextBilling      *string `json:"next_billing,omitempty"`
	LastCharged      *string `json:"last_charged,omitempty"`
	TransactionCount int     `json:"transaction_count"`
}

// TransactionHistoryInput is a transaction for subscription detection.
type TransactionHistoryInput struct {
	Name         string  `json:"name"`
	MerchantName *string `json:"merchant_name,omitempty"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Type         string  `json:"type"`
}

type detectSubsRequest struct {
	Transactions []TransactionHistoryInput `json:"transactions"`
}

type detectSubsResponse struct {
	Subscriptions []DetectedSubscription `json:"subscriptions"`
}

// AnomalyInput is a transaction for anomaly detection.
type AnomalyInput struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MerchantName *string `json:"merchant_name,omitempty"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Category     *string `json:"category,omitempty"`
}

// AnomalyResult is the result of anomaly detection.
type AnomalyResult struct {
	TransactionID string   `json:"transaction_id"`
	IsAnomaly     bool     `json:"is_anomaly"`
	Reason        string   `json:"reason"`
	ZScore        *float64 `json:"z_score,omitempty"`
}

type anomalyRequest struct {
	Transactions []AnomalyInput `json:"transactions"`
	History      []AnomalyInput `json:"history"`
}

type anomalyResponse struct {
	Anomalies []AnomalyResult `json:"anomalies"`
}

// DailySpendingInput is a daily spending entry for forecasting.
type DailySpendingInput struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category *string `json:"category,omitempty"`
}

// ForecastResult is the spending forecast.
type ForecastResult struct {
	PredictedTotal    float64              `json:"predicted_total"`
	DailyPredictions  []DailySpendingInput `json:"daily_predictions"`
	CategoryBreakdown map[string]float64   `json:"category_breakdown,omitempty"`
}

type forecastRequest struct {
	History []DailySpendingInput `json:"history"`
	Days    int                  `json:"days"`
}

type forecastResponse struct {
	Forecast ForecastResult `json:"forecast"`
}

// NegotiationTip is a bill negotiation suggestion.
type NegotiationTip struct {
	Tip              string  `json:"tip"`
	TypicalRange     *string `json:"typical_range,omitempty"`
	PotentialSavings *string `json:"potential_savings,omitempty"`
}

// BillInput is a bill for negotiation tips.
type BillInput struct {
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	Category  *string `json:"category,omitempty"`
	Frequency *string `json:"frequency,omitempty"`
}

type negotiateRequest struct {
	Bills []BillInput `json:"bills"`
}

type negotiateResponse struct {
	Tips []NegotiationTip `json:"tips"`
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// BudgetStatusEntry represents budget status for AI context.
type BudgetStatusEntry struct {
	Spent float64 `json:"spent"`
	Limit float64 `json:"limit"`
}

// ChatContext is the user's financial context for AI chat.
type ChatContext struct {
	MonthlyIncome   *float64                     `json:"monthly_income,omitempty"`
	MonthlySpending *float64                     `json:"monthly_spending,omitempty"`
	TopCategories   map[string]float64            `json:"top_categories,omitempty"`
	BudgetStatus    map[string]BudgetStatusEntry  `json:"budget_status,omitempty"`
	RecentAlerts    []string                      `json:"recent_alerts,omitempty"`
}

type chatRequest struct {
	Message string        `json:"message"`
	History []ChatMessage `json:"history"`
	Context *ChatContext  `json:"context,omitempty"`
}

type chatSyncResponse struct {
	Content string `json:"content"`
}

// --- Public methods ---

// CategorizeTransactions sends transactions to the AI service for categorization.
func (s *AIService) CategorizeTransactions(ctx context.Context, txns []CategorizeTransactionInput) ([]CategoryResult, error) {
	var resp categorizeResponse
	err := s.doRequest(ctx, "POST", "/api/categorize", categorizeRequest{Transactions: txns}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Categories, nil
}

// DetectSubscriptions sends transaction history for subscription detection.
func (s *AIService) DetectSubscriptions(ctx context.Context, txns []TransactionHistoryInput) ([]DetectedSubscription, error) {
	var resp detectSubsResponse
	err := s.doRequest(ctx, "POST", "/api/detect-subscriptions", detectSubsRequest{Transactions: txns}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Subscriptions, nil
}

// DetectAnomalies detects anomalous transactions.
func (s *AIService) DetectAnomalies(ctx context.Context, txns, history []AnomalyInput) ([]AnomalyResult, error) {
	var resp anomalyResponse
	err := s.doRequest(ctx, "POST", "/api/anomaly-detect", anomalyRequest{Transactions: txns, History: history}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Anomalies, nil
}

// ForecastSpending predicts future spending.
func (s *AIService) ForecastSpending(ctx context.Context, history []DailySpendingInput, days int) (*ForecastResult, error) {
	var resp forecastResponse
	err := s.doRequest(ctx, "POST", "/api/forecast", forecastRequest{History: history, Days: days}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Forecast, nil
}

// GetNegotiationTips generates bill negotiation suggestions.
func (s *AIService) GetNegotiationTips(ctx context.Context, bills []BillInput) ([]NegotiationTip, error) {
	var resp negotiateResponse
	err := s.doRequest(ctx, "POST", "/api/negotiate-tips", negotiateRequest{Bills: bills}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Tips, nil
}

// ChatSync sends a chat message and returns the full response (non-streaming).
func (s *AIService) ChatSync(ctx context.Context, message string, history []ChatMessage, chatCtx *ChatContext) (string, error) {
	var resp chatSyncResponse
	err := s.doRequest(ctx, "POST", "/api/chat/sync", chatRequest{
		Message: message,
		History: history,
		Context: chatCtx,
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// IsAvailable reports whether the circuit breaker currently allows requests.
// This is a hint only — the circuit may change state between this call and actual use.
func (s *AIService) IsAvailable() bool {
	return s.cb.State() != CircuitOpen
}

// DecimalToFloat converts decimal.Decimal to float64 for AI service requests.
// Note: float64 is acceptable for AI categorization/forecasting where exact precision is not required.
func DecimalToFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

// --- Internal ---

func (s *AIService) doRequest(ctx context.Context, method, path string, body, result any) error {
	if !s.cb.Allow() {
		return &apperr.DomainError{Code: apperr.CodeAIUnavailable, Message: "AI service temporarily unavailable"}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return apperr.NewInternalError("failed to process request")
	}

	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return apperr.NewInternalError("failed to process request")
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("X-Service-Key", s.apiKey)
	}
	if s.hmacSecret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		// Sign timestamp + body to prevent replay attacks
		mac := hmac.New(sha256.New, []byte(s.hmacSecret))
		mac.Write([]byte(timestamp))
		mac.Write([]byte("."))
		mac.Write(jsonBody)
		req.Header.Set("X-Signature", hex.EncodeToString(mac.Sum(nil)))
		req.Header.Set("X-Timestamp", timestamp)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.cb.RecordFailure()
		log.Error().Err(err).Str("path", path).Msg("AI service request failed")
		return &apperr.DomainError{Code: apperr.CodeAIUnavailable, Message: "AI service unavailable"}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		s.cb.RecordFailure()
		return apperr.NewInternalError("failed to read AI response")
	}

	if resp.StatusCode >= 500 {
		s.cb.RecordFailure()
		log.Error().Int("status", resp.StatusCode).Str("path", path).Msg("AI service error")
		return &apperr.DomainError{Code: apperr.CodeAIUnavailable, Message: "AI service error"}
	}

	if resp.StatusCode >= 400 {
		log.Warn().Int("status", resp.StatusCode).Str("path", path).Msg("AI service rejected request")
		return apperr.NewInternalError("AI request failed")
	}

	s.cb.RecordSuccess()

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			log.Error().Err(err).Str("path", path).Msg("failed to decode AI response")
			return apperr.NewInternalError("failed to process AI response")
		}
	}

	return nil
}

