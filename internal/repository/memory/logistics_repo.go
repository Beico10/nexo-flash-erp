package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/nexoone/nexo-one/internal/modules/logistics"
)

type LogisticsRepo struct {
	mu        sync.RWMutex
	contracts map[string]*logistics.Contract
}

func NewLogisticsRepo() *LogisticsRepo {
	return &LogisticsRepo{contracts: make(map[string]*logistics.Contract)}
}

func (r *LogisticsRepo) GetApplicable(_ context.Context, tenantID, shipperID string, vt logistics.VehicleType) (*logistics.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var best *logistics.Contract
	for _, c := range r.contracts {
		if c.TenantID != tenantID || c.VehicleType != vt {
			continue
		}
		if c.ShipperID != nil && *c.ShipperID == shipperID {
			return c, nil
		}
		if c.ShipperID == nil && best == nil {
			best = c
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, fmt.Errorf("nenhum contrato encontrado para veiculo %s", vt)
}
