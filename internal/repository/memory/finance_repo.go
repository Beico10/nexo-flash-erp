// Package memory — repositório in-memory de dados financeiros.
package memory

import (
	"context"
	"time"

	"github.com/nexoone/nexo-one/internal/finance"
)

type FinanceRepo struct{}

func NewFinanceRepo() *FinanceRepo { return &FinanceRepo{} }

func (r *FinanceRepo) GetMonthlyRevenue(ctx context.Context, tenantID string, year, month int) ([]finance.DRELine, error) {
	// Demo: receitas simuladas por categoria
	return []finance.DRELine{
		{Category: "os_servicos", Label: "Ordens de Serviço", Amount: 8500.00, IsPositive: true},
		{Category: "pecas",       Label: "Venda de Peças",    Amount: 2300.00, IsPositive: true},
		{Category: "mensalidade", Label: "Contratos Mensais", Amount: 1200.00, IsPositive: true},
	}, nil
}

func (r *FinanceRepo) GetMonthlyExpenses(ctx context.Context, tenantID string, year, month int) ([]finance.DRELine, error) {
	return []finance.DRELine{
		{Category: "aluguel",    Label: "Aluguel",          Amount: 3500.00, IsPositive: false},
		{Category: "fornecedor", Label: "Peças/Fornecedores", Amount: 2100.00, IsPositive: false},
		{Category: "folha",      Label: "Folha de Pagamento", Amount: 2800.00, IsPositive: false},
		{Category: "imposto",    Label: "Impostos",          Amount: 780.00,  IsPositive: false},
		{Category: "servico",    Label: "Serviços/Utilities", Amount: 450.00,  IsPositive: false},
	}, nil
}

func (r *FinanceRepo) GetCashFlowTransactions(ctx context.Context, tenantID string, from, to time.Time) ([]finance.FinancialTransaction, error) {
	now := time.Now()
	return []finance.FinancialTransaction{
		{Date: now.AddDate(0, 0, 1),  Description: "OS #1042 — João Silva",    Party: "João Silva",         Amount: 850.00,  FlowType: "inflow",  Category: "servico",   Status: "pending"},
		{Date: now.AddDate(0, 0, 2),  Description: "Aluguel do galpão",         Party: "Imobiliária Central", Amount: 3500.00, FlowType: "outflow", Category: "aluguel",   Status: "pending"},
		{Date: now.AddDate(0, 0, 3),  Description: "SIMPLES Nacional",           Party: "Receita Federal",    Amount: 780.00,  FlowType: "outflow", Category: "imposto",   Status: "pending"},
		{Date: now.AddDate(0, 0, 5),  Description: "OS #1043 — Maria Oliveira", Party: "Maria Oliveira",     Amount: 420.00,  FlowType: "inflow",  Category: "servico",   Status: "pending"},
		{Date: now.AddDate(0, 0, 7),  Description: "Fornecedor Peças Silva",    Party: "Auto Peças Silva",   Amount: 1850.00, FlowType: "outflow", Category: "fornecedor", Status: "pending"},
		{Date: now.AddDate(0, 0, 10), Description: "Mensalidade Frotas ABC",    Party: "Transportadora ABC", Amount: 1200.00, FlowType: "inflow",  Category: "mensalidade", Status: "pending"},
		{Date: now.AddDate(0, 0, 12), Description: "Folha de Pagamento",        Party: "Funcionários",       Amount: 2800.00, FlowType: "outflow", Category: "folha",     Status: "pending"},
		{Date: now.AddDate(0, 0, 15), Description: "OS #1050 — Carlos Souza",   Party: "Carlos Souza",       Amount: 650.00,  FlowType: "inflow",  Category: "servico",   Status: "pending"},
		{Date: now.AddDate(0, 0, 20), Description: "OS #1055 — Lote 5 clientes", Party: "Clientes Diversos", Amount: 3200.00, FlowType: "inflow",  Category: "servico",   Status: "pending"},
		{Date: now.AddDate(0, 0, 25), Description: "Conta de Energia",          Party: "CPFL Energia",       Amount: 450.00,  FlowType: "outflow", Category: "servico",   Status: "pending"},
	}, nil
}
