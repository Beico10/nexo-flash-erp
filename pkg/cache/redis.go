// Package cache implementa o cache Redis do Nexo Flash.
//
// Uso principal: cache de alíquotas NCM (TTL 1h).
// As alíquotas mudam raramente (publicação no DOU), então 1h é seguro.
// Ao atualizar a tabela fiscal_ncm_rates, invalide o cache com InvalidateNCM().
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// TTLAliquota é o tempo de vida do cache de alíquotas NCM.
	// 1 hora: seguro pois alíquotas só mudam com publicação oficial.
	TTLAliquota = 1 * time.Hour

	// TTLSession é o TTL de sessões e tokens de aprovação WhatsApp.
	TTLSession = 24 * time.Hour

	// TTLRateLimit é o janela de rate limiting por IP/tenant.
	TTLRateLimit = 1 * time.Minute
)

// RedisClient é a interface mínima de acesso ao Redis.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// NCMCache gerencia o cache de alíquotas por NCM.
type NCMCache struct {
	redis RedisClient
}

func NewNCMCache(r RedisClient) *NCMCache { return &NCMCache{redis: r} }

// keyNCM monta a chave Redis para um NCM.
// Inclui data para invalidação por mudança de alíquota na data.
func keyNCM(ncm string, date time.Time) string {
	return fmt.Sprintf("ncm:rate:%s:%s", ncm, date.Format("2006-01-02"))
}

// Get busca alíquota em cache. Retorna nil se não encontrado (cache miss).
func (c *NCMCache) Get(ctx context.Context, ncm string, date time.Time) (map[string]any, error) {
	raw, err := c.redis.Get(ctx, keyNCM(ncm, date))
	if err != nil {
		return nil, nil // cache miss — sem erro
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, nil
	}
	return result, nil
}

// Set armazena alíquota em cache com TTL de 1h.
func (c *NCMCache) Set(ctx context.Context, ncm string, date time.Time, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, keyNCM(ncm, date), string(raw), TTLAliquota)
}

// InvalidateNCM remove o cache de um NCM específico (após atualização de alíquota).
func (c *NCMCache) InvalidateNCM(ctx context.Context, ncm string) error {
	// Invalida para os próximos 3 dias por segurança
	for i := 0; i < 3; i++ {
		date := time.Now().AddDate(0, 0, i)
		if err := c.redis.Del(ctx, keyNCM(ncm, date)); err != nil {
			return err
		}
	}
	return nil
}

// SessionCache gerencia tokens de sessão e aprovação WhatsApp.
type SessionCache struct {
	redis RedisClient
}

func NewSessionCache(r RedisClient) *SessionCache { return &SessionCache{redis: r} }

func keySession(token string) string     { return "session:" + token }
func keyApproval(token string) string    { return "approval:whatsapp:" + token }

// SetApprovalToken armazena o token de aprovação WhatsApp com TTL de 24h.
func (s *SessionCache) SetApprovalToken(ctx context.Context, token, osID string) error {
	return s.redis.Set(ctx, keyApproval(token), osID, TTLSession)
}

// GetApprovalToken recupera o ID da OS associada ao token de aprovação.
func (s *SessionCache) GetApprovalToken(ctx context.Context, token string) (string, error) {
	return s.redis.Get(ctx, keyApproval(token))
}

// DeleteApprovalToken invalida o token após uso.
func (s *SessionCache) DeleteApprovalToken(ctx context.Context, token string) error {
	return s.redis.Del(ctx, keyApproval(token))
}

// RateLimiter implementa rate limiting por chave (IP ou tenant_id).
type RateLimiter struct {
	redis   RedisClient
	maxReqs int64
	window  time.Duration
}

func NewRateLimiter(r RedisClient, maxReqs int64, window time.Duration) *RateLimiter {
	return &RateLimiter{redis: r, maxReqs: maxReqs, window: window}
}

// Allow verifica se a chave está dentro do limite de requisições.
// Retorna true se permitido, false se limite atingido.
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rkey := "ratelimit:" + key
	count, err := rl.redis.Incr(ctx, rkey)
	if err != nil {
		return true, nil // fail open: se Redis cair, não bloqueia
	}
	if count == 1 {
		// Primeira requisição na janela: configura TTL
		_ = rl.redis.Expire(ctx, rkey, rl.window)
	}
	return count <= rl.maxReqs, nil
}
