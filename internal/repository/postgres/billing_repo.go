// Package postgres — repositório de billing (planos e assinaturas).
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nexoone/nexo-one/internal/billing"
)

// BillingRepo implementa billing.BillingRepository.
type BillingRepo struct {
	db *DB
}

func NewBillingRepo(db *DB) *BillingRepo {
	return &BillingRepo{db: db}
}

// ════════════════════════════════════════════════════════════
// PLANOS
// ════════════════════════════════════════════════════════════

func (r *BillingRepo) ListPlans(ctx context.Context, activeOnly bool) ([]*billing.Plan, error) {
	query := `
		SELECT id, code, name, description, price_monthly, price_yearly, setup_fee,
		       max_users, max_transactions, max_products, max_invoices, max_storage_mb,
		       features, allowed_niches, display_order, is_active, is_featured
		FROM nexo.billing_plans
		WHERE ($1 = FALSE OR is_active = TRUE)
		ORDER BY display_order`

	rows, err := r.db.pool.QueryContext(ctx, query, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("BillingRepo.ListPlans: %w", err)
	}
	defer rows.Close()

	var plans []*billing.Plan
	for rows.Next() {
		p, err := r.scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *BillingRepo) GetPlan(ctx context.Context, planID string) (*billing.Plan, error) {
	row := r.db.pool.QueryRowContext(ctx, `
		SELECT id, code, name, description, price_monthly, price_yearly, setup_fee,
		       max_users, max_transactions, max_products, max_invoices, max_storage_mb,
		       features, allowed_niches, display_order, is_active, is_featured
		FROM nexo.billing_plans WHERE id = $1`, planID)
	return r.scanPlanRow(row)
}

func (r *BillingRepo) GetPlanByCode(ctx context.Context, code string) (*billing.Plan, error) {
	row := r.db.pool.QueryRowContext(ctx, `
		SELECT id, code, name, description, price_monthly, price_yearly, setup_fee,
		       max_users, max_transactions, max_products, max_invoices, max_storage_mb,
		       features, allowed_niches, display_order, is_active, is_featured
		FROM nexo.billing_plans WHERE code = $1 AND is_active = TRUE`, code)
	return r.scanPlanRow(row)
}

func (r *BillingRepo) UpdatePlan(ctx context.Context, plan *billing.Plan) error {
	features, _ := json.Marshal(plan.Features)
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE nexo.billing_plans SET
			name = $2, description = $3, price_monthly = $4, price_yearly = $5,
			setup_fee = $6, max_users = $7, max_transactions = $8, max_products = $9,
			max_invoices = $10, max_storage_mb = $11, features = $12,
			is_active = $13, is_featured = $14, updated_at = NOW()
		WHERE id = $1`,
		plan.ID, plan.Name, plan.Description, plan.PriceMonthly, plan.PriceYearly,
		plan.SetupFee, plan.MaxUsers, plan.MaxTransactions, plan.MaxProducts,
		plan.MaxInvoices, plan.MaxStorageMB, features, plan.IsActive, plan.IsFeatured)
	return err
}

func (r *BillingRepo) scanPlan(rows *sql.Rows) (*billing.Plan, error) {
	var p billing.Plan
	var featuresJSON, nichesJSON []byte
	var priceYearly, setupFee sql.NullFloat64
	var maxUsers, maxTx, maxProd, maxInv, maxStorage sql.NullInt64

	err := rows.Scan(
		&p.ID, &p.Code, &p.Name, &p.Description, &p.PriceMonthly, &priceYearly, &setupFee,
		&maxUsers, &maxTx, &maxProd, &maxInv, &maxStorage,
		&featuresJSON, &nichesJSON, &p.DisplayOrder, &p.IsActive, &p.IsFeatured,
	)
	if err != nil {
		return nil, err
	}

	if priceYearly.Valid {
		p.PriceYearly = priceYearly.Float64
	}
	if setupFee.Valid {
		p.SetupFee = setupFee.Float64
	}
	if maxUsers.Valid {
		v := int(maxUsers.Int64)
		p.MaxUsers = &v
	}
	if maxTx.Valid {
		v := int(maxTx.Int64)
		p.MaxTransactions = &v
	}
	if maxProd.Valid {
		v := int(maxProd.Int64)
		p.MaxProducts = &v
	}
	if maxInv.Valid {
		v := int(maxInv.Int64)
		p.MaxInvoices = &v
	}
	if maxStorage.Valid {
		v := int(maxStorage.Int64)
		p.MaxStorageMB = &v
	}

	json.Unmarshal(featuresJSON, &p.Features)
	json.Unmarshal(nichesJSON, &p.AllowedNiches)

	return &p, nil
}

func (r *BillingRepo) scanPlanRow(row *sql.Row) (*billing.Plan, error) {
	var p billing.Plan
	var featuresJSON, nichesJSON []byte
	var priceYearly, setupFee sql.NullFloat64
	var maxUsers, maxTx, maxProd, maxInv, maxStorage sql.NullInt64

	err := row.Scan(
		&p.ID, &p.Code, &p.Name, &p.Description, &p.PriceMonthly, &priceYearly, &setupFee,
		&maxUsers, &maxTx, &maxProd, &maxInv, &maxStorage,
		&featuresJSON, &nichesJSON, &p.DisplayOrder, &p.IsActive, &p.IsFeatured,
	)
	if err == sql.ErrNoRows {
		return nil, billing.ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}

	if priceYearly.Valid {
		p.PriceYearly = priceYearly.Float64
	}
	if setupFee.Valid {
		p.SetupFee = setupFee.Float64
	}
	if maxUsers.Valid {
		v := int(maxUsers.Int64)
		p.MaxUsers = &v
	}
	if maxTx.Valid {
		v := int(maxTx.Int64)
		p.MaxTransactions = &v
	}
	if maxProd.Valid {
		v := int(maxProd.Int64)
		p.MaxProducts = &v
	}
	if maxInv.Valid {
		v := int(maxInv.Int64)
		p.MaxInvoices = &v
	}
	if maxStorage.Valid {
		v := int(maxStorage.Int64)
		p.MaxStorageMB = &v
	}

	json.Unmarshal(featuresJSON, &p.Features)
	json.Unmarshal(nichesJSON, &p.AllowedNiches)

	return &p, nil
}

// ════════════════════════════════════════════════════════════
// ASSINATURAS
// ════════════════════════════════════════════════════════════

func (r *BillingRepo) GetSubscription(ctx context.Context, tenantID string) (*billing.Subscription, error) {
	var s billing.Subscription
	var trialEnds sql.NullTime
	var maxUsers, maxTx, maxProd, maxInv sql.NullInt64
	var discountReason sql.NullString

	err := r.db.pool.QueryRowContext(ctx, `
		SELECT bs.id, bs.tenant_id, bs.plan_id, bp.code, bp.name, bs.status,
		       bs.trial_ends_at, bs.current_period_start, bs.current_period_end,
		       bs.billing_cycle, bs.discount_percent, bs.discount_reason,
		       bs.current_users, bs.current_transactions, bs.current_products, bs.current_invoices,
		       bp.max_users, bp.max_transactions, bp.max_products, bp.max_invoices,
		       CASE WHEN bs.billing_cycle = 'yearly' THEN bp.price_yearly ELSE bp.price_monthly END
		FROM nexo.billing_subscriptions bs
		JOIN nexo.billing_plans bp ON bp.id = bs.plan_id
		WHERE bs.tenant_id = $1`, tenantID).Scan(
		&s.ID, &s.TenantID, &s.PlanID, &s.PlanCode, &s.PlanName, &s.Status,
		&trialEnds, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.BillingCycle, &s.DiscountPercent, &discountReason,
		&s.CurrentUsers, &s.CurrentTransactions, &s.CurrentProducts, &s.CurrentInvoices,
		&maxUsers, &maxTx, &maxProd, &maxInv, &s.Price,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("assinatura não encontrada")
	}
	if err != nil {
		return nil, err
	}

	if trialEnds.Valid {
		s.TrialEndsAt = &trialEnds.Time
	}
	if discountReason.Valid {
		s.DiscountReason = discountReason.String
	}
	if maxUsers.Valid {
		v := int(maxUsers.Int64)
		s.MaxUsers = &v
	}
	if maxTx.Valid {
		v := int(maxTx.Int64)
		s.MaxTransactions = &v
	}
	if maxProd.Valid {
		v := int(maxProd.Int64)
		s.MaxProducts = &v
	}
	if maxInv.Valid {
		v := int(maxInv.Int64)
		s.MaxInvoices = &v
	}

	return &s, nil
}

func (r *BillingRepo) CreateSubscription(ctx context.Context, sub *billing.Subscription) error {
	return r.db.pool.QueryRowContext(ctx, `
		INSERT INTO nexo.billing_subscriptions 
		(tenant_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, billing_cycle)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		sub.TenantID, sub.PlanID, sub.Status, sub.TrialEndsAt,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.BillingCycle,
	).Scan(&sub.ID)
}

func (r *BillingRepo) UpdateSubscription(ctx context.Context, sub *billing.Subscription) error {
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE nexo.billing_subscriptions SET
			plan_id = $2, status = $3, trial_ends_at = $4,
			current_period_start = $5, current_period_end = $6,
			billing_cycle = $7, discount_percent = $8, discount_reason = $9,
			updated_at = NOW()
		WHERE tenant_id = $1`,
		sub.TenantID, sub.PlanID, sub.Status, sub.TrialEndsAt,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.BillingCycle, sub.DiscountPercent, nullString(sub.DiscountReason),
	)
	return err
}

// ════════════════════════════════════════════════════════════
// USO E LIMITES
// ════════════════════════════════════════════════════════════

func (r *BillingRepo) IncrementUsage(ctx context.Context, tenantID, metric string, delta int) error {
	column := metricToColumn(metric)
	if column == "" {
		return fmt.Errorf("métrica inválida: %s", metric)
	}

	query := fmt.Sprintf(`
		UPDATE nexo.billing_subscriptions 
		SET %s = %s + $2, updated_at = NOW()
		WHERE tenant_id = $1`, column, column)
	_, err := r.db.pool.ExecContext(ctx, query, tenantID, delta)
	return err
}

func (r *BillingRepo) GetUsage(ctx context.Context, tenantID string) (map[string]int, error) {
	var users, tx, products, invoices int
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT current_users, current_transactions, current_products, current_invoices
		FROM nexo.billing_subscriptions WHERE tenant_id = $1`, tenantID).
		Scan(&users, &tx, &products, &invoices)
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"users":        users,
		"transactions": tx,
		"products":     products,
		"invoices":     invoices,
	}, nil
}

func (r *BillingRepo) CheckLimit(ctx context.Context, tenantID, metric string) (bool, error) {
	var canProceed bool
	err := r.db.pool.QueryRowContext(ctx,
		`SELECT nexo.check_plan_limit($1, $2)`, tenantID, metric).Scan(&canProceed)
	return canProceed, err
}

func (r *BillingRepo) GetFeatures(ctx context.Context, tenantID string) (*billing.PlanFeatures, error) {
	var featuresJSON []byte
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT bp.features
		FROM nexo.billing_subscriptions bs
		JOIN nexo.billing_plans bp ON bp.id = bs.plan_id
		WHERE bs.tenant_id = $1 AND bs.status IN ('trialing', 'active')`,
		tenantID).Scan(&featuresJSON)
	if err != nil {
		return nil, err
	}
	return billing.ParseFeatures(featuresJSON)
}

// ════════════════════════════════════════════════════════════
// CUPONS
// ════════════════════════════════════════════════════════════

func (r *BillingRepo) GetCoupon(ctx context.Context, code string) (*billing.Coupon, error) {
	var c billing.Coupon
	var validUntil sql.NullTime

	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, code, description, discount_type, discount_value, duration_months,
		       valid_until,
		       (is_active AND (max_uses IS NULL OR current_uses < max_uses)
		        AND (valid_until IS NULL OR valid_until > NOW())) as is_valid
		FROM nexo.billing_coupons
		WHERE code = $1`, code).Scan(
		&c.ID, &c.Code, &c.Description, &c.DiscountType, &c.DiscountValue,
		&c.DurationMonths, &validUntil, &c.IsValid,
	)
	if err == sql.ErrNoRows {
		return nil, billing.ErrCouponInvalid
	}
	if err != nil {
		return nil, err
	}

	if validUntil.Valid {
		c.ValidUntil = &validUntil.Time
	}
	return &c, nil
}

func (r *BillingRepo) UseCoupon(ctx context.Context, code, tenantID string) error {
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE nexo.billing_coupons 
		SET current_uses = current_uses + 1
		WHERE code = $1`, code)
	return err
}

// ════════════════════════════════════════════════════════════
// HELPERS
// ════════════════════════════════════════════════════════════

func metricToColumn(metric string) string {
	switch metric {
	case "users":
		return "current_users"
	case "transactions":
		return "current_transactions"
	case "products":
		return "current_products"
	case "invoices":
		return "current_invoices"
	case "storage_mb":
		return "current_storage_mb"
	default:
		return ""
	}
}
