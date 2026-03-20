package handlers

import (
	"net/http"

	"github.com/nexoone/nexo-one/pkg/middleware"
)

// DashboardStats agrega KPIs de todos os modulos do tenant.
type DashboardStats struct {
	MechanicOS      ModuleStats      `json:"mechanic_os"`
	BakeryProducts  int              `json:"bakery_products"`
	Appointments    int              `json:"appointments"`
	PendingSugg     int              `json:"pending_suggestions"`
	Revenue         DashboardRevenue `json:"revenue"`
}

type ModuleStats struct {
	Total     int `json:"total"`
	Open      int `json:"open"`
	InProgress int `json:"in_progress"`
	AwaitApproval int `json:"await_approval"`
	Done      int `json:"done"`
}

type DashboardRevenue struct {
	Today    float64           `json:"today"`
	Week     float64           `json:"week"`
	Chart    []DashboardDay    `json:"chart"`
	ByModule []ModuleActivity  `json:"by_module"`
}

type DashboardDay struct {
	Day     string  `json:"day"`
	Revenue float64 `json:"revenue"`
	Tax     float64 `json:"tax"`
}

type ModuleActivity struct {
	Module string `json:"module"`
	Count  int    `json:"count"`
}

type DashboardDataProvider interface {
	GetDashboardStats(tenantID string) DashboardStats
}

type DashboardHandler struct {
	provider DashboardDataProvider
}

func NewDashboardHandler(p DashboardDataProvider) *DashboardHandler {
	return &DashboardHandler{provider: p}
}

func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/dashboard/stats", h.GetStats)
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant nao identificado")
		return
	}
	stats := h.provider.GetDashboardStats(tenantID)
	respondJSON(w, http.StatusOK, stats)
}
