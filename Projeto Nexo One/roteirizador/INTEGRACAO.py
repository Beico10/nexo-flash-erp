# =============================================================================
# COMO INTEGRAR O ROTEIRIZADOR NO server.py EXISTENTE
# =============================================================================

# 1. Adicionar imports no topo do server.py (após os imports existentes):
#
#    from router_module import router_api
#
# 2. Registrar o router no app FastAPI (após a criação do `app`):
#
#    app.include_router(router_api)
#
# 3. Instalar dependências (se ainda não tiver):
#
#    pip install httpx pydantic

# =============================================================================
# TRECHO EXATO PARA ADICIONAR NO server.py
# Copie e cole logo após a linha: app = FastAPI(title="Nexo One Proxy")
# =============================================================================

TRECHO = """
# ── Roteirizador Inteligente ─────────────────────────────────────────────────
from router_module import router_api
app.include_router(router_api)
"""

# =============================================================================
# ENDPOINTS DISPONÍVEIS APÓS INTEGRAÇÃO
# =============================================================================
ENDPOINTS = """
POST /api/v1/logistics/route/optimize
    Otimiza rota com 2-opt para 100+ paradas
    Body: { origin, destinations, vehicle_capacity_kg, optimize }

GET  /api/v1/logistics/routes
    Lista rotas salvas

POST /api/v1/logistics/routes
    Salva rota planejada

GET  /api/v1/logistics/distance
    Distância real entre 2 pontos
    Query: ?from_lat=&from_lng=&to_lat=&to_lng=

POST /api/v1/logistics/route/recalculate
    Recalcula rota a partir da posição atual do motorista
"""

# =============================================================================
# FRONTEND
# =============================================================================
FRONTEND = """
Arquivo: frontend/src/app/logistics/page.tsx
Substituir pelo arquivo logistics_page.tsx fornecido.

Tecnologias:
  - Leaflet.js (mapa open source, zero API key)
  - OpenStreetMap (tiles gratuitos)
  - OSRM (rotas reais, gratuito)
  - Algoritmo 2-opt nativo em Python

Funcionalidades:
  - Mapa interativo com rota desenhada
  - Marcadores numerados por ordem de visita
  - Navegação passo a passo (estilo Waze)
  - Botão "Entregue" para avançar na rota
  - Barra de progresso da entrega
  - Recalculo automático se motorista desviar
  - Fallback para linha reta se OSRM cair
"""
