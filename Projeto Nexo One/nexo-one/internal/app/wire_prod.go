// Package app — wire de PRODUÇÃO do Nexo One.
// Usa PostgreSQL com RLS para isolamento multi-tenant real.
// Para desenvolvimento sem banco: use wire.go (in-memory).
//
// Ativar: renomear este arquivo para wire.go antes do deploy.
package app

import (
	"context"
	"fmt"

	"github.com/nexoone/nexo-one/internal/ai"
	"github.com/nexoone/nexo-one/internal/auth"
	"github.com/nexoone/nexo-one/internal/baas"
	"github.com/nexoone/nexo-one/internal/billing"
	"github.com/nexoone/nexo-one/internal/expenses"
	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/internal/journey"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/bakery"
	"github.com/nexoone/nexo-one/internal/modules/industry"
	"github.com/nexoone/nexo-one/internal/modules/logistics"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
	"github.com/nexoone/nexo-one/internal/modules/shoes"
	pgRepo "github.com/nexoone/nexo-one/internal/repository/postgres"
	"github.com/nexoone/nexo-one/internal/tax"
	"github.com/nexoone/nexo-one/internal/trial"
	"github.com/nexoone/nexo-one/pkg/cache"
)

// ConfigProd configuração de produção (requer banco e Redis reais).
type ConfigProd struct {
	DatabaseURL   string
	RedisURL      string
	NatsURL       string
	JWTSecret     string
	BaseURL       string
	Port          string
}

// WireProd inicializa todas as dependências de produção.
// PostgreSQL + Redis + RLS ativo.
func WireProd(cfg ConfigProd) (*Container, error) {
	// ── 1. PostgreSQL ─────────────────────────────────────────────────────────
	db, err := pgRepo.New(pgRepo.Config{DSN: cfg.DatabaseURL})
	if err != nil {
		return nil, fmt.Errorf("WireProd: PostgreSQL: %w", err)
	}

	// ── 2. Redis ──────────────────────────────────────────────────────────────
	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("WireProd: Redis: %w", err)
	}

	// ── 3. Repositórios PostgreSQL ────────────────────────────────────────────
	tenantRepo     := pgRepo.NewTenantRepo(db)
	userRepo       := pgRepo.NewUserRepo(db)
	mechanicRepo   := pgRepo.NewMechanicRepo(db)
	bakeryRepo     := pgRepo.NewBakeryRepo(db)
	taxRateRepo    := pgRepo.NewTaxRateRepo(db, redisClient)
	aiRepo         := pgRepo.NewAIRepo(db)
	logisticsRepo  := pgRepo.NewLogisticsRepo(db)
	paymentRepo    := pgRepo.NewPaymentRepo(db)
	aestheticsRepo := pgRepo.NewAestheticsRepo(db)
	billingRepo    := pgRepo.NewBillingRepo(db)
	trialRepo      := pgRepo.NewTrialRepo(db)
	journeyRepo    := pgRepo.NewJourneyRepo(db)
	tokenStore     := auth.NewRedisTokenStore(redisClient)

	// ── 4. Serviços ───────────────────────────────────────────────────────────
	taxEngine   := tax.NewEngine(taxRateRepo)
	waClient    := mechanic.NewWALinkSender(cfg.BaseURL)
	osSvc       := mechanic.NewOSService(mechanicRepo, waClient, cfg.BaseURL)
	aiGateway   := ai.NewGateway(aiRepo)
	concierge   := ai.NewConcierge(aiGateway)
	contractSvc := logistics.NewContractService(logisticsRepo)
	agendaSvc   := aesthetics.NewAgendaService(aestheticsRepo)
	baasGW      := baas.NewNoOpGateway()
	paymentSvc  := baas.NewPaymentService(baasGW, paymentRepo)
	scaleReader := pgRepo.NewNoOpScaleReader()
	bakeryService := bakery.NewPDVService(bakeryRepo, scaleReader)
	billingSvc  := billing.NewService(billingRepo)
	trialSvc    := trial.NewService(trialRepo)
	journeySvc  := journey.NewService(journeyRepo)
	sefazScraper := expenses.NewSEFAZScraper()
	expenseSvc  := expenses.NewService(
		pgRepo.NewExpenseRepo(db),
		sefazScraper,
	)
	industryRepo := pgRepo.NewIndustryRepo(db)
	shoesRepo    := pgRepo.NewShoesRepo(db)

	pcpSvc  := industry.NewPCPService(industryRepo, industryRepo, industryRepo)
	gridSvc := shoes.NewGridService(shoesRepo)

	authSvc := auth.NewService(
		cfg.JWTSecret,
		&prodUserProvider{tenants: tenantRepo, users: userRepo},
		tokenStore,
	)

	// ── 5. Handlers ───────────────────────────────────────────────────────────
	return &Container{
		MechanicHandler:   handlers.NewMechanicHandler(osSvc),
		BakeryHandler:     handlers.NewBakeryHandler(bakeryService),
		TaxHandler:        handlers.NewTaxHandler(taxEngine),
		LogisticsHandler:  handlers.NewLogisticsHandler(contractSvc),
		AestheticsHandler: handlers.NewAestheticsHandler(agendaSvc),
		AIHandler:         handlers.NewAIHandler(aiGateway, concierge),
		PaymentHandler:    handlers.NewPaymentHandler(paymentSvc),
		SimulatorHandler:  handlers.NewSimulatorHandler(taxEngine, taxRateRepo.ListRates),
		DashboardHandler:  handlers.NewDashboardHandler(
			pgRepo.NewDashboardProvider(mechanicRepo, bakeryRepo, aestheticsRepo, aiRepo),
		),
		BillingHandler:    handlers.NewBillingHandler(billingSvc),
		TrialHandler:      handlers.NewTrialHandler(trialSvc, journeySvc),
		OnboardingHandler: handlers.NewOnboardingHandler(journeySvc),
		TrackingHandler:   handlers.NewTrackingHandler(journeySvc),
		AnalyticsHandler:  handlers.NewAnalyticsHandler(journeySvc),
		ExpenseHandler:    handlers.NewExpenseHandler(expenseSvc),
		IndustryHandler:   handlers.NewIndustryHandler(pcpSvc, industryRepo),
		ShoesHandler:      handlers.NewShoesHandler(gridSvc, shoesRepo),
		NFEHandler:        handlers.NewNFEHandler(),
		AuthService:       authSvc,
		TaxEngine:         taxEngine,
		Close: func() {
			db.Close()
			redisClient.Close()
		},
	}, nil
}

// prodUserProvider adapta repos PostgreSQL para auth.UserProvider.
type prodUserProvider struct {
	tenants *pgRepo.TenantRepo
	users   *pgRepo.UserRepo
}

func (a *prodUserProvider) GetByEmail(ctx context.Context, tenantID, email string) (*auth.UserAuth, error) {
	u, err := a.users.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	return &auth.UserAuth{
		ID: u.ID, TenantID: u.TenantID, Email: u.Email,
		Name: u.Name, Role: u.Role, PasswordHash: u.PasswordHash, Active: u.Active,
	}, nil
}

func (a *prodUserProvider) GetTenantBySlug(ctx context.Context, slug string) (*auth.TenantAuth, error) {
	t, err := a.tenants.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant não encontrado")
	}
	return &auth.TenantAuth{
		ID: t.ID, Slug: t.Slug, BusinessType: t.BusinessType, Plan: t.Plan,
	}, nil
}
