package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/bakery"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type BakeryHandler struct {
	service *bakery.PDVService
}

func NewBakeryHandler(s *bakery.PDVService) *BakeryHandler {
	return &BakeryHandler{service: s}
}

func (h *BakeryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/bakery/sale", h.CompleteSale)
	mux.HandleFunc("GET /api/v1/bakery/scale/{scaleId}", h.ReadScale)
	mux.HandleFunc("POST /api/v1/bakery/loss", h.RegisterLoss)
	mux.HandleFunc("GET /api/v1/bakery/loss/summary", h.LossSummary)
	mux.HandleFunc("GET /api/v1/bakery/products", h.ListProducts)
}

func (h *BakeryHandler) ReadScale(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	scaleID := r.PathValue("scaleId")
	item, err := h.service.ReadFromScale(r.Context(), tenantID, scaleID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *BakeryHandler) CompleteSale(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	var req struct {
		Items []struct {
			ProductID string  `json:"product_id"`
			Quantity  float64 `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
			Discount  float64 `json:"discount"`
		} `json:"items"`
		PaymentMethod string  `json:"payment_method"`
		Discount      float64 `json:"discount"`
		OperatorID    string  `json:"operator_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}
	var items []bakery.PDVSaleItem
	for _, i := range req.Items {
		items = append(items, bakery.PDVSaleItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			UnitPrice: i.UnitPrice,
			Discount:  i.Discount,
		})
	}
	sale := &bakery.PDVSale{
		TenantID:      tenantID,
		Items:         items,
		Discount:      req.Discount,
		PaymentMethod: req.PaymentMethod,
		OperatorID:    req.OperatorID,
	}
	if err := h.service.CompleteSale(r.Context(), sale); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, sale)
}

func (h *BakeryHandler) RegisterLoss(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	var req struct {
		ProductID string            `json:"product_id"`
		Quantity  float64           `json:"quantity"`
		Unit      string            `json:"unit"`
		Reason    bakery.LossReason `json:"reason"`
		CostValue float64           `json:"cost_value"`
		Notes     string            `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON invalido")
		return
	}
	claims, _ := middleware.GetClaims(r.Context())
	loss := &bakery.LossRecord{
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		Unit:       req.Unit,
		Reason:     req.Reason,
		CostValue:  req.CostValue,
		Notes:      req.Notes,
		RecordedBy: claims.UserID,
	}
	if err := h.service.RegisterLoss(r.Context(), tenantID, loss); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, loss)
}

func (h *BakeryHandler) LossSummary(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "parametro 'from' invalido (use YYYY-MM-DD)")
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "parametro 'to' invalido (use YYYY-MM-DD)")
		return
	}
	summary, err := h.service.LossSummary(r.Context(), tenantID, from, to)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (h *BakeryHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	products, err := h.service.ListProducts(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": products, "total": len(products)})
}
