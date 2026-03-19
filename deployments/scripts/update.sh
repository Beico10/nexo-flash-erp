#!/bin/bash
# =============================================================================
# NEXO FLASH ERP — Atualização sem downtime
# Execute para atualizar o sistema após novos commits no GitHub:
#   bash /opt/nexo-flash/deployments/scripts/update.sh
# =============================================================================

set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${BLUE}[→]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }

APP_DIR="/opt/nexo-flash"
COMPOSE="docker compose -f $APP_DIR/deployments/docker/docker-compose.yml --env-file $APP_DIR/.env"

echo ""
echo "══════════════════════════════════════"
echo "  Nexo Flash — Atualizando sistema"
echo "  $(date '+%d/%m/%Y %H:%M:%S')"
echo "══════════════════════════════════════"
echo ""

# 1. Backup antes de atualizar
info "Fazendo backup do banco..."
/opt/backup-nexo.sh
log "Backup concluído"

# 2. Puxar código novo
info "Atualizando código do GitHub..."
cd "$APP_DIR"
git fetch origin
git reset --hard origin/master
log "Código atualizado"

# 3. Build nova imagem
info "Buildando nova imagem..."
$COMPOSE build api --no-cache
log "Build concluído"

# 4. Rolling restart da API (sem derrubar banco/redis/nats)
info "Reiniciando API (zero downtime)..."
$COMPOSE up -d --no-deps api
log "API reiniciada"

# 5. Aguardar health check
info "Verificando saúde..."
for i in {1..15}; do
    if curl -sf "http://localhost:8080/health" | grep -q "ok"; then
        log "API saudável ✓"
        break
    fi
    sleep 2
    warn "Aguardando... ($i/15)"
done

# 6. Limpar imagens antigas
docker image prune -f --filter "until=24h" > /dev/null

echo ""
log "Atualização concluída em $(date '+%H:%M:%S')"
echo ""
