// Package handlers — endpoints de Contas a Pagar.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nexoone/nexo-one/internal/payables"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

// PayablesHandler gerencia endpoints de contas a pagar.
type PayablesHandler struct {
	svc *payables.Service
}

func NewPayablesHandler(svc *payables.Service) *PayablesHandler {
	return &PayablesHandler{svc: svc}
}

func (h *PayablesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/payables", h.Create)
	mux.HandleFunc("GET /api/v1/payables", h.List)
	mux.HandleFunc("GET /api/v1/payables/summary", h.Summary)
	mux.HandleFunc("GET /api/v1/payables/cashflow", h.CashFlow)
	mux.HandleFunc("GET /api/v1/payables/{id}", h.GetByID)
	mux.HandleFunc("POST /api/v1/payables/{id}/pay", h.Pay)
	mux.HandleFunc("DELETE /api/v1/payables/{id}", h.Cancel)
}

// Create POST /api/v1/payables
// Cria conta a pagar. Suporta parcelamento automático.
func (h *PayablesHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		Description       string  `json:"description"`
		SupplierName      string  `json:"supplier_name"`
		SupplierCNPJ      string  `json:"supplier_cnpj"`
		Category          string  `json:"category"`
		Amount            float64 `json:"amount"`
		DueDate           string  `json:"due_date"` // YYYY-MM-DD
		Installments      int     `json:"installments"`
		Recurrence        string  `json:"recurrence"`
		Notes             string  `json:"notes"`
		CostCenter        string  `json:"cost_center"`
		NFEKey            string  `json:"nfe_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Description == "" {
		respondError(w, http.StatusBadRequest, "Descrição é obrigatória")
		return
	}
	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "Valor deve ser maior que zero")
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Data de vencimento inválida. Use YYYY-MM-DD")
		return
	}

	if req.Installments <= 0 {
		req.Installments = 1
	}
	if req.Recurrence == "" {
		req.Recurrence = payables.RecurrenceNone
	}

	p := &payables.Payable{
		TenantID:          tenantID,
		Description:       req.Description,
		SupplierName:      req.SupplierName,
		SupplierCNPJ:      req.SupplierCNPJ,
		Category:          req.Category,
		Amount:            req.Amount,
		DueDate:           dueDate,
		TotalInstallments: req.Installments,
		Recurrence:        req.Recurrence,
		Notes:             req.Notes,
		CostCenter:        req.CostCenter,
		NFEKey:            req.NFEKey,
		CreatedBy:         userID,
	}

	created, err := h.svc.Create(r.Context(), p)
	if err != nil {
		switch err {
		case payables.ErrInvalidAmount:
			respondError(w, http.StatusBadRequest, "Valor inválido")
		case payables.ErrInvalidDueDate:
			respondError(w, http.StatusBadRequest, "Data de vencimento inválida")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	formatted := make([]map[string]interface{}, len(created))
	for i, item := range created {
		formatted[i] = formatPayable(item)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":      "Conta(s) criada(s) com sucesso",
		"payables":     formatted,
		"count":        len(formatted),
		"total_amount": req.Amount,
	})
}

// List GET /api/v1/payables
func (h *PayablesHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	q := r.URL.Query()

	filter := payables.PayableFilter{
		Status:   q.Get("status"),
		Category: q.Get("category"),
		Limit:    50,
	}

	if q.Get("overdue") == "true" {
		filter.Overdue = true
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}
	if from := q.Get("due_from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.DueFrom = &t
		}
	}
	if to := q.Get("due_to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.DueTo = &t
		}
	}

	list, err := h.svc.List(r.Context(), tenantID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(list))
	for i, item := range list {
		formatted[i] = formatPayable(item)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payables": formatted,
		"count":    len(formatted),
	})
}

// GetByID GET /api/v1/payables/{id}
func (h *PayablesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	p, err := h.svc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Conta não encontrada")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payable": formatPayable(p),
	})
}

// Pay POST /api/v1/payables/{id}/pay
// Realiza baixa da conta.
func (h *PayablesHandler) Pay(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req struct {
		AmountPaid    float64 `json:"amount_paid"`
		PaymentMethod string  `json:"payment_method"` // pix, boleto, cartao, dinheiro
		PaidAt        string  `json:"paid_at"`        // YYYY-MM-DD (opcional)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	var paidAt *time.Time
	if req.PaidAt != "" {
		if t, err := time.Parse("2006-01-02", req.PaidAt); err == nil {
			paidAt = &t
		}
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "pix"
	}

	paid, err := h.svc.Pay(r.Context(), tenantID, id, req.AmountPaid, req.PaymentMethod, paidAt)
	if err != nil {
		switch err {
		case payables.ErrPayableNotFound:
			respondError(w, http.StatusNotFound, "Conta não encontrada")
		case payables.ErrAlreadyPaid:
			respondError(w, http.StatusConflict, "Esta conta já foi paga")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Pagamento registrado com sucesso!",
		"payable": formatPayable(paid),
	})
}

// Cancel DELETE /api/v1/payables/{id}
func (h *PayablesHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Cancelado pelo usuário"
	}

	if err := h.svc.Cancel(r.Context(), tenantID, id, req.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Conta cancelada",
	})
}

// Summary GET /api/v1/payables/summary
func (h *PayablesHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	summary, err := h.svc.GetSummary(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// CashFlow GET /api/v1/payables/cashflow
func (h *PayablesHandler) CashFlow(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	q := r.URL.Query()

	from := time.Now()
	to := from.AddDate(0, 1, 0) // próximo mês por padrão

	if f := q.Get("from"); f != "" {
		if t, err := time.Parse("2006-01-02", f); err == nil {
			from = t
		}
	}
	if t := q.Get("to"); t != "" {
		if parsed, err := time.Parse("2006-01-02", t); err == nil {
			to = parsed
		}
	}

	items, err := h.svc.GetCashFlow(r.Context(), tenantID, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calcular totais
	var totalPending, totalPaid float64
	for _, item := range items {
		if item.Status == payables.StatusPaid {
			totalPaid += item.Amount
		} else {
			totalPending += item.Amount
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"period": map[string]string{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"items":         items,
		"total_pending": totalPending,
		"total_paid":    totalPaid,
		"total":         totalPending + totalPaid,
	})
}

// ── HELPERS ───────────────────────────────────────────────────────────────────

func formatPayable(p *payables.Payable) map[string]interface{} {
	result := map[string]interface{}{
		"id":                 p.ID,
		"description":        p.Description,
		"supplier_name":      p.SupplierName,
		"supplier_cnpj":      p.SupplierCNPJ,
		"category":           p.Category,
		"amount":             p.Amount,
		"amount_paid":        p.AmountPaid,
		"due_date":           p.DueDate.Format("2006-01-02"),
		"payment_method":     p.PaymentMethod,
		"installment":        p.Installment,
		"total_installments": p.TotalInstallments,
		"recurrence":         p.Recurrence,
		"status":             p.Status,
		"notes":              p.Notes,
		"cost_center":        p.CostCenter,
		"nfe_key":            p.NFEKey,
		"is_overdue":         p.Status == payables.StatusOverdue || (p.Status == payables.StatusPending && time.Now().After(p.DueDate)),
		"days_until_due":     int(time.Until(p.DueDate).Hours() / 24),
		"created_at":         p.CreatedAt.Format(time.RFC3339),
	}

	if p.PaidAt != nil {
		result["paid_at"] = p.PaidAt.Format("2006-01-02")
	}

	return result
}
