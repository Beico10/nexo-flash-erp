# Nexo One ERP — PRD (Product Requirements Document)

## 1. Visão Geral

**Nome:** Nexo One ERP  
**Versão:** 1.0.0  
**Data:** Janeiro 2026

### Propósito
ERP SaaS Multi-Tenant que atende desde **indústrias** (nível TOTVS Protheus) até **micro-empreendedores** (manicure em casa), com foco na Reforma Tributária Brasil 2026 (IBS/CBS).

### Diretrizes Fundamentais
1. **Máxima Segurança** — RLS obrigatório, IA com aprovação humana, auditoria imutável
2. **Mínimo Custo** — 100% open source, infraestrutura enxuta, zero licenças

---

## 2. Personas

| Persona | Descrição | Necessidade Principal |
|---------|-----------|----------------------|
| **Indústria** | Fábrica média/grande | PCP, BOM, gestão de insumos, compliance fiscal |
| **Transportadora** | Frota própria/terceirizada | CT-e, contratos multi-cliente, DRE da viagem |
| **Mecânica** | Oficina de bairro | OS digital, peças, aprovação rápida |
| **Padaria** | Comércio de alimentos | PDV rápido, balança, cesta básica zero |
| **Estética** | Salão/autônomo | Agenda sem conflito, split pagamento |
| **Calçados** | Loja varejo | Grade cor/tamanho, comissões |

---

## 3. Requisitos Funcionais

### 3.1 Core (Todos os nichos)
- [x] Multi-tenant com RLS (Row Level Security)
- [x] Autenticação JWT + Refresh Token
- [x] Motor fiscal IBS/CBS 2026
- [x] IA Concierge (Human-in-the-Loop)
- [x] Event Bus (NATS JetStream)
- [x] BaaS (PIX, Boleto, Split)

### 3.2 Por Módulo
| Módulo | Status | Pendências |
|--------|--------|------------|
| Mecânica | ✅ 90% | WhatsApp real |
| Padaria | ✅ 80% | Integração balanças |
| Indústria | ✅ 70% | - |
| Logística | ✅ 75% | Roteirizador, CT-e SEFAZ |
| Estética | ✅ 85% | - |
| Calçados | ✅ 70% | - |

---

## 4. O Que Foi Implementado

### Janeiro 2026
- [x] Estrutura multi-tenant com RLS PostgreSQL
- [x] 6 módulos de negócio funcionais
- [x] Motor fiscal IBS/CBS 2026 com Cesta Básica
- [x] Sistema de aprovação IA (Human-in-the-Loop)
- [x] JWT + Refresh Token com rotação
- [x] BaaS interface (PIX/Boleto)
- [x] Event Bus NATS JetStream
- [x] Frontend Next.js 14
- [x] Docker Compose completo
- [x] Dockerfile multi-stage otimizado
- [x] Migrations corrigidas
- [x] Renomeação para Nexo One

---

## 5. Backlog Priorizado

### P0 (Crítico)
1. [ ] Integração SEFAZ NF-e/NFC-e
2. [ ] CT-e completo com XML assinado
3. [ ] Testes de integração end-to-end

### P1 (Alto)
1. [ ] WhatsApp Business API real
2. [ ] Roteirizador inteligente (OSRM)
3. [ ] Integração com balanças Toledo/Elgin
4. [ ] MDF-e

### P2 (Médio)
1. [ ] Relatórios gerenciais avançados
2. [ ] Dashboard analytics
3. [ ] App mobile (React Native)
4. [ ] Multi-idioma (i18n)

### P3 (Baixo)
1. [ ] Integração contábil
2. [ ] BI embarcado
3. [ ] Marketplace de módulos

---

## 6. Stack Técnica

| Componente | Tecnologia |
|------------|-----------|
| Backend | Go 1.22 (Clean Architecture) |
| Frontend | Next.js 14 + TailwindCSS |
| Banco | PostgreSQL 16 + RLS |
| Cache | Redis 7 |
| Mensageria | NATS JetStream |
| Container | Docker + distroless |
| Infraestrutura | Hetzner VPS |

---

## 7. Métricas de Sucesso

| Métrica | Meta |
|---------|------|
| Tempo de build | < 30s |
| Imagem Docker | < 20MB |
| Tempo de resposta API | < 100ms (p95) |
| Cobertura de testes | > 80% |
| Uptime | 99.9% |

---

## 8. Próximas Ações

1. **Testar compilação Go** em ambiente local
2. **Rodar migrations** em PostgreSQL real
3. **Implementar NF-e** para produção
4. **Criar testes E2E** com Playwright

---

*Última atualização: Janeiro 2026*

---

## 9. Sistema de Planos e Assinaturas (Implementado Jan/2026)

### Modelo de Preços

| Plano | Mensal | Anual | Setup | Público-Alvo |
|-------|--------|-------|-------|--------------|
| **Micro** | R$ 47 | R$ 470 | Grátis | Autônomos, MEI |
| **Starter** | R$ 97 | R$ 970 | Grátis | Pequenos negócios |
| **Pro** | R$ 197 | R$ 1.970 | Grátis | PMEs estabelecidas |
| **Business** | R$ 397 | R$ 3.970 | R$ 297 | Indústrias pequenas |
| **Enterprise** | R$ 997+ | R$ 9.970+ | R$ 497 | Indústrias médias |

### Funcionalidades Self-Service

- ✅ Trial 7 dias automático
- ✅ Conversão sem intervenção humana
- ✅ Upgrade/Downgrade pelo cliente
- ✅ Verificação de limites em tempo real
- ✅ Cupons de desconto
- ✅ Admin Master configura preços (sem código)

### Arquivos Criados

- `/migrations/006_billing_plans.sql` — Tabelas de planos e assinaturas
- `/internal/billing/service.go` — Lógica de negócio
- `/internal/repository/postgres/billing_repo.go` — Persistência
- `/internal/handlers/billing_handler.go` — Endpoints API
- `/frontend/src/app/pricing/page.tsx` — Página de preços
- `/frontend/src/app/dashboard/subscription/page.tsx` — Gerenciar assinatura

### Endpoints API

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/api/billing/plans` | Lista planos (público) |
| POST | `/api/billing/coupon/validate` | Valida cupom |
| GET | `/api/billing/subscription` | Assinatura do tenant |
| POST | `/api/billing/convert` | Converte trial → pago |
| POST | `/api/billing/change-plan` | Muda plano |
| GET | `/api/billing/usage` | Uso atual |
| PUT | `/api/admin/billing/plans/{id}` | Admin atualiza plano |


---

## 10. Sistema de Trial e Jornada do Usuário (Implementado Jan/2026)

### Verificação por WhatsApp (Custo Zero)

```
1. Usuário cadastra com telefone
2. Sistema gera código 6 dígitos (Redis, 5 min TTL)
3. Abre WhatsApp: "Meu código Nexo One: 847293"
4. Webhook recebe → Valida → Trial liberado
5. 1 trial por telefone (hash SHA256 para LGPD)
```

### Anti-Abuso

| Controle | Implementação |
|----------|---------------|
| 1 trial/telefone | Hash SHA256 único |
| Device fingerprint | Detecta múltiplos trials |
| Rate limit | 5 tentativas de código |
| Abuse score | 0-100, bloqueio automático |

### Tracking de Jornada

**Eventos rastreados:**
- page_view, signup_started, signup_completed
- verification_sent, verification_completed
- onboarding_started, onboarding_step_completed
- first_action, trial_converted

**Analytics disponíveis:**
- Funil de conversão diário
- Taxa de conversão por nicho
- Pontos de travamento (drop points)
- Tempo médio de ativação

### Onboarding Adaptativo

| Nicho | Passos | Recompensa |
|-------|--------|------------|
| Mecânica | 5 passos | +3 dias trial |
| Padaria | 5 passos | +3 dias trial |
| Estética | 5 passos | +3 dias trial |
| Logística | 5 passos | +4 dias trial |
| Indústria | 5 passos | +5 dias trial |
| Calçados | 5 passos | +4 dias trial |

### Endpoints Criados

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/auth/verify/start` | Inicia verificação WhatsApp |
| POST | `/api/auth/verify/confirm` | Confirma código |
| POST | `/api/webhooks/whatsapp` | Webhook para receber código |
| GET | `/api/onboarding/steps` | Passos do onboarding |
| GET | `/api/onboarding/progress` | Progresso atual |
| POST | `/api/onboarding/complete` | Marca passo completo |
| POST | `/api/onboarding/skip` | Pula onboarding |
| POST | `/api/track` | Envia evento de tracking |
| GET | `/api/admin/analytics/funnel` | Métricas do funil |
| GET | `/api/admin/analytics/drops` | Usuários travados |

