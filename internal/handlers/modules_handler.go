// Package handlers — endpoints do marketplace de módulos.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nexoone/nexo-one/internal/modules"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

type ModulesHandler struct {
	svc *modules.Service
}

func NewModulesHandler(svc *modules.Service) *ModulesHandler {
	return &ModulesHandler{svc: svc}
}

func (h *ModulesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/modules/catalog", h.GetCatalog)
	mux.HandleFunc("GET /api/v1/modules/subscriptions", h.ListSubscriptions)
	mux.HandleFunc("POST /api/v1/modules/subscribe", h.Subscribe)
	mux.HandleFunc("POST /api/v1/modules/{id}/cancel", h.Cancel)
	mux.HandleFunc("GET /api/v1/modules/active", h.ActiveModules)
}

func (h *ModulesHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	businessType := "mechanic"
	if claims != nil {
		businessType = claims.BusinessType
	}

	catalog := h.svc.GetCatalog(businessType)

	formatted := make([]map[string]interface{}, len(catalog))
	for i, m := range catalog {
		formatted[i] = map[string]interface{}{
			"id":           m.ID,
			"name":         m.Name,
			"description":  m.Description,
			"icon":         m.Icon,
			"price":        m.Price,
			"price_yearly": m.PriceYearly,
			"category":     m.Category,
			"for_niches":   m.ForNiches,
			"features":     m.Features,
			"is_popular":   m.IsPopular,
			"trial_days":   m.TrialDays,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"modules": formatted,
		"count":   len(formatted),
	})
}

func (h *ModulesHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	tm, err := h.svc.GetTenantModules(r.Context(), tenantID)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"subscriptions": []interface{}{}})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": tm.AddonModules,
		"plan_modules":  tm.PlanModules,
		"all_modules":   tm.AllModules,
	})
}

func (h *ModulesHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	var req struct {
		ModuleID string `json:"module_id"`
		Cycle    string `json:"cycle"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Cycle == "" {
		req.Cycle = "monthly"
	}

	sub, err := h.svc.Subscribe(r.Context(), tenantID, modules.ModuleID(req.ModuleID), req.Cycle)
	if err != nil {
		switch err {
		case modules.ErrModuleNotFound:
			respondError(w, http.StatusNotFound, "Módulo não encontrado")
		case modules.ErrAlreadySubscribed:
			respondError(w, http.StatusConflict, "Módulo já contratado")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	module, _ := h.svc.GetModule(modules.ModuleID(req.ModuleID))

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":      "Módulo ativado! Trial iniciado.",
		"subscription": sub,
		"trial_days":   module.TrialDays,
		"module":       module,
	})
}

func (h *ModulesHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	moduleID := r.PathValue("id")

	if err := h.svc.CancelModule(r.Context(), tenantID, modules.ModuleID(moduleID)); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Módulo cancelado. Acesso até o final do período.",
	})
}

func (h *ModulesHandler) ActiveModules(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	tm, err := h.svc.GetTenantModules(r.Context(), tenantID)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"modules": []interface{}{}})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"modules": tm.AllModules,
		"count":   len(tm.AllModules),
	})
}
