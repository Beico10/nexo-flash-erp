package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nexoone/nexo-one/internal/modules/industry"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

// IndustryDataProvider provides data access for the industry handler.
type IndustryDataProvider interface {
	ListByStatus(ctx context.Context, tenantID string, status industry.ProductionOrderStatus) ([]*industry.ProductionOrder, error)
	GetByID(ctx context.Context, tenantID, id string) (*industry.ProductionOrder, error)
	Create(ctx context.Context, op *industry.ProductionOrder) error
	Update(ctx context.Context, op *industry.ProductionOrder) error
	ListBOMs(ctx context.Context, tenantID string) ([]*industry.BOM, error)
	ListMaterials() map[string]float64
}

type IndustryHandler struct {
	svc  *industry.PCPService
	data IndustryDataProvider
}

func NewIndustryHandler(svc *industry.PCPService, data IndustryDataProvider) *IndustryHandler {
	return &IndustryHandler{svc: svc, data: data}
}

func (h *IndustryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/industry/orders", h.ListOrders)
	mux.HandleFunc("GET /api/v1/industry/orders/{id}", h.GetOrder)
	mux.HandleFunc("POST /api/v1/industry/orders", h.CreateOrder)
	mux.HandleFunc("PATCH /api/v1/industry/orders/{id}/status", h.UpdateStatus)
	mux.HandleFunc("GET /api/v1/industry/boms", h.ListBOMs)
	mux.HandleFunc("GET /api/v1/industry/materials", h.ListMaterials)
	mux.HandleFunc("POST /api/v1/industry/bom/explode", h.ExplodeBOM)
}

func (h *IndustryHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	status := industry.ProductionOrderStatus(r.URL.Query().Get("status"))
	orders, err := h.data.ListByStatus(r.Context(), tenantID, status)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }

	result := make([]map[string]interface{}, len(orders))
	for i, o := range orders {
		result[i] = map[string]interface{}{
			"id": o.ID, "number": o.Number, "product_name": o.ProductName,
			"planned_qty": o.PlannedQty, "produced_qty": o.ProducedQty, "unit": o.Unit,
			"status": o.Status, "planned_start": o.PlannedStart, "planned_end": o.PlannedEnd,
			"progress": 0,
		}
		if o.PlannedQty > 0 { result[i]["progress"] = int(o.ProducedQty / o.PlannedQty * 100) }
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"orders": result, "count": len(result)})
}

func (h *IndustryHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	order, err := h.data.GetByID(r.Context(), tenantID, id)
	if err != nil { respondError(w, http.StatusNotFound, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"order": order})
}

func (h *IndustryHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	var req struct {
		ProductID   string  `json:"product_id"`
		ProductName string  `json:"product_name"`
		BOMID       string  `json:"bom_id"`
		PlannedQty  float64 `json:"planned_qty"`
		Unit        string  `json:"unit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	op := &industry.ProductionOrder{
		TenantID: tenantID, ProductID: req.ProductID, ProductName: req.ProductName,
		BOMID: req.BOMID, PlannedQty: req.PlannedQty, Unit: req.Unit, Status: industry.POStatusPlanned,
	}
	if err := h.data.Create(r.Context(), op); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]interface{}{"order": op})
}

func (h *IndustryHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	var req struct { Status string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	order, err := h.data.GetByID(r.Context(), tenantID, id)
	if err != nil { respondError(w, http.StatusNotFound, err.Error()); return }
	order.Status = industry.ProductionOrderStatus(req.Status)
	h.data.Update(r.Context(), order)
	respondJSON(w, http.StatusOK, map[string]interface{}{"updated": true})
}

func (h *IndustryHandler) ListBOMs(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	boms, _ := h.data.ListBOMs(r.Context(), tenantID)
	result := make([]map[string]interface{}, len(boms))
	for i, b := range boms {
		result[i] = map[string]interface{}{
			"id": b.ID, "product_name": b.ProductName, "version": b.Version,
			"total_cost": b.TotalCost, "items_count": len(b.Items),
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"boms": result})
}

func (h *IndustryHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	materials := h.data.ListMaterials()
	respondJSON(w, http.StatusOK, map[string]interface{}{"materials": materials})
}

func (h *IndustryHandler) ExplodeBOM(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	var req struct {
		ProductID string  `json:"product_id"`
		Quantity  float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	reservations, totalCost, err := h.svc.ExplodeBOM(r.Context(), tenantID, req.ProductID, req.Quantity)
	if err != nil { respondError(w, http.StatusBadRequest, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reservations": reservations, "total_cost": totalCost,
	})
}
