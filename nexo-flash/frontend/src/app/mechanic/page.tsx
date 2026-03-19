'use client'
import { useState } from 'react'
import { Plus, Search, Wrench, MessageCircle, Eye, Clock, CheckCircle, AlertCircle, XCircle } from 'lucide-react'

const mockOS = [
  { id: 'OS-2026-001842', plate: 'BRA2E19', model: 'Civic 2021', km: 45200, customer: 'Carlos Silva', phone: '11999887766', status: 'await_approval', complaint: 'Barulho no freio dianteiro', total: 850, createdAt: '2026-03-19 08:30' },
  { id: 'OS-2026-001841', plate: 'ABC1D23', model: 'HB20 2022',  km: 32100, customer: 'Maria Santos',  phone: '11988776655', status: 'in_progress',    complaint: 'Troca de correia dentada',  total: 1200, createdAt: '2026-03-19 07:15' },
  { id: 'OS-2026-001840', plate: 'XYZ9K87', model: 'Gol 2019',   km: 78500, customer: 'João Lima',    phone: '11977665544', status: 'done',            complaint: 'Revisão 80.000 km',       total: 430,  createdAt: '2026-03-18 14:00' },
  { id: 'OS-2026-001839', plate: 'DEF4G56', model: 'Onix 2023',  km: 12300, customer: 'Ana Costa',    phone: '11966554433', status: 'open',            complaint: 'Luz do motor acesa',      total: 0,    createdAt: '2026-03-19 09:45' },
  { id: 'OS-2026-001838', plate: 'GHI7J89', model: 'Fox 2018',   km: 95000, customer: 'Pedro Rocha',  phone: '11955443322', status: 'diagnosed',       complaint: 'Suspensão dianteira',     total: 2100, createdAt: '2026-03-18 11:00' },
]

const statusConfig: Record<string, { label: string; cls: string; icon: React.ReactNode }> = {
  open:           { label: 'Aberta',              cls: 'badge-open',    icon: <Clock size={10} /> },
  diagnosed:      { label: 'Diagnosticada',       cls: 'badge-pending', icon: <AlertCircle size={10} /> },
  await_approval: { label: 'Aguard. Aprovação',   cls: 'badge-pending', icon: <MessageCircle size={10} /> },
  in_progress:    { label: 'Em andamento',         cls: 'badge-approved',icon: <Wrench size={10} /> },
  done:           { label: 'Concluída',            cls: 'badge-done',    icon: <CheckCircle size={10} /> },
  invoiced:       { label: 'Faturada',             cls: 'badge-done',    icon: <CheckCircle size={10} /> },
}

export default function MechanicPage() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('all')
  const [showModal, setShowModal] = useState(false)

  const filtered = mockOS.filter(os => {
    const matchSearch = os.plate.includes(search.toUpperCase()) ||
      os.customer.toLowerCase().includes(search.toLowerCase()) ||
      os.id.includes(search)
    const matchFilter = filter === 'all' || os.status === filter
    return matchSearch && matchFilter
  })

  return (
    <div className="space-y-5 animate-fade-in">

      {/* Stats row */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: 'Abertas', value: '4', color: 'text-nexo-600', bg: 'bg-nexo-50' },
          { label: 'Em andamento', value: '2', color: 'text-emerald-600', bg: 'bg-emerald-50' },
          { label: 'Aguard. aprovação', value: '1', color: 'text-amber-600', bg: 'bg-amber-50' },
          { label: 'Concluídas hoje', value: '3', color: 'text-slate-600', bg: 'bg-slate-50' },
        ].map(s => (
          <div key={s.label} className={`card p-4 flex items-center gap-3`}>
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
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="Placa, cliente ou número..."
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
        <button onClick={() => setShowModal(true)} className="btn-primary ml-auto">
          <Plus size={15} />
          Nova OS
        </button>
      </div>

      {/* OS Table */}
      <div className="card overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-50 border-b border-slate-100">
            <tr>
              <th className="table-header">Número</th>
              <th className="table-header">Placa / Modelo</th>
              <th className="table-header">Cliente</th>
              <th className="table-header">Reclamação</th>
              <th className="table-header">Status</th>
              <th className="table-header">KM</th>
              <th className="table-header text-right">Valor</th>
              <th className="table-header text-center">Ações</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((os) => {
              const s = statusConfig[os.status]
              return (
                <tr key={os.id} className="hover:bg-slate-50 transition-colors group">
                  <td className="table-cell">
                    <span className="font-mono text-xs text-nexo-600 font-medium">{os.id}</span>
                    <p className="text-[10px] text-slate-400 mt-0.5">{os.createdAt}</p>
                  </td>
                  <td className="table-cell">
                    <span className="font-mono font-bold text-slate-800 text-sm">{os.plate}</span>
                    <p className="text-xs text-slate-400">{os.model}</p>
                  </td>
                  <td className="table-cell">
                    <p className="font-medium text-slate-700">{os.customer}</p>
                    <p className="text-xs text-slate-400">{os.phone}</p>
                  </td>
                  <td className="table-cell max-w-[200px]">
                    <p className="text-sm text-slate-600 truncate">{os.complaint}</p>
                  </td>
                  <td className="table-cell">
                    <span className={s.cls}>{s.icon}{s.label}</span>
                  </td>
                  <td className="table-cell font-mono text-xs text-slate-500">
                    {os.km.toLocaleString('pt-BR')} km
                  </td>
                  <td className="table-cell text-right font-semibold text-slate-800">
                    {os.total > 0 ? `R$ ${os.total.toLocaleString('pt-BR')}` : '—'}
                  </td>
                  <td className="table-cell text-center">
                    <div className="flex items-center justify-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button className="w-7 h-7 flex items-center justify-center rounded-lg bg-nexo-50 text-nexo-600 hover:bg-nexo-100 transition-colors">
                        <Eye size={13} />
                      </button>
                      {os.status === 'diagnosed' && (
                        <button className="w-7 h-7 flex items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 hover:bg-emerald-100 transition-colors" title="Enviar aprovação WhatsApp">
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
        {filtered.length === 0 && (
          <div className="py-12 text-center">
            <Wrench size={32} className="text-slate-200 mx-auto mb-3" />
            <p className="text-sm text-slate-400">Nenhuma OS encontrada</p>
          </div>
        )}
      </div>

      {/* New OS Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-nexo-xl w-full max-w-lg animate-slide-up">
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
              <h2 className="font-semibold text-slate-800">Nova Ordem de Serviço</h2>
              <button onClick={() => setShowModal(false)} className="text-slate-400 hover:text-slate-600">
                <XCircle size={20} />
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Placa do Veículo *</label>
                  <input className="input font-mono uppercase" placeholder="BRA2E19" />
                </div>
                <div>
                  <label className="label">KM Atual</label>
                  <input className="input" type="number" placeholder="45200" />
                </div>
              </div>
              <div>
                <label className="label">Modelo do Veículo</label>
                <input className="input" placeholder="Civic 2021" />
              </div>
              <div>
                <label className="label">Cliente</label>
                <input className="input" placeholder="Nome do cliente" />
              </div>
              <div>
                <label className="label">WhatsApp do Cliente</label>
                <input className="input" placeholder="(11) 99999-9999" />
              </div>
              <div>
                <label className="label">Reclamação do Cliente *</label>
                <textarea className="input resize-none" rows={3} placeholder="Descreva o problema relatado..." />
              </div>
            </div>
            <div className="flex gap-3 px-6 py-4 border-t border-slate-100">
              <button onClick={() => setShowModal(false)} className="btn-ghost flex-1">Cancelar</button>
              <button className="btn-primary flex-1">Criar OS</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
