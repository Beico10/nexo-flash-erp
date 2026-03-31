-- =============================================================================
-- NEXO FLASH ERP — Migração 000: Controle de migrations
-- Esta migration DEVE ser aplicada primeiro, antes de todas as outras.
-- Cria a tabela de controle para evitar re-aplicar migrations.
-- =============================================================================

-- Tabela de controle de migrations (aplicada pelo deploy.sh)
CREATE TABLE IF NOT EXISTS schema_migrations (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Criar usuário da aplicação se não existir
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user LOGIN PASSWORD 'PLACEHOLDER_REPLACED_BY_DEPLOY';
    END IF;
END$$;

-- Permissões mínimas para app_user
GRANT CONNECT ON DATABASE nexo_one TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_user;

-- Segurança: app_user NUNCA bypassa RLS
-- (BYPASSRLS é atributo de role, não de schema — controlado na criação do role)

COMMENT ON TABLE schema_migrations IS 'Controle de migrations — gerenciado pelo deploy.sh';
