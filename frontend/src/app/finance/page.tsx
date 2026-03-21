'use client'
import { useState, useEffect } from 'react'
import { TrendingUp, TrendingDown, DollarSign, BarChart3, Calendar, AlertTriangle } from 'lucide-react'

interface DRELine { category: string; label: string; amount: number; percentage: number; is_positive: boolean }
interface DREMonth {
  year: number; month: number; month_name: string
  gross_revenue: number; revenue_lines: DRELine[]
  total_expenses: number; expense_lines: DRELine[]
  net_result: number; net_margin: number; is_profit: boolean
  prev_month_result: number; result_variation: number
}
interface CashFlowDay {
  date: string; day_label: string
  inflows: number; outflows: number
  balance: number; cumulative: number
  is_today: boolean; is_past: boolean
}
interface CashFlowSummary {
  period: string; total_inflows: number; total_outflows: number
  net_cash_flow: number; days: CashFlowDay[]; critical_days: CashFlowDay[]
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
const fmtPct = (v: number) => `${v >= 0 ? '+' : ''}${v.toFixed(1)}%`

export default function FinancePage() {
  const [dre, setDre] = useState<DREMonth | null>(null)
  const [cashflow, setCashflow] = useState<CashFlowSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<'dre' | 'cashflow'>('dre')
  const [cfDays, setCfDays] = useState(30)
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  useEffect(() => { fetchData() }, [cfDays])

  const fetchData = async () => {
    setLoading(true)
    const headers = { Authorization: `Bearer ${token}` }
    const now = new Date()
    try {
      const [dreRes, cfRes] = await Promise.all([
        fetch(`/api/v1/finance/dre?year=${now.getFullYear()}&month=${now.getMonth() + 1}`, { headers }),
        fetch(`/api/v1/finance/cashflow?days=${cfDays}`, { headers }),
      ])
      if (dreRes.ok) setDre(await dreRes.json())
      if (cfRes.ok) setCashflow(await cfRes.json())
    } finally { setLoading(false) }
  }

  const maxCumulative = cashflow ? Math.max(...cashflow.days.map(d => Math.abs(d.cumulative)), 1) : 1

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>Financeiro</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>DRE e Fluxo de Caixa</p>
        </div>
        <div style={{ display: 'flex', background: '#F0F2F8', borderRadius: 10, padding: 4, gap: 4 }}>
          {(['dre', 'cashflow'] as const).map(t => (
            <button key={t} onClick={() => setTab(t)} style={{
              padding: '8px 20px', borderRadius: 8, border: 'none', cursor: 'pointer',
              fontWeight: 600, fontSize: 13,
              background: tab === t ? 'white' : 'transparent',
              color: tab === t ? '#0A3D8F' : '#757575',
              boxShadow: tab === t ? '0 1px 4px rgba(0,0,0,0.1)' : 'none',
            }}>
              {t === 'dre' ? '📊 DRE' : '📈 Fluxo de Caixa'}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#757575' }}>Carregando...</div>
      ) : tab === 'dre' && dre ? (
        <>
          {/* Cards DRE */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 28 }}>
            <div style={{ background: '#E8F5E9', borderRadius: 14, padding: 20, border: '1.5px solid #A5D6A7' }}>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#2E7D32', textTransform: 'uppercase', marginBottom: 8 }}>Receita Bruta</div>
              <div style={{ fontSize: 26, fontWeight: 700, color: '#1B5E20' }}>{fmt(dre.gross_revenue)}</div>
              <div style={{ fontSize: 12, color: '#388E3C', marginTop: 4 }}>{dre.month_name} {dre.year}</div>
            </div>
            <div style={{ background: '#FFEBEE', borderRadius: 14, padding: 20, border: '1.5px solid #EF9A9A' }}>
              <div style={{ fontSize: 11, fontWeight: 700, color: '#B71C1C', textTransform: 'uppercase', marginBottom: 8 }}>Total Despesas</div>
              <div style={{ fontSize: 26, fontWeight: 700, color: '#B71C1C' }}>{fmt(dre.total_expenses)}</div>
              <div style={{ fontSize: 12, color: '#C62828', marginTop: 4 }}>{(dre.total_expenses / dre.gross_revenue * 100).toFixed(1)}% da receita</div>
            </div>
            <div style={{
              background: dre.is_profit ? '#E8F5E9' : '#FFEBEE',
              borderRadius: 14, padding: 20,
              border: `1.5px solid ${dre.is_profit ? '#A5D6A7' : '#EF9A9A'}`,
            }}>
              <div style={{ fontSize: 11, fontWeight: 700, color: dre.is_profit ? '#2E7D32' : '#B71C1C', textTransform: 'uppercase', marginBottom: 8 }}>
                {dre.is_profit ? '✅ Lucro Líquido' : '❌ Prejuízo'}
              </div>
              <div style={{ fontSize: 26, fontWeight: 700, color: dre.is_profit ? '#1B5E20' : '#B71C1C' }}>{fmt(dre.net_result)}</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
                <span style={{ fontSize: 12, color: '#757575' }}>Margem: {dre.net_margin.toFixed(1)}%</span>
                <span style={{
                  fontSize: 11, fontWeight: 700, padding: '1px 8px', borderRadius: 100,
                  background: dre.result_variation >= 0 ? '#E8F5E9' : '#FFEBEE',
                  color: dre.result_variation >= 0 ? '#2E7D32' : '#B71C1C',
                }}>
                  {dre.result_variation >= 0 ? '↑' : '↓'} {fmtPct(dre.result_variation)} vs mês ant.
                </span>
              </div>
            </div>
          </div>

          {/* Detalhamento */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
            {/* Receitas */}
            <div style={{ background: 'white', borderRadius: 14, padding: 20, border: '1.5px solid #E0E4F0' }}>
              <h3 style={{ fontSize: 13, fontWeight: 700, color: '#2E7D32', marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
                <TrendingUp size={16} /> RECEITAS
              </h3>
              {dre.revenue_lines.map((line, i) => (
                <div key={i} style={{ marginBottom: 12 }}>
                  <div className="flex justify-between" style={{ marginBottom: 4 }}>
                    <span style={{ fontSize: 13, color: '#424242' }}>{line.label}</span>
                    <span style={{ fontSize: 13, fontWeight: 700, color: '#2E7D32' }}>{fmt(line.amount)}</span>
                  </div>
                  <div style={{ height: 6, background: '#F0F2F8', borderRadius: 3 }}>
                    <div style={{ height: '100%', background: '#4CAF50', borderRadius: 3, width: `${line.percentage}%` }} />
                  </div>
                  <div style={{ fontSize: 11, color: '#9E9E9E', marginTop: 2 }}>{line.percentage.toFixed(1)}% do total</div>
                </div>
              ))}
            </div>

            {/* Despesas */}
            <div style={{ background: 'white', borderRadius: 14, padding: 20, border: '1.5px solid #E0E4F0' }}>
              <h3 style={{ fontSize: 13, fontWeight: 700, color: '#B71C1C', marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
                <TrendingDown size={16} /> DESPESAS
              </h3>
              {dre.expense_lines.map((line, i) => (
                <div key={i} style={{ marginBottom: 12 }}>
                  <div className="flex justify-between" style={{ marginBottom: 4 }}>
                    <span style={{ fontSize: 13, color: '#424242' }}>{line.label}</span>
                    <span style={{ fontSize: 13, fontWeight: 700, color: '#B71C1C' }}>{fmt(line.amount)}</span>
                  </div>
                  <div style={{ height: 6, background: '#F0F2F8', borderRadius: 3 }}>
                    <div style={{ height: '100%', background: '#EF5350', borderRadius: 3, width: `${Math.min(line.percentage, 100)}%` }} />
                  </div>
                  <div style={{ fontSize: 11, color: '#9E9E9E', marginTop: 2 }}>{line.percentage.toFixed(1)}% da receita</div>
                </div>
              ))}
            </div>
          </div>
        </>
      ) : cashflow ? (
        <>
          {/* Seletor de período */}
          <div style={{ display: 'flex', gap: 8, marginBottom: 20 }}>
            {[30, 60, 90].map(d => (
              <button key={d} onClick={() => setCfDays(d)} style={{
                padding: '7px 16px', borderRadius: 8, border: '1.5px solid',
                borderColor: cfDays === d ? '#1565C0' : '#E0E4F0',
                background: cfDays === d ? '#E3F2FD' : 'white',
                color: cfDays === d ? '#1565C0' : '#757575',
                fontSize: 12, fontWeight: 600, cursor: 'pointer',
              }}>
                {d} dias
              </button>
            ))}
            <span style={{ fontSize: 12, color: '#9E9E9E', padding: '7px 0', marginLeft: 8 }}>{cashflow.period}</span>
          </div>

          {/* Cards fluxo */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 24 }}>
            {[
              { label: 'Entradas Previstas', value: fmt(cashflow.total_inflows), color: '#2E7D32', bg: '#E8F5E9', icon: '↑' },
              { label: 'Saídas Previstas',   value: fmt(cashflow.total_outflows), color: '#B71C1C', bg: '#FFEBEE', icon: '↓' },
              { label: 'Saldo Líquido',      value: fmt(cashflow.net_cash_flow), color: cashflow.net_cash_flow >= 0 ? '#2E7D32' : '#B71C1C', bg: cashflow.net_cash_flow >= 0 ? '#E8F5E9' : '#FFEBEE', icon: cashflow.net_cash_flow >= 0 ? '✅' : '⚠️' },
            ].map((c, i) => (
              <div key={i} style={{ background: c.bg, borderRadius: 14, padding: 16, border: `1.5px solid ${c.color}33` }}>
                <div style={{ fontSize: 11, fontWeight: 700, color: c.color, textTransform: 'uppercase', marginBottom: 6 }}>{c.icon} {c.label}</div>
                <div style={{ fontSize: 22, fontWeight: 700, color: c.color }}>{c.value}</div>
              </div>
            ))}
          </div>

          {/* Alerta dias críticos */}
          {cashflow.critical_days.length > 0 && (
            <div style={{ background: '#FFF8E1', border: '1.5px solid #F9A825', borderRadius: 12, padding: '12px 16px', marginBottom: 20, display: 'flex', gap: 12 }}>
              <AlertTriangle size={18} style={{ color: '#F57F17', flexShrink: 0, marginTop: 2 }} />
              <div>
                <p style={{ fontSize: 13, fontWeight: 700, color: '#E65100' }}>
                  {cashflow.critical_days.length} dia(s) com saldo negativo projetado
                </p>
                <p style={{ fontSize: 12, color: '#757575', marginTop: 2 }}>
                  {cashflow.critical_days.map(d => d.day_label).join(', ')}
                </p>
              </div>
            </div>
          )}

          {/* Gráfico de barras simples */}
          <div style={{ background: 'white', borderRadius: 14, padding: 20, border: '1.5px solid #E0E4F0', overflowX: 'auto' }}>
            <h3 style={{ fontSize: 13, fontWeight: 700, color: '#424242', marginBottom: 16 }}>Saldo Acumulado Projetado</h3>
            <div style={{ display: 'flex', alignItems: 'flex-end', gap: 4, height: 120, minWidth: cashflow.days.length * 20 }}>
              {cashflow.days.map((day, i) => {
                const height = Math.abs(day.cumulative) / maxCumulative * 100
                const isNeg = day.cumulative < 0
                return (
                  <div key={i} title={`${day.day_label}: ${fmt(day.cumulative)}`} style={{
                    flex: 1, minWidth: 12,
                    height: `${Math.max(height, 4)}%`,
                    background: day.is_today ? '#FF9800' : isNeg ? '#EF5350' : '#42A5F5',
                    borderRadius: '3px 3px 0 0',
                    opacity: day.is_past ? 0.5 : 1,
                    cursor: 'pointer',
                    transition: 'opacity 0.2s',
                  }} />
                )
              })}
            </div>
            <div style={{ display: 'flex', gap: 16, marginTop: 12, fontSize: 11 }}>
              {[
                { color: '#42A5F5', label: 'Saldo positivo' },
                { color: '#EF5350', label: 'Saldo negativo' },
                { color: '#FF9800', label: 'Hoje' },
              ].map((l, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <div style={{ width: 10, height: 10, borderRadius: 2, background: l.color }} />
                  <span style={{ color: '#757575' }}>{l.label}</span>
                </div>
              ))}
            </div>
          </div>
        </>
      ) : null}
    </div>
  )
}
