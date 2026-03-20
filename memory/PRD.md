# Nexo One ERP — Product Requirements Document

## Visao
ERP SaaS multi-nicho adaptativo ("Camaleao") para Reforma Tributaria Brasil 2026 (IBS/CBS).
Do TOTVS ao cafezinho — gestao para todos. Metade do preco, mesma entrega.

## Nicho alvo
Industria, Logistica, Mecanica, Estetica, Padaria, Calcados

## Tech Stack
- **Backend:** Go 1.22, Clean Architecture, HTTP Router
- **Frontend:** Next.js 14, React, TypeScript, Tailwind CSS
- **Database (preview):** In-memory repositories (volatile, seeded on startup)
- **Database (producao):** PostgreSQL com Row Level Security (RLS)

## Credenciais demo
- Tenant: `demo` | Email: `admin@demo.com` | Senha: `demo123`
- Cupons: `NEXO20` (20% off), `PRIMEIRO` (100% primeiro mes)

## O que esta implementado

### Fase 1 — Motor Fiscal + Frontend Integration (Completa)
- [x] Motor Fiscal IBS/CBS 2026 completo
- [x] Auth JWT, Multi-tenancy
- [x] Dashboard, Mecanica, Padaria, Estetica, Logistica, IA conectados ao backend
- [x] Simulador Fiscal publico
- [x] Redirect automatico sem auth

### Fase 2 — Self-Service & Onboarding (Completa)
- [x] 4 planos (Starter R$497, Pro R$997, Business R$1.997, Enterprise R$2.997)
- [x] Setup gratis nos 2 primeiros planos
- [x] Trial 7 dias, conversao, upgrade/downgrade
- [x] Cupons com desconto
- [x] Trial com verificacao WhatsApp (MOCKED)
- [x] Onboarding por nicho (mecanica 5 etapas, padaria 3, estetica 3)
- [x] Journey tracking + funnel analytics
- [x] Pricing page, Subscription page, Onboarding page

### Fase 2.5 — Admin Plans + Fase 3 QR Scanner (Completa)
- [x] **Admin Plans** (/admin/plans): Editor de planos com campos editaveis (nome, preco mensal/anual, setup, limites, nichos, features)
- [x] **Partial update** seguro — campos nao enviados nao sao resetados
- [x] **QR Code NFC-e Scanner:** Parse de QR codes de NFC-e, NF-e, SAT, CT-e
- [x] **Expenses CRUD:** 5 despesas demo seeded, lista com filtros, detalhe com itens
- [x] **8 categorias** de despesa com auto-categorizacao por NCM
- [x] **Resumo fiscal:** Creditos IBS/CBS por categoria
- [x] **Tax report:** Relatorio para IR separando dedutiveis/nao dedutiveis
- [x] **SEFAZ scraper** real (web scraping) — funcional mas SEFAZ externa indisponivel no ambiente
- [x] Frontend: Expenses list, Scanner page, Admin Plans
- [x] Sidebar com Despesas e Gestao de Planos
- [x] 25/25 testes backend + 100% frontend (iteracao 5)

## Backlog Priorizado

### P1 — Emissao NF-e/CT-e
- [ ] Emissao de NF-e/CT-e via SEFAZ (integracao web services)

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
│   ├── expenses/               # QR scanner + SEFAZ scraper
│   ├── trial/service.go        # Trial & WhatsApp verification
│   ├── journey/service.go      # Journey tracking & onboarding
│   ├── handlers/               # API handlers
│   ├── repository/memory/      # In-memory repos + seed data
│   └── tax/engine.go           # Motor Fiscal IBS/CBS 2026
└── test_reports/               # Test reports (iterations 1-5)
```
