# Nexo One ERP — Product Requirements Document

## Visao
ERP SaaS multi-nicho adaptativo ("Camaleao") para Reforma Tributaria Brasil 2026 (IBS/CBS).
Do TOTVS ao cafezinho — gestao para todos.

## Nicho alvo
Industria, Logistica, Mecanica, Estetica, Padaria, Calcados

## Tech Stack
- **Backend:** Go 1.22, Clean Architecture, HTTP Router
- **Frontend:** Next.js 14, React, TypeScript, Tailwind CSS
- **Database (preview):** In-memory repositories (volatile, seeded on startup)
- **Database (producao):** PostgreSQL com Row Level Security (RLS)
- **Cache (producao):** Redis
- **Event Bus (producao):** NATS

## Credenciais demo
- Tenant: `demo` | Email: `admin@demo.com` | Senha: `demo123`

## O que esta implementado

### 2026-03-19 — Motor Fiscal + Backend Go Funcional
- [x] Go 1.22 compilando no ambiente ARM64
- [x] Backend Go rodando na porta 8002 + proxy FastAPI 8001
- [x] Motor Fiscal IBS/CBS 2026 completo (IBS 9.25% + CBS 3.75%, Cesta Basica, Seletivo, Transicao, Cashback)
- [x] Auth JWT (login, refresh, logout)
- [x] Multi-tenancy
- [x] Modulo Mecanica CRUD

### 2026-03-20 — Simulador Fiscal + Integracao Frontend (Fase 1 Completa)
- [x] Simulador Fiscal publico (/simulador-fiscal)
- [x] Mecanica conectada ao backend real
- [x] Dashboard conectado ao backend (KPIs, graficos receita vs impostos, status OS, atividade por modulo)
- [x] Padaria conectada ao backend (lista de produtos com badges CESTA)
- [x] Estetica conectada ao backend (timeline de agendamentos por profissional)
- [x] Aprovacoes IA conectadas ao backend (sugestoes com aprovar/rejeitar)
- [x] Logistica com calculadora de frete (UI funcional, requer contratos no backend)
- [x] Industria pagina estatica
- [x] Calcados pagina estatica
- [x] Sidebar com navegacao completa
- [x] Auth.ts corrigido (URLs relativas com prefixo /api)
- [x] Frontend rebuild com todas as paginas
- [x] 29/29 testes backend + 100% frontend passando (iteracao 3)

## API Endpoints
| Metodo | Rota | Auth | Descricao |
|--------|------|------|-----------|
| GET | /api/health | N | Health check |
| POST | /api/auth/login | N | Login |
| POST | /api/auth/refresh | S | Refresh token |
| POST | /api/auth/logout | S | Logout |
| GET | /api/auth/me | S | User info |
| GET | /api/v1/tax/ncm-list | N | Lista NCMs |
| POST | /api/v1/tax/simulate | N | Simulador fiscal publico |
| POST | /api/v1/tax/calculate | S | Calculo fiscal autenticado |
| GET | /api/v1/dashboard/stats | S | Dashboard KPIs |
| POST | /api/v1/mechanic/os | S | Criar OS |
| GET | /api/v1/mechanic/os | S | Listar OS abertas |
| GET | /api/v1/mechanic/os/{id} | S | Buscar OS |
| PATCH | /api/v1/mechanic/os/{id}/status | S | Atualizar status |
| GET | /api/v1/mechanic/os/plate/{plate} | S | Buscar por placa |
| GET | /api/v1/bakery/products | S | Listar produtos padaria |
| POST | /api/v1/bakery/sale | S | Registrar venda |
| GET | /api/v1/aesthetics/appointments | S | Listar agendamentos |
| POST | /api/v1/aesthetics/appointments | S | Criar agendamento |
| GET | /api/v1/ai/suggestions | S | Listar sugestoes IA |
| POST | /api/v1/ai/suggestions/{id}/approve | S | Aprovar sugestao |
| POST | /api/v1/ai/suggestions/{id}/reject | S | Rejeitar sugestao |
| POST | /api/v1/logistics/freight/calculate | S | Calcular frete |
| POST | /api/v1/payments/pix | S | Gerar PIX |
| POST | /api/v1/payments/boleto | S | Gerar boleto |

## Backlog Priorizado

### P1 — Fase 2: Self-Service & Onboarding
- [ ] Sistema de Billing/Assinaturas
- [ ] Free trial com verificacao WhatsApp
- [ ] Jornada do usuario (event tracking)

### P1 — Fase 3: Features de Receita
- [ ] Scanner QR Code NFC-e para despesas
- [ ] Emissao de NF-e/CT-e via SEFAZ

### P2 — Futuro
- [ ] IA Co-Piloto (sugestoes proativas)
- [ ] Roteirizador inteligente
- [ ] Integracao com balancas
- [ ] App mobile
- [ ] Internacionalizacao (i18n)

## Arquitetura
```
/app/
├── backend/server.py           # Python proxy (FastAPI → Go binary)
├── cmd/api/main.go             # Go entrypoint, in-memory DI
├── frontend/                   # Next.js 14
├── internal/
│   ├── app/wire.go             # Dependency injection
│   ├── auth/                   # JWT auth handler
│   ├── handlers/               # API route handlers
│   ├── repository/memory/      # In-memory repos + seed data
│   └── tax/engine.go           # Motor Fiscal IBS/CBS 2026
└── test_reports/               # Test reports
```
