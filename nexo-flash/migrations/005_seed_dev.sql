-- =============================================================================
-- NEXO FLASH ERP — Migration 005: Seed de Dados Iniciais
-- Cria o primeiro tenant de teste e usuário admin.
-- Execute apenas em desenvolvimento / primeira instalação.
-- =============================================================================

-- Função para criar tenant + usuário admin com senha hashada
-- Uso: SELECT seed_tenant('mecanica-joao', 'mechanic', 'Mecânica do João', 'admin@mecanica.com', 'senha123');
CREATE OR REPLACE FUNCTION seed_tenant(
    p_slug         TEXT,
    p_type         TEXT,
    p_name         TEXT,
    p_admin_email  TEXT,
    p_admin_pass   TEXT  -- bcrypt hash gerado externamente
) RETURNS TEXT
LANGUAGE plpgsql AS $$
DECLARE
    v_tenant_id UUID;
    v_user_id   UUID;
BEGIN
    -- Criar tenant
    INSERT INTO tenants (slug, business_type, name, plan)
    VALUES (p_slug, p_type, p_name, 'pro')
    ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO v_tenant_id;

    -- Configurar app.tenant_id para RLS
    PERFORM set_config('app.tenant_id', v_tenant_id::text, TRUE);

    -- Criar usuário owner
    INSERT INTO users (tenant_id, email, name, role, password_hash)
    VALUES (v_tenant_id, p_admin_email, 'Administrador', 'owner', p_admin_pass)
    ON CONFLICT (tenant_id, email) DO NOTHING
    RETURNING id INTO v_user_id;

    RETURN format('Tenant: %s | User: %s', v_tenant_id, v_user_id);
END;
$$;

-- =============================================================================
-- Dados de exemplo para desenvolvimento
-- Senhas geradas com bcrypt cost=12
-- Senha padrão para todos: "nexo@2026"
-- =============================================================================

DO $$
DECLARE
    -- bcrypt hash de "nexo@2026" — troque em produção!
    v_hash TEXT := '$2a$12$LpVzGwQwT6k0dRqhsMCn5.tHdV5k3N0lF6LM3sU/PjRzFNnS6TkMC';
    v_result TEXT;
BEGIN
    v_result := seed_tenant('mecanica-demo', 'mechanic',  'Mecânica Demo',  'admin@mecanica.demo',  v_hash);
    RAISE NOTICE 'Mechanic:  %', v_result;

    v_result := seed_tenant('padaria-demo',  'bakery',    'Padaria Demo',   'admin@padaria.demo',   v_hash);
    RAISE NOTICE 'Bakery:    %', v_result;

    v_result := seed_tenant('industria-demo','industry',  'Indústria Demo', 'admin@industria.demo', v_hash);
    RAISE NOTICE 'Industry:  %', v_result;

    v_result := seed_tenant('logistica-demo','logistics', 'Logística Demo', 'admin@logistica.demo', v_hash);
    RAISE NOTICE 'Logistics: %', v_result;

    v_result := seed_tenant('estetica-demo', 'aesthetics','Estética Demo',  'admin@estetica.demo',  v_hash);
    RAISE NOTICE 'Aesthetics:%', v_result;

    v_result := seed_tenant('calcados-demo', 'shoes',     'Calçados Demo',  'admin@calcados.demo',  v_hash);
    RAISE NOTICE 'Shoes:     %', v_result;
END $$;

-- Inserir alíquotas IBS/CBS 2026 básicas para desenvolvimento
INSERT INTO fiscal_ncm_rates (ncm_code, ncm_description, ibs_rate, cbs_rate, selective_rate, basket_reduced, basket_type, transition_year, transition_factor, valid_from, source_lei)
VALUES
    -- Produtos industriais gerais
    ('84715010', 'Computadores e notebooks',         0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('87032310', 'Automóveis até 1000cc',             0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    -- Cesta Básica — alíquota zero (LC 194/2022 + LC 214/2025)
    ('10063021', 'Arroz beneficiado',                0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('07133310', 'Feijão preto',                     0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('04011000', 'Leite integral',                   0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    ('19052000', 'Pão francês',                      0.0925, 0.0375, 0, TRUE,  'zero', 2026, 0.10, '2026-01-01', 'LC 194/2022'),
    -- Peças automotivas (mecânica)
    ('87083000', 'Freios e partes',                  0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    ('87088010', 'Amortecedores',                    0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    -- Bebidas (Imposto Seletivo)
    ('22030000', 'Cerveja de malte',                 0.0925, 0.0375, 0.10, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025 IS'),
    ('22084000', 'Rum e aguardente',                 0.0925, 0.0375, 0.20, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025 IS'),
    -- Calçados
    ('64041900', 'Calçados com sola de borracha',    0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025'),
    -- Serviços de estética (futuro ISS→IBS)
    ('96021000', 'Serviços de cabeleireiro',         0.0925, 0.0375, 0, FALSE, NULL, 2026, 0.10, '2026-01-01', 'LC 214/2025')
ON CONFLICT (ncm_code, valid_from) DO NOTHING;

COMMENT ON FUNCTION seed_tenant IS
    'Cria tenant + usuário admin para desenvolvimento. '
    'NÃO execute em produção com senha padrão.';
