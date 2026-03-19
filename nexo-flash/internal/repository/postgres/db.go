// Package postgres implementa a camada de acesso ao banco de dados PostgreSQL.
//
// SEGURANÇA CRÍTICA:
// Toda query DEVE ser executada dentro de uma transação que executa primeiro:
//   SET LOCAL app.tenant_id = '<uuid>'
//
// Isso ativa o Row Level Security (RLS) do PostgreSQL.
// A função WithTenant() garante isso automaticamente.
//
// NUNCA execute queries com db.QueryContext() diretamente sem usar WithTenant().
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // driver PostgreSQL
)

// DB encapsula o pool de conexões PostgreSQL com helpers de segurança.
type DB struct {
	pool *sql.DB
}

// Config agrupa as configurações de conexão.
type Config struct {
	DSN             string        // ex: "postgres://app_user:pass@localhost/nexoflash?sslmode=require"
	MaxOpenConns    int           // padrão: 25
	MaxIdleConns    int           // padrão: 10
	ConnMaxLifetime time.Duration // padrão: 5min
	ConnMaxIdleTime time.Duration // padrão: 1min
}

// New cria um novo pool de conexões PostgreSQL.
func New(cfg Config) (*DB, error) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 1 * time.Minute
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres.New: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verificar conectividade
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres.New: ping falhou: %w", err)
	}

	return &DB{pool: db}, nil
}

// WithTenant executa fn dentro de uma transação com o tenant_id configurado.
// OBRIGATÓRIO para todas as queries de negócio — ativa o RLS.
//
// Uso:
//
//	err := db.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
//	    return tx.QueryRowContext(ctx, "SELECT ...").Scan(...)
//	})
func (d *DB) WithTenant(ctx context.Context, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := d.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("postgres.WithTenant: begin: %w", err)
	}

	// SET LOCAL: afeta apenas esta transação — thread-safe com pool de conexões
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID); err != nil {
		tx.Rollback()
		return fmt.Errorf("postgres.WithTenant: set tenant_id: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// QueryRow executa uma query com tenant isolado e retorna uma linha.
// Atalho para queries simples de leitura.
func (d *DB) QueryRow(ctx context.Context, tenantID, query string, args ...any) (*sql.Row, error) {
	var row *sql.Row
	err := d.WithTenant(ctx, tenantID, func(tx *sql.Tx) error {
		row = tx.QueryRowContext(ctx, query, args...)
		return nil
	})
	return row, err
}

// Pool retorna o pool raw para migrações e operações de superusuário.
// NÃO use em código de negócio — bypassa o RLS.
func (d *DB) Pool() *sql.DB { return d.pool }

// Close fecha o pool de conexões.
func (d *DB) Close() error { return d.pool.Close() }

// Ping verifica a conectividade com o banco.
func (d *DB) Ping(ctx context.Context) error { return d.pool.PingContext(ctx) }
