-- =============================================================================
-- NEXO ONE ERP — Migration 011: Contas a Receber + View Fluxo de Caixa
-- =============================================================================

CREATE TABLE IF NOT EXISTS receivables (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    description         TEXT NOT NULL,
    customer_name       TEXT,
    customer_phone      TEXT,
    customer_document   TEXT,
    category            TEXT DEFAULT 'servico'
                        CHECK (category IN ('servico','produto','mensalidade','contrato','outros')),
    amount              NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    amount_received     NUMERIC(15,2) DEFAULT 0,
    due_date            DATE NOT NULL,
    received_at         DATE,
    payment_method      TEXT,
    installment         INT DEFAULT 1,
    total_installments  INT DEFAULT 1,
    parent_id           UUID REFERENCES receivables(id),
    recurrence          TEXT DEFAULT 'none'
                        CHECK (recurrence IN ('none','weekly','monthly','yearly')),
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','received','overdue','cancelled','partial')),
    notes               TEXT,
    reference_id        TEXT,
    reference_type      TEXT,
    alert_sent          BOOLEAN DEFAULT FALSE,
    alert_count         INT DEFAULT 0,
    created_by          TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- RLS
ALTER TABLE receivables ENABLE ROW LEVEL SECURITY;
CREATE POLICY receivables_iso ON receivables
    USING (tenant_id = current_tenant_id());

-- Índices
CREATE INDEX IF NOT EXISTS idx_receivables_tenant_due    ON receivables(tenant_id, due_date);
CREATE INDEX IF NOT EXISTS idx_receivables_tenant_status ON receivables(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_receivables_customer      ON receivables(tenant_id, customer_name);
CREATE INDEX IF NOT EXISTS idx_receivables_overdue       ON receivables(tenant_id, due_date) WHERE status = 'pending';

-- Trigger updated_at
CREATE OR REPLACE FUNCTION update_receivables_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_receivables_updated_at
    BEFORE UPDATE ON receivables
    FOR EACH ROW EXECUTE FUNCTION update_receivables_updated_at();

-- =============================================================================
-- VIEW: Fluxo de Caixa consolidado (Pagar + Receber)
-- =============================================================================
CREATE OR REPLACE VIEW vw_cashflow AS
SELECT
    tenant_id,
    due_date       AS date,
    description,
    supplier_name  AS party,
    amount,
    'payable'      AS flow_type,  -- saída
    category,
    status
FROM payables
WHERE status NOT IN ('cancelled')

UNION ALL

SELECT
    tenant_id,
    due_date,
    description,
    customer_name,
    amount,
    'receivable',  -- entrada
    category,
    status
FROM receivables
WHERE status NOT IN ('cancelled')

ORDER BY date, flow_type;

-- =============================================================================
-- SEED DEMO
-- =============================================================================
INSERT INTO receivables (tenant_id, description, customer_name, customer_phone, category, amount, due_date, recurrence, status)
SELECT
    t.id,
    unnest(ARRAY[
        'OS #1042 — Troca de embreagem',
        'Mensalidade Frotas — Transportadora ABC',
        'OS #1039 — Revisão completa — VW Gol'
    ]),
    unnest(ARRAY['João da Silva', 'Transportadora ABC', 'Maria Oliveira']),
    unnest(ARRAY['5511991234567', '5511955551234', '5511987654321']),
    unnest(ARRAY['servico','mensalidade','servico']),
    unnest(ARRAY[850.00, 1200.00, 420.00]),
    unnest(ARRAY[
        CURRENT_DATE + 5,
        CURRENT_DATE + 10,
        CURRENT_DATE - 4
    ]::DATE[]),
    unnest(ARRAY['none','monthly','none']),
    unnest(ARRAY['pending','pending','overdue'])
FROM tenants t WHERE t.slug = 'demo'
ON CONFLICT DO NOTHING;
