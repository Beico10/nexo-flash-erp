// Package handlers — endpoints de Estoque.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexoone/nexo-one/internal/inventory"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type InventoryHandler struct {
	svc *inventory.Service
}

func NewInventoryHandler(svc *inventory.Service) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) RegisterRoutes(mux *http.ServeMux) {
	// Configuração do nicho (camaleão)
	mux.HandleFunc("GET /api/v1/inventory/config", h.GetConfig)

	// Produtos
	mux.HandleFunc("POST /api/v1/inventory/products", h.CreateProduct)
	mux.HandleFunc("GET /api/v1/inventory/products", h.ListProducts)
	mux.HandleFunc("GET /api/v1/inventory/products/summary", h.Summary)
	mux.HandleFunc("GET /api/v1/inventory/products/low-stock", h.LowStock)
	mux.HandleFunc("GET /api/v1/inventory/products/{id}", h.GetProduct)
	mux.HandleFunc("PUT /api/v1/inventory/products/{id}", h.UpdateProduct)
	mux.HandleFunc("DELETE /api/v1/inventory/products/{id}", h.DeleteProduct)

	// Movimentações
	mux.HandleFunc("POST /api/v1/inventory/products/{id}/add", h.AddStock)
	mux.HandleFunc("POST /api/v1/inventory/products/{id}/remove", h.RemoveStock)
	mux.HandleFunc("POST /api/v1/inventory/products/{id}/adjust", h.AdjustStock)
	mux.HandleFunc("GET /api/v1/inventory/products/{id}/movements", h.ListMovements)

	// Entrada de NF-e
	mux.HandleFunc("POST /api/v1/inventory/nfe-entry", h.ProcessNFeEntry)
}

// GetConfig GET /api/v1/inventory/config
// Retorna configuração do estoque para o nicho do tenant (camaleão).
func (h *InventoryHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	businessType := "mechanic"
	if claims != nil {
		businessType = claims.BusinessType
	}

	config := inventory.NichoCamaleao(businessType)
	respondJSON(w, http.StatusOK, config)
}

func (h *InventoryHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())

	var p inventory.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	p.TenantID = tenantID
	if claims != nil {
		p.BusinessType = claims.BusinessType
		p.CreatedBy = claims.UserID
	}

	if err := h.svc.CreateProduct(r.Context(), &p); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Produto cadastrado com sucesso",
		"product": formatProduct(&p),
	})
}

func (h *InventoryHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	q := r.URL.Query()

	filter := inventory.ProductFilter{
		Category:   q.Get("category"),
		Search:     q.Get("search"),
		LowStock:   q.Get("low_stock") == "true",
		OutOfStock: q.Get("out_of_stock") == "true",
		Limit:      50,
	}

	if limit, err := strconv.Atoi(q.Get("limit")); err == nil && limit > 0 {
		filter.Limit = limit
	}

	products, err := h.svc.ListProducts(r.Context(), tenantID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(products))
	for i, p := range products {
		formatted[i] = formatProduct(p)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"products": formatted,
		"count":    len(formatted),
	})
}

func (h *InventoryHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	p, err := h.svc.GetProduct(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Produto não encontrado")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"product": formatProduct(p)})
}

func (h *InventoryHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	var p inventory.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	p.TenantID = tenantID
	p.ID = id

	if err := h.svc.UpdateProduct(r.Context(), &p); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "Produto atualizado"})
}

func (h *InventoryHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	p, err := h.svc.GetProduct(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Produto não encontrado")
		return
	}

	p.IsActive = false
	h.svc.UpdateProduct(r.Context(), p)

	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "Produto desativado"})
}

func (h *InventoryHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	id := r.PathValue("id")
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		Quantity      float64 `json:"quantity"`
		UnitCost      float64 `json:"unit_cost"`
		ReferenceID   string  `json:"reference_id"`
		ReferenceType string  `json:"reference_type"`
		NFeKey        string  `json:"nfe_key"`
		Notes         string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	movement, err := h.svc.AddStock(r.Context(), tenantID, id,
		req.Quantity, req.UnitCost,
		req.ReferenceID, req.ReferenceType, req.NFeKey,
		req.Notes, userID,
	)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Entrada registrada com sucesso",
		"movement": formatMovement(movement),
	})
}

func (h *InventoryHandler) RemoveStock(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	id := r.PathValue("id")
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		Quantity      float64 `json:"quantity"`
		ReferenceID   string  `json:"reference_id"`
		ReferenceType string  `json:"reference_type"`
		Notes         string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	movement, err := h.svc.RemoveStock(r.Context(), tenantID, id,
		req.Quantity, req.ReferenceID, req.ReferenceType, req.Notes, userID,
	)
	if err != nil {
		switch err {
		case inventory.ErrInsufficientStock:
			respondError(w, http.StatusConflict, err.Error())
		case inventory.ErrProductNotFound:
			respondError(w, http.StatusNotFound, "Produto não encontrado")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Saída registrada com sucesso",
		"movement": formatMovement(movement),
	})
}

func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	id := r.PathValue("id")
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		NewQuantity float64 `json:"new_quantity"`
		Notes       string  `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	movement, err := h.svc.AdjustStock(r.Context(), tenantID, id, req.NewQuantity, req.Notes, userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Ajuste registrado com sucesso",
		"movement": formatMovement(movement),
	})
}

func (h *InventoryHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	movements, err := h.svc.ListMovements(r.Context(), tenantID, id, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(movements))
	for i, m := range movements {
		formatted[i] = formatMovement(m)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"movements": formatted,
		"count":     len(formatted),
	})
}

func (h *InventoryHandler) Summary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	summary, err := h.svc.GetSummary(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (h *InventoryHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	filter := inventory.ProductFilter{LowStock: true, Limit: 100}
	products, err := h.svc.ListProducts(r.Context(), tenantID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	formatted := make([]map[string]interface{}, len(products))
	for i, p := range products {
		formatted[i] = formatProduct(p)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"products": formatted,
		"count":    len(formatted),
	})
}

func (h *InventoryHandler) ProcessNFeEntry(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		NFeKey string              `json:"nfe_key"`
		Items  []inventory.NFeItem `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	movements, err := h.svc.ProcessNFeEntry(r.Context(), tenantID, userID, req.Items, req.NFeKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Entrada NF-e processada com sucesso",
		"movements": len(movements),
	})
}

func formatProduct(p *inventory.Product) map[string]interface{} {
	return map[string]interface{}{
		"id":            p.ID,
		"code":          p.Code,
		"barcode":       p.Barcode,
		"name":          p.Name,
		"description":   p.Description,
		"category":      p.Category,
		"unit":          p.Unit,
		"quantity":      p.Quantity,
		"min_quantity":  p.MinQuantity,
		"max_quantity":  p.MaxQuantity,
		"location":      p.Location,
		"cost_price":    p.CostPrice,
		"sale_price":    p.SalePrice,
		"total_value":   p.TotalValue,
		"ncm":           p.NCM,
		"is_active":     p.IsActive,
		"is_low_stock":  p.MinQuantity > 0 && p.Quantity <= p.MinQuantity,
		"is_out_of_stock": p.Quantity <= 0,
		"extra":         p.Extra,
		"created_at":    p.CreatedAt,
	}
}

func formatMovement(m *inventory.Movement) map[string]interface{} {
	return map[string]interface{}{
		"id":             m.ID,
		"product_id":     m.ProductID,
		"product_name":   m.ProductName,
		"type":           m.Type,
		"quantity":       m.Quantity,
		"unit_cost":      m.UnitCost,
		"total_cost":     m.TotalCost,
		"previous_stock": m.PreviousStock,
		"new_stock":      m.NewStock,
		"previous_cmp":   m.PreviousCMP,
		"new_cmp":        m.NewCMP,
		"reference_id":   m.ReferenceID,
		"reference_type": m.ReferenceType,
		"notes":          m.Notes,
		"created_at":     m.CreatedAt,
	}
}
