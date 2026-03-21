// Package finance - DRE e Fluxo de Caixa do Nexo One ERP
package finance

import (
	"context"
	"time"
)

// DRELine representa uma linha do DRE
type DRELine struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Percentage  float64 `json:"percentage"`
}

// DRE representa o Demonstrativo de Resultado do Exercício
type DRE struct {
	TenantID       string    `json:"tenant_id"`
	Period         string    `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	
	// Receitas
	GrossRevenue   float64   `json:"gross_revenue"`
	Deductions     float64   `json:"deductions"`
	NetRevenue     float64   `json:"net_revenue"`
	
	// Custos
	COGS           float64   `json:"cogs"` // Custo dos Produtos Vendidos
	GrossProfit    float64   `json:"gross_profit"`
	GrossMargin    float64   `json:"gross_margin_pct"`
	
	// Despesas Operacionais
	OperatingExpenses float64 `json:"operating_expenses"`
	AdminExpenses     float64 `json:"admin_expenses"`
	SalesExpenses     float64 `json:"sales_expenses"`
	
	// Resultado
	EBITDA         float64   `json:"ebitda"`
	Depreciation   float64   `json:"depreciation"`
	EBIT           float64   `json:"ebit"`
	FinancialResult float64  `json:"financial_result"`
	EBT            float64   `json:"ebt"`
	Taxes          float64   `json:"taxes"`
	NetIncome      float64   `json:"net_income"`
	NetMargin      float64   `json:"net_margin_pct"`
	
	Lines          []DRELine `json:"lines"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// CashFlowEntry representa uma entrada no fluxo de caixa
type CashFlowEntry struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Type        string    `json:"type"` // inflow, outflow
	Amount      float64   `json:"amount"`
	Balance     float64   `json:"balance"`
}

// CashFlow representa o fluxo de caixa
type CashFlow struct {
	TenantID      string          `json:"tenant_id"`
	Period        string          `json:"period"`
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	
	OpeningBalance float64        `json:"opening_balance"`
	TotalInflows   float64        `json:"total_inflows"`
	TotalOutflows  float64        `json:"total_outflows"`
	NetCashFlow    float64        `json:"net_cash_flow"`
	ClosingBalance float64        `json:"closing_balance"`
	
	Entries        []CashFlowEntry `json:"entries"`
	GeneratedAt    time.Time       `json:"generated_at"`
}

// CashFlowProjection projeção de fluxo de caixa
type CashFlowProjection struct {
	TenantID        string    `json:"tenant_id"`
	ProjectionDays  int       `json:"projection_days"`
	CurrentBalance  float64   `json:"current_balance"`
	ExpectedInflows float64   `json:"expected_inflows"`
	ExpectedOutflows float64  `json:"expected_outflows"`
	ProjectedBalance float64  `json:"projected_balance"`
	
	// Alertas
	NegativeBalanceDate *time.Time `json:"negative_balance_date,omitempty"`
	LowBalanceAlert     bool       `json:"low_balance_alert"`
	
	DailyProjection []DailyBalance `json:"daily_projection"`
	GeneratedAt     time.Time      `json:"generated_at"`
}

// DailyBalance balanço diário projetado
type DailyBalance struct {
	Date     time.Time `json:"date"`
	Inflows  float64   `json:"inflows"`
	Outflows float64   `json:"outflows"`
	Balance  float64   `json:"balance"`
}

// DataProvider interface para obter dados financeiros
type DataProvider interface {
	GetRevenues(ctx context.Context, tenantID string, start, end time.Time) (float64, error)
	GetExpenses(ctx context.Context, tenantID string, start, end time.Time) (float64, error)
	GetPayables(ctx context.Context, tenantID string, start, end time.Time) ([]CashFlowEntry, error)
	GetReceivables(ctx context.Context, tenantID string, start, end time.Time) ([]CashFlowEntry, error)
	GetCurrentBalance(ctx context.Context, tenantID string) (float64, error)
}

// Service para finanças
type Service struct {
	provider DataProvider
}

// NewService cria um novo service de finanças
func NewService(provider DataProvider) *Service {
	return &Service{provider: provider}
}

// GenerateDRE gera o DRE para um período
func (s *Service) GenerateDRE(ctx context.Context, tenantID string, start, end time.Time) (*DRE, error) {
	revenues, _ := s.provider.GetRevenues(ctx, tenantID, start, end)
	expenses, _ := s.provider.GetExpenses(ctx, tenantID, start, end)
	
	// Valores simulados para demo
	deductions := revenues * 0.0925 // PIS/COFINS
	netRevenue := revenues - deductions
	cogs := revenues * 0.35
	grossProfit := netRevenue - cogs
	grossMargin := 0.0
	if netRevenue > 0 {
		grossMargin = (grossProfit / netRevenue) * 100
	}
	
	operatingExpenses := expenses * 0.6
	adminExpenses := expenses * 0.25
	salesExpenses := expenses * 0.15
	
	ebitda := grossProfit - operatingExpenses - adminExpenses - salesExpenses
	depreciation := revenues * 0.02
	ebit := ebitda - depreciation
	financialResult := -revenues * 0.01
	ebt := ebit + financialResult
	taxes := ebt * 0.34
	if taxes < 0 {
		taxes = 0
	}
	netIncome := ebt - taxes
	netMargin := 0.0
	if netRevenue > 0 {
		netMargin = (netIncome / netRevenue) * 100
	}
	
	return &DRE{
		TenantID:          tenantID,
		Period:            start.Format("2006-01") + " a " + end.Format("2006-01"),
		StartDate:         start,
		EndDate:           end,
		GrossRevenue:      revenues,
		Deductions:        deductions,
		NetRevenue:        netRevenue,
		COGS:              cogs,
		GrossProfit:       grossProfit,
		GrossMargin:       grossMargin,
		OperatingExpenses: operatingExpenses,
		AdminExpenses:     adminExpenses,
		SalesExpenses:     salesExpenses,
		EBITDA:            ebitda,
		Depreciation:      depreciation,
		EBIT:              ebit,
		FinancialResult:   financialResult,
		EBT:               ebt,
		Taxes:             taxes,
		NetIncome:         netIncome,
		NetMargin:         netMargin,
		GeneratedAt:       time.Now(),
	}, nil
}

// GenerateCashFlow gera o fluxo de caixa para um período
func (s *Service) GenerateCashFlow(ctx context.Context, tenantID string, start, end time.Time) (*CashFlow, error) {
	payables, _ := s.provider.GetPayables(ctx, tenantID, start, end)
	receivables, _ := s.provider.GetReceivables(ctx, tenantID, start, end)
	currentBalance, _ := s.provider.GetCurrentBalance(ctx, tenantID)
	
	var entries []CashFlowEntry
	totalInflows := 0.0
	totalOutflows := 0.0
	
	for _, r := range receivables {
		r.Type = "inflow"
		entries = append(entries, r)
		totalInflows += r.Amount
	}
	
	for _, p := range payables {
		p.Type = "outflow"
		entries = append(entries, p)
		totalOutflows += p.Amount
	}
	
	return &CashFlow{
		TenantID:       tenantID,
		Period:         start.Format("02/01/2006") + " a " + end.Format("02/01/2006"),
		StartDate:      start,
		EndDate:        end,
		OpeningBalance: currentBalance,
		TotalInflows:   totalInflows,
		TotalOutflows:  totalOutflows,
		NetCashFlow:    totalInflows - totalOutflows,
		ClosingBalance: currentBalance + totalInflows - totalOutflows,
		Entries:        entries,
		GeneratedAt:    time.Now(),
	}, nil
}

// ProjectCashFlow projeta o fluxo de caixa para os próximos dias
func (s *Service) ProjectCashFlow(ctx context.Context, tenantID string, days int) (*CashFlowProjection, error) {
	currentBalance, _ := s.provider.GetCurrentBalance(ctx, tenantID)
	
	start := time.Now()
	end := start.AddDate(0, 0, days)
	
	payables, _ := s.provider.GetPayables(ctx, tenantID, start, end)
	receivables, _ := s.provider.GetReceivables(ctx, tenantID, start, end)
	
	expectedInflows := 0.0
	expectedOutflows := 0.0
	
	for _, r := range receivables {
		expectedInflows += r.Amount
	}
	for _, p := range payables {
		expectedOutflows += p.Amount
	}
	
	projectedBalance := currentBalance + expectedInflows - expectedOutflows
	
	var negativeDate *time.Time
	lowAlert := projectedBalance < currentBalance*0.2
	
	if projectedBalance < 0 {
		t := time.Now().AddDate(0, 0, days/2)
		negativeDate = &t
	}
	
	return &CashFlowProjection{
		TenantID:            tenantID,
		ProjectionDays:      days,
		CurrentBalance:      currentBalance,
		ExpectedInflows:     expectedInflows,
		ExpectedOutflows:    expectedOutflows,
		ProjectedBalance:    projectedBalance,
		NegativeBalanceDate: negativeDate,
		LowBalanceAlert:     lowAlert,
		GeneratedAt:         time.Now(),
	}, nil
}
