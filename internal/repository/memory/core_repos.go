package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/tax"
	"golang.org/x/crypto/bcrypt"
)

// Tenant representa um tenant do sistema.
type Tenant struct {
	ID             string
	Slug           string
	BusinessType   string
	Name           string
	CNPJ           string
	Plan           string
	ModulesEnabled []string
	Timezone       string
	Currency       string
	IsActive       bool
	CreatedAt      time.Time
}

// User representa um usuario do sistema.
type User struct {
	ID           string
	TenantID     string
	Email        string
	Name         string
	Role         string
	PasswordHash string
	Active       bool
	CreatedAt    time.Time
}

// TenantRepo gerencia tenants in-memory.
type TenantRepo struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	bySlug  map[string]*Tenant
}

func NewTenantRepo() *TenantRepo {
	repo := &TenantRepo{
		tenants: make(map[string]*Tenant),
		bySlug:  make(map[string]*Tenant),
	}
	repo.seedDemoData()
	return repo
}

func (r *TenantRepo) GetByID(_ context.Context, id string) (*Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant %s nao encontrado", id)
	}
	return t, nil
}

func (r *TenantRepo) GetBySlug(_ context.Context, slug string) (*Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.bySlug[slug]
	if !ok {
		return nil, fmt.Errorf("tenant '%s' nao encontrado", slug)
	}
	return t, nil
}

// UserRepo gerencia usuarios in-memory.
type UserRepo struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewUserRepo() *UserRepo {
	repo := &UserRepo{users: make(map[string]*User)}
	repo.seedDemoData()
	return repo
}

func (r *UserRepo) GetByEmail(_ context.Context, tenantID, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.TenantID == tenantID && u.Email == email && u.Active {
			return u, nil
		}
	}
	return nil, fmt.Errorf("usuario nao encontrado")
}

// TaxRateRepo implementa tax.RateRepository in-memory com dados da Reforma 2026.
type TaxRateRepo struct {
	mu    sync.RWMutex
	rates map[string]*tax.NCMRate
}

func NewTaxRateRepo() *TaxRateRepo {
	repo := &TaxRateRepo{rates: make(map[string]*tax.NCMRate)}
	repo.seedNCMRates()
	return repo
}

func (r *TaxRateRepo) GetRate(_ context.Context, ncm string, _ time.Time) (*tax.NCMRate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rate, ok := r.rates[ncm]
	if !ok {
		return nil, fmt.Errorf("NCM %s nao encontrado", ncm)
	}
	return rate, nil
}

// ListRates retorna todas as aliquotas NCM cadastradas.
func (r *TaxRateRepo) ListRates() []*tax.NCMRate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*tax.NCMRate, 0, len(r.rates))
	for _, rate := range r.rates {
		result = append(result, rate)
	}
	return result
}

// seedNCMRates carrega aliquotas base da Reforma 2026.
func (r *TaxRateRepo) seedNCMRates() {
	rates := []tax.NCMRate{
		// Cesta Basica Nacional - Aliquota Zero (Art. 8 LC 214/2025)
		{NCMCode: "19052000", NCMDescription: "Pao frances", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
		{NCMCode: "04011010", NCMDescription: "Leite fluido", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
		{NCMCode: "10063021", NCMDescription: "Arroz beneficiado", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
		{NCMCode: "07132319", NCMDescription: "Feijao preto", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
		{NCMCode: "02023000", NCMDescription: "Carne bovina desossada", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "zero", TransitionFactor: 1.0},
		// Cesta Estendida - Reducao 60%
		{NCMCode: "17019900", NCMDescription: "Acucar cristal", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "reduced_60", TransitionFactor: 1.0},
		{NCMCode: "15079011", NCMDescription: "Oleo de soja", IBSRate: 0.0925, CBSRate: 0.0375, BasketReduced: true, BasketType: "reduced_60", TransitionFactor: 1.0},
		// Produtos Industriais Padrao
		{NCMCode: "84715010", NCMDescription: "Computador portatil", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		{NCMCode: "87032100", NCMDescription: "Automovel ate 1000cc", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		// Imposto Seletivo
		{NCMCode: "22030000", NCMDescription: "Cerveja de malte", IBSRate: 0.0925, CBSRate: 0.0375, SelectiveRate: 0.10, TransitionFactor: 1.0},
		{NCMCode: "24012030", NCMDescription: "Tabaco", IBSRate: 0.0925, CBSRate: 0.0375, SelectiveRate: 0.25, TransitionFactor: 1.0},
		// Transicao 2026 (fator 10%)
		{NCMCode: "61051000", NCMDescription: "Camisa masculina algodao", IBSRate: 0.0925, CBSRate: 0.0375, TransitionYear: 2026, TransitionFactor: 0.10},
		// Pecas automotivas (mecanica)
		{NCMCode: "87083090", NCMDescription: "Pastilha de freio", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		{NCMCode: "84099190", NCMDescription: "Filtro de oleo", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		{NCMCode: "40111000", NCMDescription: "Pneu automovel", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		// Servicos
		{NCMCode: "00000000", NCMDescription: "Servico generico", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
		// Calcados
		{NCMCode: "64039990", NCMDescription: "Calcado couro", IBSRate: 0.0925, CBSRate: 0.0375, TransitionFactor: 1.0},
	}
	for i := range rates {
		r.rates[rates[i].NCMCode] = &rates[i]
	}
}

func (r *TenantRepo) seedDemoData() {
	demoTenantID := "00000000-0000-0000-0000-000000000001"
	t := &Tenant{
		ID:             demoTenantID,
		Slug:           "demo",
		BusinessType:   "mechanic",
		Name:           "Mecanica Demo Nexo One",
		Plan:           "trial",
		ModulesEnabled: []string{"mechanic", "bakery", "aesthetics", "logistics", "industry", "shoes"},
		Timezone:       "America/Sao_Paulo",
		Currency:       "BRL",
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
	}
	r.tenants[t.ID] = t
	r.bySlug[t.Slug] = t
}

func (r *UserRepo) seedDemoData() {
	demoTenantID := "00000000-0000-0000-0000-000000000001"
	hash, _ := bcrypt.GenerateFromPassword([]byte("demo123"), 10)
	u := &User{
		ID:           uuid.New().String(),
		TenantID:     demoTenantID,
		Email:        "admin@demo.com",
		Name:         "Admin Demo",
		Role:         "owner",
		PasswordHash: string(hash),
		Active:       true,
		CreatedAt:    time.Now().UTC(),
	}
	r.users[u.ID] = u
}
