package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/components"
	"github.com/nghiaduong/finai/templates/pages"
)

type BudgetHandler struct {
	budgetService *service.BudgetService
}

func NewBudgetHandler(budgetService *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: budgetService}
}

// BudgetsPage renders the full budgets page.
func (h *BudgetHandler) BudgetsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Budgets(csrf)).ServeHTTP(w, r)
}

// ListBudgets returns budget progress bars as HTML partial.
// GET /api/v1/budgets
func (h *BudgetHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	budgets, err := h.budgetService.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.BudgetList(budgets, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"budgets": budgets})
}

// CreateBudget handles budget creation from form.
// POST /api/v1/budgets
func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	if err := r.ParseForm(); err != nil {
		respondError(w, apperr.NewValidationError("invalid form data"))
		return
	}

	amountLimit, err := decimal.NewFromString(r.FormValue("amount_limit"))
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

	if amountLimit.IsNegative() || amountLimit.IsZero() {
		respondError(w, apperr.NewValidationError("amount limit must be positive"))
		return
	}

	alertThreshold := decimal.NewFromFloat(0.80)
	if v := r.FormValue("alert_threshold"); v != "" {
		parsed, err := decimal.NewFromString(v)
		if err != nil {
			respondError(w, apperr.NewValidationError("invalid alert threshold"))
			return
		}
		if parsed.IsNegative() || parsed.GreaterThan(decimal.NewFromInt(1)) {
			respondError(w, apperr.NewValidationError("alert threshold must be between 0 and 1"))
			return
		}
		alertThreshold = parsed
	}

	input := service.CreateBudgetInput{
		CategoryID:     categoryID,
		Name:           r.FormValue("name"),
		AmountLimit:    amountLimit,
		Period:         r.FormValue("period"),
		AlertThreshold: alertThreshold,
	}

	budget, err := h.budgetService.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "budgetCreated")
		templ.Handler(components.BudgetProgressBar(budget)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusCreated, budget)
}

// DeleteBudget removes a budget.
// DELETE /api/v1/budgets/{id}
func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid budget ID"))
		return
	}

	if err := h.budgetService.Delete(r.Context(), userID, budgetID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
