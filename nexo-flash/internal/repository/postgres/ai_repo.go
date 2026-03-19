// Package postgres — repositório de sugestões de IA (Human-in-the-Loop).
// Implementa ai.ApprovalRepository.
//
// REGRA CRÍTICA: Nenhuma IA persiste dados diretamente.
// Toda sugestão entra com status='pending'.
// Apenas Approve() muda o status para 'approved' e persiste a ação.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexoflash/nexo-flash/internal/ai"
)

// AIRepo implementa ai.ApprovalRepository.
type AIRepo struct {
	db *DB
}

func NewAIRepo(db *DB) *AIRepo { return &AIRepo{db: db} }

// CreatePending insere uma sugestão com status='pending'.
// Único ponto de entrada para sugestões de IA.
func (r *AIRepo) CreatePending(ctx context.Context, s *ai.Suggestion) error {
	data, err := json.Marshal(s.SuggestedData)
	if err != nil {
		return fmt.Errorf("AIRepo.CreatePending: marshal: %w", err)
	}

	return r.db.WithTenant(ctx, s.TenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO ai_suggestions (
				id, tenant_id, suggestion_type, target_table, target_id,
				suggested_data, reason, confidence, created_by_ai,
				status, created_at, expires_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',$10,$11)`,
			s.ID, s.TenantID, string(s.Type),
			s.TargetTable, nullString(s.TargetID),
			data, s.Reason, s.Confidence, s.CreatedByAI,
			s.CreatedAt, s.ExpiresAt,
		)
		return err
	})
}

// GetPending retorna todas as sugestões pendentes do tenant, ordenadas por urgência.
func (r *AIRepo) GetPending(ctx context.Context, tenantID string) ([]*ai.Suggestion, error) {
	var list []*ai.Suggestion

	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, tenant_id, suggestion_type, target_table,
			       COALESCE(target_id::text,''), suggested_data,
			       reason, confidence, created_by_ai,
			       status, created_at, expires_at
			FROM ai_suggestions
			WHERE tenant_id = $1
			  AND status = 'pending'
			  AND expires_at > NOW()
			ORDER BY confidence DESC, created_at ASC`, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var s ai.Suggestion
			var rawData []byte
			if err := rows.Scan(
				&s.ID, &s.TenantID, &s.Type, &s.TargetTable,
				&s.TargetID, &rawData,
				&s.Reason, &s.Confidence, &s.CreatedByAI,
				&s.Status, &s.CreatedAt, &s.ExpiresAt,
			); err != nil {
				return err
			}
			if err := json.Unmarshal(rawData, &s.SuggestedData); err != nil {
				return err
			}
			list = append(list, &s)
		}
		return rows.Err()
	})
	return list, err
}

// GetByID busca uma sugestão pelo ID.
func (r *AIRepo) GetByID(ctx context.Context, tenantID, id string) (*ai.Suggestion, error) {
	var s ai.Suggestion
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		var rawData []byte
		err := tx.QueryRowContext(ctx, `
			SELECT id, tenant_id, suggestion_type, target_table,
			       COALESCE(target_id::text,''), suggested_data,
			       reason, confidence, created_by_ai,
			       status, created_at, expires_at
			FROM ai_suggestions
			WHERE id = $1 AND tenant_id = $2`, id, tenantID).
			Scan(
				&s.ID, &s.TenantID, &s.Type, &s.TargetTable,
				&s.TargetID, &rawData,
				&s.Reason, &s.Confidence, &s.CreatedByAI,
				&s.Status, &s.CreatedAt, &s.ExpiresAt,
			)
		if err == sql.ErrNoRows {
			return fmt.Errorf("sugestão %s não encontrada", id)
		}
		if err != nil {
			return err
		}
		return json.Unmarshal(rawData, &s.SuggestedData)
	})
	return &s, err
}

// Approve marca como aprovada e registra quem aprovou.
// HUMAN-IN-THE-LOOP: este é o único método que permite a uma sugestão
// de IA se tornar uma ação real no sistema.
func (r *AIRepo) Approve(ctx context.Context, suggestionID, approvedByUserID string) error {
	// Precisamos do tenant_id para usar WithTenant — buscamos sem RLS
	// (a política de auditoria cobre isso)
	row := r.db.pool.QueryRowContext(ctx,
		`SELECT tenant_id FROM ai_suggestions WHERE id = $1 AND status = 'pending'`,
		suggestionID)
	var tenantID string
	if err := row.Scan(&tenantID); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("sugestão %s não encontrada ou já processada", suggestionID)
		}
		return err
	}

	now := time.Now().UTC()
	return r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
			UPDATE ai_suggestions SET
				status      = 'approved',
				approved_by = $3,
				approved_at = $4
			WHERE id = $1 AND tenant_id = $2 AND status = 'pending'`,
			suggestionID, tenantID, approvedByUserID, now)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("sugestão %s não pôde ser aprovada (já processada?)", suggestionID)
		}
		return nil
	})
}

// Reject rejeita uma sugestão sem nenhuma alteração nos dados de negócio.
func (r *AIRepo) Reject(ctx context.Context, suggestionID, userID, reason string) error {
	row := r.db.pool.QueryRowContext(ctx,
		`SELECT tenant_id FROM ai_suggestions WHERE id = $1`, suggestionID)
	var tenantID string
	if err := row.Scan(&tenantID); err != nil {
		return fmt.Errorf("sugestão não encontrada: %w", err)
	}

	return r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE ai_suggestions SET
				status           = 'rejected',
				rejection_reason = $3,
				approved_by      = $4,
				approved_at      = NOW()
			WHERE id = $1 AND tenant_id = $2 AND status = 'pending'`,
			suggestionID, tenantID, reason, userID)
		return err
	})
}

// ExpireStale expira sugestões que passaram do prazo (job periódico).
func (r *AIRepo) ExpireStale(ctx context.Context) (int64, error) {
	result, err := r.db.pool.ExecContext(ctx, `
		UPDATE ai_suggestions SET status = 'expired'
		WHERE status = 'pending' AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
