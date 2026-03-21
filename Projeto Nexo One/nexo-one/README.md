# Nexo One ERP

Sistema ERP SaaS Multi-Tenant para pequenos negócios brasileiros.
Motor fiscal IBS/CBS 2026 integrado.

## Modos de Operação

### Desenvolvimento (in-memory — sem banco)
```bash
# Roda sem PostgreSQL. Dados em memória, reinicia limpos.
go run ./cmd/api
```
Credenciais demo: `tenant=demo | email=admin@demo.com | senha=demo123`

### Produção (PostgreSQL + Redis + RLS)
```bash
# 1. Configurar variáveis
export DATABASE_URL="postgres://user:senha@host:5432/nexoone?sslmode=require"
export REDIS_URL="redis://:senha@host:6379/0"
export JWT_SECRET="seu-segredo-aqui"

# 2. Rodar migrations
psql $DATABASE_URL -f migrations/001_foundation_rls.sql
psql $DATABASE_URL -f migrations/002_business_modules.sql
psql $DATABASE_URL -f migrations/003_baas_ai_panel.sql
psql $DATABASE_URL -f migrations/004_cte_indexes_audit.sql
psql $DATABASE_URL -f migrations/005_seed_dev.sql
psql $DATABASE_URL -f migrations/006_billing_plans.sql
psql $DATABASE_URL -f migrations/007_trial_journey.sql
psql $DATABASE_URL -f migrations/008_expenses_qrcode.sql
psql $DATABASE_URL -f migrations/009_billing_trial_journey.sql

# 3. Trocar wire.go para produção
cp internal/app/wire_prod.go internal/app/wire_prod_active.go
# Editar main.go: trocar app.Wire() por app.WireProd()

# 4. Compilar e rodar
go build -o nexo-one ./cmd/api
./nexo-one
```

## Stack

| Componente | Desenvolvimento | Produção |
|---|---|---|
| Banco | In-memory (sem configuração) | PostgreSQL 16 + RLS |
| Cache/Auth | In-memory | Redis 7 |
| Event Bus | — | NATS JetStream |

## Módulos

- **Mecânica** — OS digital, WhatsApp, aprovação de orçamento
- **Padaria** — PDV, balança Toledo/Elgin, gestão de perdas
- **Indústria** — PCP, ficha técnica (BOM), ordens de produção
- **Logística** — contratos de frete, DRE da viagem, CT-e
- **Estética** — agenda com trava de conflito, split de pagamento
- **Calçados** — matriz de grade Cor/Tamanho/SKU, comissões

## Segurança

- **RLS** em 100% das tabelas — dados de tenants completamente isolados
- **JWT HS256** (15 min) + Refresh Token Redis (7 dias, rotação automática)
- **bcrypt cost=12** para senhas
- **Audit log** imutável via trigger

## Repositório

`github.com/Beico10/nexo-flash-erp`
