-- =============================================================================
-- Nexo One ERP — Migration 013: Usuário Master / Dono da Plataforma
-- Cria o tenant "nexo-plataforma" e o usuário superadmin do sistema.
-- Execute uma única vez em produção.
-- =============================================================================

BEGIN;

SET search_path = nexo;

DO $$
DECLARE
    v_tenant_id UUID;
    v_user_id   UUID;
    v_hash      TEXT;
BEGIN
    -- Gera hash bcrypt cost=12 usando pgcrypto
    v_hash := crypt('Antonio#6325', gen_salt('bf', 12));

    -- 1. Cria (ou atualiza) o tenant da plataforma
    INSERT INTO tenants (slug, name, business_type, plan, cnpj, razao_social)
    VALUES (
        'nexo-plataforma',
        'Gestão Para Todos — Plataforma',
        'platform',
        'enterprise',
        NULL,
        'Gestão Para Todos Tecnologia'
    )
    ON CONFLICT (slug) DO UPDATE
        SET name       = EXCLUDED.name,
            plan       = EXCLUDED.plan,
            updated_at = NOW()
    RETURNING id INTO v_tenant_id;

    -- 2. Cria (ou atualiza) o usuário master
    INSERT INTO users (tenant_id, email, name, full_name, role, password_hash)
    VALUES (
        v_tenant_id,
        'antinybeico10@gmail.com',
        'Antonio',
        'Antonio — Dono da Plataforma',
        'owner',
        v_hash
    )
    ON CONFLICT (tenant_id, email) DO UPDATE
        SET password_hash = EXCLUDED.password_hash,
            role          = 'owner',
            is_active     = TRUE
    RETURNING id INTO v_user_id;

    RAISE NOTICE '✅ Usuário master criado com sucesso!';
    RAISE NOTICE '   tenant_id : %', v_tenant_id;
    RAISE NOTICE '   user_id   : %', v_user_id;
    RAISE NOTICE '   email     : antinybeico10@gmail.com';
    RAISE NOTICE '   slug      : nexo-plataforma';
    RAISE NOTICE '   role      : owner (acesso total)';
END $$;

COMMIT;
