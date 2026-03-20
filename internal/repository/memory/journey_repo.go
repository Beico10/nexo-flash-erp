package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/journey"
)

// JourneyRepo implementa journey.JourneyRepository in-memory.
type JourneyRepo struct {
	mu              sync.RWMutex
	events          []*journey.Event
	funnelMetrics   map[string]*journey.FunnelMetrics // dateKey -> metrics
	dropPoints      []*journey.DropPoint
	onboardingSteps map[string][]*journey.OnboardingStep    // businessType -> steps
	progress        map[string]*journey.OnboardingProgress  // tenantID -> progress
}

func NewJourneyRepo() *JourneyRepo {
	r := &JourneyRepo{
		funnelMetrics:   make(map[string]*journey.FunnelMetrics),
		onboardingSteps: make(map[string][]*journey.OnboardingStep),
		progress:        make(map[string]*journey.OnboardingProgress),
	}
	r.seedOnboardingSteps()
	r.seedDemoProgress()
	return r
}

func (r *JourneyRepo) seedOnboardingSteps() {
	r.onboardingSteps["mechanic"] = []*journey.OnboardingStep{
		{ID: "ob-1", BusinessType: "mechanic", StepCode: "company_info", StepOrder: 1, Title: "Dados da empresa", Description: "Preencha CNPJ e endereco", Icon: "building", IsRequired: true, EstimatedTime: 120, ActionType: "form", RewardText: "+1 dia de trial"},
		{ID: "ob-2", BusinessType: "mechanic", StepCode: "first_os", StepOrder: 2, Title: "Crie sua primeira OS", Description: "Abra uma ordem de servico", Icon: "wrench", IsRequired: true, EstimatedTime: 60, ActionType: "action", RewardText: "+1 dia de trial"},
		{ID: "ob-3", BusinessType: "mechanic", StepCode: "invite_team", StepOrder: 3, Title: "Convide sua equipe", Description: "Adicione mecanicos e recepcionistas", Icon: "users", IsRequired: false, IsSkippable: true, EstimatedTime: 90, ActionType: "action", RewardText: "+1 dia de trial"},
		{ID: "ob-4", BusinessType: "mechanic", StepCode: "setup_whatsapp", StepOrder: 4, Title: "Configure WhatsApp", Description: "Ative notificacoes para clientes", Icon: "message-circle", IsRequired: false, IsSkippable: true, EstimatedTime: 60, ActionType: "integration", RewardText: "Recurso premium desbloqueado"},
		{ID: "ob-5", BusinessType: "mechanic", StepCode: "first_approval", StepOrder: 5, Title: "Envie aprovacao ao cliente", Description: "Envie um orcamento para aprovacao", Icon: "check-circle", IsRequired: false, IsSkippable: true, EstimatedTime: 30, ActionType: "action", RewardText: "+1 dia de trial"},
	}
	r.onboardingSteps["bakery"] = []*journey.OnboardingStep{
		{ID: "ob-b1", BusinessType: "bakery", StepCode: "company_info", StepOrder: 1, Title: "Dados da padaria", Description: "Preencha CNPJ e endereco", Icon: "building", IsRequired: true, EstimatedTime: 120, ActionType: "form"},
		{ID: "ob-b2", BusinessType: "bakery", StepCode: "add_products", StepOrder: 2, Title: "Cadastre produtos", Description: "Adicione paes, bolos e doces", Icon: "wheat", IsRequired: true, EstimatedTime: 180, ActionType: "action"},
		{ID: "ob-b3", BusinessType: "bakery", StepCode: "first_sale", StepOrder: 3, Title: "Registre uma venda", Description: "Faca sua primeira venda no PDV", Icon: "shopping-cart", IsRequired: true, EstimatedTime: 30, ActionType: "action"},
	}
	r.onboardingSteps["aesthetics"] = []*journey.OnboardingStep{
		{ID: "ob-a1", BusinessType: "aesthetics", StepCode: "company_info", StepOrder: 1, Title: "Dados do salao", Description: "Preencha CNPJ e endereco", Icon: "building", IsRequired: true, EstimatedTime: 120, ActionType: "form"},
		{ID: "ob-a2", BusinessType: "aesthetics", StepCode: "add_professionals", StepOrder: 2, Title: "Cadastre profissionais", Description: "Adicione cabeleireiras e esteticistas", Icon: "users", IsRequired: true, EstimatedTime: 90, ActionType: "action"},
		{ID: "ob-a3", BusinessType: "aesthetics", StepCode: "first_booking", StepOrder: 3, Title: "Agende um cliente", Description: "Crie o primeiro agendamento", Icon: "calendar", IsRequired: true, EstimatedTime: 30, ActionType: "action"},
	}
}

func (r *JourneyRepo) seedDemoProgress() {
	now := time.Now()
	r.progress["00000000-0000-0000-0000-000000000001"] = &journey.OnboardingProgress{
		TenantID:       "00000000-0000-0000-0000-000000000001",
		UserID:         "c1d4741e-7af6-4ba1-a1a3-9c37377c7524",
		BusinessType:   "mechanic",
		CurrentStep:    "invite_team",
		TotalSteps:     5,
		CompletedSteps: []string{"company_info", "first_os"},
		StartedAt:      now.AddDate(0, 0, -2),
		LastActivity:   now.AddDate(0, 0, -1),
	}
}

// ── Interface implementation ─────────────────────────────────

func (r *JourneyRepo) TrackEvent(_ context.Context, event *journey.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	event.ID = int64(len(r.events) + 1)
	r.events = append(r.events, event)
	return nil
}

func (r *JourneyRepo) GetEvents(_ context.Context, tenantID string, since time.Time) ([]*journey.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*journey.Event
	for _, e := range r.events {
		if e.TenantID == tenantID && e.OccurredAt.After(since) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *JourneyRepo) GetFunnelMetrics(_ context.Context, date time.Time, businessType string) (*journey.FunnelMetrics, error) {
	key := fmt.Sprintf("%s:%s", date.Format("2006-01-02"), businessType)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.funnelMetrics[key]; ok {
		return m, nil
	}
	return &journey.FunnelMetrics{Date: date, BusinessType: businessType, Visits: 142, SignupsStarted: 38, SignupsCompleted: 22, PhoneVerified: 18, OnboardingStarted: 15, OnboardingCompleted: 9, FirstAction: 7, TrialConverted: 3, ConversionRate: 2.1}, nil
}

func (r *JourneyRepo) GetFunnelRange(_ context.Context, from, to time.Time, businessType string) ([]*journey.FunnelMetrics, error) {
	var result []*journey.FunnelMetrics
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		m, _ := r.GetFunnelMetrics(context.Background(), d, businessType)
		result = append(result, m)
	}
	return result, nil
}

func (r *JourneyRepo) UpdateFunnelDaily(_ context.Context, metrics *journey.FunnelMetrics) error {
	key := fmt.Sprintf("%s:%s", metrics.Date.Format("2006-01-02"), metrics.BusinessType)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funnelMetrics[key] = metrics
	return nil
}

func (r *JourneyRepo) GetDropPoints(_ context.Context, stage string, minDaysStuck int) ([]*journey.DropPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*journey.DropPoint
	for _, dp := range r.dropPoints {
		if (stage == "" || dp.Stage == stage) && dp.DaysStuck >= minDaysStuck {
			result = append(result, dp)
		}
	}
	return result, nil
}

func (r *JourneyRepo) MarkDropResolved(_ context.Context, tenantID, resolution string) error {
	return nil
}

func (r *JourneyRepo) CreateDropPoint(_ context.Context, dp *journey.DropPoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dropPoints = append(r.dropPoints, dp)
	return nil
}

func (r *JourneyRepo) GetOnboardingSteps(_ context.Context, businessType string) ([]*journey.OnboardingStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	steps, ok := r.onboardingSteps[businessType]
	if !ok {
		return []*journey.OnboardingStep{}, nil
	}
	return steps, nil
}

func (r *JourneyRepo) GetOnboardingProgress(_ context.Context, tenantID string) (*journey.OnboardingProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.progress[tenantID]
	if !ok {
		return nil, fmt.Errorf("onboarding nao encontrado")
	}
	return p, nil
}

func (r *JourneyRepo) UpdateOnboardingProgress(_ context.Context, progress *journey.OnboardingProgress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.progress[progress.TenantID] = progress
	return nil
}
