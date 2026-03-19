// Package tax implementa o Motor Fiscal Brasil 2026.
//
// Legislacao de referencia:
//   - PEC 45/2019 (Reforma Tributaria) - IBS e CBS substituem PIS/COFINS/ICMS/ISS
//   - Lei Complementar 214/2025 - regulamentacao do IBS/CBS
//   - Decreto 11.787/2023 - tabela NCM vigente
//   - Art. 8 da LC 214/2025 - aliquota zero para cesta basica
//
// DIRETRIZ CRITICA: Nenhum valor fiscal e persistido sem aprovacao humana.
// Toda sugestao de IA passa pelo estado PENDING no banco de dados.
package tax

import (
	"context"
	"fmt"
	"math"
	"time"
	"unicode"
)

// OperationType define se a operacao e entrada (compra/credito) ou saida (venda/debito).
type OperationType string

const (
	OperationEntry OperationType = "credit_entry" // compra = gera credito
	OperationExit  OperationType = "debit_exit"   // venda = gera debito
)

// NCMRate contem as aliquotas de um codigo NCM para um periodo.
type NCMRate struct {
	NCMCode          string  `json:"ncm_code"`
	NCMDescription   string  `json:"ncm_description,omitempty"`
	IBSRate          float64 `json:"ibs_rate"`           // ex: 0.0925 = 9.25%
	CBSRate          float64 `json:"cbs_rate"`           // ex: 0.0375 = 3.75%
	SelectiveRate    float64 `json:"selective_rate"`      // Imposto Seletivo (bebidas, tabaco)
	BasketReduced    bool    `json:"basket_reduced"`      // true = cesta basica
	BasketType       string  `json:"basket_type"`         // "zero" | "reduced_60"
	TransitionYear   int     `json:"transition_year"`     // ano de referencia transicao
	TransitionFactor float64 `json:"transition_factor"`   // 2026=0.10, 2027=0.20 ... 2033=1.0
}

// RateRepository abstrai a busca de aliquotas por NCM.
type RateRepository interface {
	GetRate(ctx context.Context, ncm string, referenceDate time.Time) (*NCMRate, error)
}

// TaxInput e a entrada do motor fiscal para um calculo.
type TaxInput struct {
	TenantID      string        `json:"tenant_id"`
	NCMCode       string        `json:"ncm_code"`
	BaseValue     float64       `json:"base_value"`
	Operation     OperationType `json:"operation"`
	ReferenceDate time.Time     `json:"reference_date"`
}

// TaxResult e o resultado do calculo fiscal.
type TaxResult struct {
	NCMCode         string  `json:"ncm_code"`
	BaseValue       float64 `json:"base_value"`
	IBSRate         float64 `json:"ibs_rate"`
	CBSRate         float64 `json:"cbs_rate"`
	SelectiveRate   float64 `json:"selective_rate"`
	IBSAmount       float64 `json:"ibs_amount"`
	CBSAmount       float64 `json:"cbs_amount"`
	SelectiveAmount float64 `json:"selective_amount"`
	TotalTax        float64 `json:"total_tax"`
	CashbackAmount  float64 `json:"cashback_amount"`
	IsBasketItem    bool    `json:"is_basket_item"`
	LegalBasis      string  `json:"legal_basis"`
	ApprovalStatus  string  `json:"approval_status"`
}

// Engine e o motor fiscal principal.
type Engine struct {
	rates RateRepository
}

// NewEngine cria uma nova instancia do motor fiscal.
func NewEngine(rates RateRepository) *Engine {
	return &Engine{rates: rates}
}

// Calculate calcula IBS, CBS, Imposto Seletivo e cashback tributario.
// Retorna sugestao com ApprovalStatus = "PENDING" - nunca persiste diretamente.
func (e *Engine) Calculate(ctx context.Context, input TaxInput) (*TaxResult, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	rate, err := e.rates.GetRate(ctx, input.NCMCode, input.ReferenceDate)
	if err != nil {
		return nil, fmt.Errorf("tax: falha ao obter aliquota para NCM %s: %w", input.NCMCode, err)
	}
	if rate == nil {
		return nil, fmt.Errorf("tax: NCM %s nao encontrado", input.NCMCode)
	}

	ibsRate := rate.IBSRate
	cbsRate := rate.CBSRate
	selectiveRate := rate.SelectiveRate
	legalBasis := "LC 214/2025 - Regime padrao IBS+CBS"
	isBasket := false

	// Cesta Basica
	if rate.BasketReduced {
		isBasket = true
		switch rate.BasketType {
		case "zero":
			ibsRate = 0
			cbsRate = 0
			legalBasis = "Art. 8 LC 214/2025 - Cesta Basica Nacional (aliquota zero)"
		case "reduced_60":
			ibsRate *= 0.40 // reducao de 60%
			cbsRate *= 0.40
			legalBasis = "Art. 9 LC 214/2025 - Cesta Basica Estendida (reducao 60%)"
		}
	}

	// Fator de transicao (2026 = 10% da aliquota plena)
	if rate.TransitionFactor > 0 && rate.TransitionFactor < 1.0 {
		ibsRate *= rate.TransitionFactor
		cbsRate *= rate.TransitionFactor
		selectiveRate *= rate.TransitionFactor
		legalBasis += fmt.Sprintf(" [Transicao %d: fator %.0f%%]", rate.TransitionYear, rate.TransitionFactor*100)
	}

	ibsAmount := round2(input.BaseValue * ibsRate)
	cbsAmount := round2(input.BaseValue * cbsRate)
	selectiveAmount := round2(input.BaseValue * selectiveRate)
	totalTax := round2(ibsAmount + cbsAmount + selectiveAmount)

	// Cashback: entrada = credito positivo, saida = debito negativo
	cashback := totalTax
	if input.Operation == OperationExit {
		cashback = -totalTax
	}

	return &TaxResult{
		NCMCode:         input.NCMCode,
		BaseValue:       input.BaseValue,
		IBSRate:         ibsRate,
		CBSRate:         cbsRate,
		SelectiveRate:   selectiveRate,
		IBSAmount:       ibsAmount,
		CBSAmount:       cbsAmount,
		SelectiveAmount: selectiveAmount,
		TotalTax:        totalTax,
		CashbackAmount:  cashback,
		IsBasketItem:    isBasket,
		LegalBasis:      legalBasis,
		ApprovalStatus:  "PENDING",
	}, nil
}

func validateInput(input TaxInput) error {
	if len(input.NCMCode) != 8 {
		return fmt.Errorf("tax: NCM deve ter 8 digitos, recebido %d", len(input.NCMCode))
	}
	for _, r := range input.NCMCode {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("tax: NCM deve conter apenas digitos")
		}
	}
	if input.BaseValue < 0 {
		return fmt.Errorf("tax: BaseValue nao pode ser negativo")
	}
	return nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
