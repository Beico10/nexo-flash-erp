// Package shoes implementa o módulo de Calçados do Nexo One.
//
// Funcionalidades (Briefing Mestre §1 — Nicho Calçados):
//   - Matriz de Grade (Cor × Tamanho × SKU) para gestão de estoque precisa
//   - Sistema de Comissões por vendedor (percentual por venda/meta)
package shoes

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GridCell representa uma célula da Matriz de Grade (Cor + Tamanho).
type GridCell struct {
	SKU       string  `json:"sku"`
	Color     string  `json:"color"`
	Size      string  `json:"size"`
	Stock     int     `json:"stock"`
	Price     float64 `json:"price"`
	CostPrice float64 `json:"cost_price"`
	Barcode   string  `json:"barcode"`
	Active    bool    `json:"active"`
}

// ProductGrid é a grade completa de um modelo de calçado.
// Estrutura: Modelo → Grade[Cor][Tamanho] → GridCell
type ProductGrid struct {
	ID        string                       `json:"id"`
	TenantID  string                       `json:"-"`
	ModelCode string                       `json:"model_code"`
	ModelName string                       `json:"model_name"`
	Brand     string                       `json:"brand"`
	NCMCode   string                       `json:"ncm_code"`
	// Grid[cor][tamanho] → célula
	Grid      map[string]map[string]GridCell `json:"grid"`
	// Listas ordenadas de cores e tamanhos para exibição na UI
	Colors    []string                     `json:"colors"`
	Sizes     []string                     `json:"sizes"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}

// GridSummary é um resumo da grade para listagem rápida.
type GridSummary struct {
	ModelCode    string
	ModelName    string
	TotalSKUs    int
	InStockSKUs  int
	TotalUnits   int
	TotalValue   float64 // estoque × preço de custo
}

// CommissionRule define a regra de comissão de um vendedor.
type CommissionRule struct {
	ID           string  `json:"id"`
	TenantID     string  `json:"-"`
	SellerID     string  `json:"seller_id"`
	SellerName   string  `json:"seller_name"`
	BaseRate     float64 `json:"base_rate"`
	BonusRate    float64 `json:"bonus_rate"`
	MonthlTarget float64 `json:"monthly_target"`
	Active       bool    `json:"active"`
}

// CommissionResult é o resultado do cálculo de comissão.
type CommissionResult struct {
	SellerID        string    `json:"seller_id"`
	PeriodFrom      time.Time `json:"period_from"`
	PeriodTo        time.Time `json:"period_to"`
	TotalSales      float64   `json:"total_sales"`
	BaseCommission  float64   `json:"base_commission"`
	BonusCommission float64   `json:"bonus_commission"`
	TotalCommission float64   `json:"total_commission"`
	MetAchieved     bool      `json:"met_achieved"`
}

// GridRepository acessa grades de calçados.
type GridRepository interface {
	Create(ctx context.Context, grid *ProductGrid) error
	Update(ctx context.Context, grid *ProductGrid) error
	GetByModel(ctx context.Context, tenantID, modelCode string) (*ProductGrid, error)
	UpdateStock(ctx context.Context, tenantID, sku string, delta int) error
	GetBySKU(ctx context.Context, tenantID, sku string) (*GridCell, error)
}

// SalesRepository acessa vendas para cálculo de comissão.
type SalesRepository interface {
	SumBySeller(ctx context.Context, tenantID, sellerID string, from, to time.Time) (float64, error)
}

// GridService gerencia a Matriz de Grade de Calçados.
type GridService struct {
	repo GridRepository
}

func NewGridService(r GridRepository) *GridService { return &GridService{repo: r} }

// BuildGrid constrói a grade a partir de listas de cores e tamanhos.
// Gera SKUs automaticamente no padrão: MODELO-COR-TAMANHO
// Ex: "SANDAL001-PRETO-38" → SKU único por célula.
func BuildGrid(tenantID, modelCode, modelName string, colors, sizes []string, basePrice, baseCost float64) *ProductGrid {
	grid := &ProductGrid{
		TenantID:  tenantID,
		ModelCode: modelCode,
		ModelName: modelName,
		Colors:    colors,
		Sizes:     sizes,
		Grid:      make(map[string]map[string]GridCell),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	for _, color := range colors {
		grid.Grid[color] = make(map[string]GridCell)
		for _, size := range sizes {
			sku := fmt.Sprintf("%s-%s-%s",
				strings.ToUpper(modelCode),
				strings.ToUpper(normalizeColor(color)),
				size)
			grid.Grid[color][size] = GridCell{
				SKU:       sku,
				Color:     color,
				Size:      size,
				Stock:     0,
				Price:     basePrice,
				CostPrice: baseCost,
				Active:    true,
			}
		}
	}
	return grid
}

// Summary retorna o resumo estatístico de uma grade.
func Summary(grid *ProductGrid) GridSummary {
	var totalSKUs, inStockSKUs, totalUnits int
	var totalValue float64

	for _, sizes := range grid.Grid {
		for _, cell := range sizes {
			if !cell.Active {
				continue
			}
			totalSKUs++
			if cell.Stock > 0 {
				inStockSKUs++
				totalUnits += cell.Stock
				totalValue += float64(cell.Stock) * cell.CostPrice
			}
		}
	}

	return GridSummary{
		ModelCode:   grid.ModelCode,
		ModelName:   grid.ModelName,
		TotalSKUs:   totalSKUs,
		InStockSKUs: inStockSKUs,
		TotalUnits:  totalUnits,
		TotalValue:  totalValue,
	}
}

// CalculateCommission calcula a comissão de um vendedor no período.
func CalculateCommission(ctx context.Context, sales SalesRepository, rule CommissionRule, from, to time.Time) (*CommissionResult, error) {
	total, err := sales.SumBySeller(ctx, rule.TenantID, rule.SellerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("shoes.CalculateCommission: %w", err)
	}

	base := total * rule.BaseRate
	var bonus float64
	metAchieved := total >= rule.MonthlTarget
	if metAchieved && rule.BonusRate > 0 {
		bonus = total * rule.BonusRate
	}

	return &CommissionResult{
		SellerID:        rule.SellerID,
		PeriodFrom:      from,
		PeriodTo:        to,
		TotalSales:      total,
		BaseCommission:  roundPrice(base),
		BonusCommission: roundPrice(bonus),
		TotalCommission: roundPrice(base + bonus),
		MetAchieved:     metAchieved,
	}, nil
}

func normalizeColor(c string) string {
	c = strings.ReplaceAll(c, " ", "")
	return c
}

func roundPrice(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
