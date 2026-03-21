# Nexo One ERP - PRD

## Problem Statement
Integrar módulo dispatch no sistema ERP existente (Go backend + Next.js frontend)

## Architecture
- Backend: Go puro com handlers modulares
- Frontend: Next.js com componentes React
- Database: Memory repositories (preview) / PostgreSQL (prod)

## What's Been Implemented
- [Jan 2026] Módulo Dispatch conectado:
  - wire.go: DispatchHandler adicionado ao Container
  - main.go: Rotas registradas em protectedMux
  - Sidebar.tsx: Link /dispatch com ícone Truck

## Backlog
- P0: Compilar Go localmente (`go build -o /app/nexo-one ./cmd/api/`)
- P1: Testar fluxo de despacho em lote
- P2: Validar integração com outros módulos (estoque, NF-e)

## Next Tasks
1. Executar build Go no ambiente local
2. Testar endpoint /api/v1/dispatch/*
3. Validar UI da página /dispatch
