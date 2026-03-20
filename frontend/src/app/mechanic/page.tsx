'use client'
import { useState, useEffect, useCallback } from 'react'
import { Plus, Search, Wrench, MessageCircle, Eye, Clock, CheckCircle, AlertCircle, XCircle, RefreshCw, Loader2 } from 'lucide-react'

interface ServiceOrder {
  id: string
  tenant_id: string
  number: string
  vehicle_plate: string
  vehicle_km: number
  vehicle_model: string
  vehicle_year: number
  customer_id: string
  customer_phone: string
  status: string
  complaint: string
  diagnosis: string
  created_at: string
  updated_at: string
}

const statusConfig: Record<string, { label: string; cls: string; icon: React.ReactNode }> = {
  open:           { label: 'Aberta',              cls: 'badge-open',     icon: <Clock size={10} /> },
  diagnosed:      { label: 'Diagnosticada',       cls: 'badge-pending',  icon: <AlertCircle size={10} /> },
  await_approval: { label: 'Aguard. Aprovacao',   cls: 'badge-pending',  icon: <MessageCircle size={10} /> },
  in_progress:    { label: 'Em andamento',         cls: 'badge-approved', icon: <Wrench size={10} /> },
  done:           { label: 'Concluida',            cls: 'badge-done',     icon: <CheckCircle size={10} /> },
  invoiced:       { label: 'Faturada',             cls: 'badge-done',     icon: <CheckCircle size={10} /> },
}

function getToken() {
  if (typeof window !== 'undefined') {
    return sessionStorage.getItem('access_token') || ''
  }
  return ''
}

async function apiFetch(path: string, opts: RequestInit = {}) {
  const token = getToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(opts.headers as Record<string, string> || {}) }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(path, { ...opts, headers })
  return res
}

export default function MechanicPage() {
  const [orders, setOrders] = useState<ServiceOrder[]>([])
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('all')
  const [showModal, setShowModal] = useState(false)
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState({ vehicle_plate: '', vehicle_km: '', vehicle_model: '', vehicle_year: '', customer_id: '', customer_phone: '', complaint: '' })

  const fetchOrders = useCallback(async () => {
    const token = getToken()
    if (!token) { window.location.href = '/login'; return }
    setLoading(true)
    try {
      const res = await apiFetch('/api/v1/mechanic/os')
      if (res.status === 401) { window.location.href = '/login'; return }
      if (res.ok) {
        const data = await res.json()
        setOrders(data.data || [])
      }
    } catch {} finally { setLoading(false) }
  }, [])

  useEffect(() => { fetchOrders() }, [fetchOrders])

  async function createOS() {
    if (!form.vehicle_plate || !form.customer_id || !form.complaint) return
    setCreating(true)
    try {
      const res = await apiFetch('/api/v1/mechanic/os', {
        method: 'POST',
        body: JSON.stringify({
          vehicle_plate: form.vehicle_plate.toUpperCase(),
          vehicle_km: parseInt(form.vehicle_km) || 0,
          vehicle_model: form.vehicle_model,
          vehicle_year: parseInt(form.vehicle_year) || 0,
          customer_id: form.customer_id,
          customer_phone: form.customer_phone,
          complaint: form.complaint,
        }),
      })
      if (res.ok) {
        setShowModal(false)
        setForm({ vehicle_plate: '', vehicle_km: '', vehicle_model: '', vehicle_year: '', customer_id: '', customer_phone: '', complaint: '' })
        fetchOrders()
      }
    } catch {} finally { setCreating(false) }
  }

  const filtered = orders.filter(os => {
    const matchSearch = os.vehicle_plate?.toLowerCase().includes(search.toLowerCase()) ||
      os.customer_id?.toLowerCase().includes(search.toLowerCase()) ||
      os.number?.includes(search)
    const matchFilter = filter === 'all' || os.status === filter
    return matchSearch && matchFilter
  })

  const stats = {
    open: orders.filter(o => o.status === 'open').length,
    in_progress: orders.filter(o => o.status === 'in_progress').length,
    await_approval: orders.filter(o => o.status === 'await_approval').length,
    done: orders.filter(o => o.status === 'done' || o.status === 'invoiced').length,
  }

  const formatDate = (d: string) => {
    try { return new Date(d).toLocaleString('pt-BR', { day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit' }) } catch { return d }
  }

  return (
    <div className="space-y-5 animate-fade-in">
      {/* Stats row */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: 'Abertas', value: stats.open, color: 'text-nexo-600', bg: 'bg-nexo-50' },
          { label: 'Em andamento', value: stats.in_progress, color: 'text-emerald-600', bg: 'bg-emerald-50' },
          { label: 'Aguard. aprovacao', value: stats.await_approval, color: 'text-amber-600', bg: 'bg-amber-50' },
          { label: 'Concluidas', value: stats.done, color: 'text-slate-600', bg: 'bg-slate-50' },
        ].map(s => (
          <div key={s.label} className={`card p-4 flex items-center gap-3`} data-testid={`stat-${s.label}`}>
            <div className={`w-10 h-10 ${s.bg} rounded-xl flex items-center justify-center`}>
              <span className={`text-lg font-display font-700 ${s.color}`}>{s.value}</span>
            </div>
            <p className="text-xs font-medium text-slate-500">{s.label}</p>
          </div>
        ))}
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl flex-1 max-w-sm">
          <Search size={14} className="text-slate-400" />
          <input
            data-testid="os-search"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Placa, cliente ou numero..."
            className="bg-transparent text-sm outline-none flex-1 text-slate-700 placeholder-slate-400"
          />
        </div>
        <div className="flex items-center gap-1 bg-white border border-slate-200 rounded-xl p-1">
          {['all', 'open', 'await_approval', 'in_progress', 'done'].map(s => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                filter === s ? 'bg-nexo-500 text-white shadow-nexo-sm' : 'text-slate-500 hover:bg-slate-50'
              }`}
            >
              {s === 'all' ? 'Todas' : statusConfig[s]?.label ?? s}
            </button>
          ))}
        </div>
        <button onClick={fetchOrders} className="btn-ghost" title="Atualizar">
          <RefreshCw size={15} className={loading ? 'animate-spin' : ''} />
        </button>
        <button data-testid="new-os-btn" onClick={() => setShowModal(true)} className="btn-primary ml-auto">
          <Plus size={15} /> Nova OS
        </button>
      </div>

      {/* OS Table */}
      <div className="card overflow-hidden">
        <table className="w-full" data-testid="os-table">
          <thead className="bg-slate-50 border-b border-slate-100">
            <tr>
              <th className="table-header">Numero</th>
              <th className="table-header">Placa / Modelo</th>
              <th className="table-header">Cliente</th>
              <th className="table-header">Reclamacao</th>
              <th className="table-header">Status</th>
              <th className="table-header">KM</th>
              <th className="table-header text-center">Acoes</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr><td colSpan={7} className="py-8 text-center"><Loader2 size={24} className="text-nexo-500 animate-spin mx-auto" /></td></tr>
            )}
            {!loading && filtered.map((os) => {
              const s = statusConfig[os.status] || statusConfig.open
              return (
                <tr key={os.id} className="hover:bg-slate-50 transition-colors group" data-testid={`os-row-${os.id}`}>
                  <td className="table-cell">
                    <span className="font-mono text-xs text-nexo-600 font-medium">{os.number}</span>
                    <p className="text-[10px] text-slate-400 mt-0.5">{formatDate(os.created_at)}</p>
                  </td>
                  <td className="table-cell">
                    <span className="font-mono font-bold text-slate-800 text-sm">{os.vehicle_plate}</span>
                    <p className="text-xs text-slate-400">{os.vehicle_model} {os.vehicle_year > 0 ? os.vehicle_year : ''}</p>
                  </td>
                  <td className="table-cell">
                    <p className="font-medium text-slate-700">{os.customer_id}</p>
                    <p className="text-xs text-slate-400">{os.customer_phone}</p>
                  </td>
                  <td className="table-cell max-w-[200px]">
                    <p className="text-sm text-slate-600 truncate">{os.complaint}</p>
                  </td>
                  <td className="table-cell">
                    <span className={s.cls}>{s.icon}{s.label}</span>
                  </td>
                  <td className="table-cell font-mono text-xs text-slate-500">
                    {os.vehicle_km > 0 ? `${os.vehicle_km.toLocaleString('pt-BR')} km` : '-'}
                  </td>
                  <td className="table-cell text-center">
                    <div className="flex items-center justify-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button className="w-7 h-7 flex items-center justify-center rounded-lg bg-nexo-50 text-nexo-600 hover:bg-nexo-100 transition-colors">
                        <Eye size={13} />
                      </button>
                      {os.status === 'diagnosed' && (
                        <button className="w-7 h-7 flex items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 hover:bg-emerald-100 transition-colors" title="Enviar aprovacao WhatsApp">
                          <MessageCircle size={13} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
        {!loading && filtered.length === 0 && (
          <div className="py-12 text-center">
            <Wrench size={32} className="text-slate-200 mx-auto mb-3" />
            <p className="text-sm text-slate-400">Nenhuma OS encontrada</p>
            <button onClick={() => setShowModal(true)} className="btn-primary mt-4 mx-auto">
              <Plus size={14} /> Criar primeira OS
            </button>
          </div>
        )}
      </div>

      {/* New OS Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-nexo-xl w-full max-w-lg animate-slide-up">
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
              <h2 className="font-semibold text-slate-800">Nova Ordem de Servico</h2>
              <button onClick={() => setShowModal(false)} className="text-slate-400 hover:text-slate-600">
                <XCircle size={20} />
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Placa do Veiculo *</label>
                  <input data-testid="input-plate" className="input font-mono uppercase" placeholder="BRA2E19" value={form.vehicle_plate} onChange={e => setForm({ ...form, vehicle_plate: e.target.value })} />
                </div>
                <div>
                  <label className="label">KM Atual</label>
                  <input data-testid="input-km" className="input" type="number" placeholder="45200" value={form.vehicle_km} onChange={e => setForm({ ...form, vehicle_km: e.target.value })} />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Modelo do Veiculo</label>
                  <input data-testid="input-model" className="input" placeholder="Civic 2021" value={form.vehicle_model} onChange={e => setForm({ ...form, vehicle_model: e.target.value })} />
                </div>
                <div>
                  <label className="label">Ano</label>
                  <input data-testid="input-year" className="input" type="number" placeholder="2022" value={form.vehicle_year} onChange={e => setForm({ ...form, vehicle_year: e.target.value })} />
                </div>
              </div>
              <div>
                <label className="label">ID do Cliente *</label>
                <input data-testid="input-customer" className="input" placeholder="Nome ou ID do cliente" value={form.customer_id} onChange={e => setForm({ ...form, customer_id: e.target.value })} />
              </div>
              <div>
                <label className="label">WhatsApp do Cliente</label>
                <input data-testid="input-phone" className="input" placeholder="5511999999999" value={form.customer_phone} onChange={e => setForm({ ...form, customer_phone: e.target.value })} />
              </div>
              <div>
                <label className="label">Reclamacao do Cliente *</label>
                <textarea data-testid="input-complaint" className="input resize-none" rows={3} placeholder="Descreva o problema relatado..." value={form.complaint} onChange={e => setForm({ ...form, complaint: e.target.value })} />
              </div>
            </div>
            <div className="flex gap-3 px-6 py-4 border-t border-slate-100">
              <button onClick={() => setShowModal(false)} className="btn-ghost flex-1">Cancelar</button>
              <button data-testid="create-os-btn" onClick={createOS} disabled={creating || !form.vehicle_plate || !form.customer_id || !form.complaint} className="btn-primary flex-1 disabled:opacity-50">
                {creating ? <Loader2 size={15} className="animate-spin" /> : <Plus size={15} />}
                {creating ? 'Criando...' : 'Criar OS'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
