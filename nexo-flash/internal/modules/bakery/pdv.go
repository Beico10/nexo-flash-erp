// Package bakery implementa o módulo de Padaria do Nexo Flash.
//
// Funcionalidades (Briefing Mestre §1 — Nicho Padaria):
//   - PDV Rápido (venda por peso, unidade e combo)
//   - Integração com Balanças (protocolo Toledo/Elgin via serial/TCP)
//   - Gestão de Perdas (vencimento, descarte, produção excedente)
//   - Alíquota Zero/Reduzida automática para Cesta Básica Nacional
package bakery

import (
	"context"
	"fmt"
	"time"
)

// SaleType define como o item é vendido no PDV.
type SaleType string

const (
	SaleByWeight SaleType = "weight" // balança (kg)
	SaleByUnit   SaleType = "unit"   // unidade
	SaleByCombo  SaleType = "combo"  // combo fixo
)

// LossReason define os motivos de perda de produção.
type LossReason string

const (
	LossExpired    LossReason = "expired"    // venceu
	LossDiscarded  LossReason = "discarded"  // descartado (qualidade)
	LossOverprod   LossReason = "overprod"   // sobra de produção
	LossDamaged    LossReason = "damaged"    // danificado
)

// BakeryProduct representa um produto da padaria com integração fiscal.
type BakeryProduct struct {
	ID              string
	TenantID        string
	SKU             string
	Name            string
	SaleType        SaleType
	UnitPrice       float64    // preço por kg OU por unidade
	NCMCode         string     // para cálculo IBS/CBS
	IsBasketItem    bool       // true = Cesta Básica → alíquota zero/reduzida
	BasketCategory  string     // ex: "pao_frances", "leite", "arroz"
	ScaleCode       string     // código para a balança (PLU)
	CurrentStock    float64    // kg ou unidades
	MinStock        float64    // alerta de estoque mínimo
	Active          bool
}

// PDVSaleItem representa um item numa venda do PDV.
type PDVSaleItem struct {
	ProductID   string
	ProductName string
	SaleType    SaleType
	Quantity    float64  // kg se weight, unidades se unit
	UnitPrice   float64
	Discount    float64
	TotalPrice  float64
	// Fiscal — calculado automaticamente pelo motor fiscal
	IBSAmount   float64
	CBSAmount   float64
	IsBasket    bool
}

// PDVSale representa uma venda completa no PDV da padaria.
type PDVSale struct {
	ID          string
	TenantID    string
	Number      string
	Items       []PDVSaleItem
	Subtotal    float64
	Discount    float64
	TotalAmount float64
	TotalTax    float64
	PaymentMethod string // "pix"|"cash"|"credit"|"debit"
	SoldAt      time.Time
	OperatorID  string
}

// ScaleReading representa uma leitura da balança.
type ScaleReading struct {
	ScaleID    string
	PLUCode    string  // código do produto na balança
	WeightKG   float64
	ReadAt     time.Time
	IsStable   bool    // peso estabilizado
}

// LossRecord registra uma perda de produção.
type LossRecord struct {
	ID          string
	TenantID    string
	ProductID   string
	ProductName string
	Quantity    float64
	Unit        string // "kg" | "unidade"
	Reason      LossReason
	CostValue   float64 // custo da perda
	Notes       string
	RecordedAt  time.Time
	RecordedBy  string
}

// ScaleReader é a interface para comunicação com balanças.
// Implementações: ToledoPrix, ElginDP, SerialPort
type ScaleReader interface {
	ReadWeight(ctx context.Context, scaleID string) (*ScaleReading, error)
	IsConnected(ctx context.Context, scaleID string) bool
}

// BakeryRepository é o contrato de persistência do módulo.
type BakeryRepository interface {
	GetProduct(ctx context.Context, tenantID, id string) (*BakeryProduct, error)
	GetProductByPLU(ctx context.Context, tenantID, plu string) (*BakeryProduct, error)
	ListProducts(ctx context.Context, tenantID string) ([]*BakeryProduct, error)
	CreateSale(ctx context.Context, sale *PDVSale) error
	CreateLossRecord(ctx context.Context, loss *LossRecord) error
	GetLossByPeriod(ctx context.Context, tenantID string, from, to time.Time) ([]*LossRecord, error)
}

// PDVService é o serviço de PDV rápido da padaria.
type PDVService struct {
	repo  BakeryRepository
	scale ScaleReader
}

func NewPDVService(repo BakeryRepository, scale ScaleReader) *PDVService {
	return &PDVService{repo: repo, scale: scale}
}

// ReadFromScale lê o peso da balança e retorna o produto + total a cobrar.
// Integração direta: operador pesa → sistema já monta o item de venda.
func (s *PDVService) ReadFromScale(ctx context.Context, tenantID, scaleID string) (*PDVSaleItem, error) {
	reading, err := s.scale.ReadWeight(ctx, scaleID)
	if err != nil {
		return nil, fmt.Errorf("bakery.ReadFromScale: erro na balança: %w", err)
	}
	if !reading.IsStable {
		return nil, fmt.Errorf("bakery.ReadFromScale: peso ainda não estabilizou")
	}

	product, err := s.repo.GetProductByPLU(ctx, tenantID, reading.PLUCode)
	if err != nil {
		return nil, fmt.Errorf("bakery.ReadFromScale: produto PLU '%s' não encontrado: %w",
			reading.PLUCode, err)
	}

	total := reading.WeightKG * product.UnitPrice

	return &PDVSaleItem{
		ProductID:   product.ID,
		ProductName: product.Name,
		SaleType:    SaleByWeight,
		Quantity:    reading.WeightKG,
		UnitPrice:   product.UnitPrice,
		TotalPrice:  roundPrice(total),
		IsBasket:    product.IsBasketItem,
		// IBS/CBS calculado pelo motor fiscal (tax.Engine) — injetado no handler
	}, nil
}

// CompleteSale finaliza uma venda no PDV, desconta estoque e persiste.
func (s *PDVService) CompleteSale(ctx context.Context, sale *PDVSale) error {
	if len(sale.Items) == 0 {
		return fmt.Errorf("bakery.CompleteSale: venda sem itens")
	}

	// Recalcular totais para garantir consistência
	var subtotal, totalTax float64
	for i, item := range sale.Items {
		sale.Items[i].TotalPrice = roundPrice(item.Quantity * item.UnitPrice * (1 - item.Discount))
		subtotal += sale.Items[i].TotalPrice
		totalTax += item.IBSAmount + item.CBSAmount
	}
	sale.Subtotal = roundPrice(subtotal)
	sale.TotalAmount = roundPrice(subtotal - sale.Discount)
	sale.TotalTax = roundPrice(totalTax)
	sale.SoldAt = time.Now().UTC()
	sale.Number = generateSaleNumber()

	return s.repo.CreateSale(ctx, sale)
}

// RegisterLoss registra uma perda e calcula o impacto financeiro.
func (s *PDVService) RegisterLoss(ctx context.Context, tenantID string, loss *LossRecord) error {
	if loss.ProductID == "" || loss.Quantity <= 0 {
		return fmt.Errorf("bakery.RegisterLoss: produto e quantidade obrigatórios")
	}
	loss.TenantID = tenantID
	loss.RecordedAt = time.Now().UTC()
	return s.repo.CreateLossRecord(ctx, loss)
}

// LossSummary retorna um resumo das perdas por período para análise gerencial.
func (s *PDVService) LossSummary(ctx context.Context, tenantID string, from, to time.Time) (map[LossReason]float64, error) {
	records, err := s.repo.GetLossByPeriod(ctx, tenantID, from, to)
	if err != nil {
		return nil, err
	}
	summary := make(map[LossReason]float64)
	for _, r := range records {
		summary[r.Reason] += r.CostValue
	}
	return summary, nil
}

func generateSaleNumber() string {
	return fmt.Sprintf("PDV-%d%06d", time.Now().Year(), time.Now().UnixMilli()%1000000)
}

func roundPrice(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// ListProducts lista todos os produtos ativos da padaria.
func (s *PDVService) ListProducts(ctx context.Context, tenantID string) ([]*BakeryProduct, error) {
	return s.repo.ListProducts(ctx, tenantID)
}
