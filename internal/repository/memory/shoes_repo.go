package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/shoes"
)

type ShoesRepo struct {
	mu          sync.RWMutex
	grids       []*shoes.ProductGrid
	commissions []shoes.CommissionRule
	sales       map[string]float64 // sellerID -> total
}

func NewShoesRepo() *ShoesRepo {
	r := &ShoesRepo{sales: make(map[string]float64)}
	r.seed()
	return r
}

func (r *ShoesRepo) seed() {
	tid := "00000000-0000-0000-0000-000000000001"

	g1 := shoes.BuildGrid(tid, "SANDAL001", "Sandalia Conforto Premium", []string{"Preto", "Marrom", "Nude"}, []string{"34", "35", "36", "37", "38", "39", "40"}, 129.90, 45.00)
	g1.ID = "grid-1"
	g1.Brand = "Nexo Comfort"
	g1.NCMCode = "64029990"
	// Seed some stock
	for color := range g1.Grid {
		for size := range g1.Grid[color] {
			cell := g1.Grid[color][size]
			cell.Stock = 5 + len(size)
			g1.Grid[color][size] = cell
		}
	}

	g2 := shoes.BuildGrid(tid, "BOOT002", "Bota Industrial CA", []string{"Preto"}, []string{"38", "39", "40", "41", "42", "43", "44"}, 189.90, 72.00)
	g2.ID = "grid-2"
	g2.Brand = "Nexo Safe"
	g2.NCMCode = "64035190"
	for size := range g2.Grid["Preto"] {
		cell := g2.Grid["Preto"][size]
		cell.Stock = 10 + len(size)
		g2.Grid["Preto"][size] = cell
	}

	g3 := shoes.BuildGrid(tid, "SNEAK003", "Tenis Casual Urban", []string{"Branco", "Preto", "Azul", "Verde"}, []string{"36", "37", "38", "39", "40", "41", "42"}, 219.90, 85.00)
	g3.ID = "grid-3"
	g3.Brand = "Nexo Street"
	g3.NCMCode = "64041190"
	for color := range g3.Grid {
		for size := range g3.Grid[color] {
			cell := g3.Grid[color][size]
			cell.Stock = 3 + len(color)
			g3.Grid[color][size] = cell
		}
	}

	r.grids = []*shoes.ProductGrid{g1, g2, g3}

	r.commissions = []shoes.CommissionRule{
		{ID: "com-1", TenantID: tid, SellerID: "seller-1", SellerName: "Carlos Silva", BaseRate: 0.05, BonusRate: 0.02, MonthlTarget: 15000, Active: true},
		{ID: "com-2", TenantID: tid, SellerID: "seller-2", SellerName: "Ana Oliveira", BaseRate: 0.06, BonusRate: 0.03, MonthlTarget: 20000, Active: true},
	}
	r.sales["seller-1"] = 18500.00
	r.sales["seller-2"] = 22300.00
}

func (r *ShoesRepo) Create(_ context.Context, grid *shoes.ProductGrid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if grid.ID == "" { grid.ID = fmt.Sprintf("grid-%d", time.Now().UnixNano()) }
	r.grids = append(r.grids, grid)
	return nil
}

func (r *ShoesRepo) Update(_ context.Context, grid *shoes.ProductGrid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, g := range r.grids {
		if g.ID == grid.ID { r.grids[i] = grid; return nil }
	}
	return fmt.Errorf("grade nao encontrada")
}

func (r *ShoesRepo) GetByModel(_ context.Context, tenantID, modelCode string) (*shoes.ProductGrid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, g := range r.grids {
		if g.TenantID == tenantID && g.ModelCode == modelCode { return g, nil }
	}
	return nil, fmt.Errorf("grade nao encontrada")
}

func (r *ShoesRepo) UpdateStock(_ context.Context, tenantID, sku string, delta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.grids {
		if g.TenantID != tenantID { continue }
		for color := range g.Grid {
			for size, cell := range g.Grid[color] {
				if cell.SKU == sku {
					cell.Stock += delta
					g.Grid[color][size] = cell
					return nil
				}
			}
		}
	}
	return fmt.Errorf("SKU nao encontrado")
}

func (r *ShoesRepo) GetBySKU(_ context.Context, tenantID, sku string) (*shoes.GridCell, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, g := range r.grids {
		if g.TenantID != tenantID { continue }
		for _, sizes := range g.Grid {
			for _, cell := range sizes {
				if cell.SKU == sku { return &cell, nil }
			}
		}
	}
	return nil, fmt.Errorf("SKU nao encontrado")
}

func (r *ShoesRepo) ListGrids(_ context.Context, tenantID string) ([]*shoes.ProductGrid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*shoes.ProductGrid
	for _, g := range r.grids {
		if g.TenantID == tenantID { result = append(result, g) }
	}
	return result, nil
}

func (r *ShoesRepo) SumBySeller(_ context.Context, _, sellerID string, _, _ time.Time) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sales[sellerID], nil
}

func (r *ShoesRepo) GetCommissions(_ context.Context, tenantID string) ([]shoes.CommissionRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []shoes.CommissionRule
	for _, c := range r.commissions {
		if c.TenantID == tenantID { result = append(result, c) }
	}
	return result, nil
}
