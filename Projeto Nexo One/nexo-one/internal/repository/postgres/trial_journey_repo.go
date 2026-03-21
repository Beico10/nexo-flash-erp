// Package postgres — repositórios de trial e journey.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/journey"
	"github.com/nexoone/nexo-one/internal/trial"
)

// ── TRIAL REPO ────────────────────────────────────────────────────────────────

type TrialRepo struct{ db *DB }

func NewTrialRepo(db *DB) *TrialRepo { return &TrialRepo{db: db} }

func (r *TrialRepo) GetByPhoneHash(ctx context.Context, hash string) (*trial.TrialControl, error) {
	var tc trial.TrialControl
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, phone_number, phone_hash, COALESCE(email,''), COALESCE(cnpj,''),
		COALESCE(verification_code,''), code_expires_at, verified_at,
		COALESCE(device_hash,''), COALESCE(ip_address,''),
		COALESCE(tenant_id::text,''), is_blocked, abuse_score, created_at
		FROM trial_controls WHERE phone_hash=$1`, hash).Scan(
		&tc.ID, &tc.PhoneNumber, &tc.PhoneHash, &tc.Email, &tc.CNPJ,
		&tc.VerificationCode, &tc.CodeExpiresAt, &tc.VerifiedAt,
		&tc.DeviceHash, &tc.IPAddress, &tc.TenantID,
		&tc.IsBlocked, &tc.AbuseScore, &tc.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trial não encontrado")
	}
	return &tc, err
}

func (r *TrialRepo) GetByDeviceHash(ctx context.Context, hash string, since time.Time) ([]*trial.TrialControl, error) {
	rows, err := r.db.pool.QueryContext(ctx, `
		SELECT id, phone_hash, email, device_hash, created_at
		FROM trial_controls WHERE device_hash=$1 AND created_at >= $2`, hash, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*trial.TrialControl
	for rows.Next() {
		var tc trial.TrialControl
		rows.Scan(&tc.ID, &tc.PhoneHash, &tc.Email, &tc.DeviceHash, &tc.CreatedAt)
		list = append(list, &tc)
	}
	return list, rows.Err()
}

func (r *TrialRepo) Create(ctx context.Context, tc *trial.TrialControl) error {
	return r.db.pool.QueryRowContext(ctx, `
		INSERT INTO trial_controls
		(phone_number, phone_hash, email, cnpj, device_hash, ip_address, abuse_score)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
		tc.PhoneNumber, tc.PhoneHash, tc.Email, tc.CNPJ,
		tc.DeviceHash, tc.IPAddress, tc.AbuseScore,
	).Scan(&tc.ID, &tc.CreatedAt)
}

func (r *TrialRepo) Update(ctx context.Context, tc *trial.TrialControl) error {
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE trial_controls SET
			verified_at=$2, is_blocked=$3, abuse_score=$4, tenant_id=$5
		WHERE id=$1`,
		tc.ID, tc.VerifiedAt, tc.IsBlocked, tc.AbuseScore,
		nullString(tc.TenantID))
	return err
}

// SaveCode salva código no banco com TTL (usando expires_at).
func (r *TrialRepo) SaveCode(ctx context.Context, phoneHash, code string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE trial_controls SET verification_code=$2, code_expires_at=$3
		WHERE phone_hash=$1`, phoneHash, code, expiresAt)
	return err
}

func (r *TrialRepo) GetCode(ctx context.Context, phoneHash string) (string, error) {
	var code string
	var expiresAt time.Time
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT COALESCE(verification_code,''), COALESCE(code_expires_at, NOW())
		FROM trial_controls WHERE phone_hash=$1`, phoneHash).
		Scan(&code, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return "", fmt.Errorf("código expirado")
	}
	return code, nil
}

func (r *TrialRepo) IncrementAttempts(ctx context.Context, phoneHash string) (int, error) {
	var attempts int
	err := r.db.pool.QueryRowContext(ctx, `
		UPDATE trial_controls SET abuse_score = abuse_score + 10
		WHERE phone_hash=$1 RETURNING abuse_score`, phoneHash).Scan(&attempts)
	return attempts / 10, err
}

// ── JOURNEY REPO ──────────────────────────────────────────────────────────────

type JourneyRepo struct{ db *DB }

func NewJourneyRepo(db *DB) *JourneyRepo { return &JourneyRepo{db: db} }

func (r *JourneyRepo) TrackEvent(ctx context.Context, e *journey.Event) error {
	propsJSON, _ := json.Marshal(e.Properties)
	_, err := r.db.pool.ExecContext(ctx, `
		INSERT INTO journey_events
		(tenant_id, user_id, anonymous_id, event_name, event_category,
		page_path, page_title, referrer, properties, funnel_stage,
		device_type, browser, os_name, session_id, occurred_at, time_on_page)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		nullString(e.TenantID), nullString(e.UserID), nullString(e.AnonymousID),
		e.EventName, e.EventCategory,
		nullString(e.PagePath), nullString(e.PageTitle), nullString(e.Referrer),
		propsJSON, nullString(e.FunnelStage),
		nullString(e.DeviceType), nullString(e.Browser), nullString(e.OS),
		nullString(e.SessionID), e.OccurredAt, e.TimeOnPage,
	)
	return err
}

func (r *JourneyRepo) GetEvents(ctx context.Context, tenantID string, since time.Time) ([]*journey.Event, error) {
	rows, err := r.db.pool.QueryContext(ctx, `
		SELECT id, event_name, event_category, page_path, funnel_stage, occurred_at
		FROM journey_events WHERE tenant_id=$1 AND occurred_at >= $2
		ORDER BY occurred_at DESC LIMIT 100`, tenantID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*journey.Event
	for rows.Next() {
		var e journey.Event
		rows.Scan(&e.ID, &e.EventName, &e.EventCategory, &e.PagePath, &e.FunnelStage, &e.OccurredAt)
		list = append(list, &e)
	}
	return list, rows.Err()
}

func (r *JourneyRepo) GetFunnelMetrics(ctx context.Context, date time.Time, businessType string) (*journey.FunnelMetrics, error) {
	var m journey.FunnelMetrics
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT date, COALESCE(business_type,''), visits, signups_started, signups_completed,
		phone_verified, onboarding_started, onboarding_completed, first_action, trial_converted,
		conversion_rate
		FROM journey_funnel_daily
		WHERE date=$1 AND ($2='' OR business_type=$2)
		LIMIT 1`, date.Format("2006-01-02"), businessType).Scan(
		&m.Date, &m.BusinessType, &m.Visits, &m.SignupsStarted, &m.SignupsCompleted,
		&m.PhoneVerified, &m.OnboardingStarted, &m.OnboardingCompleted,
		&m.FirstAction, &m.TrialConverted, &m.ConversionRate,
	)
	if err == sql.ErrNoRows {
		m.Date = date
		m.BusinessType = businessType
		return &m, nil
	}
	return &m, err
}

func (r *JourneyRepo) GetFunnelRange(ctx context.Context, from, to time.Time, businessType string) ([]*journey.FunnelMetrics, error) {
	rows, err := r.db.pool.QueryContext(ctx, `
		SELECT date, visits, signups_completed, trial_converted, conversion_rate
		FROM journey_funnel_daily
		WHERE date BETWEEN $1 AND $2 AND ($3='' OR business_type=$3)
		ORDER BY date`, from, to, businessType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*journey.FunnelMetrics
	for rows.Next() {
		var m journey.FunnelMetrics
		rows.Scan(&m.Date, &m.Visits, &m.SignupsCompleted, &m.TrialConverted, &m.ConversionRate)
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *JourneyRepo) UpdateFunnelDaily(ctx context.Context, m *journey.FunnelMetrics) error {
	_, err := r.db.pool.ExecContext(ctx, `
		INSERT INTO journey_funnel_daily
		(date, business_type, visits, signups_started, signups_completed,
		phone_verified, onboarding_started, onboarding_completed, first_action, trial_converted)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (date, business_type) DO UPDATE SET
			visits=EXCLUDED.visits, signups_completed=EXCLUDED.signups_completed`,
		m.Date, m.BusinessType, m.Visits, m.SignupsStarted, m.SignupsCompleted,
		m.PhoneVerified, m.OnboardingStarted, m.OnboardingCompleted, m.FirstAction, m.TrialConverted,
	)
	return err
}

func (r *JourneyRepo) GetDropPoints(ctx context.Context, stage string, minDaysStuck int) ([]*journey.DropPoint, error) {
	rows, err := r.db.pool.QueryContext(ctx, `
		SELECT tenant_id, COALESCE(user_id,''), stage, step_code, days_stuck
		FROM journey_drop_points
		WHERE ($1='' OR stage=$1) AND days_stuck >= $2
		ORDER BY days_stuck DESC`, stage, minDaysStuck)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*journey.DropPoint
	for rows.Next() {
		var dp journey.DropPoint
		rows.Scan(&dp.TenantID, &dp.UserID, &dp.Stage, &dp.StepCode, &dp.DaysStuck)
		list = append(list, &dp)
	}
	return list, rows.Err()
}

func (r *JourneyRepo) MarkDropResolved(ctx context.Context, tenantID, resolution string) error {
	_, err := r.db.pool.ExecContext(ctx, `
		UPDATE journey_drop_points SET resolved_at=NOW(), resolution=$2
		WHERE tenant_id=$1 AND resolved_at IS NULL`, tenantID, resolution)
	return err
}

func (r *JourneyRepo) CreateDropPoint(ctx context.Context, dp *journey.DropPoint) error {
	_, err := r.db.pool.ExecContext(ctx, `
		INSERT INTO journey_drop_points (tenant_id, user_id, stage, step_code, days_stuck)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT (tenant_id, stage) DO NOTHING`,
		dp.TenantID, nullString(dp.UserID), dp.Stage, dp.StepCode, dp.DaysStuck)
	return err
}

func (r *JourneyRepo) GetOnboardingSteps(ctx context.Context, businessType string) ([]*journey.OnboardingStep, error) {
	rows, err := r.db.pool.QueryContext(ctx, `
		SELECT id, business_type, step_code, step_order, title, description,
		COALESCE(icon,''), is_required, is_skippable, COALESCE(estimated_time,0),
		COALESCE(action_type,''), COALESCE(reward_text,''), COALESCE(reward_days,0)
		FROM onboarding_steps WHERE business_type=$1 ORDER BY step_order`, businessType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []*journey.OnboardingStep
	for rows.Next() {
		var s journey.OnboardingStep
		rows.Scan(&s.ID, &s.BusinessType, &s.StepCode, &s.StepOrder,
			&s.Title, &s.Description, &s.Icon, &s.IsRequired, &s.IsSkippable,
			&s.EstimatedTime, &s.ActionType, &s.RewardText, &s.RewardDays)
		steps = append(steps, &s)
	}
	return steps, rows.Err()
}

func (r *JourneyRepo) GetOnboardingProgress(ctx context.Context, tenantID string) (*journey.OnboardingProgress, error) {
	var p journey.OnboardingProgress
	var completedJSON []byte
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT tenant_id, COALESCE(user_id,''), COALESCE(business_type,''),
		COALESCE(current_step,''), total_steps, completed_steps,
		started_at, completed_at, last_activity, skipped
		FROM onboarding_progress WHERE tenant_id=$1`, tenantID).Scan(
		&p.TenantID, &p.UserID, &p.BusinessType, &p.CurrentStep,
		&p.TotalSteps, &completedJSON, &p.StartedAt, &p.CompletedAt,
		&p.LastActivity, &p.Skipped,
	)
	if err == sql.ErrNoRows {
		return &journey.OnboardingProgress{TenantID: tenantID}, nil
	}
	json.Unmarshal(completedJSON, &p.CompletedSteps)
	return &p, err
}

func (r *JourneyRepo) UpdateOnboardingProgress(ctx context.Context, p *journey.OnboardingProgress) error {
	completedJSON, _ := json.Marshal(p.CompletedSteps)
	_, err := r.db.pool.ExecContext(ctx, `
		INSERT INTO onboarding_progress
		(tenant_id, current_step, total_steps, completed_steps, last_activity, completed_at, skipped)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			current_step=EXCLUDED.current_step,
			completed_steps=EXCLUDED.completed_steps,
			last_activity=EXCLUDED.last_activity,
			completed_at=EXCLUDED.completed_at,
			skipped=EXCLUDED.skipped`,
		p.TenantID, p.CurrentStep, p.TotalSteps, completedJSON,
		p.LastActivity, p.CompletedAt, p.Skipped,
	)
	return err
}
