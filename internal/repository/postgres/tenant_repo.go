// Package postgres — repositório de tenants e usuários.
// Operações de tenant NÃO usam RLS (são operações de superusuário/admin).
// Operações de usuário USAM RLS via WithTenant.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Tenant representa um tenant do sistema.
type Tenant struct {
	ID             string
	Slug           string
	BusinessType   string
	Name           string
	CNPJ           string
	RazaoSocial    string
	Plan           string
	ModulesEnabled []string
	Timezone       string
	Currency       string
	IsActive       bool
	CreatedAt      time.Time
}

// User representa um usuário do sistema.
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

// TenantRepo gerencia tenants (sem RLS — operações administrativas).
type TenantRepo struct {
	db *DB
}

func NewTenantRepo(db *DB) *TenantRepo { return &TenantRepo{db: db} }

// Create cria um novo tenant.
func (r *TenantRepo) Create(ctx context.Context, t *Tenant) error {
	modules, _ := json.Marshal(t.ModulesEnabled)
	return r.db.pool.QueryRowContext(ctx, `
		INSERT INTO nexo.tenants (slug, business_type, name, cnpj, razao_social, plan, active_modules, timezone, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at`,
		t.Slug, t.BusinessType, t.Name,
		nullString(t.CNPJ), nullString(t.RazaoSocial), t.Plan,
		modules, t.Timezone, t.Currency,
	).Scan(&t.ID, &t.CreatedAt)
}

// GetByID busca um tenant pelo ID (sem RLS).
func (r *TenantRepo) GetByID(ctx context.Context, id string) (*Tenant, error) {
	var t Tenant
	var modules []byte
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, slug, business_type, name, COALESCE(cnpj,''),
		       COALESCE(razao_social,''), plan, active_modules, timezone, currency, is_active, created_at
		FROM nexo.tenants WHERE id = $1`, id).
		Scan(&t.ID, &t.Slug, &t.BusinessType, &t.Name, &t.CNPJ,
			&t.RazaoSocial, &t.Plan, &modules, &t.Timezone, &t.Currency,
			&t.IsActive, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant %s não encontrado", id)
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(modules, &t.ModulesEnabled)
	return &t, nil
}

// GetBySlug busca tenant pelo slug (usado no subdomínio/login).
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var t Tenant
	var modules []byte
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, slug, business_type, name, COALESCE(cnpj,''),
		       plan, active_modules, timezone, currency, created_at
		FROM nexo.tenants WHERE slug = $1 AND is_active = TRUE`, slug).
		Scan(&t.ID, &t.Slug, &t.BusinessType, &t.Name, &t.CNPJ,
			&t.Plan, &modules, &t.Timezone, &t.Currency, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant '%s' não encontrado ou inativo", slug)
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(modules, &t.ModulesEnabled)
	return &t, nil
}

// UserRepo gerencia usuários com RLS.
type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo { return &UserRepo{db: db} }

// Create cria um novo usuário no tenant.
func (r *UserRepo) Create(ctx context.Context, u *User) error {
	return r.db.WithTenant(ctx, u.TenantID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `
			INSERT INTO nexo.users (tenant_id, email, name, role, password_hash)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id, created_at`,
			u.TenantID, u.Email, u.Name, u.Role, u.PasswordHash,
		).Scan(&u.ID, &u.CreatedAt)
	})
}

// GetByEmail busca usuário por email dentro do tenant (para login).
// Não usa RLS pois é chamado antes da autenticação — usa tenant_id diretamente.
func (r *UserRepo) GetByEmail(ctx context.Context, tenantID, email string) (*User, error) {
	var u User
	err := r.db.pool.QueryRowContext(ctx, `
		SELECT id, tenant_id, email, name, role, password_hash, is_active, created_at
		FROM nexo.users
		WHERE tenant_id = $1 AND email = $2 AND is_active = TRUE`, tenantID, email).
		Scan(&u.ID, &u.TenantID, &u.Email, &u.Name,
			&u.Role, &u.PasswordHash, &u.Active, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuário não encontrado")
	}
	return &u, err
}

// ListByTenant lista todos os usuários do tenant.
func (r *UserRepo) ListByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	var users []*User
	err := r.db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, email, name, role, is_active, created_at
			FROM nexo.users WHERE tenant_id = $1 ORDER BY created_at`, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var u User
			u.TenantID = tenantID
			if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt); err != nil {
				return err
			}
			users = append(users, &u)
		}
		return rows.Err()
	})
	return users, err
}
