#!/bin/bash
# =============================================================================
# NEXO FLASH ERP — Deploy Automático no Hetzner VPS
# Execute UMA VEZ no servidor novo:
#   bash deploy.sh
#
# O que este script faz:
#   1. Instala dependências (Docker, Git, Nginx, Certbot)
#   2. Configura firewall (UFW)
#   3. Clona o repositório do GitHub
#   4. Gera senhas seguras automaticamente
#   5. Sobe todos os serviços com Docker Compose
#   6. Configura HTTPS com Let's Encrypt
#   7. Configura backup automático diário
# =============================================================================

set -euo pipefail

# ─── Cores para output ───────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
info() { echo -e "${BLUE}[i]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }

# ─── Variáveis — EDITE ANTES DE RODAR ────────────────────────────────────────
GITHUB_REPO="${GITHUB_REPO:-https://github.com/Beico10/nexo-flash-erp.git}"
DOMAIN="${DOMAIN:-api.nexoflash.com.br}"
FRONTEND_DOMAIN="${FRONTEND_DOMAIN:-app.nexoflash.com.br}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:-seu@email.com.br}"
APP_DIR="/opt/nexo-flash"
DEPLOY_USER="deploy"

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║     Nexo Flash ERP — Deploy Hetzner v1.0     ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""

# ─── 1. Verificar root ───────────────────────────────────────────────────────
[[ $EUID -ne 0 ]] && err "Execute como root: sudo bash deploy.sh"

# ─── 2. Atualizar sistema ────────────────────────────────────────────────────
info "Atualizando sistema..."
apt-get update -qq && apt-get upgrade -y -qq
apt-get install -y -qq curl wget git ufw fail2ban ca-certificates gnupg lsb-release

# ─── 3. Instalar Docker ──────────────────────────────────────────────────────
info "Instalando Docker..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker.gpg] \
        https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
        > /etc/apt/sources.list.d/docker.list
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
    systemctl enable docker
    log "Docker instalado"
else
    log "Docker já instalado"
fi

# ─── 4. Instalar Certbot ─────────────────────────────────────────────────────
info "Instalando Certbot (Let's Encrypt)..."
apt-get install -y -qq certbot python3-certbot-nginx
log "Certbot instalado"

# ─── 5. Criar usuário deploy ─────────────────────────────────────────────────
info "Criando usuário deploy..."
if ! id "$DEPLOY_USER" &>/dev/null; then
    useradd -m -s /bin/bash "$DEPLOY_USER"
    usermod -aG docker "$DEPLOY_USER"
    log "Usuário '$DEPLOY_USER' criado"
else
    log "Usuário '$DEPLOY_USER' já existe"
fi

# ─── 6. Configurar firewall ──────────────────────────────────────────────────
info "Configurando firewall (UFW)..."
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp   # HTTP (redireciona para HTTPS)
ufw allow 443/tcp  # HTTPS
ufw --force enable
log "Firewall configurado"

# ─── 7. Configurar Fail2ban ──────────────────────────────────────────────────
info "Configurando Fail2ban..."
cat > /etc/fail2ban/jail.local << 'F2B'
[DEFAULT]
bantime  = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port    = ssh

[nginx-http-auth]
enabled = true
F2B
systemctl enable fail2ban
systemctl restart fail2ban
log "Fail2ban configurado"

# ─── 8. Clonar repositório ───────────────────────────────────────────────────
info "Clonando repositório..."
mkdir -p "$APP_DIR"
if [ -d "$APP_DIR/.git" ]; then
    cd "$APP_DIR" && git pull origin main
    log "Repositório atualizado"
else
    git clone "$GITHUB_REPO" "$APP_DIR"
    log "Repositório clonado"
fi
chown -R "$DEPLOY_USER:$DEPLOY_USER" "$APP_DIR"

# ─── 9. Gerar .env com senhas seguras ────────────────────────────────────────
info "Gerando configuração segura..."
ENV_FILE="$APP_DIR/.env"

if [ ! -f "$ENV_FILE" ]; then
    # Gerar senhas criptograficamente seguras
    PG_PASS=$(openssl rand -hex 24)
    APP_DB_PASS=$(openssl rand -hex 24)
    REDIS_PASS=$(openssl rand -hex 24)
    NATS_PASS=$(openssl rand -hex 24)
    JWT_SECRET=$(openssl rand -hex 32)

    cat > "$ENV_FILE" << EOF
# ── Banco de dados ──────────────────────────────────────────
POSTGRES_PASSWORD=${PG_PASS}
APP_DB_PASSWORD=${APP_DB_PASS}
DATABASE_URL=postgres://app_user:${APP_DB_PASS}@postgres:5432/nexoflash?sslmode=require

# ── Redis ───────────────────────────────────────────────────
REDIS_PASSWORD=${REDIS_PASS}
REDIS_URL=redis://:${REDIS_PASS}@redis:6379/0

# ── NATS ────────────────────────────────────────────────────
NATS_USER=nexo
NATS_PASSWORD=${NATS_PASS}
NATS_URL=nats://nexo:${NATS_PASS}@nats:4222

# ── JWT ─────────────────────────────────────────────────────
JWT_SECRET=${JWT_SECRET}

# ── Domínio ─────────────────────────────────────────────────
API_HOST=${DOMAIN}
FRONTEND_HOST=${FRONTEND_DOMAIN}
BASE_URL=https://${DOMAIN}

# ── Ambiente ────────────────────────────────────────────────
APP_ENV=production
LOG_LEVEL=info
PORT=8080
EOF
    chmod 600 "$ENV_FILE"
    log ".env gerado com senhas seguras"

    # Salvar as senhas para referência do administrador
    cat > /root/nexo-flash-credentials.txt << EOF
═══════════════════════════════════════════════
  NEXO FLASH ERP — Credenciais do Servidor
  Gerado em: $(date)
  GUARDE ESTE ARQUIVO EM LOCAL SEGURO!
═══════════════════════════════════════════════

PostgreSQL Admin Password : ${PG_PASS}
App DB Password           : ${APP_DB_PASS}
Redis Password            : ${REDIS_PASS}
NATS Password             : ${NATS_PASS}
JWT Secret                : ${JWT_SECRET}

Domínio API               : https://${DOMAIN}
Domínio Frontend          : https://${FRONTEND_DOMAIN}
EOF
    chmod 600 /root/nexo-flash-credentials.txt
    warn "Credenciais salvas em /root/nexo-flash-credentials.txt — GUARDE EM LOCAL SEGURO!"
else
    log ".env já existe — mantendo configuração atual"
fi

# ─── 10. Build e subir serviços ──────────────────────────────────────────────
info "Subindo serviços com Docker Compose..."
cd "$APP_DIR"
docker compose -f deployments/docker/docker-compose.yml --env-file .env pull
docker compose -f deployments/docker/docker-compose.yml --env-file .env up -d --build

# Aguardar PostgreSQL inicializar
info "Aguardando PostgreSQL inicializar..."
sleep 10
for i in {1..30}; do
    if docker compose -f deployments/docker/docker-compose.yml exec -T postgres \
        pg_isready -U postgres -d nexoflash &>/dev/null; then
        log "PostgreSQL pronto"
        break
    fi
    sleep 2
done

log "Todos os serviços no ar"

# ─── 11. Configurar Nginx como proxy reverso ─────────────────────────────────
info "Configurando Nginx..."
apt-get install -y -qq nginx

cat > /etc/nginx/sites-available/nexo-flash << NGINX
# Nexo Flash ERP — Nginx Proxy
# API Backend
server {
    listen 80;
    server_name ${DOMAIN};

    location /.well-known/acme-challenge/ { root /var/www/certbot; }
    location / { return 301 https://\$host\$request_uri; }
}

server {
    listen 443 ssl http2;
    server_name ${DOMAIN};

    ssl_certificate     /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin";

    # Rate limiting
    limit_req_zone \$binary_remote_addr zone=api:10m rate=60r/m;
    limit_req zone=api burst=20 nodelay;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
        proxy_connect_timeout 10s;
    }
}

# Frontend Next.js
server {
    listen 80;
    server_name ${FRONTEND_DOMAIN};
    location /.well-known/acme-challenge/ { root /var/www/certbot; }
    location / { return 301 https://\$host\$request_uri; }
}

server {
    listen 443 ssl http2;
    server_name ${FRONTEND_DOMAIN};

    ssl_certificate     /etc/letsencrypt/live/${FRONTEND_DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${FRONTEND_DOMAIN}/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
NGINX

ln -sf /etc/nginx/sites-available/nexo-flash /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx
log "Nginx configurado"

# ─── 12. Obter certificado SSL ───────────────────────────────────────────────
info "Obtendo certificado SSL (Let's Encrypt)..."
mkdir -p /var/www/certbot
certbot --nginx \
    -d "$DOMAIN" \
    -d "$FRONTEND_DOMAIN" \
    --email "$LETSENCRYPT_EMAIL" \
    --agree-tos \
    --non-interactive \
    --redirect \
    && log "SSL configurado" \
    || warn "SSL falhou — configure o DNS primeiro e rode: certbot --nginx -d $DOMAIN -d $FRONTEND_DOMAIN"

# ─── 13. Renovação automática do SSL ─────────────────────────────────────────
(crontab -l 2>/dev/null; echo "0 3 * * * certbot renew --quiet && systemctl reload nginx") | crontab -
log "Renovação automática de SSL agendada (3h diariamente)"

# ─── 14. Backup automático ───────────────────────────────────────────────────
info "Configurando backup automático..."
cat > /opt/backup-nexo.sh << 'BACKUP'
#!/bin/bash
BACKUP_DIR="/opt/backups/nexo-flash"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"

# Backup do PostgreSQL
docker exec nexo-flash-postgres-1 \
    pg_dump -U postgres nexoflash | \
    gzip > "$BACKUP_DIR/db_${DATE}.sql.gz"

# Manter apenas os últimos 7 dias
find "$BACKUP_DIR" -name "*.gz" -mtime +7 -delete

echo "Backup concluído: $BACKUP_DIR/db_${DATE}.sql.gz"
BACKUP
chmod +x /opt/backup-nexo.sh

# Backup diário às 2h da manhã
(crontab -l 2>/dev/null; echo "0 2 * * * /opt/backup-nexo.sh >> /var/log/nexo-backup.log 2>&1") | crontab -
log "Backup automático configurado (diário às 2h)"

# ─── 15. Health check ────────────────────────────────────────────────────────
info "Verificando saúde dos serviços..."
sleep 5
if curl -sf "http://localhost:8080/health" | grep -q "ok"; then
    log "API respondendo corretamente"
else
    warn "API ainda inicializando — verifique com: docker compose logs api"
fi

# ─── Resumo final ────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║           Deploy Concluído! ✓                ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}API Backend${NC}  : https://${DOMAIN}"
echo -e "  ${GREEN}Frontend${NC}     : https://${FRONTEND_DOMAIN}"
echo -e "  ${GREEN}Health check${NC} : https://${DOMAIN}/health"
echo ""
echo -e "  ${YELLOW}Credenciais${NC} : /root/nexo-flash-credentials.txt"
echo -e "  ${YELLOW}Logs${NC}        : docker compose -f ${APP_DIR}/deployments/docker/docker-compose.yml logs -f"
echo -e "  ${YELLOW}Backups${NC}     : /opt/backups/nexo-flash/"
echo ""
echo -e "  Para atualizar o sistema:"
echo -e "  ${BLUE}bash /opt/nexo-flash/deployments/scripts/update.sh${NC}"
echo ""
