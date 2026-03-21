// Package memory — repositório in-memory de Estoque.
package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/inventory"
)

type InventoryRepo struct {
	mu        sync.RWMutex
	products  map[string]*inventory.Product
	movements map[string][]*inventory.Movement // key: tenantID:productID
}

func NewInventoryRepo() *InventoryRepo {
	r := &InventoryRepo{
		products:  make(map[string]*inventory.Product),
		movements: make(map[string][]*inventory.Movement),
	}
	r.seed()
	return r
}

func (r *InventoryRepo) pKey(tenantID, id string) string { return tenantID + ":" + id }

func (r *InventoryRepo) seed() {
	now := time.Now()

	// Produtos demo — Mecânica
	products := []*inventory.Product{
		{
			ID: uuid.NewString(), TenantID: "demo", BusinessType: "mechanic",
			Code: "FLT-001", Name: "Filtro de Óleo — Bosch", Category: "filtros",
			Unit: inventory.UnitUN, Quantity: 8, MinQuantity: 5, MaxQuantity: 30,
			CostPrice: 18.50, SalePrice: 35.00, Location: "A1-P2",
			Extra: inventory.BusinessFields{ManufacturerCode: "0986AF0030", Application: "Fiat/VW 1.0-1.6", Brand: "Bosch"},
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo", BusinessType: "mechanic",
			Code: "OLE-001", Name: "Óleo Motor 5W30 Sintético 1L", Category: "fluidos",
			Unit: inventory.UnitL, Quantity: 2, MinQuantity: 10, MaxQuantity: 50,
			CostPrice: 28.00, SalePrice: 48.00, Location: "B2-P1",
			Extra: inventory.BusinessFields{Brand: "Mobil 1"},
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo", BusinessType: "mechanic",
			Code: "PAS-001", Name: "Pastilha de Freio Dianteira", Category: "freios",
			Unit: inventory.UnitUN, Quantity: 12, MinQuantity: 4, MaxQuantity: 20,
			CostPrice: 45.00, SalePrice: 89.00, Location: "C1-P3",
			Extra: inventory.BusinessFields{ManufacturerCode: "TRW-GDB1234", Application: "Gol/Palio/Uno", Brand: "TRW"},
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: uuid.NewString(), TenantID: "demo", BusinessType: "mechanic",
			Code: "VEL-001", Name: "Vela de Ignição NGK", Category: "elétrica",
			Unit: inventory.UnitUN, Quantity: 0, MinQuantity: 8, MaxQuantity: 40,
			CostPrice: 12.00, SalePrice: 22.00, Location: "A2-P1",
			Extra: inventory.BusinessFields{ManufacturerCode: "BKR5EGP", Application: "Universal", Brand: "NGK"},
			IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
	}

	// Calcular TotalValue
	for _, p := range products {
		p.TotalValue = p.Quantity * p.CostPrice
		r.products[r.pKey(p.TenantID, p.ID)] = p
	}
}

// ── PRODUTOS ──────────────────────────────────────────────────────────────────

func (r *InventoryRepo) CreateProduct(ctx context.Context, p *inventory.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.TotalValue = p.Quantity * p.CostPrice
	copy := *p
	r.products[r.pKey(p.TenantID, p.ID)] = &copy
	return nil
}

func (r *InventoryRepo) GetProduct(ctx context.Context, tenantID, id string) (*inventory.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[r.pKey(tenantID, id)]
	if !ok {
		return nil, inventory.ErrProductNotFound
	}
	copy := *p
	return &copy, nil
}

func (r *InventoryRepo) GetProductByCode(ctx context.Context, tenantID, code string) (*inventory.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := tenantID + ":"
	for k, p := range r.products {
		if strings.HasPrefix(k, prefix) && p.Code == code {
			copy := *p
			return &copy, nil
		}
	}
	return nil, inventory.ErrProductNotFound
}

func (r *InventoryRepo) ListProducts(ctx context.Context, tenantID string, filter inventory.ProductFilter) ([]*inventory.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := tenantID + ":"
	var result []*inventory.Product

	for k, p := range r.products {
		if !strings.HasPrefix(k, prefix) || !p.IsActive {
			continue
		}
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.LowStock && !(p.MinQuantity > 0 && p.Quantity <= p.MinQuantity) {
			continue
		}
		if filter.OutOfStock && p.Quantity > 0 {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(filter.Search)) {
			continue
		}
		copy := *p
		result = append(result, &copy)
	}

	// Ordenar por nome
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Name > result[j].Name {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InventoryRepo) UpdateProduct(ctx context.Context, p *inventory.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.products[r.pKey(p.TenantID, p.ID)]; !ok {
		return inventory.ErrProductNotFound
	}
	p.UpdatedAt = time.Now()
	p.TotalValue = p.Quantity * p.CostPrice
	copy := *p
	r.products[r.pKey(p.TenantID, p.ID)] = &copy
	return nil
}

func (r *InventoryRepo) DeleteProduct(ctx context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.products, r.pKey(tenantID, id))
	return nil
}

// ── MOVIMENTAÇÕES ─────────────────────────────────────────────────────────────

func (r *InventoryRepo) CreateMovement(ctx context.Context, m *inventory.Movement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	key := r.pKey(m.TenantID, m.ProductID)
	copy := *m
	r.movements[key] = append([]*inventory.Movement{&copy}, r.movements[key]...)
	return nil
}

func (r *InventoryRepo) ListMovements(ctx context.Context, tenantID, productID string, limit int) ([]*inventory.Movement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := r.pKey(tenantID, productID)
	all := r.movements[key]
	if limit > 0 && limit < len(all) {
		all = all[:limit]
	}
	result := make([]*inventory.Movement, len(all))
	for i, m := range all {
		copy := *m
		result[i] = &copy
	}
	return result, nil
}

// ── ALERTAS E RELATÓRIOS ──────────────────────────────────────────────────────

func (r *InventoryRepo) GetLowStockProducts(ctx context.Context, tenantID string) ([]*inventory.Product, error) {
	return r.ListProducts(ctx, tenantID, inventory.ProductFilter{LowStock: true, Limit: 100})
}

func (r *InventoryRepo) GetSummary(ctx context.Context, tenantID string) (*inventory.InventorySummary, error) {
	all, _ := r.ListProducts(ctx, tenantID, inventory.ProductFilter{Limit: 1000})

	summary := &inventory.InventorySummary{}
	for _, p := range all {
		summary.TotalProducts++
		summary.TotalValue += p.TotalValue
		if p.MinQuantity > 0 && p.Quantity <= p.MinQuantity {
			summary.LowStockCount++
		}
		if p.Quantity <= 0 {
			summary.OutOfStockCount++
		}
	}

	// Top 5 por valor
	top := make([]*inventory.Product, len(all))
	copy(top, all)
	for i := 0; i < len(top)-1; i++ {
		for j := i + 1; j < len(top); j++ {
			if top[i].TotalValue < top[j].TotalValue {
				top[i], top[j] = top[j], top[i]
			}
		}
	}
	if len(top) > 5 {
		top = top[:5]
	}
	summary.TopProducts = top

	return summary, nil
}

func (r *InventoryRepo) MarkAlertSent(ctx context.Context, tenantID, productID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.products[r.pKey(tenantID, productID)]; ok {
		p.AlertSent = true
	}
	return nil
}
