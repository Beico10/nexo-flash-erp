'use client'
import { useState, useEffect } from 'react'
import { BarChart3, TrendingUp, Wrench, Clock, AlertTriangle, CheckCircle, Brain, Wheat, Scissors, ArrowUpRight, ArrowDownRight, Loader2 } from 'lucide-react'

interface DashboardStats {
  mechanic_os: { total: number; open: number; in_progress: number; await_approval: number; done: number }
  bakery_products: number
  appointments: number
  pending_suggestions: number
  revenue: { today: number; week: number; chart: { day: string; revenue: number; tax: number }[]; by_module: { module: string; count: number }[] }
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : ''
    if (!token) { window.location.href = '/login'; return }
    fetch('/api/v1/dashboard/stats', { headers: { Authorization: `Bearer ${token}` } })
      .then(r => { if (r.status === 401) { window.location.href = '/login'; return null }; return r.json() })
      .then(d => { if (d) { setStats(d); setLoading(false) } })
      .catch(() => setLoading(false))
  }, [])

  if (loading) return <div className="flex items-center justify-center h-64"><Loader2 size={32} className="text-nexo-500 animate-spin" /></div>
  if (!stats) return <div className="text-center py-16 text-slate-400">Falha ao carregar dados</div>

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  const maxRev = Math.max(...stats.revenue.chart.map(d => d.revenue), 1)

  return (
    <div className="space-y-5 animate-fade-in" data-testid="dashboard-page">
      {/* KPI Row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <KPICard icon={<BarChart3 size={18} />} label="Faturamento Hoje" value={fmt(stats.revenue.today)} sub={`Semana: ${fmt(stats.revenue.week)}`} trend={+12.5} color="nexo" />
        <KPICard icon={<Wrench size={18} />} label="OS Abertas" value={String(stats.mechanic_os.open + stats.mechanic_os.in_progress)} sub={`${stats.mechanic_os.total} total`} trend={-5} color="amber" />
        <KPICard icon={<Brain size={18} />} label="Sugestoes IA" value={String(stats.pending_suggestions)} sub="aguardando aprovacao" color="violet" />
        <KPICard icon={<Scissors size={18} />} label="Agendamentos Hoje" value={String(stats.appointments)} sub={`${stats.bakery_products} produtos padaria`} trend={+8} color="emerald" />
      </div>

      <div className="grid lg:grid-cols-3 gap-5">
        {/* Revenue Chart */}
        <div className="lg:col-span-2 card p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold text-slate-800 text-sm">Faturamento Semanal</h3>
            <span className="text-xs text-slate-400">Receita vs Impostos</span>
          </div>
          <div className="flex items-end gap-2 h-44">
            {stats.revenue.chart.map((d, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-1">
                <div className="w-full flex flex-col items-center gap-0.5" style={{ height: 160 }}>
                  <div className="w-full rounded-t-lg bg-nexo-500 transition-all" style={{ height: `${(d.revenue / maxRev) * 100}%`, minHeight: 4 }} />
                  <div className="w-full rounded-b-lg bg-red-400/40 transition-all" style={{ height: `${(d.tax / maxRev) * 100}%`, minHeight: 2 }} />
                </div>
                <span className="text-[10px] text-slate-400 font-medium">{d.day}</span>
              </div>
            ))}
          </div>
          <div className="flex items-center gap-4 mt-3 pt-3 border-t border-slate-100">
            <div className="flex items-center gap-1.5 text-xs text-slate-500"><div className="w-2.5 h-2.5 rounded bg-nexo-500" /> Receita</div>
            <div className="flex items-center gap-1.5 text-xs text-slate-500"><div className="w-2.5 h-2.5 rounded bg-red-400/40" /> Impostos</div>
          </div>
        </div>

        {/* Mechanic OS Status */}
        <div className="card p-5">
          <h3 className="font-semibold text-slate-800 text-sm mb-4">Status das Ordens</h3>
          <div className="space-y-3">
            <StatusRow icon={<Clock size={13} />} label="Abertas" value={stats.mechanic_os.open} total={stats.mechanic_os.total} color="#3B82F6" />
            <StatusRow icon={<AlertTriangle size={13} />} label="Aguard. Aprovacao" value={stats.mechanic_os.await_approval} total={stats.mechanic_os.total} color="#F59E0B" />
            <StatusRow icon={<Wrench size={13} />} label="Em Andamento" value={stats.mechanic_os.in_progress} total={stats.mechanic_os.total} color="#8B5CF6" />
            <StatusRow icon={<CheckCircle size={13} />} label="Concluidas" value={stats.mechanic_os.done} total={stats.mechanic_os.total} color="#10B981" />
          </div>
          <div className="mt-4 pt-3 border-t border-slate-100">
            <div className="flex items-center justify-between">
              <span className="text-xs text-slate-400">Total de OS</span>
              <span className="text-lg font-bold text-slate-800">{stats.mechanic_os.total}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Module Activity */}
      <div className="grid lg:grid-cols-3 gap-5">
        <div className="card p-5 lg:col-span-2">
          <h3 className="font-semibold text-slate-800 text-sm mb-3">Atividade por Modulo</h3>
          <div className="space-y-2">
            {stats.revenue.by_module.map((m, i) => {
              const icons = [<Wrench size={14} key="w" />, <Wheat size={14} key="wh" />, <Scissors size={14} key="s" />]
              const colors = ['bg-nexo-500', 'bg-amber-500', 'bg-pink-500']
              const maxCount = Math.max(...stats.revenue.by_module.map(x => x.count), 1)
              return (
                <div key={i} className="flex items-center gap-3">
                  <div className={`w-8 h-8 rounded-lg ${colors[i] || 'bg-slate-400'} flex items-center justify-center text-white`}>
                    {icons[i]}
                  </div>
                  <span className="text-sm text-slate-700 w-24 font-medium">{m.module}</span>
                  <div className="flex-1 h-6 rounded-lg bg-slate-100 overflow-hidden">
                    <div className={`h-full rounded-lg ${colors[i] || 'bg-slate-400'} transition-all flex items-center px-2`} style={{ width: `${(m.count / maxCount) * 100}%`, minWidth: 30 }}>
                      <span className="text-[10px] font-bold text-white">{m.count}</span>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
        <div className="card p-5">
          <h3 className="font-semibold text-slate-800 text-sm mb-3">IA Co-Piloto</h3>
          <div className="text-center py-4">
            <div className="w-14 h-14 mx-auto rounded-2xl bg-violet-50 flex items-center justify-center mb-3">
              <Brain size={24} className="text-violet-500" />
            </div>
            <p className="text-2xl font-bold text-slate-800">{stats.pending_suggestions}</p>
            <p className="text-xs text-slate-400 mt-1">sugestoes pendentes</p>
            <a href="/ai-approvals" className="btn-primary mt-4 text-xs inline-flex">Ver sugestoes</a>
          </div>
        </div>
      </div>
    </div>
  )
}

function KPICard({ icon, label, value, sub, trend, color }: { icon: React.ReactNode; label: string; value: string; sub: string; trend?: number; color: string }) {
  const colorMap: Record<string, string> = { nexo: 'bg-nexo-50 text-nexo-600', amber: 'bg-amber-50 text-amber-600', violet: 'bg-violet-50 text-violet-600', emerald: 'bg-emerald-50 text-emerald-600' }
  return (
    <div className="card p-4">
      <div className="flex items-center justify-between mb-2">
        <div className={`w-9 h-9 rounded-xl ${colorMap[color]} flex items-center justify-center`}>{icon}</div>
        {trend !== undefined && (
          <span className={`text-xs font-semibold flex items-center gap-0.5 ${trend >= 0 ? 'text-emerald-600' : 'text-red-500'}`}>
            {trend >= 0 ? <ArrowUpRight size={12} /> : <ArrowDownRight size={12} />}
            {Math.abs(trend)}%
          </span>
        )}
      </div>
      <p className="text-xl font-bold text-slate-800">{value}</p>
      <p className="text-xs text-slate-400 mt-0.5">{sub}</p>
    </div>
  )
}

function StatusRow({ icon, label, value, total, color }: { icon: React.ReactNode; label: string; value: number; total: number; color: string }) {
  const pct = total > 0 ? (value / total) * 100 : 0
  return (
    <div className="flex items-center gap-2">
      <div className="w-6 h-6 rounded-lg flex items-center justify-center" style={{ background: `${color}15`, color }}>{icon}</div>
      <span className="text-xs text-slate-600 flex-1">{label}</span>
      <span className="text-sm font-bold text-slate-800 w-6 text-right">{value}</span>
      <div className="w-20 h-1.5 rounded-full bg-slate-100 overflow-hidden">
        <div className="h-full rounded-full transition-all" style={{ width: `${pct}%`, background: color }} />
      </div>
    </div>
  )
}
