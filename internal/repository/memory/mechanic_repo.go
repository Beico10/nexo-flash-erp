// Package memory implementa repositorios in-memory para ambiente de preview.
// Em producao, substituir por postgres.* via injecao de dependencia no wire.go.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
)

type MechanicRepo struct {
	mu     sync.RWMutex
	orders map[string]*mechanic.ServiceOrder
}

func NewMechanicRepo() *MechanicRepo {
	return &MechanicRepo{orders: make(map[string]*mechanic.ServiceOrder)}
}

func (r *MechanicRepo) Create(_ context.Context, os *mechanic.ServiceOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	os.ID = uuid.New().String()
	os.CreatedAt = time.Now().UTC()
	os.UpdatedAt = os.CreatedAt
	r.orders[os.ID] = os
	return nil
}

func (r *MechanicRepo) Update(_ context.Context, os *mechanic.ServiceOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[os.ID]; !ok {
		return fmt.Errorf("OS %s nao encontrada", os.ID)
	}
	r.orders[os.ID] = os
	return nil
}

func (r *MechanicRepo) GetByID(_ context.Context, tenantID, id string) (*mechanic.ServiceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	os, ok := r.orders[id]
	if !ok || os.TenantID != tenantID {
		return nil, fmt.Errorf("OS %s nao encontrada", id)
	}
	return os, nil
}

func (r *MechanicRepo) GetByPlate(_ context.Context, tenantID, plate string) ([]*mechanic.ServiceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*mechanic.ServiceOrder
	for _, os := range r.orders {
		if os.TenantID == tenantID && os.VehiclePlate == plate {
			result = append(result, os)
		}
	}
	return result, nil
}

func (r *MechanicRepo) ListOpen(_ context.Context, tenantID string) ([]*mechanic.ServiceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*mechanic.ServiceOrder
	for _, os := range r.orders {
		if os.TenantID == tenantID && os.Status != mechanic.OSStatusDone && os.Status != mechanic.OSStatusInvoiced {
			result = append(result, os)
		}
	}
	return result, nil
}
