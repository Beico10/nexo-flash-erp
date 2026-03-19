// Package cache — implementação real do cliente Redis.
// Conecta ao Redis via URL e implementa todas as interfaces do sistema.
package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client é o cliente Redis do Nexo One.
// Implementa: cache.RedisClient, auth.SimpleRedis, repository.RedisClient
type Client struct {
	rdb *redis.Client
}

// NewRedisClient cria um novo cliente Redis a partir de uma URL.
// URL formato: redis://:senha@host:6379/0
func NewRedisClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache.NewRedisClient: URL inválida: %w", err)
	}

	rdb := redis.NewClient(opts)

	// Verificar conectividade
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache.NewRedisClient: ping falhou: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Get busca um valor pelo key. Retorna erro se não encontrado.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key não encontrada: %s", key)
	}
	return val, err
}

// Set armazena um valor com TTL.
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Del remove uma ou mais keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Incr incrementa um contador e retorna o novo valor.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire define o TTL de uma key existente.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// SAdd adiciona membros a um SET Redis.
func (c *Client) SAdd(ctx context.Context, key string, members ...string) error {
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.rdb.SAdd(ctx, key, args...).Err()
}

// SMembers retorna todos os membros de um SET Redis.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SRem remove membros de um SET Redis.
func (c *Client) SRem(ctx context.Context, key string, members ...string) error {
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.rdb.SRem(ctx, key, args...).Err()
}

// Close fecha a conexão com o Redis.
func (c *Client) Close() error { return c.rdb.Close() }

// Ping verifica a conectividade.
func (c *Client) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

// FlushPrefix remove todas as keys com um prefixo (uso administrativo).
func (c *Client) FlushPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// SetTenantSession configura o tenant_id na sessão PostgreSQL.
// Implementa pkg/middleware.DBSessionSetter.
func (c *Client) SetTenantSession(_ context.Context, _ string) error {
	// Redis não tem sessão SQL — este método é implementado pelo postgres.DB
	return nil
}

// helper interno
func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}
