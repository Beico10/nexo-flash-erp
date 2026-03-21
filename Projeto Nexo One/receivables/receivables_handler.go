// Package handlers — endpoints de Contas a Receber.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nexoone/nexo-one/internal/receivables"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type ReceivablesHandler struct {
	svc *receivables.Service
}

func NewReceivablesHandler(svc *receivables.Service) *ReceivablesHandler {
	return &ReceivablesHandler{svc: svc}
}

func (h *ReceivablesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/receivables", h.Create)
	mux.HandleFunc("GET /api/v1/receivables", h.List)
	mux.HandleFunc("GET /api/v1/receivables/summary", h.Summary)
	mux.HandleFunc("GET /api/v1/receivables/{id}", h.GetByID)
	mux.HandleFunc("POST /api/v1/receivables/{id}/receive", h.Receive)
	mux.HandleFunc("DELETE /api/v1/receivables/{id}", h.Cancel)
}

func (h *ReceivablesHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		Description      string  `json:"description"`
		CustomerName     string  `json:"customer_name"`
		CustomerPhone    string  `json:"customer_phone"`
		CustomerDocument string  `json:"customer_document"`
		Category         string  `json:"category"`
		Amount           float64 `json:"amount"`
		DueDate          string  `json:"due_date"`
		Installments     int     `json:"installments"`
		Recurrence       string  `json:"recurrence"`
		Notes            string  `json:"notes"`
		ReferenceID      string  `json:"reference_id"`
		ReferenceType    string  `json:"reference_type"`
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
		respondError(w, http.StatusBadRequest, "Data inválida. Use YYYY-MM-DD")
		return
	}

	if req.Installments <= 0 {
		req.Installments = 1
	}
	if req.Recurrence == "" {
		req.Recurrence = receivables.RecurrenceNone
	}

	rec := &receivables.Receivable{
		TenantID:          tenantID,
		Description:       req.Description,
		CustomerName:      req.CustomerName,
		CustomerPhone:     req.CustomerPhone,
		CustomerDocument:  req.CustomerDocument,
		Category:          req.Category,
		Amount:            req.Amount,
		DueDate:           dueDate,
		TotalInstallments: req.Installments,
		Recurrence:        req.Recurrence,
		Notes:             req.Notes,
		ReferenceID:       req.ReferenceID,
		ReferenceType:     req.ReferenceType,
		CreatedBy:         userID,
	}

	created, err := h.svc.Create(r.Context(), rec)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(created))
	for i, item := range created {
		formatted[i] = formatReceivable(item)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":      "Conta(s) criada(s) com sucesso",
		"receivables":  formatted,
		"count":        len(formatted),
		"total_amount": req.Amount,
	})
}

func (h *ReceivablesHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	q := r.URL.Query()

	filter := receivables.ReceivableFilter{
		Status:   q.Get("status"),
		Category: q.Get("category"),
		Customer: q.Get("customer"),
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

	list, err := h.svc.List(r.Context(), tenantID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(list))
	for i, item := range list {
		formatted[i] = formatReceivable(item)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"receivables": formatted,
		"count":       len(formatted),
	})
}

func (h *ReceivablesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	rec, err := h.svc.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Conta não encontrada")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"receivable": formatReceivable(rec),
	})
}

func (h *ReceivablesHandler) Receive(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var req struct {
		AmountReceived float64 `json:"amount_received"`
		PaymentMethod  string  `json:"payment_method"`
		ReceivedAt     string  `json:"received_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	var receivedAt *time.Time
	if req.ReceivedAt != "" {
		if t, err := time.Parse("2006-01-02", req.ReceivedAt); err == nil {
			receivedAt = &t
		}
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "pix"
	}

	received, err := h.svc.Receive(r.Context(), tenantID, id, req.AmountReceived, req.PaymentMethod, receivedAt)
	if err != nil {
		switch err {
		case receivables.ErrNotFound:
			respondError(w, http.StatusNotFound, "Conta não encontrada")
		case receivables.ErrAlreadyReceived:
			respondError(w, http.StatusConflict, "Conta já foi recebida")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Recebimento registrado com sucesso!",
		"receivable": formatReceivable(received),
	})
}

func (h *ReceivablesHandler) Cancel(w http.ResponseWriter, r *http.Request) {
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

	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "Conta cancelada"})
}

func (h *ReceivablesHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	summary, err := h.svc.GetSummary(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func formatReceivable(r *receivables.Receivable) map[string]interface{} {
	result := map[string]interface{}{
		"id":                 r.ID,
		"description":        r.Description,
		"customer_name":      r.CustomerName,
		"customer_phone":     r.CustomerPhone,
		"customer_document":  r.CustomerDocument,
		"category":           r.Category,
		"amount":             r.Amount,
		"amount_received":    r.AmountReceived,
		"due_date":           r.DueDate.Format("2006-01-02"),
		"payment_method":     r.PaymentMethod,
		"installment":        r.Installment,
		"total_installments": r.TotalInstallments,
		"recurrence":         r.Recurrence,
		"status":             r.Status,
		"notes":              r.Notes,
		"reference_id":       r.ReferenceID,
		"reference_type":     r.ReferenceType,
		"is_overdue":         r.Status == receivables.StatusOverdue || (r.Status == receivables.StatusPending && time.Now().After(r.DueDate)),
		"days_until_due":     int(time.Until(r.DueDate).Hours() / 24),
		"created_at":         r.CreatedAt.Format(time.RFC3339),
	}
	if r.ReceivedAt != nil {
		result["received_at"] = r.ReceivedAt.Format("2006-01-02")
	}
	return result
}
