#!/bin/bash
# =============================================================================
# NEXO FLASH ERP — Setup inicial do servidor Hetzner VPS
# Execute UMA VEZ no servidor recém-criado como root:
#   curl -fsSL https://raw.githubusercontent.com/Beico10/nexo-flash-erp/master/deployments/hetzner/setup-server.sh | bash
#
# O que este script faz:
#   1. Atualiza o sistema (Ubuntu 24.04)
#   2. Instala Docker + Docker Compose
#   3. Cria usuário 'deploy' sem root
#   4. Configura firewall (UFW) — apenas 80, 443, 22
#   5. Instala Fail2ban (proteção contra força bruta)
#   6. Configura swap de 2GB
#   7. Clona o repositório e prepara o ambiente
# =============================================================================

set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }
step() { echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; echo -e "${BLUE}  $1${NC}"; echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; }

[ "$EUID" -ne 0 ] && err "Execute como root: sudo bash setup-server.sh"

step "1/7 — Atualizando sistema"
apt-get update -qq && apt-get upgrade -y -qq
apt-get install -y -qq curl wget git ufw fail2ban htop unzip jq
log "Sistema atualizado"

step "2/7 — Instalando Docker"
curl -fsSL https://get.docker.com | sh
systemctl enable docker && systemctl start docker
# Docker Compose v2 (plugin)
apt-get install -y -qq docker-compose-plugin
docker --version && docker compose version
log "Docker instalado"

step "3/7 — Criando usuário 'deploy'"
if ! id "deploy" &>/dev/null; then
    useradd -m -s /bin/bash -G docker,sudo deploy
    # Gerar senha aleatória para o usuário deploy
    DEPLOY_PASS=$(openssl rand -base64 16)
    echo "deploy:$DEPLOY_PASS" | chpasswd
    warn "Senha do usuário deploy: $DEPLOY_PASS"
    warn "SALVE ESTA SENHA EM LOCAL SEGURO!"
fi

# Criar diretório do app
mkdir -p /opt/nexoflash
chown deploy:deploy /opt/nexoflash
log "Usuário 'deploy' criado"

step "4/7 — Configurando Firewall (UFW)"
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp   comment 'SSH'
ufw allow 80/tcp   comment 'HTTP → redireciona para HTTPS'
ufw allow 443/tcp  comment 'HTTPS'
ufw --force enable
ufw status
log "Firewall configurado"

step "5/7 — Configurando Fail2ban"
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime  = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port    = ssh
logpath = %(sshd_log)s

[nginx-http-auth]
enabled = true
EOF
systemctl enable fail2ban && systemctl restart fail2ban
log "Fail2ban configurado"

step "6/7 — Configurando Swap (2GB)"
if [ ! -f /swapfile ]; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    sysctl vm.swappiness=10
    echo 'vm.swappiness=10' >> /etc/sysctl.conf
    log "Swap de 2GB configurado"
else
    warn "Swap já existe — ignorando"
fi

step "7/7 — Preparando diretório do app"
cd /opt/nexoflash

# Criar .env template
cat > /opt/nexoflash/.env.template << 'EOF'
# PREENCHA TODOS OS VALORES ANTES DE FAZER O DEPLOY
# cp .env.template .env && nano .env

# ─── Banco de dados ────────────────────────────────
POSTGRES_PASSWORD=GERE_COM_openssl_rand_base64_32
APP_DB_PASSWORD=GERE_COM_openssl_rand_base64_32

# ─── Redis ────────────────────────────────────────
REDIS_PASSWORD=GERE_COM_openssl_rand_base64_32

# ─── NATS ─────────────────────────────────────────
NATS_USER=nexo
NATS_PASSWORD=GERE_COM_openssl_rand_base64_32

# ─── JWT (gere: openssl rand -hex 32) ─────────────
JWT_SECRET=GERE_COM_openssl_rand_hex_32

# ─── Domínio ──────────────────────────────────────
API_HOST=api.seudominio.com.br
FRONTEND_HOST=app.seudominio.com.br

# ─── Ambiente ─────────────────────────────────────
APP_ENV=production
LOG_LEVEL=info
PORT=8080
EOF

chown deploy:deploy /opt/nexoflash/.env.template

echo ""
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}  ✅ Servidor configurado com sucesso!   ${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo ""
echo "Próximos passos:"
echo "  1. su - deploy"
echo "  2. cd /opt/nexoflash"
echo "  3. cp .env.template .env && nano .env  (preencher secrets)"
echo "  4. bash deploy.sh"
