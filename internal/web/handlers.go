// Package web — handlers HTTP para páginas web (templates).
package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nexoone/nexo-one/internal/handlers"
	"github.com/nexoone/nexo-one/pkg/middleware"
)

// PageHandler serve as páginas HTML via templates.
type PageHandler struct {
	renderer  *TemplateRenderer
	dashboard handlers.DashboardDataProvider
}

// NewPageHandler cria um novo handler de páginas.
func NewPageHandler(renderer *TemplateRenderer, dashboard handlers.DashboardDataProvider) *PageHandler {
	return &PageHandler{
		renderer:  renderer,
		dashboard: dashboard,
	}
}

// RegisterRoutes registra as rotas de páginas web.
func (h *PageHandler) RegisterRoutes(mux *http.ServeMux) {
	// Não registra mais aqui - registrado diretamente no main.go
}

// RegisterProtectedRoutes registra rotas que precisam de autenticação.
func (h *PageHandler) RegisterProtectedRoutes(mux *http.ServeMux) {
	// Não registra mais aqui - registrado diretamente no main.go
}

func (h *PageHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *PageHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title:      "Login",
		ShowLayout: false,
	}
	if err := h.renderer.Render(w, "login", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PageHandler) HandlePricing(w http.ResponseWriter, r *http.Request) {
	// TODO: implementar página de pricing
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *PageHandler) HandleSimulador(w http.ResponseWriter, r *http.Request) {
	// TODO: implementar página do simulador fiscal
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *PageHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	claims, _ := middleware.GetClaims(r.Context())
	
	// Busca estatísticas
	stats := h.dashboard.GetDashboardStats(tenantID)
	
	// Converte para view model
	statsView := h.convertStatsToView(stats)
	
	// Extrai iniciais do nome
	initials := "U"
	userName := "Usuario"
	if claims != nil {
		userName = claims.TenantSlug
		parts := strings.Split(claims.TenantSlug, "-")
		if len(parts) > 0 {
			initials = strings.ToUpper(string(parts[0][0]))
		}
	}
	
	data := TemplateData{
		Title:        "Dashboard",
		PageTitle:    "Dashboard",
		PageSubtitle: "Visao geral do negocio",
		ActivePage:   "dashboard",
		ShowLayout:   true,
		UserInitials: initials,
		UserName:     userName,
		TenantName:   tenantID,
		Stats:        statsView,
	}
	
	if err := h.renderer.Render(w, "dashboard", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PageHandler) handleModulePage(w http.ResponseWriter, r *http.Request, page, title, subtitle string) {
	claims, _ := middleware.GetClaims(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())
	
	initials := "U"
	userName := "Usuario"
	if claims != nil {
		userName = claims.TenantSlug
		parts := strings.Split(claims.TenantSlug, "-")
		if len(parts) > 0 && len(parts[0]) > 0 {
			initials = strings.ToUpper(string(parts[0][0]))
		}
	}
	
	data := TemplateData{
		Title:        title,
		PageTitle:    title,
		PageSubtitle: subtitle,
		ActivePage:   page,
		ShowLayout:   true,
		UserInitials: initials,
		UserName:     userName,
		TenantName:   tenantID,
	}
	
	// Tenta renderizar o template específico, senão redireciona para dashboard
	if err := h.renderer.Render(w, page, data); err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

func (h *PageHandler) HandleMechanic(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "mechanic", "Mecanica", "Gestao de Ordens de Servico")
}

func (h *PageHandler) HandleBakery(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "bakery", "Padaria", "Producao e Vendas")
}

func (h *PageHandler) HandleAesthetics(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "aesthetics", "Estetica", "Agendamentos e Servicos")
}

func (h *PageHandler) HandleLogistics(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "logistics", "Logistica", "Entregas e Rotas")
}

func (h *PageHandler) HandleIndustry(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "industry", "Industria PCP", "Producao e Materiais")
}

func (h *PageHandler) HandleShoes(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "shoes", "Calcados", "Grades e Comissoes")
}

func (h *PageHandler) HandleNFE(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "nfe", "Emissao NF-e", "Notas Fiscais")
}

func (h *PageHandler) HandleExpenses(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "expenses", "Despesas", "Scanner QR Code")
}

func (h *PageHandler) HandleCopilot(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())
	
	initials := "U"
	userName := "Usuario"
	if claims != nil {
		userName = claims.TenantSlug
		parts := strings.Split(claims.TenantSlug, "-")
		if len(parts) > 0 && len(parts[0]) > 0 {
			initials = strings.ToUpper(string(parts[0][0]))
		}
	}
	
	data := TemplateData{
		Title:        "Co-Piloto IA",
		PageTitle:    "Co-Piloto IA",
		PageSubtitle: "Assistente inteligente",
		ActivePage:   "copilot",
		ShowLayout:   true,
		UserInitials: initials,
		UserName:     userName,
		TenantName:   tenantID,
	}
	
	if err := h.renderer.Render(w, "copilot", data); err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

func (h *PageHandler) HandleAIApprovals(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "ai-approvals", "Aprovacoes IA", "Sugestoes Pendentes")
}

func (h *PageHandler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	h.handleModulePage(w, r, "settings", "Configuracoes", "Preferencias do Sistema")
}

func (h *PageHandler) convertStatsToView(stats handlers.DashboardStats) *DashboardStatsView {
	// Encontra o máximo para calcular porcentagens do gráfico
	maxRevenue := 1.0
	for _, d := range stats.Revenue.Chart {
		if d.Revenue > maxRevenue {
			maxRevenue = d.Revenue
		}
	}
	
	// Converte chart
	chart := make([]ChartDayView, len(stats.Revenue.Chart))
	for i, d := range stats.Revenue.Chart {
		chart[i] = ChartDayView{
			Day:            d.Day,
			Revenue:        d.Revenue,
			Tax:            d.Tax,
			RevenuePercent: (d.Revenue / maxRevenue) * 100,
			TaxPercent:     (d.Tax / maxRevenue) * 100,
		}
	}
	
	// Encontra max para módulos
	maxModule := 1
	for _, m := range stats.Revenue.ByModule {
		if m.Count > maxModule {
			maxModule = m.Count
		}
	}
	
	// Converte módulos
	modules := make([]ModuleView, len(stats.Revenue.ByModule))
	for i, m := range stats.Revenue.ByModule {
		modules[i] = ModuleView{
			Module:  m.Module,
			Count:   m.Count,
			Percent: (float64(m.Count) / float64(maxModule)) * 100,
		}
	}
	
	// Calcula porcentagens das OS
	total := stats.MechanicOS.Total
	if total == 0 {
		total = 1
	}
	
	return &DashboardStatsView{
		Revenue: RevenueView{
			TodayFmt: fmt.Sprintf("R$ %.2f", stats.Revenue.Today),
			WeekFmt:  fmt.Sprintf("R$ %.2f", stats.Revenue.Week),
			Chart:    chart,
			ByModule: modules,
		},
		MechanicOS: MechanicOSView{
			Total:         stats.MechanicOS.Total,
			Open:          stats.MechanicOS.Open,
			InProgress:    stats.MechanicOS.InProgress,
			AwaitApproval: stats.MechanicOS.AwaitApproval,
			Done:          stats.MechanicOS.Done,
			OpenCount:     stats.MechanicOS.Open + stats.MechanicOS.InProgress,
			OpenPct:       float64(stats.MechanicOS.Open) / float64(total) * 100,
			InProgressPct: float64(stats.MechanicOS.InProgress) / float64(total) * 100,
			AwaitPct:      float64(stats.MechanicOS.AwaitApproval) / float64(total) * 100,
			DonePct:       float64(stats.MechanicOS.Done) / float64(total) * 100,
		},
		BakeryProducts:     stats.BakeryProducts,
		Appointments:       stats.Appointments,
		PendingSuggestions: stats.PendingSugg,
	}
}
