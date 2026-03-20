package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/industry"
)

type IndustryRepo struct {
	mu        sync.RWMutex
	boms      []*industry.BOM
	orders    []*industry.ProductionOrder
	materials map[string]float64 // materialID -> balance
}

func NewIndustryRepo() *IndustryRepo {
	r := &IndustryRepo{materials: make(map[string]float64)}
	r.seed()
	return r
}

func (r *IndustryRepo) seed() {
	tid := "00000000-0000-0000-0000-000000000001"
	now := time.Now()

	r.boms = []*industry.BOM{
		{
			ID: "bom-1", TenantID: tid, ProductID: "prod-mesa", ProductName: "Mesa de Escritorio 1.20m", Version: 1,
			Items: []industry.BOMItem{
				{ID: "bi-1", ParentProductID: "prod-mesa", ComponentID: "mat-mdf", ComponentName: "MDF 15mm", NCMCode: "44101190", Quantity: 2.4, Unit: "m2", ScrapFactor: 0.05, UnitCost: 45.00},
				{ID: "bi-2", ParentProductID: "prod-mesa", ComponentID: "mat-tubo", ComponentName: "Tubo metalon 30x30", NCMCode: "73066100", Quantity: 6.0, Unit: "m", ScrapFactor: 0.02, UnitCost: 12.50},
				{ID: "bi-3", ParentProductID: "prod-mesa", ComponentID: "mat-parafuso", ComponentName: "Parafuso 5x40mm", NCMCode: "73181500", Quantity: 16, Unit: "un", ScrapFactor: 0.10, UnitCost: 0.35},
				{ID: "bi-4", ParentProductID: "prod-mesa", ComponentID: "mat-verniz", ComponentName: "Verniz PU transparente", NCMCode: "32091090", Quantity: 0.5, Unit: "L", ScrapFactor: 0.15, UnitCost: 38.00},
			},
			TotalCost: 230.16, ValidFrom: now.AddDate(0, -3, 0), CreatedBy: "admin", CreatedAt: now.AddDate(0, -3, 0),
		},
		{
			ID: "bom-2", TenantID: tid, ProductID: "prod-cadeira", ProductName: "Cadeira Giratoria Ergonomica", Version: 1,
			Items: []industry.BOMItem{
				{ID: "bi-5", ParentProductID: "prod-cadeira", ComponentID: "mat-tecido", ComponentName: "Tecido mesh preto", NCMCode: "59032090", Quantity: 1.2, Unit: "m2", ScrapFactor: 0.08, UnitCost: 55.00},
				{ID: "bi-6", ParentProductID: "prod-cadeira", ComponentID: "mat-espuma", ComponentName: "Espuma D33 5cm", NCMCode: "39211990", Quantity: 0.8, Unit: "m2", ScrapFactor: 0.05, UnitCost: 42.00},
				{ID: "bi-7", ParentProductID: "prod-cadeira", ComponentID: "mat-base", ComponentName: "Base giratoria cromada", NCMCode: "94017900", Quantity: 1, Unit: "un", ScrapFactor: 0, UnitCost: 85.00},
			},
			TotalCost: 189.52, ValidFrom: now.AddDate(0, -2, 0), CreatedBy: "admin", CreatedAt: now.AddDate(0, -2, 0),
		},
	}

	r.orders = []*industry.ProductionOrder{
		{ID: "op-1", TenantID: tid, Number: "OP-2026-000001", ProductID: "prod-mesa", ProductName: "Mesa de Escritorio 1.20m", BOMID: "bom-1", PlannedQty: 50, ProducedQty: 32, Unit: "un", Status: industry.POStatusInProgress, PlannedStart: now.AddDate(0, 0, -5), PlannedEnd: now.AddDate(0, 0, 5), CreatedAt: now.AddDate(0, 0, -7)},
		{ID: "op-2", TenantID: tid, Number: "OP-2026-000002", ProductID: "prod-cadeira", ProductName: "Cadeira Giratoria Ergonomica", BOMID: "bom-2", PlannedQty: 100, ProducedQty: 0, Unit: "un", Status: industry.POStatusPlanned, PlannedStart: now.AddDate(0, 0, 3), PlannedEnd: now.AddDate(0, 0, 15), CreatedAt: now.AddDate(0, 0, -1)},
		{ID: "op-3", TenantID: tid, Number: "OP-2026-000003", ProductID: "prod-mesa", ProductName: "Mesa de Escritorio 1.20m", BOMID: "bom-1", PlannedQty: 20, ProducedQty: 20, Unit: "un", Status: industry.POStatusDone, PlannedStart: now.AddDate(0, 0, -15), PlannedEnd: now.AddDate(0, 0, -8), CreatedAt: now.AddDate(0, 0, -18)},
	}

	r.materials = map[string]float64{
		"mat-mdf": 120.0, "mat-tubo": 350.0, "mat-parafuso": 5000.0, "mat-verniz": 25.0,
		"mat-tecido": 80.0, "mat-espuma": 60.0, "mat-base": 150.0,
	}
}

// BOMRepository (renamed to avoid Create conflict)
func (r *IndustryRepo) GetCurrent(_ context.Context, tenantID, productID string) (*industry.BOM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := len(r.boms) - 1; i >= 0; i-- {
		if r.boms[i].TenantID == tenantID && r.boms[i].ProductID == productID {
			return r.boms[i], nil
		}
	}
	return nil, fmt.Errorf("BOM nao encontrada")
}

func (r *IndustryRepo) CreateBOM(_ context.Context, bom *industry.BOM) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if bom.ID == "" { bom.ID = fmt.Sprintf("bom-%d", time.Now().UnixNano()) }
	r.boms = append(r.boms, bom)
	return nil
}

func (r *IndustryRepo) ListVersions(_ context.Context, tenantID, productID string) ([]*industry.BOM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*industry.BOM
	for _, b := range r.boms {
		if b.TenantID == tenantID && b.ProductID == productID { result = append(result, b) }
	}
	return result, nil
}

// BOMAdapter wraps IndustryRepo to satisfy industry.BOMRepository.
type BOMAdapter struct{ R *IndustryRepo }

func (a *BOMAdapter) GetCurrent(ctx context.Context, tenantID, productID string) (*industry.BOM, error) {
	return a.R.GetCurrent(ctx, tenantID, productID)
}
func (a *BOMAdapter) Create(ctx context.Context, bom *industry.BOM) error {
	return a.R.CreateBOM(ctx, bom)
}
func (a *BOMAdapter) ListVersions(ctx context.Context, tenantID, productID string) ([]*industry.BOM, error) {
	return a.R.ListVersions(ctx, tenantID, productID)
}

// ProductionRepository
func (r *IndustryRepo) Create(_ context.Context, op *industry.ProductionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if op.ID == "" { op.ID = fmt.Sprintf("op-%d", time.Now().UnixNano()) }
	op.CreatedAt = time.Now()
	r.orders = append(r.orders, op)
	return nil
}

func (r *IndustryRepo) Update(_ context.Context, op *industry.ProductionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, o := range r.orders {
		if o.ID == op.ID { r.orders[i] = op; return nil }
	}
	return fmt.Errorf("OP nao encontrada")
}

func (r *IndustryRepo) GetByID(_ context.Context, tenantID, id string) (*industry.ProductionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.orders {
		if o.ID == id && o.TenantID == tenantID { return o, nil }
	}
	return nil, fmt.Errorf("OP nao encontrada")
}

func (r *IndustryRepo) ListByStatus(_ context.Context, tenantID string, status industry.ProductionOrderStatus) ([]*industry.ProductionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*industry.ProductionOrder
	for _, o := range r.orders {
		if o.TenantID == tenantID && (status == "" || o.Status == status) { result = append(result, o) }
	}
	return result, nil
}

// MaterialRepository
func (r *IndustryRepo) GetBalance(_ context.Context, _, materialID string) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.materials[materialID], nil
}

func (r *IndustryRepo) Reserve(_ context.Context, _, materialID, _ string, qty float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.materials[materialID] < qty { return fmt.Errorf("estoque insuficiente") }
	r.materials[materialID] -= qty
	return nil
}

func (r *IndustryRepo) Consume(_ context.Context, _, materialID, _ string, qty float64) error {
	return r.Reserve(context.Background(), "", materialID, "", qty)
}

// ListBOMs retorna todas as BOMs do tenant.
func (r *IndustryRepo) ListBOMs(_ context.Context, tenantID string) ([]*industry.BOM, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*industry.BOM
	for _, b := range r.boms {
		if b.TenantID == tenantID { result = append(result, b) }
	}
	return result, nil
}

// ListMaterials retorna saldos de materiais.
func (r *IndustryRepo) ListMaterials() map[string]float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make(map[string]float64)
	for k, v := range r.materials { cp[k] = v }
	return cp
}
