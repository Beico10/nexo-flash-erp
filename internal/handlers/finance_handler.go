// Package handlers — endpoints de DRE e Fluxo de Caixa.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/nexoone/nexo-one/internal/finance"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type FinanceHandler struct {
	svc *finance.Service
}

func NewFinanceHandler(svc *finance.Service) *FinanceHandler {
	return &FinanceHandler{svc: svc}
}

func (h *FinanceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/finance/dre", h.GetDRE)
	mux.HandleFunc("GET /api/v1/finance/cashflow", h.GetCashFlow)
}

// GetDRE GET /api/v1/finance/dre?year=2026&month=3
func (h *FinanceHandler) GetDRE(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	q := r.URL.Query()

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := q.Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil {
			year = parsed
		}
	}
	if m := q.Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	dre, err := h.svc.GetDRE(r.Context(), tenantID, year, month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dre)
}

// GetCashFlow GET /api/v1/finance/cashflow?days=30
func (h *FinanceHandler) GetCashFlow(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	cf, err := h.svc.GetCashFlow(r.Context(), tenantID, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, cf)
}
