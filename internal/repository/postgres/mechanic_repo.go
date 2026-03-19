// Package postgres — repositório PostgreSQL do módulo de Mecânica.
// Implementa mechanic.OSRepository.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/mechanic"
)

// MechanicRepo implementa mechanic.OSRepository usando PostgreSQL + RLS.
type MechanicRepo struct {
	db *DB
}

func NewMechanicRepo(db *DB) *MechanicRepo {
	return &MechanicRepo{db: db}
}

// Create insere uma nova Ordem de Serviço.
func (r *MechanicRepo) Create(ctx context.Context, os *mechanic.ServiceOrder) error {
	return r.db.WithTenant(ctx, os.TenantID, func(tx *sql.Tx) error {
		query := `
			INSERT INTO mechanic_service_orders (
				id, tenant_id, number, vehicle_plate, vehicle_km, vehicle_model,
				vehicle_year, customer_id, customer_phone, status, complaint,
				created_at, updated_at
			) VALUES (
				gen_random_uuid(), $1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10,
				NOW(), NOW()
			) RETURNING id, created_at`

		return tx.QueryRowContext(ctx, query,
			os.TenantID, os.Number, os.VehiclePlate, os.VehicleKM, os.VehicleModel,
			os.VehicleYear, nullString(os.CustomerID), os.CustomerPhone, string(os.Status), os.Complaint,
		).Scan(&os.ID, &os.CreatedAt)
	})
}

// Update atualiza uma OS existente.
func (r *MechanicRepo) Update(ctx context.Context, os *mechanic.ServiceOrder) error {
	return r.db.WithTenant(ctx, os.TenantID, func(tx *sql.Tx) error {
		query := `
			UPDATE mechanic_service_orders SET
				status         = $3,
				diagnosis      = $4,
				approval_token = $5,
				approval_url   = $6,
				approved_at    = $7,
				total_parts    = $8,
				total_labor    = $9,
				total_amount   = $10,
				updated_at     = NOW()
			WHERE id = $1 AND tenant_id = $2`

		_, err := tx.ExecContext(ctx, query,
			os.ID, os.TenantID,
			string(os.Status), nullString(os.Diagnosis),
			nullString(os.ApprovalToken), nullString(os.ApprovalURL),
			nullTime(os.ApprovedAt),
			totalParts(os), totalLabor(os), totalOS(os),
		)
		return err
	})
}

// GetByID busca uma OS pelo ID, incluindo peças e mão de obra.
func (r *MechanicRepo) GetByID(ctx context.Context, tenantID, id string) (*mechanic.ServiceOrder, error) {
	var os mechanic.ServiceOrder
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		query := `
			SELECT id, tenant_id, number, vehicle_plate, vehicle_km, vehicle_model,
			       vehicle_year, COALESCE(customer_id::text,''), customer_phone,
			       status, COALESCE(complaint,''), COALESCE(diagnosis,''),
			       COALESCE(approval_token,''), COALESCE(approval_url,''),
			       approved_at, total_parts, total_labor, total_amount,
			       created_at, updated_at
			FROM mechanic_service_orders
			WHERE id = $1 AND tenant_id = $2`

		var approvedAt sql.NullTime
		err := tx.QueryRowContext(ctx, query, id, tenantID).Scan(
			&os.ID, &os.TenantID, &os.Number, &os.VehiclePlate, &os.VehicleKM,
			&os.VehicleModel, &os.VehicleYear, &os.CustomerID, &os.CustomerPhone,
			&os.Status, &os.Complaint, &os.Diagnosis,
			&os.ApprovalToken, &os.ApprovalURL,
			&approvedAt, new(float64), new(float64), new(float64),
			&os.CreatedAt, &os.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			return fmt.Errorf("OS %s não encontrada", id)
		}
		if approvedAt.Valid {
			os.ApprovedAt = &approvedAt.Time
		}

		// Buscar peças
		parts, err2 := r.loadParts(ctx, tx, tenantID, id)
		if err2 != nil {
			return err2
		}
		os.Parts = parts

		// Buscar mão de obra
		labor, err3 := r.loadLabor(ctx, tx, tenantID, id)
		if err3 != nil {
			return err3
		}
		os.LaborItems = labor
		return err
	})
	return &os, err
}

// GetByPlate retorna histórico de OSs por placa do veículo.
func (r *MechanicRepo) GetByPlate(ctx context.Context, tenantID, plate string) ([]*mechanic.ServiceOrder, error) {
	var list []*mechanic.ServiceOrder
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, number, vehicle_plate, vehicle_model, vehicle_km,
			       status, total_amount, created_at
			FROM mechanic_service_orders
			WHERE tenant_id = $1 AND vehicle_plate = $2
			ORDER BY created_at DESC
			LIMIT 50`, tenantID, plate)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var os mechanic.ServiceOrder
			if err := rows.Scan(&os.ID, &os.Number, &os.VehiclePlate,
				&os.VehicleModel, &os.VehicleKM, &os.Status,
				new(float64), &os.CreatedAt); err != nil {
				return err
			}
			os.TenantID = tenantID
			list = append(list, &os)
		}
		return rows.Err()
	})
	return list, err
}

// ListOpen retorna todas as OSs não concluídas do tenant.
func (r *MechanicRepo) ListOpen(ctx context.Context, tenantID string) ([]*mechanic.ServiceOrder, error) {
	var list []*mechanic.ServiceOrder
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, number, vehicle_plate, vehicle_model, vehicle_km,
			       COALESCE(customer_id::text,''), customer_phone,
			       status, COALESCE(complaint,''), total_amount, created_at
			FROM mechanic_service_orders
			WHERE tenant_id = $1
			  AND status NOT IN ('done','invoiced')
			ORDER BY created_at DESC`, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var os mechanic.ServiceOrder
			if err := rows.Scan(
				&os.ID, &os.Number, &os.VehiclePlate, &os.VehicleModel, &os.VehicleKM,
				&os.CustomerID, &os.CustomerPhone,
				&os.Status, &os.Complaint, new(float64), &os.CreatedAt,
			); err != nil {
				return err
			}
			os.TenantID = tenantID
			list = append(list, &os)
		}
		return rows.Err()
	})
	return list, err
}

// loadParts carrega as peças de uma OS (deve ser chamado dentro de WithTenant).
func (r *MechanicRepo) loadParts(ctx context.Context, tx *sql.Tx, tenantID, osID string) ([]mechanic.OSPart, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, COALESCE(part_code,''), description, quantity,
		       unit_cost, unit_price, total_price, COALESCE(ncm_code,'')
		FROM mechanic_os_parts
		WHERE os_id = $1 AND tenant_id = $2`, osID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []mechanic.OSPart
	for rows.Next() {
		var p mechanic.OSPart
		if err := rows.Scan(&p.ID, &p.PartCode, &p.Description, &p.Quantity,
			&p.UnitCost, &p.UnitPrice, &p.TotalPrice, &p.NCMCode); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}

// loadLabor carrega os itens de mão de obra de uma OS.
func (r *MechanicRepo) loadLabor(ctx context.Context, tx *sql.Tx, tenantID, osID string) ([]mechanic.OSLabor, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, description, hours, hourly_rate, total_price,
		       COALESCE(technician_id::text,'')
		FROM mechanic_os_labor
		WHERE os_id = $1 AND tenant_id = $2`, osID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var labor []mechanic.OSLabor
	for rows.Next() {
		var l mechanic.OSLabor
		if err := rows.Scan(&l.ID, &l.Description, &l.Hours,
			&l.HourlyRate, &l.TotalPrice, &l.TechnicianID); err != nil {
			return nil, err
		}
		labor = append(labor, l)
	}
	return labor, rows.Err()
}

// AddPart adiciona uma peça a uma OS.
func (r *MechanicRepo) AddPart(ctx context.Context, tenantID, osID string, part mechanic.OSPart) error {
	return r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO mechanic_os_parts
			(id, tenant_id, os_id, part_code, description, quantity, unit_cost, unit_price, total_price, ncm_code)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			tenantID, osID, nullString(part.PartCode), part.Description,
			part.Quantity, part.UnitCost, part.UnitPrice, part.TotalPrice, nullString(part.NCMCode))
		return err
	})
}

// GetByApprovalToken busca uma OS pelo token de aprovação WhatsApp.
func (r *MechanicRepo) GetByApprovalToken(ctx context.Context, token string) (*mechanic.ServiceOrder, error) {
	// Esta query não usa tenant_id pois o token é o autenticador
	// O tenant_id é resolvido a partir do token no banco
	row := r.db.pool.QueryRowContext(ctx, `
		SELECT id, tenant_id FROM mechanic_service_orders
		WHERE approval_token = $1 AND status = 'await_approval'`, token)
	var id, tenantID string
	if err := row.Scan(&id, &tenantID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token inválido ou expirado")
		}
		return nil, err
	}
	return r.GetByID(ctx, tenantID, id)
}

// helpers
func nullString(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
func totalParts(os *mechanic.ServiceOrder) float64 {
	var t float64
	for _, p := range os.Parts {
		t += p.TotalPrice
	}
	return t
}
func totalLabor(os *mechanic.ServiceOrder) float64 {
	var t float64
	for _, l := range os.LaborItems {
		t += l.TotalPrice
	}
	return t
}
func totalOS(os *mechanic.ServiceOrder) float64 { return totalParts(os) + totalLabor(os) }

// MarshalJSON helper para serialização de OS completa
func marshalJSON(v any) ([]byte, error) { return json.Marshal(v) }
