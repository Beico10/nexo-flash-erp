// Package postgres — repositório PostgreSQL do módulo de Padaria.
// Implementa bakery.BakeryRepository.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nexoflash/nexo-flash/internal/modules/bakery"
)

// BakeryRepo implementa bakery.BakeryRepository.
type BakeryRepo struct {
	db *DB
}

func NewBakeryRepo(db *DB) *BakeryRepo { return &BakeryRepo{db: db} }

// GetProduct busca um produto pelo ID.
func (r *BakeryRepo) GetProduct(ctx context.Context, tenantID, id string) (*bakery.BakeryProduct, error) {
	var p bakery.BakeryProduct
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, sku, name, sale_type, unit_price,
			       COALESCE(ncm_code,''), is_basket_item,
			       COALESCE(basket_category,''), COALESCE(scale_plu,''),
			       current_stock, min_stock, active
			FROM bakery_products
			WHERE id = $1 AND tenant_id = $2 AND active = TRUE`,
			id, tenantID).Scan(
			&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.SaleType, &p.UnitPrice,
			&p.NCMCode, &p.IsBasketItem, &p.BasketCategory, &p.ScaleCode,
			&p.CurrentStock, &p.MinStock, &p.Active,
		)
	})
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("produto %s não encontrado", id)
	}
	return &p, err
}

// GetProductByPLU busca produto pelo código PLU da balança.
func (r *BakeryRepo) GetProductByPLU(ctx context.Context, tenantID, plu string) (*bakery.BakeryProduct, error) {
	var p bakery.BakeryProduct
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, sku, name, sale_type, unit_price,
			       COALESCE(ncm_code,''), is_basket_item,
			       COALESCE(basket_category,''), COALESCE(scale_plu,''),
			       current_stock, min_stock, active
			FROM bakery_products
			WHERE tenant_id = $1 AND scale_plu = $2 AND active = TRUE`,
			tenantID, plu).Scan(
			&p.ID, &p.TenantID, &p.SKU, &p.Name, &p.SaleType, &p.UnitPrice,
			&p.NCMCode, &p.IsBasketItem, &p.BasketCategory, &p.ScaleCode,
			&p.CurrentStock, &p.MinStock, &p.Active,
		)
	})
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("produto PLU '%s' não encontrado", plu)
	}
	return &p, err
}

// ListProducts lista todos os produtos ativos da padaria.
func (r *BakeryRepo) ListProducts(ctx context.Context, tenantID string) ([]*bakery.BakeryProduct, error) {
	var list []*bakery.BakeryProduct
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, sku, name, sale_type, unit_price,
			       COALESCE(ncm_code,''), is_basket_item,
			       COALESCE(scale_plu,''), current_stock, active
			FROM bakery_products
			WHERE tenant_id = $1 AND active = TRUE
			ORDER BY name`, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p bakery.BakeryProduct
			p.TenantID = tenantID
			if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.SaleType,
				&p.UnitPrice, &p.NCMCode, &p.IsBasketItem,
				&p.ScaleCode, &p.CurrentStock, &p.Active); err != nil {
				return err
			}
			list = append(list, &p)
		}
		return rows.Err()
	})
	return list, err
}

// CreateSale persiste uma venda do PDV e atualiza o estoque.
func (r *BakeryRepo) CreateSale(ctx context.Context, sale *bakery.PDVSale) error {
	return r.db.WithTenant(ctx, sale.TenantID, func(tx *sql.Tx) error {
		// Inserir venda
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO bakery_sales
			(tenant_id, number, subtotal, discount, total_amount, total_tax, payment_method, operator_id, sold_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id`,
			sale.TenantID, sale.Number, sale.Subtotal, sale.Discount,
			sale.TotalAmount, sale.TotalTax, sale.PaymentMethod,
			nullString(sale.OperatorID), sale.SoldAt,
		).Scan(&sale.ID); err != nil {
			return fmt.Errorf("CreateSale: inserir venda: %w", err)
		}

		// Inserir itens e atualizar estoque
		for _, item := range sale.Items {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO bakery_sale_items
				(tenant_id, sale_id, product_id, sale_type, quantity, unit_price, discount, total_price, ibs_amount, cbs_amount, is_basket)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
				sale.TenantID, sale.ID, item.ProductID, string(item.SaleType),
				item.Quantity, item.UnitPrice, item.Discount, item.TotalPrice,
				item.IBSAmount, item.CBSAmount, item.IsBasket,
			)
			if err != nil {
				return fmt.Errorf("CreateSale: inserir item: %w", err)
			}

			// Atualizar estoque
			_, err = tx.ExecContext(ctx, `
				UPDATE bakery_products
				SET current_stock = current_stock - $3
				WHERE id = $1 AND tenant_id = $2`,
				item.ProductID, sale.TenantID, item.Quantity)
			if err != nil {
				return fmt.Errorf("CreateSale: atualizar estoque: %w", err)
			}
		}
		return nil
	})
}

// CreateLossRecord registra uma perda de produção.
func (r *BakeryRepo) CreateLossRecord(ctx context.Context, loss *bakery.LossRecord) error {
	return r.db.WithTenant(ctx, loss.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO bakery_losses
			(tenant_id, product_id, quantity, unit, reason, cost_value, notes, recorded_by, recorded_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id`,
			loss.TenantID, loss.ProductID, loss.Quantity, loss.Unit,
			string(loss.Reason), loss.CostValue, nullString(loss.Notes),
			nullString(loss.RecordedBy), loss.RecordedAt,
		).Scan(&loss.ID)
	})
}

// GetLossByPeriod retorna perdas de um período para análise.
func (r *BakeryRepo) GetLossByPeriod(ctx context.Context, tenantID string, from, to time.Time) ([]*bakery.LossRecord, error) {
	var list []*bakery.LossRecord
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT l.id, l.product_id, p.name, l.quantity, l.unit,
			       l.reason, l.cost_value, COALESCE(l.notes,''), l.recorded_at
			FROM bakery_losses l
			JOIN bakery_products p ON p.id = l.product_id
			WHERE l.tenant_id = $1
			  AND l.recorded_at BETWEEN $2 AND $3
			ORDER BY l.recorded_at DESC`,
			tenantID, from, to)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l bakery.LossRecord
			l.TenantID = tenantID
			if err := rows.Scan(&l.ID, &l.ProductID, &l.ProductName,
				&l.Quantity, &l.Unit, &l.Reason, &l.CostValue,
				&l.Notes, &l.RecordedAt); err != nil {
				return err
			}
			list = append(list, &l)
		}
		return rows.Err()
	})
	return list, err
}

// =============================================================================
// NoOpScaleReader — balança simulada para desenvolvimento
// =============================================================================

// NoOpScaleReader simula uma balança para dev/testes.
type NoOpScaleReader struct{}

func NewNoOpScaleReader() *NoOpScaleReader { return &NoOpScaleReader{} }

func (n *NoOpScaleReader) ReadWeight(_ context.Context, scaleID string) (*bakery.ScaleReading, error) {
	// Simula leitura de 0.350 kg estável
	return &bakery.ScaleReading{
		ScaleID:  scaleID,
		PLUCode:  "P001", // pão francês
		WeightKG: 0.350,
		ReadAt:   time.Now(),
		IsStable: true,
	}, nil
}

func (n *NoOpScaleReader) IsConnected(_ context.Context, _ string) bool { return true }
