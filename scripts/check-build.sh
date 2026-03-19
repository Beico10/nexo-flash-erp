#!/bin/bash
# ============================================================
# NEXO ONE ERP — Script de Verificação e Build Local
# ============================================================
# DIRETRIZES: Máxima segurança, Mínimo custo
#
# Uso:
#   ./scripts/check-build.sh        # Verifica tudo
#   ./scripts/check-build.sh --fast # Apenas sintaxe Go
# ============================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}       NEXO ONE ERP — Verificação de Build         ${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════${NC}"
echo ""

# Diretório raiz do projeto
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# ════════════════════════════════════════════════════════════
# 1. VERIFICAR FERRAMENTAS
# ════════════════════════════════════════════════════════════
echo -e "${YELLOW}[1/5] Verificando ferramentas...${NC}"

check_command() {
    if command -v "$1" &> /dev/null; then
        echo -e "  ✅ $1 encontrado"
        return 0
    else
        echo -e "  ${RED}❌ $1 não encontrado${NC}"
        return 1
    fi
}

MISSING=0
check_command "go" || MISSING=1
check_command "node" || true  # opcional para backend
check_command "docker" || true  # opcional

if [ $MISSING -eq 1 ]; then
    echo -e "${RED}Erro: Go é obrigatório. Instale com: https://go.dev/dl/${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "  📌 Go versão: ${GO_VERSION}"

# ════════════════════════════════════════════════════════════
# 2. VERIFICAR MÓDULO GO
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${YELLOW}[2/5] Verificando dependências Go...${NC}"

if [ ! -f "go.mod" ]; then
    echo -e "${RED}Erro: go.mod não encontrado${NC}"
    exit 1
fi

go mod download
go mod verify
echo -e "  ✅ Dependências OK"

# ════════════════════════════════════════════════════════════
# 3. VERIFICAR SINTAXE (go vet)
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${YELLOW}[3/5] Verificando sintaxe (go vet)...${NC}"

if ! go vet ./...; then
    echo -e "${RED}Erro: go vet encontrou problemas${NC}"
    exit 1
fi
echo -e "  ✅ Sintaxe OK"

# Modo rápido: para aqui
if [ "$1" == "--fast" ]; then
    echo ""
    echo -e "${GREEN}✅ Verificação rápida concluída!${NC}"
    exit 0
fi

# ════════════════════════════════════════════════════════════
# 4. COMPILAR BACKEND
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${YELLOW}[4/5] Compilando backend...${NC}"

BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"

VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o "$BUILD_DIR/nexo-one" \
    ./cmd/api

SIZE=$(ls -lh "$BUILD_DIR/nexo-one" | awk '{print $5}')
echo -e "  ✅ Binário criado: $BUILD_DIR/nexo-one (${SIZE})"

# ════════════════════════════════════════════════════════════
# 5. VERIFICAR FRONTEND (se existir)
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${YELLOW}[5/5] Verificando frontend...${NC}"

if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
    cd frontend
    if command -v yarn &> /dev/null; then
        yarn install --frozen-lockfile 2>/dev/null || yarn install
        yarn build
        echo -e "  ✅ Frontend OK"
    elif command -v npm &> /dev/null; then
        npm ci 2>/dev/null || npm install
        npm run build
        echo -e "  ✅ Frontend OK"
    else
        echo -e "  ${YELLOW}⚠️ Node/Yarn não encontrado, pulando frontend${NC}"
    fi
    cd ..
else
    echo -e "  ${YELLOW}⚠️ Frontend não encontrado${NC}"
fi

# ════════════════════════════════════════════════════════════
# RESUMO
# ════════════════════════════════════════════════════════════
echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}       ✅ BUILD CONCLUÍDO COM SUCESSO!              ${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
echo ""
echo "Artefatos:"
echo "  📦 Backend:  $BUILD_DIR/nexo-one"
if [ -d "frontend/.next" ]; then
echo "  📦 Frontend: frontend/.next"
fi
echo ""
echo "Próximos passos:"
echo "  1. Configure as variáveis de ambiente (veja .env.example)"
echo "  2. Execute: docker compose up -d"
echo "  3. Acesse: http://localhost:3000"
echo ""
