package handler

import (
	"net/http"
	"time"

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

type BillHandler struct {
	billService *service.BillService
}

func NewBillHandler(billService *service.BillService) *BillHandler {
	return &BillHandler{billService: billService}
}

// BillsPage renders the full bills page.
func (h *BillHandler) BillsPage(w http.ResponseWriter, r *http.Request) {
	csrf := middleware.GetCSRFToken(r)
	templ.Handler(pages.Bills(csrf)).ServeHTTP(w, r)
}

// ListBills returns bill cards as HTML partial.
// GET /api/v1/bills
func (h *BillHandler) ListBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	bills, err := h.billService.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		csrf := middleware.GetCSRFToken(r)
		templ.Handler(components.BillList(bills, csrf)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"bills": bills})
}

// CreateBill handles bill creation from form.
// POST /api/v1/bills
func (h *BillHandler) CreateBill(w http.ResponseWriter, r *http.Request) {
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

	dueDate, err := time.Parse("2006-01-02", r.FormValue("due_date"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid due date"))
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

	input := service.CreateBillInput{
		Name:         r.FormValue("name"),
		Amount:       amount,
		DueDate:      dueDate,
		Frequency:    r.FormValue("frequency"),
		CategoryID:   categoryID,
		IsAutopay:    r.FormValue("is_autopay") == "on",
		ReminderDays: 3,
	}

	bill, err := h.billService.Create(r.Context(), userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "billCreated")
		templ.Handler(components.BillCard(bill)).ServeHTTP(w, r)
		return
	}
	respondJSON(w, http.StatusCreated, bill)
}

// MarkBillPaid marks a bill as paid.
// POST /api/v1/bills/{id}/pay
func (h *BillHandler) MarkBillPaid(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	billID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid bill ID"))
		return
	}

	if err := h.billService.MarkPaid(r.Context(), userID, billID); err != nil {
		respondError(w, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "billUpdated")
		w.WriteHeader(http.StatusOK)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "paid"})
}

// DeleteBill removes a bill.
// DELETE /api/v1/bills/{id}
func (h *BillHandler) DeleteBill(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, apperr.NewUnauthorizedError("not authenticated"))
		return
	}

	billID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, apperr.NewValidationError("invalid bill ID"))
		return
	}

	if err := h.billService.Delete(r.Context(), userID, billID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
