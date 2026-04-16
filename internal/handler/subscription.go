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

type SubscriptionHandler struct {
	subService *service.SubscriptionService
}

func NewSubscriptionHandler(subService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

// SubscriptionsPage renders the full subscriptions page.
func (h *SubscriptionHandler) SubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Subscriptions(csrf)).ServeHTTP(w, r)
}

// ListSubscriptions returns subscription cards as HTML partial.
// GET /api/v1/subscriptions
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	subs, err := h.subService.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.SubscriptionList(subs, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"subscriptions": subs})
}

// CreateSubscription handles subscription creation from form.
// POST /api/v1/subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
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

	input := service.CreateSubscriptionInput{
		Name:         r.FormValue("name"),
		Amount:       amount,
		CurrencyCode: "USD",
		Frequency:    r.FormValue("frequency"),
		CategoryID:   categoryID,
		Status:       "active",
	}

	sub, err := h.subService.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "subscriptionCreated")
		templ.Handler(components.SubscriptionCard(sub)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusCreated, sub)
}

// CancelSubscription marks a subscription as cancelled.
// POST /api/v1/subscriptions/{id}/cancel
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid subscription ID"))
		return
	}

	if err := h.subService.Cancel(r.Context(), userID, subID); err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "subscriptionUpdated")
		w.WriteHeader(http.StatusOK)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// DeleteSubscription removes a subscription.
// DELETE /api/v1/subscriptions/{id}
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid subscription ID"))
		return
	}

	if err := h.subService.Delete(r.Context(), userID, subID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
