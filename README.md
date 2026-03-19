# Nexo Flash ERP — Fundação v1.0

ERP SaaS Multi-Tenant, Multi-Nicho, focado na **Reforma Tributária Brasil 2026 (IBS/CBS)**.

## Arquitetura rápida

```
internal/core/module_registry.go   → micro-kernel: contratos e registro de módulos
internal/tax/engine.go             → Motor Fiscal IBS/CBS 2026 (NCM → alíquota → cashback)
internal/ai/gateway.go             → Human-in-the-Loop: toda ação de IA passa aqui
internal/modules/logistics/        → CT-e, contratos multi-cliente, DRE da viagem
pkg/eventbus/bus.go                → NATS JetStream: desacoplamento entre módulos
migrations/001_foundation_rls.sql  → PostgreSQL RLS obrigatório
deployments/docker/docker-compose.yml
Dockerfile                         → imagem scratch ~10MB, usuário não-root
```

## Setup rápido

```bash
cp .env.example .env   # preencha as variáveis
docker compose -f deployments/docker/docker-compose.yml up -d
```

## Para commitar no GitHub

```bash
cd nexo-flash
git init
git add .
git commit -m "feat: nexo flash foundation v1.0"
git remote add origin https://github.com/SEU_USUARIO/nexo-flash.git
git push -u origin main
```
