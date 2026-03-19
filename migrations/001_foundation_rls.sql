-- ============================================================
-- Nexo One ERP — Migration 001: Fundação Multi-Tenant + RLS
-- ============================================================
-- LGPD (Lei 13.709/2018): isolamento obrigatório de dados por tenant
-- RLS garante que nenhum tenant acessa dados de outro — mesmo com bug na API
-- ============================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- Roles de banco (princípio do menor privilégio)
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'nexo_api') THEN
    CREATE ROLE nexo_api LOGIN PASSWORD 'TROQUE_VIA_VAULT';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'nexo_migrator') THEN
    CREATE ROLE nexo_migrator LOGIN PASSWORD 'TROQUE_VIA_VAULT';
  END IF;
END $$;

CREATE SCHEMA IF NOT EXISTS nexo AUTHORIZATION nexo_migrator;
SET search_path = nexo;

-- ─────────────────────────────────────────────
-- FUNÇÃO HELPER: current_tenant_id()
-- Usada em todas as policies RLS
-- ─────────────────────────────────────────────
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID
LANGUAGE sql STABLE SECURITY DEFINER AS $$
    SELECT NULLIF(current_setting('nexo.current_tenant_id', true), '')::UUID;
$$;

COMMENT ON FUNCTION current_tenant_id IS 
    'Retorna o tenant_id da sessão atual. Usado em todas as policies RLS.';

-- ─────────────────────────────────────────────
-- tenants
-- ─────────────────────────────────────────────
CREATE TABLE tenants (
  id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug           VARCHAR(100) NOT NULL UNIQUE,
  cnpj           VARCHAR(14) UNIQUE,
  razao_social   VARCHAR(255),
  name           VARCHAR(255) NOT NULL,
  business_type  VARCHAR(50) NOT NULL,
  tax_regime     VARCHAR(30) NOT NULL DEFAULT 'simples_nacional',
  plan           VARCHAR(30) NOT NULL DEFAULT 'starter',
  active_modules JSONB NOT NULL DEFAULT '[]',
  timezone       VARCHAR(50) NOT NULL DEFAULT 'America/Sao_Paulo',
  currency       VARCHAR(3) NOT NULL DEFAULT 'BRL',
  is_active      BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- users
-- ─────────────────────────────────────────────
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email         VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255) NOT NULL,
  full_name     VARCHAR(255),
  role          VARCHAR(50) NOT NULL DEFAULT 'operator',
  is_active     BOOLEAN NOT NULL DEFAULT true,
  last_login_at TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, email)
);

-- ─────────────────────────────────────────────
-- products (catálogo multi-nicho via JSONB)
-- ─────────────────────────────────────────────
CREATE TABLE products (
  id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id      UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  sku            VARCHAR(100) NOT NULL,
  name           VARCHAR(255) NOT NULL,
  ncm            VARCHAR(8),
  unit           VARCHAR(10) NOT NULL DEFAULT 'UN',
  cost_price     NUMERIC(15,2) NOT NULL DEFAULT 0,
  sale_price     NUMERIC(15,2) NOT NULL DEFAULT 0,
  niche_data     JSONB NOT NULL DEFAULT '{}',
  stock_quantity NUMERIC(15,3) NOT NULL DEFAULT 0,
  is_active      BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, sku)
);

-- ─────────────────────────────────────────────
-- tax_results — auditoria fiscal IMUTÁVEL
-- APPEND-ONLY: sem UPDATE, sem DELETE (reforçado pelo RLS)
-- ─────────────────────────────────────────────
CREATE TABLE tax_results (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  ncm             VARCHAR(8) NOT NULL,
  category        VARCHAR(30) NOT NULL,
  base_value      NUMERIC(15,2) NOT NULL,
  ibs_rate        NUMERIC(8,6) NOT NULL,
  cbs_rate        NUMERIC(8,6) NOT NULL,
  ibs_amount      NUMERIC(15,2) NOT NULL,
  cbs_amount      NUMERIC(15,2) NOT NULL,
  total_tax       NUMERIC(15,2) NOT NULL,
  credit_amount   NUMERIC(15,2) NOT NULL,
  debit_amount    NUMERIC(15,2) NOT NULL,
  is_zero_rated   BOOLEAN NOT NULL DEFAULT false,
  legal_basis     TEXT NOT NULL,
  approval_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  approved_by     UUID REFERENCES users(id),
  approved_at     TIMESTAMPTZ,
  calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- ai_suggestions — Human-in-the-Loop obrigatório
-- IA nunca altera dados sem aprovação humana
-- ─────────────────────────────────────────────
CREATE TABLE ai_suggestions (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  suggestion_type VARCHAR(100) NOT NULL,
  target_table    VARCHAR(100),
  target_id       UUID,
  suggested_data  JSONB NOT NULL,
  reason          TEXT,
  confidence      NUMERIC(3,2) DEFAULT 0.5,
  created_by_ai   VARCHAR(50) DEFAULT 'system',
  expires_at      TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '7 days'),
  status          VARCHAR(20) NOT NULL DEFAULT 'pending',
  reviewed_by     UUID REFERENCES users(id),
  reviewed_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- invoices — documentos fiscais
-- ─────────────────────────────────────────────
CREATE TABLE invoices (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  doc_type        VARCHAR(10) NOT NULL,
  status          VARCHAR(20) NOT NULL DEFAULT 'draft',
  number          INTEGER,
  series          INTEGER NOT NULL DEFAULT 1,
  access_key      VARCHAR(44) UNIQUE,
  xml_storage_key VARCHAR(500),
  total_value     NUMERIC(15,2) NOT NULL DEFAULT 0,
  tax_result_id   UUID REFERENCES tax_results(id),
  is_contingency  BOOLEAN NOT NULL DEFAULT false,
  emitted_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- audit_log — Auditoria imutável
-- ─────────────────────────────────────────────
CREATE TABLE audit_log (
  id          BIGSERIAL PRIMARY KEY,
  tenant_id   UUID,
  action      VARCHAR(20) NOT NULL,
  table_name  VARCHAR(100) NOT NULL,
  record_id   UUID,
  old_data    JSONB,
  new_data    JSONB,
  user_id     UUID,
  ip_address  INET,
  user_agent  TEXT,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- fiscal_ncm_rates — Tabela de alíquotas NCM
-- ─────────────────────────────────────────────
CREATE TABLE fiscal_ncm_rates (
  id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  ncm_code         VARCHAR(8) NOT NULL,
  ncm_description  TEXT,
  ibs_rate         NUMERIC(8,6) NOT NULL DEFAULT 0.0925,
  cbs_rate         NUMERIC(8,6) NOT NULL DEFAULT 0.0375,
  selective_rate   NUMERIC(8,6) DEFAULT 0,
  basket_reduced   BOOLEAN NOT NULL DEFAULT FALSE,
  basket_type      VARCHAR(20),
  transition_year  INTEGER DEFAULT 2026,
  transition_factor NUMERIC(4,2) DEFAULT 0.10,
  valid_from       DATE NOT NULL DEFAULT CURRENT_DATE,
  valid_until      DATE,
  source_lei       VARCHAR(100),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(ncm_code, valid_from)
);

-- ─────────────────────────────────────────────
-- ROW LEVEL SECURITY — isolamento total por tenant
-- ─────────────────────────────────────────────
ALTER TABLE tenants        ENABLE ROW LEVEL SECURITY;
ALTER TABLE users          ENABLE ROW LEVEL SECURITY;
ALTER TABLE products       ENABLE ROW LEVEL SECURITY;
ALTER TABLE tax_results    ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_suggestions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices       ENABLE ROW LEVEL SECURITY;

-- Policy para tenants (admin pode ver apenas o próprio)
CREATE POLICY rls_tenants ON tenants
  USING (id = current_tenant_id() OR current_tenant_id() IS NULL);

CREATE POLICY rls_users ON users
  USING (tenant_id = current_tenant_id());

CREATE POLICY rls_products ON products
  USING (tenant_id = current_tenant_id());

CREATE POLICY rls_tax_read ON tax_results FOR SELECT
  USING (tenant_id = current_tenant_id());
CREATE POLICY rls_tax_insert ON tax_results FOR INSERT
  WITH CHECK (tenant_id = current_tenant_id());

CREATE POLICY rls_ai ON ai_suggestions
  USING (tenant_id = current_tenant_id());

CREATE POLICY rls_invoices ON invoices
  USING (tenant_id = current_tenant_id());

-- ─────────────────────────────────────────────
-- Permissões nexo_api (menor privilégio)
-- ─────────────────────────────────────────────
GRANT USAGE ON SCHEMA nexo TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON tenants TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON users TO nexo_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON products TO nexo_api;
GRANT SELECT, INSERT ON tax_results TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON ai_suggestions TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON invoices TO nexo_api;
GRANT SELECT ON fiscal_ncm_rates TO nexo_api;
GRANT INSERT ON audit_log TO nexo_api;

-- ─────────────────────────────────────────────
-- Índices
-- ─────────────────────────────────────────────
CREATE INDEX idx_tenants_slug         ON tenants(slug);
CREATE INDEX idx_users_tenant         ON users(tenant_id);
CREATE INDEX idx_users_email          ON users(tenant_id, email);
CREATE INDEX idx_products_tenant      ON products(tenant_id);
CREATE INDEX idx_products_sku         ON products(tenant_id, sku);
CREATE INDEX idx_products_ncm         ON products(ncm);
CREATE INDEX idx_products_name_trgm   ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_tax_tenant           ON tax_results(tenant_id);
CREATE INDEX idx_tax_pending          ON tax_results(approval_status) WHERE approval_status = 'PENDING';
CREATE INDEX idx_ai_pending           ON ai_suggestions(tenant_id, status) WHERE status = 'pending';
CREATE INDEX idx_invoices_tenant      ON invoices(tenant_id);
CREATE INDEX idx_invoices_access_key  ON invoices(access_key) WHERE access_key IS NOT NULL;
CREATE INDEX idx_audit_tenant_date    ON audit_log(tenant_id, occurred_at DESC);
CREATE INDEX idx_ncm_rates_code       ON fiscal_ncm_rates(ncm_code);

-- Trigger updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW(); RETURN NEW; END; $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tenants_updated  BEFORE UPDATE ON tenants  FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_products_updated BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at();

COMMIT;
