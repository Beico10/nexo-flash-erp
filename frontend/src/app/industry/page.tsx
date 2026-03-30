'use client'
import { isDemoMode, promptLogin } from '@/lib/demo'
import { useState, useEffect } from 'react'
import { Factory, Play, Pause, CheckCircle2, Clock, Package, AlertTriangle } from 'lucide-react'

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

interface Order { id: string; number: string; product_name: string; planned_qty: number; produced_qty: number; status: string; planned_start: string; planned_end: string; progress: number }
interface BOM { id: string; product_name: string; version: number; total_cost: number; items_count: number }

const statusConfig: Record<string, { label: string; color: string; icon: any }> = {
  planned: { label: 'Planejada', color: 'bg-blue-100 text-blue-700', icon: Clock },
  in_progress: { label: 'Em Producao', color: 'bg-amber-100 text-amber-700', icon: Play },
  paused: { label: 'Pausada', color: 'bg-gray-100 text-gray-700', icon: Pause },
  done: { label: 'Concluida', color: 'bg-green-100 text-green-700', icon: CheckCircle2 },
}

export default function IndustryPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [boms, setBoms] = useState<BOM[]>([])
  const [materials, setMaterials] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    if (isDemoMode()) { setOrders([{ id: "1", number: "OP-001", product_name: "Cadeira Escritorio Premium", planned_qty: 50, produced_qty: 32, status: "in_progress", planned_start: new Date().toISOString(), planned_end: new Date().toISOString(), progress: 64 },{ id: "2", number: "OP-002", product_name: "Mesa de Reuniao", planned_qty: 20, produced_qty: 20, status: "done", planned_start: new Date().toISOString(), planned_end: new Date().toISOString(), progress: 100 },{ id: "3", number: "OP-003", product_name: "Estante Modular", planned_qty: 30, produced_qty: 0, status: "planned", planned_start: new Date().toISOString(), planned_end: new Date().toISOString(), progress: 0 }]); setBoms([{ id: "1", product_name: "Cadeira Escritorio Premium", version: 2, total_cost: 380.50, items_count: 12 },{ id: "2", product_name: "Mesa de Reuniao", version: 1, total_cost: 620.00, items_count: 8 }]); setMaterials({ "mat-aco-carbono": 245, "mat-espuma-d33": 38, "mat-tecido-mesh": 120, "mat-parafuso-m6": 500 }); setLoading(false); return } if (!token) { promptLogin(); setLoading(false); return }
    const h = { Authorization: `Bearer ${token}` }
    Promise.all([
      fetch('/api/v1/industry/orders', { headers: h }).then(r => r.json()),
      fetch('/api/v1/industry/boms', { headers: h }).then(r => r.json()),
      fetch('/api/v1/industry/materials', { headers: h }).then(r => r.json()),
    ]).then(([o, b, m]) => {
      setOrders(o.orders || [])
      setBoms(b.boms || [])
      setMaterials(m.materials || {})
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  if (loading) return <div className="flex items-center justify-center min-h-[400px]"><div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" /></div>

  return (
    <div className="max-w-6xl mx-auto py-6 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900" data-testid="industry-title">Industria - PCP</h1>
          <p className="text-sm text-gray-500">Planejamento e Controle de Producao</p>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Ordens Ativas</p><p className="text-2xl font-bold text-gray-900" data-testid="active-orders">{orders.filter(o => o.status === 'in_progress').length}</p></div>
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">BOMs Cadastradas</p><p className="text-2xl font-bold text-gray-900">{boms.length}</p></div>
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Materiais Rastreados</p><p className="text-2xl font-bold text-gray-900">{Object.keys(materials).length}</p></div>
      </div>

      <div className="bg-white rounded-xl border">
        <div className="px-5 py-3 border-b"><h2 className="font-semibold text-gray-900">Ordens de Producao</h2></div>
        <div className="divide-y">
          {orders.map(o => {
            const cfg = statusConfig[o.status] || statusConfig.planned
            const Icon = cfg.icon
            return (
              <div key={o.id} data-testid={`order-${o.id}`} className="px-5 py-4 flex items-center gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs text-gray-400">{o.number}</span>
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium flex items-center gap-1 ${cfg.color}`}><Icon size={11} />{cfg.label}</span>
                  </div>
                  <p className="font-medium text-gray-900 mt-0.5">{o.product_name}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-mono">{o.produced_qty}/{o.planned_qty} {o.status === 'in_progress' ? 'un' : ''}</p>
                  <div className="w-32 h-1.5 bg-gray-100 rounded-full mt-1 overflow-hidden">
                    <div className="h-full bg-blue-500 rounded-full" style={{ width: `${o.progress}%` }} />
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-white rounded-xl border">
          <div className="px-5 py-3 border-b"><h2 className="font-semibold text-gray-900">Fichas Tecnicas (BOM)</h2></div>
          <div className="divide-y">
            {boms.map(b => (
              <div key={b.id} className="px-5 py-3 flex justify-between items-center">
                <div><p className="font-medium text-sm">{b.product_name}</p><p className="text-xs text-gray-400">v{b.version} - {b.items_count} componentes</p></div>
                <p className="font-mono text-sm">R$ {b.total_cost.toFixed(2)}</p>
              </div>
            ))}
          </div>
        </div>
        <div className="bg-white rounded-xl border">
          <div className="px-5 py-3 border-b flex items-center gap-2"><Package size={16} /><h2 className="font-semibold text-gray-900">Estoque de Materiais</h2></div>
          <div className="divide-y">
            {Object.entries(materials).map(([k, v]) => (
              <div key={k} className="px-5 py-3 flex justify-between items-center">
                <p className="text-sm">{k.replace('mat-', '').replace('-', ' ')}</p>
                <div className="flex items-center gap-2">
                  {v < 50 && <AlertTriangle size={12} className="text-amber-500" />}
                  <span className={`font-mono text-sm ${v < 50 ? 'text-amber-600 font-bold' : ''}`}>{v}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

