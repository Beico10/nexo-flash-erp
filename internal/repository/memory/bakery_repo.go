package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/modules/bakery"
)

type BakeryRepo struct {
	mu       sync.RWMutex
	products map[string]*bakery.BakeryProduct
	sales    map[string]*bakery.PDVSale
	losses   []*bakery.LossRecord
}

func NewBakeryRepo() *BakeryRepo {
	return &BakeryRepo{
		products: make(map[string]*bakery.BakeryProduct),
		sales:    make(map[string]*bakery.PDVSale),
	}
}

func (r *BakeryRepo) GetProduct(_ context.Context, tenantID, id string) (*bakery.BakeryProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok || p.TenantID != tenantID {
		return nil, fmt.Errorf("produto %s nao encontrado", id)
	}
	return p, nil
}

func (r *BakeryRepo) GetProductByPLU(_ context.Context, tenantID, plu string) (*bakery.BakeryProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.products {
		if p.TenantID == tenantID && p.ScaleCode == plu {
			return p, nil
		}
	}
	return nil, fmt.Errorf("produto PLU %s nao encontrado", plu)
}

func (r *BakeryRepo) ListProducts(_ context.Context, tenantID string) ([]*bakery.BakeryProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*bakery.BakeryProduct
	for _, p := range r.products {
		if p.TenantID == tenantID && p.Active {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *BakeryRepo) CreateSale(_ context.Context, sale *bakery.PDVSale) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sale.ID = uuid.New().String()
	r.sales[sale.ID] = sale
	return nil
}

func (r *BakeryRepo) CreateLossRecord(_ context.Context, loss *bakery.LossRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	loss.ID = uuid.New().String()
	r.losses = append(r.losses, loss)
	return nil
}

func (r *BakeryRepo) GetLossByPeriod(_ context.Context, tenantID string, from, to time.Time) ([]*bakery.LossRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*bakery.LossRecord
	for _, l := range r.losses {
		if l.TenantID == tenantID && !l.RecordedAt.Before(from) && !l.RecordedAt.After(to) {
			result = append(result, l)
		}
	}
	return result, nil
}
