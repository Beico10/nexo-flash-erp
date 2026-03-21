-- =============================================================================
-- NEXO ONE ERP — Migration 012: Estoque (Inventory)
-- Sistema camaleão — uma tabela para todos os nichos
-- =============================================================================

CREATE TABLE IF NOT EXISTS inventory_products (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    business_type   TEXT NOT NULL,

    -- Dados básicos
    code            TEXT,
    barcode         TEXT,
    name            TEXT NOT NULL,
    description     TEXT,
    category        TEXT,
    unit            TEXT NOT NULL DEFAULT 'un',

    -- Estoque
    quantity        NUMERIC(15,3) NOT NULL DEFAULT 0,
    min_quantity    NUMERIC(15,3) DEFAULT 0,
    max_quantity    NUMERIC(15,3) DEFAULT 0,
    location        TEXT,

    -- Financeiro
    cost_price      NUMERIC(15,4) DEFAULT 0,  -- Custo Médio Ponderado
    sale_price      NUMERIC(15,2) DEFAULT 0,
    total_value     NUMERIC(15,2) GENERATED ALWAYS AS (quantity * cost_price) STORED,

    -- Fiscal
    ncm             TEXT,
    cfop            TEXT,

    -- Campos específicos do nicho (JSONB — camaleão)
    extra           JSONB DEFAULT '{}',

    -- Controle
    is_active       BOOLEAN DEFAULT TRUE,
    alert_sent      BOOLEAN DEFAULT FALSE,
    last_movement   TIMESTAMPTZ,
    created_by      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE (tenant_id, code)
);

-- RLS
ALTER TABLE inventory_products ENABLE ROW LEVEL SECURITY;
CREATE POLICY inventory_products_iso ON inventory_products
    USING (tenant_id = current_tenant_id());

-- Índices
CREATE INDEX IF NOT EXISTS idx_inv_tenant_category ON inventory_products(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_inv_tenant_barcode  ON inventory_products(tenant_id, barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_inv_low_stock       ON inventory_products(tenant_id, quantity, min_quantity) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_inv_ncm             ON inventory_products(ncm) WHERE ncm IS NOT NULL;

-- Trigger updated_at
CREATE TRIGGER trg_inventory_updated_at
    BEFORE UPDATE ON inventory_products
    FOR EACH ROW EXECUTE FUNCTION update_payables_updated_at();

-- =============================================================================
-- MOVIMENTAÇÕES DE ESTOQUE
-- =============================================================================

CREATE TABLE IF NOT EXISTS inventory_movements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    product_id      UUID NOT NULL REFERENCES inventory_products(id),
    product_name    TEXT NOT NULL,

    type            TEXT NOT NULL
                    CHECK (type IN ('in','out','adjust','loss','transfer')),
    quantity        NUMERIC(15,3) NOT NULL,
    unit_cost       NUMERIC(15,4) DEFAULT 0,
    total_cost      NUMERIC(15,2) DEFAULT 0,

    -- Saldo antes/depois
    previous_stock  NUMERIC(15,3) NOT NULL,
    new_stock       NUMERIC(15,3) NOT NULL,
    previous_cmp    NUMERIC(15,4) DEFAULT 0,
    new_cmp         NUMERIC(15,4) DEFAULT 0,

    -- Referência
    reference_id    TEXT,
    reference_type  TEXT,  -- os, nfe, sale, adjustment
    nfe_key         TEXT,

    notes           TEXT,
    created_by      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- RLS
ALTER TABLE inventory_movements ENABLE ROW LEVEL SECURITY;
CREATE POLICY inventory_movements_iso ON inventory_movements
    USING (tenant_id = current_tenant_id());

-- Índices
CREATE INDEX IF NOT EXISTS idx_inv_mov_product   ON inventory_movements(tenant_id, product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_inv_mov_type      ON inventory_movements(tenant_id, type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_inv_mov_reference ON inventory_movements(reference_id) WHERE reference_id IS NOT NULL;

-- =============================================================================
-- SEED DEMO — Mecânica
-- =============================================================================
INSERT INTO inventory_products
    (tenant_id, business_type, code, name, category, unit, quantity, min_quantity, cost_price, sale_price, location, extra)
SELECT
    t.id, 'mechanic',
    unnest(ARRAY['FLT-001','OLE-001','PAS-001','VEL-001']),
    unnest(ARRAY['Filtro de Óleo Bosch','Óleo Motor 5W30 1L','Pastilha de Freio Dianteira','Vela de Ignição NGK']),
    unnest(ARRAY['filtros','fluidos','freios','elétrica']),
    unnest(ARRAY['un','l','un','un']),
    unnest(ARRAY[8, 2, 12, 0]::NUMERIC[]),
    unnest(ARRAY[5, 10, 4, 8]::NUMERIC[]),
    unnest(ARRAY[18.50, 28.00, 45.00, 12.00]::NUMERIC[]),
    unnest(ARRAY[35.00, 48.00, 89.00, 22.00]::NUMERIC[]),
    unnest(ARRAY['A1-P2','B2-P1','C1-P3','A2-P1']),
    unnest(ARRAY[
        '{"manufacturer_code":"0986AF0030","application":"Fiat/VW 1.0-1.6","brand":"Bosch"}',
        '{"brand":"Mobil 1"}',
        '{"manufacturer_code":"TRW-GDB1234","application":"Gol/Palio/Uno","brand":"TRW"}',
        '{"manufacturer_code":"BKR5EGP","application":"Universal","brand":"NGK"}'
    ]::JSONB[])
FROM tenants t WHERE t.slug = 'demo'
ON CONFLICT (tenant_id, code) DO NOTHING;
