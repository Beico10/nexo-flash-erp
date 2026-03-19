-- ============================================================
-- NEXO ONE ERP — Migration 007: Trial Control + Journey Tracking
-- ============================================================
-- Verificação por WhatsApp (custo zero)
-- Tracking completo da jornada do usuário
-- Analytics para identificar onde o lead trava
-- ============================================================

BEGIN;

SET search_path = nexo;

-- ═══════════════════════════════════════════════════════════
-- CONTROLE DE TRIAL (1 por pessoa)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE trial_controls (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Identificadores únicos
    phone_number    VARCHAR(20) NOT NULL,         -- +5511999999999
    phone_hash      VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 do telefone (LGPD)
    email           VARCHAR(255),
    cnpj            VARCHAR(14),
    
    -- Verificação WhatsApp
    verification_code   VARCHAR(6),
    code_expires_at     TIMESTAMPTZ,
    verified_at         TIMESTAMPTZ,
    verification_method VARCHAR(20) DEFAULT 'whatsapp',
    
    -- Device Fingerprint (detectar tentativas de burlar)
    device_hash     VARCHAR(64),
    ip_address      INET,
    user_agent      TEXT,
    
    -- Vínculo com tenant
    tenant_id       UUID REFERENCES tenants(id),
    
    -- Flags de controle
    is_blocked      BOOLEAN DEFAULT FALSE,
    block_reason    TEXT,
    abuse_score     INT DEFAULT 0,  -- 0-100, quanto maior mais suspeito
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trial_phone_hash ON trial_controls(phone_hash);
CREATE INDEX idx_trial_device ON trial_controls(device_hash) WHERE device_hash IS NOT NULL;
CREATE INDEX idx_trial_tenant ON trial_controls(tenant_id);

-- ═══════════════════════════════════════════════════════════
-- PROGRESSO DO ONBOARDING (por nicho)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE onboarding_progress (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id),
    business_type   VARCHAR(50) NOT NULL,
    
    -- Progresso
    current_step    VARCHAR(50) NOT NULL DEFAULT 'welcome',
    total_steps     INT NOT NULL DEFAULT 5,
    completed_steps JSONB NOT NULL DEFAULT '[]',
    -- Ex: ["welcome", "import_xml", "first_sale"]
    
    -- Tempo gasto
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    last_activity   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Comportamento
    skipped         BOOLEAN DEFAULT FALSE,
    skipped_at      TIMESTAMPTZ,
    returned_count  INT DEFAULT 0,  -- quantas vezes voltou ao onboarding
    
    UNIQUE(tenant_id)
);

-- ═══════════════════════════════════════════════════════════
-- DEFINIÇÃO DOS PASSOS POR NICHO
-- ═══════════════════════════════════════════════════════════
CREATE TABLE onboarding_steps (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_type   VARCHAR(50) NOT NULL,
    step_code       VARCHAR(50) NOT NULL,
    step_order      INT NOT NULL,
    
    -- Conteúdo
    title           VARCHAR(100) NOT NULL,
    description     TEXT,
    icon            VARCHAR(50),
    
    -- Configuração
    is_required     BOOLEAN DEFAULT FALSE,  -- Obrigatório para usar o sistema?
    is_skippable    BOOLEAN DEFAULT TRUE,   -- Pode pular?
    estimated_time  INT DEFAULT 60,         -- Segundos estimados
    
    -- Ação
    action_type     VARCHAR(50) NOT NULL,   -- 'import_xml', 'form', 'wizard', 'video'
    action_config   JSONB,                  -- Configuração específica da ação
    
    -- Incentivo
    reward_text     TEXT,                   -- "Complete e ganhe +1 dia de trial"
    reward_days     INT DEFAULT 0,          -- Dias extras de trial
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(business_type, step_code)
);

-- ═══════════════════════════════════════════════════════════
-- TRACKING DE JORNADA (Analytics)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE journey_events (
    id              BIGSERIAL PRIMARY KEY,
    
    -- Quem
    tenant_id       UUID REFERENCES tenants(id),
    user_id         UUID REFERENCES users(id),
    anonymous_id    VARCHAR(64),            -- Para eventos antes do cadastro
    
    -- O quê
    event_name      VARCHAR(100) NOT NULL,  -- 'page_view', 'button_click', 'form_submit', etc.
    event_category  VARCHAR(50) NOT NULL,   -- 'onboarding', 'activation', 'engagement', 'conversion'
    
    -- Onde
    page_path       VARCHAR(255),
    page_title      VARCHAR(255),
    referrer        VARCHAR(500),
    
    -- Contexto
    properties      JSONB DEFAULT '{}',
    -- Ex: {"step": "import_xml", "action": "skip", "reason": "no_xml_available"}
    
    -- Funil
    funnel_stage    VARCHAR(50),            -- 'awareness', 'interest', 'decision', 'action', 'retention'
    
    -- Device
    device_type     VARCHAR(20),            -- 'desktop', 'mobile', 'tablet'
    browser         VARCHAR(50),
    os              VARCHAR(50),
    screen_size     VARCHAR(20),
    
    -- Sessão
    session_id      VARCHAR(64),
    session_start   TIMESTAMPTZ,
    
    -- Tempo
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    time_on_page    INT,                    -- Segundos na página anterior
    
    -- Performance
    page_load_time  INT                     -- Milissegundos
);

-- Índices para queries de analytics
CREATE INDEX idx_journey_tenant ON journey_events(tenant_id, occurred_at DESC);
CREATE INDEX idx_journey_event ON journey_events(event_name, occurred_at DESC);
CREATE INDEX idx_journey_funnel ON journey_events(funnel_stage, occurred_at DESC);
CREATE INDEX idx_journey_session ON journey_events(session_id);
CREATE INDEX idx_journey_anonymous ON journey_events(anonymous_id) WHERE anonymous_id IS NOT NULL;

-- Particionar por mês para performance (dados crescem rápido)
-- Em produção, considerar TimescaleDB ou particionamento nativo

-- ═══════════════════════════════════════════════════════════
-- FUNIL DE CONVERSÃO (Agregado)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE funnel_daily (
    id              BIGSERIAL PRIMARY KEY,
    date            DATE NOT NULL,
    business_type   VARCHAR(50),            -- NULL = todos
    
    -- Métricas do funil
    visits              INT DEFAULT 0,      -- Visitou o site
    signups_started     INT DEFAULT 0,      -- Começou cadastro
    signups_completed   INT DEFAULT 0,      -- Completou cadastro
    phone_verified      INT DEFAULT 0,      -- Verificou WhatsApp
    onboarding_started  INT DEFAULT 0,      -- Iniciou onboarding
    onboarding_completed INT DEFAULT 0,     -- Completou onboarding
    first_action        INT DEFAULT 0,      -- Primeira ação real (OS, venda, etc)
    trial_converted     INT DEFAULT 0,      -- Converteu para pago
    
    -- Abandono por etapa
    drop_signup         INT DEFAULT 0,      -- Abandonou no cadastro
    drop_verification   INT DEFAULT 0,      -- Não verificou telefone
    drop_onboarding     INT DEFAULT 0,      -- Abandonou onboarding
    drop_activation     INT DEFAULT 0,      -- Não fez primeira ação
    
    -- Tempo médio
    avg_time_to_verify  INT,                -- Segundos
    avg_time_to_first_action INT,
    avg_onboarding_duration INT,
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(date, business_type)
);

-- ═══════════════════════════════════════════════════════════
-- PONTOS DE TRAVAMENTO (Onde o usuário para)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE drop_points (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID REFERENCES tenants(id),
    user_id         UUID REFERENCES users(id),
    
    -- Onde travou
    stage           VARCHAR(50) NOT NULL,   -- 'signup', 'verification', 'onboarding', 'activation'
    step_code       VARCHAR(50),            -- Passo específico do onboarding
    page_path       VARCHAR(255),
    
    -- Quanto tempo ficou parado
    stuck_since     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    days_stuck      INT GENERATED ALWAYS AS (EXTRACT(DAY FROM (NOW() - stuck_since))) STORED,
    
    -- Tentativas de reengajamento
    reminder_sent   BOOLEAN DEFAULT FALSE,
    reminder_count  INT DEFAULT 0,
    
    -- Resolução
    resolved        BOOLEAN DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    resolution      VARCHAR(50),            -- 'completed', 'skipped', 'churned', 'support_helped'
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_drop_unresolved ON drop_points(stage, days_stuck) WHERE resolved = FALSE;
CREATE INDEX idx_drop_tenant ON drop_points(tenant_id);

-- ═══════════════════════════════════════════════════════════
-- SEED: Passos de Onboarding por Nicho
-- ═══════════════════════════════════════════════════════════

-- MECÂNICA
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('mechanic', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('mechanic', 'import_parts', 2, 'Importar Peças', 'Importe o XML de uma NF-e de compra e cadastre suas peças automaticamente', 'upload', TRUE, 120, 'import_xml', 1),
('mechanic', 'labor_rates', 3, 'Tabela de Mão de Obra', 'Configure o valor da hora de cada tipo de serviço', 'clock', TRUE, 60, 'form', 0),
('mechanic', 'first_os', 4, 'Primeira O.S.', 'Crie uma ordem de serviço de teste', 'clipboard', TRUE, 120, 'wizard', 1),
('mechanic', 'invite_team', 5, 'Convidar Equipe', 'Adicione mecânicos e atendentes', 'users', TRUE, 60, 'form', 1);

-- PADARIA
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('bakery', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('bakery', 'import_products', 2, 'Importar Produtos', 'Importe o XML de uma NF-e e cadastre produtos automaticamente', 'upload', TRUE, 120, 'import_xml', 1),
('bakery', 'basket_items', 3, 'Cesta Básica', 'Marque os itens com alíquota zero (pão, leite, etc)', 'shopping-basket', TRUE, 60, 'form', 0),
('bakery', 'first_sale', 4, 'Primeira Venda', 'Faça uma venda de teste no PDV', 'shopping-cart', TRUE, 90, 'wizard', 1),
('bakery', 'invite_team', 5, 'Convidar Equipe', 'Adicione operadores de caixa', 'users', TRUE, 60, 'form', 1);

-- ESTÉTICA
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('aesthetics', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('aesthetics', 'professionals', 2, 'Cadastrar Profissionais', 'Adicione os profissionais do seu salão', 'users', FALSE, 90, 'form', 0),
('aesthetics', 'services', 3, 'Cadastrar Serviços', 'Configure serviços, preços e duração', 'scissors', FALSE, 120, 'form', 1),
('aesthetics', 'schedule', 4, 'Horários', 'Defina os horários de funcionamento', 'calendar', TRUE, 60, 'form', 0),
('aesthetics', 'first_appointment', 5, 'Primeiro Agendamento', 'Agende um cliente de teste', 'calendar-plus', TRUE, 60, 'wizard', 1);

-- LOGÍSTICA
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('logistics', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('logistics', 'vehicles', 2, 'Cadastrar Veículos', 'Adicione sua frota (placa, tipo, capacidade)', 'truck', FALSE, 120, 'form', 1),
('logistics', 'freight_table', 3, 'Tabela de Frete', 'Configure sua tabela de preços padrão', 'table', TRUE, 120, 'form', 1),
('logistics', 'first_shipper', 4, 'Primeiro Embarcador', 'Cadastre um cliente (embarcador)', 'building', TRUE, 60, 'form', 0),
('logistics', 'first_cte', 5, 'Primeiro CT-e', 'Simule a emissão de um CT-e de teste', 'file-text', TRUE, 120, 'wizard', 1);

-- INDÚSTRIA
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('industry', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('industry', 'import_materials', 2, 'Importar Insumos', 'Importe o XML de uma NF-e de compra de matéria-prima', 'upload', TRUE, 120, 'import_xml', 1),
('industry', 'first_product', 3, 'Primeiro Produto', 'Cadastre um produto acabado', 'package', FALSE, 90, 'form', 0),
('industry', 'first_bom', 4, 'Primeira Ficha Técnica', 'Crie a ficha técnica (BOM) do produto', 'list', TRUE, 180, 'wizard', 2),
('industry', 'first_order', 5, 'Primeira Ordem', 'Crie uma ordem de produção de teste', 'clipboard', TRUE, 120, 'wizard', 1);

-- CALÇADOS
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_skippable, estimated_time, action_type, reward_days) VALUES
('shoes', 'welcome', 1, 'Bem-vindo!', 'Conheça seu novo sistema de gestão', 'rocket', FALSE, 30, 'video', 0),
('shoes', 'import_products', 2, 'Importar Produtos', 'Importe o XML de uma NF-e de compra', 'upload', TRUE, 120, 'import_xml', 1),
('shoes', 'first_grid', 3, 'Primeira Grade', 'Configure a grade de cores e tamanhos de um produto', 'grid', FALSE, 120, 'wizard', 1),
('shoes', 'sellers', 4, 'Cadastrar Vendedores', 'Adicione vendedores e configure comissões', 'users', TRUE, 90, 'form', 0),
('shoes', 'first_sale', 5, 'Primeira Venda', 'Faça uma venda de teste', 'shopping-cart', TRUE, 90, 'wizard', 1);

-- ═══════════════════════════════════════════════════════════
-- VIEWS PARA DASHBOARD ADMIN
-- ═══════════════════════════════════════════════════════════

-- Funil de hoje
CREATE OR REPLACE VIEW funnel_today AS
SELECT
    COALESCE(t.business_type, 'all') as business_type,
    COUNT(DISTINCT tc.id) FILTER (WHERE tc.created_at::date = CURRENT_DATE) as signups_today,
    COUNT(DISTINCT tc.id) FILTER (WHERE tc.verified_at::date = CURRENT_DATE) as verified_today,
    COUNT(DISTINCT op.id) FILTER (WHERE op.completed_at::date = CURRENT_DATE) as onboarding_completed_today,
    COUNT(DISTINCT bs.id) FILTER (WHERE bs.status = 'active' AND bs.updated_at::date = CURRENT_DATE) as converted_today
FROM trial_controls tc
LEFT JOIN tenants t ON t.id = tc.tenant_id
LEFT JOIN onboarding_progress op ON op.tenant_id = tc.tenant_id
LEFT JOIN billing_subscriptions bs ON bs.tenant_id = tc.tenant_id
GROUP BY ROLLUP(t.business_type);

-- Usuários travados
CREATE OR REPLACE VIEW stuck_users AS
SELECT
    dp.tenant_id,
    t.name as tenant_name,
    t.business_type,
    u.email,
    u.name as user_name,
    dp.stage,
    dp.step_code,
    dp.days_stuck,
    dp.reminder_count,
    tc.phone_number
FROM drop_points dp
JOIN tenants t ON t.id = dp.tenant_id
JOIN users u ON u.id = dp.user_id
LEFT JOIN trial_controls tc ON tc.tenant_id = dp.tenant_id
WHERE dp.resolved = FALSE
ORDER BY dp.days_stuck DESC;

-- ═══════════════════════════════════════════════════════════
-- FUNÇÃO: Registrar evento de jornada
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE FUNCTION track_event(
    p_tenant_id UUID,
    p_user_id UUID,
    p_event_name VARCHAR(100),
    p_event_category VARCHAR(50),
    p_properties JSONB DEFAULT '{}'
) RETURNS BIGINT
LANGUAGE plpgsql AS $$
DECLARE
    v_id BIGINT;
BEGIN
    INSERT INTO journey_events (tenant_id, user_id, event_name, event_category, properties)
    VALUES (p_tenant_id, p_user_id, p_event_name, p_event_category, p_properties)
    RETURNING id INTO v_id;
    
    RETURN v_id;
END;
$$;

-- ═══════════════════════════════════════════════════════════
-- PERMISSÕES
-- ═══════════════════════════════════════════════════════════
GRANT SELECT, INSERT, UPDATE ON trial_controls TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON onboarding_progress TO nexo_api;
GRANT SELECT ON onboarding_steps TO nexo_api;
GRANT INSERT, SELECT ON journey_events TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON funnel_daily TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON drop_points TO nexo_api;
GRANT SELECT ON funnel_today TO nexo_api;
GRANT SELECT ON stuck_users TO nexo_api;

COMMIT;
