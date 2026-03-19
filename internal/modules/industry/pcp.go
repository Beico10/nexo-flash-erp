// Package industry implementa o módulo de Indústria do Nexo One.
//
// Funcionalidades (Briefing Mestre §1 — Nicho Indústria):
//   - PCP: Planejamento e Controle da Produção
//   - Ficha Técnica (BOM — Bill of Materials): estrutura de produto
//   - Gestão de Insumos: entradas, saídas e saldo por centro de custo
package industry

import (
	"context"
	"fmt"
	"time"
)

// ProductionOrderStatus estados de uma Ordem de Produção.
type ProductionOrderStatus string

const (
	POStatusPlanned    ProductionOrderStatus = "planned"
	POStatusReleased   ProductionOrderStatus = "released"
	POStatusInProgress ProductionOrderStatus = "in_progress"
	POStatusDone       ProductionOrderStatus = "done"
	POStatusCancelled  ProductionOrderStatus = "cancelled"
)

// BOMItem representa um componente na Ficha Técnica (Bill of Materials).
type BOMItem struct {
	ID              string
	ParentProductID string  // produto final
	ComponentID     string  // insumo/matéria-prima
	ComponentName   string
	NCMCode         string
	Quantity        float64 // quantidade por unidade produzida
	Unit            string  // kg, L, un, m²
	ScrapFactor     float64 // fator de perda (ex: 0.05 = 5% de refugo)
	NetQuantity     float64 // quantity * (1 + scrapFactor)
	UnitCost        float64
	TotalCost       float64
}

// BOM representa a Ficha Técnica completa de um produto.
type BOM struct {
	ID          string
	TenantID    string
	ProductID   string
	ProductName string
	Version     int
	Items       []BOMItem
	TotalCost   float64
	ValidFrom   time.Time
	ValidUntil  *time.Time
	CreatedBy   string
	CreatedAt   time.Time
}

// ProductionOrder representa uma Ordem de Produção (OP).
type ProductionOrder struct {
	ID           string
	TenantID     string
	Number       string
	ProductID    string
	ProductName  string
	BOMID        string
	PlannedQty   float64
	ProducedQty  float64
	Unit         string
	Status       ProductionOrderStatus
	PlannedStart time.Time
	PlannedEnd   time.Time
	ActualStart  *time.Time
	ActualEnd    *time.Time
	// Insumos reservados para esta OP (explodido da BOM)
	MaterialReservations []MaterialReservation
	CreatedAt   time.Time
}

// MaterialReservation reserva insumos para uma Ordem de Produção.
type MaterialReservation struct {
	MaterialID  string
	MaterialName string
	RequiredQty float64
	Unit        string
	ReservedQty float64
	IsAvailable bool
}

// BOMRepository acessa fichas técnicas.
type BOMRepository interface {
	GetCurrent(ctx context.Context, tenantID, productID string) (*BOM, error)
	Create(ctx context.Context, bom *BOM) error
	ListVersions(ctx context.Context, tenantID, productID string) ([]*BOM, error)
}

// ProductionRepository acessa ordens de produção.
type ProductionRepository interface {
	Create(ctx context.Context, op *ProductionOrder) error
	Update(ctx context.Context, op *ProductionOrder) error
	GetByID(ctx context.Context, tenantID, id string) (*ProductionOrder, error)
	ListByStatus(ctx context.Context, tenantID string, status ProductionOrderStatus) ([]*ProductionOrder, error)
}

// MaterialRepository acessa saldos de insumos.
type MaterialRepository interface {
	GetBalance(ctx context.Context, tenantID, materialID string) (float64, error)
	Reserve(ctx context.Context, tenantID, materialID, opID string, qty float64) error
	Consume(ctx context.Context, tenantID, materialID, opID string, qty float64) error
}

// PCPService é o serviço de Planejamento e Controle da Produção.
type PCPService struct {
	boms      BOMRepository
	orders    ProductionRepository
	materials MaterialRepository
}

func NewPCPService(b BOMRepository, o ProductionRepository, m MaterialRepository) *PCPService {
	return &PCPService{boms: b, orders: o, materials: m}
}

// ExplodeBOM calcula todos os insumos necessários para uma quantidade de produção,
// aplicando o fator de perda (scrap factor) de cada componente.
func (s *PCPService) ExplodeBOM(ctx context.Context, tenantID, productID string, qty float64) ([]MaterialReservation, float64, error) {
	bom, err := s.boms.GetCurrent(ctx, tenantID, productID)
	if err != nil {
		return nil, 0, fmt.Errorf("industry.ExplodeBOM: BOM não encontrada: %w", err)
	}

	var reservations []MaterialReservation
	var totalCost float64

	for _, item := range bom.Items {
		// Quantidade bruta considerando perda
		netQty := item.Quantity * (1 + item.ScrapFactor) * qty
		balance, err := s.materials.GetBalance(ctx, tenantID, item.ComponentID)
		if err != nil {
			return nil, 0, err
		}
		cost := netQty * item.UnitCost
		totalCost += cost
		reservations = append(reservations, MaterialReservation{
			MaterialID:   item.ComponentID,
			MaterialName: item.ComponentName,
			RequiredQty:  roundQty(netQty),
			Unit:         item.Unit,
			IsAvailable:  balance >= netQty,
		})
	}

	return reservations, roundPrice(totalCost), nil
}

// ReleaseOrder libera uma OP para produção e reserva os insumos.
func (s *PCPService) ReleaseOrder(ctx context.Context, op *ProductionOrder) error {
	reservations, _, err := s.ExplodeBOM(ctx, op.TenantID, op.ProductID, op.PlannedQty)
	if err != nil {
		return err
	}

	// Verifica disponibilidade de todos os insumos
	for _, r := range reservations {
		if !r.IsAvailable {
			return fmt.Errorf("industry.ReleaseOrder: insumo '%s' insuficiente (necessário: %.3f %s)",
				r.MaterialName, r.RequiredQty, r.Unit)
		}
	}

	// Reserva os insumos
	for _, r := range reservations {
		if err := s.materials.Reserve(ctx, op.TenantID, r.MaterialID, op.ID, r.RequiredQty); err != nil {
			return fmt.Errorf("industry.ReleaseOrder: reserva falhou: %w", err)
		}
	}

	op.MaterialReservations = reservations
	op.Status = POStatusReleased
	op.Number = generateOPNumber()
	return s.orders.Update(ctx, op)
}

func generateOPNumber() string {
	return fmt.Sprintf("OP-%d-%06d", time.Now().Year(), time.Now().UnixMilli()%1000000)
}

func roundQty(v float64) float64 { return float64(int(v*1000+0.5)) / 1000 }
func roundPrice(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
