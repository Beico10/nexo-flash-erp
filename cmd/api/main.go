// Package main e o ponto de entrada do servidor Nexo One ERP.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexoone/nexo-one/internal/app"
	"github.com/nexoone/nexo-one/internal/auth"
	"github.com/nexoone/nexo-one/pkg/middleware"

	// Registra todos os modulos via init()
	_ "github.com/nexoone/nexo-one/internal/modules/aesthetics"
	_ "github.com/nexoone/nexo-one/internal/modules/bakery"
	_ "github.com/nexoone/nexo-one/internal/modules/industry"
	_ "github.com/nexoone/nexo-one/internal/modules/logistics"
	_ "github.com/nexoone/nexo-one/internal/modules/mechanic"
	_ "github.com/nexoone/nexo-one/internal/modules/shoes"
)

var version = "1.0.0"

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8002")
	jwtSecret := getEnv("JWT_SECRET", "nexo-one-dev-secret-2026")

	slog.Info("Nexo One ERP iniciando", "version", version, "port", port)

	container, err := app.Wire(app.Config{
		JWTSecret: jwtSecret,
		BaseURL:   getEnv("BASE_URL", "http://localhost:"+port),
		Port:      port,
	})
	if err != nil {
		slog.Error("falha ao inicializar dependencias", "err", err)
		os.Exit(1)
	}
	defer container.Close()

	mux := buildRouter(container)

	srv := &http.Server{
		Addr:         ":" + port,
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown forcado", "err", err)
	}
	slog.Info("servidor encerrado")
}

func buildRouter(c *app.Container) http.Handler {
	mux := http.NewServeMux()
	authMW := middleware.AuthMiddleware(c.AuthService)

	// Health check (publico)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"` + version + `","engine":"fiscal_ibs_cbs_2026"}`))
	})

	// Auth (publico)
	authHandler := auth.NewHandler(c.AuthService)
	authHandler.RegisterRoutes(mux)

	// Rotas protegidas - precisam de JWT via AuthMiddleware
	protectedMux := http.NewServeMux()
	c.TaxHandler.RegisterRoutes(protectedMux)
	c.MechanicHandler.RegisterRoutes(protectedMux)
	c.BakeryHandler.RegisterRoutes(protectedMux)
	c.LogisticsHandler.RegisterRoutes(protectedMux)
	c.AestheticsHandler.RegisterRoutes(protectedMux)
	c.AIHandler.RegisterRoutes(protectedMux)
	c.PaymentHandler.RegisterRoutes(protectedMux)

	// Monta as rotas protegidas no mux principal com middleware
	mux.Handle("/api/v1/", authMW(protectedMux))

	return corsMiddleware(mux)
}

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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
