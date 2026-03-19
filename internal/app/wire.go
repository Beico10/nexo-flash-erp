// Package app implementa a injeção de dependências do Nexo Flash.
//
// Wire conecta:
//   Repositórios PostgreSQL → Serviços → Handlers HTTP
//
// Uso em main.go:
//
//	container, err := app.Wire(cfg)
//	container.MechanicHandler.RegisterRoutes(mux)
//	container.TaxHandler.RegisterRoutes(mux)
//	...
package app

import (
	"fmt"

	"github.com/nexoflash/nexo-flash/internal/ai"
	"github.com/nexoflash/nexo-flash/internal/baas"
	"github.com/nexoflash/nexo-flash/internal/handlers"
	"github.com/nexoflash/nexo-flash/internal/modules/aesthetics"
	"github.com/nexoflash/nexo-flash/internal/modules/logistics"
	"github.com/nexoflash/nexo-flash/internal/modules/mechanic"
	"github.com/nexoflash/nexo-flash/internal/repository/postgres"
	"github.com/nexoflash/nexo-flash/internal/tax"
	"github.com/nexoflash/nexo-flash/pkg/cache"
)

// Config agrupa toda a configuração da aplicação.
type Config struct {
	DatabaseURL string
	RedisURL    string
	NatsURL     string
	JWTSecret   string
	BaseURL     string // ex: "https://app.nexoflash.com.br"
	WhatsAppToken string
}

// Container agrupa todos os handlers HTTP prontos para registrar rotas.
type Container struct {
	// Handlers
	MechanicHandler  *handlers.MechanicHandler
	BakeryHandler    *handlers.BakeryHandler
	TaxHandler       *handlers.TaxHandler
	LogisticsHandler *handlers.LogisticsHandler
	AestheticsHandler *handlers.AestheticsHandler
	AIHandler        *handlers.AIHandler
	PaymentHandler   *handlers.PaymentHandler

	// Repositórios (expostos para testes)
	DB        *postgres.DB
	TenantRepo *postgres.TenantRepo
	UserRepo   *postgres.UserRepo

	// Cleanup
	Close func()
}

// Wire inicializa toda a cadeia de dependências.
func Wire(cfg Config) (*Container, error) {
	// 1. PostgreSQL
	db, err := postgres.New(postgres.Config{DSN: cfg.DatabaseURL})
	if err != nil {
		return nil, fmt.Errorf("app.Wire: db: %w", err)
	}

	// 2. Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app.Wire: redis: %w", err)
	}

	// 3. Repositórios
	tenantRepo   := postgres.NewTenantRepo(db)
	userRepo     := postgres.NewUserRepo(db)
	mechanicRepo := postgres.NewMechanicRepo(db)
	taxRateRepo  := postgres.NewTaxRateRepo(db, redisClient)
	aiRepo       := postgres.NewAIRepo(db)
	logisticsRepo:= postgres.NewLogisticsRepo(db)
	paymentRepo  := postgres.NewPaymentRepo(db)
	sessionCache := cache.NewSessionCache(redisClient)

	// 4. Serviços
	taxEngine := tax.NewEngine(taxRateRepo)

	waClient := mechanic.NewNoOpWhatsApp() // trocar por implementação real
	_ = sessionCache // usado pelo auth middleware

	osSvc := mechanic.NewOSService(mechanicRepo, waClient, cfg.BaseURL)
	aiGateway := ai.NewGateway(aiRepo)
	concierge := ai.NewConcierge(aiGateway)
	contractSvc := logistics.NewContractService(logisticsRepo)
	agendaSvc := aesthetics.NewAgendaService(postgres.NewAestheticsRepo(db))

	// BaaS — gateway plugável (implementar Celcoin/BMP em produção)
	baasGateway := baas.NewNoOpGateway()
	paymentSvc := baas.NewPaymentService(baasGateway, paymentRepo)

	// 5. Handlers HTTP
	c := &Container{
		MechanicHandler:   handlers.NewMechanicHandler(osSvc),
		TaxHandler:        handlers.NewTaxHandler(taxEngine),
		LogisticsHandler:  handlers.NewLogisticsHandler(contractSvc),
		AestheticsHandler: handlers.NewAestheticsHandler(agendaSvc),
		AIHandler:         handlers.NewAIHandler(aiGateway, concierge),
		PaymentHandler:    handlers.NewPaymentHandler(paymentSvc),
		DB:                db,
		TenantRepo:        tenantRepo,
		UserRepo:          userRepo,
		Close: func() {
			db.Close()
		},
	}

	_ = tenantRepo
	_ = userRepo

	return c, nil
}
