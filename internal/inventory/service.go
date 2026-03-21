// Package inventory implementa o módulo de Estoque do Nexo One.
//
// Sistema camaleão — o mesmo módulo se adapta a cada nicho:
//   - Mecânica: peças e insumos com código do fabricante
//   - Padaria: ingredientes e produtos prontos com validade
//   - Indústria: matéria-prima e produtos acabados com BOM
//   - Estética: produtos de beleza com validade
//   - Calçados: grade cor/tamanho/SKU
//   - Logística: não usa estoque (desabilitado)
//
// Custo médio ponderado (CMP) recalculado automaticamente a cada entrada.
// Alertas de estoque mínimo via WhatsApp.
package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ── CONSTANTES ────────────────────────────────────────────────────────────────

const (
	MovementTypeIn       = "in"        // entrada de mercadoria
	MovementTypeOut      = "out"       // saída (venda, consumo, OS)
	MovementTypeAdjust   = "adjust"    // ajuste de inventário
	MovementTypeLoss     = "loss"      // perda/quebra (padaria)
	MovementTypeTransfer = "transfer"  // transferência entre filiais

	UnitUN  = "un"   // unidade
	UnitKG  = "kg"   // quilograma
	UnitG   = "g"    // grama
	UnitL   = "l"    // litro
	UnitML  = "ml"   // mililitro
	UnitM   = "m"    // metro
	UnitCX  = "cx"   // caixa
	UnitPAR = "par"  // par (calçados)
	UnitPC  = "pc"   // peça
)

// BusinessFields campos específicos por nicho.
type BusinessFields struct {
	// Mecânica
	ManufacturerCode string `json:"manufacturer_code,omitempty"` // código do fabricante
	Application      string `json:"application,omitempty"`        // ex: "Fiat Uno 2010-2015"
	Brand            string `json:"brand,omitempty"`

	// Padaria
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	IsIngredient  bool       `json:"is_ingredient,omitempty"`  // ingrediente ou produto final
	LossPercent   float64    `json:"loss_percent,omitempty"`   // % de perda estimada

	// Indústria
	IsBOM         bool     `json:"is_bom,omitempty"`          // faz parte de ficha técnica
	BOMParentID   string   `json:"bom_parent_id,omitempty"`   // produto pai na BOM

	// Calçados
	ColorCode     string   `json:"color_code,omitempty"`
	Size          string   `json:"size,omitempty"`
	GridParentID  string   `json:"grid_parent_id,omitempty"` // produto pai na grade

	// Estética
	Volume        float64  `json:"volume,omitempty"` // volume em ml
	IsConsumable  bool     `json:"is_consumable,omitempty"`
}

// Product representa um produto no estoque.
type Product struct {
	ID           string
	TenantID     string
	BusinessType string // mechanic, bakery, industry, aesthetics, shoes

	// Dados básicos
	Code         string  // código interno
	Barcode      string  // código de barras EAN
	Name         string
	Description  string
	Category     string
	Unit         string  // un, kg, l, m, par, etc.

	// Estoque
	Quantity     float64
	MinQuantity  float64 // estoque mínimo para alerta
	MaxQuantity  float64 // estoque máximo
	Location     string  // prateleira, setor, depósito

	// Financeiro
	CostPrice    float64 // custo médio ponderado (CMP)
	SalePrice    float64
	TotalValue   float64 // Quantity × CostPrice

	// Fiscal
	NCM          string
	CFOP         string

	// Campos específicos do nicho
	Extra        BusinessFields

	// Controle
	IsActive     bool
	AlertSent    bool
	LastMovement *time.Time
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Movement representa uma movimentação de estoque.
type Movement struct {
	ID            string
	TenantID      string
	ProductID     string
	ProductName   string
	Type          string  // in, out, adjust, loss, transfer
	Quantity      float64
	UnitCost      float64 // custo unitário na entrada
	TotalCost     float64
	PreviousStock float64
	NewStock      float64
	PreviousCMP   float64 // custo médio anterior
	NewCMP        float64 // novo custo médio após entrada

	// Referência
	ReferenceID   string // ID da OS, NF-e, venda, etc.
	ReferenceType string // os, nfe, sale, adjustment
	NFeKey        string

	Notes         string
	CreatedBy     string
	CreatedAt     time.Time
}

// StockAlert alerta de estoque mínimo.
type StockAlert struct {
	ProductID   string
	ProductName string
	Current     float64
	Minimum     float64
	Unit        string
	Deficit     float64 // quanto falta para atingir o mínimo
}

// InventorySummary resumo do estoque.
type InventorySummary struct {
	TotalProducts   int     `json:"total_products"`
	TotalValue      float64 `json:"total_value"`
	LowStockCount   int     `json:"low_stock_count"`
	OutOfStockCount int     `json:"out_of_stock_count"`
	TopProducts     []*Product `json:"top_products"`
}

// ── ERROS ─────────────────────────────────────────────────────────────────────

var (
	ErrProductNotFound    = errors.New("produto não encontrado")
	ErrInsufficientStock  = errors.New("estoque insuficiente")
	ErrInvalidQuantity    = errors.New("quantidade inválida")
	ErrDuplicateCode      = errors.New("código já cadastrado")
)

// ── FILTROS ───────────────────────────────────────────────────────────────────

type ProductFilter struct {
	Category    string
	LowStock    bool
	OutOfStock  bool
	Search      string
	Limit       int
	Offset      int
}

// ── REPOSITÓRIO ───────────────────────────────────────────────────────────────

type Repository interface {
	// Produtos
	CreateProduct(ctx context.Context, p *Product) error
	GetProduct(ctx context.Context, tenantID, id string) (*Product, error)
	GetProductByCode(ctx context.Context, tenantID, code string) (*Product, error)
	ListProducts(ctx context.Context, tenantID string, filter ProductFilter) ([]*Product, error)
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, tenantID, id string) error

	// Movimentações
	CreateMovement(ctx context.Context, m *Movement) error
	ListMovements(ctx context.Context, tenantID, productID string, limit int) ([]*Movement, error)

	// Alertas e relatórios
	GetLowStockProducts(ctx context.Context, tenantID string) ([]*Product, error)
	GetSummary(ctx context.Context, tenantID string) (*InventorySummary, error)
	MarkAlertSent(ctx context.Context, tenantID, productID string) error
}

// WhatsAppNotifier alerta de estoque mínimo.
type WhatsAppNotifier interface {
	SendLowStockAlert(ctx context.Context, phone string, alerts []*StockAlert) error
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

type Service struct {
	repo     Repository
	notifier WhatsAppNotifier
}

func NewService(repo Repository, notifier WhatsAppNotifier) *Service {
	return &Service{repo: repo, notifier: notifier}
}

// ── PRODUTOS ──────────────────────────────────────────────────────────────────

// CreateProduct cadastra um novo produto.
func (s *Service) CreateProduct(ctx context.Context, p *Product) error {
	if p.Name == "" {
		return errors.New("nome do produto é obrigatório")
	}
	if p.Unit == "" {
		p.Unit = UnitUN
	}

	p.IsActive = true
	p.TotalValue = p.Quantity * p.CostPrice
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	return s.repo.CreateProduct(ctx, p)
}

// UpdateProduct atualiza dados do produto (sem mexer no estoque).
func (s *Service) UpdateProduct(ctx context.Context, p *Product) error {
	existing, err := s.repo.GetProduct(ctx, p.TenantID, p.ID)
	if err != nil {
		return ErrProductNotFound
	}

	// Preservar quantidade e CMP — só mudam via movimentação
	p.Quantity = existing.Quantity
	p.CostPrice = existing.CostPrice
	p.TotalValue = existing.TotalValue
	p.UpdatedAt = time.Now()

	return s.repo.UpdateProduct(ctx, p)
}

// ── MOVIMENTAÇÕES ─────────────────────────────────────────────────────────────

// AddStock entrada de mercadoria com cálculo automático de CMP.
func (s *Service) AddStock(ctx context.Context, tenantID, productID string, quantity, unitCost float64, refID, refType, nfeKey, notes, userID string) (*Movement, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	product, err := s.repo.GetProduct(ctx, tenantID, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	prevStock := product.Quantity
	prevCMP := product.CostPrice

	// Custo Médio Ponderado (CMP)
	// CMP = (Qtd_atual × CMP_atual + Qtd_entrada × Custo_entrada) / (Qtd_atual + Qtd_entrada)
	newCMP := prevCMP
	if unitCost > 0 {
		totalValue := (prevStock * prevCMP) + (quantity * unitCost)
		newStock := prevStock + quantity
		if newStock > 0 {
			newCMP = totalValue / newStock
		}
	}

	newStock := prevStock + quantity

	// Atualizar produto
	product.Quantity = newStock
	product.CostPrice = newCMP
	product.TotalValue = newStock * newCMP
	product.AlertSent = false // reset alerta após entrada
	now := time.Now()
	product.LastMovement = &now
	product.UpdatedAt = now

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("inventory.AddStock: %w", err)
	}

	movement := &Movement{
		TenantID:      tenantID,
		ProductID:     productID,
		ProductName:   product.Name,
		Type:          MovementTypeIn,
		Quantity:      quantity,
		UnitCost:      unitCost,
		TotalCost:     quantity * unitCost,
		PreviousStock: prevStock,
		NewStock:      newStock,
		PreviousCMP:   prevCMP,
		NewCMP:        newCMP,
		ReferenceID:   refID,
		ReferenceType: refType,
		NFeKey:        nfeKey,
		Notes:         notes,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateMovement(ctx, movement); err != nil {
		return nil, fmt.Errorf("inventory.AddStock movement: %w", err)
	}

	return movement, nil
}

// RemoveStock saída de estoque (venda, consumo, OS).
func (s *Service) RemoveStock(ctx context.Context, tenantID, productID string, quantity float64, refID, refType, notes, userID string) (*Movement, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	product, err := s.repo.GetProduct(ctx, tenantID, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if product.Quantity < quantity {
		return nil, fmt.Errorf("%w: disponível %.2f %s", ErrInsufficientStock, product.Quantity, product.Unit)
	}

	prevStock := product.Quantity
	newStock := prevStock - quantity

	product.Quantity = newStock
	product.TotalValue = newStock * product.CostPrice
	now := time.Now()
	product.LastMovement = &now
	product.UpdatedAt = now

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("inventory.RemoveStock: %w", err)
	}

	movement := &Movement{
		TenantID:      tenantID,
		ProductID:     productID,
		ProductName:   product.Name,
		Type:          MovementTypeOut,
		Quantity:      quantity,
		UnitCost:      product.CostPrice,
		TotalCost:     quantity * product.CostPrice,
		PreviousStock: prevStock,
		NewStock:      newStock,
		PreviousCMP:   product.CostPrice,
		NewCMP:        product.CostPrice,
		ReferenceID:   refID,
		ReferenceType: refType,
		Notes:         notes,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateMovement(ctx, movement); err != nil {
		return nil, err
	}

	// Verificar estoque mínimo após saída
	if product.MinQuantity > 0 && newStock <= product.MinQuantity && !product.AlertSent {
		s.triggerLowStockAlert(ctx, tenantID, product)
	}

	return movement, nil
}

// AdjustStock ajuste manual de inventário.
func (s *Service) AdjustStock(ctx context.Context, tenantID, productID string, newQuantity float64, notes, userID string) (*Movement, error) {
	product, err := s.repo.GetProduct(ctx, tenantID, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	prevStock := product.Quantity
	diff := newQuantity - prevStock

	product.Quantity = newQuantity
	product.TotalValue = newQuantity * product.CostPrice
	now := time.Now()
	product.LastMovement = &now
	product.UpdatedAt = now

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}

	movType := MovementTypeAdjust
	if diff < 0 {
		movType = MovementTypeLoss
	}

	movement := &Movement{
		TenantID:      tenantID,
		ProductID:     productID,
		ProductName:   product.Name,
		Type:          movType,
		Quantity:      diff,
		UnitCost:      product.CostPrice,
		TotalCost:     diff * product.CostPrice,
		PreviousStock: prevStock,
		NewStock:      newQuantity,
		PreviousCMP:   product.CostPrice,
		NewCMP:        product.CostPrice,
		Notes:         notes,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
	}

	return movement, s.repo.CreateMovement(ctx, movement)
}

// ProcessNFeEntry processa entrada de NF-e e dá entrada automática no estoque.
func (s *Service) ProcessNFeEntry(ctx context.Context, tenantID, userID string, items []NFeItem, nfeKey string) ([]*Movement, error) {
	var movements []*Movement

	for _, item := range items {
		// Buscar produto pelo código de barras ou NCM
		product, err := s.findOrCreateProduct(ctx, tenantID, item)
		if err != nil {
			continue
		}

		movement, err := s.AddStock(ctx, tenantID, product.ID,
			item.Quantity, item.UnitPrice,
			nfeKey, "nfe", nfeKey,
			fmt.Sprintf("Entrada NF-e — %s", item.Description),
			userID,
		)
		if err != nil {
			continue
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

// NFeItem item de uma NF-e para entrada no estoque.
type NFeItem struct {
	Code        string
	Barcode     string
	Description string
	NCM         string
	Quantity    float64
	Unit        string
	UnitPrice   float64
}

// findOrCreateProduct busca produto existente ou cria novo.
func (s *Service) findOrCreateProduct(ctx context.Context, tenantID string, item NFeItem) (*Product, error) {
	// Tentar encontrar pelo código
	if item.Code != "" {
		p, err := s.repo.GetProductByCode(ctx, tenantID, item.Code)
		if err == nil {
			return p, nil
		}
	}

	// Criar novo produto
	p := &Product{
		TenantID:    tenantID,
		Code:        item.Code,
		Barcode:     item.Barcode,
		Name:        item.Description,
		Unit:        item.Unit,
		NCM:         item.NCM,
		CostPrice:   item.UnitPrice,
		IsActive:    true,
	}

	if err := s.repo.CreateProduct(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ── ALERTAS ───────────────────────────────────────────────────────────────────

func (s *Service) triggerLowStockAlert(ctx context.Context, tenantID string, product *Product) {
	if s.notifier == nil {
		return
	}

	alert := &StockAlert{
		ProductID:   product.ID,
		ProductName: product.Name,
		Current:     product.Quantity,
		Minimum:     product.MinQuantity,
		Unit:        product.Unit,
		Deficit:     product.MinQuantity - product.Quantity,
	}

	s.notifier.SendLowStockAlert(ctx, "", []*StockAlert{alert})
	s.repo.MarkAlertSent(ctx, tenantID, product.ID)
}

// CheckLowStock verifica todos os produtos com estoque baixo e envia alertas.
func (s *Service) CheckLowStock(ctx context.Context, tenantID, phone string) error {
	if s.notifier == nil {
		return nil
	}

	lowStock, err := s.repo.GetLowStockProducts(ctx, tenantID)
	if err != nil {
		return err
	}

	var alerts []*StockAlert
	for _, p := range lowStock {
		if p.AlertSent {
			continue
		}
		alerts = append(alerts, &StockAlert{
			ProductID:   p.ID,
			ProductName: p.Name,
			Current:     p.Quantity,
			Minimum:     p.MinQuantity,
			Unit:        p.Unit,
			Deficit:     p.MinQuantity - p.Quantity,
		})
	}

	if len(alerts) == 0 {
		return nil
	}

	return s.notifier.SendLowStockAlert(ctx, phone, alerts)
}

// ── CONSULTAS ─────────────────────────────────────────────────────────────────

func (s *Service) ListProducts(ctx context.Context, tenantID string, filter ProductFilter) ([]*Product, error) {
	return s.repo.ListProducts(ctx, tenantID, filter)
}

func (s *Service) GetProduct(ctx context.Context, tenantID, id string) (*Product, error) {
	return s.repo.GetProduct(ctx, tenantID, id)
}

func (s *Service) GetSummary(ctx context.Context, tenantID string) (*InventorySummary, error) {
	return s.repo.GetSummary(ctx, tenantID)
}

func (s *Service) ListMovements(ctx context.Context, tenantID, productID string, limit int) ([]*Movement, error) {
	return s.repo.ListMovements(ctx, tenantID, productID, limit)
}

// NichoCamaleao retorna configurações visuais do estoque por nicho.
func NichoCamaleao(businessType string) map[string]interface{} {
	configs := map[string]map[string]interface{}{
		"mechanic": {
			"name":         "Peças e Insumos",
			"icon":         "🔧",
			"unit_default": UnitUN,
			"categories":   []string{"peças", "filtros", "fluidos", "pneus", "elétrica", "freios", "outros"},
			"extra_fields": []string{"manufacturer_code", "application", "brand"},
			"show_expiry":  false,
		},
		"bakery": {
			"name":         "Ingredientes e Produtos",
			"icon":         "🍞",
			"unit_default": UnitKG,
			"categories":   []string{"farinhas", "açúcares", "gorduras", "laticínios", "ovos", "embalagens", "produtos_prontos"},
			"extra_fields": []string{"expiry_date", "is_ingredient", "loss_percent"},
			"show_expiry":  true,
		},
		"industry": {
			"name":         "Matéria-Prima e Produtos",
			"icon":         "🏭",
			"unit_default": UnitKG,
			"categories":   []string{"materia_prima", "produto_acabado", "embalagens", "insumos"},
			"extra_fields": []string{"is_bom", "bom_parent_id"},
			"show_expiry":  false,
		},
		"aesthetics": {
			"name":         "Produtos de Beleza",
			"icon":         "💇",
			"unit_default": UnitML,
			"categories":   []string{"tintura", "tratamento", "finalizador", "shampoo", "outros"},
			"extra_fields": []string{"expiry_date", "volume", "brand"},
			"show_expiry":  true,
		},
		"shoes": {
			"name":         "Calçados e Acessórios",
			"icon":         "👟",
			"unit_default": UnitPAR,
			"categories":   []string{"feminino", "masculino", "infantil", "esportivo", "acessórios"},
			"extra_fields": []string{"color_code", "size", "grid_parent_id"},
			"show_expiry":  false,
		},
		"logistics": {
			"name":         "",
			"disabled":     true, // logística não usa estoque
		},
	}

	if cfg, ok := configs[businessType]; ok {
		return cfg
	}
	return configs["mechanic"] // default
}
