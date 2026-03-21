// Package handlers — endpoints Enterprise.
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
	// Planos
	mux.HandleFunc("GET /api/v1/enterprise/plans", h.ListPlans)
	mux.HandleFunc("GET /api/v1/enterprise/plans/{id}", h.GetPlan)
	mux.HandleFunc("POST /api/v1/enterprise/plans", h.CreatePlan)
	mux.HandleFunc("PUT /api/v1/enterprise/plans/{id}", h.UpdatePlan)
	
	// Licenças
	mux.HandleFunc("GET /api/v1/enterprise/license", h.GetLicense)
	mux.HandleFunc("POST /api/v1/enterprise/license/activate", h.ActivateLicense)
	
	// Uso
	mux.HandleFunc("GET /api/v1/enterprise/usage", h.GetUsage)
	mux.HandleFunc("GET /api/v1/enterprise/dashboard", h.GetDashboard)
}

func (h *EnterpriseHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
		"count": len(plans),
	})
}

func (h *EnterpriseHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	plan, err := h.svc.GetPlan(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Plano não encontrado")
		return
	}
	respondJSON(w, http.StatusOK, plan)
}

func (h *EnterpriseHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan enterprise.Plan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	
	if err := h.svc.CreatePlan(r.Context(), &plan); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Plano criado com sucesso",
		"plan":    plan,
	})
}

func (h *EnterpriseHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var plan enterprise.Plan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	plan.ID = id
	
	if err := h.svc.UpdatePlan(r.Context(), &plan); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Plano atualizado",
		"plan":    plan,
	})
}

func (h *EnterpriseHandler) GetLicense(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	
	license, err := h.svc.GetLicense(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Licença não encontrada")
		return
	}
	respondJSON(w, http.StatusOK, license)
}

func (h *EnterpriseHandler) ActivateLicense(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	
	var req struct {
		PlanID string `json:"plan_id"`
		Months int    `json:"months"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	
	if req.Months == 0 {
		req.Months = 12
	}
	
	license, err := h.svc.ActivateLicense(r.Context(), tenantID, req.PlanID, req.Months)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Licença ativada com sucesso",
		"license": license,
	})
}

func (h *EnterpriseHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	
	usage, err := h.svc.GetUsage(r.Context(), tenantID, month)
	if err != nil {
		// Retorna uso zerado se não encontrado
		usage = &enterprise.Usage{
			TenantID:    tenantID,
			Month:       month,
			APIRequests: 0,
			Storage:     0,
			NFesEmitted: 0,
			UsersActive: 1,
		}
	}
	respondJSON(w, http.StatusOK, usage)
}

func (h *EnterpriseHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	month := time.Now().Format("2006-01")
	
	license, _ := h.svc.GetLicense(r.Context(), tenantID)
	usage, _ := h.svc.GetUsage(r.Context(), tenantID, month)
	plans, _ := h.svc.ListPlans(r.Context())
	
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"license":       license,
		"current_usage": usage,
		"plans":         plans,
		"tenant_id":     tenantID,
	})
}
