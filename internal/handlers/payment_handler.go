// Package handlers — BaaS payment handlers HTTP.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/baas"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type PaymentHandler struct {
	service *baas.PaymentService
}

func NewPaymentHandler(s *baas.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: s}
}

func (h *PaymentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/payments/pix", h.CreatePix)
	mux.HandleFunc("POST /api/v1/payments/boleto", h.CreateBoleto)
	mux.HandleFunc("POST /api/v1/webhooks/payment", h.Webhook)
}

// CreatePix gera um PIX dinâmico com QR Code.
// POST /api/v1/payments/pix
func (h *PaymentHandler) CreatePix(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	var req struct {
		Amount        float64 `json:"amount"`
		Description   string  `json:"description"`
		PayerName     string  `json:"payer_name"`
		PayerDocument string  `json:"payer_document"`
		ExpiresInHours int    `json:"expires_in_hours"`
		Split         []struct {
			AccountID  string  `json:"account_id"`
			Amount     float64 `json:"amount"`
			Percentage float64 `json:"percentage"`
		} `json:"split,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 24 // padrão: 24h
	}

	var split []baas.SplitRecipient
	for _, s := range req.Split {
		split = append(split, baas.SplitRecipient{
			AccountID:  s.AccountID,
			Amount:     s.Amount,
			Percentage: s.Percentage,
		})
	}

	charge, err := h.service.CreatePixCharge(r.Context(), tenantID, baas.PixChargeInput{
		Amount:          req.Amount,
		Description:     req.Description,
		PayerName:       req.PayerName,
		PayerDocument:   req.PayerDocument,
		ExpiresIn:       time.Duration(req.ExpiresInHours) * time.Hour,
		SplitRecipients: split,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, charge)
}

// CreateBoleto gera um boleto híbrido (boleto + PIX embutido).
// POST /api/v1/payments/boleto
func (h *PaymentHandler) CreateBoleto(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant não identificado")
		return
	}
	var req struct {
		Amount        float64 `json:"amount"`
		DueDate       string  `json:"due_date"` // YYYY-MM-DD
		Description   string  `json:"description"`
		PayerName     string  `json:"payer_name"`
		PayerDocument string  `json:"payer_document"`
		PayerAddress  string  `json:"payer_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "due_date inválido (use YYYY-MM-DD)")
		return
	}

	boleto, err := h.service.CreateBoleto(r.Context(), tenantID, baas.BoletoInput{
		Amount:        req.Amount,
		DueDate:       dueDate,
		Description:   req.Description,
		PayerName:     req.PayerName,
		PayerDocument: req.PayerDocument,
		PayerAddress:  req.PayerAddress,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, boleto)
}

// Webhook recebe confirmações de pagamento do gateway BaaS.
// POST /api/v1/webhooks/payment
// Este endpoint é público mas valida a assinatura HMAC do gateway.
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-Gateway-Signature")
	if signature == "" {
		respondError(w, http.StatusUnauthorized, "assinatura ausente")
		return
	}

	buf := make([]byte, r.ContentLength)
	n, _ := r.Body.Read(buf)
	payload := buf[:n]

	if err := h.service.ProcessWebhook(r.Context(), payload, signature); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Gateway espera 200 OK para confirmar recebimento
	w.WriteHeader(http.StatusOK)
}
