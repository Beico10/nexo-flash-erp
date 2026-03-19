// Package postgres — repositório de alíquotas fiscais com cache Redis.
// Implementa tax.RateRepository.
//
// Fluxo de cache:
//  1. Busca no Redis (TTL 1h)
//  2. Se não encontrado: busca no PostgreSQL
//  3. Persiste no Redis com TTL 1h
//  4. Retorna o resultado
//
// Ao atualizar alíquotas no banco: chame InvalidateNCM() para limpar o cache.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexoone/nexo-one/internal/tax"
)

// RedisClient interface mínima para o cache de alíquotas.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// TaxRateRepo implementa tax.RateRepository com cache Redis.
type TaxRateRepo struct {
	db    *DB
	redis RedisClient
}

func NewTaxRateRepo(db *DB, redis RedisClient) *TaxRateRepo {
	return &TaxRateRepo{db: db, redis: redis}
}

const taxRateTTL = 1 * time.Hour

func cacheKey(ncm string, date time.Time) string {
	return fmt.Sprintf("ncm:rate:%s:%s", ncm, date.Format("2006-01-02"))
}

// GetRate implementa tax.RateRepository.
// Busca alíquotas com prioridade: Redis → PostgreSQL.
func (r *TaxRateRepo) GetRate(ctx context.Context, ncm string, referenceDate time.Time) (*tax.NCMRate, error) {
	key := cacheKey(ncm, referenceDate)

	// 1. Tentar cache Redis
	if raw, err := r.redis.Get(ctx, key); err == nil && raw != "" {
		var rate tax.NCMRate
		if err := json.Unmarshal([]byte(raw), &rate); err == nil {
			return &rate, nil
		}
	}

	// 2. Buscar no PostgreSQL
	rate, err := r.fetchFromDB(ctx, ncm, referenceDate)
	if err != nil {
		return nil, err
	}

	// 3. Cachear no Redis
	if data, err := json.Marshal(rate); err == nil {
		_ = r.redis.Set(ctx, key, string(data), taxRateTTL)
	}

	return rate, nil
}

// fetchFromDB busca alíquotas diretamente no banco.
// Respeita o período de transição: busca a alíquota vigente para a data de referência.
func (r *TaxRateRepo) fetchFromDB(ctx context.Context, ncm string, ref time.Time) (*tax.NCMRate, error) {
	query := `
		SELECT
			ncm_code, ibs_rate, cbs_rate, selective_rate,
			basket_reduced, COALESCE(basket_type,''),
			COALESCE(transition_year, 0),
			COALESCE(transition_factor, 1.0)
		FROM fiscal_ncm_rates
		WHERE ncm_code = $1
		  AND valid_from <= $2
		  AND (valid_until IS NULL OR valid_until >= $2)
		ORDER BY valid_from DESC
		LIMIT 1`

	var rate tax.NCMRate
	err := r.db.pool.QueryRowContext(ctx, query, ncm, ref.Format("2006-01-02")).Scan(
		&rate.NCMCode,
		&rate.IBSRate,
		&rate.CBSRate,
		&rate.SelectiveRate,
		&rate.BasketReduced,
		&rate.BasketType,
		&rate.TransitionYear,
		&rate.TransitionFactor,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("NCM %s não encontrado para %s", ncm, ref.Format("2006-01-02"))
	}
	return &rate, err
}

// InvalidateNCM remove do cache as alíquotas de um NCM específico.
// Deve ser chamado após atualizar fiscal_ncm_rates.
func (r *TaxRateRepo) InvalidateNCM(ctx context.Context, ncm string) error {
	// Invalida para os próximos 3 dias (janela de segurança)
	for i := 0; i < 3; i++ {
		date := time.Now().AddDate(0, 0, i)
		if err := r.redis.Del(ctx, cacheKey(ncm, date)); err != nil {
			return err
		}
	}
	return nil
}

// BulkInsertRates insere ou atualiza múltiplas alíquotas (para importação da SEFAZ).
func (r *TaxRateRepo) BulkInsertRates(ctx context.Context, rates []tax.NCMRate, sourceLei string) error {
	tx, err := r.db.pool.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO fiscal_ncm_rates (
			ncm_code, ncm_description, ibs_rate, cbs_rate, selective_rate,
			basket_reduced, basket_type, transition_year, transition_factor,
			valid_from, source_lei, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_DATE, $10, NOW())
		ON CONFLICT (ncm_code, valid_from) DO UPDATE SET
			ibs_rate         = EXCLUDED.ibs_rate,
			cbs_rate         = EXCLUDED.cbs_rate,
			selective_rate   = EXCLUDED.selective_rate,
			basket_reduced   = EXCLUDED.basket_reduced,
			basket_type      = EXCLUDED.basket_type,
			transition_factor= EXCLUDED.transition_factor,
			source_lei       = EXCLUDED.source_lei,
			updated_at       = NOW()`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, rate := range rates {
		_, err := stmt.ExecContext(ctx,
			rate.NCMCode,
			"", // description — preencher via SEFAZ API
			rate.IBSRate,
			rate.CBSRate,
			rate.SelectiveRate,
			rate.BasketReduced,
			nullString(rate.BasketType),
			sql.NullInt64{Int64: int64(rate.TransitionYear), Valid: rate.TransitionYear > 0},
			rate.TransitionFactor,
			sourceLei,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("InsertRate NCM %s: %w", rate.NCMCode, err)
		}
	}

	return tx.Commit()
}
