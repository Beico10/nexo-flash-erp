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
	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/bakery"
	"github.com/nexoone/nexo-one/internal/modules/logistics"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
	"github.com/nexoone/nexo-one/internal/repository/memory"
	"github.com/nexoone/nexo-one/internal/tax"
)

type Config struct {
	JWTSecret     string
	BaseURL       string
	Port          string
}

type Container struct {
	MechanicHandler   *handlers.MechanicHandler
	BakeryHandler     *handlers.BakeryHandler
	TaxHandler        *handlers.TaxHandler
	LogisticsHandler  *handlers.LogisticsHandler
	AestheticsHandler *handlers.AestheticsHandler
	AIHandler         *handlers.AIHandler
	PaymentHandler    *handlers.PaymentHandler
	SimulatorHandler  *handlers.SimulatorHandler
	DashboardHandler  *handlers.DashboardHandler
	AuthService       *auth.Service
	TaxEngine         *tax.Engine
	tenantRepo        *memory.TenantRepo
	userRepo          *memory.UserRepo
	tokenStore        *auth.RedisTokenStore
	Close             func()
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

	// Seed demo data
	memory.SeedAllDemoData(mechanicRepo, bakeryRepo, aestheticsRepo, aiRepo)

	// Auth
	authSvc := auth.NewService(cfg.JWTSecret, &userProviderAdapter{tenants: tenantRepo, users: userRepo}, tokenStore)

	return &Container{
		MechanicHandler:   handlers.NewMechanicHandler(osSvc),
		BakeryHandler:     handlers.NewBakeryHandler(bakeryService),
		TaxHandler:        handlers.NewTaxHandler(taxEngine),
		LogisticsHandler:  handlers.NewLogisticsHandler(contractSvc),
		AestheticsHandler: handlers.NewAestheticsHandler(agendaSvc),
		AIHandler:         handlers.NewAIHandler(aiGateway, concierge),
		PaymentHandler:    handlers.NewPaymentHandler(paymentSvc),
		SimulatorHandler:  handlers.NewSimulatorHandler(taxEngine, taxRateRepo.ListRates),
		DashboardHandler:  handlers.NewDashboardHandler(dashboardProvider),
		AuthService:       authSvc,
		TaxEngine:         taxEngine,
		tenantRepo:        tenantRepo,
		userRepo:          userRepo,
		tokenStore:        tokenStore,
		Close:             func() { cache.Close() },
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
