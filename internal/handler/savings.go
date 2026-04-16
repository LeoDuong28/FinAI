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

type SavingsHandler struct {
	savingsService *service.SavingsService
}

func NewSavingsHandler(savingsService *service.SavingsService) *SavingsHandler {
	return &SavingsHandler{savingsService: savingsService}
}

// SavingsPage renders the full savings page.
func (h *SavingsHandler) SavingsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Savings(csrf)).ServeHTTP(w, r)
}

// ListSavingsGoals returns savings goal cards as HTML partial.
// GET /api/v1/savings
func (h *SavingsHandler) ListSavingsGoals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	goals, err := h.savingsService.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.SavingsList(goals, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"savings_goals": goals})
}

// CreateSavingsGoal handles savings goal creation from form.
// POST /api/v1/savings
func (h *SavingsHandler) CreateSavingsGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	if err := r.ParseForm(); err != nil {
		respondError(w, apperr.NewValidationError("invalid form data"))
		return
	}

	targetAmount, err := decimal.NewFromString(r.FormValue("target_amount"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid target amount"))
		return
	}

	var icon, color *string
	if v := r.FormValue("icon"); v != "" {
		icon = &v
	}
	if v := r.FormValue("color"); v != "" {
		color = &v
	}

	input := service.CreateSavingsInput{
		Name:         r.FormValue("name"),
		TargetAmount: targetAmount,
		Icon:         icon,
		Color:        color,
	}

	goal, err := h.savingsService.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "savingsCreated")
		templ.Handler(components.SavingsGoalCard(goal)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusCreated, goal)
}

// AddFunds adds funds to a savings goal.
// POST /api/v1/savings/{id}/add
func (h *SavingsHandler) AddFunds(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid savings goal ID"))
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

	goal, err := h.savingsService.AddFunds(r.Context(), userID, goalID, amount)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "savingsUpdated")
		templ.Handler(components.SavingsGoalCard(goal)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, goal)
}

// WithdrawFunds removes funds from a savings goal.
// POST /api/v1/savings/{id}/withdraw
func (h *SavingsHandler) WithdrawFunds(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid savings goal ID"))
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

	goal, err := h.savingsService.WithdrawFunds(r.Context(), userID, goalID, amount)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "savingsUpdated")
		templ.Handler(components.SavingsGoalCard(goal)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, goal)
}

// DeleteSavingsGoal removes a savings goal.
// DELETE /api/v1/savings/{id}
func (h *SavingsHandler) DeleteSavingsGoal(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	goalID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid savings goal ID"))
		return
	}

	if err := h.savingsService.Delete(r.Context(), userID, goalID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
