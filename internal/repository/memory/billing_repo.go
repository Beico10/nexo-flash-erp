package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/billing"
)

// BillingRepo implementa billing.BillingRepository in-memory.
type BillingRepo struct {
	mu            sync.RWMutex
	plans         []*billing.Plan
	subscriptions map[string]*billing.Subscription // keyed by tenantID
	usage         map[string]map[string]int        // tenantID -> metric -> count
	coupons       map[string]*billing.Coupon        // keyed by code
	usedCoupons   map[string]bool                   // code:tenantID -> used
}

func NewBillingRepo() *BillingRepo {
	r := &BillingRepo{
		subscriptions: make(map[string]*billing.Subscription),
		usage:         make(map[string]map[string]int),
		coupons:       make(map[string]*billing.Coupon),
		usedCoupons:   make(map[string]bool),
	}
	r.seedPlans()
	r.seedCoupons()
	r.seedDemoSubscription()
	return r
}

func (r *BillingRepo) seedPlans() {
	intPtr := func(v int) *int { return &v }
	r.plans = []*billing.Plan{
		{
			ID: "plan-starter", Code: "starter", Name: "Starter", Description: "Para quem esta comecando",
			PriceMonthly: 497, PriceYearly: 4970, SetupFee: 0,
			MaxUsers: intPtr(3), MaxTransactions: intPtr(500), MaxProducts: intPtr(200), MaxInvoices: intPtr(100), MaxStorageMB: intPtr(2000),
			Features: billing.PlanFeatures{Fiscal2026: true, BaaSPix: true, WhatsApp: true, AIConcierge: true},
			AllowedNiches: []string{"mechanic", "bakery", "aesthetics", "logistics", "shoes"}, DisplayOrder: 1, IsActive: true,
		},
		{
			ID: "plan-pro", Code: "pro", Name: "Pro", Description: "Para negocios em expansao",
			PriceMonthly: 997, PriceYearly: 9970, SetupFee: 0,
			MaxUsers: intPtr(10), MaxTransactions: intPtr(5000), MaxProducts: intPtr(2000), MaxInvoices: intPtr(1000), MaxStorageMB: intPtr(20000),
			Features: billing.PlanFeatures{Fiscal2026: true, BaaSPix: true, BaaSBoleto: true, WhatsApp: true, AICopilot: true, AIConcierge: true, MultiPDV: true, CustomReports: true},
			AllowedNiches: []string{"mechanic", "bakery", "aesthetics", "logistics", "shoes", "industry"}, DisplayOrder: 2, IsActive: true, IsFeatured: true,
		},
		{
			ID: "plan-business", Code: "business", Name: "Business", Description: "Gestao completa para media empresa",
			PriceMonthly: 1997, PriceYearly: 19970, SetupFee: 5000,
			MaxUsers: intPtr(30), MaxTransactions: nil, MaxProducts: nil, MaxInvoices: nil, MaxStorageMB: intPtr(100000),
			Features: billing.PlanFeatures{Fiscal2026: true, BaaSPix: true, BaaSBoleto: true, BaaSSplit: true, WhatsApp: true, AICopilot: true, AIConcierge: true, Roteirizador: true, MultiPDV: true, APIAccess: true, PrioritySupport: true, CustomReports: true},
			AllowedNiches: []string{"mechanic", "bakery", "aesthetics", "logistics", "shoes", "industry"}, DisplayOrder: 3, IsActive: true,
		},
		{
			ID: "plan-enterprise", Code: "enterprise", Name: "Enterprise", Description: "Para industria e multi-filial",
			PriceMonthly: 2997, PriceYearly: 29970, SetupFee: 10000,
			MaxUsers: nil, MaxTransactions: nil, MaxProducts: nil, MaxInvoices: nil, MaxStorageMB: nil,
			Features: billing.PlanFeatures{Fiscal2026: true, BaaSPix: true, BaaSBoleto: true, BaaSSplit: true, WhatsApp: true, AICopilot: true, AIConcierge: true, Roteirizador: true, MultiPDV: true, APIAccess: true, PrioritySupport: true, CustomReports: true, DedicatedSupport: true, SLA999: true},
			AllowedNiches: []string{"mechanic", "bakery", "aesthetics", "logistics", "shoes", "industry"}, DisplayOrder: 4, IsActive: true,
		},
	}
}

func (r *BillingRepo) seedCoupons() {
	validUntil := time.Now().AddDate(0, 6, 0)
	r.coupons["NEXO20"] = &billing.Coupon{
		ID: "coupon-1", Code: "NEXO20", Description: "20% de desconto",
		DiscountType: "percent", DiscountValue: 20, DurationMonths: 3,
		ValidUntil: &validUntil, IsValid: true,
	}
	r.coupons["PRIMEIRO"] = &billing.Coupon{
		ID: "coupon-2", Code: "PRIMEIRO", Description: "Primeiro mes gratis",
		DiscountType: "percent", DiscountValue: 100, DurationMonths: 1,
		ValidUntil: &validUntil, IsValid: true,
	}
}

func (r *BillingRepo) seedDemoSubscription() {
	trialEnd := time.Now().AddDate(0, 0, 5)
	r.subscriptions["00000000-0000-0000-0000-000000000001"] = &billing.Subscription{
		ID: "sub-demo", TenantID: "00000000-0000-0000-0000-000000000001",
		PlanID: "plan-pro", PlanCode: "pro", PlanName: "Pro",
		Status: billing.StatusTrialing, TrialEndsAt: &trialEnd,
		CurrentPeriodStart: time.Now(), CurrentPeriodEnd: trialEnd,
		BillingCycle: billing.CycleMonthly, Price: 997,
		CurrentUsers: 3, CurrentTransactions: 47, CurrentProducts: 12, CurrentInvoices: 8,
		MaxUsers: intPtr(10), MaxTransactions: intPtr(5000), MaxProducts: intPtr(2000), MaxInvoices: intPtr(1000),
	}
	r.usage["00000000-0000-0000-0000-000000000001"] = map[string]int{
		"users": 3, "transactions": 47, "products": 12, "invoices": 8,
	}
}

func intPtr(v int) *int { return &v }

// ── Interface implementation ─────────────────────────────────

func (r *BillingRepo) ListPlans(_ context.Context, activeOnly bool) ([]*billing.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*billing.Plan
	for _, p := range r.plans {
		if activeOnly && !p.IsActive {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *BillingRepo) GetPlan(_ context.Context, planID string) (*billing.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.plans {
		if p.ID == planID {
			return p, nil
		}
	}
	return nil, billing.ErrPlanNotFound
}

func (r *BillingRepo) GetPlanByCode(_ context.Context, code string) (*billing.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.plans {
		if p.Code == code {
			return p, nil
		}
	}
	return nil, billing.ErrPlanNotFound
}

func (r *BillingRepo) UpdatePlan(_ context.Context, plan *billing.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, p := range r.plans {
		if p.ID == plan.ID {
			r.plans[i] = plan
			return nil
		}
	}
	return billing.ErrPlanNotFound
}

func (r *BillingRepo) GetSubscription(_ context.Context, tenantID string) (*billing.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.subscriptions[tenantID]
	if !ok {
		return nil, fmt.Errorf("assinatura nao encontrada")
	}
	return sub, nil
}

func (r *BillingRepo) CreateSubscription(_ context.Context, sub *billing.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if sub.ID == "" {
		sub.ID = fmt.Sprintf("sub-%d", time.Now().UnixNano())
	}
	r.subscriptions[sub.TenantID] = sub
	return nil
}

func (r *BillingRepo) UpdateSubscription(_ context.Context, sub *billing.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subscriptions[sub.TenantID] = sub
	return nil
}

func (r *BillingRepo) IncrementUsage(_ context.Context, tenantID, metric string, delta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.usage[tenantID] == nil {
		r.usage[tenantID] = make(map[string]int)
	}
	r.usage[tenantID][metric] += delta
	return nil
}

func (r *BillingRepo) GetUsage(_ context.Context, tenantID string) (map[string]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.usage[tenantID]; ok {
		return u, nil
	}
	return map[string]int{}, nil
}

func (r *BillingRepo) CheckLimit(_ context.Context, tenantID, metric string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.subscriptions[tenantID]
	if !ok {
		return false, fmt.Errorf("assinatura nao encontrada")
	}
	current := 0
	if u, ok := r.usage[tenantID]; ok {
		current = u[metric]
	}
	var limit *int
	switch metric {
	case "users":
		limit = sub.MaxUsers
	case "transactions":
		limit = sub.MaxTransactions
	case "products":
		limit = sub.MaxProducts
	case "invoices":
		limit = sub.MaxInvoices
	}
	if limit == nil {
		return true, nil
	}
	return current < *limit, nil
}

func (r *BillingRepo) GetCoupon(_ context.Context, code string) (*billing.Coupon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.coupons[code]
	if !ok {
		return nil, billing.ErrCouponInvalid
	}
	return c, nil
}

func (r *BillingRepo) UseCoupon(_ context.Context, code, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := code + ":" + tenantID
	if r.usedCoupons[key] {
		return billing.ErrCouponUsed
	}
	r.usedCoupons[key] = true
	return nil
}

func (r *BillingRepo) GetFeatures(_ context.Context, tenantID string) (*billing.PlanFeatures, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.subscriptions[tenantID]
	if !ok {
		return &billing.PlanFeatures{}, nil
	}
	for _, p := range r.plans {
		if p.ID == sub.PlanID {
			return &p.Features, nil
		}
	}
	return &billing.PlanFeatures{}, nil
}
