#!/bin/bash
# =============================================================================
# NEXO FLASH ERP — Deploy com zero downtime
# Execute como usuário 'deploy' no servidor:
#   cd /opt/nexoflash && bash deploy.sh
#
# O que este script faz:
#   1. Valida que o .env está preenchido
#   2. Puxa a versão mais recente do GitHub
#   3. Builda a imagem Docker do backend
#   4. Builda o frontend Next.js
#   5. Sobe os serviços em ordem (banco → redis → nats → api → frontend)
#   6. Roda as migrations do banco
#   7. Verifica saúde de todos os serviços
#   8. Em caso de falha: rollback automático para a versão anterior
# =============================================================================

set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()      { echo -e "${GREEN}[✓]${NC} $1"; }
warn()     { echo -e "${YELLOW}[!]${NC} $1"; }
err()      { echo -e "${RED}[✗]${NC} $1"; }
step()     { echo -e "\n${BLUE}▶ $1${NC}"; }
success()  { echo -e "\n${GREEN}████████████████████████████████████${NC}"; echo -e "${GREEN}  $1${NC}"; echo -e "${GREEN}████████████████████████████████████${NC}\n"; }

APP_DIR="/opt/nexoflash"
REPO="https://github.com/Beico10/nexo-flash-erp.git"
COMPOSE="docker compose -f deployments/docker/docker-compose.yml --env-file .env"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

cd "$APP_DIR"

# ─── Validações ───────────────────────────────────────────────────────────────

step "Validando configuração..."

[ ! -f .env ] && { err ".env não encontrado! Execute: cp .env.template .env && nano .env"; exit 1; }

# Verificar se os secrets foram preenchidos (não podem ter o valor padrão)
if grep -q "GERE_COM" .env; then
    err "O arquivo .env contém valores não preenchidos (GERE_COM_...)!"
    err "Preencha todos os secrets antes de fazer deploy."
    grep "GERE_COM" .env | awk -F= '{print "  → " $1}'
    exit 1
fi
log "Configuração válida"

# ─── Backup da versão atual ───────────────────────────────────────────────────

step "Salvando versão atual para rollback..."
CURRENT_IMAGE=$(docker images nexoflash-api --format "{{.ID}}" | head -1 || true)
if [ -n "$CURRENT_IMAGE" ]; then
    docker tag "$CURRENT_IMAGE" "nexoflash-api:rollback-$TIMESTAMP" 2>/dev/null || true
    log "Backup da imagem atual salvo como nexoflash-api:rollback-$TIMESTAMP"
fi

# ─── Atualizar código ─────────────────────────────────────────────────────────

step "Atualizando código do GitHub..."
if [ -d ".git" ]; then
    git fetch origin
    git reset --hard origin/master
    log "Código atualizado: $(git log -1 --format='%h %s')"
else
    git clone "$REPO" .
    log "Repositório clonado"
fi

# ─── Build do Backend ─────────────────────────────────────────────────────────

step "Buildando imagem Docker do backend..."
docker build \
    --tag nexoflash-api:latest \
    --tag "nexoflash-api:$TIMESTAMP" \
    --build-arg VERSION="$(git describe --tags --always 2>/dev/null || echo dev)" \
    --cache-from nexoflash-api:latest \
    . 2>&1 | tail -5
log "Backend buildado"

# ─── Build do Frontend ────────────────────────────────────────────────────────

step "Buildando frontend Next.js..."
cd frontend

# Instalar dependências se necessário
if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
    npm ci --silent
fi

# Build de produção
NEXT_PUBLIC_API_URL="https://$(grep API_HOST "$APP_DIR/.env" | cut -d= -f2)" \
    npm run build 2>&1 | tail -5

cd "$APP_DIR"
log "Frontend buildado"

# ─── Subir serviços ───────────────────────────────────────────────────────────

step "Subindo serviços com Docker Compose..."

# Ordem importa: dependências primeiro
$COMPOSE up -d postgres redis nats
log "Banco de dados, Redis e NATS iniciados"

# Aguardar banco ficar saudável
echo -n "  Aguardando PostgreSQL..."
for i in $(seq 1 30); do
    if docker compose -f deployments/docker/docker-compose.yml exec -T postgres \
        pg_isready -U postgres -d nexoflash &>/dev/null; then
        echo " OK"
        break
    fi
    echo -n "."
    sleep 2
    [ $i -eq 30 ] && { echo ""; err "PostgreSQL não iniciou em 60s!"; exit 1; }
done

# ─── Migrations ───────────────────────────────────────────────────────────────

step "Executando migrations do banco..."
for migration in migrations/*.sql; do
    MIGRATION_NAME=$(basename "$migration")
    # Verificar se migration já foi aplicada
    APPLIED=$(docker compose -f deployments/docker/docker-compose.yml exec -T postgres \
        psql -U postgres -d nexoflash -tAc \
        "SELECT COUNT(*) FROM schema_migrations WHERE name = '$MIGRATION_NAME'" 2>/dev/null || echo "0")
    
    if [ "$APPLIED" = "0" ]; then
        echo "  Aplicando $MIGRATION_NAME..."
        docker compose -f deployments/docker/docker-compose.yml exec -T postgres \
            psql -U postgres -d nexoflash < "$migration"
        docker compose -f deployments/docker/docker-compose.yml exec -T postgres \
            psql -U postgres -d nexoflash -c \
            "INSERT INTO schema_migrations (name, applied_at) VALUES ('$MIGRATION_NAME', NOW())" 2>/dev/null || true
        log "$MIGRATION_NAME aplicado"
    else
        echo "  ↷ $MIGRATION_NAME já aplicado"
    fi
done

# ─── Subir API ────────────────────────────────────────────────────────────────

step "Subindo API backend..."
$COMPOSE up -d api
sleep 3

# Health check da API
echo -n "  Verificando saúde da API..."
for i in $(seq 1 15); do
    if curl -sf http://localhost:8080/health | grep -q '"status":"ok"'; then
        echo " OK"
        break
    fi
    echo -n "."
    sleep 2
    if [ $i -eq 15 ]; then
        echo ""
        err "API não iniciou! Verificando logs..."
        $COMPOSE logs api --tail=20
        # Rollback automático
        if [ -n "$CURRENT_IMAGE" ]; then
            warn "Fazendo rollback para versão anterior..."
            docker tag "nexoflash-api:rollback-$TIMESTAMP" nexoflash-api:latest
            $COMPOSE up -d api
        fi
        exit 1
    fi
done

# ─── Status final ─────────────────────────────────────────────────────────────

step "Verificando status de todos os serviços..."
$COMPOSE ps

# Limpar imagens antigas (manter últimas 3)
docker images nexoflash-api --format "{{.Tag}}" | \
    grep -v "latest\|rollback" | \
    sort -r | \
    tail -n +4 | \
    xargs -I{} docker rmi "nexoflash-api:{}" 2>/dev/null || true

success "✅ Deploy concluído com sucesso! ($TIMESTAMP)"
echo "  API:      https://$(grep API_HOST .env | cut -d= -f2)"
echo "  Frontend: https://$(grep FRONTEND_HOST .env | cut -d= -f2 2>/dev/null || echo 'configure FRONTEND_HOST')"
echo "  Logs:     docker compose -f deployments/docker/docker-compose.yml logs -f api"
