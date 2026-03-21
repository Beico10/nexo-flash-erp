// Package memory — repositório in-memory Enterprise.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/enterprise"
)

type EnterpriseRepo struct {
	mu          sync.RWMutex
	apiKeys     map[string]*enterprise.APIKey      // key: hash
	subsidiaries map[string][]*enterprise.Subsidiary // key: tenantID
	webhooks    map[string][]*enterprise.WebhookEndpoint
}

func NewEnterpriseRepo() *EnterpriseRepo {
	r := &EnterpriseRepo{
		apiKeys:      make(map[string]*enterprise.APIKey),
		subsidiaries: make(map[string][]*enterprise.Subsidiary),
		webhooks:     make(map[string][]*enterprise.WebhookEndpoint),
	}
	r.seed()
	return r
}

func (r *EnterpriseRepo) seed() {
	now := time.Now()
	// Filial demo
	r.subsidiaries["demo"] = []*enterprise.Subsidiary{
		{
			ID: uuid.NewString(), TenantID: "demo",
			Name: "Matriz — São Paulo", CNPJ: "12.345.678/0001-90",
			City: "São Paulo", State: "SP",
			IsHeadquarter: true, IsActive: true, CreatedAt: now,
		},
	}
}

// API Keys
func (r *EnterpriseRepo) CreateAPIKey(ctx context.Context, key *enterprise.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if key.ID == "" { key.ID = uuid.NewString() }
	copy := *key
	r.apiKeys[key.KeyHash] = &copy
	return nil
}

func (r *EnterpriseRepo) GetAPIKeyByHash(ctx context.Context, hash string) (*enterprise.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key, ok := r.apiKeys[hash]
	if !ok { return nil, enterprise.ErrKeyNotFound }
	copy := *key
	return &copy, nil
}

func (r *EnterpriseRepo) ListAPIKeys(ctx context.Context, tenantID string) ([]*enterprise.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*enterprise.APIKey
	for _, k := range r.apiKeys {
		if k.TenantID == tenantID {
			copy := *k
			result = append(result, &copy)
		}
	}
	return result, nil
}

func (r *EnterpriseRepo) UpdateAPIKey(ctx context.Context, key *enterprise.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *key
	r.apiKeys[key.KeyHash] = &copy
	return nil
}

func (r *EnterpriseRepo) DeleteAPIKey(ctx context.Context, tenantID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, k := range r.apiKeys {
		if k.TenantID == tenantID && k.ID == id {
			delete(r.apiKeys, hash)
			return nil
		}
	}
	return nil
}

func (r *EnterpriseRepo) CountAPIKeys(ctx context.Context, tenantID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, k := range r.apiKeys {
		if k.TenantID == tenantID && k.IsActive {
			count++
		}
	}
	return count, nil
}

// Subsidiaries
func (r *EnterpriseRepo) CreateSubsidiary(ctx context.Context, s *enterprise.Subsidiary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" { s.ID = uuid.NewString() }
	copy := *s
	r.subsidiaries[s.TenantID] = append(r.subsidiaries[s.TenantID], &copy)
	return nil
}

func (r *EnterpriseRepo) ListSubsidiaries(ctx context.Context, tenantID string) ([]*enterprise.Subsidiary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	subs := r.subsidiaries[tenantID]
	result := make([]*enterprise.Subsidiary, len(subs))
	for i, s := range subs {
		copy := *s
		result[i] = &copy
	}
	return result, nil
}

func (r *EnterpriseRepo) GetSubsidiary(ctx context.Context, tenantID, id string) (*enterprise.Subsidiary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.subsidiaries[tenantID] {
		if s.ID == id {
			copy := *s
			return &copy, nil
		}
	}
	return nil, enterprise.ErrKeyNotFound
}

// Webhooks
func (r *EnterpriseRepo) CreateWebhook(ctx context.Context, w *enterprise.WebhookEndpoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if w.ID == "" { w.ID = uuid.NewString() }
	copy := *w
	r.webhooks[w.TenantID] = append(r.webhooks[w.TenantID], &copy)
	return nil
}

func (r *EnterpriseRepo) ListWebhooks(ctx context.Context, tenantID string) ([]*enterprise.WebhookEndpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	hooks := r.webhooks[tenantID]
	result := make([]*enterprise.WebhookEndpoint, len(hooks))
	for i, h := range hooks {
		copy := *h
		result[i] = &copy
	}
	return result, nil
}

func (r *EnterpriseRepo) UpdateWebhook(ctx context.Context, w *enterprise.WebhookEndpoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existing := range r.webhooks[w.TenantID] {
		if existing.ID == w.ID {
			copy := *w
			r.webhooks[w.TenantID][i] = &copy
			return nil
		}
	}
	return nil
}
