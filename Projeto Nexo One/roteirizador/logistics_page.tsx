'use client'
import { useState, useEffect, useRef, useCallback } from 'react'
import { Navigation, MapPin, Clock, Route, ChevronRight, ChevronLeft, RotateCcw, Play, Pause, AlertTriangle, CheckCircle2, Truck } from 'lucide-react'

// ── TIPOS ─────────────────────────────────────────────────────────────────────
interface Stop {
  rank: number
  label: string
  lat: number
  lng: number
  weight_kg: number
  priority: number
}

interface Leg {
  step: number
  from: string
  to: string
  distance_km: number
  duration_min: number
  from_lat: number
  from_lng: number
  to_lat: number
  to_lng: number
}

interface RouteData {
  total_distance_km: number
  total_duration_min: number
  total_duration_h: string
  total_stops: number
  total_weight_kg: number
  stops_ordered: Stop[]
  legs: Leg[]
  geometry: any
  source: string
  algorithm: string
}

// ── COMPONENTE PRINCIPAL ─────────────────────────────────────────────────────
export default function RoteirizadorPage() {
  const mapRef = useRef<any>(null)
  const leafletRef = useRef<any>(null)
  const markersRef = useRef<any[]>([])
  const routeLayerRef = useRef<any>(null)

  const [route, setRoute] = useState<RouteData | null>(null)
  const [loading, setLoading] = useState(false)
  const [navigating, setNavigating] = useState(false)
  const [currentStop, setCurrentStop] = useState(0)
  const [completedStops, setCompletedStops] = useState<number[]>([])
  const [mapReady, setMapReady] = useState(false)
  const [originInput, setOriginInput] = useState('Centro de Distribuição SP')
  const [destinations, setDestinations] = useState([
    { label: 'Cliente A — Vila Mariana', lat: -23.5931, lng: -46.6395, weight_kg: 150 },
    { label: 'Cliente B — Pinheiros',    lat: -23.5631, lng: -46.6911, weight_kg: 80  },
    { label: 'Cliente C — Lapa',         lat: -23.5223, lng: -46.7070, weight_kg: 200 },
    { label: 'Cliente D — Santo André',  lat: -23.6644, lng: -46.5382, weight_kg: 120 },
    { label: 'Cliente E — Guarulhos',    lat: -23.4543, lng: -46.5337, weight_kg: 95  },
  ])

  // ── INICIALIZAR MAPA ──────────────────────────────────────────────────────
  useEffect(() => {
    if (typeof window === 'undefined' || mapReady) return

    const loadLeaflet = async () => {
      // Carregar CSS do Leaflet
      if (!document.getElementById('leaflet-css')) {
        const link = document.createElement('link')
        link.id = 'leaflet-css'
        link.rel = 'stylesheet'
        link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css'
        document.head.appendChild(link)
      }

      // Carregar JS do Leaflet
      await new Promise<void>((resolve) => {
        if ((window as any).L) { resolve(); return }
        const script = document.createElement('script')
        script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js'
        script.onload = () => resolve()
        document.head.appendChild(script)
      })

      const L = (window as any).L
      leafletRef.current = L

      // Inicializar mapa centrado em São Paulo
      const map = L.map('nexo-map', { zoomControl: true }).setView([-23.5505, -46.6333], 11)

      // Tiles OpenStreetMap — gratuito, sem API key
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        maxZoom: 19,
      }).addTo(map)

      mapRef.current = map
      setMapReady(true)
    }

    loadLeaflet()
  }, [mapReady])

  // ── DESENHAR ROTA NO MAPA ────────────────────────────────────────────────
  const drawRoute = useCallback((routeData: RouteData) => {
    const L = leafletRef.current
    const map = mapRef.current
    if (!L || !map) return

    // Limpar camadas anteriores
    markersRef.current.forEach(m => map.removeLayer(m))
    markersRef.current = []
    if (routeLayerRef.current) map.removeLayer(routeLayerRef.current)

    // Desenhar linha da rota (geometry GeoJSON do OSRM)
    if (routeData.geometry) {
      const geoLayer = L.geoJSON(routeData.geometry, {
        style: { color: '#1565C0', weight: 4, opacity: 0.8, dashArray: '0' }
      }).addTo(map)
      routeLayerRef.current = geoLayer
    } else {
      // Fallback: linha reta entre paradas
      const latlngs = routeData.stops_ordered.map(s => [s.lat, s.lng])
      const poly = L.polyline(latlngs, { color: '#1565C0', weight: 3, opacity: 0.7 }).addTo(map)
      routeLayerRef.current = poly
    }

    // Marcadores numerados por ordem
    routeData.stops_ordered.forEach((stop, idx) => {
      const isCompleted = completedStops.includes(idx)
      const isCurrent = idx === currentStop

      const color = isCompleted ? '#2E7D32' : isCurrent ? '#E65100' : '#1565C0'
      const num = stop.rank

      const icon = L.divIcon({
        html: `<div style="
          background:${color};color:white;
          width:28px;height:28px;border-radius:50%;
          display:flex;align-items:center;justify-content:center;
          font-size:11px;font-weight:700;
          border:2px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.3);
          ${isCurrent ? 'transform:scale(1.3);' : ''}
        ">${num}</div>`,
        iconSize: [28, 28],
        iconAnchor: [14, 14],
        className: '',
      })

      const marker = L.marker([stop.lat, stop.lng], { icon })
        .bindPopup(`<b>${stop.label}</b><br>${stop.weight_kg}kg`)
        .addTo(map)

      markersRef.current.push(marker)
    })

    // Ajustar zoom para mostrar toda a rota
    if (markersRef.current.length > 0) {
      const group = L.featureGroup(markersRef.current)
      map.fitBounds(group.getBounds().pad(0.1))
    }
  }, [completedStops, currentStop])

  useEffect(() => {
    if (route && mapReady) drawRoute(route)
  }, [route, mapReady, drawRoute])

  // ── CALCULAR ROTA ─────────────────────────────────────────────────────────
  const calcularRota = async () => {
    setLoading(true)
    try {
      const body = {
        origin: { lat: -23.5505, lng: -46.6333, label: originInput, weight_kg: 0 },
        destinations,
        vehicle_capacity_kg: 1000,
        optimize: true,
      }

      const res = await fetch('/api/v1/logistics/route/optimize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      if (!res.ok) throw new Error('Erro ao calcular rota')
      const data = await res.json()
      setRoute(data.route)
      setCurrentStop(0)
      setCompletedStops([])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  // ── NAVEGAÇÃO PASSO A PASSO ───────────────────────────────────────────────
  const proximaParada = () => {
    if (!route) return
    setCompletedStops(prev => [...prev, currentStop])
    if (currentStop < route.stops_ordered.length - 1) {
      setCurrentStop(prev => prev + 1)
      // Centralizar no próximo ponto
      const next = route.stops_ordered[currentStop + 1]
      if (mapRef.current) mapRef.current.setView([next.lat, next.lng], 14)
    }
  }

  const paradaAnterior = () => {
    if (currentStop > 0) {
      setCompletedStops(prev => prev.filter(s => s !== currentStop - 1))
      setCurrentStop(prev => prev - 1)
    }
  }

  const reiniciarNavegacao = () => {
    setCurrentStop(0)
    setCompletedStops([])
    setNavigating(false)
    if (route) drawRoute(route)
  }

  const paradaAtual = route?.stops_ordered[currentStop]
  const progresso = route ? Math.round((completedStops.length / route.stops_ordered.length) * 100) : 0

  return (
    <div className="flex h-screen" style={{ background: '#F0F2F8' }}>

      {/* ── PAINEL LATERAL ── */}
      <div style={{ width: 380, background: 'white', borderRight: '1px solid #E0E4F0', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

        {/* Header */}
        <div style={{ padding: '16px 20px', background: 'linear-gradient(135deg, #0A3D8F, #1565C0)', color: 'white' }}>
          <div className="flex items-center gap-2 mb-1">
            <Route size={18} />
            <span style={{ fontSize: 15, fontWeight: 700 }}>Roteirizador Inteligente</span>
          </div>
          <p style={{ fontSize: 11, opacity: 0.7 }}>OSRM • OpenStreetMap • Algoritmo 2-opt</p>
        </div>

        {/* Se não tem rota — formulário */}
        {!route && (
          <div style={{ padding: 20, flex: 1, overflowY: 'auto' }}>
            <p style={{ fontSize: 12, fontWeight: 700, color: '#424242', marginBottom: 8, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Origem</p>
            <input
              value={originInput}
              onChange={e => setOriginInput(e.target.value)}
              style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 10, fontSize: 13, marginBottom: 16, outline: 'none' }}
            />

            <div className="flex items-center justify-between mb-2">
              <p style={{ fontSize: 12, fontWeight: 700, color: '#424242', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Destinos ({destinations.length})
              </p>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 16 }}>
              {destinations.map((d, i) => (
                <div key={i} style={{ padding: '8px 12px', background: '#F5F7FF', borderRadius: 8, border: '1px solid #E0E4F0', fontSize: 12 }}>
                  <div style={{ fontWeight: 600, color: '#212121' }}>{d.label}</div>
                  <div style={{ color: '#757575', marginTop: 2 }}>{d.weight_kg}kg</div>
                </div>
              ))}
            </div>

            <button
              onClick={calcularRota}
              disabled={loading}
              style={{
                width: '100%', padding: '13px', borderRadius: 12,
                background: loading ? '#90CAF9' : 'linear-gradient(135deg, #0A3D8F, #1565C0)',
                color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: loading ? 'wait' : 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
              }}
            >
              {loading ? (
                <><div style={{ width: 16, height: 16, border: '2px solid rgba(255,255,255,0.3)', borderTopColor: 'white', borderRadius: '50%', animation: 'spin 0.7s linear infinite' }} /> Calculando...</>
              ) : (
                <><Navigation size={16} /> Otimizar Rota</>
              )}
            </button>
          </div>
        )}

        {/* Se tem rota — painel de navegação */}
        {route && (
          <>
            {/* Métricas */}
            <div style={{ padding: '12px 16px', background: '#F5F7FF', borderBottom: '1px solid #E0E4F0', display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
              {[
                { icon: <Route size={14} />, val: `${route.total_distance_km}km`, label: 'Distância' },
                { icon: <Clock size={14} />, val: route.total_duration_h, label: 'Tempo' },
                { icon: <MapPin size={14} />, val: `${route.total_stops}`, label: 'Paradas' },
              ].map((m, i) => (
                <div key={i} style={{ textAlign: 'center', padding: 6 }}>
                  <div style={{ color: '#1565C0', display: 'flex', justifyContent: 'center', marginBottom: 2 }}>{m.icon}</div>
                  <div style={{ fontSize: 14, fontWeight: 700, color: '#0A3D8F' }}>{m.val}</div>
                  <div style={{ fontSize: 10, color: '#757575' }}>{m.label}</div>
                </div>
              ))}
            </div>

            {/* Progresso */}
            <div style={{ padding: '10px 16px', borderBottom: '1px solid #E0E4F0' }}>
              <div className="flex items-center justify-between mb-1">
                <span style={{ fontSize: 11, color: '#757575' }}>Progresso da rota</span>
                <span style={{ fontSize: 11, fontWeight: 700, color: '#1565C0' }}>{progresso}%</span>
              </div>
              <div style={{ height: 6, background: '#E0E4F0', borderRadius: 3 }}>
                <div style={{ height: '100%', background: '#1565C0', borderRadius: 3, width: `${progresso}%`, transition: 'width 0.3s' }} />
              </div>
            </div>

            {/* Parada atual */}
            {navigating && paradaAtual && (
              <div style={{ margin: 16, padding: 14, background: '#E3F2FD', borderRadius: 12, border: '2px solid #1565C0' }}>
                <div style={{ fontSize: 10, fontWeight: 700, color: '#1565C0', textTransform: 'uppercase', marginBottom: 4 }}>
                  Parada {currentStop + 1} de {route.stops_ordered.length}
                </div>
                <div style={{ fontSize: 14, fontWeight: 700, color: '#0A3D8F' }}>{paradaAtual.label}</div>
                <div style={{ fontSize: 12, color: '#424242', marginTop: 4 }}>{paradaAtual.weight_kg}kg</div>

                <div className="flex gap-2 mt-3">
                  <button onClick={paradaAnterior} disabled={currentStop === 0}
                    style={{ flex: 1, padding: '8px', borderRadius: 8, background: '#E0E4F0', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 4, fontSize: 12 }}>
                    <ChevronLeft size={14} /> Anterior
                  </button>
                  <button onClick={proximaParada}
                    style={{ flex: 2, padding: '8px', borderRadius: 8, background: '#1565C0', color: 'white', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 4, fontSize: 12, fontWeight: 700 }}>
                    <CheckCircle2 size={14} /> Entregue <ChevronRight size={14} />
                  </button>
                </div>
              </div>
            )}

            {/* Lista de paradas */}
            <div style={{ flex: 1, overflowY: 'auto', padding: '8px 0' }}>
              {route.stops_ordered.map((stop, idx) => {
                const done = completedStops.includes(idx)
                const current = idx === currentStop && navigating
                return (
                  <div key={idx} style={{
                    padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 10,
                    background: current ? '#E3F2FD' : done ? '#F1F8E9' : 'white',
                    borderLeft: current ? '3px solid #1565C0' : done ? '3px solid #2E7D32' : '3px solid transparent',
                  }}>
                    <div style={{
                      width: 26, height: 26, borderRadius: '50%', flexShrink: 0,
                      background: done ? '#2E7D32' : current ? '#E65100' : '#E0E4F0',
                      color: done || current ? 'white' : '#424242',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontSize: 11, fontWeight: 700,
                    }}>
                      {done ? '✓' : stop.rank}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: 13, fontWeight: 600, color: done ? '#2E7D32' : '#212121' }}>{stop.label}</div>
                      <div style={{ fontSize: 11, color: '#757575' }}>{stop.weight_kg}kg</div>
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Botões de ação */}
            <div style={{ padding: 12, borderTop: '1px solid #E0E4F0', display: 'flex', gap: 8 }}>
              <button onClick={reiniciarNavegacao}
                style={{ flex: 1, padding: '10px', borderRadius: 10, background: '#F5F7FF', border: '1px solid #E0E4F0', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, fontSize: 12 }}>
                <RotateCcw size={14} /> Resetar
              </button>
              <button onClick={() => setRoute(null)}
                style={{ flex: 1, padding: '10px', borderRadius: 10, background: '#F5F7FF', border: '1px solid #E0E4F0', cursor: 'pointer', fontSize: 12 }}>
                Nova rota
              </button>
              <button
                onClick={() => setNavigating(v => !v)}
                style={{ flex: 2, padding: '10px', borderRadius: 10, background: navigating ? '#E65100' : '#1565C0', color: 'white', border: 'none', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6, fontSize: 12, fontWeight: 700 }}>
                {navigating ? <><Pause size={14} /> Pausar</> : <><Play size={14} /> Iniciar</>}
              </button>
            </div>
          </>
        )}
      </div>

      {/* ── MAPA ── */}
      <div style={{ flex: 1, position: 'relative' }}>
        <div id="nexo-map" style={{ width: '100%', height: '100%' }} />

        {/* Badge fonte */}
        {route && (
          <div style={{
            position: 'absolute', bottom: 24, right: 24,
            background: 'white', padding: '6px 12px', borderRadius: 8,
            boxShadow: '0 2px 12px rgba(0,0,0,0.15)',
            fontSize: 11, color: '#424242',
            display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <Truck size={12} />
            {route.source === 'osrm' ? 'Rota real via OSRM' : 'Rota estimada (OSRM indisponível)'}
            {' · '}Algoritmo: {route.algorithm === 'exact_tsp' ? 'TSP Exato' : '2-opt'}
          </div>
        )}

        {/* Loading overlay */}
        {loading && (
          <div style={{ position: 'absolute', inset: 0, background: 'rgba(255,255,255,0.8)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: 12 }}>
            <div style={{ width: 40, height: 40, border: '4px solid #E0E4F0', borderTopColor: '#1565C0', borderRadius: '50%', animation: 'spin 0.7s linear infinite' }} />
            <p style={{ fontSize: 14, fontWeight: 600, color: '#1565C0' }}>Calculando rota otimizada...</p>
          </div>
        )}
      </div>

      <style jsx global>{`
        @keyframes spin { to { transform: rotate(360deg); } }
      `}</style>
    </div>
  )
}
