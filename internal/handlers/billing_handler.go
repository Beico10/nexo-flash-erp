// Package handlers — endpoints de billing (planos, assinaturas, upgrade).
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nexoone/nexo-one/internal/billing"
)

type BillingHandler struct {
	svc *billing.Service
}

func NewBillingHandler(svc *billing.Service) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// ════════════════════════════════════════════════════════════
// ENDPOINTS PÚBLICOS (Página de Preços)
// ════════════════════════════════════════════════════════════

// ListPlans GET /api/billing/plans - Lista planos para página de preços.
func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{
		"plans": plans,
	})
}

// ValidateCoupon POST /api/billing/coupon/validate - Valida cupom.
func (h *BillingHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code     string `json:"code"`
		PlanCode string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	coupon, err := h.svc.ValidateCoupon(r.Context(), req.Code, req.PlanCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{
		"valid":          true,
		"discount_type":  coupon.DiscountType,
		"discount_value": coupon.DiscountValue,
		"description":    coupon.Description,
	})
}

// ════════════════════════════════════════════════════════════
// ENDPOINTS AUTENTICADOS (Tenant)
// ════════════════════════════════════════════════════════════

// GetSubscription GET /api/billing/subscription - Retorna assinatura atual.
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	sub, err := h.svc.GetSubscription(r.Context(), tenantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	usage, _ := h.svc.GetUsageStatus(r.Context(), tenantID)

	jsonOK(w, map[string]interface{}{
		"subscription": sub,
		"usage":        usage,
	})
}

// ConvertTrial POST /api/billing/convert - Converte trial em assinatura paga.
func (h *BillingHandler) ConvertTrial(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	var req struct {
		PaymentMethod string `json:"payment_method"` // "pix", "credit_card", "boleto"
		CouponCode    string `json:"coupon_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	sub, err := h.svc.ConvertTrial(r.Context(), tenantID, req.PaymentMethod, req.CouponCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{
		"subscription": sub,
		"message":      "Assinatura ativada com sucesso!",
	})
}

// ChangePlan POST /api/billing/change-plan - Upgrade ou downgrade de plano.
func (h *BillingHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	var req struct {
		PlanCode string `json:"plan_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	sub, err := h.svc.ChangePlan(r.Context(), tenantID, req.PlanCode)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{
		"subscription": sub,
		"message":      "Plano alterado com sucesso!",
	})
}

// GetUsage GET /api/billing/usage - Retorna uso atual vs limites.
func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	usage, err := h.svc.GetUsageStatus(r.Context(), tenantID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"usage": usage,
	})
}

// CheckFeature GET /api/billing/feature/{feature} - Verifica se tem acesso a recurso.
func (h *BillingHandler) CheckFeature(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	feature := r.URL.Query().Get("feature")

	if feature == "" {
		jsonError(w, "feature é obrigatório", http.StatusBadRequest)
		return
	}

	hasAccess, err := h.svc.HasFeature(r.Context(), tenantID, feature)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"feature":    feature,
		"has_access": hasAccess,
	})
}

// ════════════════════════════════════════════════════════════
// ENDPOINTS ADMIN MASTER (Gerenciar Planos)
// ════════════════════════════════════════════════════════════

// AdminListPlans GET /api/admin/billing/plans - Lista todos os planos (ativos e inativos).
func (h *BillingHandler) AdminListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListAllPlans(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{
		"plans": plans,
	})
}

// AdminUpdatePlan PUT /api/admin/billing/plans/{id} - Atualiza plano.
func (h *BillingHandler) AdminUpdatePlan(w http.ResponseWriter, r *http.Request) {
	var plan billing.Plan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		jsonError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdatePlan(r.Context(), &plan); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"message": "Plano atualizado com sucesso",
		"plan":    plan,
	})
}

// ════════════════════════════════════════════════════════════
// MIDDLEWARE: Verificar Limite antes de Ação
// ════════════════════════════════════════════════════════════

// LimitMiddleware verifica se tenant pode executar ação.
func (h *BillingHandler) LimitMiddleware(metric string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			if err := h.svc.CanPerform(r.Context(), tenantID, metric); err != nil {
				jsonError(w, err.Error(), http.StatusPaymentRequired)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// FeatureMiddleware verifica se tenant tem acesso a recurso.
func (h *BillingHandler) FeatureMiddleware(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			hasAccess, err := h.svc.HasFeature(r.Context(), tenantID, feature)
			if err != nil || !hasAccess {
				jsonError(w, "Recurso não disponível no seu plano. Faça upgrade para ter acesso.", http.StatusPaymentRequired)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func (h *BillingHandler) GetService() *billing.Service {
	return h.svc
}
