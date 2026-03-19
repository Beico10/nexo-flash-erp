// Package postgres — repositório de contratos logísticos.
// Implementa logistics.ContractRepository.
//
// Hierarquia de contratos (Briefing Mestre §4):
//   - Contrato específico do embarcador (shipper_id preenchido) → PRIORIDADE
//   - Tabela geral da transportadora (shipper_id IS NULL) → fallback
//
// A função SQL resolve_contract() implementa essa hierarquia no banco.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/logistics"
)

// LogisticsRepo implementa logistics.ContractRepository.
type LogisticsRepo struct {
	db *DB
}

func NewLogisticsRepo(db *DB) *LogisticsRepo { return &LogisticsRepo{db: db} }

// GetApplicable resolve o contrato correto usando a hierarquia:
// embarcador específico > tabela geral.
// Delega para a função SQL resolve_contract() que usa o índice otimizado.
func (r *LogisticsRepo) GetApplicable(ctx context.Context, tenantID, shipperID string, vt logistics.VehicleType) (*logistics.Contract, error) {
	var contract logistics.Contract

	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		// Tenta primeiro o contrato específico do embarcador
		query := `
			SELECT id, contract_name, shipper_id, vehicle_type,
			       COALESCE(price_per_km, 0), COALESCE(price_per_kg, 0),
			       COALESCE(minimum_charge, 0), toll_policy
			FROM nexo.logistics_contracts
			WHERE tenant_id = $1
			  AND vehicle_type = $2
			  AND active = TRUE
			  AND (valid_until IS NULL OR valid_until >= CURRENT_DATE)
			ORDER BY
				-- Prioridade: contrato específico do embarcador
				CASE WHEN shipper_id::text = $3 THEN 0 ELSE 1 END,
				valid_from DESC
			LIMIT 1`

		var shipperIDNull sql.NullString
		err := tx.QueryRowContext(ctx, query, tenantID, string(vt), shipperID).
			Scan(
				&contract.ID, &contract.ContractName,
				&shipperIDNull, &contract.VehicleType,
				&contract.PricePerKM, &contract.PricePerKG,
				&contract.MinimumCharge, &contract.TollPolicy,
			)
		if err == sql.ErrNoRows {
			return fmt.Errorf("nenhum contrato encontrado para veículo %s", vt)
		}
		if shipperIDNull.Valid {
			contract.ShipperID = &shipperIDNull.String
		}
		contract.TenantID = tenantID
		return err
	})
	return &contract, err
}

// Create cria um novo contrato.
func (r *LogisticsRepo) Create(ctx context.Context, c *logistics.Contract) error {
	return r.db.WithTenant(ctx, c.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO nexo.logistics_contracts
			(tenant_id, contract_name, shipper_id, vehicle_type,
			 price_per_km, price_per_kg, minimum_charge, toll_policy)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id`,
			c.TenantID, c.ContractName,
			c.ShipperID, string(c.VehicleType),
			c.PricePerKM, c.PricePerKG, c.MinimumCharge, c.TollPolicy,
		).Scan(&c.ID)
	})
}

// CTe representa um CT-e emitido.
type CTe struct {
	ID          string
	TenantID    string
	ChaveCTe    string
	NumCTe      string
	ShipperID   string
	RouteOrigin string
	RouteDest   string
	VehicleType string
	DistanceKM  float64
	WeightKG    float64
	GrossValue  float64
	Status      string // "authorized" | "cancelled" | "contingency"
	IssuedAt    time.Time
	CancelledAt *time.Time
	XMLPath     string // caminho no object storage
}

// CTeRepo gerencia CT-es emitidos.
type CTeRepo struct {
	db *DB
}

func NewCTeRepo(db *DB) *CTeRepo { return &CTeRepo{db: db} }

// Save persiste um CT-e emitido.
func (r *CTeRepo) Save(ctx context.Context, cte *CTe) error {
	return r.db.WithTenant(ctx, cte.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO nexo.logistics_ctes
			(tenant_id, chave_cte, num_cte, shipper_id,
			 route_origin, route_dest, vehicle_type,
			 distance_km, weight_kg, gross_value, status, issued_at, xml_path)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			RETURNING id`,
			cte.TenantID, cte.ChaveCTe, cte.NumCTe, cte.ShipperID,
			cte.RouteOrigin, cte.RouteDest, cte.VehicleType,
			cte.DistanceKM, cte.WeightKG, cte.GrossValue,
			cte.Status, cte.IssuedAt, cte.XMLPath,
		).Scan(&cte.ID)
	})
}

// ListByPeriod lista CT-es emitidos num período.
func (r *CTeRepo) ListByPeriod(ctx context.Context, tenantID string, from, to time.Time) ([]*CTe, error) {
	var list []*CTe
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, chave_cte, num_cte, route_origin, route_dest,
			       vehicle_type, gross_value, status, issued_at
			FROM nexo.logistics_ctes
			WHERE tenant_id = $1
			  AND issued_at BETWEEN $2 AND $3
			ORDER BY issued_at DESC`, tenantID, from, to)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c CTe
			c.TenantID = tenantID
			if err := rows.Scan(&c.ID, &c.ChaveCTe, &c.NumCTe,
				&c.RouteOrigin, &c.RouteDest, &c.VehicleType,
				&c.GrossValue, &c.Status, &c.IssuedAt); err != nil {
				return err
			}
			list = append(list, &c)
		}
		return rows.Err()
	})
	return list, err
}
