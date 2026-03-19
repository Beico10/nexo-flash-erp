// Package postgres — repositório PostgreSQL do módulo de Estética.
// Implementa aesthetics.AgendaRepository.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/modules/aesthetics"
)

// AestheticsRepo implementa aesthetics.AgendaRepository.
type AestheticsRepo struct {
	db *DB
}

func NewAestheticsRepo(db *DB) *AestheticsRepo {
	return &AestheticsRepo{db: db}
}

// Create insere um novo agendamento.
// A trava de conflito é garantida pela constraint EXCLUDE no banco
// (migration 002: EXCLUDE USING gist com tstzrange).
func (r *AestheticsRepo) Create(ctx context.Context, apt *aesthetics.Appointment) error {
	return r.db.WithTenant(ctx, apt.TenantID, func(tx *sql.Tx) error {
		query := `
			INSERT INTO nexo.aesthetics_appointments (
				tenant_id, professional_id, customer_id, customer_name,
				customer_phone, service_id, service_price,
				start_time, end_time, duration_min, status,
				notes, split_enabled, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
			RETURNING id`

		err := tx.QueryRowContext(ctx, query,
			apt.TenantID,
			apt.ProfessionalID,
			nullString(apt.CustomerID),
			apt.CustomerName,
			nullString(apt.CustomerPhone),
			apt.ServiceID,
			apt.ServicePrice,
			apt.StartTime,
			apt.EndTime,
			apt.DurationMin,
			string(apt.Status),
			nullString(apt.Notes),
			apt.SplitEnabled,
		).Scan(&apt.ID)

		// Se for erro de constraint EXCLUDE = conflito de agenda
		if err != nil && isConflictError(err) {
			return &aesthetics.ConflictError{
				ProfessionalID:  apt.ProfessionalID,
				ConflictingTime: apt.StartTime,
			}
		}
		return err
	})
}

// Update atualiza um agendamento existente.
func (r *AestheticsRepo) Update(ctx context.Context, apt *aesthetics.Appointment) error {
	return r.db.WithTenant(ctx, apt.TenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE nexo.aesthetics_appointments SET
				start_time  = $3,
				end_time    = $4,
				status      = $5,
				notes       = $6
			WHERE id = $1 AND tenant_id = $2`,
			apt.ID, apt.TenantID,
			apt.StartTime, apt.EndTime,
			string(apt.Status), nullString(apt.Notes),
		)
		if err != nil && isConflictError(err) {
			return &aesthetics.ConflictError{
				ProfessionalID:  apt.ProfessionalID,
				ConflictingTime: apt.StartTime,
			}
		}
		return err
	})
}

// GetByID busca um agendamento pelo ID.
func (r *AestheticsRepo) GetByID(ctx context.Context, tenantID, id string) (*aesthetics.Appointment, error) {
	var apt aesthetics.Appointment
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, professional_id,
			       COALESCE(customer_id::text,''), customer_name,
			       COALESCE(customer_phone,''), service_id, service_price,
			       start_time, end_time, duration_min, status,
			       COALESCE(notes,''), split_enabled, created_at
			FROM nexo.aesthetics_appointments
			WHERE id = $1 AND tenant_id = $2`, id, tenantID).
			Scan(
				&apt.ID, &apt.TenantID, &apt.ProfessionalID,
				&apt.CustomerID, &apt.CustomerName,
				&apt.CustomerPhone, &apt.ServiceID, &apt.ServicePrice,
				&apt.StartTime, &apt.EndTime, &apt.DurationMin,
				&apt.Status, &apt.Notes, &apt.SplitEnabled, &apt.CreatedAt,
			)
	})
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agendamento %s não encontrado", id)
	}
	return &apt, err
}

// FindConflicts busca agendamentos que se sobrepõem ao intervalo dado.
// Usa o índice GiST da constraint EXCLUDE para performance.
func (r *AestheticsRepo) FindConflicts(ctx context.Context, tenantID, professionalID string, start, end time.Time, excludeID string) ([]*aesthetics.Appointment, error) {
	var list []*aesthetics.Appointment
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, start_time, end_time, customer_name, status
			FROM nexo.aesthetics_appointments
			WHERE tenant_id       = $1
			  AND professional_id = $2
			  AND status NOT IN ('cancelled','no_show')
			  AND tstzrange(start_time, end_time) && tstzrange($3, $4)
			  AND ($5 = '' OR id::text != $5)`,
			tenantID, professionalID, start, end, excludeID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var apt aesthetics.Appointment
			if err := rows.Scan(&apt.ID, &apt.StartTime, &apt.EndTime,
				&apt.CustomerName, &apt.Status); err != nil {
				return err
			}
			list = append(list, &apt)
		}
		return rows.Err()
	})
	return list, err
}

// ListByProfessional lista agendamentos de um profissional em um dia.
func (r *AestheticsRepo) ListByProfessional(ctx context.Context, tenantID, professionalID string, date time.Time) ([]*aesthetics.Appointment, error) {
	return r.listByDate(ctx, tenantID, date, professionalID)
}

// ListByDate lista todos os agendamentos do tenant em um dia.
func (r *AestheticsRepo) ListByDate(ctx context.Context, tenantID string, date time.Time) ([]*aesthetics.Appointment, error) {
	return r.listByDate(ctx, tenantID, date, "")
}

func (r *AestheticsRepo) listByDate(ctx context.Context, tenantID string, date time.Time, professionalID string) ([]*aesthetics.Appointment, error) {
	var list []*aesthetics.Appointment
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		query := `
			SELECT id, tenant_id, professional_id,
			       COALESCE(customer_id::text,''), customer_name,
			       COALESCE(customer_phone,''), service_id, service_price,
			       start_time, end_time, duration_min, status, created_at
			FROM nexo.aesthetics_appointments
			WHERE tenant_id = $1
			  AND start_time >= $2 AND start_time < $3
			  AND ($4 = '' OR professional_id::text = $4)
			ORDER BY start_time`

		rows, err := tx.QueryContext(ctx, query, tenantID, startOfDay, endOfDay, professionalID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var apt aesthetics.Appointment
			if err := rows.Scan(
				&apt.ID, &apt.TenantID, &apt.ProfessionalID,
				&apt.CustomerID, &apt.CustomerName, &apt.CustomerPhone,
				&apt.ServiceID, &apt.ServicePrice,
				&apt.StartTime, &apt.EndTime, &apt.DurationMin,
				&apt.Status, &apt.CreatedAt,
			); err != nil {
				return err
			}
			list = append(list, &apt)
		}
		return rows.Err()
	})
	return list, err
}

// isConflictError detecta erros de constraint EXCLUDE do PostgreSQL.
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	// Código PostgreSQL 23P01 = exclusion_violation
	return len(err.Error()) > 0 &&
		(contains(err.Error(), "23P01") ||
			contains(err.Error(), "no_double_booking") ||
			contains(err.Error(), "exclusion constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())
}
