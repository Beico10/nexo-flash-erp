// Package postgres — repositório de pagamentos BaaS (PIX e Boleto).
// Implementa baas.PaymentRepository.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/nexoone/nexo-one/internal/baas"
)

// PaymentRepo implementa baas.PaymentRepository.
type PaymentRepo struct {
	db *DB
}

func NewPaymentRepo(db *DB) *PaymentRepo { return &PaymentRepo{db: db} }

// SavePixCharge persiste uma cobrança PIX.
func (r *PaymentRepo) SavePixCharge(ctx context.Context, charge *baas.PixCharge) error {
	splitJSON, _ := json.Marshal(charge.SplitRecipients)

	return r.db.WithTenant(ctx, charge.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO nexo.baas_pix_charges
			(tenant_id, tx_id, amount, description, payer_name, payer_document,
			 expires_at, status, qr_code_text, qr_code_image, split_data, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			RETURNING id`,
			charge.TenantID, charge.TxID, charge.Amount,
			nullString(charge.Description), nullString(charge.PayerName),
			nullString(charge.PayerDocument), charge.ExpiresAt,
			string(charge.Status),
			nullString(charge.QRCodeText), nullString(charge.QRCodeImage),
			splitJSON, charge.CreatedAt,
		).Scan(&charge.ID)
	})
}

// UpdatePixStatus atualiza o status de uma cobrança PIX (após webhook).
func (r *PaymentRepo) UpdatePixStatus(ctx context.Context, tenantID, txID string, status baas.PaymentStatus, paidAt *time.Time) error {
	// Buscar tenant_id via txID se não fornecido
	if tenantID == "" {
		r.db.pool.QueryRowContext(ctx,
			`SELECT tenant_id FROM nexo.baas_pix_charges WHERE tx_id = $1`, txID).
			Scan(&tenantID)
	}
	if tenantID == "" {
		return nil // não encontrado — ignorar webhook
	}

	return r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE nexo.baas_pix_charges SET
				status  = $3,
				paid_at = $4
			WHERE tx_id = $1 AND tenant_id = $2`,
			txID, tenantID, string(status), nullTime(paidAt))
		return err
	})
}

// SaveBoleto persiste um boleto.
func (r *PaymentRepo) SaveBoleto(ctx context.Context, boleto *baas.BoletoCharge) error {
	return r.db.WithTenant(ctx, boleto.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO nexo.baas_boletos
			(tenant_id, our_number, amount, due_date, description,
			 payer_name, payer_document, payer_address,
			 status, bar_code, digitable_line, qr_code_text, pdf_url, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			RETURNING id`,
			boleto.TenantID, boleto.OurNumber, boleto.Amount, boleto.DueDate,
			nullString(boleto.Description), boleto.PayerName, boleto.PayerDocument,
			nullString(boleto.PayerAddress), string(boleto.Status),
			nullString(boleto.BarCode), nullString(boleto.DigitableLine),
			nullString(boleto.QRCodeText), nullString(boleto.PDFURL),
			boleto.CreatedAt,
		).Scan(&boleto.ID)
	})
}

// UpdateBoletoStatus atualiza o status de um boleto.
func (r *PaymentRepo) UpdateBoletoStatus(ctx context.Context, tenantID, ourNumber string, status baas.PaymentStatus, paidAt *time.Time) error {
	if tenantID == "" {
		r.db.pool.QueryRowContext(ctx,
			`SELECT tenant_id FROM nexo.baas_boletos WHERE our_number = $1`, ourNumber).
			Scan(&tenantID)
	}
	return r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			UPDATE nexo.baas_boletos SET status = $3, paid_at = $4
			WHERE our_number = $1 AND tenant_id = $2`,
			ourNumber, tenantID, string(status), nullTime(paidAt))
		return err
	})
}

// LogWebhook registra o payload bruto do webhook para auditoria.
func (r *PaymentRepo) LogWebhook(ctx context.Context, tenantID, eventType, chargeID, txID string, amount float64, paidAt time.Time, rawPayload []byte) error {
	raw, _ := json.RawMessage(rawPayload).MarshalJSON()
	_, err := r.db.pool.ExecContext(ctx, `
		INSERT INTO nexo.baas_webhook_events
		(tenant_id, event_type, charge_id, tx_id, amount, paid_at, raw_payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		nullString(tenantID), eventType, nullString(chargeID),
		nullString(txID), amount, paidAt, raw)
	return err
}

// GetPendingPix retorna cobranças PIX pendentes que estão prestes a expirar.
// Usado por job periódico para alertar sobre cobranças não pagas.
func (r *PaymentRepo) GetPendingPix(ctx context.Context, tenantID string, expiringBefore time.Time) ([]*baas.PixCharge, error) {
	var list []*baas.PixCharge
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, tx_id, amount, description, payer_name, expires_at, status
			FROM nexo.baas_pix_charges
			WHERE tenant_id = $1
			  AND status = 'pending'
			  AND expires_at < $2
			ORDER BY expires_at ASC`, tenantID, expiringBefore)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c baas.PixCharge
			c.TenantID = tenantID
			var desc, payer sql.NullString
			if err := rows.Scan(&c.ID, &c.TxID, &c.Amount, &desc, &payer, &c.ExpiresAt, &c.Status); err != nil {
				return err
			}
			c.Description = desc.String
			c.PayerName = payer.String
			list = append(list, &c)
		}
		return rows.Err()
	})
	return list, err
}
