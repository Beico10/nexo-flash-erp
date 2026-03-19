-- =============================================================================
-- NEXO FLASH ERP — Migração 004: CT-e, Agenda Estética e Índices de Performance
-- =============================================================================

-- =============================================================================
-- LOGÍSTICA: Tabela de CT-e emitidos
-- =============================================================================

CREATE TABLE logistics_ctes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    chave_cte       TEXT UNIQUE,                   -- 44 dígitos chave CT-e
    num_cte         TEXT NOT NULL,
    shipper_id      UUID,
    route_origin    TEXT NOT NULL,
    route_dest      TEXT NOT NULL,
    vehicle_type    TEXT NOT NULL,
    distance_km     NUMERIC(10,2),
    weight_kg       NUMERIC(10,3),
    gross_value     NUMERIC(15,2) NOT NULL,
    status          TEXT NOT NULL DEFAULT 'authorized'
                    CHECK (status IN ('authorized','cancelled','contingency')),
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at    TIMESTAMPTZ,
    xml_path        TEXT,                          -- caminho no object storage
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE logistics_ctes ENABLE ROW LEVEL SECURITY;
CREATE POLICY cte_iso ON logistics_ctes USING (tenant_id = current_tenant_id());

CREATE INDEX idx_cte_tenant_date ON logistics_ctes(tenant_id, issued_at DESC);
CREATE INDEX idx_cte_chave       ON logistics_ctes(chave_cte) WHERE chave_cte IS NOT NULL;

-- =============================================================================
-- ESTÉTICA: Índice para trava de conflito de agenda
-- (A constraint EXCLUDE já cria um índice GiST — este é um índice adicional)
-- =============================================================================

CREATE INDEX idx_apt_professional_date
    ON aesthetics_appointments(tenant_id, professional_id, start_time)
    WHERE status NOT IN ('cancelled', 'no_show');

-- =============================================================================
-- PERFORMANCE: Índices adicionais baseados em queries frequentes
-- =============================================================================

-- Busca de OS por placa (hot path da mecânica)
CREATE INDEX IF NOT EXISTS idx_mec_os_plate_date
    ON mechanic_service_orders(tenant_id, vehicle_plate, created_at DESC);

-- Busca de produtos por PLU na padaria (hit na balança)
CREATE INDEX IF NOT EXISTS idx_bak_product_plu
    ON bakery_products(tenant_id, scale_plu)
    WHERE scale_plu IS NOT NULL AND active = TRUE;

-- Conciliação PIX por data
CREATE INDEX IF NOT EXISTS idx_pix_tenant_date
    ON baas_pix_charges(tenant_id, created_at DESC);

-- =============================================================================
-- FUNÇÃO: audit_trigger()
-- Dispara log de auditoria em qualquer UPDATE ou DELETE em tabelas críticas.
-- =============================================================================

CREATE OR REPLACE FUNCTION audit_trigger_fn() RETURNS TRIGGER
    LANGUAGE plpgsql SECURITY DEFINER AS
$$
BEGIN
    INSERT INTO audit_log (tenant_id, action, table_name, record_id, old_data, new_data, occurred_at)
    VALUES (
        current_tenant_id(),
        TG_OP,
        TG_TABLE_NAME,
        COALESCE(OLD.id, NEW.id),
        CASE WHEN TG_OP IN ('UPDATE','DELETE') THEN to_jsonb(OLD) ELSE NULL END,
        CASE WHEN TG_OP IN ('INSERT','UPDATE') THEN to_jsonb(NEW) ELSE NULL END,
        NOW()
    );
    RETURN COALESCE(NEW, OLD);
END;
$$;

-- Aplicar trigger em tabelas financeiras críticas
CREATE TRIGGER audit_pix
    AFTER INSERT OR UPDATE ON baas_pix_charges
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER audit_boleto
    AFTER INSERT OR UPDATE ON baas_boletos
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_fn();

CREATE TRIGGER audit_ai_approval
    AFTER UPDATE ON ai_suggestions
    FOR EACH ROW WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE FUNCTION audit_trigger_fn();

-- =============================================================================
-- VIEW: dashboard_summary
-- Agregação para o dashboard principal — atualizada em tempo real.
-- =============================================================================

CREATE OR REPLACE VIEW dashboard_summary AS
SELECT
    t.id AS tenant_id,
    -- Mecânica
    COUNT(DISTINCT os.id) FILTER (WHERE os.status NOT IN ('done','invoiced'))     AS os_open,
    COUNT(DISTINCT os.id) FILTER (WHERE os.status = 'await_approval')            AS os_await_approval,
    SUM(os.total_amount)  FILTER (WHERE os.created_at::date = CURRENT_DATE)      AS os_revenue_today,
    -- Padaria
    COUNT(DISTINCT bs.id) FILTER (WHERE bs.sold_at::date = CURRENT_DATE)         AS bakery_sales_today,
    SUM(bs.total_amount)  FILTER (WHERE bs.sold_at::date = CURRENT_DATE)         AS bakery_revenue_today,
    -- IA
    COUNT(DISTINCT ai.id) FILTER (WHERE ai.status = 'pending')                   AS ai_pending,
    -- Pagamentos
    SUM(pix.amount) FILTER (WHERE pix.status = 'paid' AND pix.paid_at::date = CURRENT_DATE) AS pix_received_today
FROM tenants t
LEFT JOIN mechanic_service_orders os ON os.tenant_id = t.id
LEFT JOIN bakery_sales bs             ON bs.tenant_id = t.id
LEFT JOIN ai_suggestions ai           ON ai.tenant_id = t.id AND ai.expires_at > NOW()
LEFT JOIN baas_pix_charges pix        ON pix.tenant_id = t.id
WHERE t.id = current_tenant_id()
GROUP BY t.id;

COMMENT ON VIEW dashboard_summary IS
    'Agregação em tempo real para o dashboard. '
    'Usada pelo endpoint GET /api/v1/dashboard/summary com RLS ativo.';
