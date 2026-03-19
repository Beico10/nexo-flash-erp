// Package app — injeção de dependências completa do Nexo One.
package app

import (
	"context"
	"fmt"

	"github.com/nexoone/nexo-one/internal/ai"
	"github.com/nexoone/nexo-one/internal/auth"
	"github.com/nexoone/nexo-one/internal/baas"
	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
	"github.com/nexoone/nexo-one/internal/modules/logistics"
	"github.com/nexoone/nexo-one/internal/modules/mechanic"
	"github.com/nexoone/nexo-one/internal/repository/postgres"
	"github.com/nexoone/nexo-one/internal/tax"
	"github.com/nexoone/nexo-one/pkg/cache"
)

type Config struct {
	DatabaseURL   string
	RedisURL      string
	NatsURL       string
	JWTSecret     string
	BaseURL       string
	WhatsAppToken string
}

type Container struct {
	MechanicHandler   *handlers.MechanicHandler
	BakeryHandler     *handlers.BakeryHandler
	TaxHandler        *handlers.TaxHandler
	LogisticsHandler  *handlers.LogisticsHandler
	AestheticsHandler *handlers.AestheticsHandler
	AIHandler         *handlers.AIHandler
	PaymentHandler    *handlers.PaymentHandler
	DB                *postgres.DB
	Redis             *cache.Client
	tenantRepo        *postgres.TenantRepo
	userRepo          *postgres.UserRepo
	tokenStore        *auth.RedisTokenStore
	Close             func()
}

func Wire(cfg Config) (*Container, error) {
	db, err := postgres.New(postgres.Config{DSN: cfg.DatabaseURL})
	if err != nil {
		return nil, fmt.Errorf("app.Wire: PostgreSQL: %w", err)
	}

	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app.Wire: Redis: %w", err)
	}

	// Repositórios
	tenantRepo     := postgres.NewTenantRepo(db)
	userRepo       := postgres.NewUserRepo(db)
	mechanicRepo   := postgres.NewMechanicRepo(db)
	taxRateRepo    := postgres.NewTaxRateRepo(db, redisClient)
	aiRepo         := postgres.NewAIRepo(db)
	logisticsRepo  := postgres.NewLogisticsRepo(db)
	paymentRepo    := postgres.NewPaymentRepo(db)
	aestheticsRepo := postgres.NewAestheticsRepo(db)
	tokenStore     := auth.NewRedisTokenStore(redisClient)

	// Serviços
	taxEngine   := tax.NewEngine(taxRateRepo)
	waClient    := mechanic.NewWALinkSender(cfg.BaseURL)
	osSvc       := mechanic.NewOSService(mechanicRepo, waClient, cfg.BaseURL)
	aiGateway   := ai.NewGateway(aiRepo)
	concierge   := ai.NewConcierge(aiGateway)
	contractSvc := logistics.NewContractService(logisticsRepo)
	agendaSvc   := aesthetics.NewAgendaService(aestheticsRepo)
	baasGW      := baas.NewNoOpGateway()
	paymentSvc  := baas.NewPaymentService(baasGW, paymentRepo)

	return &Container{
		MechanicHandler:   handlers.NewMechanicHandler(osSvc),
		BakeryHandler:     handlers.NewBakeryHandler(nil),
		TaxHandler:        handlers.NewTaxHandler(taxEngine),
		LogisticsHandler:  handlers.NewLogisticsHandler(contractSvc),
		AestheticsHandler: handlers.NewAestheticsHandler(agendaSvc),
		AIHandler:         handlers.NewAIHandler(aiGateway, concierge),
		PaymentHandler:    handlers.NewPaymentHandler(paymentSvc),
		DB:                db,
		Redis:             redisClient,
		tenantRepo:        tenantRepo,
		userRepo:          userRepo,
		tokenStore:        tokenStore,
		Close:             func() { db.Close(); redisClient.Close() },
	}, nil
}

func (c *Container) UserProvider() auth.UserProvider {
	return &userProviderAdapter{tenants: c.tenantRepo, users: c.userRepo}
}

func (c *Container) TokenStore() auth.TokenStore { return c.tokenStore }

type userProviderAdapter struct {
	tenants *postgres.TenantRepo
	users   *postgres.UserRepo
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
		return nil, err
	}
	return &auth.TenantAuth{
		ID: t.ID, Slug: t.Slug, BusinessType: t.BusinessType, Plan: t.Plan,
	}, nil
}
