package handler

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

type SettingsHandler struct {
	settingsService *service.SettingsService
}

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: svc}
}

// SettingsPage renders the full settings page.
// GET /app/settings
func (h *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Settings(csrf)).ServeHTTP(w, r)
}

// GetProfile returns the current user's profile.
// GET /api/v1/settings/profile
func (h *SettingsHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	user, err := h.settingsService.GetProfile(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"currency":   user.Currency,
		"timezone":   user.Timezone,
		"theme":      user.Theme,
	})
}

// UpdateProfile updates the current user's profile.
// PUT /api/v1/settings/profile
func (h *SettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Currency  string `json:"currency"`
		Timezone  string `json:"timezone"`
		Theme     string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	user, err := h.settingsService.UpdateProfile(r.Context(), userID, service.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Currency:  req.Currency,
		Timezone:  req.Timezone,
		Theme:     req.Theme,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"currency":   user.Currency,
		"timezone":   user.Timezone,
		"theme":      user.Theme,
	})
}

// ChangePassword updates the current user's password.
// PUT /api/v1/settings/password
func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, apperr.NewValidationError("invalid request body"))
		return
	}

	err := h.settingsService.ChangePassword(r.Context(), userID, service.ChangePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}
