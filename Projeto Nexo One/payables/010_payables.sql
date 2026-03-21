-- =============================================================================
-- NEXO ONE ERP — Migration 010: Contas a Pagar
-- =============================================================================

CREATE TABLE IF NOT EXISTS payables (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    description         TEXT NOT NULL,
    supplier_name       TEXT,
    supplier_cnpj       TEXT,
    category            TEXT DEFAULT 'outros'
                        CHECK (category IN ('aluguel','fornecedor','imposto','folha','servico','outros')),
    amount              NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    amount_paid         NUMERIC(15,2) DEFAULT 0,
    due_date            DATE NOT NULL,
    paid_at             DATE,
    payment_method      TEXT CHECK (payment_method IN ('pix','boleto','cartao','dinheiro','transferencia') OR payment_method IS NULL),
    bank_account        TEXT,
    installment         INT DEFAULT 1,
    total_installments  INT DEFAULT 1,
    parent_id           UUID REFERENCES payables(id),
    recurrence          TEXT DEFAULT 'none'
                        CHECK (recurrence IN ('none','weekly','monthly','yearly')),
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','paid','overdue','cancelled')),
    notes               TEXT,
    nfe_key             TEXT,
    cost_center         TEXT,
    alert_sent          BOOLEAN DEFAULT FALSE,
    created_by          TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- RLS — isolamento total por tenant
ALTER TABLE payables ENABLE ROW LEVEL SECURITY;
CREATE POLICY payables_iso ON payables
    USING (tenant_id = current_tenant_id());

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_payables_tenant_due     ON payables(tenant_id, due_date);
CREATE INDEX IF NOT EXISTS idx_payables_tenant_status  ON payables(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_payables_tenant_category ON payables(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_payables_overdue         ON payables(tenant_id, due_date) WHERE status = 'pending';

-- Trigger updated_at automático
CREATE OR REPLACE FUNCTION update_payables_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_payables_updated_at
    BEFORE UPDATE ON payables
    FOR EACH ROW EXECUTE FUNCTION update_payables_updated_at();

-- Seed demo
INSERT INTO payables (tenant_id, description, supplier_name, category, amount, due_date, recurrence, status)
SELECT
    t.id,
    unnest(ARRAY[
        'Aluguel do galpão — Março/2026',
        'SIMPLES Nacional — Fevereiro/2026',
        'Conta de Energia — Fevereiro/2026'
    ]),
    unnest(ARRAY[
        'Imobiliária Central',
        'Receita Federal',
        'CPFL Energia'
    ]),
    unnest(ARRAY['aluguel','imposto','servico']),
    unnest(ARRAY[3500.00, 780.00, 450.00]),
    unnest(ARRAY[
        CURRENT_DATE + 3,
        CURRENT_DATE + 1,
        CURRENT_DATE + 10
    ]::DATE[]),
    unnest(ARRAY['monthly','monthly','monthly']),
    'pending'
FROM tenants t WHERE t.slug = 'demo'
ON CONFLICT DO NOTHING;
