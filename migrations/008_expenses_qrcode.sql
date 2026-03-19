-- ============================================================
-- NEXO ONE ERP — Migration 008: Módulo de Despesas (QR Code NF-e)
-- ============================================================
-- Leitor de QR Code de NFC-e/NF-e para registrar despesas
-- MEI aponta câmera → Sistema registra gasto → Abate no imposto
-- ============================================================

BEGIN;

SET search_path = nexo;

-- ═══════════════════════════════════════════════════════════
-- TABELA: expenses (Despesas do negócio)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE expenses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Origem da despesa
    source          VARCHAR(30) NOT NULL DEFAULT 'manual'
                    CHECK (source IN ('qrcode', 'xml_upload', 'manual', 'recurring')),
    
    -- Dados da NF-e/NFC-e (se origem for qrcode/xml)
    nfe_key         VARCHAR(44) UNIQUE,           -- Chave de acesso 44 dígitos
    nfe_number      VARCHAR(20),
    nfe_series      VARCHAR(5),
    nfe_type        VARCHAR(10),                  -- 'nfe', 'nfce', 'cte', 'sat'
    nfe_url         TEXT,                         -- URL do QR Code
    nfe_xml         TEXT,                         -- XML completo (se disponível)
    
    -- Fornecedor
    supplier_cnpj   VARCHAR(14),
    supplier_name   VARCHAR(255) NOT NULL,
    supplier_ie     VARCHAR(20),
    
    -- Valores
    total_products  NUMERIC(15,2) NOT NULL,       -- Total dos produtos
    total_discount  NUMERIC(15,2) DEFAULT 0,
    total_shipping  NUMERIC(15,2) DEFAULT 0,
    total_other     NUMERIC(15,2) DEFAULT 0,
    total_amount    NUMERIC(15,2) NOT NULL,       -- Valor total da nota
    
    -- Impostos pagos (para crédito)
    icms_amount     NUMERIC(15,2) DEFAULT 0,
    ipi_amount      NUMERIC(15,2) DEFAULT 0,
    pis_amount      NUMERIC(15,2) DEFAULT 0,
    cofins_amount   NUMERIC(15,2) DEFAULT 0,
    ibs_credit      NUMERIC(15,2) DEFAULT 0,      -- Crédito IBS (2026+)
    cbs_credit      NUMERIC(15,2) DEFAULT 0,      -- Crédito CBS (2026+)
    
    -- Categorização
    category        VARCHAR(50) NOT NULL DEFAULT 'outros',
    -- Categorias: mercadorias, materiais, servicos, combustivel, alimentacao, 
    --             manutencao, equipamentos, aluguel, energia, telefone, outros
    subcategory     VARCHAR(50),
    tags            TEXT[],
    
    -- Para qual módulo/finalidade
    module_ref      VARCHAR(50),                  -- 'mechanic', 'bakery', etc.
    reference_type  VARCHAR(50),                  -- 'os', 'production_order', etc.
    reference_id    UUID,                         -- ID da OS, ordem de produção, etc.
    
    -- Pagamento
    payment_method  VARCHAR(30),                  -- 'pix', 'dinheiro', 'cartao', 'boleto', 'prazo'
    paid            BOOLEAN DEFAULT TRUE,
    due_date        DATE,
    paid_at         TIMESTAMPTZ,
    
    -- Datas
    issue_date      TIMESTAMPTZ NOT NULL,         -- Data de emissão da nota
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    registered_by   UUID REFERENCES users(id),
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'cancelled', 'duplicate', 'pending_review')),
    
    -- Observações
    notes           TEXT,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: expense_items (Itens de cada despesa)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE expense_items (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    expense_id      UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    
    -- Dados do produto
    item_order      INT NOT NULL,
    product_code    VARCHAR(60),                  -- Código do fornecedor
    ean             VARCHAR(14),                  -- Código de barras
    description     VARCHAR(500) NOT NULL,
    ncm             VARCHAR(8),                   -- NCM do produto
    cfop            VARCHAR(4),
    
    -- Quantidades
    quantity        NUMERIC(15,4) NOT NULL,
    unit            VARCHAR(10) NOT NULL DEFAULT 'UN',
    unit_price      NUMERIC(15,4) NOT NULL,
    discount        NUMERIC(15,2) DEFAULT 0,
    total_price     NUMERIC(15,2) NOT NULL,
    
    -- Impostos do item
    icms_base       NUMERIC(15,2) DEFAULT 0,
    icms_rate       NUMERIC(5,2) DEFAULT 0,
    icms_amount     NUMERIC(15,2) DEFAULT 0,
    ipi_amount      NUMERIC(15,2) DEFAULT 0,
    
    -- Categorização automática
    auto_category   VARCHAR(50),                  -- Sugerido pelo NCM
    
    -- Vínculo com estoque (opcional)
    product_id      UUID,                         -- Se vincular com produto interno
    added_to_stock  BOOLEAN DEFAULT FALSE
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: expense_categories (Categorias configuráveis)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE expense_categories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID REFERENCES tenants(id),  -- NULL = global
    
    code            VARCHAR(50) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    icon            VARCHAR(50),
    color           VARCHAR(7),                   -- Hex color
    
    -- Para IR/Imposto
    tax_deductible  BOOLEAN DEFAULT TRUE,
    tax_code        VARCHAR(20),                  -- Código para IRPF/IRPJ
    
    -- NCMs que se enquadram nesta categoria
    ncm_patterns    TEXT[],                       -- Ex: ['8703%', '8704%'] = veículos
    
    parent_id       UUID REFERENCES expense_categories(id),
    display_order   INT DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    
    UNIQUE(tenant_id, code)
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: expense_recurring (Despesas recorrentes)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE expense_recurring (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    
    description     VARCHAR(255) NOT NULL,
    supplier_name   VARCHAR(255),
    category        VARCHAR(50) NOT NULL,
    amount          NUMERIC(15,2) NOT NULL,
    
    -- Recorrência
    frequency       VARCHAR(20) NOT NULL,         -- 'monthly', 'weekly', 'yearly'
    day_of_month    INT,                          -- Dia do mês (1-31)
    day_of_week     INT,                          -- Dia da semana (0-6)
    
    -- Controle
    next_due_date   DATE NOT NULL,
    last_generated  DATE,
    is_active       BOOLEAN DEFAULT TRUE,
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- TABELA: qrcode_scan_log (Log de leituras de QR Code)
-- ═══════════════════════════════════════════════════════════
CREATE TABLE qrcode_scan_log (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID REFERENCES tenants(id),
    user_id         UUID REFERENCES users(id),
    
    -- Dados do scan
    qr_content      TEXT NOT NULL,                -- Conteúdo bruto do QR
    qr_type         VARCHAR(20),                  -- 'nfce', 'nfe', 'pix', 'unknown'
    
    -- Resultado
    success         BOOLEAN NOT NULL,
    expense_id      UUID REFERENCES expenses(id),
    error_message   TEXT,
    
    -- Contexto
    device_type     VARCHAR(20),
    scanned_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ═══════════════════════════════════════════════════════════
-- RLS
-- ═══════════════════════════════════════════════════════════
ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;
ALTER TABLE expense_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE expense_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE expense_recurring ENABLE ROW LEVEL SECURITY;
ALTER TABLE qrcode_scan_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY exp_isolation ON expenses USING (tenant_id = current_tenant_id());
CREATE POLICY exp_items_isolation ON expense_items USING (tenant_id = current_tenant_id());
CREATE POLICY exp_cat_isolation ON expense_categories USING (tenant_id = current_tenant_id() OR tenant_id IS NULL);
CREATE POLICY exp_rec_isolation ON expense_recurring USING (tenant_id = current_tenant_id());
CREATE POLICY qr_log_isolation ON qrcode_scan_log USING (tenant_id = current_tenant_id());

-- ═══════════════════════════════════════════════════════════
-- ÍNDICES
-- ═══════════════════════════════════════════════════════════
CREATE INDEX idx_expenses_tenant ON expenses(tenant_id);
CREATE INDEX idx_expenses_date ON expenses(tenant_id, issue_date DESC);
CREATE INDEX idx_expenses_category ON expenses(tenant_id, category);
CREATE INDEX idx_expenses_supplier ON expenses(tenant_id, supplier_cnpj);
CREATE INDEX idx_expenses_nfe_key ON expenses(nfe_key) WHERE nfe_key IS NOT NULL;
CREATE INDEX idx_expense_items_expense ON expense_items(expense_id);
CREATE INDEX idx_expense_items_ncm ON expense_items(ncm) WHERE ncm IS NOT NULL;

-- ═══════════════════════════════════════════════════════════
-- TRIGGERS
-- ═══════════════════════════════════════════════════════════
CREATE TRIGGER trg_expenses_updated 
    BEFORE UPDATE ON expenses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ═══════════════════════════════════════════════════════════
-- SEED: Categorias padrão
-- ═══════════════════════════════════════════════════════════
INSERT INTO expense_categories (tenant_id, code, name, icon, color, tax_deductible, ncm_patterns, display_order) VALUES
(NULL, 'mercadorias', 'Mercadorias para Revenda', 'package', '#3B82F6', TRUE, '{}', 1),
(NULL, 'materiais', 'Materiais e Insumos', 'box', '#10B981', TRUE, '{}', 2),
(NULL, 'pecas', 'Peças e Componentes', 'cog', '#6366F1', TRUE, ARRAY['8708%', '8409%', '4011%'], 3),
(NULL, 'combustivel', 'Combustível', 'fuel', '#F59E0B', TRUE, ARRAY['2710%'], 4),
(NULL, 'alimentacao', 'Alimentação', 'utensils', '#EF4444', FALSE, ARRAY['19%', '20%', '21%', '22%'], 5),
(NULL, 'manutencao', 'Manutenção e Reparos', 'wrench', '#8B5CF6', TRUE, '{}', 6),
(NULL, 'equipamentos', 'Equipamentos e Ferramentas', 'tool', '#EC4899', TRUE, ARRAY['82%', '84%', '85%'], 7),
(NULL, 'aluguel', 'Aluguel', 'home', '#14B8A6', TRUE, '{}', 8),
(NULL, 'energia', 'Energia Elétrica', 'zap', '#FBBF24', TRUE, '{}', 9),
(NULL, 'agua', 'Água', 'droplet', '#06B6D4', TRUE, '{}', 10),
(NULL, 'telefone', 'Telefone e Internet', 'phone', '#84CC16', TRUE, '{}', 11),
(NULL, 'servicos', 'Serviços de Terceiros', 'users', '#A855F7', TRUE, '{}', 12),
(NULL, 'outros', 'Outros', 'more-horizontal', '#6B7280', TRUE, '{}', 99);

-- ═══════════════════════════════════════════════════════════
-- VIEW: Resumo de despesas por período
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE VIEW expense_summary AS
SELECT
    e.tenant_id,
    DATE_TRUNC('month', e.issue_date) AS month,
    e.category,
    COUNT(*) AS count,
    SUM(e.total_amount) AS total,
    SUM(e.ibs_credit) AS ibs_credit,
    SUM(e.cbs_credit) AS cbs_credit,
    SUM(e.icms_amount) AS icms_paid
FROM expenses e
WHERE e.status = 'active'
GROUP BY e.tenant_id, DATE_TRUNC('month', e.issue_date), e.category;

-- ═══════════════════════════════════════════════════════════
-- VIEW: Despesas para relatório de IR
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE VIEW expense_tax_report AS
SELECT
    e.tenant_id,
    EXTRACT(YEAR FROM e.issue_date) AS year,
    EXTRACT(MONTH FROM e.issue_date) AS month,
    c.code AS category_code,
    c.name AS category_name,
    c.tax_deductible,
    SUM(e.total_amount) AS total,
    SUM(e.ibs_credit + e.cbs_credit) AS tax_credit,
    COUNT(*) AS doc_count
FROM expenses e
JOIN expense_categories c ON c.code = e.category AND (c.tenant_id = e.tenant_id OR c.tenant_id IS NULL)
WHERE e.status = 'active'
GROUP BY e.tenant_id, EXTRACT(YEAR FROM e.issue_date), EXTRACT(MONTH FROM e.issue_date), 
         c.code, c.name, c.tax_deductible;

-- ═══════════════════════════════════════════════════════════
-- FUNÇÃO: Categorizar automaticamente pelo NCM
-- ═══════════════════════════════════════════════════════════
CREATE OR REPLACE FUNCTION auto_categorize_expense(p_tenant_id UUID, p_ncm VARCHAR(8))
RETURNS VARCHAR(50)
LANGUAGE plpgsql AS $$
DECLARE
    v_category VARCHAR(50);
BEGIN
    -- Buscar categoria que tem o padrão NCM
    SELECT code INTO v_category
    FROM expense_categories
    WHERE (tenant_id = p_tenant_id OR tenant_id IS NULL)
      AND is_active = TRUE
      AND EXISTS (
          SELECT 1 FROM unnest(ncm_patterns) pattern
          WHERE p_ncm LIKE REPLACE(pattern, '%', '')
      )
    ORDER BY tenant_id NULLS LAST
    LIMIT 1;
    
    RETURN COALESCE(v_category, 'outros');
END;
$$;

-- ═══════════════════════════════════════════════════════════
-- PERMISSÕES
-- ═══════════════════════════════════════════════════════════
GRANT SELECT, INSERT, UPDATE ON expenses TO nexo_api;
GRANT SELECT, INSERT ON expense_items TO nexo_api;
GRANT SELECT ON expense_categories TO nexo_api;
GRANT SELECT, INSERT, UPDATE ON expense_recurring TO nexo_api;
GRANT INSERT, SELECT ON qrcode_scan_log TO nexo_api;
GRANT SELECT ON expense_summary TO nexo_api;
GRANT SELECT ON expense_tax_report TO nexo_api;

COMMIT;
