// Package journey implementa o tracking da jornada do usuário.
//
// Objetivo: Saber exatamente onde o lead trava para otimizar conversão.
//
// Eventos rastreados:
//   - page_view: Visualização de página
//   - signup_started: Iniciou cadastro
//   - signup_completed: Completou cadastro
//   - verification_sent: Código enviado
//   - verification_completed: Telefone verificado
//   - onboarding_started: Iniciou onboarding
//   - onboarding_step_completed: Completou passo do onboarding
//   - onboarding_step_skipped: Pulou passo
//   - onboarding_completed: Completou onboarding
//   - first_action: Primeira ação real (OS, venda, etc)
//   - trial_converted: Converteu para pago
package journey

import (
	"context"
	"time"
)

// Event representa um evento da jornada.
type Event struct {
	ID            int64
	TenantID      string
	UserID        string
	AnonymousID   string // Para eventos antes do cadastro
	EventName     string
	EventCategory string // onboarding, activation, engagement, conversion
	PagePath      string
	PageTitle     string
	Referrer      string
	Properties    map[string]any
	FunnelStage   string // awareness, interest, decision, action, retention
	DeviceType    string
	Browser       string
	OS            string
	SessionID     string
	OccurredAt    time.Time
	TimeOnPage    int // segundos na página anterior
}

// FunnelMetrics métricas do funil de conversão.
type FunnelMetrics struct {
	Date                 time.Time
	BusinessType         string
	Visits               int
	SignupsStarted       int
	SignupsCompleted     int
	PhoneVerified        int
	OnboardingStarted    int
	OnboardingCompleted  int
	FirstAction          int
	TrialConverted       int
	DropSignup           int
	DropVerification     int
	DropOnboarding       int
	DropActivation       int
	ConversionRate       float64 // % de visits que converteram
}

// DropPoint onde o usuário travou.
type DropPoint struct {
	TenantID      string
	UserID        string
	TenantName    string
	BusinessType  string
	Email         string
	UserName      string
	Phone         string
	Stage         string
	StepCode      string
	DaysStuck     int
	ReminderCount int
}

// OnboardingProgress progresso do onboarding.
type OnboardingProgress struct {
	TenantID       string
	UserID         string
	BusinessType   string
	CurrentStep    string
	TotalSteps     int
	CompletedSteps []string
	StartedAt      time.Time
	CompletedAt    *time.Time
	LastActivity   time.Time
	Skipped        bool
}

// OnboardingStep definição de um passo.
type OnboardingStep struct {
	ID            string
	BusinessType  string
	StepCode      string
	StepOrder     int
	Title         string
	Description   string
	Icon          string
	IsRequired    bool
	IsSkippable   bool
	EstimatedTime int // segundos
	ActionType    string
	RewardText    string
	RewardDays    int
}

// JourneyRepository interface de persistência.
type JourneyRepository interface {
	// Eventos
	TrackEvent(ctx context.Context, event *Event) error
	GetEvents(ctx context.Context, tenantID string, since time.Time) ([]*Event, error)
	
	// Funil
	GetFunnelMetrics(ctx context.Context, date time.Time, businessType string) (*FunnelMetrics, error)
	GetFunnelRange(ctx context.Context, from, to time.Time, businessType string) ([]*FunnelMetrics, error)
	UpdateFunnelDaily(ctx context.Context, metrics *FunnelMetrics) error
	
	// Drop Points
	GetDropPoints(ctx context.Context, stage string, minDaysStuck int) ([]*DropPoint, error)
	MarkDropResolved(ctx context.Context, tenantID, resolution string) error
	CreateDropPoint(ctx context.Context, dp *DropPoint) error
	
	// Onboarding
	GetOnboardingSteps(ctx context.Context, businessType string) ([]*OnboardingStep, error)
	GetOnboardingProgress(ctx context.Context, tenantID string) (*OnboardingProgress, error)
	UpdateOnboardingProgress(ctx context.Context, progress *OnboardingProgress) error
}

// Service gerencia tracking de jornada.
type Service struct {
	repo JourneyRepository
}

func NewService(repo JourneyRepository) *Service {
	return &Service{repo: repo}
}

// ════════════════════════════════════════════════════════════
// TRACKING DE EVENTOS
// ════════════════════════════════════════════════════════════

// Track registra um evento da jornada.
func (s *Service) Track(ctx context.Context, event *Event) error {
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	
	// Determinar estágio do funil
	event.FunnelStage = s.determineFunnelStage(event.EventName)
	
	// Salvar evento
	if err := s.repo.TrackEvent(ctx, event); err != nil {
		return err
	}
	
	// Verificar se é ponto de travamento
	s.checkDropPoint(ctx, event)
	
	return nil
}

// TrackPageView atalho para page_view.
func (s *Service) TrackPageView(ctx context.Context, tenantID, userID, path, title string) error {
	return s.Track(ctx, &Event{
		TenantID:      tenantID,
		UserID:        userID,
		EventName:     "page_view",
		EventCategory: "engagement",
		PagePath:      path,
		PageTitle:     title,
	})
}

// TrackSignupStarted marca início do cadastro.
func (s *Service) TrackSignupStarted(ctx context.Context, anonymousID string) error {
	return s.Track(ctx, &Event{
		AnonymousID:   anonymousID,
		EventName:     "signup_started",
		EventCategory: "conversion",
		FunnelStage:   "interest",
	})
}

// TrackSignupCompleted marca cadastro completo.
func (s *Service) TrackSignupCompleted(ctx context.Context, tenantID, userID string) error {
	return s.Track(ctx, &Event{
		TenantID:      tenantID,
		UserID:        userID,
		EventName:     "signup_completed",
		EventCategory: "conversion",
		FunnelStage:   "decision",
	})
}

// TrackOnboardingStep marca passo do onboarding.
func (s *Service) TrackOnboardingStep(ctx context.Context, tenantID, userID, stepCode string, skipped bool) error {
	eventName := "onboarding_step_completed"
	if skipped {
		eventName = "onboarding_step_skipped"
	}
	
	return s.Track(ctx, &Event{
		TenantID:      tenantID,
		UserID:        userID,
		EventName:     eventName,
		EventCategory: "onboarding",
		Properties:    map[string]any{"step": stepCode, "skipped": skipped},
	})
}

// TrackFirstAction marca primeira ação real do usuário.
func (s *Service) TrackFirstAction(ctx context.Context, tenantID, userID, actionType string) error {
	return s.Track(ctx, &Event{
		TenantID:      tenantID,
		UserID:        userID,
		EventName:     "first_action",
		EventCategory: "activation",
		FunnelStage:   "action",
		Properties:    map[string]any{"action_type": actionType},
	})
}

// TrackTrialConverted marca conversão do trial.
func (s *Service) TrackTrialConverted(ctx context.Context, tenantID, userID, planCode string) error {
	return s.Track(ctx, &Event{
		TenantID:      tenantID,
		UserID:        userID,
		EventName:     "trial_converted",
		EventCategory: "conversion",
		FunnelStage:   "retention",
		Properties:    map[string]any{"plan": planCode},
	})
}

// ════════════════════════════════════════════════════════════
// ONBOARDING
// ════════════════════════════════════════════════════════════

// GetOnboardingSteps retorna os passos do onboarding para o nicho.
func (s *Service) GetOnboardingSteps(ctx context.Context, businessType string) ([]*OnboardingStep, error) {
	return s.repo.GetOnboardingSteps(ctx, businessType)
}

// GetOnboardingProgress retorna o progresso do tenant.
func (s *Service) GetOnboardingProgress(ctx context.Context, tenantID string) (*OnboardingProgress, error) {
	return s.repo.GetOnboardingProgress(ctx, tenantID)
}

// CompleteOnboardingStep marca um passo como completo.
func (s *Service) CompleteOnboardingStep(ctx context.Context, tenantID, userID, stepCode string, skipped bool) error {
	progress, err := s.repo.GetOnboardingProgress(ctx, tenantID)
	if err != nil {
		return err
	}
	
	// Adicionar aos passos completados
	progress.CompletedSteps = append(progress.CompletedSteps, stepCode)
	progress.LastActivity = time.Now()
	
	// Verificar se completou todos
	steps, _ := s.repo.GetOnboardingSteps(ctx, progress.BusinessType)
	if len(progress.CompletedSteps) >= len(steps) {
		now := time.Now()
		progress.CompletedAt = &now
	}
	
	// Track evento
	s.TrackOnboardingStep(ctx, tenantID, userID, stepCode, skipped)
	
	return s.repo.UpdateOnboardingProgress(ctx, progress)
}

// SkipOnboarding marca onboarding como pulado.
func (s *Service) SkipOnboarding(ctx context.Context, tenantID string) error {
	progress, err := s.repo.GetOnboardingProgress(ctx, tenantID)
	if err != nil {
		return err
	}
	
	progress.Skipped = true
	progress.LastActivity = time.Now()
	
	return s.repo.UpdateOnboardingProgress(ctx, progress)
}

// ════════════════════════════════════════════════════════════
// ANALYTICS
// ════════════════════════════════════════════════════════════

// GetFunnelToday retorna métricas do funil de hoje.
func (s *Service) GetFunnelToday(ctx context.Context, businessType string) (*FunnelMetrics, error) {
	return s.repo.GetFunnelMetrics(ctx, time.Now(), businessType)
}

// GetFunnelRange retorna métricas do funil em um período.
func (s *Service) GetFunnelRange(ctx context.Context, from, to time.Time, businessType string) ([]*FunnelMetrics, error) {
	return s.repo.GetFunnelRange(ctx, from, to, businessType)
}

// GetDropPoints retorna usuários travados.
func (s *Service) GetDropPoints(ctx context.Context, stage string, minDaysStuck int) ([]*DropPoint, error) {
	return s.repo.GetDropPoints(ctx, stage, minDaysStuck)
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func (s *Service) determineFunnelStage(eventName string) string {
	switch eventName {
	case "page_view":
		return "awareness"
	case "signup_started":
		return "interest"
	case "signup_completed", "verification_completed":
		return "decision"
	case "onboarding_completed", "first_action":
		return "action"
	case "trial_converted":
		return "retention"
	default:
		return "engagement"
	}
}

func (s *Service) checkDropPoint(ctx context.Context, event *Event) {
	// Implementar lógica de detecção de travamento
	// Por exemplo: se signup_started mas não signup_completed em 10 min
	// → Criar drop_point
}
