'use client'
import { TrendingUp, TrendingDown, DollarSign, FileText, Clock, CheckCircle, AlertTriangle, Brain } from 'lucide-react'
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import Link from 'next/link'

const revenueData = [
  { day: 'Seg', receita: 4200, impostos: 546 },
  { day: 'Ter', receita: 5800, impostos: 754 },
  { day: 'Qua', receita: 3900, impostos: 507 },
  { day: 'Qui', receita: 6700, impostos: 871 },
  { day: 'Sex', receita: 7200, impostos: 936 },
  { day: 'Sáb', receita: 8100, impostos: 1053 },
  { day: 'Dom', receita: 5400, impostos: 702 },
]

const moduleActivity = [
  { module: 'Mecânica', os: 12 },
  { module: 'Padaria', os: 89 },
  { module: 'Logística', os: 7 },
  { module: 'Estética', os: 23 },
  { module: 'Calçados', os: 15 },
]

const recentOS = [
  { id: 'OS-2026-001842', plate: 'BRA2E19', customer: 'Carlos Silva', status: 'await_approval', value: 850 },
  { id: 'OS-2026-001841', plate: 'ABC1D23', customer: 'Maria Santos', status: 'in_progress', value: 1200 },
  { id: 'OS-2026-001840', plate: 'XYZ9K87', customer: 'João Lima', status: 'done', value: 430 },
  { id: 'OS-2026-001839', plate: 'DEF4G56', customer: 'Ana Costa', status: 'open', value: 0 },
]

const statusBadge: Record<string, { label: string; cls: string }> = {
  open:           { label: 'Aberta',          cls: 'badge-open' },
  await_approval: { label: 'Aguard. aprovação', cls: 'badge-pending' },
  in_progress:    { label: 'Em andamento',     cls: 'badge-approved' },
  done:           { label: 'Concluída',        cls: 'badge-done' },
}

const aiSuggestions = [
  { id: '1', type: 'Mão de obra faltante', os: 'OS-001842', confidence: 0.94, urgency: 'high' },
  { id: '2', type: 'Correção de NCM',      os: 'Produto A', confidence: 0.87, urgency: 'medium' },
  { id: '3', type: 'Produto em NF-e',      os: '12 itens',  confidence: 0.92, urgency: 'low' },
]

export default function DashboardPage() {
  return (
    <div className="space-y-6 animate-fade-in">

      {/* KPI Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <KPICard
          title="Receita Hoje"
          value="R$ 8.100"
          change="+12.4%"
          positive
          icon={<DollarSign size={18} className="text-nexo-500" />}
          sub="vs. ontem R$ 7.200"
        />
        <KPICard
          title="OS Abertas"
          value="14"
          change="+3"
          positive={false}
          icon={<FileText size={18} className="text-amber-500" />}
          sub="4 aguardando aprovação"
        />
        <KPICard
          title="IBS/CBS Devido"
          value="R$ 1.053"
          change="13,0%"
          positive={true}
          icon={<CheckCircle size={18} className="text-emerald-500" />}
          sub="alíquota efetiva"
        />
        <KPICard
          title="Cashback Tributário"
          value="R$ 432"
          change="crédito"
          positive={true}
          icon={<TrendingUp size={18} className="text-nexo-500" />}
          sub="acumulado no mês"
        />
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">

        {/* Revenue Chart */}
        <div className="card p-5 lg:col-span-2">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="font-semibold text-slate-800 text-sm">Receita × Impostos — 7 dias</h2>
              <p className="text-xs text-slate-400 mt-0.5">IBS + CBS calculados automaticamente por NCM</p>
            </div>
            <span className="badge-approved">Esta semana</span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={revenueData}>
              <defs>
                <linearGradient id="gradReceita" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#1A6BFF" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#1A6BFF" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="gradImpostos" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#F59E0B" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#F59E0B" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#EEF0F8" />
              <XAxis dataKey="day" tick={{ fontSize: 11, fill: '#8892C8' }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 11, fill: '#8892C8' }} axisLine={false} tickLine={false} tickFormatter={v => `R$${(v/1000).toFixed(0)}k`} />
              <Tooltip
                contentStyle={{ background: '#fff', border: '1px solid #EEF0F8', borderRadius: 12, fontSize: 12 }}
                formatter={(v: number, n: string) => [`R$ ${v.toLocaleString('pt-BR')}`, n === 'receita' ? 'Receita' : 'IBS+CBS']}
              />
              <Area type="monotone" dataKey="receita"  stroke="#1A6BFF" strokeWidth={2} fill="url(#gradReceita)" />
              <Area type="monotone" dataKey="impostos" stroke="#F59E0B" strokeWidth={2} fill="url(#gradImpostos)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Module Activity */}
        <div className="card p-5">
          <h2 className="font-semibold text-slate-800 text-sm mb-4">Atividade por Módulo</h2>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={moduleActivity} layout="vertical">
              <XAxis type="number" tick={{ fontSize: 11, fill: '#8892C8' }} axisLine={false} tickLine={false} />
              <YAxis type="category" dataKey="module" tick={{ fontSize: 11, fill: '#8892C8' }} axisLine={false} tickLine={false} width={70} />
              <Tooltip
                contentStyle={{ background: '#fff', border: '1px solid #EEF0F8', borderRadius: 12, fontSize: 12 }}
                formatter={(v: number) => [v, 'Transações']}
              />
              <Bar dataKey="os" fill="#1A6BFF" radius={[0, 6, 6, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Bottom row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">

        {/* Recent OS */}
        <div className="card lg:col-span-2">
          <div className="flex items-center justify-between px-5 py-4 border-b border-slate-50">
            <h2 className="font-semibold text-slate-800 text-sm">Últimas Ordens de Serviço</h2>
            <Link href="/mechanic" className="text-xs font-medium text-nexo-500 hover:text-nexo-700">Ver todas →</Link>
          </div>
          <table className="w-full">
            <thead>
              <tr>
                <th className="table-header">Número</th>
                <th className="table-header">Placa</th>
                <th className="table-header">Cliente</th>
                <th className="table-header">Status</th>
                <th className="table-header text-right">Valor</th>
              </tr>
            </thead>
            <tbody>
              {recentOS.map((os) => {
                const s = statusBadge[os.status]
                return (
                  <tr key={os.id} className="hover:bg-slate-50 transition-colors cursor-pointer">
                    <td className="table-cell font-mono text-xs text-nexo-600">{os.id}</td>
                    <td className="table-cell font-mono font-semibold text-slate-800">{os.plate}</td>
                    <td className="table-cell">{os.customer}</td>
                    <td className="table-cell"><span className={s.cls}>{s.label}</span></td>
                    <td className="table-cell text-right font-medium">
                      {os.value > 0 ? `R$ ${os.value.toLocaleString('pt-BR')}` : '—'}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        {/* AI Suggestions Panel */}
        <div className="card p-5">
          <div className="flex items-center gap-2 mb-4">
            <div className="w-7 h-7 bg-nexo-50 rounded-lg flex items-center justify-center">
              <Brain size={14} className="text-nexo-500" />
            </div>
            <h2 className="font-semibold text-slate-800 text-sm">IA — Pendentes</h2>
            <span className="ml-auto text-[10px] font-bold bg-nexo-500 text-white px-1.5 py-0.5 rounded-full">
              {aiSuggestions.length}
            </span>
          </div>
          <div className="space-y-2.5">
            {aiSuggestions.map((s) => (
              <div key={s.id} className="p-3 bg-slate-50 rounded-xl border border-slate-100 hover:border-nexo-200 transition-colors cursor-pointer">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-xs font-semibold text-slate-700">{s.type}</p>
                    <p className="text-xs text-slate-400 mt-0.5">{s.os}</p>
                  </div>
                  <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded-full ${
                    s.urgency === 'high' ? 'bg-red-100 text-red-600' :
                    s.urgency === 'medium' ? 'bg-amber-100 text-amber-600' :
                    'bg-nexo-100 text-nexo-600'
                  }`}>
                    {Math.round(s.confidence * 100)}%
                  </span>
                </div>
                <div className="flex gap-2 mt-2.5">
                  <button className="flex-1 text-xs font-medium py-1 bg-nexo-500 text-white rounded-lg hover:bg-nexo-600 transition-colors">
                    Aprovar
                  </button>
                  <button className="text-xs font-medium py-1 px-3 bg-white text-slate-500 border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors">
                    Ver
                  </button>
                </div>
              </div>
            ))}
          </div>
          <Link href="/ai-approvals" className="mt-3 flex items-center justify-center text-xs font-medium text-nexo-500 hover:text-nexo-700 py-2">
            Ver todas as sugestões →
          </Link>
        </div>
      </div>
    </div>
  )
}

function KPICard({ title, value, change, positive, icon, sub }: {
  title: string; value: string; change: string; positive: boolean; icon: React.ReactNode; sub: string
}) {
  return (
    <div className="stat-card animate-slide-up">
      <div className="flex items-center justify-between">
        <p className="text-xs font-semibold text-slate-400 uppercase tracking-wide">{title}</p>
        <div className="w-8 h-8 bg-slate-50 rounded-xl flex items-center justify-center">{icon}</div>
      </div>
      <div>
        <p className="text-2xl font-display font-700 text-slate-900">{value}</p>
        <div className="flex items-center gap-1.5 mt-1">
          {positive
            ? <TrendingUp size={12} className="text-emerald-500" />
            : <TrendingDown size={12} className="text-red-400" />
          }
          <span className={`text-xs font-medium ${positive ? 'text-emerald-600' : 'text-red-500'}`}>{change}</span>
          <span className="text-xs text-slate-400">{sub}</span>
        </div>
      </div>
    </div>
  )
}
