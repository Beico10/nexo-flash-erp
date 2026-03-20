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

var version = "2.0.0-go"

func main() {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8001")
	jwtSecret := getEnv("JWT_SECRET", "nexo-one-dev-secret-2026")

	slog.Info("Nexo One ERP iniciando (Go puro)", "version", version, "port", port)

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
	webAuthMW := middleware.WebAuthMiddleware(c.AuthService)

	// ═══════════════════════════════════════════════════════════════════
	// PAGINAS WEB (Templates Go + HTMX)
	// ═══════════════════════════════════════════════════════════════════
	
	// Páginas públicas (sem auth)
	mux.HandleFunc("GET /{$}", c.PageHandler.HandleRoot) // {$} = match apenas raiz exata
	mux.HandleFunc("GET /login", c.PageHandler.HandleLogin)
	mux.HandleFunc("GET /pricing", c.PageHandler.HandlePricing)
	mux.HandleFunc("GET /simulador-fiscal", c.PageHandler.HandleSimulador)
	
	// Páginas protegidas (com auth via cookie/header)
	protectedPagesMux := http.NewServeMux()
	c.PageHandler.RegisterProtectedRoutes(protectedPagesMux)
	mux.Handle("GET /dashboard", webAuthMW(http.HandlerFunc(c.PageHandler.HandleDashboard)))
	mux.Handle("GET /mechanic", webAuthMW(http.HandlerFunc(c.PageHandler.HandleMechanic)))
	mux.Handle("GET /bakery", webAuthMW(http.HandlerFunc(c.PageHandler.HandleBakery)))
	mux.Handle("GET /aesthetics", webAuthMW(http.HandlerFunc(c.PageHandler.HandleAesthetics)))
	mux.Handle("GET /logistics", webAuthMW(http.HandlerFunc(c.PageHandler.HandleLogistics)))
	mux.Handle("GET /industry", webAuthMW(http.HandlerFunc(c.PageHandler.HandleIndustry)))
	mux.Handle("GET /shoes", webAuthMW(http.HandlerFunc(c.PageHandler.HandleShoes)))
	mux.Handle("GET /nfe", webAuthMW(http.HandlerFunc(c.PageHandler.HandleNFE)))
	mux.Handle("GET /expenses", webAuthMW(http.HandlerFunc(c.PageHandler.HandleExpenses)))
	mux.Handle("GET /copilot", webAuthMW(http.HandlerFunc(c.PageHandler.HandleCopilot)))
	mux.Handle("GET /ai-approvals", webAuthMW(http.HandlerFunc(c.PageHandler.HandleAIApprovals)))
	mux.Handle("GET /settings", webAuthMW(http.HandlerFunc(c.PageHandler.HandleSettings)))

	// ═══════════════════════════════════════════════════════════════════
	// API REST
	// ═══════════════════════════════════════════════════════════════════

	// Health check (publico)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"` + version + `","engine":"fiscal_ibs_cbs_2026","stack":"go-puro"}`))
	})

	// Auth (publico)
	authHandler := auth.NewHandler(c.AuthService)
	authHandler.RegisterRoutes(mux)

	// Simulador Fiscal (publico - sem autenticacao)
	c.SimulatorHandler.RegisterPublicRoutes(mux)

	// Billing publico (planos e cupons)
	mux.HandleFunc("GET /api/billing/plans", c.BillingHandler.ListPlans)
	mux.HandleFunc("POST /api/billing/coupon/validate", c.BillingHandler.ValidateCoupon)

	// Trial & Verificacao WhatsApp (publico)
	mux.HandleFunc("POST /api/auth/verify/start", c.TrialHandler.StartVerification)
	mux.HandleFunc("POST /api/auth/verify/confirm", c.TrialHandler.VerifyCode)
	mux.HandleFunc("POST /api/webhooks/whatsapp", c.TrialHandler.WhatsAppWebhook)

	// Rotas protegidas - precisam de JWT via AuthMiddleware
	protectedMux := http.NewServeMux()
	c.DashboardHandler.RegisterRoutes(protectedMux)
	c.TaxHandler.RegisterRoutes(protectedMux)
	c.MechanicHandler.RegisterRoutes(protectedMux)
	c.BakeryHandler.RegisterRoutes(protectedMux)
	c.LogisticsHandler.RegisterRoutes(protectedMux)
	c.AestheticsHandler.RegisterRoutes(protectedMux)
	c.AIHandler.RegisterRoutes(protectedMux)
	c.PaymentHandler.RegisterRoutes(protectedMux)
	c.IndustryHandler.RegisterRoutes(protectedMux)
	c.ShoesHandler.RegisterRoutes(protectedMux)
	c.NFEHandler.RegisterRoutes(protectedMux)
	c.CopilotHandler.RegisterRoutes(protectedMux)

	// Billing autenticado
	protectedMux.HandleFunc("GET /api/v1/billing/subscription", c.BillingHandler.GetSubscription)
	protectedMux.HandleFunc("POST /api/v1/billing/convert", c.BillingHandler.ConvertTrial)
	protectedMux.HandleFunc("POST /api/v1/billing/change-plan", c.BillingHandler.ChangePlan)
	protectedMux.HandleFunc("GET /api/v1/billing/usage", c.BillingHandler.GetUsage)
	protectedMux.HandleFunc("GET /api/v1/billing/feature", c.BillingHandler.CheckFeature)

	// Admin - Gestao de planos
	protectedMux.HandleFunc("GET /api/v1/admin/plans", c.BillingHandler.GetAllPlansAdmin)
	protectedMux.HandleFunc("PUT /api/v1/admin/plans", c.BillingHandler.UpdatePlan)

	// Onboarding autenticado
	protectedMux.HandleFunc("GET /api/v1/onboarding/steps", c.OnboardingHandler.GetSteps)
	protectedMux.HandleFunc("GET /api/v1/onboarding/progress", c.OnboardingHandler.GetProgress)
	protectedMux.HandleFunc("POST /api/v1/onboarding/complete", c.OnboardingHandler.CompleteStep)
	protectedMux.HandleFunc("POST /api/v1/onboarding/skip", c.OnboardingHandler.SkipOnboarding)

	// Journey tracking autenticado
	protectedMux.HandleFunc("POST /api/v1/track", c.TrackingHandler.TrackEvent)
	protectedMux.HandleFunc("GET /api/v1/analytics/funnel", c.AnalyticsHandler.GetFunnel)
	protectedMux.HandleFunc("GET /api/v1/analytics/drops", c.AnalyticsHandler.GetDropPoints)

	// Despesas - QR Code Scanner
	protectedMux.HandleFunc("POST /api/v1/expenses/scan", c.ExpenseHandler.ScanQRCode)
	protectedMux.HandleFunc("POST /api/v1/expenses/parse-qr", c.ExpenseHandler.ParseQRCode)
	protectedMux.HandleFunc("POST /api/v1/expenses/upload-xml", c.ExpenseHandler.UploadXML)
	protectedMux.HandleFunc("POST /api/v1/expenses", c.ExpenseHandler.CreateExpense)
	protectedMux.HandleFunc("GET /api/v1/expenses", c.ExpenseHandler.ListExpenses)
	protectedMux.HandleFunc("GET /api/v1/expenses/{id}", c.ExpenseHandler.GetExpense)
	protectedMux.HandleFunc("DELETE /api/v1/expenses/{id}", c.ExpenseHandler.DeleteExpense)
	protectedMux.HandleFunc("GET /api/v1/expenses/categories", c.ExpenseHandler.GetCategories)
	protectedMux.HandleFunc("GET /api/v1/expenses/summary", c.ExpenseHandler.GetSummary)
	protectedMux.HandleFunc("GET /api/v1/expenses/tax-report", c.ExpenseHandler.GetTaxReport)

	// Monta as rotas protegidas no mux principal com middleware
	// Usar handler que aceita qualquer método
	mux.Handle("/api/v1/", authMW(http.StripPrefix("", protectedMux)))

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
