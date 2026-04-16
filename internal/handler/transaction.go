package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/components"
	"github.com/nghiaduong/finai/templates/pages"
)

type TransactionHandler struct {
	txnService *service.TransactionService
}

func NewTransactionHandler(txnService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txnService: txnService}
}

// TransactionsPage renders the full transactions page.
// GET /app/transactions
func (h *TransactionHandler) TransactionsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Transactions(csrf)).ServeHTTP(w, r)
}

// ListTransactions returns transaction rows as HTML partial.
// GET /api/v1/transactions
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	filter := parseTransactionFilter(r)

	txns, cursor, err := h.txnService.List(r.Context(), userID, filter)
	if err != nil {
		respondError(w, err)
		return
	}

	// If HTMX request, return HTML partial.
	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.TransactionList(txns, cursor, csrf)).ServeHTTP(w, r)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"transactions": txns, "cursor": cursor})
}

// GetTransaction returns a single transaction.
// GET /api/v1/transactions/{id}
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid transaction ID"))
		return
	}

	txn, err := h.txnService.GetByID(r.Context(), userID, txnID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.TransactionDetail(txn, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, txn)
}

// CreateTransaction handles manual transaction entry.
// POST /api/v1/transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	if err := r.ParseForm(); err != nil {
		respondError(w, apperr.NewValidationError("invalid form data"))
		return
	}

	amount, err := decimal.NewFromString(r.FormValue("amount"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid amount"))
		return
	}

	var categoryID *uuid.UUID
	if catStr := r.FormValue("category_id"); catStr != "" {
		id, err := uuid.Parse(catStr)
		if err != nil {
			respondError(w, apperr.NewValidationError("invalid category ID"))
			return
		}
		categoryID = &id
	}

	date := time.Now().UTC()
	if dateStr := r.FormValue("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = parsed
		}
	}

	var notes *string
	if n := r.FormValue("notes"); n != "" {
		notes = &n
	}

	input := service.CreateTransactionInput{
		CategoryID: categoryID,
		Amount:     amount,
		Date:       date,
		Name:       r.FormValue("name"),
		Type:       r.FormValue("type"),
		Notes:      notes,
	}

	txn, err := h.txnService.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "transactionCreated")
		templ.Handler(components.TransactionRow(txn)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusCreated, txn)
}

// UpdateTransactionCategory updates a transaction's category.
// PATCH /api/v1/transactions/{id}/category
func (h *TransactionHandler) UpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid transaction ID"))
		return
	}

	var req struct {
		CategoryID *uuid.UUID `json:"category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	if err := h.txnService.UpdateCategory(r.Context(), userID, txnID, req.CategoryID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateTransactionNotes updates a transaction's notes.
// PATCH /api/v1/transactions/{id}/notes
func (h *TransactionHandler) UpdateTransactionNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid transaction ID"))
		return
	}

	var req struct {
		Notes *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	if err := h.txnService.UpdateNotes(r.Context(), userID, txnID, req.Notes); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTransaction removes a transaction.
// DELETE /api/v1/transactions/{id}
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	txnID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid transaction ID"))
		return
	}

	if err := h.txnService.Delete(r.Context(), userID, txnID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchTransactions performs a fuzzy search.
// GET /api/v1/transactions/search?q=...
func (h *TransactionHandler) SearchTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	query := r.URL.Query().Get("q")
	txns, err := h.txnService.Search(r.Context(), userID, query)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.TransactionList(txns, nil, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"transactions": txns})
}

func parseTransactionFilter(r *http.Request) domain.TransactionFilter {
	q := r.URL.Query()
	filter := domain.TransactionFilter{}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if v := q.Get("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CategoryID = &id
		}
	}
	if v := q.Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AccountID = &id
		}
	}
	if v := q.Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := q.Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateTo = &t
		}
	}
	if v := q.Get("type"); v != "" {
		filter.Type = &v
	}

	// Cursor
	if cursorDate := q.Get("cursor_date"); cursorDate != "" {
		if cursorID := q.Get("cursor_id"); cursorID != "" {
			d, dErr := time.Parse("2006-01-02", cursorDate)
			id, idErr := uuid.Parse(cursorID)
			if dErr == nil && idErr == nil {
				filter.Cursor = &domain.TransactionCursor{Date: d, ID: id}
			}
		}
	}

	if v := q.Get("search"); v != "" {
		filter.Search = &v
		log.Debug().Str("search", v).Msg("transaction search filter")
	}

	return filter
}
