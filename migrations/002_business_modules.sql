-- =============================================================================
-- NEXO ONE ERP — Migração 002: Tabelas de Negócio (6 Nichos)
-- Todas as tabelas têm RLS habilitado.
-- =============================================================================

BEGIN;

SET search_path = nexo;

-- =============================================================================
-- MÓDULO: MECÂNICA
-- =============================================================================

CREATE TABLE mechanic_service_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    number          TEXT NOT NULL,
    vehicle_plate   TEXT NOT NULL,
    vehicle_km      INT,
    vehicle_model   TEXT,
    vehicle_year    INT,
    customer_id     UUID,
    customer_name   TEXT,
    customer_phone  TEXT,
    status          TEXT NOT NULL DEFAULT 'open'
                    CHECK (status IN ('open','diagnosed','await_approval','approved','rejected','in_progress','done','invoiced')),
    complaint       TEXT,
    diagnosis       TEXT,
    approval_token  TEXT UNIQUE,
    approval_url    TEXT,
    approved_at     TIMESTAMPTZ,
    total_parts     NUMERIC(15,2) DEFAULT 0,
    total_labor     NUMERIC(15,2) DEFAULT 0,
    total_amount    NUMERIC(15,2) DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mechanic_os_parts (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    os_id       UUID NOT NULL REFERENCES mechanic_service_orders(id) ON DELETE CASCADE,
    part_code   TEXT,
    description TEXT NOT NULL,
    quantity    NUMERIC(10,3) NOT NULL,
    unit_cost   NUMERIC(10,2) NOT NULL,
    unit_price  NUMERIC(10,2) NOT NULL,
    total_price NUMERIC(10,2) NOT NULL,
    ncm_code    TEXT
);

CREATE TABLE mechanic_os_labor (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id),
    os_id         UUID NOT NULL REFERENCES mechanic_service_orders(id) ON DELETE CASCADE,
    description   TEXT NOT NULL,
    hours         NUMERIC(6,2) NOT NULL,
    hourly_rate   NUMERIC(10,2) NOT NULL,
    total_price   NUMERIC(10,2) NOT NULL,
    technician_id UUID
);

ALTER TABLE mechanic_service_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE mechanic_os_parts ENABLE ROW LEVEL SECURITY;
ALTER TABLE mechanic_os_labor ENABLE ROW LEVEL SECURITY;

CREATE POLICY mec_os_isolation ON mechanic_service_orders USING (tenant_id = current_tenant_id());
CREATE POLICY mec_parts_isolation ON mechanic_os_parts USING (tenant_id = current_tenant_id());
CREATE POLICY mec_labor_isolation ON mechanic_os_labor USING (tenant_id = current_tenant_id());

CREATE INDEX idx_mec_os_tenant ON mechanic_service_orders(tenant_id);
CREATE INDEX idx_mec_os_plate ON mechanic_service_orders(tenant_id, vehicle_plate);
CREATE INDEX idx_mec_os_status ON mechanic_service_orders(tenant_id, status) WHERE status NOT IN ('done','invoiced');

-- =============================================================================
-- MÓDULO: PADARIA
-- =============================================================================

CREATE TABLE bakery_products (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    sku             TEXT NOT NULL,
    name            TEXT NOT NULL,
    sale_type       TEXT NOT NULL CHECK (sale_type IN ('weight','unit','combo')),
    unit_price      NUMERIC(10,2) NOT NULL,
    ncm_code        TEXT,
    is_basket_item  BOOLEAN NOT NULL DEFAULT FALSE,
    basket_category TEXT,
    scale_plu       TEXT,
    current_stock   NUMERIC(12,3) DEFAULT 0,
    min_stock       NUMERIC(12,3) DEFAULT 0,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, sku)
);

CREATE TABLE bakery_sales (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    number          TEXT NOT NULL,
    subtotal        NUMERIC(15,2) NOT NULL,
    discount        NUMERIC(15,2) DEFAULT 0,
    total_amount    NUMERIC(15,2) NOT NULL,
    total_tax       NUMERIC(15,2) DEFAULT 0,
    payment_method  TEXT NOT NULL CHECK (payment_method IN ('pix','cash','credit','debit','mixed')),
    operator_id     UUID,
    sold_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bakery_sale_items (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    sale_id     UUID NOT NULL REFERENCES bakery_sales(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES bakery_products(id),
    sale_type   TEXT NOT NULL,
    quantity    NUMERIC(10,3) NOT NULL,
    unit_price  NUMERIC(10,2) NOT NULL,
    discount    NUMERIC(10,2) DEFAULT 0,
    total_price NUMERIC(10,2) NOT NULL,
    ibs_amount  NUMERIC(10,2) DEFAULT 0,
    cbs_amount  NUMERIC(10,2) DEFAULT 0,
    is_basket   BOOLEAN DEFAULT FALSE
);

CREATE TABLE bakery_losses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    product_id  UUID NOT NULL REFERENCES bakery_products(id),
    quantity    NUMERIC(10,3) NOT NULL,
    unit        TEXT NOT NULL,
    reason      TEXT NOT NULL CHECK (reason IN ('expired','discarded','overprod','damaged')),
    cost_value  NUMERIC(10,2) NOT NULL,
    notes       TEXT,
    recorded_by UUID,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE bakery_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE bakery_sales ENABLE ROW LEVEL SECURITY;
ALTER TABLE bakery_sale_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE bakery_losses ENABLE ROW LEVEL SECURITY;

CREATE POLICY bak_products_iso ON bakery_products USING (tenant_id = current_tenant_id());
CREATE POLICY bak_sales_iso ON bakery_sales USING (tenant_id = current_tenant_id());
CREATE POLICY bak_items_iso ON bakery_sale_items USING (tenant_id = current_tenant_id());
CREATE POLICY bak_losses_iso ON bakery_losses USING (tenant_id = current_tenant_id());

CREATE INDEX idx_bak_products_tenant ON bakery_products(tenant_id);
CREATE INDEX idx_bak_sales_tenant ON bakery_sales(tenant_id);

-- =============================================================================
-- MÓDULO: ESTÉTICA
-- =============================================================================

CREATE TABLE aesthetics_professionals (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     UUID REFERENCES users(id),
    name        TEXT NOT NULL,
    phone       TEXT,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE aesthetics_services (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    name            TEXT NOT NULL,
    duration_min    INT NOT NULL,
    price           NUMERIC(10,2) NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE aesthetics_appointments (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id),
    professional_id   UUID NOT NULL REFERENCES aesthetics_professionals(id),
    customer_id       UUID,
    customer_name     TEXT NOT NULL,
    customer_phone    TEXT,
    service_id        UUID NOT NULL REFERENCES aesthetics_services(id),
    service_price     NUMERIC(10,2) NOT NULL,
    start_time        TIMESTAMPTZ NOT NULL,
    end_time          TIMESTAMPTZ NOT NULL,
    duration_min      INT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'scheduled'
                      CHECK (status IN ('scheduled','confirmed','in_progress','done','cancelled','no_show')),
    notes             TEXT,
    split_enabled     BOOLEAN DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT no_double_booking EXCLUDE USING gist (
        tenant_id WITH =,
        professional_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    ) WHERE (status NOT IN ('cancelled','no_show'))
);

CREATE TABLE aesthetics_split_rules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    appointment_id  UUID NOT NULL REFERENCES aesthetics_appointments(id),
    recipient_id    UUID NOT NULL,
    recipient_type  TEXT NOT NULL CHECK (recipient_type IN ('professional','salon')),
    percentage      NUMERIC(5,2),
    fixed_amount    NUMERIC(10,2),
    baas_account_id TEXT
);

ALTER TABLE aesthetics_professionals ENABLE ROW LEVEL SECURITY;
ALTER TABLE aesthetics_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE aesthetics_appointments ENABLE ROW LEVEL SECURITY;
ALTER TABLE aesthetics_split_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY aes_prof_iso ON aesthetics_professionals USING (tenant_id = current_tenant_id());
CREATE POLICY aes_srv_iso ON aesthetics_services USING (tenant_id = current_tenant_id());
CREATE POLICY aes_apt_iso ON aesthetics_appointments USING (tenant_id = current_tenant_id());
CREATE POLICY aes_split_iso ON aesthetics_split_rules USING (tenant_id = current_tenant_id());

CREATE INDEX idx_aes_prof_tenant ON aesthetics_professionals(tenant_id);
CREATE INDEX idx_aes_apt_tenant ON aesthetics_appointments(tenant_id);

-- =============================================================================
-- MÓDULO: CALÇADOS
-- =============================================================================

CREATE TABLE shoes_products (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    model_code  TEXT NOT NULL,
    model_name  TEXT NOT NULL,
    brand       TEXT,
    ncm_code    TEXT,
    colors      TEXT[] NOT NULL DEFAULT '{}',
    sizes       TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, model_code)
);

CREATE TABLE shoes_grid_cells (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    product_id  UUID NOT NULL REFERENCES shoes_products(id),
    sku         TEXT NOT NULL,
    color       TEXT NOT NULL,
    size        TEXT NOT NULL,
    stock       INT NOT NULL DEFAULT 0,
    price       NUMERIC(10,2) NOT NULL,
    cost_price  NUMERIC(10,2),
    barcode     TEXT,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (tenant_id, sku)
);

CREATE TABLE shoes_commission_rules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    seller_id       UUID NOT NULL REFERENCES users(id),
    base_rate       NUMERIC(5,4) NOT NULL,
    bonus_rate      NUMERIC(5,4) DEFAULT 0,
    monthly_target  NUMERIC(15,2) DEFAULT 0,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

ALTER TABLE shoes_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE shoes_grid_cells ENABLE ROW LEVEL SECURITY;
ALTER TABLE shoes_commission_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY shoes_prod_iso ON shoes_products USING (tenant_id = current_tenant_id());
CREATE POLICY shoes_grid_iso ON shoes_grid_cells USING (tenant_id = current_tenant_id());
CREATE POLICY shoes_comm_iso ON shoes_commission_rules USING (tenant_id = current_tenant_id());

CREATE INDEX idx_shoes_prod_tenant ON shoes_products(tenant_id);
CREATE INDEX idx_shoes_grid_tenant ON shoes_grid_cells(tenant_id);

-- =============================================================================
-- MÓDULO: INDÚSTRIA
-- =============================================================================

CREATE TABLE industry_bom (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    product_id  UUID NOT NULL,
    product_name TEXT NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    total_cost  NUMERIC(15,2) DEFAULT 0,
    valid_from  DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until DATE,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE industry_bom_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    bom_id          UUID NOT NULL REFERENCES industry_bom(id),
    component_id    UUID NOT NULL,
    component_name  TEXT NOT NULL,
    ncm_code        TEXT,
    quantity        NUMERIC(12,4) NOT NULL,
    unit            TEXT NOT NULL,
    scrap_factor    NUMERIC(5,4) DEFAULT 0,
    unit_cost       NUMERIC(10,4) NOT NULL,
    total_cost      NUMERIC(12,4) NOT NULL
);

CREATE TABLE industry_production_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    number          TEXT NOT NULL,
    product_id      UUID NOT NULL,
    product_name    TEXT NOT NULL,
    bom_id          UUID REFERENCES industry_bom(id),
    planned_qty     NUMERIC(12,3) NOT NULL,
    produced_qty    NUMERIC(12,3) DEFAULT 0,
    unit            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'planned'
                    CHECK (status IN ('planned','released','in_progress','done','cancelled')),
    planned_start   TIMESTAMPTZ,
    planned_end     TIMESTAMPTZ,
    actual_start    TIMESTAMPTZ,
    actual_end      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE industry_bom ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_bom_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_production_orders ENABLE ROW LEVEL SECURITY;

CREATE POLICY ind_bom_iso ON industry_bom USING (tenant_id = current_tenant_id());
CREATE POLICY ind_bom_items_iso ON industry_bom_items USING (tenant_id = current_tenant_id());
CREATE POLICY ind_po_iso ON industry_production_orders USING (tenant_id = current_tenant_id());

CREATE INDEX idx_ind_bom_tenant ON industry_bom(tenant_id);
CREATE INDEX idx_ind_po_tenant ON industry_production_orders(tenant_id);

COMMIT;
