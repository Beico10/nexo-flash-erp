# Nexo One ERP — Product Requirements Document

## Visao
ERP SaaS multi-nicho adaptativo ("Camaleao") para Reforma Tributaria Brasil 2026 (IBS/CBS).
Do TOTVS ao cafezinho — gestao para todos. Metade do preco, mesma entrega.

## Nicho alvo
Industria, Logistica, Mecanica, Estetica, Padaria, Calcados

## Tech Stack (ATUALIZADO - 20/03/2026)
- **Backend:** Go 1.22, Clean Architecture, HTTP Router (100% Go)
- **Frontend:** HTMX + html/template + Tailwind CSS via CDN
- **Database (preview):** In-memory repositories (volatile, seeded on startup)
- **Database (producao):** PostgreSQL com Row Level Security (RLS)
- **AI:** Gemini 3 Flash via emergentintegrations (Emergent LLM Key) - micro-servico Python

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
- [x] **Admin Plans** (/admin/plans): Editor de planos com campos editaveis
- [x] **Partial update** seguro — campos nao enviados nao sao resetados
- [x] **QR Code NFC-e Scanner:** Parse de QR codes de NFC-e, NF-e, SAT, CT-e
- [x] **Expenses CRUD:** 5 despesas demo seeded, lista com filtros, detalhe com itens
- [x] **8 categorias** de despesa com auto-categorizacao por NCM
- [x] **Resumo fiscal:** Creditos IBS/CBS por categoria
- [x] **Tax report:** Relatorio para IR separando dedutiveis/nao dedutiveis
- [x] **SEFAZ scraper** real (web scraping) — funcional mas SEFAZ externa indisponivel no ambiente
- [x] Frontend: Expenses list, Scanner page, Admin Plans
- [x] Sidebar com Despesas e Gestao de Planos

### Fase 4 — NF-e, Co-Piloto IA, Industria PCP, Calcados (Completa - 20/03/2026)
- [x] **Emissao NF-e/NFC-e/CT-e:** CRUD completo com emissao e cancelamento (homologacao)
- [x] 4 documentos demo seeded (NF-e, NFC-e, CT-e)
- [x] Chave de acesso gerada automaticamente
- [x] **IA Co-Piloto:** Chat com Gemini 3 Flash para sugestoes de negocio
- [x] Session-based conversation, sugestoes rapidas
- [x] **Industria PCP:** Ordens de producao, fichas tecnicas (BOM), estoque de materiais
- [x] Explosao de BOM com calculo de insumos e custos
- [x] 3 OPs demo + 2 BOMs com componentes detalhados + 7 materiais
- [x] **Calcados Grade:** Matriz Cor x Tamanho x SKU, comissoes de vendedores
- [x] 3 grades demo (Sandalia, Bota, Tenis) com estoque por SKU
- [x] 2 vendedores com regras de comissao e calculo automatico
- [x] Sidebar atualizada com links para Emissao NF-e e Co-Piloto IA
- [x] JSON tags snake_case em todos os structs Go
- [x] 17/17 testes backend + 100% frontend (iteracao 6)

### Fase 5 — Migracao para Go Puro (EM ANDAMENTO - 20/03/2026)
- [x] **Backend 100% Go:** Eliminado proxy Python complexo
- [x] **Templates Go:** html/template + HTMX para paginas web
- [x] **Tailwind CSS via CDN:** Estilizacao responsiva
- [x] **Paginas migradas:** Login, Dashboard (com grafico e KPIs)
- [x] **Co-Piloto IA em Go:** Cliente HTTP chamando micro-servico Python
- [x] **API funcional:** Health, Auth, Dashboard, Copilot (4/4 testes passando)
- [ ] **Migrar demais modulos:** Mecanica, Padaria, Estetica, etc.
- [ ] **Eliminar micro-servico Python:** Integrar Gemini diretamente em Go

## Backlog Priorizado

### P0 — Atual
- [ ] Migrar paginas restantes para templates Go (mecanica, padaria, etc.)
- [ ] Configurar ingress para servir frontend via porta 8001

### P1 — Proximo
- [ ] Roteirizador inteligente (integracao OSRM)
- [ ] Integracao com balancas (Padaria/Industria)
- [ ] Eliminar micro-servico Python de IA

### P2 — Futuro
- [ ] App mobile
- [ ] Internacionalizacao (i18n)

## Arquitetura (ATUALIZADA)
```
/app/
├── nexo-one                    # Binario Go compilado
├── cmd/api/main.go             # Go entrypoint, routes, DI
├── templates/                  # Templates Go (html/template)
│   ├── layouts/base.html       # Layout base com HTMX/Tailwind
│   ├── pages/login.html        # Pagina de login
│   ├── pages/dashboard.html    # Dashboard com KPIs
│   ├── pages/copilot.html      # Chat Co-Piloto IA
│   └── partials/               # Sidebar, header, etc.
├── internal/
│   ├── web/                    # Handlers de paginas web
│   ├── gemini/client.go        # Cliente HTTP para servico de IA
│   ├── app/wire.go             # Dependency injection
│   ├── auth/                   # JWT auth
│   ├── billing/                # Billing & subscriptions
│   ├── expenses/               # QR scanner + SEFAZ scraper
│   ├── trial/                  # Trial & WhatsApp verification
│   ├── journey/                # Journey tracking & onboarding
│   ├── handlers/               # API handlers
│   ├── modules/                # Modulos de negocio
│   ├── repository/memory/      # In-memory repos + seed data
│   └── tax/engine.go           # Motor Fiscal IBS/CBS 2026
├── backend/
│   ├── server.py               # Proxy Python (temporario)
│   └── ai_service.py           # Micro-servico de IA (Gemini)
└── test_reports/               # Test reports (iterations 1-7)
```

## Changelog
- **20/03/2026 (Fase 5):** Iniciada migracao para Go puro. Login e Dashboard funcionando com templates Go + HTMX. Co-Piloto IA operacional.
- **20/03/2026 (Fase 4):** NF-e, Co-Piloto IA, Industria PCP, Calcados implementados.
