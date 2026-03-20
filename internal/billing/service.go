// Package billing implementa o sistema de planos e assinaturas do Nexo One.
//
// Modelo Self-Service:
//   - Trial 7 dias automático
//   - Conversão sem intervenção humana
//   - Upgrade/Downgrade pelo próprio cliente
//   - Limites verificados em tempo real
package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ════════════════════════════════════════════════════════════
// TIPOS E CONSTANTES
// ════════════════════════════════════════════════════════════

const (
	TrialDays = 7

	StatusTrialing = "trialing"
	StatusActive   = "active"
	StatusPastDue  = "past_due"
	StatusCancelled = "cancelled"
	StatusExpired  = "expired"

	CycleMonthly = "monthly"
	CycleYearly  = "yearly"
)

// Plan representa um plano configurável pelo Admin Master.
type Plan struct {
	ID              string          `json:"id"`
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	PriceMonthly    float64         `json:"price_monthly"`
	PriceYearly     float64         `json:"price_yearly"`
	SetupFee        float64         `json:"setup_fee"`
	MaxUsers        *int            `json:"max_users"`        // nil = ilimitado
	MaxTransactions *int            `json:"max_transactions"`
	MaxProducts     *int            `json:"max_products"`
	MaxInvoices     *int            `json:"max_invoices"`
	MaxStorageMB    *int            `json:"max_storage_mb"`
	Features        PlanFeatures    `json:"features"`
	AllowedNiches   []string        `json:"allowed_niches"`
	DisplayOrder    int             `json:"display_order"`
	IsActive        bool            `json:"is_active"`
	IsFeatured      bool            `json:"is_featured"`
}

// PlanFeatures define os recursos habilitados em cada plano.
type PlanFeatures struct {
	Fiscal2026       bool `json:"fiscal_2026"`
	BaaSPix          bool `json:"baas_pix"`
	BaaSBoleto       bool `json:"baas_boleto"`
	BaaSSplit        bool `json:"baas_split"`
	WhatsApp         bool `json:"whatsapp"`
	AICopilot        bool `json:"ai_copilot"`
	AIConcierge      bool `json:"ai_concierge"`
	Roteirizador     bool `json:"roteirizador"`
	MultiPDV         bool `json:"multi_pdv"`
	APIAccess        bool `json:"api_access"`
	PrioritySupport  bool `json:"priority_support"`
	CustomReports    bool `json:"custom_reports"`
	DedicatedSupport bool `json:"dedicated_support,omitempty"`
	SLA999           bool `json:"sla_99_9,omitempty"`
}

// Subscription representa a assinatura de um tenant.
type Subscription struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	PlanID              string     `json:"plan_id"`
	PlanCode            string     `json:"plan_code"`
	PlanName            string     `json:"plan_name"`
	Status              string     `json:"status"`
	TrialEndsAt         *time.Time `json:"trial_ends_at"`
	CurrentPeriodStart  time.Time  `json:"current_period_start"`
	CurrentPeriodEnd    time.Time  `json:"current_period_end"`
	BillingCycle        string     `json:"billing_cycle"`
	DiscountPercent     float64    `json:"discount_percent"`
	DiscountReason      string     `json:"discount_reason"`
	
	// Uso atual
	CurrentUsers        int `json:"current_users"`
	CurrentTransactions int `json:"current_transactions"`
	CurrentProducts     int `json:"current_products"`
	CurrentInvoices     int `json:"current_invoices"`
	
	// Limites do plano (para exibição)
	MaxUsers        *int `json:"max_users"`
	MaxTransactions *int `json:"max_transactions"`
	MaxProducts     *int `json:"max_products"`
	MaxInvoices     *int `json:"max_invoices"`
	
	// Preço atual
	Price float64 `json:"price"`
}

// UsageStatus retorna o status de uso de um recurso.
type UsageStatus struct {
	Metric   string `json:"metric"`
	Current  int    `json:"current"`
	Limit    *int   `json:"limit"`    // nil = ilimitado
	Percent  int    `json:"percent"`  // 0-100
	IsAtLimit bool  `json:"is_at_limit"`
}

// Coupon representa um cupom de desconto.
type Coupon struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	Description    string     `json:"description"`
	DiscountType   string     `json:"discount_type"`  // "percent" ou "fixed"
	DiscountValue  float64    `json:"discount_value"`
	DurationMonths int        `json:"duration_months"`
	ValidUntil     *time.Time `json:"valid_until"`
	IsValid        bool       `json:"is_valid"`
}

// ════════════════════════════════════════════════════════════
// ERROS
// ════════════════════════════════════════════════════════════

var (
	ErrPlanNotFound       = errors.New("plano não encontrado")
	ErrSubscriptionExists = errors.New("tenant já possui assinatura")
	ErrTrialExpired       = errors.New("período de teste expirado")
	ErrLimitReached       = errors.New("limite do plano atingido")
	ErrFeatureDisabled    = errors.New("recurso não disponível no seu plano")
	ErrCouponInvalid      = errors.New("cupom inválido ou expirado")
	ErrCouponUsed         = errors.New("cupom já foi utilizado")
	ErrDowngradeLimit     = errors.New("uso atual excede limites do plano inferior")
)

// ════════════════════════════════════════════════════════════
// REPOSITÓRIO (Interface)
// ════════════════════════════════════════════════════════════

type BillingRepository interface {
	// Planos
	ListPlans(ctx context.Context, activeOnly bool) ([]*Plan, error)
	GetPlan(ctx context.Context, planID string) (*Plan, error)
	GetPlanByCode(ctx context.Context, code string) (*Plan, error)
	UpdatePlan(ctx context.Context, plan *Plan) error
	
	// Assinaturas
	GetSubscription(ctx context.Context, tenantID string) (*Subscription, error)
	CreateSubscription(ctx context.Context, sub *Subscription) error
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	
	// Uso
	IncrementUsage(ctx context.Context, tenantID, metric string, delta int) error
	GetUsage(ctx context.Context, tenantID string) (map[string]int, error)
	CheckLimit(ctx context.Context, tenantID, metric string) (bool, error)
	
	// Cupons
	GetCoupon(ctx context.Context, code string) (*Coupon, error)
	UseCoupon(ctx context.Context, code, tenantID string) error
	
	// Features
	GetFeatures(ctx context.Context, tenantID string) (*PlanFeatures, error)
}

// ════════════════════════════════════════════════════════════
// SERVIÇO
// ════════════════════════════════════════════════════════════

type Service struct {
	repo BillingRepository
}

func NewService(repo BillingRepository) *Service {
	return &Service{repo: repo}
}

// ────────────────────────────────────────────────────────────
// PLANOS (Leitura para clientes, Escrita para Admin)
// ────────────────────────────────────────────────────────────

// ListPlans retorna todos os planos ativos (para página de preços).
func (s *Service) ListPlans(ctx context.Context) ([]*Plan, error) {
	return s.repo.ListPlans(ctx, true)
}

// ListAllPlans retorna todos os planos (Admin Master).
func (s *Service) ListAllPlans(ctx context.Context) ([]*Plan, error) {
	return s.repo.ListPlans(ctx, false)
}

// GetSubscription retorna a assinatura de um tenant.
func (s *Service) GetSubscription(ctx context.Context, tenantID string) (*Subscription, error) {
	return s.repo.GetSubscription(ctx, tenantID)
}

// UpdatePlan atualiza um plano (Admin Master).
func (s *Service) UpdatePlan(ctx context.Context, plan *Plan) error {
	return s.repo.UpdatePlan(ctx, plan)
}

// GetPlanByCode busca plano pelo codigo.
func (s *Service) GetPlanByCode(ctx context.Context, code string) (*Plan, error) {
	return s.repo.GetPlanByCode(ctx, code)
}

// ────────────────────────────────────────────────────────────
// TRIAL (7 dias grátis)
// ────────────────────────────────────────────────────────────

// StartTrial inicia o período de teste de 7 dias.
// Chamado automaticamente no registro do tenant.
func (s *Service) StartTrial(ctx context.Context, tenantID, planCode string) (*Subscription, error) {
	// Verificar se já tem assinatura
	existing, _ := s.repo.GetSubscription(ctx, tenantID)
	if existing != nil {
		return nil, ErrSubscriptionExists
	}
	
	// Buscar plano (default: starter se não especificado)
	if planCode == "" {
		planCode = "starter"
	}
	plan, err := s.repo.GetPlanByCode(ctx, planCode)
	if err != nil {
		return nil, ErrPlanNotFound
	}
	
	now := time.Now().UTC()
	trialEnd := now.AddDate(0, 0, TrialDays)
	
	sub := &Subscription{
		TenantID:           tenantID,
		PlanID:             plan.ID,
		PlanCode:           plan.Code,
		PlanName:           plan.Name,
		Status:             StatusTrialing,
		TrialEndsAt:        &trialEnd,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   trialEnd,
		BillingCycle:       CycleMonthly,
		Price:              plan.PriceMonthly,
	}
	
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("billing.StartTrial: %w", err)
	}
	
	return sub, nil
}

// ────────────────────────────────────────────────────────────
// CONVERSÃO E UPGRADE (Self-Service)
// ────────────────────────────────────────────────────────────

// ConvertTrial converte trial em assinatura paga.
func (s *Service) ConvertTrial(ctx context.Context, tenantID string, paymentMethod string, couponCode string) (*Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	
	if sub.Status != StatusTrialing {
		return nil, fmt.Errorf("assinatura não está em trial")
	}
	
	// Aplicar cupom se fornecido
	if couponCode != "" {
		coupon, err := s.ValidateCoupon(ctx, couponCode, sub.PlanCode)
		if err != nil {
			return nil, err
		}
		sub.DiscountPercent = coupon.DiscountValue
		sub.DiscountReason = coupon.Description
		
		if err := s.repo.UseCoupon(ctx, couponCode, tenantID); err != nil {
			return nil, err
		}
	}
	
	now := time.Now().UTC()
	sub.Status = StatusActive
	sub.TrialEndsAt = nil
	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = now.AddDate(0, 1, 0) // +1 mês
	
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	
	return sub, nil
}

// ChangePlan altera o plano do tenant (upgrade ou downgrade).
func (s *Service) ChangePlan(ctx context.Context, tenantID, newPlanCode string) (*Subscription, error) {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	
	newPlan, err := s.repo.GetPlanByCode(ctx, newPlanCode)
	if err != nil {
		return nil, ErrPlanNotFound
	}
	
	// Verificar se é downgrade e se uso atual permite
	if err := s.validateDowngrade(ctx, tenantID, sub, newPlan); err != nil {
		return nil, err
	}
	
	sub.PlanID = newPlan.ID
	sub.PlanCode = newPlan.Code
	sub.PlanName = newPlan.Name
	
	if sub.BillingCycle == CycleYearly {
		sub.Price = newPlan.PriceYearly
	} else {
		sub.Price = newPlan.PriceMonthly
	}
	
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	
	return sub, nil
}

// validateDowngrade verifica se downgrade é possível.
func (s *Service) validateDowngrade(ctx context.Context, tenantID string, sub *Subscription, newPlan *Plan) error {
	usage, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	// Verificar cada limite
	if newPlan.MaxUsers != nil && usage["users"] > *newPlan.MaxUsers {
		return fmt.Errorf("%w: você tem %d usuários, plano permite %d", 
			ErrDowngradeLimit, usage["users"], *newPlan.MaxUsers)
	}
	if newPlan.MaxProducts != nil && usage["products"] > *newPlan.MaxProducts {
		return fmt.Errorf("%w: você tem %d produtos, plano permite %d",
			ErrDowngradeLimit, usage["products"], *newPlan.MaxProducts)
	}
	
	return nil
}

// ────────────────────────────────────────────────────────────
// VERIFICAÇÃO DE LIMITES (Chamado antes de cada ação)
// ────────────────────────────────────────────────────────────

// CanPerform verifica se tenant pode executar uma ação.
func (s *Service) CanPerform(ctx context.Context, tenantID, metric string) error {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		return err
	}
	
	// Verificar status
	switch sub.Status {
	case StatusCancelled, StatusExpired:
		return fmt.Errorf("assinatura inativa")
	case StatusTrialing:
		if sub.TrialEndsAt != nil && time.Now().After(*sub.TrialEndsAt) {
			return ErrTrialExpired
		}
	case StatusPastDue:
		// Permitir por grace period (3 dias)
		gracePeriod := sub.CurrentPeriodEnd.AddDate(0, 0, 3)
		if time.Now().After(gracePeriod) {
			return fmt.Errorf("pagamento pendente há mais de 3 dias")
		}
	}
	
	// Verificar limite
	canProceed, err := s.repo.CheckLimit(ctx, tenantID, metric)
	if err != nil {
		return err
	}
	if !canProceed {
		return ErrLimitReached
	}
	
	return nil
}

// HasFeature verifica se tenant tem acesso a um recurso.
func (s *Service) HasFeature(ctx context.Context, tenantID, feature string) (bool, error) {
	features, err := s.repo.GetFeatures(ctx, tenantID)
	if err != nil {
		return false, err
	}
	
	switch feature {
	case "fiscal_2026":
		return features.Fiscal2026, nil
	case "baas_pix":
		return features.BaaSPix, nil
	case "baas_boleto":
		return features.BaaSBoleto, nil
	case "baas_split":
		return features.BaaSSplit, nil
	case "whatsapp":
		return features.WhatsApp, nil
	case "ai_copilot":
		return features.AICopilot, nil
	case "ai_concierge":
		return features.AIConcierge, nil
	case "roteirizador":
		return features.Roteirizador, nil
	case "multi_pdv":
		return features.MultiPDV, nil
	case "api_access":
		return features.APIAccess, nil
	case "priority_support":
		return features.PrioritySupport, nil
	default:
		return false, nil
	}
}

// GetUsageStatus retorna status de uso de todos os recursos.
func (s *Service) GetUsageStatus(ctx context.Context, tenantID string) ([]*UsageStatus, error) {
	sub, err := s.repo.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	
	metrics := []struct {
		name    string
		current int
		limit   *int
	}{
		{"users", sub.CurrentUsers, sub.MaxUsers},
		{"transactions", sub.CurrentTransactions, sub.MaxTransactions},
		{"products", sub.CurrentProducts, sub.MaxProducts},
		{"invoices", sub.CurrentInvoices, sub.MaxInvoices},
	}
	
	var result []*UsageStatus
	for _, m := range metrics {
		status := &UsageStatus{
			Metric:  m.name,
			Current: m.current,
			Limit:   m.limit,
		}
		
		if m.limit != nil && *m.limit > 0 {
			status.Percent = (m.current * 100) / *m.limit
			status.IsAtLimit = m.current >= *m.limit
		}
		
		result = append(result, status)
	}
	
	return result, nil
}

// ────────────────────────────────────────────────────────────
// CUPONS
// ────────────────────────────────────────────────────────────

// ValidateCoupon verifica se cupom é válido.
func (s *Service) ValidateCoupon(ctx context.Context, code, planCode string) (*Coupon, error) {
	coupon, err := s.repo.GetCoupon(ctx, code)
	if err != nil {
		return nil, ErrCouponInvalid
	}
	
	if !coupon.IsValid {
		return nil, ErrCouponInvalid
	}
	
	if coupon.ValidUntil != nil && time.Now().After(*coupon.ValidUntil) {
		return nil, ErrCouponInvalid
	}
	
	return coupon, nil
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

// ParseFeatures converte JSONB para struct.
func ParseFeatures(data []byte) (*PlanFeatures, error) {
	var f PlanFeatures
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
