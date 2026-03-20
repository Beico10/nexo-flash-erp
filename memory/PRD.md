# Nexo One ERP — Product Requirements Document

## Visao
ERP SaaS multi-nicho adaptativo ("Camaleao") para Reforma Tributaria Brasil 2026 (IBS/CBS).
Do TOTVS ao cafezinho — gestao para todos.

## Nicho alvo
Industria, Logistica, Mecanica, Estetica, Padaria, Calcados

## Tech Stack
- **Backend:** Go 1.22, Clean Architecture, HTTP Router
- **Frontend:** Next.js 14, React, TypeScript, Tailwind CSS, Recharts
- **Database (preview):** In-memory repositories
- **Database (producao):** PostgreSQL com Row Level Security (RLS)
- **Cache (producao):** Redis
- **Event Bus (producao):** NATS

## O que esta implementado

### 2026-03-19 — Motor Fiscal + Backend Go Funcional
- [x] Go 1.22 compilando no ambiente ARM64
- [x] Backend Go rodando na porta 8002 + proxy FastAPI 8001
- [x] **Motor Fiscal IBS/CBS 2026 completo:**
  - IBS (9.25%) + CBS (3.75%) por NCM
  - Cesta Basica Nacional (aliquota zero)
  - Cesta Estendida (reducao 60%)
  - Imposto Seletivo (bebidas, tabaco)
  - Fator de transicao 2026 (10%)
  - Cashback tributario
  - Base legal, approval PENDING
  - 10/10 testes Go
- [x] Auth JWT (login, refresh, logout)
- [x] Multi-tenancy
- [x] Modulo Mecanica CRUD

### 2026-03-20 — Simulador Fiscal + Integracao Frontend
- [x] **Simulador Fiscal publico** (/simulador-fiscal)
  - Pagina publica sem auth necessario
  - Dropdown NCM com busca
  - Calculo em tempo real IBS/CBS
  - Badges Cesta Basica / Seletivo
  - Barra de composicao tributaria
  - Base legal por calculo
  - Cashback credito/debito
- [x] **Mecanica conectada ao backend real**
  - Criar OS via modal → backend Go
  - Listar OS abertas do backend
  - Stats em tempo real
  - Filtro por status
  - Busca por placa/cliente
- [x] Sidebar com link Simulador Fiscal
- [x] Layout condicional (simulador sem sidebar)
- [x] 16/16 testes passando (iteracao 2)

### Credenciais demo
- Tenant: `demo` | Email: `admin@demo.com` | Senha: `demo123`

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
| POST | /api/v1/mechanic/os | S | Criar OS |
| GET | /api/v1/mechanic/os | S | Listar OS abertas |
| GET | /api/v1/mechanic/os/{id} | S | Buscar OS |
| PATCH | /api/v1/mechanic/os/{id}/status | S | Atualizar status |
| GET | /api/v1/mechanic/os/plate/{plate} | S | Buscar por placa |

## Backlog Priorizado

### P0 — Proximo
- [ ] Conectar Dashboard a dados reais (KPIs do backend)
- [ ] Conectar outros modulos ao backend (Padaria, Estetica)

### P1 — Medio prazo
- [ ] Sistema de Billing/Assinaturas
- [ ] Free trial com verificacao WhatsApp
- [ ] Scanner QR Code NFC-e para despesas
- [ ] Emissao de NF-e/CT-e via SEFAZ

### P2 — Futuro
- [ ] IA Co-Piloto
- [ ] Roteirizador inteligente
- [ ] Integracao com balancas
- [ ] App mobile
