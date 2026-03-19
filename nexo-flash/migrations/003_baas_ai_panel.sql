-- =============================================================================
-- NEXO FLASH ERP — Migração 003: Painel de Aprovação IA + Pagamentos BaaS
-- =============================================================================

-- =============================================================================
-- MÓDULO: BAAS — Cobranças PIX e Boleto
-- =============================================================================

CREATE TABLE baas_pix_charges (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    tx_id           TEXT NOT NULL UNIQUE,          -- ID BACEN (max 35 chars)
    amount          NUMERIC(15,2) NOT NULL,
    description     TEXT,
    payer_name      TEXT,
    payer_document  TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','paid','expired','cancelled','refunded')),
    qr_code_text    TEXT,                          -- payload EMV copia-e-cola
    qr_code_image   TEXT,                          -- base64 PNG
    paid_at         TIMESTAMPTZ,
    split_data      JSONB,                         -- destinatários do split
    reference_id    TEXT,                          -- ID da OS / venda / agendamento
    reference_type  TEXT,                          -- 'mechanic_os'|'bakery_sale'|'aesthetics_apt'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE baas_boletos (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    our_number      TEXT NOT NULL,
    amount          NUMERIC(15,2) NOT NULL,
    due_date        DATE NOT NULL,
    description     TEXT,
    payer_name      TEXT NOT NULL,
    payer_document  TEXT NOT NULL,
    payer_address   TEXT,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','paid','expired','cancelled')),
    bar_code        TEXT,
    digitable_line  TEXT,
    qr_code_text    TEXT,                          -- PIX embutido
    pdf_url         TEXT,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Conciliação automática: log de webhooks recebidos
CREATE TABLE baas_webhook_events (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID,
    event_type      TEXT NOT NULL,
    charge_id       TEXT,
    tx_id           TEXT,
    amount          NUMERIC(15,2),
    paid_at         TIMESTAMPTZ,
    raw_payload     JSONB NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE baas_pix_charges ENABLE ROW LEVEL SECURITY;
ALTER TABLE baas_boletos ENABLE ROW LEVEL SECURITY;

CREATE POLICY pix_iso ON baas_pix_charges USING (tenant_id = current_tenant_id());
CREATE POLICY boleto_iso ON baas_boletos USING (tenant_id = current_tenant_id());

-- Índice para busca rápida por tx_id (hot path de conciliação)
CREATE INDEX idx_pix_txid ON baas_pix_charges(tx_id);
CREATE INDEX idx_pix_status ON baas_pix_charges(tenant_id, status) WHERE status = 'pending';
CREATE INDEX idx_boleto_status ON baas_boletos(tenant_id, status) WHERE status = 'pending';

-- =============================================================================
-- PAINEL DE APROVAÇÃO IA — View para o frontend
-- Consolida sugestões pendentes com contexto legível
-- =============================================================================

CREATE VIEW ai_pending_panel AS
SELECT
    s.id,
    s.tenant_id,
    s.suggestion_type,
    s.target_table,
    s.target_id,
    s.suggested_data,
    s.reason,
    s.confidence,
    s.created_by_ai,
    s.created_at,
    s.expires_at,
    -- Quantos dias restam para expirar
    EXTRACT(DAY FROM (s.expires_at - NOW())) AS days_until_expiry,
    -- Nível de urgência baseado na confiança e tipo
    CASE
        WHEN s.confidence >= 0.9 AND s.suggestion_type = 'ncm_correction' THEN 'high'
        WHEN s.confidence >= 0.8 THEN 'medium'
        ELSE 'low'
    END AS urgency,
    -- Descrição amigável do tipo
    CASE s.suggestion_type
        WHEN 'missing_labor_cost'  THEN 'Mão de obra faltante na OS'
        WHEN 'ncm_correction'      THEN 'Correção de NCM'
        WHEN 'price_anomaly'       THEN 'Anomalia de preço'
        WHEN 'stock_low_reorder'   THEN 'Reposição de estoque'
        WHEN 'route_optimization'  THEN 'Otimização de rota'
        WHEN 'onboard_field'       THEN 'Produto detectado em NF-e'
        ELSE s.suggestion_type
    END AS type_label
FROM ai_suggestions s
WHERE s.status = 'pending'
  AND s.expires_at > NOW();

-- =============================================================================
-- MÓDULO: LOGÍSTICA — Tabela de Contratos no banco
-- =============================================================================

CREATE TABLE logistics_contracts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    contract_name   TEXT NOT NULL,
    shipper_id      UUID,                          -- NULL = tabela geral
    vehicle_type    TEXT NOT NULL DEFAULT 'general'
                    CHECK (vehicle_type IN ('general','vuc','truck','carreta','van')),
    price_per_km    NUMERIC(10,4),
    price_per_kg    NUMERIC(10,4),
    minimum_charge  NUMERIC(10,2),
    toll_policy     TEXT DEFAULT 'included'
                    CHECK (toll_policy IN ('included','excluded','split')),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from      DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until     DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE logistics_contracts ENABLE ROW LEVEL SECURITY;
CREATE POLICY logi_contract_iso ON logistics_contracts
    USING (tenant_id = current_tenant_id());

-- Índice para resolução de contrato (hot path do cálculo de frete)
-- Prioriza shipper_id específico sobre NULL (tabela geral)
CREATE INDEX idx_logi_contract_lookup
    ON logistics_contracts(tenant_id, shipper_id NULLS LAST, vehicle_type)
    WHERE active = TRUE;

-- =============================================================================
-- FUNÇÃO: resolve_contract()
-- Retorna o contrato mais específico para um par (embarcador, veículo).
-- Hierarquia: shipper específico > tabela geral (shipper_id IS NULL)
-- =============================================================================

CREATE OR REPLACE FUNCTION resolve_contract(
    p_tenant_id UUID,
    p_shipper_id UUID,
    p_vehicle_type TEXT
) RETURNS UUID
LANGUAGE sql STABLE SECURITY DEFINER AS
$$
    SELECT id FROM logistics_contracts
    WHERE tenant_id = p_tenant_id
      AND vehicle_type = p_vehicle_type
      AND active = TRUE
      AND (valid_until IS NULL OR valid_until >= CURRENT_DATE)
    ORDER BY
        -- Contrato específico do embarcador tem prioridade máxima
        CASE WHEN shipper_id = p_shipper_id THEN 0 ELSE 1 END,
        -- Dentro do mesmo nível, o mais recente vence
        valid_from DESC
    LIMIT 1;
$$;

COMMENT ON FUNCTION resolve_contract IS
    'Implementa a hierarquia de contratos do Briefing Mestre §4: '
    'contrato específico do embarcador tem prioridade sobre a tabela geral.';
