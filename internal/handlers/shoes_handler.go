package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/shoes"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

// ShoesDataProvider provides data access for the shoes handler.
type ShoesDataProvider interface {
	ListGrids(ctx context.Context, tenantID string) ([]*shoes.ProductGrid, error)
	GetByModel(ctx context.Context, tenantID, modelCode string) (*shoes.ProductGrid, error)
	Create(ctx context.Context, grid *shoes.ProductGrid) error
	UpdateStock(ctx context.Context, tenantID, sku string, delta int) error
	GetCommissions(ctx context.Context, tenantID string) ([]shoes.CommissionRule, error)
	SumBySeller(ctx context.Context, tenantID, sellerID string, from, to time.Time) (float64, error)
}

type ShoesHandler struct {
	svc  *shoes.GridService
	data ShoesDataProvider
}

func NewShoesHandler(svc *shoes.GridService, data ShoesDataProvider) *ShoesHandler {
	return &ShoesHandler{svc: svc, data: data}
}

func (h *ShoesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/shoes/grids", h.ListGrids)
	mux.HandleFunc("GET /api/v1/shoes/grids/{model}", h.GetGrid)
	mux.HandleFunc("POST /api/v1/shoes/grids", h.CreateGrid)
	mux.HandleFunc("PATCH /api/v1/shoes/stock", h.UpdateStock)
	mux.HandleFunc("GET /api/v1/shoes/commissions", h.GetCommissions)
}

func (h *ShoesHandler) ListGrids(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	grids, _ := h.data.ListGrids(r.Context(), tenantID)
	result := make([]map[string]interface{}, len(grids))
	for i, g := range grids {
		s := shoes.Summary(g)
		result[i] = map[string]interface{}{
			"id": g.ID, "model_code": g.ModelCode, "model_name": g.ModelName, "brand": g.Brand,
			"colors": g.Colors, "sizes": g.Sizes, "total_skus": s.TotalSKUs,
			"in_stock_skus": s.InStockSKUs, "total_units": s.TotalUnits, "total_value": s.TotalValue,
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"grids": result, "count": len(result)})
}

func (h *ShoesHandler) GetGrid(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	model := r.PathValue("model")
	grid, err := h.data.GetByModel(r.Context(), tenantID, model)
	if err != nil { respondError(w, http.StatusNotFound, err.Error()); return }
	s := shoes.Summary(grid)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"grid": grid, "summary": s,
	})
}

func (h *ShoesHandler) CreateGrid(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	var req struct {
		ModelCode string   `json:"model_code"`
		ModelName string   `json:"model_name"`
		Brand     string   `json:"brand"`
		Colors    []string `json:"colors"`
		Sizes     []string `json:"sizes"`
		BasePrice float64  `json:"base_price"`
		BaseCost  float64  `json:"base_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	grid := shoes.BuildGrid(tenantID, req.ModelCode, req.ModelName, req.Colors, req.Sizes, req.BasePrice, req.BaseCost)
	grid.Brand = req.Brand
	if err := h.data.Create(r.Context(), grid); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]interface{}{"grid": grid})
}

func (h *ShoesHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	var req struct {
		SKU   string `json:"sku"`
		Delta int    `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondError(w, http.StatusBadRequest, "JSON invalido"); return }
	if err := h.data.UpdateStock(r.Context(), tenantID, req.SKU, req.Delta); err != nil { respondError(w, http.StatusBadRequest, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"updated": true})
}

func (h *ShoesHandler) GetCommissions(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	commissions, _ := h.data.GetCommissions(r.Context(), tenantID)
	results := make([]map[string]interface{}, len(commissions))
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for i, c := range commissions {
		cr, _ := shoes.CalculateCommission(r.Context(), h.data, c, monthStart, now)
		results[i] = map[string]interface{}{
			"seller_name": c.SellerName, "base_rate": c.BaseRate * 100, "bonus_rate": c.BonusRate * 100,
			"monthly_target": c.MonthlTarget, "total_sales": cr.TotalSales,
			"base_commission": cr.BaseCommission, "bonus_commission": cr.BonusCommission,
			"total_commission": cr.TotalCommission, "met_achieved": cr.MetAchieved,
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"commissions": results})
}
