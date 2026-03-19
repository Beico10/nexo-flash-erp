# Nexo One ERP

**Do TOTVS Protheus ao vendedor de cafezinho — um sistema que atende todos.**

ERP SaaS Multi-Tenant, Multi-Nicho — Reforma Tributária Brasil 2026 (IBS/CBS).

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://postgresql.org)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)]()

---

## 📋 Visão Geral

O **Nexo One** é um ERP modular que atende **6 nichos de negócio** com a mesma base de código:

| Nicho | Funcionalidades Principais |
|-------|---------------------------|
| **Mecânica** | OS digital, peças, aprovação via WhatsApp |
| **Padaria** | PDV rápido, balanças, gestão de perdas, cesta básica |
| **Indústria** | PCP, BOM (ficha técnica), gestão de insumos |
| **Logística** | CT-e, contratos multi-cliente, DRE da viagem |
| **Estética** | Agenda com trava de conflito, split de pagamento |
| **Calçados** | Matriz de grade Cor/Tamanho/SKU, comissões |

---

## 🔒 Princípios de Segurança

1. **RLS Obrigatório** — Row Level Security em 100% das tabelas de negócio
2. **Human-in-the-Loop** — IA nunca altera dados sem aprovação humana
3. **Auditoria Imutável** — Todas alterações financeiras são logadas
4. **JWT com Rotação** — Access token 15min + Refresh token 7 dias
5. **Zero Trust** — Cada request é validado independentemente

---

## 🚀 Quick Start (Desenvolvimento)

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) e Docker Compose
- [Node.js 20+](https://nodejs.org/) (para frontend)
- [Yarn](https://yarnpkg.com/)

### 1. Clone e configure

```bash
git clone https://github.com/Beico10/nexo-one-erp.git
cd nexo-one-erp

# Copie o arquivo de exemplo e configure
cp .env.example .env
# Edite .env com suas configurações
```

### 2. Inicie os serviços

```bash
# Opção A: Docker Compose (recomendado)
docker compose up -d

# Opção B: Serviços separados
docker compose up -d postgres redis nats
./scripts/check-build.sh
./build/nexo-one
```

### 3. Acesse

- **Frontend:** http://localhost:3000
- **API:** http://localhost:8080
- **NATS Monitor:** http://localhost:8222

### Credenciais de Teste

| Tenant | Email | Senha |
|--------|-------|-------|
| mecanica-demo | admin@mecanica.demo | nexo@2026 |
| padaria-demo | admin@padaria.demo | nexo@2026 |
| industria-demo | admin@industria.demo | nexo@2026 |
| logistica-demo | admin@logistica.demo | nexo@2026 |
| estetica-demo | admin@estetica.demo | nexo@2026 |
| calcados-demo | admin@calcados.demo | nexo@2026 |

---

## 📁 Estrutura do Projeto

```
nexo-one-erp/
├── cmd/api/                  # Entry point da API
├── internal/
│   ├── ai/                   # IA Concierge + Gateway (Human-in-the-Loop)
│   ├── auth/                 # JWT + Refresh Token
│   ├── baas/                 # Banking as a Service (PIX, Boleto)
│   ├── core/                 # Module Registry
│   ├── handlers/             # HTTP Handlers
│   ├── modules/              # Módulos de negócio
│   │   ├── aesthetics/       # Estética/Salão
│   │   ├── bakery/           # Padaria
│   │   ├── industry/         # Indústria
│   │   ├── logistics/        # Transportadora
│   │   ├── mechanic/         # Mecânica
│   │   └── shoes/            # Calçados
│   ├── repository/postgres/  # Repositórios com RLS
│   └── tax/                  # Motor Fiscal IBS/CBS 2026
├── pkg/
│   ├── cache/                # Redis client
│   ├── eventbus/             # NATS JetStream
│   └── middleware/           # Auth + Tenant Middleware
├── migrations/               # PostgreSQL migrations
├── frontend/                 # Next.js 14
├── scripts/                  # Scripts utilitários
├── Dockerfile                # Build multi-stage otimizado
├── docker-compose.yml        # Ambiente completo
└── .env.example              # Template de configuração
```

---

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENTES                                  │
│              (Browser, Mobile, Integrações)                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     FRONTEND (Next.js 14)                       │
│                     http://localhost:3000                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      BACKEND API (Go)                           │
│                    http://localhost:8080                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Middleware: Auth → TenantDB (SET LOCAL tenant_id)      │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌───────────┬───────────┬───────────┬───────────┬─────────┐   │
│  │ Mechanic  │  Bakery   │ Logistics │ Aesthetics│  Tax    │   │
│  │  Module   │  Module   │  Module   │  Module   │ Engine  │   │
│  └───────────┴───────────┴───────────┴───────────┴─────────┘   │
└─────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  PostgreSQL 16  │  │    Redis 7      │  │  NATS JetStream │
│    + RLS        │  │ (cache/sessões) │  │   (Event Bus)   │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## 🧪 Testes

```bash
# Backend
go test ./... -v -cover

# Frontend
cd frontend && yarn test

# Verificação completa
./scripts/check-build.sh
```

---

## 🔧 Scripts Úteis

| Script | Descrição |
|--------|-----------|
| `./scripts/check-build.sh` | Verifica e compila tudo |
| `./scripts/check-build.sh --fast` | Apenas go vet (rápido) |
| `docker compose up -d` | Inicia ambiente completo |
| `docker compose logs -f api` | Ver logs do backend |
| `docker compose down -v` | Para e remove volumes |

---

## 📊 Motor Fiscal Brasil 2026

O Nexo One implementa o motor fiscal conforme a **Reforma Tributária (LC 214/2025)**:

- **IBS** (Imposto sobre Bens e Serviços) — Estadual/Municipal
- **CBS** (Contribuição sobre Bens e Serviços) — Federal
- **Cashback Tributário** — Crédito na entrada / Débito na saída
- **Cesta Básica Nacional** — Alíquota zero para itens essenciais
- **Imposto Seletivo** — Aplicado a bebidas alcoólicas, tabaco, etc.

```go
// Exemplo de cálculo
result := taxEngine.Calculate(ctx, TaxInput{
    NCM:       "19052000", // pão francês
    BaseValue: 100.00,
    TenantTaxRegime: "simples_nacional",
})
// result.IsZeroRated = true (Cesta Básica)
```

---

## 🛡️ Segurança RLS

Toda query de negócio passa pelo middleware que configura o `tenant_id`:

```sql
-- Automático em cada conexão
SET LOCAL nexo.current_tenant_id = '<uuid>';

-- Policy RLS (todas as tabelas)
CREATE POLICY isolation ON tabela
  USING (tenant_id = current_tenant_id());
```

**Resultado:** Mesmo com bug na API, um tenant NUNCA acessa dados de outro.

---

## 📈 Roadmap

- [x] Multi-tenant com RLS
- [x] Motor fiscal IBS/CBS 2026
- [x] Módulos: Mecânica, Padaria, Indústria, Logística, Estética, Calçados
- [x] IA com Human-in-the-Loop
- [x] BaaS (PIX, Boleto, Split)
- [ ] Integração SEFAZ (NF-e, NFC-e, CT-e)
- [ ] Roteirizador inteligente
- [ ] Integração com balanças
- [ ] App mobile

---

## 📄 Licença

Proprietário — © 2026 Nexo One. Todos os direitos reservados.

---

## 🤝 Suporte

- **Email:** suporte@nexoone.com.br
- **Documentação:** https://docs.nexoone.com.br
