// Package postgres — repositório PostgreSQL de billing.
// Implementa billing.BillingRepository com RLS.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/billing"
)

// BillingRepo implementa billing.BillingRepository.
type BillingRepo struct {
	db *DB
}

func NewBillingRepo(db *DB) *BillingRepo { return &BillingRepo{db: db} }

// ── PLANOS ────────────────────────────────────────────────────────────────────

func (r *BillingRepo) ListPlans(ctx context.Context, activeOnly bool) ([]*billing.Plan, error) {
	query := `SELECT id, code, name, description, price_monthly, price_yearly,
		COALESCE(setup_fee,0), max_users, max_transactions, max_products,
		max_invoices, max_storage_mb, features, allowed_niches,
		display_order, is_active, is_featured
		FROM billing_plans`
	if activeOnly {
		query += " WHERE is_active = TRUE"
	}
	query += " ORDER BY display_order"

	rows, err := r.db.pool.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*billing.Plan
	for rows.Next() {
		var p billing.Plan
		var featuresJSON []byte
		var nichesJSON []byte
		if err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.Description,
			&p.PriceMonthly, &p.PriceYearly, &p.SetupFee,
			&p.MaxUsers, &p.MaxTransactions, &p.MaxProducts,
			&p.MaxInvoices, &p.MaxStorageMB,
			&featuresJSON, &nichesJSON,
			&p.DisplayOrder, &p.IsActive, &p.IsFeatured,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(featuresJSON, &p.Features)
		json.Unmarshal(nichesJSON, &p.AllowedNiches)
		plans = append(plans, &p)
	}
	return plans, rows.Err()
}

func (r *BillingRepo) GetPlan(ctx context.Context, planID string) (*billing.Plan, error) {
	return r.getPlanBy(ctx, "id", planID)
}

func (r *BillingRepo) GetPlanByCode(ctx context.Context, code string) (*billing.Plan, error) {
	return r.getPlanBy(ctx, "code", code)
}

func (r *BillingRepo) getPlanBy(ctx context.Context, field, value string) (*billing.Plan, error) {
	var p billing.Plan
	var featuresJSON, nichesJSON []byte
	err := r.db.pool.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT id, code, name, description, price_monthly, price_yearly,
		COALESCE(setup_fee,0), max_users, max_transactions, max_products,
		max_invoices, max_storage_mb, features, allowed_niches,
		display_order, is_active, is_featured
		FROM billing_plans WHERE %s = $1`, field), value).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.PriceMonthly, &p.PriceYearly, &p.SetupFee,
		&p.MaxUsers, &p.MaxTransactions, &p.MaxProducts,
		&p.MaxInvoices, &p.MaxStorageMB,
		&featuresJSON, &nichesJSON,
		&p.DisplayOrder, &p.IsActive, &p.IsFeatured,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plano não encontrado")
	}
	json.Unmarshal(featuresJSON, &p.Features)
	json.Unmarshal(nichesJSON, &p.AllowedNiches)
	return &p, err
}

func (r *BillingRepo) UpdatePlan(ctx context.Context, plan *billing.Plan) error {
	featuresJSON, _ := json.Marshal(plan.Features)
	nichesJSON, _ := json.Marshal(plan.AllowedNiches)
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE billing_plans SET
			name=$2, description=$3, price_monthly=$4, price_yearly=$5,
			setup_fee=$6, max_users=$7, max_transactions=$8, max_products=$9,
			max_invoices=$10, max_storage_mb=$11, features=$12,
			allowed_niches=$13, display_order=$14, is_active=$15, is_featured=$16
		WHERE id=$1`,
		plan.ID, plan.Name, plan.Description, plan.PriceMonthly, plan.PriceYearly,
		plan.SetupFee, plan.MaxUsers, plan.MaxTransactions, plan.MaxProducts,
		plan.MaxInvoices, plan.MaxStorageMB, featuresJSON, nichesJSON,
		plan.DisplayOrder, plan.IsActive, plan.IsFeatured,
	)
	return err
}

// ── ASSINATURAS ───────────────────────────────────────────────────────────────

func (r *BillingRepo) GetSubscription(ctx context.Context, tenantID string) (*billing.Subscription, error) {
	var s billing.Subscription
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, tenant_id, plan_id, plan_code, plan_name, status,
		trial_ends_at, current_period_start, current_period_end,
		billing_cycle, discount_percent, COALESCE(discount_reason,''),
		current_users, current_transactions, current_products, current_invoices,
		max_users, max_transactions, max_products, max_invoices, price
		FROM billing_subscriptions WHERE tenant_id=$1`, tenantID).Scan(
		&s.ID, &s.TenantID, &s.PlanID, &s.PlanCode, &s.PlanName, &s.Status,
		&s.TrialEndsAt, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.BillingCycle, &s.DiscountPercent, &s.DiscountReason,
		&s.CurrentUsers, &s.CurrentTransactions, &s.CurrentProducts, &s.CurrentInvoices,
		&s.MaxUsers, &s.MaxTransactions, &s.MaxProducts, &s.MaxInvoices, &s.Price,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("assinatura não encontrada")
	}
	return &s, err
}

func (r *BillingRepo) CreateSubscription(ctx context.Context, s *billing.Subscription) error {
	return r.db.pool.QueryRowContext(ctx, `
		INSERT INTO billing_subscriptions
		(tenant_id, plan_id, plan_code, plan_name, status,
		trial_ends_at, current_period_start, current_period_end,
		billing_cycle, discount_percent, discount_reason, price)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id`,
		s.TenantID, s.PlanID, s.PlanCode, s.PlanName, s.Status,
		s.TrialEndsAt, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.BillingCycle, s.DiscountPercent, s.DiscountReason, s.Price,
	).Scan(&s.ID)
}

func (r *BillingRepo) UpdateSubscription(ctx context.Context, s *billing.Subscription) error {
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE billing_subscriptions SET
			plan_id=$2, plan_code=$3, plan_name=$4, status=$5,
			trial_ends_at=$6, current_period_start=$7, current_period_end=$8,
			billing_cycle=$9, discount_percent=$10, discount_reason=$11, price=$12
		WHERE id=$1`,
		s.ID, s.PlanID, s.PlanCode, s.PlanName, s.Status,
		s.TrialEndsAt, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.BillingCycle, s.DiscountPercent, s.DiscountReason, s.Price,
	)
	return err
}

// ── USO ───────────────────────────────────────────────────────────────────────

func (r *BillingRepo) IncrementUsage(ctx context.Context, tenantID, metric string, delta int) error {
	col := usageColumn(metric)
	if col == "" {
		return nil
	}
	_, err := r.db.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE billing_subscriptions SET %s = %s + $2 WHERE tenant_id=$1", col, col),
		tenantID, delta)
	return err
}

func (r *BillingRepo) GetUsage(ctx context.Context, tenantID string) (map[string]int, error) {
	var users, txns, products, invoices int
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT current_users, current_transactions, current_products, current_invoices
		FROM billing_subscriptions WHERE tenant_id=$1`, tenantID).
		Scan(&users, &txns, &products, &invoices)
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"users": users, "transactions": txns,
		"products": products, "invoices": invoices,
	}, nil
}

func (r *BillingRepo) CheckLimit(ctx context.Context, tenantID, metric string) (bool, error) {
	col := usageColumn(metric)
	maxCol := "max_" + metric + "s"
	if col == "" {
		return true, nil
	}
	var current int
	var limit sql.NullInt64
	err := r.db.pool.QueryRowContext(ctx,
		fmt.Sprintf("SELECT %s, %s FROM billing_subscriptions WHERE tenant_id=$1", col, maxCol),
		tenantID).Scan(&current, &limit)
	if err != nil {
		return true, nil
	}
	if !limit.Valid {
		return true, nil // ilimitado
	}
	return current < int(limit.Int64), nil
}

// ── CUPONS ────────────────────────────────────────────────────────────────────

func (r *BillingRepo) GetCoupon(ctx context.Context, code string) (*billing.Coupon, error) {
	var c billing.Coupon
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, code, description, discount_type, discount_value,
		duration_months, valid_until, is_valid
		FROM billing_coupons WHERE code=$1`, code).Scan(
		&c.ID, &c.Code, &c.Description, &c.DiscountType, &c.DiscountValue,
		&c.DurationMonths, &c.ValidUntil, &c.IsValid,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cupom não encontrado")
	}
	return &c, err
}

func (r *BillingRepo) UseCoupon(ctx context.Context, code, tenantID string) error {
	_, err := r.db.pool.ExecContext(ctx,
		"INSERT INTO billing_coupon_uses (coupon_code, tenant_id, used_at) VALUES ($1,$2,$3)",
		code, tenantID, time.Now())
	return err
}

// ── FEATURES ──────────────────────────────────────────────────────────────────

func (r *BillingRepo) GetFeatures(ctx context.Context, tenantID string) (*billing.PlanFeatures, error) {
	var featuresJSON []byte
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT bp.features FROM billing_subscriptions bs
		JOIN billing_plans bp ON bp.id = bs.plan_id
		WHERE bs.tenant_id=$1`, tenantID).Scan(&featuresJSON)
	if err != nil {
		return nil, err
	}
	var f billing.PlanFeatures
	json.Unmarshal(featuresJSON, &f)
	return &f, nil
}

func usageColumn(metric string) string {
	switch metric {
	case "users":        return "current_users"
	case "transactions": return "current_transactions"
	case "products":     return "current_products"
	case "invoices":     return "current_invoices"
	default:             return ""
	}
}
