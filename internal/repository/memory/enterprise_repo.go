// Package memory — repositório in-memory para Enterprise.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/enterprise"
)

type EnterpriseRepo struct {
	mu       sync.RWMutex
	plans    map[string]*enterprise.Plan
	licenses map[string]*enterprise.License
	usage    map[string]*enterprise.Usage // key: tenantID-month
}

func NewEnterpriseRepo() *EnterpriseRepo {
	repo := &EnterpriseRepo{
		plans:    make(map[string]*enterprise.Plan),
		licenses: make(map[string]*enterprise.License),
		usage:    make(map[string]*enterprise.Usage),
	}
	repo.seedPlans()
	return repo
}

func (r *EnterpriseRepo) seedPlans() {
	defaultPlans := []*enterprise.Plan{
		{
			ID:         "starter",
			Name:       "Starter",
			Type:       "starter",
			MaxUsers:   3,
			MaxTenants: 1,
			Features:   []string{"dashboard", "nfe_basic", "reports_basic"},
			PriceMonth: 97,
			PriceYear:  970,
			IsActive:   true,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "professional",
			Name:       "Professional",
			Type:       "professional",
			MaxUsers:   10,
			MaxTenants: 3,
			Features:   []string{"dashboard", "nfe_full", "reports_advanced", "api_access", "integrations"},
			PriceMonth: 197,
			PriceYear:  1970,
			IsActive:   true,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "enterprise",
			Name:       "Enterprise",
			Type:       "enterprise",
			MaxUsers:   100,
			MaxTenants: 50,
			Features:   []string{"dashboard", "nfe_full", "reports_advanced", "api_access", "integrations", "white_label", "sla_priority", "dedicated_support"},
			PriceMonth: 497,
			PriceYear:  4970,
			IsActive:   true,
			CreatedAt:  time.Now(),
		},
	}
	
	for _, p := range defaultPlans {
		r.plans[p.ID] = p
	}
}

func (r *EnterpriseRepo) ListPlans(ctx context.Context) ([]*enterprise.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plans := make([]*enterprise.Plan, 0, len(r.plans))
	for _, p := range r.plans {
		if p.IsActive {
			plans = append(plans, p)
		}
	}
	return plans, nil
}

func (r *EnterpriseRepo) GetPlan(ctx context.Context, id string) (*enterprise.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plan, ok := r.plans[id]
	if !ok {
		return nil, fmt.Errorf("plano não encontrado: %s", id)
	}
	return plan, nil
}

func (r *EnterpriseRepo) CreatePlan(ctx context.Context, plan *enterprise.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.plans[plan.ID] = plan
	return nil
}

func (r *EnterpriseRepo) UpdatePlan(ctx context.Context, plan *enterprise.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.plans[plan.ID]; !ok {
		return fmt.Errorf("plano não encontrado: %s", plan.ID)
	}
	r.plans[plan.ID] = plan
	return nil
}

func (r *EnterpriseRepo) GetLicense(ctx context.Context, tenantID string) (*enterprise.License, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	license, ok := r.licenses[tenantID]
	if !ok {
		return nil, fmt.Errorf("licença não encontrada para tenant: %s", tenantID)
	}
	return license, nil
}

func (r *EnterpriseRepo) CreateLicense(ctx context.Context, license *enterprise.License) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.licenses[license.TenantID] = license
	return nil
}

func (r *EnterpriseRepo) UpdateLicense(ctx context.Context, license *enterprise.License) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.licenses[license.TenantID] = license
	return nil
}

func (r *EnterpriseRepo) GetUsage(ctx context.Context, tenantID, month string) (*enterprise.Usage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	key := tenantID + "-" + month
	usage, ok := r.usage[key]
	if !ok {
		return nil, fmt.Errorf("uso não encontrado")
	}
	return usage, nil
}

func (r *EnterpriseRepo) RecordUsage(ctx context.Context, usage *enterprise.Usage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := usage.TenantID + "-" + usage.Month
	r.usage[key] = usage
	return nil
}
