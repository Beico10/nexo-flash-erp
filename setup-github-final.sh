#!/bin/bash
# =============================================================================
# Nexo Flash ERP — Setup COMPLETO v1.0 (todos os módulos)
# Execute: bash setup-github-final.sh SEU_USUARIO NOME_DO_REPO
# Exemplo: bash setup-github-final.sh antonio nexo-flash-erp
# =============================================================================

set -e
GITHUB_USER="${1:?Uso: bash setup-github-final.sh USUARIO REPO}"
REPO_NAME="${2:?Uso: bash setup-github-final.sh USUARIO REPO}"

echo ""
echo "=================================================="
echo " Nexo Flash ERP — Configurando repositório GitHub"
echo " Usuario : $GITHUB_USER"
echo " Repo    : $REPO_NAME"
echo "=================================================="

# Extrai o projeto
if [ -f "nexo-flash-foundation-complete.tar.gz" ]; then
    echo "[1/6] Extraindo arquivos..."
    tar -xzf nexo-flash-foundation-complete.tar.gz
    cd nexo-flash
elif [ -d "nexo-flash" ]; then
    cd nexo-flash
else
    echo "ERRO: arquivo nexo-flash-foundation-complete.tar.gz não encontrado"
    exit 1
fi

echo "[2/6] Criando .gitignore seguro..."
cat > .gitignore << 'EOF'
# Binários compilados
nexo-flash
*.exe
*.out

# Variáveis de ambiente — NUNCA commitar secrets
.env
.env.local
.env.production
!.env.example

# Go
vendor/
*.test
coverage.out
dist/

# Docker volumes locais
data/
postgres_data/
redis_data/

# IDE
.idea/
.vscode/
*.swp
*.swo
.DS_Store
EOF

echo "[3/6] Criando .env.example..."
cat > .env.example << 'EOF'
# ================================================================
# Nexo Flash ERP — Variáveis de Ambiente
# Copie para .env e preencha com valores reais
# NUNCA commite o .env no git
# ================================================================

# Banco de dados PostgreSQL
POSTGRES_PASSWORD=COLOQUE_UMA_SENHA_FORTE_AQUI
APP_DB_PASSWORD=COLOQUE_OUTRA_SENHA_FORTE

# Redis
REDIS_PASSWORD=COLOQUE_UMA_SENHA_FORTE

# NATS JetStream
NATS_USER=nexo
NATS_PASSWORD=COLOQUE_UMA_SENHA_FORTE

# JWT (gere com: openssl rand -hex 32)
JWT_SECRET=GERE_COM_OPENSSL_RAND_HEX_32

# Domínio
API_HOST=api.seudominio.com.br

# Ambiente
APP_ENV=development
LOG_LEVEL=info
PORT=8080
EOF

echo "[4/6] Inicializando repositório Git..."
git init
git checkout -b main

echo "[5/6] Criando commit completo..."
git add .
git commit -m "feat: nexo flash erp foundation completo v1.0

Módulos implementados:
- core/module_registry.go    → micro-kernel plug-and-play
- modules/mechanic           → OS digital, peças, aprovação WhatsApp
- modules/bakery             → PDV rápido, balanças, gestão de perdas
- modules/industry           → PCP, BOM (ficha técnica), insumos
- modules/aesthetics         → agenda (trava de conflito), split pagamento
- modules/shoes              → matriz de grade Cor/Tamanho/SKU, comissões
- modules/logistics          → CT-e, contratos multi-cliente, DRE da viagem
- tax/engine.go              → Motor Fiscal IBS/CBS 2026 (NCM + Cesta Básica)
- ai/gateway.go              → Human-in-the-Loop (toda IA passa por aprovação)
- ai/concierge.go            → Leitura de XML NF-e (onboarding 90% automático)
- pkg/middleware/auth.go     → JWT + RLS por tenant em cada request
- pkg/cache/redis.go         → Cache alíquotas NCM (TTL 1h) + sessões
- pkg/eventbus/bus.go        → NATS JetStream event bus
- migrations/001_foundation_rls.sql  → PostgreSQL RLS obrigatório
- migrations/002_business_modules.sql → Tabelas de todos os módulos
- deployments/docker/docker-compose.yml → Stack completa
- Dockerfile                 → Imagem scratch ~10MB, usuário não-root

Stack: Go 1.22 · PostgreSQL 16 · Redis 7 · NATS JetStream · Traefik v3"

echo "[6/6] Enviando para GitHub..."
echo ""
echo "  IMPORTANTE: Se pedir senha, use um Personal Access Token"
echo "  (GitHub > Settings > Developer settings > Personal access tokens)"
echo ""
git remote add origin "https://github.com/$GITHUB_USER/$REPO_NAME.git"
git push -u origin main

echo ""
echo "=================================================="
echo " PRONTO! Repositório disponível em:"
echo " https://github.com/$GITHUB_USER/$REPO_NAME"
echo "=================================================="
echo ""
echo "Proximos passos recomendados:"
echo "  1. Settings > Branches > Add rule > main: exigir PR antes de merge"
echo "  2. Settings > Security > Enable Dependabot alerts"
echo "  3. Preencha o .env com senhas reais e rode: docker compose up -d"
