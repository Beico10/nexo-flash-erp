-- ============================================================
-- NEXO ONE ERP — Migration 006: Sistema de Planos e Assinaturas
-- ============================================================
-- Self-service: Trial 7 dias → Conversão → Upgrade automático
-- Admin Master configura preços sem mexer em código
-- ============================================================

BEGIN;

SET search_path = nexo;

-- ═══════════════════════════════════════════════════════════
-- TABELA: billing_plans (Configurável pelo Admin Master)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE billing_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            VARCHAR(50) NOT NULL UNIQUE,  -- 'micro', 'starter', 'pro', 'business', 'enterprise'
    name            VARCHAR(100) NOT NULL,        -- 'Micro', 'Starter', etc.
    description     TEXT,
    
    -- Preços (Admin Master altera aqui)
    price_monthly   NUMERIC(10,2) NOT NULL,       -- Preço mensal
    price_yearly    NUMERIC(10,2),                -- Preço anual (desconto)
    setup_fee       NUMERIC(10,2) DEFAULT 0,      -- Taxa de adesão
    
    -- Limites (NULL = ilimitado)
    max_users           INT,                      -- Usuários permitidos
    max_transactions    INT,                      -- Transações/mês
    max_products        INT,                      -- Produtos cadastrados
    max_invoices        INT,                      -- Notas fiscais/mês
    max_storage_mb      INT,                      -- Armazenamento em MB
    
    -- Recursos habilitados (JSONB flexível)
    features        JSONB NOT NULL DEFAULT '{
        "fiscal_2026": false,
        "baas_pix": false,
        "baas_boleto": false,
        "baas_split": false,
        "whatsapp": false,
        "ai_copilot": false,
        "ai_concierge": false,
        "roteirizador": false,
        "multi_pdv": false,
        "api_access": false,
        "priority_support": false,
        "custom_reports": false
    }',
    
    -- Nichos permitidos (vazio = todos)
    allowed_niches  TEXT[] DEFAULT '{}',
    
    -- Ordenação e status
    display_order   INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured     BOOLEAN DEFAULT FALSE,        -- Destacado na página de preços
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: billing_subscriptions (Assinatura de cada Tenant)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE billing_subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id         UUID NOT NULL REFERENCES billing_plans(id),
    
    -- Status
    status          VARCHAR(30) NOT NULL DEFAULT 'trialing'
                    CHECK (status IN ('trialing', 'active', 'past_due', 'cancelled', 'expired')),
    
    -- Datas importantes
    trial_ends_at   TIMESTAMPTZ,                  -- Fim do trial (7 dias após criação)
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancelled_at    TIMESTAMPTZ,
    
    -- Pagamento
    payment_method  VARCHAR(30),                  -- 'pix', 'credit_card', 'boleto'
    billing_cycle   VARCHAR(20) NOT NULL DEFAULT 'monthly'
                    CHECK (billing_cycle IN ('monthly', 'yearly')),
    
    -- Desconto (cupom, early adopter, etc.)
    discount_percent    NUMERIC(5,2) DEFAULT 0,
    discount_reason     VARCHAR(100),
    discount_expires_at TIMESTAMPTZ,
    
    -- Uso atual (atualizado por triggers/jobs)
    current_users       INT DEFAULT 0,
    current_transactions INT DEFAULT 0,
    current_products    INT DEFAULT 0,
    current_invoices    INT DEFAULT 0,
    current_storage_mb  INT DEFAULT 0,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(tenant_id)  -- Um tenant = uma assinatura
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: billing_invoices (Faturas geradas)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE billing_invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    subscription_id UUID NOT NULL REFERENCES billing_subscriptions(id),
    
    -- Valores
    amount          NUMERIC(10,2) NOT NULL,
    discount        NUMERIC(10,2) DEFAULT 0,
    total           NUMERIC(10,2) NOT NULL,
    
    -- Status
    status          VARCHAR(30) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'failed', 'refunded', 'cancelled')),
    
    -- Pagamento
    payment_method  VARCHAR(30),
    paid_at         TIMESTAMPTZ,
    pix_tx_id       TEXT,                         -- ID da transação PIX
    boleto_url      TEXT,
    
    -- Período cobrado
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    
    -- Vencimento
    due_date        DATE NOT NULL,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: billing_usage_log (Histórico de uso para analytics)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE billing_usage_log (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    metric          VARCHAR(50) NOT NULL,         -- 'users', 'transactions', 'products', etc.
    value           INT NOT NULL,
    recorded_at     DATE NOT NULL DEFAULT CURRENT_DATE,
    
    UNIQUE(tenant_id, metric, recorded_at)
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: billing_coupons (Cupons de desconto)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE billing_coupons (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            VARCHAR(50) NOT NULL UNIQUE,  -- 'EARLY30', 'BLACKFRIDAY50'
    description     TEXT,
    
    -- Desconto
    discount_type   VARCHAR(20) NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
    discount_value  NUMERIC(10,2) NOT NULL,       -- 30 (%) ou 50.00 (R$)
    
    -- Restrições
    max_uses        INT,                          -- NULL = ilimitado
    current_uses    INT DEFAULT 0,
    min_plan_price  NUMERIC(10,2),                -- Preço mínimo do plano
    allowed_plans   TEXT[],                       -- Planos específicos
    
    -- Validade
    valid_from      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until     TIMESTAMPTZ,
    
    -- Duração do desconto
    duration_months INT DEFAULT 1,                -- Quantos meses o desconto vale
    
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- ÍNDICES
-- ═══════════════════════════════════════════════════════════
CREATE INDEX idx_subscriptions_tenant ON billing_subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_status ON billing_subscriptions(status) WHERE status IN ('trialing', 'past_due');
CREATE INDEX idx_subscriptions_trial_ends ON billing_subscriptions(trial_ends_at) WHERE status = 'trialing';
CREATE INDEX idx_invoices_tenant ON billing_invoices(tenant_id);
CREATE INDEX idx_invoices_status ON billing_invoices(status, due_date) WHERE status = 'pending';
CREATE INDEX idx_usage_tenant_date ON billing_usage_log(tenant_id, recorded_at DESC);
CREATE INDEX idx_coupons_code ON billing_coupons(code) WHERE is_active = TRUE;

-- ═══════════════════════════════════════════════════════════
-- TRIGGERS
-- ═══════════════════════════════════════════════════════════
CREATE TRIGGER trg_plans_updated 
    BEFORE UPDATE ON billing_plans 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_subscriptions_updated 
    BEFORE UPDATE ON billing_subscriptions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ═══════════════════════════════════════════════════════════
-- FUNÇÃO: check_plan_limit()
-- Verifica se tenant pode executar ação baseado no plano
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE FUNCTION check_plan_limit(
    p_tenant_id UUID,
    p_metric VARCHAR(50)
) RETURNS BOOLEAN
LANGUAGE plpgsql SECURITY DEFINER AS $$
DECLARE
    v_limit INT;
    v_current INT;
    v_status VARCHAR(30);
BEGIN
    -- Buscar limite do plano e uso atual
    SELECT 
        CASE p_metric
            WHEN 'users' THEN bp.max_users
            WHEN 'transactions' THEN bp.max_transactions
            WHEN 'products' THEN bp.max_products
            WHEN 'invoices' THEN bp.max_invoices
            WHEN 'storage_mb' THEN bp.max_storage_mb
        END,
        CASE p_metric
            WHEN 'users' THEN bs.current_users
            WHEN 'transactions' THEN bs.current_transactions
            WHEN 'products' THEN bs.current_products
            WHEN 'invoices' THEN bs.current_invoices
            WHEN 'storage_mb' THEN bs.current_storage_mb
        END,
        bs.status
    INTO v_limit, v_current, v_status
    FROM billing_subscriptions bs
    JOIN billing_plans bp ON bp.id = bs.plan_id
    WHERE bs.tenant_id = p_tenant_id;
    
    -- Se não encontrou, bloqueia
    IF v_status IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Se expirado ou cancelado, bloqueia
    IF v_status IN ('expired', 'cancelled') THEN
        RETURN FALSE;
    END IF;
    
    -- Se limite é NULL, é ilimitado
    IF v_limit IS NULL THEN
        RETURN TRUE;
    END IF;
    
    -- Verifica se está dentro do limite
    RETURN v_current < v_limit;
END;
$$;

-- ═══════════════════════════════════════════════════════════
-- FUNÇÃO: get_tenant_features()
-- Retorna features habilitadas do plano do tenant
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE FUNCTION get_tenant_features(p_tenant_id UUID)
RETURNS JSONB
LANGUAGE sql STABLE SECURITY DEFINER AS $$
    SELECT COALESCE(bp.features, '{}'::jsonb)
    FROM billing_subscriptions bs
    JOIN billing_plans bp ON bp.id = bs.plan_id
    WHERE bs.tenant_id = p_tenant_id
      AND bs.status IN ('trialing', 'active');
$$;

-- ═══════════════════════════════════════════════════════════
-- SEED: Planos Iniciais (Admin Master pode alterar depois)
-- ═══════════════════════════════════════════════════════════
INSERT INTO billing_plans (code, name, description, price_monthly, price_yearly, setup_fee, max_users, max_transactions, max_products, max_invoices, features, display_order, is_featured) VALUES

-- MICRO: Manicure, Pipoqueiro, Autônomo
('micro', 'Micro', 'Para autônomos e micro-empreendedores', 
 47.00, 470.00, 0,
 1, 100, 50, 30,
 '{
    "fiscal_2026": true,
    "baas_pix": true,
    "baas_boleto": false,
    "baas_split": false,
    "whatsapp": false,
    "ai_copilot": false,
    "ai_concierge": false,
    "roteirizador": false,
    "multi_pdv": false,
    "api_access": false,
    "priority_support": false,
    "custom_reports": false
 }',
 1, false),

-- STARTER: Oficina pequena, Salão individual
('starter', 'Starter', 'Para pequenos negócios em crescimento',
 97.00, 970.00, 0,
 2, NULL, 200, 100,
 '{
    "fiscal_2026": true,
    "baas_pix": true,
    "baas_boleto": true,
    "baas_split": false,
    "whatsapp": true,
    "ai_copilot": false,
    "ai_concierge": true,
    "roteirizador": false,
    "multi_pdv": false,
    "api_access": false,
    "priority_support": false,
    "custom_reports": false
 }',
 2, false),

-- PRO: Padaria, Loja, Salão com equipe
('pro', 'Pro', 'Para negócios estabelecidos que querem crescer',
 197.00, 1970.00, 0,
 5, NULL, 1000, 500,
 '{
    "fiscal_2026": true,
    "baas_pix": true,
    "baas_boleto": true,
    "baas_split": true,
    "whatsapp": true,
    "ai_copilot": true,
    "ai_concierge": true,
    "roteirizador": false,
    "multi_pdv": true,
    "api_access": false,
    "priority_support": false,
    "custom_reports": true
 }',
 3, true),  -- DESTACADO

-- BUSINESS: Indústria pequena, Transportadora
('business', 'Business', 'Para empresas com operações complexas',
 397.00, 3970.00, 297.00,
 10, NULL, 5000, NULL,
 '{
    "fiscal_2026": true,
    "baas_pix": true,
    "baas_boleto": true,
    "baas_split": true,
    "whatsapp": true,
    "ai_copilot": true,
    "ai_concierge": true,
    "roteirizador": true,
    "multi_pdv": true,
    "api_access": true,
    "priority_support": true,
    "custom_reports": true
 }',
 4, false),

-- ENTERPRISE: Indústria média/grande
('enterprise', 'Enterprise', 'Para grandes operações com SLA dedicado',
 997.00, 9970.00, 497.00,
 NULL, NULL, NULL, NULL,
 '{
    "fiscal_2026": true,
    "baas_pix": true,
    "baas_boleto": true,
    "baas_split": true,
    "whatsapp": true,
    "ai_copilot": true,
    "ai_concierge": true,
    "roteirizador": true,
    "multi_pdv": true,
    "api_access": true,
    "priority_support": true,
    "custom_reports": true,
    "dedicated_support": true,
    "sla_99_9": true,
    "custom_integrations": true
 }',
 5, false);

-- ═══════════════════════════════════════════════════════════
-- SEED: Cupom Early Adopter (100% desconto no setup)
-- ═══════════════════════════════════════════════════════════
INSERT INTO billing_coupons (code, description, discount_type, discount_value, max_uses, duration_months, valid_until)
VALUES 
('EARLY30', 'Early Adopter - 30 primeiros parceiros', 'percent', 100, 30, 12, NOW() + INTERVAL '90 days'),
('LANCAMENTO50', 'Lançamento - 50% desconto 3 meses', 'percent', 50, 100, 3, NOW() + INTERVAL '60 days');

-- ═══════════════════════════════════════════════════════════
-- PERMISSÕES
-- ═══════════════════════════════════════════════════════════
GRANT SELECT ON billing_plans TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON billing_subscriptions TO nexo_api;
GRANT SELECT, INSERT ON billing_invoices TO nexo_api;
GRANT SELECT, INSERT ON billing_usage_log TO nexo_api;
GRANT SELECT, UPDATE ON billing_coupons TO nexo_api;

COMMIT;
