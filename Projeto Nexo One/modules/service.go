// Package modules implementa o marketplace de módulos avulsos do Nexo One.
//
// Permite que clientes contratem apenas o módulo que precisam
// sem precisar do plano completo.
//
// Estratégia Land & Expand:
//   1. Cliente entra pelo módulo mais barato (ex: WhatsApp R$49/mês)
//   2. Experimenta, confia, adiciona módulos gradualmente
//   3. Eventualmente migra para plano completo
package modules

import (
	"context"
	"errors"
	"time"
)

// ── MÓDULOS DISPONÍVEIS ───────────────────────────────────────────────────────

// ModuleID identificador único de cada módulo.
type ModuleID string

const (
	ModuleRoteirizador  ModuleID = "roteirizador"
	ModuleWhatsApp      ModuleID = "whatsapp_os"
	ModulePDVPadaria    ModuleID = "pdv_padaria"
	ModuleFiscalIBSCBS  ModuleID = "fiscal_ibs_cbs"
	ModuleDespachoLote  ModuleID = "despacho_lote"
	ModuleEstoque       ModuleID = "estoque"
	ModuleCoPilotoIA    ModuleID = "copiloto_ia"
	ModuleAPIAccess     ModuleID = "api_access"
)

// Module representa um módulo do sistema.
type Module struct {
	ID          ModuleID `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Price       float64  `json:"price"`       // R$/mês módulo avulso
	PriceYearly float64  `json:"price_yearly"` // R$/mês no anual (20% desc)
	Category    string   `json:"category"`    // operacional, fiscal, logistica, ia
	ForNiches   []string `json:"for_niches"`  // nichos compatíveis (vazio = todos)
	Features    []string `json:"features"`    // lista de funcionalidades
	IsPopular   bool     `json:"is_popular"`
	TrialDays   int      `json:"trial_days"`
	VideoURL    string   `json:"video_url"`
}

// Catálogo completo de módulos avulsos.
var Catalog = map[ModuleID]*Module{
	ModuleRoteirizador: {
		ID:          ModuleRoteirizador,
		Name:        "Roteirizador Inteligente",
		Description: "Otimize rotas de entrega com algoritmo 2-opt. Economize até 30% em combustível.",
		Icon:        "🗺️",
		Price:       97.00,
		PriceYearly: 77.60,
		Category:    "logistica",
		ForNiches:   []string{"logistics", "industry", "bakery"},
		Features: []string{
			"Otimização 2-opt para 100+ paradas",
			"Mapa interativo estilo Waze",
			"Navegação passo a passo para motorista",
			"Recalculo automático por desvio",
			"Estimativa de tempo e distância real",
			"Custo: R$ 0,00 (OpenStreetMap + OSRM)",
		},
		IsPopular: true,
		TrialDays: 7,
	},
	ModuleWhatsApp: {
		ID:          ModuleWhatsApp,
		Name:        "WhatsApp de Aprovação",
		Description: "Cliente aprova orçamento direto pelo WhatsApp. Sem ligação, sem espera.",
		Icon:        "💬",
		Price:       49.00,
		PriceYearly: 39.20,
		Category:    "operacional",
		ForNiches:   []string{"mechanic", "aesthetics"},
		Features: []string{
			"Link de aprovação automático por OS",
			"Cliente aprova ou rejeita pelo celular",
			"Notificação imediata ao mecânico",
			"Histórico de aprovações",
			"Sem limite de mensagens",
			"Integrado com Evolution API",
		},
		IsPopular: true,
		TrialDays: 14,
	},
	ModulePDVPadaria: {
		ID:          ModulePDVPadaria,
		Name:        "PDV para Padaria",
		Description: "Ponto de venda com integração de balança Toledo/Elgin e código PLU.",
		Icon:        "🍞",
		Price:       79.00,
		PriceYearly: 63.20,
		Category:    "operacional",
		ForNiches:   []string{"bakery"},
		Features: []string{
			"Integração balança Toledo e Elgin",
			"Código PLU para produtos por peso",
			"Frente de caixa simplificada",
			"Controle de perdas e quebras",
			"Relatório de vendas por produto",
			"Fechamento de caixa diário",
		},
		IsPopular: false,
		TrialDays: 7,
	},
	ModuleFiscalIBSCBS: {
		ID:          ModuleFiscalIBSCBS,
		Name:        "Motor Fiscal IBS/CBS 2026",
		Description: "Único motor fiscal 100% alinhado com a Reforma Tributária. Não pague imposto duas vezes.",
		Icon:        "⚡",
		Price:       67.00,
		PriceYearly: 53.60,
		Category:    "fiscal",
		ForNiches:   []string{},
		Features: []string{
			"Cálculo automático IBS + CBS por NCM",
			"Crédito fiscal nas entradas (NF-e)",
			"Relatório de créditos acumulados",
			"Simulador de impacto da Reforma",
			"Alíquotas atualizadas em tempo real",
			"Exportação para contador (SPED)",
		},
		IsPopular: true,
		TrialDays: 30,
	},
	ModuleDespachoLote: {
		ID:          ModuleDespachoLote,
		Name:        "Despacho em Lote",
		Description: "Importe 500 notas de uma vez. XML, CSV, EDI. Distribua por veículo automaticamente.",
		Icon:        "🚛",
		Price:       147.00,
		PriceYearly: 117.60,
		Category:    "logistica",
		ForNiches:   []string{"logistics", "industry"},
		Features: []string{
			"Importação XML NF-e em massa",
			"Suporte CSV, EDI ANSI X12 e PDF",
			"Scanner de código de barras avulso",
			"Distribuição automática por veículo",
			"Critério: peso + cubagem + região",
			"Integração direta com roteirizador",
		},
		IsPopular: false,
		TrialDays: 7,
	},
	ModuleEstoque: {
		ID:          ModuleEstoque,
		Name:        "Estoque Inteligente",
		Description: "Custo Médio Ponderado automático. Alertas de mínimo via WhatsApp. Para qualquer nicho.",
		Icon:        "📦",
		Price:       89.00,
		PriceYearly: 71.20,
		Category:    "operacional",
		ForNiches:   []string{},
		Features: []string{
			"Custo Médio Ponderado automático",
			"Entrada via QR Code de NF-e",
			"Alertas de estoque mínimo WhatsApp",
			"Campos específicos por nicho",
			"Relatório de giro de estoque",
			"Integração com entrada de NF-e",
		},
		IsPopular: true,
		TrialDays: 7,
	},
}

// ── TIPOS ─────────────────────────────────────────────────────────────────────

// ModuleSubscription assinatura de um módulo avulso.
type ModuleSubscription struct {
	ID         string
	TenantID   string
	ModuleID   ModuleID
	Status     string    // trialing, active, cancelled
	TrialEndsAt *time.Time
	RenewsAt   time.Time
	Price      float64
	Cycle      string    // monthly, yearly
	CreatedAt  time.Time
}

// TenantModules módulos ativos de um tenant.
type TenantModules struct {
	TenantID      string
	PlanModules   []ModuleID            // módulos do plano base
	AddonModules  []ModuleSubscription  // módulos avulsos adicionais
	AllModules    []ModuleID            // união de tudo
}

// ── ERROS ─────────────────────────────────────────────────────────────────────

var (
	ErrModuleNotFound    = errors.New("módulo não encontrado")
	ErrAlreadySubscribed = errors.New("módulo já contratado")
	ErrModuleIncompatible = errors.New("módulo incompatível com seu nicho")
)

// ── REPOSITÓRIO ───────────────────────────────────────────────────────────────

type Repository interface {
	GetTenantModules(ctx context.Context, tenantID string) (*TenantModules, error)
	AddModule(ctx context.Context, sub *ModuleSubscription) error
	CancelModule(ctx context.Context, tenantID string, moduleID ModuleID) error
	ListModuleSubscriptions(ctx context.Context, tenantID string) ([]*ModuleSubscription, error)
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetCatalog retorna todos os módulos disponíveis.
func (s *Service) GetCatalog(businessType string) []*Module {
	var result []*Module
	for _, m := range Catalog {
		// Módulo universal ou compatível com o nicho
		if len(m.ForNiches) == 0 {
			result = append(result, m)
			continue
		}
		for _, n := range m.ForNiches {
			if n == businessType {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

// GetModule retorna um módulo específico.
func (s *Service) GetModule(id ModuleID) (*Module, error) {
	m, ok := Catalog[id]
	if !ok {
		return nil, ErrModuleNotFound
	}
	return m, nil
}

// Subscribe ativa um módulo avulso para um tenant.
func (s *Service) Subscribe(ctx context.Context, tenantID string, moduleID ModuleID, cycle string) (*ModuleSubscription, error) {
	module, err := s.GetModule(moduleID)
	if err != nil {
		return nil, err
	}

	price := module.Price
	if cycle == "yearly" {
		price = module.PriceYearly
	}

	now := time.Now()
	trialEnd := now.AddDate(0, 0, module.TrialDays)

	sub := &ModuleSubscription{
		TenantID:    tenantID,
		ModuleID:    moduleID,
		Status:      "trialing",
		TrialEndsAt: &trialEnd,
		RenewsAt:    trialEnd,
		Price:       price,
		Cycle:       cycle,
		CreatedAt:   now,
	}

	if err := s.repo.AddModule(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// HasModule verifica se tenant tem acesso a um módulo.
func (s *Service) HasModule(ctx context.Context, tenantID string, moduleID ModuleID) bool {
	modules, err := s.repo.GetTenantModules(ctx, tenantID)
	if err != nil {
		return false
	}
	for _, m := range modules.AllModules {
		if m == moduleID {
			return true
		}
	}
	return false
}

// GetTenantModules retorna todos os módulos ativos de um tenant.
func (s *Service) GetTenantModules(ctx context.Context, tenantID string) (*TenantModules, error) {
	return s.repo.GetTenantModules(ctx, tenantID)
}

// CancelModule cancela um módulo avulso.
func (s *Service) CancelModule(ctx context.Context, tenantID string, moduleID ModuleID) error {
	return s.repo.CancelModule(ctx, tenantID, moduleID)
}

// CalculateMonthlyCost calcula o custo total dos módulos avulsos.
func (s *Service) CalculateMonthlyCost(subs []*ModuleSubscription) float64 {
	total := 0.0
	for _, sub := range subs {
		if sub.Status == "active" || sub.Status == "trialing" {
			total += sub.Price
		}
	}
	return total
}

// ModulesIncludedInPlan retorna quais módulos já estão incluídos em cada plano.
func ModulesIncludedInPlan(planCode string) []ModuleID {
	plans := map[string][]ModuleID{
		"starter": {
			ModuleFiscalIBSCBS,
			ModuleWhatsApp,
		},
		"pro": {
			ModuleFiscalIBSCBS,
			ModuleWhatsApp,
			ModuleEstoque,
			ModuleCoPilotoIA,
			ModuleRoteirizador,
		},
		"business": {
			ModuleFiscalIBSCBS,
			ModuleWhatsApp,
			ModuleEstoque,
			ModuleCoPilotoIA,
			ModuleRoteirizador,
			ModuleDespachoLote,
			ModuleAPIAccess,
		},
		"enterprise": {
			ModuleFiscalIBSCBS,
			ModuleWhatsApp,
			ModuleEstoque,
			ModuleCoPilotoIA,
			ModuleRoteirizador,
			ModuleDespachoLote,
			ModuleAPIAccess,
			ModulePDVPadaria,
		},
	}
	if modules, ok := plans[planCode]; ok {
		return modules
	}
	return []ModuleID{}
}
