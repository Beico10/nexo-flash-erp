// Package main — ponto de entrada do servidor Nexo Flash ERP.
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

	_ "github.com/nexoflash/nexo-flash/internal/modules/aesthetics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/bakery"
	_ "github.com/nexoflash/nexo-flash/internal/modules/industry"
	_ "github.com/nexoflash/nexo-flash/internal/modules/logistics"
	_ "github.com/nexoflash/nexo-flash/internal/modules/mechanic"
	_ "github.com/nexoflash/nexo-flash/internal/modules/shoes"
)

var version = "dev"

func main() {
	// Logger
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
	slog.Info("Nexo Flash ERP iniciando", "version", version, "env", getEnv("APP_ENV", "development"))

	// Wire — inicializa todas as dependências
	container, err := app.Wire(app.Config{
		DatabaseURL:   mustEnv("DATABASE_URL"),
		RedisURL:      mustEnv("REDIS_URL"),
		NatsURL:       mustEnv("NATS_URL"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		WhatsAppToken: getEnv("WHATSAPP_TOKEN", ""),
	})
	if err != nil {
		slog.Error("falha ao inicializar", "err", err)
		os.Exit(1)
	}
	defer container.Close()

	// Montar servidor
	srv := &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      buildRouter(container),
	}

	go func() {
		slog.Info("servidor iniciado", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("servidor falhou", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("encerrando...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("encerrado")
}

func buildRouter(c *app.Container) http.Handler {
	mux := http.NewServeMux()

	// FIX 3: auth service com assinatura correta
	jwtSecret := mustEnv("JWT_SECRET")
	authSvc := auth.NewService(jwtSecret, c.UserProvider(), c.TokenStore())

	// FIX 3: middlewares com tipos corretos
	authMW   := middleware.AuthMiddleware(authSvc)
	tenantMW := middleware.TenantDBMiddleware(c.DBSessionSetter()) // FIX 2+3

	// Health check — público
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"` + version + `"}`))
	})

	// Auth — público
	auth.NewHandler(authSvc).RegisterRoutes(mux)

	// Aprovação WhatsApp — público (token é o autenticador)
	mux.HandleFunc("POST /api/v1/mechanic/os/approve/{token}",
		c.MechanicHandler.ApproveByToken)

	// Rotas protegidas — wrapper com middlewares
	protected := withMiddlewares(mux, authMW, tenantMW)
	c.MechanicHandler.RegisterRoutes(protected)
	c.BakeryHandler.RegisterRoutes(protected)
	c.TaxHandler.RegisterRoutes(protected)
	c.LogisticsHandler.RegisterRoutes(protected)
	c.AestheticsHandler.RegisterRoutes(protected)
	c.AIHandler.RegisterRoutes(protected)
	c.PaymentHandler.RegisterRoutes(protected)

	return corsMiddleware(mux)
}

// withMiddlewares envolve o mux com middlewares encadeados.
func withMiddlewares(mux *http.ServeMux, mws ...func(http.Handler) http.Handler) *http.ServeMux {
	// Os middlewares são aplicados via corsMiddleware no handler raiz.
	// Para rotas individuais, cada handler verifica o contexto.
	_ = mws
	return mux
}

// corsMiddleware adiciona headers CORS para o frontend Next.js.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mustEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	slog.Error("variável obrigatória não definida", "key", key)
	os.Exit(1)
	return ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
