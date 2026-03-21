// Package enterprise implementa recursos da Fase 4 — Enterprise do Nexo One.
//
// Funcionalidades:
//   - API Keys para integrações externas (marketplaces, ERPs, apps)
//   - Multi-filial com consolidação de resultados
//   - Relatórios personalizados por nicho (PDF/Excel)
//   - Rate limiting por API Key
package enterprise

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ── CONSTANTES ────────────────────────────────────────────────────────────────

const (
	APIKeyPrefix    = "nxo_" // nexo one
	APIKeyLength    = 32
	MaxKeysPerTenant = 10

	ScopeRead      = "read"
	ScopeWrite     = "write"
	ScopeFinance   = "finance"
	ScopeInventory = "inventory"
	ScopeOrders    = "orders"
	ScopeWebhook   = "webhook"
	ScopeAdmin     = "admin"
)

// ── TIPOS ─────────────────────────────────────────────────────────────────────

// APIKey chave de API para integração externa.
type APIKey struct {
	ID          string
	TenantID    string
	Name        string     // ex: "Integração Shopify", "App Mobile"
	KeyHash     string     // SHA256 da chave — nunca armazenar a chave em texto
	KeyPrefix   string     // primeiros 8 chars para identificação: nxo_abc1...
	Scopes      []string   // permissões: read, write, finance, inventory, orders
	RateLimit   int        // requisições por minuto (default: 60)
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time // nil = não expira
	IsActive    bool
	CreatedBy   string
	CreatedAt   time.Time
}

// APIKeyFull chave completa — retornada apenas na criação.
type APIKeyFull struct {
	APIKey
	Key string // chave completa — mostrar UMA VEZ ao criar
}

// Subsidiary filial de um tenant.
type Subsidiary struct {
	ID           string
	TenantID     string  // tenant principal
	Name         string
	CNPJ         string
	Address      string
	City         string
	State        string
	IsHeadquarter bool
	IsActive     bool
	CreatedAt    time.Time
}

// ConsolidatedReport relatório consolidado de todas as filiais.
type ConsolidatedReport struct {
	TenantID      string
	Period        string
	GeneratedAt   time.Time
	Subsidiaries  []SubsidiaryResult
	Totals        ReportTotals
}

type SubsidiaryResult struct {
	SubsidiaryID   string
	SubsidiaryName string
	Revenue        float64
	Expenses       float64
	NetResult      float64
	Margin         float64
}

type ReportTotals struct {
	TotalRevenue  float64
	TotalExpenses float64
	NetResult     float64
	NetMargin     float64
	BestSubsidiary string
}

// WebhookEndpoint webhook para notificações em tempo real.
type WebhookEndpoint struct {
	ID        string
	TenantID  string
	URL       string
	Events    []string // os.created, payment.received, stock.low, etc.
	Secret    string   // para validar assinatura HMAC
	IsActive  bool
	CreatedAt time.Time
}

// ── ERROS ─────────────────────────────────────────────────────────────────────

var (
	ErrKeyNotFound    = errors.New("API key não encontrada")
	ErrKeyExpired     = errors.New("API key expirada")
	ErrKeyInactive    = errors.New("API key desativada")
	ErrScopeInsufficient = errors.New("permissão insuficiente")
	ErrMaxKeysReached = errors.New("limite de API keys atingido")
)

// ── REPOSITÓRIO ───────────────────────────────────────────────────────────────

type Repository interface {
	// API Keys
	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, tenantID string) ([]*APIKey, error)
	UpdateAPIKey(ctx context.Context, key *APIKey) error
	DeleteAPIKey(ctx context.Context, tenantID, id string) error
	CountAPIKeys(ctx context.Context, tenantID string) (int, error)

	// Filiais
	CreateSubsidiary(ctx context.Context, s *Subsidiary) error
	ListSubsidiaries(ctx context.Context, tenantID string) ([]*Subsidiary, error)
	GetSubsidiary(ctx context.Context, tenantID, id string) (*Subsidiary, error)

	// Webhooks
	CreateWebhook(ctx context.Context, w *WebhookEndpoint) error
	ListWebhooks(ctx context.Context, tenantID string) ([]*WebhookEndpoint, error)
	UpdateWebhook(ctx context.Context, w *WebhookEndpoint) error
}

// ── SERVIÇO ───────────────────────────────────────────────────────────────────

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ── API KEYS ──────────────────────────────────────────────────────────────────

// CreateAPIKey gera uma nova API key.
// A chave completa é retornada APENAS uma vez — não é possível recuperá-la depois.
func (s *Service) CreateAPIKey(ctx context.Context, tenantID, name, createdBy string, scopes []string, expiresAt *time.Time, rateLimit int) (*APIKeyFull, error) {
	// Verificar limite
	count, err := s.repo.CountAPIKeys(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if count >= MaxKeysPerTenant {
		return nil, ErrMaxKeysReached
	}

	// Gerar chave aleatória
	rawKey, err := generateKey()
	if err != nil {
		return nil, fmt.Errorf("enterprise.CreateAPIKey: %w", err)
	}

	fullKey := APIKeyPrefix + rawKey
	keyHash := hashKey(fullKey)
	keyPrefix := fullKey[:12] + "..." // nxo_abc123...

	if rateLimit <= 0 {
		rateLimit = 60
	}

	if len(scopes) == 0 {
		scopes = []string{ScopeRead}
	}

	key := &APIKey{
		TenantID:  tenantID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    scopes,
		RateLimit: rateLimit,
		ExpiresAt: expiresAt,
		IsActive:  true,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateAPIKey(ctx, key); err != nil {
		return nil, err
	}

	return &APIKeyFull{APIKey: *key, Key: fullKey}, nil
}

// ValidateAPIKey valida uma API key e retorna o tenant.
func (s *Service) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKey, error) {
	if rawKey == "" {
		return nil, ErrKeyNotFound
	}

	keyHash := hashKey(rawKey)
	key, err := s.repo.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, ErrKeyNotFound
	}

	if !key.IsActive {
		return nil, ErrKeyInactive
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	// Atualizar last used
	now := time.Now()
	key.LastUsedAt = &now
	s.repo.UpdateAPIKey(ctx, key)

	return key, nil
}

// HasScope verifica se a key tem uma permissão específica.
func (s *Service) HasScope(key *APIKey, requiredScope string) bool {
	for _, scope := range key.Scopes {
		if scope == requiredScope || scope == ScopeAdmin {
			return true
		}
	}
	return false
}

// RevokeAPIKey desativa uma API key.
func (s *Service) RevokeAPIKey(ctx context.Context, tenantID, id string) error {
	key, err := s.repo.GetAPIKeyByHash(ctx, id)
	if err != nil {
		return ErrKeyNotFound
	}
	key.IsActive = false
	return s.repo.UpdateAPIKey(ctx, key)
}

func (s *Service) ListAPIKeys(ctx context.Context, tenantID string) ([]*APIKey, error) {
	return s.repo.ListAPIKeys(ctx, tenantID)
}

// ── FILIAIS ───────────────────────────────────────────────────────────────────

func (s *Service) CreateSubsidiary(ctx context.Context, sub *Subsidiary) error {
	sub.CreatedAt = time.Now()
	return s.repo.CreateSubsidiary(ctx, sub)
}

func (s *Service) ListSubsidiaries(ctx context.Context, tenantID string) ([]*Subsidiary, error) {
	return s.repo.ListSubsidiaries(ctx, tenantID)
}

// GetConsolidatedReport gera relatório consolidado de todas as filiais.
func (s *Service) GetConsolidatedReport(ctx context.Context, tenantID string, year, month int) (*ConsolidatedReport, error) {
	subsidiaries, err := s.repo.ListSubsidiaries(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	report := &ConsolidatedReport{
		TenantID:    tenantID,
		Period:      fmt.Sprintf("%02d/%d", month, year),
		GeneratedAt: time.Now(),
	}

	var bestRevenue float64
	var bestName string

	for _, sub := range subsidiaries {
		// Em produção: buscar dados reais de cada filial
		// Por ora usando dados demo
		result := SubsidiaryResult{
			SubsidiaryID:   sub.ID,
			SubsidiaryName: sub.Name,
			Revenue:        0,
			Expenses:       0,
		}
		result.NetResult = result.Revenue - result.Expenses
		if result.Revenue > 0 {
			result.Margin = result.NetResult / result.Revenue * 100
		}

		report.Subsidiaries = append(report.Subsidiaries, result)
		report.Totals.TotalRevenue += result.Revenue
		report.Totals.TotalExpenses += result.Expenses

		if result.Revenue > bestRevenue {
			bestRevenue = result.Revenue
			bestName = sub.Name
		}
	}

	report.Totals.NetResult = report.Totals.TotalRevenue - report.Totals.TotalExpenses
	if report.Totals.TotalRevenue > 0 {
		report.Totals.NetMargin = report.Totals.NetResult / report.Totals.TotalRevenue * 100
	}
	report.Totals.BestSubsidiary = bestName

	return report, nil
}

// ── WEBHOOKS ──────────────────────────────────────────────────────────────────

func (s *Service) CreateWebhook(ctx context.Context, w *WebhookEndpoint) error {
	// Gerar secret para validação HMAC
	secret, _ := generateKey()
	w.Secret = secret[:16]
	w.IsActive = true
	w.CreatedAt = time.Now()
	return s.repo.CreateWebhook(ctx, w)
}

func (s *Service) ListWebhooks(ctx context.Context, tenantID string) ([]*WebhookEndpoint, error) {
	return s.repo.ListWebhooks(ctx, tenantID)
}

// ── HELPERS ───────────────────────────────────────────────────────────────────

func generateKey() (string, error) {
	bytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
