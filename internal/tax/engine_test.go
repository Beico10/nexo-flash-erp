package tax

import (
	"context"
	"testing"
	"time"
)

// ratesFixture retorna um repositorio in-memory com aliquotas de teste.
type inMemoryRateRepo struct {
	rates map[string]*NCMRate
}

func (r *inMemoryRateRepo) GetRate(_ context.Context, ncm string, _ time.Time) (*NCMRate, error) {
	rate, ok := r.rates[ncm]
	if !ok {
		return nil, nil
	}
	return rate, nil
}

func ratesFixture() RateRepository {
	return &inMemoryRateRepo{
		rates: map[string]*NCMRate{
			// Cesta basica zero
			"19052000": {NCMCode: "19052000", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
			// Cesta estendida (reducao 60%)
			"17019900": {NCMCode: "17019900", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "reduced_60", TransitionFactor: 1.0},
			// Padrao
			"84715010": {NCMCode: "84715010", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
			// Seletivo
			"22030000": {NCMCode: "22030000", IBSRate: 0.0925, CBSRate: 0.0375, SelectiveRate: 0.10, TransitionFactor: 1.0},
			// Transicao 2026
			"61051000": {NCMCode: "61051000", IBSRate: 0.0925, CBSRate: 0.0375, TransitionYear: 2026, TransitionFactor: 0.10},
		},
	}
}

func TestCalculate_CestaBasicaZero(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "19052000", BaseValue: 100, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalTax != 0 {
		t.Errorf("cesta basica zero: esperado 0, obteve %f", result.TotalTax)
	}
	if !result.IsBasketItem {
		t.Error("deveria ser item de cesta basica")
	}
	if result.ApprovalStatus != "PENDING" {
		t.Error("deveria retornar PENDING")
	}
}

func TestCalculate_AliquotaPadrao(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "84715010", BaseValue: 1000, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IBSAmount != 92.5 {
		t.Errorf("IBS esperado 92.50, obteve %.2f", result.IBSAmount)
	}
	if result.CBSAmount != 37.5 {
		t.Errorf("CBS esperado 37.50, obteve %.2f", result.CBSAmount)
	}
	if result.TotalTax != 130 {
		t.Errorf("Total esperado 130.00, obteve %.2f", result.TotalTax)
	}
}

func TestCalculate_ImpostoSeletivo(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "22030000", BaseValue: 200, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SelectiveAmount != 20 {
		t.Errorf("Seletivo esperado 20.00, obteve %.2f", result.SelectiveAmount)
	}
	if result.TotalTax != 46 {
		t.Errorf("Total esperado 46.00, obteve %.2f", result.TotalTax)
	}
}

func TestCalculate_CestaEstendida60(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "17019900", BaseValue: 100, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IBSAmount != 3.7 {
		t.Errorf("IBS (40%% de 9.25%%) esperado 3.70, obteve %.2f", result.IBSAmount)
	}
	if result.CBSAmount != 1.5 {
		t.Errorf("CBS (40%% de 3.75%%) esperado 1.50, obteve %.2f", result.CBSAmount)
	}
	if !result.IsBasketItem {
		t.Error("deveria ser item de cesta basica")
	}
}

func TestCalculate_Transicao2026(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "61051000", BaseValue: 1000, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Fator 10%: IBS = 1000 * 0.0925 * 0.10 = 9.25
	if result.IBSAmount != 9.25 {
		t.Errorf("IBS transicao esperado 9.25, obteve %.2f", result.IBSAmount)
	}
	// CBS = 1000 * 0.0375 * 0.10 = 3.75
	if result.CBSAmount != 3.75 {
		t.Errorf("CBS transicao esperado 3.75, obteve %.2f", result.CBSAmount)
	}
}

func TestCalculate_CashbackEntrada(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "84715010", BaseValue: 1000, Operation: OperationEntry,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CashbackAmount != 130 {
		t.Errorf("cashback entrada esperado 130.00, obteve %.2f", result.CashbackAmount)
	}
}

func TestCalculate_CashbackSaida(t *testing.T) {
	e := NewEngine(ratesFixture())
	result, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "84715010", BaseValue: 1000, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CashbackAmount != -130 {
		t.Errorf("cashback saida esperado -130.00, obteve %.2f", result.CashbackAmount)
	}
}

func TestCalculate_NCMInvalido(t *testing.T) {
	e := NewEngine(ratesFixture())
	_, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "123", BaseValue: 100, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Error("deveria retornar erro para NCM invalido")
	}
}

func TestCalculate_BaseValueNegativo(t *testing.T) {
	e := NewEngine(ratesFixture())
	_, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "84715010", BaseValue: -100, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Error("deveria retornar erro para BaseValue negativo")
	}
}

func TestCalculate_NCMNaoEncontrado(t *testing.T) {
	e := NewEngine(ratesFixture())
	_, err := e.Calculate(context.Background(), TaxInput{
		TenantID: "t1", NCMCode: "99999999", BaseValue: 100, Operation: OperationExit,
		ReferenceDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Error("deveria retornar erro para NCM nao encontrado")
	}
}
