package handler

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/components"
	"github.com/nghiaduong/finai/templates/pages"
)

type InsightsHandler struct {
	insightsService *service.InsightsService
}

func NewInsightsHandler(svc *service.InsightsService) *InsightsHandler {
	return &InsightsHandler{insightsService: svc}
}

// InsightsPage renders the full insights page.
// GET /app/insights
func (h *InsightsHandler) InsightsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Insights(csrf)).ServeHTTP(w, r)
}

// SpendingCard returns the monthly spending stat card partial.
// GET /api/v1/insights/spending
func (h *InsightsHandler) SpendingCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	total, err := h.insightsService.MonthlySpending(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(components.StatCard("Monthly Spending", "$"+total.StringFixed(2), "expense")).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"spending": total.StringFixed(2)})
}

// IncomeCard returns the monthly income stat card partial.
// GET /api/v1/insights/income
func (h *InsightsHandler) IncomeCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	total, err := h.insightsService.MonthlyIncome(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(components.StatCard("Monthly Income", "$"+total.StringFixed(2), "income")).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"income": total.StringFixed(2)})
}

// SubscriptionsTotalCard returns the subscriptions total stat card partial.
// GET /api/v1/insights/subscriptions-total
func (h *InsightsHandler) SubscriptionsTotalCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	total, err := h.insightsService.SubscriptionsTotal(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(components.StatCard("Subscriptions", "$"+total.StringFixed(2)+"/mo", "subscription")).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"subscriptions_monthly": total.StringFixed(2)})
}

// SavingsProgressCard returns the savings progress stat card partial.
// GET /api/v1/insights/savings-progress
func (h *InsightsHandler) SavingsProgressCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	progress, err := h.insightsService.SavingsProgress(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	value := "$" + progress.TotalSaved.StringFixed(2)
	if !progress.TotalTarget.IsZero() {
		value += " / $" + progress.TotalTarget.StringFixed(2)
	}

	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(components.StatCard("Savings", value, "savings")).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"saved":  progress.TotalSaved.StringFixed(2),
		"target": progress.TotalTarget.StringFixed(2),
	})
}

// CategoriesBreakdown returns spending by category.
// GET /api/v1/insights/categories
func (h *InsightsHandler) CategoriesBreakdown(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	from, to := service.CurrentMonthBounds()
	categories, err := h.insightsService.SpendingByCategory(r.Context(), userID, from, to)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(components.CategoryChart(categories)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"categories": categories})
}

// Forecast returns spending forecast data.
// GET /api/v1/insights/forecast?days=30
func (h *InsightsHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 90 {
			days = n
		}
	}

	forecast, err := h.insightsService.SpendingForecast(r.Context(), userID, days)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"forecast": forecast})
}
