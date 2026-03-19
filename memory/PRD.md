# Nexo One ERP — Product Requirements Document

## Visao
ERP SaaS multi-nicho adaptativo ("Camaleao") para Reforma Tributaria Brasil 2026 (IBS/CBS).
Do TOTVS ao cafezinho — gestao para todos.

## Nicho alvo
Industria, Logistica, Mecanica, Estetica, Padaria, Calcados

## Tech Stack
- **Backend:** Go 1.22, Clean Architecture, HTTP Router (Go 1.22 pattern matching)
- **Frontend:** Next.js 14, React, TypeScript, Tailwind CSS, Recharts
- **Database (preview):** In-memory repositories
- **Database (producao):** PostgreSQL com Row Level Security (RLS)
- **Cache (producao):** Redis
- **Event Bus (producao):** NATS
- **Deploy:** Docker, Hetzner (producao)

## Arquitetura
```
/app/
├── cmd/api/main.go              # Go entrypoint (porta 8002)
├── backend/server.py            # Proxy FastAPI (porta 8001) → Go (8002)
├── internal/
│   ├── tax/engine.go            # Motor Fiscal IBS/CBS 2026
│   ├── auth/                    # JWT authentication
│   ├── handlers/                # HTTP handlers (API routes)
│   ├── app/wire.go              # Dependency injection
│   ├── repository/memory/       # In-memory repos (preview)
│   ├── repository/postgres/     # PostgreSQL repos (producao)
│   ├── modules/                 # Logica de negocio por nicho
│   ├── ai/                      # IA Concierge & Sugestoes
│   ├── baas/                    # Banking as a Service
│   ├── billing/                 # Planos e assinaturas
│   ├── expenses/                # Scanner QR Code NFC-e
│   ├── trial/                   # Free trial management
│   └── journey/                 # User journey tracking
├── frontend/                    # Next.js 14
└── migrations/                  # PostgreSQL migrations
```

## O que esta implementado

### 2026-03-19 — Motor Fiscal + Backend Go Funcional
- [x] Compilador Go 1.22 instalado no ambiente
- [x] Backend Go compilando e rodando na porta 8002
- [x] Proxy Python/FastAPI na porta 8001 redirecionando para Go
- [x] **Motor Fiscal IBS/CBS 2026 funcional:**
  - Calculo IBS (9.25%) + CBS (3.75%) por NCM
  - Cesta Basica Nacional (aliquota zero) - Art. 8 LC 214/2025
  - Cesta Estendida (reducao 60%) - Art. 9 LC 214/2025
  - Imposto Seletivo (bebidas, tabaco)
  - Fator de transicao (2026 = 10% da aliquota plena)
  - Cashback tributario (credito/debito)
  - Validacao NCM (8 digitos)
  - Approval status = PENDING (human-in-the-loop)
  - Base legal em cada resultado
  - 10/10 testes Go passando
- [x] Autenticacao JWT (login, refresh, logout)
- [x] Multi-tenancy com isolamento por tenant_id
- [x] Modulo Mecanica CRUD (criar OS, listar, buscar por placa, atualizar status)
- [x] Modulo Padaria (PDV, vendas, perdas, produtos)
- [x] Modulo Estetica (agendamentos, conflitos, split)
- [x] Modulo Logistica (contratos, frete)
- [x] Modulo IA (sugestoes, aprovacao human-in-the-loop)
- [x] Modulo Pagamentos (PIX, Boleto)
- [x] Frontend login 2-step (tenant slug → credenciais)
- [x] Dashboard com KPIs, graficos, tabela OS, painel IA
- [x] Layout condicional (login sem sidebar, dashboard com sidebar)
- [x] 18/18 testes backend + frontend passando

### Credenciais demo
- Tenant: `demo`
- Email: `admin@demo.com`
- Senha: `demo123`

## API Endpoints
| Metodo | Rota | Descricao |
|--------|------|-----------|
| GET | /api/health | Health check |
| POST | /api/auth/login | Login (tenant_slug, email, password) |
| POST | /api/auth/refresh | Refresh token |
| POST | /api/auth/logout | Logout |
| GET | /api/auth/me | User info |
| POST | /api/v1/tax/calculate | Calculo fiscal IBS/CBS |
| POST | /api/v1/mechanic/os | Criar OS |
| GET | /api/v1/mechanic/os | Listar OS abertas |
| GET | /api/v1/mechanic/os/{id} | Buscar OS |
| PATCH | /api/v1/mechanic/os/{id}/status | Atualizar status |
| POST | /api/v1/mechanic/os/{id}/send-approval | Enviar aprovacao WhatsApp |
| POST | /api/v1/mechanic/approve/{token} | Aprovar por token |
| GET | /api/v1/mechanic/os/plate/{plate} | Buscar por placa |

## Backlog Priorizado

### P0 — Proximo
- [ ] Conectar frontend (mecanica, padaria, etc) ao backend real
- [ ] Pagina de Simulador Fiscal no frontend (demonstrar motor IBS/CBS)

### P1 — Medio prazo
- [ ] Sistema de Billing/Assinaturas (planos, self-service)
- [ ] Free trial com verificacao WhatsApp
- [ ] User journey tracking
- [ ] Scanner QR Code NFC-e para despesas
- [ ] Emissao de NF-e/CT-e via SEFAZ

### P2 — Futuro
- [ ] IA Co-Piloto (sugestoes operacionais)
- [ ] Roteirizador inteligente (OSRM)
- [ ] Integracao com balancas fisicas
- [ ] Internacionalizacao (i18n)
- [ ] App mobile
