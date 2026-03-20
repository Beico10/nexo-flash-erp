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
- Cupons: `NEXO20` (20% off), `PRIMEIRO` (100% primeiro mes)

## O que esta implementado

### 2026-03-19 — Motor Fiscal + Backend Go Funcional
- [x] Go 1.22 compilando no ambiente ARM64
- [x] Backend Go rodando na porta 8002 + proxy FastAPI 8001
- [x] Motor Fiscal IBS/CBS 2026 completo
- [x] Auth JWT (login, refresh, logout)
- [x] Multi-tenancy
- [x] Modulo Mecanica CRUD

### 2026-03-20 — Simulador Fiscal + Integracao Frontend (Fase 1 Completa)
- [x] Simulador Fiscal publico (/simulador-fiscal)
- [x] Dashboard conectado ao backend (KPIs, graficos)
- [x] Todos os modulos conectados (Mecanica, Padaria, Estetica, Logistica, IA)
- [x] Sidebar com navegacao completa
- [x] Auth.ts corrigido (URLs relativas)
- [x] Redirect automatico para /login sem auth
- [x] 29/29 testes backend passando (iteracao 3)

### 2026-03-20 — Fase 2: Self-Service & Onboarding (Completa)
- [x] **Billing System:** 5 planos (Micro R$49, Starter R$99, Pro R$199, Business R$399, Enterprise R$999)
- [x] **Subscription Management:** Trial 7 dias, conversao, upgrade/downgrade, verificacao de limites
- [x] **Coupon System:** Validacao de cupons com desconto percentual
- [x] **Trial WhatsApp Verification:** Gera codigo 6 digitos + URL wa.me (MOCKED - sem integracao real)
- [x] **Anti-abuso:** Device fingerprint, rate limiting, phone hash (SHA256)
- [x] **Onboarding por nicho:** 5 etapas para mecanica, 3 para padaria, 3 para estetica
- [x] **Journey Tracking:** Eventos de funil, page views, conversao
- [x] **Funnel Analytics:** Metricas diarias de conversao e drop points
- [x] **Pricing Page (/pricing):** Toggle mensal/anual, cupom, 5 cards de planos
- [x] **Subscription Page (/dashboard/subscription):** Status trial, uso do plano, botoes de acao
- [x] **Onboarding Page (/onboarding):** Progresso visual, completar/pular etapas, recompensas
- [x] **Sidebar atualizada:** Links "Minha Assinatura" e "Onboarding"
- [x] 24/24 testes backend + 100% frontend passando (iteracao 4)

## API Endpoints
| Metodo | Rota | Auth | Descricao |
|--------|------|------|-----------|
| GET | /api/health | N | Health check |
| POST | /api/auth/login | N | Login |
| POST | /api/auth/refresh | S | Refresh token |
| POST | /api/auth/logout | S | Logout |
| GET | /api/auth/me | S | User info |
| GET | /api/v1/tax/ncm-list | N | Lista NCMs |
| POST | /api/v1/tax/simulate | N | Simulador fiscal |
| POST | /api/v1/tax/calculate | S | Calculo fiscal |
| GET | /api/v1/dashboard/stats | S | Dashboard KPIs |
| POST/GET | /api/v1/mechanic/os | S | CRUD OS mecanica |
| GET | /api/v1/bakery/products | S | Lista produtos padaria |
| GET/POST | /api/v1/aesthetics/appointments | S | Agendamentos estetica |
| GET | /api/v1/ai/suggestions | S | Sugestoes IA |
| POST | /api/v1/logistics/freight/calculate | S | Calculo frete |
| POST | /api/v1/payments/pix | S | Gerar PIX |
| POST | /api/v1/payments/boleto | S | Gerar boleto |
| GET | /api/billing/plans | N | Lista planos |
| POST | /api/billing/coupon/validate | N | Valida cupom |
| GET | /api/v1/billing/subscription | S | Assinatura atual |
| POST | /api/v1/billing/convert | S | Converter trial |
| POST | /api/v1/billing/change-plan | S | Mudar plano |
| GET | /api/v1/billing/usage | S | Uso do plano |
| GET | /api/v1/billing/feature | S | Verificar feature |
| POST | /api/auth/verify/start | N | Iniciar verificacao WhatsApp |
| POST | /api/auth/verify/confirm | N | Confirmar codigo |
| POST | /api/webhooks/whatsapp | N | Webhook WhatsApp |
| GET | /api/v1/onboarding/steps | S | Passos do onboarding |
| GET | /api/v1/onboarding/progress | S | Progresso onboarding |
| POST | /api/v1/onboarding/complete | S | Completar passo |
| POST | /api/v1/onboarding/skip | S | Pular onboarding |
| POST | /api/v1/track | S | Track eventos |
| GET | /api/v1/analytics/funnel | S | Metricas funil |
| GET | /api/v1/analytics/drops | S | Drop points |

## Backlog Priorizado

### P1 — Fase 3: Features de Receita
- [ ] Scanner QR Code NFC-e para despesas
- [ ] Emissao de NF-e/CT-e via SEFAZ

### P2 — Futuro
- [ ] IA Co-Piloto (sugestoes proativas com LLM)
- [ ] Roteirizador inteligente
- [ ] Integracao com balancas
- [ ] App mobile
- [ ] Internacionalizacao (i18n)
- [ ] Modulos Industria e Calcados com CRUD real

## Arquitetura
```
/app/
├── backend/server.py           # Python proxy (FastAPI -> Go binary)
├── cmd/api/main.go             # Go entrypoint, routes, DI
├── frontend/                   # Next.js 14
├── internal/
│   ├── app/wire.go             # Dependency injection
│   ├── auth/                   # JWT auth
│   ├── billing/service.go      # Billing & subscriptions
│   ├── trial/service.go        # Trial & WhatsApp verification
│   ├── journey/service.go      # Journey tracking & onboarding
│   ├── handlers/               # API handlers
│   ├── repository/memory/      # In-memory repos + seed data
│   └── tax/engine.go           # Motor Fiscal IBS/CBS 2026
└── test_reports/               # Test reports (iterations 1-4)
```
