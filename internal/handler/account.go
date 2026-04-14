package handler

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

// AccountHandler handles bank account HTTP requests.
type AccountHandler struct {
	accountService *service.AccountService
}

// NewAccountHandler creates a new account handler.
func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: svc}
}

// AccountsPage renders the full accounts page.
// GET /app/accounts
func (h *AccountHandler) AccountsPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	csrf := middleware.GetCSRFToken(r)

	accounts, err := h.accountService.ListAccounts(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list accounts for page")
		accounts = nil // show empty state on error
	}

	templ.Handler(pages.Accounts(csrf, accounts)).ServeHTTP(w, r)
}

// CreateLinkToken returns a Plaid Link token as JSON.
// POST /api/v1/accounts/link-token
func (h *AccountHandler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("Not authenticated"))
		return
	}

	token, err := h.accountService.CreateLinkToken(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"link_token": token})
}

// ExchangeToken handles the Plaid Link callback — exchanges public_token and creates accounts.
// POST /api/v1/accounts/exchange-token
func (h *AccountHandler) ExchangeToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("Not authenticated"))
		return
	}

	var req struct {
		PublicToken     string `json:"public_token"`
		InstitutionID   string `json:"institution_id"`
		InstitutionName string `json:"institution_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("Invalid request body"))
		return
	}

	if req.PublicToken == "" {
		respondError(w, apperr.NewValidationError("public_token is required"))
		return
	}
	if req.InstitutionID == "" || req.InstitutionName == "" {
		respondError(w, apperr.NewValidationError("institution_id and institution_name are required"))
		return
	}

	accounts, err := h.accountService.LinkAccount(r.Context(), userID, req.PublicToken, service.LinkMetadata{
		InstitutionID:   req.InstitutionID,
		InstitutionName: req.InstitutionName,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"accounts": accounts,
		"count":    len(accounts),
	})
}

// ListAccounts returns all active accounts for the user.
// GET /api/v1/accounts
func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("Not authenticated"))
		return
	}

	accounts, err := h.accountService.ListAccounts(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"accounts": accounts})
}

// SyncAccount triggers a manual transaction sync for an account.
// POST /api/v1/accounts/{id}/sync
func (h *AccountHandler) SyncAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("Not authenticated"))
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("Invalid account ID"))
		return
	}

	if err := h.accountService.SyncAccount(r.Context(), userID, accountID); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "synced"})
}

// UnlinkAccount removes a linked bank account.
// DELETE /api/v1/accounts/{id}
func (h *AccountHandler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("Not authenticated"))
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("Invalid account ID"))
		return
	}

	if err := h.accountService.UnlinkAccount(r.Context(), userID, accountID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON response")
	}
}

// respondError writes an error response as JSON.
func respondError(w http.ResponseWriter, err error) {
	if domainErr, ok := apperr.IsDomainError(err); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(domainErr.HTTPStatus())
		if encErr := json.NewEncoder(w).Encode(map[string]any{
			"error": domainErr,
		}); encErr != nil {
			log.Error().Err(encErr).Msg("failed to encode error response")
		}
		return
	}

	log.Error().Err(err).Msg("unhandled error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if encErr := json.NewEncoder(w).Encode(map[string]any{
		"error": apperr.DomainError{Code: apperr.CodeInternal, Message: "An unexpected error occurred"},
	}); encErr != nil {
		log.Error().Err(encErr).Msg("failed to encode error response")
	}
}
