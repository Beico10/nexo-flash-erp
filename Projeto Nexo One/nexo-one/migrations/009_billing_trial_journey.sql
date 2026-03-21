-- =============================================================================
-- NEXO ONE ERP — Migration 009: Billing, Trial, Journey e Onboarding
-- =============================================================================

-- ── BILLING: Planos ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS billing_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    price_monthly   NUMERIC(10,2) NOT NULL DEFAULT 0,
    price_yearly    NUMERIC(10,2) NOT NULL DEFAULT 0,
    setup_fee       NUMERIC(10,2) DEFAULT 0,
    max_users       INT,           -- NULL = ilimitado
    max_transactions INT,
    max_products    INT,
    max_invoices    INT,
    max_storage_mb  INT,
    features        JSONB NOT NULL DEFAULT '{}',
    allowed_niches  JSONB NOT NULL DEFAULT '[]',
    display_order   INT DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    is_featured     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Seed dos planos iniciais
INSERT INTO billing_plans (code, name, description, price_monthly, price_yearly, features, display_order, is_featured) VALUES
('starter', 'Starter', 'Para começar sem complicação', 97.00, 970.00,
 '{"fiscal_2026":true,"baas_pix":true,"whatsapp":true,"ai_copilot":false}', 1, false),
('pro', 'Pro', 'O mais escolhido pelos clientes', 197.00, 1970.00,
 '{"fiscal_2026":true,"baas_pix":true,"baas_boleto":true,"baas_split":true,"whatsapp":true,"ai_copilot":true,"ai_concierge":true,"roteirizador":true}', 2, true),
('business', 'Business', 'Para negócios em crescimento', 297.00, 2970.00,
 '{"fiscal_2026":true,"baas_pix":true,"baas_boleto":true,"baas_split":true,"whatsapp":true,"ai_copilot":true,"ai_concierge":true,"roteirizador":true,"multi_pdv":true,"api_access":true,"priority_support":true,"custom_reports":true}', 3, false)
ON CONFLICT (code) DO NOTHING;

-- ── BILLING: Assinaturas ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id               UUID UNIQUE NOT NULL REFERENCES tenants(id),
    plan_id                 UUID NOT NULL REFERENCES billing_plans(id),
    plan_code               TEXT NOT NULL,
    plan_name               TEXT NOT NULL,
    status                  TEXT NOT NULL DEFAULT 'trialing'
                            CHECK (status IN ('trialing','active','past_due','cancelled','expired')),
    trial_ends_at           TIMESTAMPTZ,
    current_period_start    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    billing_cycle           TEXT NOT NULL DEFAULT 'monthly',
    discount_percent        NUMERIC(5,2) DEFAULT 0,
    discount_reason         TEXT,
    price                   NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- Uso atual
    current_users           INT DEFAULT 1,
    current_transactions    INT DEFAULT 0,
    current_products        INT DEFAULT 0,
    current_invoices        INT DEFAULT 0,
    -- Limites herdados do plano (cache para performance)
    max_users               INT,
    max_transactions        INT,
    max_products            INT,
    max_invoices            INT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ── BILLING: Cupons ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS billing_coupons (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            TEXT UNIQUE NOT NULL,
    description     TEXT,
    discount_type   TEXT NOT NULL CHECK (discount_type IN ('percent','fixed')),
    discount_value  NUMERIC(10,2) NOT NULL,
    duration_months INT DEFAULT 1,
    valid_until     TIMESTAMPTZ,
    max_uses        INT,
    is_valid        BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_coupon_uses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    coupon_code TEXT NOT NULL,
    tenant_id   UUID NOT NULL,
    used_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (coupon_code, tenant_id)
);

-- Cupom SEBRAE — 60 dias grátis
INSERT INTO billing_coupons (code, description, discount_type, discount_value, duration_months)
VALUES ('SEBRAE2026', '60 dias grátis — parceria SEBRAE', 'percent', 100, 2)
ON CONFLICT (code) DO NOTHING;

-- ── TRIAL: Controle anti-abuso ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS trial_controls (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_number        TEXT NOT NULL,
    phone_hash          TEXT UNIQUE NOT NULL,  -- SHA256 para LGPD
    email               TEXT,
    cnpj                TEXT,
    verification_code   TEXT,
    code_expires_at     TIMESTAMPTZ,
    verified_at         TIMESTAMPTZ,
    device_hash         TEXT,
    ip_address          TEXT,
    tenant_id           UUID REFERENCES tenants(id),
    is_blocked          BOOLEAN DEFAULT FALSE,
    abuse_score         INT DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trial_phone_hash ON trial_controls(phone_hash);
CREATE INDEX IF NOT EXISTS idx_trial_device     ON trial_controls(device_hash, created_at);

-- ── JOURNEY: Eventos de rastreamento ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS journey_events (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID,
    user_id         TEXT,
    anonymous_id    TEXT,
    event_name      TEXT NOT NULL,
    event_category  TEXT,
    page_path       TEXT,
    page_title      TEXT,
    referrer        TEXT,
    properties      JSONB DEFAULT '{}',
    funnel_stage    TEXT,
    device_type     TEXT,
    browser         TEXT,
    os_name         TEXT,
    session_id      TEXT,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    time_on_page    INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_journey_tenant_date ON journey_events(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_journey_event_name  ON journey_events(event_name, occurred_at DESC);

-- ── JOURNEY: Métricas diárias do funil ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS journey_funnel_daily (
    date                    DATE NOT NULL,
    business_type           TEXT NOT NULL DEFAULT '',
    visits                  INT DEFAULT 0,
    signups_started         INT DEFAULT 0,
    signups_completed       INT DEFAULT 0,
    phone_verified          INT DEFAULT 0,
    onboarding_started      INT DEFAULT 0,
    onboarding_completed    INT DEFAULT 0,
    first_action            INT DEFAULT 0,
    trial_converted         INT DEFAULT 0,
    drop_signup             INT DEFAULT 0,
    drop_verification       INT DEFAULT 0,
    drop_onboarding         INT DEFAULT 0,
    drop_activation         INT DEFAULT 0,
    conversion_rate         NUMERIC(5,2) DEFAULT 0,
    PRIMARY KEY (date, business_type)
);

-- ── JOURNEY: Drop points ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS journey_drop_points (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL,
    user_id         TEXT,
    stage           TEXT NOT NULL,
    step_code       TEXT,
    days_stuck      INT DEFAULT 0,
    reminder_count  INT DEFAULT 0,
    resolved_at     TIMESTAMPTZ,
    resolution      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tenant_id, stage)
);

-- ── ONBOARDING: Passos por nicho ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS onboarding_steps (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_type   TEXT NOT NULL,
    step_code       TEXT NOT NULL,
    step_order      INT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    icon            TEXT,
    is_required     BOOLEAN DEFAULT TRUE,
    is_skippable    BOOLEAN DEFAULT FALSE,
    estimated_time  INT DEFAULT 60,
    action_type     TEXT,
    reward_text     TEXT,
    reward_days     INT DEFAULT 0,
    UNIQUE (business_type, step_code)
);

-- Seed passos de onboarding — Mecânica
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_required, is_skippable, estimated_time, reward_text, reward_days) VALUES
('mechanic','cadastrar_veiculo',1,'Cadastre seu primeiro veículo','Digite a placa e o sistema busca o histórico automaticamente','🚗',TRUE,FALSE,120,'Primeira OS liberada!',0),
('mechanic','criar_os',2,'Crie sua primeira OS','Registre a reclamação do cliente e adicione as peças','🔧',TRUE,FALSE,180,'5 dias extras no trial',5),
('mechanic','enviar_whatsapp',3,'Envie o orçamento por WhatsApp','O cliente aprova ou rejeita direto no celular — grátis','💬',TRUE,FALSE,60,'WhatsApp de aprovação ativo',0),
('mechanic','configurar_preco',4,'Configure seus preços de mão de obra','Defina o valor por hora e por tipo de serviço','💰',FALSE,TRUE,120,NULL,0)
ON CONFLICT (business_type, step_code) DO NOTHING;

-- Seed — Padaria
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_required, is_skippable, estimated_time, reward_text, reward_days) VALUES
('bakery','cadastrar_produto',1,'Cadastre seus produtos','Adicione pão francês, bolos e produtos com código PLU','🍞',TRUE,FALSE,180,'PDV liberado!',0),
('bakery','configurar_balanca',2,'Configure sua balança','Conecte a Toledo ou Elgin — o peso é lido automaticamente','⚖️',FALSE,TRUE,300,'5 dias extras no trial',5),
('bakery','primeira_venda',3,'Faça sua primeira venda','O PDV está pronto — registre uma venda de teste','🛒',TRUE,FALSE,120,NULL,0)
ON CONFLICT (business_type, step_code) DO NOTHING;

-- Seed — Estética
INSERT INTO onboarding_steps (business_type, step_code, step_order, title, description, icon, is_required, is_skippable, estimated_time, reward_text, reward_days) VALUES
('aesthetics','cadastrar_profissional',1,'Cadastre seus profissionais','Adicione cada profissional ou autônoma do salão','💇',TRUE,FALSE,120,'Agenda liberada!',0),
('aesthetics','cadastrar_servico',2,'Cadastre os serviços','Corte, escova, coloração — com preço e duração','✂️',TRUE,FALSE,180,'5 dias extras',5),
('aesthetics','primeiro_agendamento',3,'Faça o primeiro agendamento','O sistema trava conflito de horário automaticamente','📅',TRUE,FALSE,60,NULL,0)
ON CONFLICT (business_type, step_code) DO NOTHING;

-- ── ONBOARDING: Progresso por tenant ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS onboarding_progress (
    tenant_id       UUID PRIMARY KEY REFERENCES tenants(id),
    user_id         TEXT,
    business_type   TEXT,
    current_step    TEXT,
    total_steps     INT DEFAULT 0,
    completed_steps JSONB DEFAULT '[]',
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    last_activity   TIMESTAMPTZ DEFAULT NOW(),
    skipped         BOOLEAN DEFAULT FALSE
);

-- ── EXPENSES: Despesas com QR Code ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS expenses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    source          TEXT DEFAULT 'manual',
    nfe_key         TEXT,
    nfe_number      TEXT,
    nfe_series      TEXT,
    nfe_type        TEXT,
    nfe_url         TEXT,
    supplier_cnpj   TEXT,
    supplier_name   TEXT,
    total_amount    NUMERIC(15,2) NOT NULL DEFAULT 0,
    ibs_credit      NUMERIC(15,2) DEFAULT 0,
    cbs_credit      NUMERIC(15,2) DEFAULT 0,
    category        TEXT DEFAULT 'outros',
    subcategory     TEXT,
    paid            BOOLEAN DEFAULT FALSE,
    due_date        DATE,
    issue_date      DATE,
    registered_at   TIMESTAMPTZ DEFAULT NOW(),
    registered_by   TEXT,
    status          TEXT DEFAULT 'active',
    notes           TEXT
);

ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;
CREATE POLICY expenses_iso ON expenses USING (tenant_id = current_tenant_id());
CREATE INDEX IF NOT EXISTS idx_expenses_tenant_date ON expenses(tenant_id, registered_at DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_nfe_key ON expenses(nfe_key) WHERE nfe_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS expense_items (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL,
    expense_id  UUID NOT NULL REFERENCES expenses(id),
    item_order  INT,
    description TEXT NOT NULL,
    ncm         TEXT,
    quantity    NUMERIC(10,3),
    unit        TEXT,
    unit_price  NUMERIC(15,2),
    total_price NUMERIC(15,2)
);

ALTER TABLE expense_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY expense_items_iso ON expense_items USING (tenant_id = current_tenant_id());
