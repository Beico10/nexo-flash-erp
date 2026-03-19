// Package app — injeção de dependências completa do Nexo Flash.
package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexoflash/nexo-flash/internal/ai"
	"github.com/nexoflash/nexo-flash/internal/auth"
	"github.com/nexoflash/nexo-flash/internal/baas"
	"github.com/nexoflash/nexo-flash/internal/handlers"
	"github.com/nexoflash/nexo-flash/internal/modules/aesthetics"
	"github.com/nexoflash/nexo-flash/internal/modules/bakery"
	"github.com/nexoflash/nexo-flash/internal/modules/logistics"
	"github.com/nexoflash/nexo-flash/internal/modules/mechanic"
	"github.com/nexoflash/nexo-flash/internal/repository/postgres"
	"github.com/nexoflash/nexo-flash/internal/tax"
	"github.com/nexoflash/nexo-flash/pkg/cache"
	"github.com/nexoflash/nexo-flash/pkg/middleware"
)

// Config agrupa toda a configuração da aplicação.
type Config struct {
	DatabaseURL   string
	RedisURL      string
	NatsURL       string
	JWTSecret     string
	BaseURL       string
	WhatsAppToken string
}

// Container agrupa todos os handlers e dependências prontos.
type Container struct {
	// Handlers HTTP
	MechanicHandler   *handlers.MechanicHandler
	BakeryHandler     *handlers.BakeryHandler
	TaxHandler        *handlers.TaxHandler
	LogisticsHandler  *handlers.LogisticsHandler
	AestheticsHandler *handlers.AestheticsHandler
	AIHandler         *handlers.AIHandler
	PaymentHandler    *handlers.PaymentHandler

	// Infraestrutura
	DB    *postgres.DB
	Redis *cache.Client

	// Privados — usados pelos adapters
	tenantRepo *postgres.TenantRepo
	userRepo   *postgres.UserRepo
	tokenStore *auth.RedisTokenStore

	Close func()
}

// Wire inicializa toda a cadeia de dependências.
func Wire(cfg Config) (*Container, error) {
	// ── 1. PostgreSQL ─────────────────────────────────────────────────────────
	db, err := postgres.New(postgres.Config{DSN: cfg.DatabaseURL})
	if err != nil {
		return nil, fmt.Errorf("app.Wire: PostgreSQL: %w", err)
	}

	// ── 2. Redis ──────────────────────────────────────────────────────────────
	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app.Wire: Redis: %w", err)
	}

	// ── 3. Repositórios ───────────────────────────────────────────────────────
	tenantRepo     := postgres.NewTenantRepo(db)
	userRepo       := postgres.NewUserRepo(db)
	mechanicRepo   := postgres.NewMechanicRepo(db)
	bakeryRepo     := postgres.NewBakeryRepo(db)        // FIX 1: injetado
	taxRateRepo    := postgres.NewTaxRateRepo(db, redisClient)
	aiRepo         := postgres.NewAIRepo(db)
	logisticsRepo  := postgres.NewLogisticsRepo(db)
	paymentRepo    := postgres.NewPaymentRepo(db)
	aestheticsRepo := postgres.NewAestheticsRepo(db)
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

	// FIX 1: PDVService com repositório e balança reais
	scaleReader := postgres.NewNoOpScaleReader()
	pdvSvc      := bakery.NewPDVService(bakeryRepo, scaleReader)

	return &Container{
		MechanicHandler:   handlers.NewMechanicHandler(osSvc),
		BakeryHandler:     handlers.NewBakeryHandler(pdvSvc), // FIX 1: não é mais nil
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

// ── Adapters para interfaces externas ────────────────────────────────────────

// UserProvider adapta os repos para auth.UserProvider.
func (c *Container) UserProvider() auth.UserProvider {
	return &userProviderAdapter{tenants: c.tenantRepo, users: c.userRepo}
}

// TokenStore retorna o token store para o auth service.
func (c *Container) TokenStore() auth.TokenStore { return c.tokenStore }

// DBSessionSetter retorna o adapter que implementa middleware.DBSessionSetter.
// FIX 2: adapter correto que seta o tenant_id por conexão.
func (c *Container) DBSessionSetter() middleware.DBSessionSetter {
	return &dbSessionAdapter{db: c.DB}
}

// userProviderAdapter adapta TenantRepo+UserRepo para auth.UserProvider.
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

// FIX 2: dbSessionAdapter implementa middleware.DBSessionSetter corretamente.
// Abre uma conexão dedicada, seta o tenant_id e a fecha após o request.
type dbSessionAdapter struct{ db *postgres.DB }

func (d *dbSessionAdapter) SetTenantSession(ctx context.Context, tenantID string) error {
	conn, err := d.db.Pool().Conn(ctx)
	if err != nil {
		return fmt.Errorf("dbSessionAdapter: obter conexão: %w", err)
	}
	defer conn.Close()
	_, err = conn.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	if err != nil {
		return fmt.Errorf("dbSessionAdapter: SET LOCAL: %w", err)
	}
	return nil
}

// Pool expõe o *sql.DB para o adapter de sessão.
func (d *dbSessionAdapter) pool() *sql.DB { return d.db.Pool() }
