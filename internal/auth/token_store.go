// Package auth — armazenamento de refresh tokens no Redis.
// Implementa auth.TokenStore.
//
// Estrutura de chaves Redis:
//   refresh:<token>          → JSON com userID e tenantID  (TTL = 7 dias)
//   user_tokens:<userID>     → SET de tokens ativos do usuário (para logout total)
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SimpleRedis é a interface mínima de Redis para o TokenStore.
type SimpleRedis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	SAdd(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...string) error
}

type tokenData struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// RedisTokenStore implementa auth.TokenStore usando Redis.
type RedisTokenStore struct {
	redis SimpleRedis
}

func NewRedisTokenStore(r SimpleRedis) *RedisTokenStore {
	return &RedisTokenStore{redis: r}
}

func refreshKey(token string) string { return "refresh:" + token }
func userTokensKey(userID string) string { return "user_tokens:" + userID }

// SaveRefreshToken armazena um refresh token com TTL.
func (s *RedisTokenStore) SaveRefreshToken(ctx context.Context, token, userID, tenantID string, ttl time.Duration) error {
	data, err := json.Marshal(tokenData{UserID: userID, TenantID: tenantID})
	if err != nil {
		return err
	}
	// Salvar o token
	if err := s.redis.Set(ctx, refreshKey(token), string(data), ttl); err != nil {
		return fmt.Errorf("RedisTokenStore.Save: %w", err)
	}
	// Adicionar ao set de tokens do usuário (para logout total)
	return s.redis.SAdd(ctx, userTokensKey(userID), token)
}

// GetRefreshToken valida e retorna os dados de um refresh token.
func (s *RedisTokenStore) GetRefreshToken(ctx context.Context, token string) (userID, tenantID string, err error) {
	raw, err := s.redis.Get(ctx, refreshKey(token))
	if err != nil || raw == "" {
		return "", "", fmt.Errorf("refresh token não encontrado ou expirado")
	}
	var td tokenData
	if err := json.Unmarshal([]byte(raw), &td); err != nil {
		return "", "", fmt.Errorf("refresh token corrompido")
	}
	return td.UserID, td.TenantID, nil
}

// DeleteRefreshToken invalida um refresh token específico.
func (s *RedisTokenStore) DeleteRefreshToken(ctx context.Context, token string) error {
	// Buscar userID antes de deletar (para remover do set)
	raw, _ := s.redis.Get(ctx, refreshKey(token))
	if raw != "" {
		var td tokenData
		if err := json.Unmarshal([]byte(raw), &td); err == nil {
			_ = s.redis.SRem(ctx, userTokensKey(td.UserID), token)
		}
	}
	return s.redis.Del(ctx, refreshKey(token))
}

// DeleteAllUserTokens invalida TODOS os refresh tokens de um usuário.
// Usado quando o usuário troca a senha ou em caso de comprometimento.
func (s *RedisTokenStore) DeleteAllUserTokens(ctx context.Context, userID string) error {
	tokens, err := s.redis.SMembers(ctx, userTokensKey(userID))
	if err != nil {
		return err
	}
	// Deletar cada token
	keys := make([]string, len(tokens))
	for i, t := range tokens {
		keys[i] = refreshKey(t)
	}
	if len(keys) > 0 {
		if err := s.redis.Del(ctx, keys...); err != nil {
			return err
		}
	}
	// Limpar o set do usuário
	return s.redis.Del(ctx, userTokensKey(userID))
}
