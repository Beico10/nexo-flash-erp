// Package handlers — endpoints Enterprise (API Keys, Filiais, Webhooks).
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexoone/nexo-one/internal/enterprise"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type EnterpriseHandler struct {
	svc *enterprise.Service
}

func NewEnterpriseHandler(svc *enterprise.Service) *EnterpriseHandler {
	return &EnterpriseHandler{svc: svc}
}

func (h *EnterpriseHandler) RegisterRoutes(mux *http.ServeMux) {
	// API Keys
	mux.HandleFunc("GET /api/v1/enterprise/api-keys", h.ListAPIKeys)
	mux.HandleFunc("POST /api/v1/enterprise/api-keys", h.CreateAPIKey)
	mux.HandleFunc("DELETE /api/v1/enterprise/api-keys/{id}", h.RevokeAPIKey)

	// Filiais
	mux.HandleFunc("GET /api/v1/enterprise/subsidiaries", h.ListSubsidiaries)
	mux.HandleFunc("POST /api/v1/enterprise/subsidiaries", h.CreateSubsidiary)
	mux.HandleFunc("GET /api/v1/enterprise/consolidated", h.ConsolidatedReport)

	// Webhooks
	mux.HandleFunc("GET /api/v1/enterprise/webhooks", h.ListWebhooks)
	mux.HandleFunc("POST /api/v1/enterprise/webhooks", h.CreateWebhook)
}

// CreateAPIKey POST /api/v1/enterprise/api-keys
func (h *EnterpriseHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	userID := ""
	if claims != nil {
		userID = claims.UserID
	}

	var req struct {
		Name      string   `json:"name"`
		Scopes    []string `json:"scopes"`
		RateLimit int      `json:"rate_limit"`
		ExpiresAt string   `json:"expires_at"` // YYYY-MM-DD ou vazio
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Nome da chave é obrigatório")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	key, err := h.svc.CreateAPIKey(r.Context(), tenantID, req.Name, userID, req.Scopes, expiresAt, req.RateLimit)
	if err != nil {
		switch err {
		case enterprise.ErrMaxKeysReached:
			respondError(w, http.StatusConflict, "Limite de 10 API keys atingido")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "API Key criada com sucesso",
		"key":        key.Key, // ⚠️ Mostrar apenas uma vez
		"key_prefix": key.KeyPrefix,
		"name":       key.Name,
		"scopes":     key.Scopes,
		"rate_limit": key.RateLimit,
		"expires_at": key.ExpiresAt,
		"warning":    "Guarde esta chave — ela não poderá ser visualizada novamente",
	})
}

// ListAPIKeys GET /api/v1/enterprise/api-keys
func (h *EnterpriseHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	keys, err := h.svc.ListAPIKeys(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Nunca retornar a chave completa na listagem
	formatted := make([]map[string]interface{}, len(keys))
	for i, k := range keys {
		formatted[i] = map[string]interface{}{
			"id":           k.ID,
			"name":         k.Name,
			"key_prefix":   k.KeyPrefix,
			"scopes":       k.Scopes,
			"rate_limit":   k.RateLimit,
			"is_active":    k.IsActive,
			"last_used_at": k.LastUsedAt,
			"expires_at":   k.ExpiresAt,
			"created_at":   k.CreatedAt,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"api_keys": formatted,
		"count":    len(formatted),
		"max":      enterprise.MaxKeysPerTenant,
	})
}

// RevokeAPIKey DELETE /api/v1/enterprise/api-keys/{id}
func (h *EnterpriseHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")

	if err := h.svc.RevokeAPIKey(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusNotFound, "API Key não encontrada")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"message": "API Key revogada"})
}

// ListSubsidiaries GET /api/v1/enterprise/subsidiaries
func (h *EnterpriseHandler) ListSubsidiaries(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	subs, err := h.svc.ListSubsidiaries(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subsidiaries": subs,
		"count":        len(subs),
	})
}

// CreateSubsidiary POST /api/v1/enterprise/subsidiaries
func (h *EnterpriseHandler) CreateSubsidiary(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	var sub enterprise.Subsidiary
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	sub.TenantID = tenantID

	if err := h.svc.CreateSubsidiary(r.Context(), &sub); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Filial criada com sucesso",
		"subsidiary": sub,
	})
}

// ConsolidatedReport GET /api/v1/enterprise/consolidated
func (h *EnterpriseHandler) ConsolidatedReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	now := time.Now()
	report, err := h.svc.GetConsolidatedReport(r.Context(), tenantID, now.Year(), int(now.Month()))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, report)
}

// ListWebhooks GET /api/v1/enterprise/webhooks
func (h *EnterpriseHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	webhooks, err := h.svc.ListWebhooks(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"webhooks": webhooks,
		"count":    len(webhooks),
	})
}

// CreateWebhook POST /api/v1/enterprise/webhooks
func (h *EnterpriseHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	var webhook enterprise.WebhookEndpoint
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	webhook.TenantID = tenantID

	if err := h.svc.CreateWebhook(r.Context(), &webhook); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Webhook criado",
		"webhook": webhook,
		"secret":  webhook.Secret,
		"warning": "Guarde o secret — use para validar assinatura HMAC das requisições",
	})
}
