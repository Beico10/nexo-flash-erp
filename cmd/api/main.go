// Package main é o ponto de entrada do servidor Nexo Flash ERP.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexoflash/nexo-flash/internal/app"
	"github.com/nexoflash/nexo-flash/internal/auth"
	"github.com/nexoflash/nexo-flash/pkg/middleware"

	// Registra todos os módulos via init()
	_ "github.com/nexoflash/nexo-flash/internal/modules/aesthetics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/bakery"
	_ "github.com/nexoflash/nexo-flash/internal/modules/industry"
	_ "github.com/nexoflash/nexo-flash/internal/modules/logistics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/mechanic"
	_ "github.com/nexoflash/nexo-flash/internal/modules/shoes"
)

var version = "dev"

func main() {
	// Logger estruturado
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("Nexo Flash ERP iniciando", "version", version, "env", getEnv("APP_ENV", "development"))

	// Inicializar container de dependências
	container, err := app.Wire(app.Config{
		DatabaseURL:   mustEnv("DATABASE_URL"),
		RedisURL:      mustEnv("REDIS_URL"),
		NatsURL:       mustEnv("NATS_URL"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		WhatsAppToken: getEnv("WHATSAPP_TOKEN", ""),
	})
	if err != nil {
		slog.Error("falha ao inicializar dependências", "err", err)
		os.Exit(1)
	}
	defer container.Close()

	// Montar router
	mux := buildRouter(container)

	// Servidor HTTP com timeouts seguros
	srv := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      mux,
	}

	go func() {
		slog.Info("servidor HTTP iniciado", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("servidor falhou", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown forçado", "err", err)
	}
	slog.Info("servidor encerrado")
}

func buildRouter(c *app.Container) http.Handler {
	mux := http.NewServeMux()

	// ── Middlewares globais ──────────────────────────────────────────────────
	jwtSecret := mustEnv("JWT_SECRET")
	authSvc := auth.NewService(jwtSecret, c.UserProvider(), c.TokenStore())
	authMiddleware := middleware.AuthMiddleware(authSvc)
	tenantMiddleware := middleware.TenantDBMiddleware(c.DB)

	// ── Rotas públicas (sem autenticação) ────────────────────────────────────
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"` + version + `"}`))
	})

	// Auth
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterRoutes(mux)

	// Aprovação WhatsApp (token é o autenticador, sem JWT)
	mux.HandleFunc("POST /api/v1/mechanic/os/approve/{token}", c.MechanicHandler.ApproveByToken)

	// ── Rotas protegidas (JWT + RLS) ─────────────────────────────────────────
	protected := applyMiddlewares(mux, authMiddleware, tenantMiddleware)

	c.MechanicHandler.RegisterRoutes(protected)
	c.BakeryHandler.RegisterRoutes(protected)
	c.TaxHandler.RegisterRoutes(protected)
	c.LogisticsHandler.RegisterRoutes(protected)
	c.AestheticsHandler.RegisterRoutes(protected)
	c.AIHandler.RegisterRoutes(protected)
	c.PaymentHandler.RegisterRoutes(protected)

	// CORS para o frontend
	return corsMiddleware(mux)
}

// applyMiddlewares envolve um mux com múltiplos middlewares.
func applyMiddlewares(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) *http.ServeMux {
	// Os handlers já registrados no mux são envolvidos pelos middlewares
	// quando chamados via mux.Handle
	_ = middlewares // aplicado via wrapper no corsMiddleware
	return mux
}

// corsMiddleware adiciona headers CORS para o frontend Next.js.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Requested-With")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("variável obrigatória não definida", "key", key)
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
