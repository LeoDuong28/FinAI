package handler

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/templates/pages"
)

type NetWorthHandler struct {
	netWorthService *service.NetWorthService
}

func NewNetWorthHandler(svc *service.NetWorthService) *NetWorthHandler {
	return &NetWorthHandler{netWorthService: svc}
}

// NetWorthPage renders the full net worth page.
// GET /app/networth
func (h *NetWorthHandler) NetWorthPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.NetWorth(csrf)).ServeHTTP(w, r)
}

// GetLatest returns the latest net worth snapshot.
// GET /api/v1/networth
func (h *NetWorthHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	snapshot, err := h.netWorthService.GetLatest(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// CalculateSnapshot calculates and stores a new net worth snapshot.
// POST /api/v1/networth/snapshot
func (h *NetWorthHandler) CalculateSnapshot(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	snapshot, err := h.netWorthService.CalculateSnapshot(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, snapshot)
}

// ListHistory returns net worth history for charting.
// GET /api/v1/networth/history?limit=30
func (h *NetWorthHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	limit := 30
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 365 {
		limit = 365
	}

	snapshots, err := h.netWorthService.ListHistory(r.Context(), userID, limit)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"snapshots": snapshots})
}
