// Package memory — repositório in-memory de módulos avulsos.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/modules"
)

type ModulesRepo struct {
	mu            sync.RWMutex
	subscriptions map[string][]*modules.ModuleSubscription // key: tenantID
}

func NewModulesRepo() *ModulesRepo {
	r := &ModulesRepo{
		subscriptions: make(map[string][]*modules.ModuleSubscription),
	}
	r.seed()
	return r
}

func (r *ModulesRepo) seed() {
	// Demo: tenant demo já tem fiscal e whatsapp ativos
	now := time.Now()
	trialEnd := now.AddDate(0, 0, 25)

	r.subscriptions["demo"] = []*modules.ModuleSubscription{
		{
			ID: uuid.NewString(), TenantID: "demo",
			ModuleID: modules.ModuleFiscalIBSCBS,
			Status: "trialing", TrialEndsAt: &trialEnd,
			Price: 67.00, Cycle: "monthly", CreatedAt: now,
		},
	}
}

func (r *ModulesRepo) GetTenantModules(ctx context.Context, tenantID string) (*modules.TenantModules, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subs := r.subscriptions[tenantID]
	var addonModules []modules.ModuleSubscription
	var allModuleIDs []modules.ModuleID

	for _, s := range subs {
		if s.Status == "active" || s.Status == "trialing" {
			addonModules = append(addonModules, *s)
			allModuleIDs = append(allModuleIDs, s.ModuleID)
		}
	}

	return &modules.TenantModules{
		TenantID:     tenantID,
		PlanModules:  []modules.ModuleID{},
		AddonModules: addonModules,
		AllModules:   allModuleIDs,
	}, nil
}

func (r *ModulesRepo) AddModule(ctx context.Context, sub *modules.ModuleSubscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == "" {
		sub.ID = uuid.NewString()
	}

	// Verificar se já existe
	for _, existing := range r.subscriptions[sub.TenantID] {
		if existing.ModuleID == sub.ModuleID && existing.Status != "cancelled" {
			return modules.ErrAlreadySubscribed
		}
	}

	copy := *sub
	r.subscriptions[sub.TenantID] = append(r.subscriptions[sub.TenantID], &copy)
	return nil
}

func (r *ModulesRepo) CancelModule(ctx context.Context, tenantID string, moduleID modules.ModuleID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, sub := range r.subscriptions[tenantID] {
		if sub.ModuleID == moduleID && sub.Status != "cancelled" {
			sub.Status = "cancelled"
			return nil
		}
	}
	return nil
}

func (r *ModulesRepo) ListModuleSubscriptions(ctx context.Context, tenantID string) ([]*modules.ModuleSubscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subs := r.subscriptions[tenantID]
	result := make([]*modules.ModuleSubscription, len(subs))
	for i, s := range subs {
		copy := *s
		result[i] = &copy
	}
	return result, nil
}
