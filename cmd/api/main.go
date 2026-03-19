// Package main é o ponto de entrada do servidor Nexo Flash ERP.
//
// Inicialização na ordem correta:
//  1. Configuração (env vars)
//  2. Banco de dados (PostgreSQL + RLS)
//  3. Cache (Redis)
//  4. Event Bus (NATS JetStream)
//  5. Módulos (via ModuleRegistry — baseado no business_type do tenant)
//  6. Servidor HTTP (Traefik faz TLS na frente)
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Importar os módulos aqui garante que seus init() são executados,
	// registrando-os automaticamente no ModuleRegistry.
	_ "github.com/nexoflash/nexo-flash/internal/modules/aesthetics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/bakery"
	_ "github.com/nexoflash/nexo-flash/internal/modules/industry"
	_ "github.com/nexoflash/nexo-flash/internal/modules/logistics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/mechanic"
	_ "github.com/nexoflash/nexo-flash/internal/modules/shoes"
)

func main() {
	// Logger estruturado (produção: JSON; desenvolvimento: texto)
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("Nexo Flash ERP iniciando",
		"env", getEnv("APP_ENV", "development"),
		"version", version,
	)

	// Configurações obrigatórias — falha imediatamente se ausentes
	dbURL := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	natsURL := mustEnv("NATS_URL")
	jwtSecret := mustEnv("JWT_SECRET")
	_ = dbURL
	_ = redisURL
	_ = natsURL
	_ = jwtSecret

	// TODO: inicializar conexões (db, redis, nats) e injetar nas dependências
	// db := database.Connect(dbURL)
	// redis := cache.Connect(redisURL)
	// bus, _ := eventbus.New(natsURL)
	// taxEngine := tax.NewEngine(tax.NewDBRateRepository(db, cache.NewNCMCache(redis)))
	// aiGateway := ai.NewGateway(ai.NewDBRepository(db))
	// concierge := ai.NewConcierge(aiGateway)

	// Servidor HTTP com timeouts seguros
	srv := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      buildRouter(),
	}

	// Graceful shutdown
	go func() {
		slog.Info("servidor HTTP iniciado", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("servidor falhou", "err", err)
			os.Exit(1)
		}
	}()

	// Aguarda sinal de encerramento (SIGTERM do Docker/Kubernetes)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown forçado", "err", err)
	}
	slog.Info("servidor encerrado com sucesso")
}

// buildRouter monta o roteador HTTP com todos os middlewares e rotas.
func buildRouter() http.Handler {
	mux := http.NewServeMux()

	// Health check — usado pelo Docker e pelo Traefik
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","version":"` + version + `"}`))
	})

	// TODO: registrar rotas de cada módulo via ModuleRegistry
	// for _, mod := range core.ModulesForTenant(businessType) {
	//     mod.Init()
	// }

	return mux
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("variável de ambiente obrigatória não definida", "key", key)
		os.Exit(1)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// version é injetado no build via -ldflags "-X main.version=v1.0.0"
var version = "dev"
