-- =============================================================================
-- NEXO ONE ERP — Migration 005: Seed de Dados Iniciais
-- Execute apenas em desenvolvimento / primeira instalação.
-- =============================================================================

BEGIN;

SET search_path = nexo;

-- =============================================================================
-- Dados de exemplo para desenvolvimento
-- Senha padrão para todos: "nexo@2026"
-- bcrypt hash gerado com cost=12
-- =============================================================================

DO $$
DECLARE
    v_hash TEXT := '$2a$12$LpVzGwQwT6k0dRqhsMCn5.tHdV5k3N0lF6LM3sU/PjRzFNnS6TkMC';
    v_tenant_id UUID;
BEGIN
    -- ═══════════════════════════════════════════════════════════════════════
    -- MECÂNICA DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('mecanica-demo', 'Mecânica Demo', 'mechanic', 'pro', '12345678000101', 'Mecânica Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@mecanica.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Mecânica Demo: %', v_tenant_id;

    -- ═══════════════════════════════════════════════════════════════════════
    -- PADARIA DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('padaria-demo', 'Padaria Demo', 'bakery', 'pro', '12345678000102', 'Padaria Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@padaria.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Padaria Demo: %', v_tenant_id;

    -- ═══════════════════════════════════════════════════════════════════════
    -- INDÚSTRIA DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('industria-demo', 'Indústria Demo', 'industry', 'pro', '12345678000103', 'Indústria Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@industria.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Indústria Demo: %', v_tenant_id;

    -- ═══════════════════════════════════════════════════════════════════════
    -- LOGÍSTICA DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('logistica-demo', 'Logística Demo', 'logistics', 'pro', '12345678000104', 'Logística Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@logistica.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Logística Demo: %', v_tenant_id;

    -- ═══════════════════════════════════════════════════════════════════════
    -- ESTÉTICA DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('estetica-demo', 'Estética Demo', 'aesthetics', 'pro', '12345678000105', 'Estética Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@estetica.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Estética Demo: %', v_tenant_id;

    -- ═══════════════════════════════════════════════════════════════════════
    -- CALÇADOS DEMO
    -- ═══════════════════════════════════════════════════════════════════════
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES ('calcados-demo', 'Calçados Demo', 'shoes', 'pro', '12345678000106', 'Calçados Demo Ltda')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    PERFORM set_config('nexo.current_tenant_id', v_tenant_id::text, TRUE);

    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, 'admin@calcados.demo', 'Administrador', 'owner', v_hash)
    ON CONFLICT (tenant_id, email) DO NOTHING;

    RAISE NOTICE 'Calçados Demo: %', v_tenant_id;

END $$;

-- =============================================================================
-- Alíquotas IBS/CBS 2026 básicas
-- =============================================================================

INSERT INTO fiscal_ncm_rates (ncm_code, ncm_description, ibs_rate, cbs_rate, selective_rate, basket_reduced, basket_type, transition_year, transition_factor, valid_from, source_lei)
VALUES
    ('84715010', 'Computadores e notebooks',         0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('87032310', 'Automóveis até 1000cc',             0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('10063021', 'Arroz beneficiado',                 0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('07133310', 'Feijão preto',                      0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('04011000', 'Leite integral',                    0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('19052000', 'Pão francês',                       0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('87083000', 'Freios e partes',                   0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('87088010', 'Amortecedores',                     0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('22030000', 'Cerveja de malte',                  0.0925, 0.0375, 0.10, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025 IS'),
    ('22084000', 'Rum e aguardente',                  0.0925, 0.0375, 0.20, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025 IS'),
    ('64041900', 'Calçados com sola de borracha',     0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('96021000', 'Serviços de cabeleireiro',          0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025')
ON CONFLICT (ncm_code, valid_from) DO NOTHING;

COMMIT;
