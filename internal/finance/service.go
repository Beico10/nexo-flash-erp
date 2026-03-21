// Package finance implementa o DRE simplificado e Fluxo de Caixa do Nexo One.
//
// Consolida Contas a Pagar + Contas a Receber em:
//   - DRE mensal: Receita − Despesa = Resultado
//   - Fluxo de Caixa: projeção diária dos próximos 30/60/90 dias
//   - Comparativo mensal: mês atual vs mês anterior
package finance

import (
	"context"
	"time"
)

// ── TIPOS ─────────────────────────────────────────────────────────────────────

// DRELine linha do DRE.
type DRELine struct {
	Category    string  `json:"category"`
	Label       string  `json:"label"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"` // % da receita bruta
	IsPositive  bool    `json:"is_positive"`
}

// DREMonth resultado do DRE de um mês.
type DREMonth struct {
	Year        int       `json:"year"`
	Month       int       `json:"month"`
	MonthName   string    `json:"month_name"`

	// Receitas
	GrossRevenue    float64    `json:"gross_revenue"`
	RevenueLines    []DRELine  `json:"revenue_lines"`

	// Despesas
	TotalExpenses   float64    `json:"total_expenses"`
	ExpenseLines    []DRELine  `json:"expense_lines"`

	// Resultado
	NetResult       float64    `json:"net_result"`
	NetMargin       float64    `json:"net_margin"` // %
	IsProfit        bool       `json:"is_profit"`

	// Comparativo com mês anterior
	PrevMonthResult  float64   `json:"prev_month_result"`
	ResultVariation  float64   `json:"result_variation"` // %
}

// CashFlowDay projeção de um dia no fluxo de caixa.
type CashFlowDay struct {
	Date        string  `json:"date"`
	DayLabel    string  `json:"day_label"`
	Inflows     float64 `json:"inflows"`     // entradas
	Outflows    float64 `json:"outflows"`    // saídas
	Balance     float64 `json:"balance"`     // saldo do dia
	Cumulative  float64 `json:"cumulative"`  // saldo acumulado
	IsToday     bool    `json:"is_today"`
	IsPast      bool    `json:"is_past"`
}

// CashFlowSummary resumo do fluxo de caixa.
type CashFlowSummary struct {
	Period       string        `json:"period"`
	TotalInflows  float64      `json:"total_inflows"`
	TotalOutflows float64      `json:"total_outflows"`
	NetCashFlow   float64      `json:"net_cash_flow"`
	Days          []CashFlowDay `json:"days"`
	CriticalDays  []CashFlowDay `json:"critical_days"` // dias com saldo negativo
}

// FinancialTransaction transação financeira unificada.
type FinancialTransaction struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Party       string    `json:"party"`       // fornecedor ou cliente
	Amount      float64   `json:"amount"`
	FlowType    string    `json:"flow_type"`   // inflow ou outflow
	Category    string    `json:"category"`
	Status      string    `json:"status"`
}

// FinanceRepository interface para buscar dados financeiros.
type FinanceRepository interface {
	GetMonthlyRevenue(ctx context.Context, tenantID string, year, month int) ([]DRELine, error)
	GetMonthlyExpenses(ctx context.Context, tenantID string, year, month int) ([]DRELine, error)
	GetCashFlowTransactions(ctx context.Context, tenantID string, from, to time.Time) ([]FinancialTransaction, error)
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

type Service struct {
	repo FinanceRepository
}

func NewService(repo FinanceRepository) *Service {
	return &Service{repo: repo}
}

// GetDRE calcula o DRE de um mês.
func (s *Service) GetDRE(ctx context.Context, tenantID string, year, month int) (*DREMonth, error) {
	revenues, err := s.repo.GetMonthlyRevenue(ctx, tenantID, year, month)
	if err != nil {
		return nil, err
	}
	expenses, err := s.repo.GetMonthlyExpenses(ctx, tenantID, year, month)
	if err != nil {
		return nil, err
	}

	var grossRevenue, totalExpenses float64
	for _, r := range revenues {
		grossRevenue += r.Amount
	}
	for _, e := range expenses {
		totalExpenses += e.Amount
	}

	// Calcular percentuais
	if grossRevenue > 0 {
		for i := range revenues {
			revenues[i].Percentage = revenues[i].Amount / grossRevenue * 100
		}
		for i := range expenses {
			expenses[i].Percentage = expenses[i].Amount / grossRevenue * 100
		}
	}

	netResult := grossRevenue - totalExpenses
	netMargin := 0.0
	if grossRevenue > 0 {
		netMargin = netResult / grossRevenue * 100
	}

	// Mês anterior para comparativo
	prevYear, prevMonth := year, month-1
	if prevMonth == 0 {
		prevMonth = 12
		prevYear--
	}
	prevRevenues, _ := s.repo.GetMonthlyRevenue(ctx, tenantID, prevYear, prevMonth)
	prevExpenses, _ := s.repo.GetMonthlyExpenses(ctx, tenantID, prevYear, prevMonth)
	var prevGross, prevExp float64
	for _, r := range prevRevenues { prevGross += r.Amount }
	for _, e := range prevExpenses { prevExp += e.Amount }
	prevResult := prevGross - prevExp

	variation := 0.0
	if prevResult != 0 {
		variation = (netResult - prevResult) / abs(prevResult) * 100
	}

	months := []string{"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
		"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}

	return &DREMonth{
		Year: year, Month: month,
		MonthName:       months[month],
		GrossRevenue:    grossRevenue,
		RevenueLines:    revenues,
		TotalExpenses:   totalExpenses,
		ExpenseLines:    expenses,
		NetResult:       netResult,
		NetMargin:       netMargin,
		IsProfit:        netResult >= 0,
		PrevMonthResult: prevResult,
		ResultVariation: variation,
	}, nil
}

// GetCashFlow calcula o fluxo de caixa projetado.
func (s *Service) GetCashFlow(ctx context.Context, tenantID string, days int) (*CashFlowSummary, error) {
	from := time.Now().Truncate(24 * time.Hour)
	to := from.AddDate(0, 0, days)

	transactions, err := s.repo.GetCashFlowTransactions(ctx, tenantID, from, to)
	if err != nil {
		return nil, err
	}

	// Agrupar por dia
	dayMap := make(map[string]*CashFlowDay)
	today := time.Now().Format("2006-01-02")

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayMap[dateStr] = &CashFlowDay{
			Date:     dateStr,
			DayLabel: d.Format("02/01"),
			IsToday:  dateStr == today,
			IsPast:   d.Before(time.Now().Truncate(24 * time.Hour)),
		}
	}

	for _, tx := range transactions {
		dateStr := tx.Date.Format("2006-01-02")
		day, ok := dayMap[dateStr]
		if !ok {
			continue
		}
		if tx.FlowType == "inflow" {
			day.Inflows += tx.Amount
		} else {
			day.Outflows += tx.Amount
		}
		day.Balance = day.Inflows - day.Outflows
	}

	// Ordenar dias e calcular acumulado
	var allDays []CashFlowDay
	var totalInflows, totalOutflows, cumulative float64
	var criticalDays []CashFlowDay

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		day := dayMap[dateStr]
		totalInflows += day.Inflows
		totalOutflows += day.Outflows
		cumulative += day.Balance
		day.Cumulative = cumulative
		allDays = append(allDays, *day)

		if cumulative < 0 {
			criticalDays = append(criticalDays, *day)
		}
	}

	return &CashFlowSummary{
		Period:        from.Format("02/01") + " a " + to.Format("02/01/2006"),
		TotalInflows:  totalInflows,
		TotalOutflows: totalOutflows,
		NetCashFlow:   totalInflows - totalOutflows,
		Days:          allDays,
		CriticalDays:  criticalDays,
	}, nil
}

func abs(x float64) float64 {
	if x < 0 { return -x }
	return x
}
