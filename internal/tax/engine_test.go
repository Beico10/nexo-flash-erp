// Package tax_test contém os testes automatizados do motor fiscal IBS/CBS 2026.
package tax_test

import (
	"context"
	"testing"
	"time"

	"github.com/nexoone/nexo-one/internal/tax"
)

// mockRateRepository implementa tax.RateRepository para testes.
type mockRateRepository struct {
	rates map[string]*tax.NCMRate
}

func (m *mockRateRepository) GetRate(_ context.Context, ncm string, _ time.Time) (*tax.NCMRate, error) {
	r, ok := m.rates[ncm]
	if !ok {
		return nil, nil
	}
	return r, nil
}

// ratesFixture retorna alíquotas de teste representando a reforma 2026.
func ratesFixture() *mockRateRepository {
	return &mockRateRepository{
		rates: map[string]*tax.NCMRate{
			// Produto industrial padrão
			"84715010": {
				NCMCode:          "84715010",
				IBSRate:          0.0925, // 9,25%
				CBSRate:          0.0375, // 3,75%
				SelectiveRate:    0,
				BasketReduced:    false,
				TransitionFactor: 1.0,
			},
			// Pão francês — Cesta Básica Nacional (alíquota zero)
			"19052000": {
				NCMCode:       "19052000",
				IBSRate:       0.0925,
				CBSRate:       0.0375,
				BasketReduced: true,
				BasketType:    "zero",
				TransitionFactor: 1.0,
			},
			// Bebida alcoólica — Imposto Seletivo
			"22030000": {
				NCMCode:       "22030000",
				IBSRate:       0.0925,
				CBSRate:       0.0375,
				SelectiveRate: 0.10, // IS 10%
				BasketReduced: false,
				TransitionFactor: 1.0,
			},
			// Produto em transição (2026 = 10% da alíquota plena)
			"61051000": {
				NCMCode:          "61051000",
				IBSRate:          0.0925,
				CBSRate:          0.0375,
				BasketReduced:    false,
				TransitionYear:   2026,
				TransitionFactor: 0.10,
			},
		},
	}
}

// TestCalculate_ProdutoIndustrial testa cálculo padrão IBS+CBS.
func TestCalculate_ProdutoIndustrial(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())
	result, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "84715010",
		BaseValue:     1000.00,
		Operation:     tax.OperationExit,
		ReferenceDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// IBS: 1000 * 9,25% = 92,50
	if result.IBSAmount != 92.50 {
		t.Errorf("IBSAmount: esperado 92.50, obtido %.2f", result.IBSAmount)
	}
	// CBS: 1000 * 3,75% = 37,50
	if result.CBSAmount != 37.50 {
		t.Errorf("CBSAmount: esperado 37.50, obtido %.2f", result.CBSAmount)
	}
	// Total: 130,00
	if result.TotalTax != 130.00 {
		t.Errorf("TotalTax: esperado 130.00, obtido %.2f", result.TotalTax)
	}
	// Saída gera débito (cashback negativo)
	if result.CashbackAmount != -130.00 {
		t.Errorf("CashbackAmount saída: esperado -130.00, obtido %.2f", result.CashbackAmount)
	}
}

// TestCalculate_CestaBasica testa alíquota zero para Cesta Básica Nacional.
func TestCalculate_CestaBasica(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())
	result, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "19052000", // pão francês
		BaseValue:     500.00,
		Operation:     tax.OperationExit,
		ReferenceDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// Cesta Básica = alíquota zero
	if result.IBSAmount != 0 {
		t.Errorf("Cesta Básica IBSAmount: esperado 0, obtido %.2f", result.IBSAmount)
	}
	if result.CBSAmount != 0 {
		t.Errorf("Cesta Básica CBSAmount: esperado 0, obtido %.2f", result.CBSAmount)
	}
	if result.TotalTax != 0 {
		t.Errorf("Cesta Básica TotalTax: esperado 0, obtido %.2f", result.TotalTax)
	}
	if !result.IsBasketItem {
		t.Error("IsBasketItem deveria ser true para pão francês")
	}
}

// TestCalculate_ImpostoSeletivo testa IS em bebida alcoólica.
func TestCalculate_ImpostoSeletivo(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())
	result, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "22030000", // cerveja
		BaseValue:     1000.00,
		Operation:     tax.OperationExit,
		ReferenceDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// IS: 1000 * 10% = 100
	if result.SelectiveAmount != 100.00 {
		t.Errorf("SelectiveAmount: esperado 100.00, obtido %.2f", result.SelectiveAmount)
	}
	// Total: IBS(92.50) + CBS(37.50) + IS(100) = 230
	if result.TotalTax != 230.00 {
		t.Errorf("TotalTax com IS: esperado 230.00, obtido %.2f", result.TotalTax)
	}
}

// TestCalculate_Transicao2026 testa fator de transição (10% da alíquota plena em 2026).
func TestCalculate_Transicao2026(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())
	result, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "61051000",
		BaseValue:     1000.00,
		Operation:     tax.OperationExit,
		ReferenceDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// IBS efetivo: 9.25% * 10% = 0.925%  → 1000 * 0.00925 = 9.25
	if result.IBSAmount != 9.25 {
		t.Errorf("Transição IBSAmount: esperado 9.25, obtido %.2f", result.IBSAmount)
	}
}

// TestCalculate_CashbackEntrada testa crédito na entrada (compra).
func TestCalculate_CashbackEntrada(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())
	result, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "84715010",
		BaseValue:     1000.00,
		Operation:     tax.OperationEntry, // compra = crédito
		ReferenceDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Entrada gera crédito positivo
	if result.CashbackAmount != 130.00 {
		t.Errorf("CashbackAmount entrada: esperado +130.00, obtido %.2f", result.CashbackAmount)
	}
}

// TestCalculate_InputInvalido testa validação de entrada.
func TestCalculate_InputInvalido(t *testing.T) {
	engine := tax.NewEngine(ratesFixture())

	// NCM com tamanho errado
	_, err := engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "123",  // inválido — deve ter 8 dígitos
		BaseValue:     100.00,
		Operation:     tax.OperationExit,
		ReferenceDate: time.Now(),
	})
	if err == nil {
		t.Error("deveria retornar erro para NCM inválido")
	}

	// Valor negativo
	_, err = engine.Calculate(context.Background(), tax.TaxInput{
		TenantID:      "tenant-001",
		NCMCode:       "84715010",
		BaseValue:     -100.00, // inválido
		Operation:     tax.OperationExit,
		ReferenceDate: time.Now(),
	})
	if err == nil {
		t.Error("deveria retornar erro para BaseValue negativo")
	}
}
