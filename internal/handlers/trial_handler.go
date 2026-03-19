// Package handlers — endpoints de trial, verificação e onboarding.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/journey"
	"github.com/nexoone/nexo-one/internal/trial"
)

// ════════════════════════════════════════════════════════════
// TRIAL & VERIFICAÇÃO WHATSAPP
// ════════════════════════════════════════════════════════════

type TrialHandler struct {
	trialSvc   *trial.Service
	journeySvc *journey.Service
}

func NewTrialHandler(t *trial.Service, j *journey.Service) *TrialHandler {
	return &TrialHandler{trialSvc: t, journeySvc: j}
}

// StartVerification POST /api/auth/verify/start
// Inicia verificação por WhatsApp.
func (h *TrialHandler) StartVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		DeviceHash  string `json:"device_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		jsonError(w, "Telefone é obrigatório", http.StatusBadRequest)
		return
	}

	// Pegar IP do request
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	// Iniciar verificação
	whatsappURL, err := h.trialSvc.StartVerification(r.Context(), req.Phone, req.Email, req.DeviceHash, ip)
	if err != nil {
		if err == trial.ErrPhoneAlreadyUsed {
			jsonError(w, "Este telefone já possui uma conta. Faça login.", http.StatusConflict)
			return
		}
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Track evento
	h.journeySvc.Track(r.Context(), &journey.Event{
		AnonymousID:   req.DeviceHash,
		EventName:     "verification_started",
		EventCategory: "conversion",
		Properties:    map[string]any{"method": "whatsapp"},
	})

	jsonOK(w, map[string]interface{}{
		"whatsapp_url": whatsappURL,
		"message":      "Clique no link para abrir o WhatsApp e enviar o código",
		"expires_in":   300, // 5 minutos
	})
}

// VerifyCode POST /api/auth/verify/confirm
// Verifica o código manualmente (fallback se webhook não funcionar).
func (h *TrialHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	tc, err := h.trialSvc.VerifyCode(r.Context(), req.Phone, req.Code)
	if err != nil {
		switch err {
		case trial.ErrCodeExpired:
			jsonError(w, "Código expirado. Solicite um novo.", http.StatusGone)
		case trial.ErrCodeInvalid:
			jsonError(w, "Código inválido. Verifique e tente novamente.", http.StatusBadRequest)
		case trial.ErrTooManyAttempts:
			jsonError(w, "Muitas tentativas. Aguarde 15 minutos.", http.StatusTooManyRequests)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Track evento
	h.journeySvc.Track(r.Context(), &journey.Event{
		TenantID:      tc.TenantID,
		EventName:     "verification_completed",
		EventCategory: "conversion",
	})

	jsonOK(w, map[string]interface{}{
		"verified":   true,
		"message":    "Telefone verificado com sucesso!",
		"tenant_id":  tc.TenantID,
	})
}

// WhatsAppWebhook POST /api/webhooks/whatsapp
// Recebe mensagens do WhatsApp (código de verificação).
func (h *TrialHandler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		From string `json:"from"` // Número que enviou
		Body string `json:"body"` // Conteúdo da mensagem
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	tc, err := h.trialSvc.ProcessWhatsAppMessage(r.Context(), payload.From, payload.Body)
	if err != nil {
		// Não retorna erro - webhook deve sempre retornar 200
		// Mas loga para análise
		jsonOK(w, map[string]interface{}{
			"processed": false,
			"reason":    err.Error(),
		})
		return
	}

	jsonOK(w, map[string]interface{}{
		"processed": true,
		"tenant_id": tc.TenantID,
	})
}

// ════════════════════════════════════════════════════════════
// ONBOARDING
// ════════════════════════════════════════════════════════════

type OnboardingHandler struct {
	journeySvc *journey.Service
}

func NewOnboardingHandler(j *journey.Service) *OnboardingHandler {
	return &OnboardingHandler{journeySvc: j}
}

// GetSteps GET /api/onboarding/steps
// Retorna os passos do onboarding para o nicho do tenant.
func (h *OnboardingHandler) GetSteps(w http.ResponseWriter, r *http.Request) {
	businessType := r.URL.Query().Get("business_type")
	if businessType == "" {
		// Pegar do contexto (tenant autenticado)
		if bt, ok := r.Context().Value("business_type").(string); ok {
			businessType = bt
		}
	}

	if businessType == "" {
		jsonError(w, "business_type é obrigatório", http.StatusBadRequest)
		return
	}

	steps, err := h.journeySvc.GetOnboardingSteps(r.Context(), businessType)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"steps": steps,
		"total": len(steps),
	})
}

// GetProgress GET /api/onboarding/progress
// Retorna o progresso do onboarding do tenant.
func (h *OnboardingHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	progress, err := h.journeySvc.GetOnboardingProgress(r.Context(), tenantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	steps, _ := h.journeySvc.GetOnboardingSteps(r.Context(), progress.BusinessType)

	// Calcular porcentagem
	percent := 0
	if len(steps) > 0 {
		percent = (len(progress.CompletedSteps) * 100) / len(steps)
	}

	jsonOK(w, map[string]interface{}{
		"progress":        progress,
		"steps":           steps,
		"percent":         percent,
		"completed":       progress.CompletedAt != nil,
		"reward_days":     calculateRewardDays(steps, progress.CompletedSteps),
	})
}

// CompleteStep POST /api/onboarding/complete
// Marca um passo como completo.
func (h *OnboardingHandler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	var req struct {
		StepCode string `json:"step_code"`
		Skipped  bool   `json:"skipped"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.journeySvc.CompleteOnboardingStep(r.Context(), tenantID, userID, req.StepCode, req.Skipped); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"message":    "Passo registrado!",
		"step_code":  req.StepCode,
		"skipped":    req.Skipped,
	})
}

// SkipOnboarding POST /api/onboarding/skip
// Pula o onboarding (usuário quer explorar sozinho).
func (h *OnboardingHandler) SkipOnboarding(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	if err := h.journeySvc.SkipOnboarding(r.Context(), tenantID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"message": "OK! Você pode voltar ao onboarding a qualquer momento pelo menu.",
	})
}

// ════════════════════════════════════════════════════════════
// TRACKING (Frontend envia eventos)
// ════════════════════════════════════════════════════════════

type TrackingHandler struct {
	journeySvc *journey.Service
}

func NewTrackingHandler(j *journey.Service) *TrackingHandler {
	return &TrackingHandler{journeySvc: j}
}

// TrackEvent POST /api/track
// Recebe eventos do frontend.
func (h *TrackingHandler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventName     string         `json:"event_name"`
		EventCategory string         `json:"event_category"`
		PagePath      string         `json:"page_path"`
		PageTitle     string         `json:"page_title"`
		Properties    map[string]any `json:"properties"`
		AnonymousID   string         `json:"anonymous_id"`
		SessionID     string         `json:"session_id"`
		DeviceType    string         `json:"device_type"`
		TimeOnPage    int            `json:"time_on_page"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Pegar tenant/user do contexto se autenticado
	var tenantID, userID string
	if tid, ok := r.Context().Value("tenant_id").(string); ok {
		tenantID = tid
	}
	if uid, ok := r.Context().Value("user_id").(string); ok {
		userID = uid
	}

	event := &journey.Event{
		TenantID:      tenantID,
		UserID:        userID,
		AnonymousID:   req.AnonymousID,
		EventName:     req.EventName,
		EventCategory: req.EventCategory,
		PagePath:      req.PagePath,
		PageTitle:     req.PageTitle,
		Properties:    req.Properties,
		SessionID:     req.SessionID,
		DeviceType:    req.DeviceType,
		TimeOnPage:    req.TimeOnPage,
		OccurredAt:    time.Now(),
	}

	if err := h.journeySvc.Track(r.Context(), event); err != nil {
		// Não falha - tracking não deve bloquear o usuário
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ════════════════════════════════════════════════════════════
// ANALYTICS (Admin)
// ════════════════════════════════════════════════════════════

type AnalyticsHandler struct {
	journeySvc *journey.Service
}

func NewAnalyticsHandler(j *journey.Service) *AnalyticsHandler {
	return &AnalyticsHandler{journeySvc: j}
}

// GetFunnel GET /api/admin/analytics/funnel
// Retorna métricas do funil.
func (h *AnalyticsHandler) GetFunnel(w http.ResponseWriter, r *http.Request) {
	businessType := r.URL.Query().Get("business_type")
	
	// Período (default: últimos 7 dias)
	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	metrics, err := h.journeySvc.GetFunnelRange(r.Context(), from, to, businessType)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calcular totais e conversão
	var totals journey.FunnelMetrics
	for _, m := range metrics {
		totals.Visits += m.Visits
		totals.SignupsStarted += m.SignupsStarted
		totals.SignupsCompleted += m.SignupsCompleted
		totals.PhoneVerified += m.PhoneVerified
		totals.OnboardingStarted += m.OnboardingStarted
		totals.OnboardingCompleted += m.OnboardingCompleted
		totals.FirstAction += m.FirstAction
		totals.TrialConverted += m.TrialConverted
	}

	if totals.Visits > 0 {
		totals.ConversionRate = float64(totals.TrialConverted) / float64(totals.Visits) * 100
	}

	jsonOK(w, map[string]interface{}{
		"period":  map[string]string{"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02")},
		"daily":   metrics,
		"totals":  totals,
	})
}

// GetDropPoints GET /api/admin/analytics/drops
// Retorna usuários travados.
func (h *AnalyticsHandler) GetDropPoints(w http.ResponseWriter, r *http.Request) {
	stage := r.URL.Query().Get("stage") // signup, verification, onboarding, activation
	
	drops, err := h.journeySvc.GetDropPoints(r.Context(), stage, 1) // Travados há pelo menos 1 dia
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Agrupar por estágio
	byStage := make(map[string][]*journey.DropPoint)
	for _, d := range drops {
		byStage[d.Stage] = append(byStage[d.Stage], d)
	}

	jsonOK(w, map[string]interface{}{
		"total":    len(drops),
		"by_stage": byStage,
		"drops":    drops,
	})
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func calculateRewardDays(steps []*journey.OnboardingStep, completed []string) int {
	total := 0
	completedSet := make(map[string]bool)
	for _, c := range completed {
		completedSet[c] = true
	}
	for _, s := range steps {
		if completedSet[s.StepCode] {
			total += s.RewardDays
		}
	}
	return total
}
