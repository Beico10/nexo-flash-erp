// Package app - injecao de dependencias do Nexo One.
// Preview mode: usa repositorios in-memory.
// Producao: trocar para postgres.* + cache.Redis
package app

import (
	"context"
	"fmt"

	"github.com/nexoone/nexo-one/internal/ai"
	"github.com/nexoone/nexo-one/internal/auth"
	"github.com/nexoone/nexo-one/internal/baas"
	"github.com/nexoone/nexo-one/internal/billing"
	"github.com/nexoone/nexo-one/internal/expenses"
	"github.com/nexoone/nexo-one/internal/gemini"
	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/internal/journey"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/bakery"
	"github.com/nexoone/nexo-one/internal/modules/industry"
	"github.com/nexoone/nexo-one/internal/modules/logistics"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
	"github.com/nexoone/nexo-one/internal/modules/shoes"
	"github.com/nexoone/nexo-one/internal/payables"
	"github.com/nexoone/nexo-one/internal/receivables"
	"github.com/nexoone/nexo-one/internal/finance"
	"github.com/nexoone/nexo-one/internal/inventory"
	"github.com/nexoone/nexo-one/internal/repository/memory"
	"github.com/nexoone/nexo-one/internal/tax"
	"github.com/nexoone/nexo-one/internal/trial"
	"github.com/nexoone/nexo-one/internal/web"
)

type Config struct {
	JWTSecret     string
	BaseURL       string
	Port          string
}

type Container struct {
	MechanicHandler    *handlers.MechanicHandler
	BakeryHandler      *handlers.BakeryHandler
	TaxHandler         *handlers.TaxHandler
	LogisticsHandler   *handlers.LogisticsHandler
	AestheticsHandler  *handlers.AestheticsHandler
	AIHandler          *handlers.AIHandler
	PaymentHandler     *handlers.PaymentHandler
	SimulatorHandler   *handlers.SimulatorHandler
	DashboardHandler   *handlers.DashboardHandler
	BillingHandler     *handlers.BillingHandler
	TrialHandler       *handlers.TrialHandler
	OnboardingHandler  *handlers.OnboardingHandler
	TrackingHandler    *handlers.TrackingHandler
	AnalyticsHandler   *handlers.AnalyticsHandler
	IndustryHandler    *handlers.IndustryHandler
	ShoesHandler       *handlers.ShoesHandler
	NFEHandler         *handlers.NFEHandler
	ExpenseHandler     *handlers.ExpenseHandler
	CopilotHandler     *handlers.CopilotHandler
	PayablesHandler    *handlers.PayablesHandler
	ReceivablesHandler *handlers.ReceivablesHandler
	FinanceHandler     *handlers.FinanceHandler
	InventoryHandler   *handlers.InventoryHandler
	PageHandler        *web.PageHandler
	TemplateRenderer   *web.TemplateRenderer
	DashboardProvider  handlers.DashboardDataProvider
	AuthService        *auth.Service
	TaxEngine          *tax.Engine
	tenantRepo         *memory.TenantRepo
	userRepo           *memory.UserRepo
	tokenStore         *auth.RedisTokenStore
	Close              func()
}

func Wire(cfg Config) (*Container, error) {
	cache := memory.NewCache()

	// Repositorios in-memory
	tenantRepo := memory.NewTenantRepo()
	userRepo := memory.NewUserRepo()
	mechanicRepo := memory.NewMechanicRepo()
	bakeryRepo := memory.NewBakeryRepo()
	taxRateRepo := memory.NewTaxRateRepo()
	aiRepo := memory.NewAIRepo()
	logisticsRepo := memory.NewLogisticsRepo()
	paymentRepo := memory.NewPaymentRepo()
	aestheticsRepo := memory.NewAestheticsRepo()
	billingRepo := memory.NewBillingRepo()
	trialRepo := memory.NewTrialRepo()
	journeyRepo := memory.NewJourneyRepo()
	expenseRepo := memory.NewExpenseRepo()
	industryRepo := memory.NewIndustryRepo()
	shoesRepo := memory.NewShoesRepo()
	payablesRepo := memory.NewPayablesRepo()
	receivablesRepo := memory.NewReceivablesRepo()
	financeRepo := memory.NewFinanceRepo()
	inventoryRepo := memory.NewInventoryRepo()
	tokenStore := auth.NewRedisTokenStore(cache)

	// Servicos
	taxEngine := tax.NewEngine(taxRateRepo)
	waClient := mechanic.NewWALinkSender(cfg.BaseURL)
	osSvc := mechanic.NewOSService(mechanicRepo, waClient, cfg.BaseURL)
	aiGateway := ai.NewGateway(aiRepo)
	concierge := ai.NewConcierge(aiGateway)
	contractSvc := logistics.NewContractService(logisticsRepo)
	agendaSvc := aesthetics.NewAgendaService(aestheticsRepo)
	baasGW := baas.NewNoOpGateway()
	paymentSvc := baas.NewPaymentService(baasGW, paymentRepo)
	bakeryService := bakery.NewPDVService(bakeryRepo, nil)
	dashboardProvider := memory.NewDashboardProvider(mechanicRepo, bakeryRepo, aestheticsRepo, aiRepo)
	billingSvc := billing.NewService(billingRepo)
	trialSvc := trial.NewService(trialRepo)
	journeySvc := journey.NewService(journeyRepo)
	sefazScraper := expenses.NewSEFAZScraper()
	expenseSvc := expenses.NewService(expenseRepo, sefazScraper)
	pcpSvc := industry.NewPCPService(&memory.BOMAdapter{R: industryRepo}, industryRepo, industryRepo)
	gridSvc := shoes.NewGridService(shoesRepo)
	payablesSvc := payables.NewService(payablesRepo, nil)
	receivablesSvc := receivables.NewService(receivablesRepo, nil)
	financeSvc := finance.NewService(financeRepo)
	inventorySvc := inventory.NewService(inventoryRepo, nil)
	
	// Cliente Gemini (Co-Piloto IA)
	geminiClient := gemini.NewClient("")
	
	// Template Renderer
	templateRenderer, err := web.NewTemplateRenderer("/app/templates")
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar templates: %w", err)
	}

	// Seed demo data
	memory.SeedAllDemoData(mechanicRepo, bakeryRepo, aestheticsRepo, aiRepo)

	// Auth
	authSvc := auth.NewService(cfg.JWTSecret, &userProviderAdapter{tenants: tenantRepo, users: userRepo}, tokenStore)

	return &Container{
		MechanicHandler:    handlers.NewMechanicHandler(osSvc),
		BakeryHandler:      handlers.NewBakeryHandler(bakeryService),
		TaxHandler:         handlers.NewTaxHandler(taxEngine),
		LogisticsHandler:   handlers.NewLogisticsHandler(contractSvc),
		AestheticsHandler:  handlers.NewAestheticsHandler(agendaSvc),
		AIHandler:          handlers.NewAIHandler(aiGateway, concierge),
		PaymentHandler:     handlers.NewPaymentHandler(paymentSvc),
		SimulatorHandler:   handlers.NewSimulatorHandler(taxEngine, taxRateRepo.ListRates),
		DashboardHandler:   handlers.NewDashboardHandler(dashboardProvider),
		BillingHandler:     handlers.NewBillingHandler(billingSvc),
		TrialHandler:       handlers.NewTrialHandler(trialSvc, journeySvc),
		OnboardingHandler:  handlers.NewOnboardingHandler(journeySvc),
		TrackingHandler:    handlers.NewTrackingHandler(journeySvc),
		AnalyticsHandler:   handlers.NewAnalyticsHandler(journeySvc),
		ExpenseHandler:     handlers.NewExpenseHandler(expenseSvc),
		IndustryHandler:    handlers.NewIndustryHandler(pcpSvc, industryRepo),
		ShoesHandler:       handlers.NewShoesHandler(gridSvc, shoesRepo),
		NFEHandler:         handlers.NewNFEHandler(),
		CopilotHandler:     handlers.NewCopilotHandler(geminiClient),
		PayablesHandler:    handlers.NewPayablesHandler(payablesSvc),
		ReceivablesHandler: handlers.NewReceivablesHandler(receivablesSvc),
		FinanceHandler:     handlers.NewFinanceHandler(financeSvc),
		InventoryHandler:   handlers.NewInventoryHandler(inventorySvc),
		PageHandler:        web.NewPageHandler(templateRenderer, dashboardProvider),
		TemplateRenderer:   templateRenderer,
		DashboardProvider:  dashboardProvider,
		AuthService:        authSvc,
		TaxEngine:          taxEngine,
		tenantRepo:         tenantRepo,
		userRepo:           userRepo,
		tokenStore:         tokenStore,
		Close:              func() { cache.Close() },
	}, nil
}

func (c *Container) UserProvider() auth.UserProvider {
	return &userProviderAdapter{tenants: c.tenantRepo, users: c.userRepo}
}

func (c *Container) TokenStore() auth.TokenStore { return c.tokenStore }

type userProviderAdapter struct {
	tenants *memory.TenantRepo
	users   *memory.UserRepo
}

func (a *userProviderAdapter) GetByEmail(ctx context.Context, tenantID, email string) (*auth.UserAuth, error) {
	u, err := a.users.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	return &auth.UserAuth{
		ID: u.ID, TenantID: u.TenantID, Email: u.Email,
		Name: u.Name, Role: u.Role, PasswordHash: u.PasswordHash, Active: u.Active,
	}, nil
}

func (a *userProviderAdapter) GetTenantBySlug(ctx context.Context, slug string) (*auth.TenantAuth, error) {
	t, err := a.tenants.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant nao encontrado")
	}
	return &auth.TenantAuth{
		ID: t.ID, Slug: t.Slug, BusinessType: t.BusinessType, Plan: t.Plan,
	}, nil
}
