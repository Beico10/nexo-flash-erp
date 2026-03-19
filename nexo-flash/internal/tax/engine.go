// Package tax implementa o Motor Fiscal Brasil 2026.
//
// Legislação de referência:
//   - PEC 45/2019 (Reforma Tributária) — IBS e CBS substituem PIS/COFINS/ICMS/ISS
//   - Lei Complementar 214/2025 — regulamentação do IBS/CBS
//   - Decreto 11.787/2023 — tabela NCM vigente
//   - Art. 8º da LC 214/2025 — alíquota zero para cesta básica
//
// DIRETRIZ CRÍTICA: Nenhum valor fiscal é persistido sem aprovação humana.
// Toda sugestão de IA passa pelo estado PENDING no banco de dados.
package tax

import (
	"context"
	"fmt"
	"time"
)

// NCM é o código da Nomenclatura Comum do Mercosul (8 dígitos).
type NCM string

// TaxRegime define o regime tributário do tenant.
type TaxRegime string

const (
	RegimeSimplesNacional TaxRegime = "simples_nacional"
	RegimeLucroPresumido  TaxRegime = "lucro_presumido"
	RegimeLucroReal       TaxRegime = "lucro_real"
	RegimeMEI             TaxRegime = "mei"
)

// TaxCategory classifica o produto para fins de alíquota.
type TaxCategory string

const (
	CategoryBasketBasic    TaxCategory = "cesta_basica"    // Alíquota zero — Art. 8º LC 214/2025
	CategoryBasketExtended TaxCategory = "cesta_estendida" // Alíquota reduzida 60%
	CategoryMedicine       TaxCategory = "medicamento"     // Alíquota reduzida 60%
	CategoryStandard       TaxCategory = "padrao"          // Alíquota padrão IBS+CBS
	CategoryLuxury         TaxCategory = "seletivo"        // IS — Imposto Seletivo
)

// Alíquotas base — referência LC 214/2025 (ajustar conforme publicação oficial 2026)
const (
	AliquotaIBSPadrao = 0.125 // 12,5% — estimativa de transição 2026
	AliquotaCBSPadrao = 0.088 // 8,8%  — previsão a partir de 2026

	ReducaoCestaBasica    = 1.00 // 100% — alíquota zero
	ReducaoCestaEstendida = 0.60 // 60% de redução
	ReducaoMedicamento    = 0.60 // 60% de redução
)

// TaxInput é a entrada do motor fiscal para um item.
type TaxInput struct {
	NCM          NCM       `json:"ncm"`
	Description  string    `json:"description"`
	UnitValue    float64   `json:"unit_value"`
	Quantity     float64   `json:"quantity"`
	TenantRegime TaxRegime `json:"tenant_regime"`
	IsService    bool      `json:"is_service"`
	StateOrigin  string    `json:"state_origin"`
	StateDestiny string    `json:"state_destiny"`
}

// TaxResult é o resultado do cálculo fiscal.
// ApprovalStatus = "PENDING" — NUNCA persiste sem aprovação humana.
type TaxResult struct {
	NCM            NCM         `json:"ncm"`
	Category       TaxCategory `json:"category"`
	BaseValue      float64     `json:"base_value"`
	IBSRate        float64     `json:"ibs_rate"`
	CBSRate        float64     `json:"cbs_rate"`
	IBSAmount      float64     `json:"ibs_amount"`
	CBSAmount      float64     `json:"cbs_amount"`
	TotalTax       float64     `json:"total_tax"`
	CreditAmount   float64     `json:"credit_amount"` // Cashback na entrada
	DebitAmount    float64     `json:"debit_amount"`  // Débito na saída
	IsZeroRated    bool        `json:"is_zero_rated"`
	LegalBasis     string      `json:"legal_basis"`
	CalculatedAt   time.Time   `json:"calculated_at"`
	ApprovalStatus string      `json:"approval_status"` // "PENDING" | "APPROVED" | "REJECTED"
}

// NCMAliquota contém as alíquotas específicas de um código NCM.
type NCMAliquota struct {
	NCM      NCM         `json:"ncm"`
	Category TaxCategory `json:"category"`
	IBSRate  float64     `json:"ibs_rate"`
	CBSRate  float64     `json:"cbs_rate"`
	ISRate   float64     `json:"is_rate,omitempty"`
}

// NCMCachePort abstrai o cache Redis de alíquotas por NCM (TTL: 24h).
type NCMCachePort interface {
	GetAliquota(ctx context.Context, ncm NCM) (*NCMAliquota, error)
	SetAliquota(ctx context.Context, ncm NCM, a *NCMAliquota) error
}

// AuditLogPort registra operações fiscais (imutável — obrigatório por lei).
type AuditLogPort interface {
	Log(ctx context.Context, tenantID string, result TaxResult) error
}

// TaxEngine é o motor fiscal principal.
type TaxEngine struct {
	ncmCache NCMCachePort
	auditLog AuditLogPort
}

// NewTaxEngine cria uma nova instância do motor fiscal.
func NewTaxEngine(cache NCMCachePort, audit AuditLogPort) *TaxEngine {
	return &TaxEngine{ncmCache: cache, auditLog: audit}
}

// Calculate calcula IBS, CBS e cashback tributário para um item.
// Retorna sugestão com ApprovalStatus = "PENDING" — nunca persiste diretamente.
func (e *TaxEngine) Calculate(ctx context.Context, tenantID string, input TaxInput) (*TaxResult, error) {
	aliquota, err := e.ncmCache.GetAliquota(ctx, input.NCM)
	if err != nil {
		return nil, fmt.Errorf("tax: falha ao obter alíquota para NCM %s: %w", input.NCM, err)
	}

	baseValue := input.UnitValue * input.Quantity
	ibsRate, cbsRate, legalBasis := e.applyReductions(aliquota)

	ibsAmount := baseValue * ibsRate
	cbsAmount := baseValue * cbsRate
	totalTax := ibsAmount + cbsAmount

	result := &TaxResult{
		NCM:            input.NCM,
		Category:       aliquota.Category,
		BaseValue:      baseValue,
		IBSRate:        ibsRate,
		CBSRate:        cbsRate,
		IBSAmount:      round2(ibsAmount),
		CBSAmount:      round2(cbsAmount),
		TotalTax:       round2(totalTax),
		CreditAmount:   round2(totalTax), // crédito na entrada
		DebitAmount:    round2(totalTax), // débito na saída
		IsZeroRated:    aliquota.Category == CategoryBasketBasic,
		LegalBasis:     legalBasis,
		CalculatedAt:   time.Now().UTC(),
		ApprovalStatus: "PENDING", // NUNCA auto-aprovado
	}

	if err := e.auditLog.Log(ctx, tenantID, *result); err != nil {
		fmt.Printf("ALERTA AUDIT: falha ao registrar cálculo fiscal: %v\n", err)
	}

	return result, nil
}

// applyReductions retorna alíquotas efetivas após reduções legais.
func (e *TaxEngine) applyReductions(a *NCMAliquota) (ibs, cbs float64, basis string) {
	switch a.Category {
	case CategoryBasketBasic:
		return 0, 0, "Art. 8º LC 214/2025 — Cesta Básica Nacional (alíquota zero)"
	case CategoryBasketExtended:
		return a.IBSRate * (1 - ReducaoCestaEstendida),
			a.CBSRate * (1 - ReducaoCestaEstendida),
			"Art. 9º LC 214/2025 — Cesta Básica Estendida (redução 60%)"
	case CategoryMedicine:
		return a.IBSRate * (1 - ReducaoMedicamento),
			a.CBSRate * (1 - ReducaoMedicamento),
			"Art. 9º §2º LC 214/2025 — Medicamento Rename (redução 60%)"
	default:
		return a.IBSRate, a.CBSRate, "LC 214/2025 — Regime padrão IBS+CBS"
	}
}

// ContingencyMode emite nota em modo SCAN offline quando SEFAZ está indisponível.
// Referência: Nota Técnica 2013.005 ENCAT
func (e *TaxEngine) ContingencyMode(ctx context.Context, tenantID string, input TaxInput) (*TaxResult, error) {
	result, err := e.Calculate(ctx, tenantID, input)
	if err != nil {
		return nil, err
	}
	result.LegalBasis += " [CONTINGÊNCIA OFFLINE — NT ENCAT 2013.005]"
	return result, nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
