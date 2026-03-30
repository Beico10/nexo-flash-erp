'use client'
import { useState, useEffect } from 'react'
import { Calendar, Clock, CheckCircle, User, Scissors, Plus } from 'lucide-react'
import { isDemoMode, promptLogin, DEMO_AESTHETICS_APPOINTMENTS } from '@/lib/demo'

interface Appointment {
  ID: string; ProfessionalID: string; CustomerName: string; ServiceName: string
  ServicePrice: number; StartTime: string; EndTime: string; DurationMin: number; Status: string
}

const statusConfig: Record<string, { label: string; color: string; bg: string }> = {
  confirmed:   { label: 'Confirmado',   color: '#059669', bg: '#ECFDF5' },
  scheduled:   { label: 'Agendado',     color: '#D97706', bg: '#FFFBEB' },
  in_progress: { label: 'Em atendimento', color: '#7C3AED', bg: '#F5F3FF' },
  completed:   { label: 'Concluido',    color: '#64748b', bg: '#F8FAFC' },
}

const professionals: Record<string, string> = {
  'prof-1': 'Ana Lima',
  'prof-2': 'Carla Souza',
  'prof-3': 'Julia Mendes',
}

export default function AestheticsPage() {
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (isDemoMode()) {
      setAppointments(DEMO_AESTHETICS_APPOINTMENTS as Appointment[])
      setLoading(false)
      return
    }
    const token = typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : ''
    if (!token) { promptLogin(); setLoading(false); return }
    fetch('/api/v1/aesthetics/appointments', { headers: { Authorization: 'Bearer ' + token } })
      .then(r => r.json()).then(d => { setAppointments(d.data || []); setLoading(false) })
      .catch(() => setLoading(false))
  }, [])

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const fmtTime = (s: string) => { try { return new Date(s).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }) } catch { return s } }

  const total = appointments.reduce((a, b) => a + b.ServicePrice, 0)
  const concluidos = appointments.filter(a => a.Status === 'completed').length
  const emAtendimento = appointments.filter(a => a.Status === 'in_progress').length

  if (loading) return null

  return (
    <div className="space-y-5">
      {isDemoMode() && (
        <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: 12, padding: '10px 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: 13, fontWeight: 600, color: '#92400E' }}>Demonstracao de Estetica</span>
            <span style={{ fontSize: 12, color: '#64748b' }}>Explore a vontade</span>
          </div>
          <button onClick={promptLogin} style={{ background: '#D97706', color: 'white', border: 'none', borderRadius: 8, padding: '6px 14px', fontSize: 12, fontWeight: 600, cursor: 'pointer' }}>Criar conta gratis</button>
        </div>
      )}

      <div className="grid grid-cols-3 gap-4">
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-purple-50 rounded-xl flex items-center justify-center"><Calendar size={18} className="text-purple-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{appointments.length}</p><p className="text-xs text-slate-400">Agendamentos hoje</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center"><CheckCircle size={18} className="text-emerald-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{concluidos}</p><p className="text-xs text-slate-400">Concluidos</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-50 rounded-xl flex items-center justify-center"><Scissors size={18} className="text-blue-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{fmt(total)}</p><p className="text-xs text-slate-400">Faturamento do dia</p></div>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-slate-700">Agenda de hoje</h2>
        <button onClick={() => isDemoMode() ? promptLogin() : undefined} className="btn-primary"><Plus size={14} /> Novo agendamento</button>
      </div>

      <div className="card overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-50 border-b border-slate-100">
            <tr>
              <th className="table-header">Horario</th>
              <th className="table-header">Cliente</th>
              <th className="table-header">Servico</th>
              <th className="table-header">Profissional</th>
              <th className="table-header">Duracao</th>
              <th className="table-header">Valor</th>
              <th className="table-header">Status</th>
            </tr>
          </thead>
          <tbody>
            {appointments.map(a => {
              const s = statusConfig[a.Status] || statusConfig.scheduled
              return (
                <tr key={a.ID} className="hover:bg-slate-50 transition-colors">
                  <td className="table-cell">
                    <div className="flex items-center gap-1">
                      <Clock size={12} className="text-slate-400" />
                      <span className="font-mono text-xs font-medium text-slate-700">{fmtTime(a.StartTime)}</span>
                    </div>
                  </td>
                  <td className="table-cell">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 bg-purple-100 rounded-full flex items-center justify-center">
                        <User size={11} className="text-purple-600" />
                      </div>
                      <span className="font-medium text-slate-700">{a.CustomerName}</span>
                    </div>
                  </td>
                  <td className="table-cell"><span className="text-sm text-slate-600">{a.ServiceName}</span></td>
                  <td className="table-cell"><span className="text-sm text-slate-500">{professionals[a.ProfessionalID] || a.ProfessionalID}</span></td>
                  <td className="table-cell"><span className="text-xs text-slate-500">{a.DurationMin} min</span></td>
                  <td className="table-cell"><span className="font-semibold text-slate-700">{fmt(a.ServicePrice)}</span></td>
                  <td className="table-cell">
                    <span style={{ background: s.bg, color: s.color, fontSize: 11, fontWeight: 600, padding: '3px 10px', borderRadius: 100 }}>{s.label}</span>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}