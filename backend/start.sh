#!/bin/bash
# Script de inicialização do Nexo One ERP (Go puro)

# Carrega variáveis de ambiente
if [ -f /app/backend/.env ]; then
    export $(cat /app/backend/.env | grep -v '^#' | xargs)
fi

# Define porta padrão
export PORT=${PORT:-8001}

# Executa o binário Go
exec /app/nexo-one
