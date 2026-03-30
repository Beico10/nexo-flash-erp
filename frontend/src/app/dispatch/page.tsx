'use client'
import { useState, useCallback, useEffect } from 'react'
import { isDemoMode, promptLogin } from '@/lib/demo'
import { Upload, Truck, MapPin, AlertTriangle, CheckCircle2, X, BarChart3, Route, QrCode } from 'lucide-react'

interface DeliveryItem {
  id: string; nfe_key: string; nfe_number: string
  recipient_name: string; full_address: string
  city: string; state: string; lat: number; lng: number
  geo_status: string; weight_kg: number; cubage_m3: number
  volumes: number; total_value: number; status: string
  vehicle_id: string; import_source: string
}

interface Vehicle {
  id: string; name: string; type: string; plate: string; driver: string
  max_weight_kg: number; max_cubage_m3: number; max_stops: number
  is_available: boolean
}

interface Assignment {
  vehicle: Vehicle; items: DeliveryItem[]
  total_stops: number; total_weight: number; total_cubage: number
  weight_used_pct: number; cubage_used_pct: number
  over_weight: boolean; over_cubage: boolean; warnings: string[]
}

const DEMO_VEHICLES: Vehicle[] = [
  { id: 'v1', name: 'Fiorino BRA-2E19', type: 'fiorino', plate: 'BRA-2E19', driver: 'Anderson Silva', max_weight_kg: 500, max_cubage_m3: 3.5, max_stops: 15, is_available: true },
  { id: 'v2', name: 'Sprinter ABC-1D23', type: 'van', plate: 'ABC-1D23', driver: 'Carlos Mendes', max_weight_kg: 1200, max_cubage_m3: 9.0, max_stops: 25, is_available: true },
  { id: 'v3', name: 'HR DEF-4G56', type: 'truck', plate: 'DEF-4G56', driver: 'Roberto Alves', max_weight_kg: 3500, max_cubage_m3: 18.0, max_stops: 12, is_available: false },
]

const DEMO_DELIVERIES: DeliveryItem[] = [
  { id: 'd1', nfe_key: '35260300000000000000550010000000011000000011', nfe_number: 'NF-000.001', recipient_name: 'Eletro Distribuidora', full_address: 'Av. Paulista 1500, Sao Paulo - SP', city: 'Sao Paulo', state: 'SP', lat: -23.5611, lng: -46.6558, geo_status: 'geocoded', weight_kg: 280, cubage_m3: 1.2, volumes: 8, total_value: 14800, status: 'pending', vehicle_id: 'v1', import_source: 'xml' },
  { id: 'd2', nfe_key: '35260300000000000000550010000000021000000021', nfe_number: 'NF-000.002', recipient_name: 'Moveis & Cia', full_address: 'Rua Augusta 800, Sao Paulo - SP', city: 'Sao Paulo', state: 'SP', lat: -23.5489, lng: -46.6424, geo_status: 'geocoded', weight_kg: 450, cubage_m3: 3.8, volumes: 5, total_value: 8200, status: 'pending', vehicle_id: 'v2', import_source: 'xml' },
  { id: 'd3', nfe_key: '35260300000000000000550010000000031000000031', nfe_number: 'NF-000.003', recipient_name: 'Supermercado Rede', full_address: 'R. Consolacao 2100, Sao Paulo - SP', city: 'Sao Paulo', state: 'SP', lat: -23.5522, lng: -46.6604, geo_status: 'geocoded', weight_kg: 180, cubage_m3: 0.9, volumes: 12, total_value: 3400, status: 'in_transit', vehicle_id: 'v1', import_source: 'xml' },
  { id: 'd4', nfe_key: '35260300000000000000550010000000041000000041', nfe_number: 'NF-000.004', recipient_name: 'Atacadao Norte', full_address: 'Av. Brasil 4500, Sao Paulo - SP', city: 'Sao Paulo', state: 'SP', lat: -23.5358, lng: -46.6255, geo_status: 'geocoded', weight_kg: 820, cubage_m3: 5.2, volumes: 20, total_value: 22100, status: 'pending', vehicle_id: 'v2', import_source: 'xml' },
  { id: 'd5', nfe_key: '35260300000000000000550010000000051000000051', nfe_number: 'NF-000.005', recipient_name: 'Farmacia Popular', full_address: 'R. da Consolacao 500, Sao Paulo - SP', city: 'Sao Paulo', state: 'SP', lat: -23.5578, lng: -46.6598, geo_status: 'geocoded', weight_kg: 45, cubage_m3: 0.3, volumes: 3, total_value: 1850, status: 'delivered', vehicle_id: 'v1', import_source: 'scan' },
]

const DEMO_ASSIGNMENTS: Assignment[] = [
  {
    vehicle: DEMO_VEHICLES[0],
    items: [DEMO_DELIVERIES[0], DEMO_DELIVERIES[2], DEMO_DELIVERIES[4]],
    total_stops: 3, total_weight: 505, total_cubage: 2.4,
    weight_used_pct: 101, cubage_used_pct: 68.6,
    over_weight: true, over_cubage: false,
    warnings: ['⚠️ Peso excede capacidade do Fiorino — mover 1 entrega para a Sprinter'],
  },
  {
    vehicle: DEMO_VEHICLES[1],
    items: [DEMO_DELIVERIES[1], DEMO_DELIVERIES[3]],
    total_stops: 2, total_weight: 1270, total_cubage: 9.0,
    weight_used_pct: 105.8, cubage_used_pct: 100,
    over_weight: true, over_cubage: false,
    warnings: ['⚠️ Cubagem no limite máximo — verificar antes de carregar'],
  },
]

const VEHICLE_ICON: Record<string, string> = {
  fiorino: '🚐', van: '🚌', truck: '🚛', carreta: '🚚', moto: '🏍️'
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export default function DispatchPage() {
  const [items, setItems] = useState<DeliveryItem[]>([])
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [assignments, setAssignments] = useState<Assignment[]>([])
  const [loading, setLoading] = useState(false)
  const [importing, setImporting] = useState(false)
  const [distributing, setDistributing] = useState(false)
  const [tab, setTab] = useState<'import' | 'vehicles' | 'distribute' | 'route'>('import')
  const [dragOver, setDragOver] = useState(false)
  const [importResult, setImportResult] = useState<any>(null)
  const [scanCode, setScanCode] = useState('')
  const [newVehicle, setNewVehicle] = useState({ name: '', type: 'van', max_weight_kg: '', max_cubage_m3: '', max_stops: '' })
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  // Carregar demo no mount
  useEffect(() => {
    if (isDemoMode()) {
      setItems(DEMO_DELIVERIES)
      setVehicles(DEMO_VEHICLES)
      setAssignments(DEMO_ASSIGNMENTS)
      setTab('route')
      setImportResult({ message: '5 entregas importadas com sucesso', total_items: 5, total_weight_kg: 1775, total_cubage_m3: 11.4, total_value: 50350, errors: [] })
    }
  }, [])

  // Carregar veículos
  const loadVehicles = async () => {
    const res = await fetch('/api/v1/dispatch/vehicles', { headers: { Authorization: `Bearer ${token}` } })
    if (res.ok) setVehicles((await res.json()).vehicles || [])
  }

  // Import em lote
  const handleFiles = useCallback(async (files: FileList) => {
    if (isDemoMode()) { promptLogin(); return }
    setImporting(true)
    const formData = new FormData()
    Array.from(files).forEach(f => formData.append('files', f))

    try {
      const res = await fetch('/api/v1/dispatch/import', {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      })
      if (res.ok) {
        const data = await res.json()
        setImportResult(data)
        setItems(data.items || [])
        if (data.items?.length > 0) setTab('vehicles')
      }
    } finally { setImporting(false) }
  }, [token])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files)
  }, [handleFiles])

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files?.length) handleFiles(e.target.files)
  }

  // Scanner
  const handleScan = async () => {
    if (scanCode.length !== 44) return
    const res = await fetch('/api/v1/dispatch/scan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ barcode: scanCode }),
    })
    if (res.ok) {
      const data = await res.json()
      setItems(prev => [...prev, data.item])
      setScanCode('')
    }
  }

  // Distribuição
  const handleDistribute = async () => {
    setDistributing(true)
    const availableVehicles = vehicles.filter(v => v.is_available)
    try {
      const res = await fetch('/api/v1/dispatch/batches/current/distribute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ items, vehicles: availableVehicles }),
      })
      if (res.ok) {
        const data = await res.json()
        setAssignments(data.assignments || [])
        setTab('route')
      }
    } finally { setDistributing(false) }
  }

  // Enviar para roteirizador
  const sendToRouter = async (assignment: Assignment) => {
    const res = await fetch('/api/v1/dispatch/batches/current/route', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        vehicle_id: assignment.vehicle.id,
        items: assignment.items,
        origin: { lat: -23.5505, lng: -46.6333, label: 'Base / CD' },
        vehicle_capacity_kg: assignment.vehicle.max_weight_kg,
      }),
    })
    if (res.ok) {
      alert(`${assignment.items.length} entregas enviadas para o roteirizador!`)
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>🚛 Despacho em Lote</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>
            Importe XML, CSV, EDI ou PDF • Distribua por veículo • Roteirize
          </p>
        </div>
        {items.length > 0 && (
          <div style={{ background: '#E3F2FD', padding: '8px 16px', borderRadius: 10, fontSize: 13, fontWeight: 700, color: '#1565C0' }}>
            {items.length} entregas na fila
          </div>
        )}
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', background: '#F0F2F8', borderRadius: 10, padding: 4, gap: 4, marginBottom: 24 }}>
        {[
          { key: 'import', label: '📥 Importar' },
          { key: 'vehicles', label: '🚛 Veículos' },
          { key: 'distribute', label: '⚖️ Distribuir' },
          { key: 'route', label: '🗺️ Roteirizar' },
        ].map(t => (
          <button key={t.key} onClick={() => { setTab(t.key as any); if (t.key === 'vehicles') loadVehicles() }} style={{
            flex: 1, padding: '8px', borderRadius: 8, border: 'none', cursor: 'pointer',
            fontWeight: 600, fontSize: 12,
            background: tab === t.key ? 'white' : 'transparent',
            color: tab === t.key ? '#0A3D8F' : '#757575',
            boxShadow: tab === t.key ? '0 1px 4px rgba(0,0,0,0.1)' : 'none',
          }}>
            {t.label}
          </button>
        ))}
      </div>

      {/* TAB: Importar */}
      {tab === 'import' && (
        <div>
          {/* Drag & Drop */}
          <div
            onDragOver={e => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
            style={{
              border: `2px dashed ${dragOver ? '#1565C0' : '#C5CAE9'}`,
              borderRadius: 16, padding: '48px 24px', textAlign: 'center',
              background: dragOver ? '#E3F2FD' : '#F5F7FF',
              transition: 'all 0.2s', cursor: 'pointer', marginBottom: 20,
            }}
            onClick={() => document.getElementById('file-input')?.click()}
          >
            <Upload size={40} style={{ color: '#1565C0', marginBottom: 12, opacity: 0.6 }} />
            <p style={{ fontSize: 16, fontWeight: 700, color: '#1565C0', marginBottom: 8 }}>
              {importing ? 'Processando...' : 'Arraste os arquivos aqui ou clique para selecionar'}
            </p>
            <p style={{ fontSize: 13, color: '#757575' }}>
              Suporta: XML NF-e • CSV • EDI ANSI X12 • PDF Romaneio
            </p>
            <p style={{ fontSize: 11, color: '#9E9E9E', marginTop: 8 }}>
              Múltiplos arquivos aceitos • Até 50MB por envio
            </p>
            <input id="file-input" type="file" multiple accept=".xml,.csv,.txt,.edi,.x12,.pdf" onChange={handleFileInput} style={{ display: 'none' }} />
          </div>

          {/* Scanner (fluxo secundário) */}
          <div style={{ background: 'white', borderRadius: 12, padding: 16, border: '1.5px solid #E0E4F0', marginBottom: 20 }}>
            <p style={{ fontSize: 12, fontWeight: 700, color: '#757575', textTransform: 'uppercase', marginBottom: 8 }}>
              <QrCode size={12} style={{ display: 'inline', marginRight: 4 }} />
              Scanner (nota avulsa)
            </p>
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                value={scanCode}
                onChange={e => setScanCode(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleScan()}
                placeholder="Bipe o código de barras da NF-e (44 dígitos)..."
                style={{ flex: 1, padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 13, outline: 'none', fontFamily: 'monospace' }}
              />
              <button onClick={handleScan} disabled={scanCode.length !== 44} style={{
                padding: '10px 16px', borderRadius: 8, border: 'none', cursor: 'pointer',
                background: scanCode.length === 44 ? '#1565C0' : '#E0E4F0',
                color: scanCode.length === 44 ? 'white' : '#9E9E9E', fontWeight: 700, fontSize: 13,
              }}>Adicionar</button>
            </div>
            {scanCode.length > 0 && scanCode.length !== 44 && (
              <p style={{ fontSize: 11, color: '#E65100', marginTop: 4 }}>{scanCode.length}/44 dígitos</p>
            )}
          </div>

          {/* Resultado da importação */}
          {importResult && (
            <div style={{ background: '#E8F5E9', borderRadius: 12, padding: 16, border: '1.5px solid #A5D6A7' }}>
              <p style={{ fontSize: 14, fontWeight: 700, color: '#2E7D32', marginBottom: 8 }}>
                ✅ {importResult.message}
              </p>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 8 }}>
                {[
                  { label: 'Entregas', value: importResult.total_items },
                  { label: 'Peso Total', value: `${importResult.total_weight_kg?.toFixed(0)}kg` },
                  { label: 'Cubagem', value: `${importResult.total_cubage_m3?.toFixed(2)}m³` },
                  { label: 'Valor Total', value: fmt(importResult.total_value || 0) },
                ].map((s, i) => (
                  <div key={i} style={{ textAlign: 'center', background: 'white', borderRadius: 8, padding: 10 }}>
                    <div style={{ fontSize: 16, fontWeight: 700, color: '#2E7D32' }}>{s.value}</div>
                    <div style={{ fontSize: 11, color: '#757575' }}>{s.label}</div>
                  </div>
                ))}
              </div>
              {importResult.errors?.length > 0 && (
                <div style={{ marginTop: 8 }}>
                  {importResult.errors.map((e: string, i: number) => (
                    <p key={i} style={{ fontSize: 11, color: '#B71C1C' }}>⚠️ {e}</p>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Lista de itens */}
          {items.length > 0 && (
            <div style={{ marginTop: 20 }}>
              <p style={{ fontSize: 13, fontWeight: 700, color: '#424242', marginBottom: 10 }}>
                Entregas na fila ({items.length})
              </p>
              <div style={{ maxHeight: 400, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 6 }}>
                {items.map((item, i) => (
                  <div key={i} style={{ background: 'white', borderRadius: 10, padding: '10px 14px', border: '1.5px solid #E0E4F0', display: 'flex', gap: 12, alignItems: 'center' }}>
                    <div style={{ width: 24, height: 24, borderRadius: '50%', background: '#E3F2FD', color: '#1565C0', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10, fontWeight: 700, flexShrink: 0 }}>
                      {i + 1}
                    </div>
                    <div style={{ flex: 1 }}>
                      <p style={{ fontSize: 13, fontWeight: 600, color: '#212121' }}>{item.recipient_name || 'Sem nome'}</p>
                      <p style={{ fontSize: 11, color: '#757575' }}>{item.full_address || item.city}</p>
                    </div>
                    <div style={{ fontSize: 12, color: '#424242', textAlign: 'right' }}>
                      <p>{item.weight_kg}kg</p>
                      <p style={{ color: '#9E9E9E' }}>{item.volumes} vol.</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* TAB: Veículos */}
      {tab === 'vehicles' && (
        <div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12, marginBottom: 20 }}>
            {vehicles.map(v => (
              <div key={v.id} style={{
                background: 'white', borderRadius: 12, padding: 16,
                border: `1.5px solid ${v.is_available ? '#A5D6A7' : '#E0E4F0'}`,
                opacity: v.is_available ? 1 : 0.6,
              }}>
                <div className="flex items-center justify-between mb-2">
                  <span style={{ fontSize: 20 }}>{VEHICLE_ICON[v.type] || '🚛'}</span>
                  <span style={{ fontSize: 11, fontWeight: 700, padding: '2px 8px', borderRadius: 100, background: v.is_available ? '#E8F5E9' : '#F5F5F5', color: v.is_available ? '#2E7D32' : '#9E9E9E' }}>
                    {v.is_available ? 'Disponível' : 'Indisponível'}
                  </span>
                </div>
                <p style={{ fontSize: 14, fontWeight: 700, color: '#212121' }}>{v.name}</p>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 6, marginTop: 10 }}>
                  {[
                    { label: 'Peso', value: `${v.max_weight_kg}kg` },
                    { label: 'Cubagem', value: `${v.max_cubage_m3}m³` },
                    { label: 'Paradas', value: v.max_stops },
                  ].map((s, i) => (
                    <div key={i} style={{ background: '#F5F7FF', borderRadius: 6, padding: '6px 8px', textAlign: 'center' }}>
                      <div style={{ fontSize: 13, fontWeight: 700 }}>{s.value}</div>
                      <div style={{ fontSize: 10, color: '#9E9E9E' }}>{s.label}</div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          {items.length > 0 && (
            <button onClick={() => setTab('distribute')} style={{
              width: '100%', padding: 14, borderRadius: 12,
              background: 'linear-gradient(135deg, #0A3D8F, #1565C0)',
              color: 'white', fontWeight: 700, fontSize: 15, border: 'none', cursor: 'pointer',
            }}>
              ⚖️ Distribuir {items.length} entregas pelos veículos →
            </button>
          )}
        </div>
      )}

      {/* TAB: Distribuir */}
      {tab === 'distribute' && (
        <div>
          <div style={{ background: '#FFF8E1', border: '1.5px solid #F9A825', borderRadius: 12, padding: 16, marginBottom: 20 }}>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#E65100' }}>⚖️ Distribuição por Peso + Cubagem + Região</p>
            <p style={{ fontSize: 13, color: '#757575', marginTop: 4 }}>
              O sistema vai sugerir a distribuição ideal. Você pode mover entregas entre veículos depois.
            </p>
          </div>

          <button onClick={handleDistribute} disabled={distributing || items.length === 0} style={{
            width: '100%', padding: 14, borderRadius: 12,
            background: distributing ? '#90CAF9' : 'linear-gradient(135deg, #1B5E20, #2E7D32)',
            color: 'white', fontWeight: 700, fontSize: 15, border: 'none', cursor: 'pointer', marginBottom: 20,
          }}>
            {distributing ? 'Distribuindo...' : `⚖️ Distribuir ${items.length} entregas automaticamente`}
          </button>
        </div>
      )}

      {/* TAB: Roteirizar */}
      {tab === 'route' && (
        <div>
          {assignments.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
              <Route size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
              <p style={{ fontSize: 15 }}>Distribua as entregas primeiro</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {assignments.map((a, i) => (
                <div key={i} style={{ background: 'white', borderRadius: 14, padding: 20, border: `1.5px solid ${a.over_weight || a.over_cubage ? '#EF9A9A' : '#E0E4F0'}` }}>
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <span style={{ fontSize: 24 }}>{VEHICLE_ICON[a.vehicle.type] || '🚛'}</span>
                      <div>
                        <p style={{ fontSize: 15, fontWeight: 700 }}>{a.vehicle.name}</p>
                        <p style={{ fontSize: 12, color: '#757575' }}>{a.total_stops} paradas • {a.total_weight.toFixed(0)}kg • {a.total_cubage.toFixed(2)}m³</p>
                      </div>
                    </div>
                    <button onClick={() => sendToRouter(a)} style={{
                      background: 'linear-gradient(135deg, #0A3D8F, #1565C0)',
                      color: 'white', padding: '8px 16px', borderRadius: 8,
                      border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 12,
                      display: 'flex', alignItems: 'center', gap: 6,
                    }}>
                      <Route size={14} /> Roteirizar
                    </button>
                  </div>

                  {/* Barras de uso */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 12 }}>
                    {[
                      { label: 'Peso', pct: a.weight_used_pct, over: a.over_weight },
                      { label: 'Cubagem', pct: a.cubage_used_pct, over: a.over_cubage },
                    ].map((b, j) => (
                      <div key={j}>
                        <div className="flex justify-between" style={{ fontSize: 11, color: '#757575', marginBottom: 4 }}>
                          <span>{b.label}</span><span style={{ fontWeight: 700, color: b.over ? '#B71C1C' : '#2E7D32' }}>{b.pct.toFixed(0)}%</span>
                        </div>
                        <div style={{ height: 6, background: '#F0F2F8', borderRadius: 3 }}>
                          <div style={{ height: '100%', background: b.over ? '#EF5350' : b.pct > 80 ? '#FF9800' : '#42A5F5', borderRadius: 3, width: `${Math.min(b.pct, 100)}%` }} />
                        </div>
                      </div>
                    ))}
                  </div>

                  {a.warnings?.map((w, j) => (
                    <p key={j} style={{ fontSize: 11, color: '#E65100', marginBottom: 4 }}>{w}</p>
                  ))}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
