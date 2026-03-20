'use client'
import { useState, useEffect } from 'react'
import { Scissors, Clock, CheckCircle, User, Loader2 } from 'lucide-react'

interface Appointment {
  ID: string; ProfessionalID: string; CustomerName: string; ServiceName: string; ServicePrice: number; StartTime: string; EndTime: string; DurationMin: number; Status: string
}

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

const statusStyles: Record<string, { label: string; cls: string }> = {
  scheduled:    { label: 'Agendado',     cls: 'bg-blue-50 text-blue-600 border-blue-100' },
  confirmed:    { label: 'Confirmado',   cls: 'bg-emerald-50 text-emerald-600 border-emerald-100' },
  in_progress:  { label: 'Em atendimento', cls: 'bg-violet-50 text-violet-600 border-violet-100' },
  completed:    { label: 'Concluido',    cls: 'bg-slate-50 text-slate-600 border-slate-100' },
  no_show:      { label: 'Nao compareceu', cls: 'bg-red-50 text-red-600 border-red-100' },
}

const profNames: Record<string, string> = { 'prof-1': 'Ana Beatriz', 'prof-2': 'Carla Souza', 'prof-3': 'Daniela Lima' }
const profColors: Record<string, string> = { 'prof-1': 'bg-pink-500', 'prof-2': 'bg-violet-500', 'prof-3': 'bg-blue-500' }

export default function AestheticsPage() {
  const [appointments, setAppointments] = useState<Appointment[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    fetch('/api/v1/aesthetics/appointments', { headers: { Authorization: `Bearer ${token}` } })
      .then(r => r.json())
      .then(d => { setAppointments(d.data || []); setLoading(false) })
      .catch(() => setLoading(false))
  }, [])

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const formatTime = (s: string) => { try { return new Date(s).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' }) } catch { return '' } }

  const professionals = [...new Set(appointments.map(a => a.ProfessionalID))]
  const totalRevenue = appointments.reduce((sum, a) => sum + a.ServicePrice, 0)

  if (loading) return <div className="flex items-center justify-center h-64"><Loader2 size={32} className="text-nexo-500 animate-spin" /></div>

  return (
    <div className="space-y-5 animate-fade-in" data-testid="aesthetics-page">
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-pink-50 rounded-xl flex items-center justify-center"><Scissors size={18} className="text-pink-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{appointments.length}</p><p className="text-xs text-slate-400">Agendamentos hoje</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-violet-50 rounded-xl flex items-center justify-center"><User size={18} className="text-violet-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{professionals.length}</p><p className="text-xs text-slate-400">Profissionais ativos</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center"><CheckCircle size={18} className="text-emerald-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{fmt(totalRevenue)}</p><p className="text-xs text-slate-400">Receita prevista</p></div>
        </div>
      </div>

      {/* Timeline by professional */}
      <div className="space-y-4">
        {professionals.map(profId => {
          const profApts = appointments.filter(a => a.ProfessionalID === profId).sort((a, b) => new Date(a.StartTime).getTime() - new Date(b.StartTime).getTime())
          return (
            <div key={profId} className="card p-4">
              <div className="flex items-center gap-3 mb-3">
                <div className={`w-8 h-8 rounded-full ${profColors[profId] || 'bg-slate-400'} flex items-center justify-center text-white text-xs font-bold`}>
                  {(profNames[profId] || profId).charAt(0)}
                </div>
                <span className="font-semibold text-slate-800 text-sm">{profNames[profId] || profId}</span>
                <span className="text-xs text-slate-400">{profApts.length} agendamentos</span>
              </div>
              <div className="flex gap-2 overflow-x-auto pb-1">
                {profApts.map(a => {
                  const st = statusStyles[a.Status] || statusStyles.scheduled
                  return (
                    <div key={a.ID} className={`flex-shrink-0 rounded-xl p-3 border ${st.cls}`} style={{ minWidth: 180 }} data-testid={`apt-${a.ID}`}>
                      <div className="flex items-center gap-1.5 mb-1">
                        <Clock size={11} />
                        <span className="text-xs font-bold">{formatTime(a.StartTime)} - {formatTime(a.EndTime)}</span>
                      </div>
                      <p className="text-sm font-medium mb-0.5">{a.ServiceName}</p>
                      <p className="text-xs opacity-70">{a.CustomerName}</p>
                      <div className="flex items-center justify-between mt-2">
                        <span className="text-[10px] font-bold uppercase">{st.label}</span>
                        <span className="text-xs font-bold">{fmt(a.ServicePrice)}</span>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
