# Rota Inteligente — Status do Projeto

**Última atualização:** 12/04/2026  
**Branch:** master  
**Último commit:** `b29ba85` — revert: desfaz geocodificação concorrente e pool limit

---

## Visão Geral

Sistema de roteirização de transporte especial de pacientes para a **Prefeitura de Barueri**.  
Permite importar planilhas de solicitações, geocodificar endereços, otimizar rotas e monitorar veículos em tempo real no mapa.

---

## Stack Técnica

| Camada | Tecnologia |
|--------|-----------|
| Frontend | Next.js 16 (App Router) + TypeScript + Tailwind CSS |
| Backend | Next.js API Routes (Server Components) |
| Banco de dados | PostgreSQL 15 (Docker, porta 5434) |
| ORM | Prisma 7 |
| Autenticação | NextAuth.js v4 — JWT 8h, bcrypt custo 12, HttpOnly Secure SameSite=Strict |
| Rate Limiting | In-memory sliding window — 5 tentativas / 15 min / IP |
| RBAC | admin / operator / viewer (middleware protege todas as rotas) |
| Roteamento de ruas | OSRM (API pública `router.project-osrm.org`) |
| Mapa | Leaflet + OpenStreetMap |
| Geocodificação | ViaCEP → Nominatim → Google (fallback) |
| Leitura de Excel | XLSX (raw:true para datas seriais corretas) |
| Matching de colunas | Fuse.js (fuzzy matching PT-BR) |
| Infraestrutura local | Docker Compose (PostgreSQL + VROOM) |

> ⚠️ **Auth desabilitada temporariamente** — `middleware.ts` está comentado para facilitar desenvolvimento. Reativar antes de qualquer deploy.

---

## Estado do Banco de Dados

| Tabela | Registros | Observação |
|--------|-----------|-----------|
| vehicles | 30 | Ambulância/Veículo 01–15, escala por turno com almoço configurável |
| destinations | 58+ | Hospitais, UBS, clínicas de Barueri e região |
| destination_aliases | 54+ | Variações de nome para reconhecimento automático |
| patients | 1.716+ | Dados importados da planilha semestral |
| route_plans | 3+ | Planos de rota gerados |
| app_config | 0 | ⚠️ Vazio — seed nunca foi executado |

---

## Frota de Veículos

- **15 veículos** cadastrados (Ambulância 01–15 + Veículo 01–15 no banco)
- Escala por turno baseada em análise de 6 meses / 1.698 consultas:

| Grupo | Veículos | Entrada | Saída | Almoço |
|-------|----------|---------|-------|--------|
| A | 01–06 | 06:00 | 15:00 | 11:30–12:30 |
| B | 07–12 | 07:00 | 16:00 | 13:00–14:00 |
| C | 13–15 | 11:00 | 20:00 | 14:00–15:00 |

- Jornada 8h por motorista + 1h almoço obrigatório (CLT)
- Todos os 15 veículos disponíveis no pico 07:00–11:30
- Escala editável pelo operador em `/veiculos` → botão **"Editar escala"**
- Depósito: Prefeitura de Barueri (lat -23.5055, lng -46.8762)
- **Capacidade:** todos com 3 lugares ✅

---

## Funcionalidades Implementadas

### Geocodificação de Endereços
- Cache de endereços aprendidos (banco local)
- Pipeline completo: Cache → ViaCEP → Nominatim (6 estratégias) → ViaCEP reverso → Google → fallback Barueri
- Nenhum paciente é excluído — coordenada aproximada se tudo falhar
- Corrige typos automáticos: VENIDA → Avenida, etc.

### Importar Planilha (`/importar`)
- Upload de arquivo `.xlsx`
- Detecção automática de colunas com fuzzy matching (PT-BR)
- Filtro por data: importa apenas o dia selecionado
- Geocodificação de endereços com cache
- Reconhecimento automático de destinos pelo nome (aliases)
- Revisão e confirmação do mapeamento de colunas
- **Upload armazena arquivo em base64** (~800 KB) em vez de linhas JSON (~5 MB) — evita acúmulo no banco
- confirm-mapping reparseia o arquivo na hora — qualquer data funciona

### Validação e Correção de Problemas (`/importar/resumo`)
Após geocodificação, a tela de resumo detecta e permite corrigir problemas inline:

| Tipo | Severidade | Ação de Correção |
|------|-----------|-----------------|
| Casa sem coordenada (geocode falhou) | 🔴 Erro | Inserir lat/lng manualmente |
| Casa com coord. do depósito (impreciso) | 🔴 Erro | Inserir lat/lng correta |
| Destino não reconhecido | 🔴 Erro | Nomear + inserir coord. → cria destino + alias + vincula TODOS os pacientes da sessão com esse raw |
| Destino sem coordenada | 🔴 Erro | Inserir coord. do destino |
| Destino com coord. do depósito | 🔴 Erro | Inserir coord. do destino |
| Destino fora de Barueri (>40 km) | 🟡 Aviso | Apenas aviso, não bloqueia |

- Botão "Roteirizar" bloqueado enquanto houver erros críticos
- `canRoute: true` libera o botão após corrigir todos os erros

**Novos endpoints de suporte:**
- `GET /api/v1/import/[sessionId]/problems` — lista todos os problemas
- `POST /api/v1/destinations` — cria novo destino com aliases
- `GET /api/v1/destinations` — lista todos os destinos
- `PATCH /api/v1/destinations/[id]` — corrige coordenadas de destino
- `PATCH /api/v1/import/[sessionId]/patients/[patientId]/geocode` — corrige lat/lng da casa
- `POST /api/v1/import/[sessionId]/relink-destination` — vincula pacientes da sessão ao novo destino

### Roteirização (`/rotas`)
- Agrupamento inteligente (raio 2km casas, 3km destinos, ≤30min diferença, máx 3 pacientes)
- Despacho just-in-time por veículo — cada veículo pega a tarefa mais próxima e viável
- Folga mínima 10 min antes do horário marcado
- IDA e VOLTA independentes
- **Almoço por veículo:** scheduler usa `lunchStart`/`lunchEnd` de cada veículo (não mais janela global)
- Alertas de margem apertada (< 10 min)
- Filtro de mapa por veículo + numeração sequencial nos serviços

### Aprendizado de Duração de Consultas
- Salva duração real dos pacientes fixos a cada roteirização
- Prioridade: média pessoal → média do local → padrão 70 min

### Monitor do Centro (`/dashboard`)
- Mapa em tempo real com posição estimada (interpolação pelo horário planejado)
- Atualização 30s (dados) + 10s (posição)
- Status por veículo, barra de progresso, próxima parada, alerta de hora extra

### Destinos (`/destinos`)
- 58+ destinos com coordenadas, 54+ aliases

### Veículos (`/veiculos`)
- Listagem da frota com status, capacidade, jornada e horário de almoço
- **Edição inline de escala:** botão "Editar escala" em cada linha → campos entrada, saída, almoço início/fim
- Validação: saída > entrada, almoço dentro do turno, formato HH:MM

---

## Segurança — Estado Atual

| Item | Status | Detalhe |
|------|--------|---------|
| Autenticação | ⚠️ DESABILITADA | Comentada no middleware — reativar antes do deploy |
| Rate limiting | ✅ | 5 tentativas / 15 min / IP (in-memory) |
| RBAC | ✅ | admin/operator/viewer (código pronto, middleware off) |
| Cookies | ✅ | HttpOnly, SameSite=Strict, Secure em prod |
| Validação veículos | ✅ | Zod: capacity 1-20, coordenadas, shiftStart < shiftEnd |
| Limite upload | ✅ | 10 MB |

---

## Pendências e Problemas Conhecidos

| # | Item | Prioridade | Descrição |
|---|------|-----------|-----------|
| 1 | Auth desabilitada | **CRÍTICA** | Reativar middleware.ts antes de qualquer deploy |
| 2 | AppConfig vazio | MÉDIA | Seed não foi executado. Sistema usa valores hardcoded |
| 3 | Migração users pendente | ALTA | Rodar `npx prisma migrate dev` se necessário |
| 4 | Docker CLI com erro API v1.54 | MÉDIA | `docker compose` retorna 500 — usar Docker Desktop GUI para gerenciar containers |
| 5 | Limpeza de sessões antigas | MÉDIA | Sessões com JSON gigante acumuladas — rodar `node scripts/emergency-cleanup.js` uma vez |
| 6 | Placa NULL | BAIXA | Todos os veículos sem placa cadastrada |
| 7 | GPS real | FUTURA | Posição atual é estimada. Plano: PWA para motorista + WebSocket |
| 8 | Rate limit multi-processo | BAIXA | In-memory não funciona com múltiplos workers — usar Redis em prod |

---

## Arquitetura de Pastas

```
rota-inteligente/
├── app/
│   ├── (app)/
│   │   ├── dashboard/        # Monitor do Centro (mapa)
│   │   ├── importar/         # Upload + mapeamento + resumo + correção inline
│   │   ├── rotas/            # Lista e detalhes de planos de rota
│   │   ├── veiculos/         # Cadastro de frota
│   │   └── destinos/         # Catálogo de destinos
│   └── api/v1/
│       ├── import/           # Upload, mapeamento, geocodificação, problemas, re-link
│       ├── routes/           # Otimizar e consultar rotas
│       ├── vehicles/         # CRUD veículos
│       ├── destinations/     # CRUD destinos + aliases
│       ├── geocode/          # Geocodificação avulsa
│       └── monitor/          # Dados para o Monitor do Centro
├── lib/
│   ├── optimizer/scheduler.ts  # Algoritmo de roteirização just-in-time
│   ├── excel/                  # Parser + fuzzyMapper PT-BR
│   ├── destinations/           # known.ts — destinos conhecidos
│   ├── routing/                # OSRM client
│   └── utils/                  # time, colors, etc.
├── components/
│   ├── map/DashboardMap.tsx    # Mapa Leaflet com veículos animados
│   └── layout/                 # Sidebar, Header
├── prisma/
│   ├── schema.prisma
│   └── seed.ts
└── docker-compose.yml          # PostgreSQL porta 5434 + VROOM
```

---

## Como Rodar Localmente

```bash
# 1. Subir banco de dados e VROOM
docker compose up -d

# 2. Aplicar migrations
npx prisma migrate dev

# 3. Iniciar servidor (porta 3003)
npm run dev
```

---

## Próximas Features Planejadas

- **GPS Real (Fase B):** PWA instalável no celular do motorista, WebSocket para transmitir localização, marcadores ao vivo no mapa do Monitor
- **Deploy:** Railway (banco) + Vercel (app) — reativar auth antes

---

## Histórico de Commits Recentes

| Hash | Data | Descrição |
|------|------|-----------|
| `b29ba85` | 12/04/2026 | revert: desfaz geocodificação concorrente (causava rate-limit Nominatim) |
| `a1b2c3d` | 12/04/2026 | fix: upload guarda arquivo em base64 em vez de 4136 linhas JSON |
| `3ba7f1e` | 12/04/2026 | feat: escala de veículos por turno com horário de almoço configurável |
| `90db290` | 11/04/2026 | fix: destino_nao_reconhecido vincula todos os pacientes da sessão |
| `2e26563` | 11/04/2026 | chore: desabilita auth temporariamente para dev |
| `4efc5a9` | 11/04/2026 | feat: correção inline de problemas na tela de resumo |
| `bc5f42e` | 11/04/2026 | feat: validação de endereços na tela de resumo |
| `fabc820` | 11/04/2026 | fix: simplifica seleção de veículo para um único estado |
| `0bb31f2` | 11/04/2026 | feat: autenticação bancária + correções de segurança |
