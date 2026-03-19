// Package memory - cache e token store in-memory.
// Substitui Redis no ambiente de preview.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Cache implementa as interfaces de cache e token store.
type Cache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

func NewCache() *Cache {
	c := &Cache{data: make(map[string]cacheEntry)}
	go c.cleanup()
	return c
}

func (c *Cache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return "", fmt.Errorf("key nao encontrada: %s", key)
	}
	return e.value, nil
}

func (c *Cache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.data[key] = cacheEntry{value: value, expiresAt: exp}
	return nil
}

func (c *Cache) Del(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.data, k)
	}
	return nil
}

func (c *Cache) Incr(_ context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	if !ok {
		c.data[key] = cacheEntry{value: "1"}
		return 1, nil
	}
	var n int64
	fmt.Sscanf(e.value, "%d", &n)
	n++
	c.data[key] = cacheEntry{value: fmt.Sprintf("%d", n), expiresAt: e.expiresAt}
	return n, nil
}

func (c *Cache) Expire(_ context.Context, key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	if ok {
		e.expiresAt = time.Now().Add(ttl)
		c.data[key] = e
	}
	return nil
}

func (c *Cache) SAdd(_ context.Context, key string, members ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	existing := e.value
	if !ok {
		existing = ""
	}
	for _, m := range members {
		existing += "|" + m
	}
	c.data[key] = cacheEntry{value: existing}
	return nil
}

func (c *Cache) SMembers(_ context.Context, key string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[key]
	if !ok {
		return nil, nil
	}
	var members []string
	current := ""
	for _, ch := range e.value {
		if ch == '|' {
			if current != "" {
				members = append(members, current)
			}
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		members = append(members, current)
	}
	return members, nil
}

func (c *Cache) SRem(_ context.Context, key string, members ...string) error {
	return nil
}

func (c *Cache) Close() error { return nil }

func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.data {
			if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}
