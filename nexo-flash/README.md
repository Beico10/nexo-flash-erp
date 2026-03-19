# Nexo Flash ERP

ERP SaaS Multi-Tenant, Multi-Nicho — Reforma Tributária Brasil 2026 (IBS/CBS).

## Deploy Hetzner — 3 passos

```bash
# 1. Acessar o servidor
ssh root@IP_DO_SERVIDOR

# 2. Configurar e rodar
export GITHUB_REPO="https://github.com/Beico10/nexo-flash-erp.git"
export DOMAIN="api.nexoflash.com.br"
export FRONTEND_DOMAIN="app.nexoflash.com.br"
export LETSENCRYPT_EMAIL="seu@email.com.br"
bash deployments/scripts/deploy.sh

# 3. Pronto — sistema no ar com HTTPS, backup e firewall
```

## Atualizar após novos commits

```bash
bash /opt/nexo-flash/deployments/scripts/update.sh
```

## Stack

Go 1.22 · PostgreSQL 16 + RLS · Redis 7 · NATS JetStream · Next.js 14 · Hetzner VPS

## Módulos

- **Mecânica** — OS digital, peças, aprovação WhatsApp
- **Padaria** — PDV rápido, balanças, gestão de perdas  
- **Indústria** — PCP, BOM (ficha técnica), insumos
- **Logística** — CT-e, contratos multi-cliente, DRE da viagem
- **Estética** — Agenda com trava de conflito, split de pagamento
- **Calçados** — Matriz de grade Cor/Tamanho/SKU, comissões

## Segurança

- RLS PostgreSQL ativo em 100% das tabelas de negócio
- JWT HS256 (15min) + Refresh Token Redis (7 dias, rotação automática)
- IA: toda sugestão → status=pending → aprovação humana obrigatória
- Auditoria imutável via trigger em tabelas financeiras
- Firewall UFW + Fail2ban + SSL Let's Encrypt automático
