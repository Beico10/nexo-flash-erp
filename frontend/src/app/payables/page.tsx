'use client'
import { useState, useEffect } from 'react'
import { Plus, CheckCircle2, AlertTriangle, Clock, TrendingDown, Filter, X, DollarSign, Calendar, Building2 } from 'lucide-react'

interface Payable {
  id: string
  description: string
  supplier_name: string
  category: string
  amount: number
  amount_paid: number
  due_date: string
  paid_at?: string
  payment_method?: string
  installment: number
  total_installments: number
  recurrence: string
  status: string
  is_overdue: boolean
  days_until_due: number
  notes?: string
}

interface Summary {
  total_pending: number
  total_overdue: number
  total_paid_month: number
  count_pending: number
  count_overdue: number
  next_due?: Payable
}

const STATUS_LABEL: Record<string, string> = {
  pending: 'Pendente', paid: 'Pago', overdue: 'Vencido', cancelled: 'Cancelado'
}
const STATUS_COLOR: Record<string, string> = {
  pending: '#1565C0', paid: '#2E7D32', overdue: '#B71C1C', cancelled: '#757575'
}
const CATEGORY_LABEL: Record<string, string> = {
  aluguel: '🏠 Aluguel', fornecedor: '📦 Fornecedor', imposto: '🏛️ Imposto',
  folha: '👥 Folha', servico: '⚙️ Serviço', outros: '📋 Outros'
}

export default function PayablesPage() {
  const [payables, setPayables] = useState<Payable[]>([])
  const [summary, setSummary] = useState<Summary | null>(null)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [showPayModal, setShowPayModal] = useState<string | null>(null)
  const [filterStatus, setFilterStatus] = useState('')
  const [token, setToken] = useState('')

  useEffect(() => {
    const t = localStorage.getItem('nexo_token') || ''
    setToken(t)
    fetchData(t)
  }, [filterStatus])

  const fetchData = async (t: string) => {
    setLoading(true)
    try {
      const headers = { Authorization: `Bearer ${t}` }
      const qs = filterStatus ? `?status=${filterStatus}` : ''

      const [listRes, summaryRes] = await Promise.all([
        fetch(`/api/v1/payables${qs}`, { headers }),
        fetch('/api/v1/payables/summary', { headers }),
      ])

      if (listRes.ok) {
        const data = await listRes.json()
        setPayables(data.payables || [])
      }
      if (summaryRes.ok) {
        const data = await summaryRes.json()
        setSummary(data)
      }
    } finally {
      setLoading(false)
    }
  }

  const handlePay = async (id: string, method: string) => {
    const res = await fetch(`/api/v1/payables/${id}/pay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ payment_method: method }),
    })
    if (res.ok) {
      setShowPayModal(null)
      fetchData(token)
    }
  }

  const fmtCurrency = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const fmtDate = (d: string) => new Date(d + 'T12:00:00').toLocaleDateString('pt-BR')

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#0A3D8F' }}>Contas a Pagar</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>Controle seus compromissos financeiros</p>
        </div>
        <button onClick={() => setShowForm(true)} style={{
          display: 'flex', alignItems: 'center', gap: 8,
          background: 'linear-gradient(135deg, #0A3D8F, #1565C0)',
          color: 'white', padding: '10px 20px', borderRadius: 10,
          border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 14,
        }}>
          <Plus size={16} /> Nova Conta
        </button>
      </div>

      {/* Cards de resumo */}
      {summary && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 24 }}>
          {[
            { label: 'A Pagar', value: fmtCurrency(summary.total_pending), count: `${summary.count_pending} contas`, color: '#1565C0', bg: '#E3F2FD', icon: <Clock size={20} /> },
            { label: 'Vencidas', value: fmtCurrency(summary.total_overdue), count: `${summary.count_overdue} contas`, color: '#B71C1C', bg: '#FFEBEE', icon: <AlertTriangle size={20} /> },
            { label: 'Pago este mês', value: fmtCurrency(summary.total_paid_month), count: 'mês atual', color: '#2E7D32', bg: '#E8F5E9', icon: <CheckCircle2 size={20} /> },
          ].map((card, i) => (
            <div key={i} style={{ background: card.bg, borderRadius: 14, padding: 20, border: `1.5px solid ${card.color}22` }}>
              <div className="flex items-center justify-between mb-2">
                <span style={{ fontSize: 12, fontWeight: 700, color: card.color, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{card.label}</span>
                <span style={{ color: card.color }}>{card.icon}</span>
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: card.color }}>{card.value}</div>
              <div style={{ fontSize: 11, color: '#757575', marginTop: 4 }}>{card.count}</div>
            </div>
          ))}
        </div>
      )}

      {/* Próximo vencimento */}
      {summary?.next_due && (
        <div style={{ background: '#FFF8E1', border: '1.5px solid #F9A825', borderRadius: 12, padding: '12px 16px', marginBottom: 20, display: 'flex', alignItems: 'center', gap: 12 }}>
          <Calendar size={18} style={{ color: '#F57F17', flexShrink: 0 }} />
          <div>
            <span style={{ fontSize: 12, fontWeight: 700, color: '#F57F17' }}>PRÓXIMO VENCIMENTO</span>
            <p style={{ fontSize: 13, color: '#333', marginTop: 2 }}>
              <strong>{summary.next_due.description}</strong> — {fmtCurrency(summary.next_due.amount)} em <strong>{fmtDate(summary.next_due.due_date)}</strong>
              {summary.next_due.days_until_due <= 3 && <span style={{ marginLeft: 8, background: '#F57F17', color: 'white', padding: '1px 8px', borderRadius: 100, fontSize: 11 }}>⚠️ {summary.next_due.days_until_due === 0 ? 'Hoje!' : `${summary.next_due.days_until_due} dias`}</span>}
            </p>
          </div>
        </div>
      )}

      {/* Filtros */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        {['', 'pending', 'overdue', 'paid'].map(s => (
          <button key={s} onClick={() => setFilterStatus(s)} style={{
            padding: '6px 14px', borderRadius: 20, border: '1.5px solid',
            borderColor: filterStatus === s ? '#1565C0' : '#E0E4F0',
            background: filterStatus === s ? '#E3F2FD' : 'white',
            color: filterStatus === s ? '#1565C0' : '#757575',
            fontSize: 12, fontWeight: 600, cursor: 'pointer',
          }}>
            {s === '' ? 'Todas' : STATUS_LABEL[s]}
          </button>
        ))}
      </div>

      {/* Lista de contas */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#757575' }}>Carregando...</div>
      ) : payables.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
          <DollarSign size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
          <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhuma conta encontrada</p>
          <p style={{ fontSize: 13, marginTop: 4 }}>Clique em "Nova Conta" para adicionar</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {payables.map(p => (
            <div key={p.id} style={{
              background: 'white', borderRadius: 12, padding: '14px 16px',
              border: `1.5px solid ${p.is_overdue ? '#EF9A9A' : '#E0E4F0'}`,
              display: 'flex', alignItems: 'center', gap: 14,
              boxShadow: p.is_overdue ? '0 2px 8px rgba(183,28,28,0.08)' : '0 1px 4px rgba(0,0,0,0.05)',
            }}>
              {/* Status indicator */}
              <div style={{
                width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
                background: STATUS_COLOR[p.status],
              }} />

              {/* Info principal */}
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#212121' }}>{p.description}</span>
                  {p.total_installments > 1 && (
                    <span style={{ fontSize: 10, background: '#E3F2FD', color: '#1565C0', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>
                      {p.installment}/{p.total_installments}
                    </span>
                  )}
                  {p.recurrence !== 'none' && (
                    <span style={{ fontSize: 10, background: '#F3E5F5', color: '#7B1FA2', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>
                      🔄 Recorrente
                    </span>
                  )}
                </div>
                <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 12, color: '#757575' }}>
                  {p.supplier_name && <span><Building2 size={11} style={{ display: 'inline', marginRight: 3 }} />{p.supplier_name}</span>}
                  <span>{CATEGORY_LABEL[p.category] || p.category}</span>
                  <span><Calendar size={11} style={{ display: 'inline', marginRight: 3 }} />{fmtDate(p.due_date)}</span>
                </div>
              </div>

              {/* Valor */}
              <div style={{ textAlign: 'right', flexShrink: 0 }}>
                <div style={{ fontSize: 16, fontWeight: 700, color: p.is_overdue ? '#B71C1C' : '#212121' }}>
                  {fmtCurrency(p.amount)}
                </div>
                <div style={{
                  fontSize: 11, fontWeight: 700, marginTop: 2,
                  color: STATUS_COLOR[p.status],
                }}>
                  {STATUS_LABEL[p.status]}
                  {p.is_overdue && p.days_until_due < 0 && ` há ${Math.abs(p.days_until_due)} dias`}
                  {p.days_until_due >= 0 && p.days_until_due <= 3 && p.status === 'pending' && ` — vence em ${p.days_until_due === 0 ? 'hoje' : `${p.days_until_due}d`}`}
                </div>
              </div>

              {/* Ação */}
              {p.status !== 'paid' && p.status !== 'cancelled' && (
                <button onClick={() => setShowPayModal(p.id)} style={{
                  background: '#E8F5E9', color: '#2E7D32',
                  border: '1.5px solid #A5D6A7', borderRadius: 8,
                  padding: '7px 14px', fontSize: 12, fontWeight: 700,
                  cursor: 'pointer', flexShrink: 0,
                }}>
                  ✓ Pagar
                </button>
              )}
              {p.status === 'paid' && p.paid_at && (
                <div style={{ fontSize: 11, color: '#2E7D32', flexShrink: 0, textAlign: 'right' }}>
                  <CheckCircle2 size={14} style={{ display: 'inline', marginRight: 3 }} />
                  {fmtDate(p.paid_at)}<br />
                  <span style={{ textTransform: 'capitalize' }}>{p.payment_method}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Modal de pagamento */}
      {showPayModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 340, boxShadow: '0 20px 60px rgba(0,0,0,0.2)' }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>Registrar Pagamento</h3>
              <button onClick={() => setShowPayModal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <p style={{ fontSize: 13, color: '#757575', marginBottom: 16 }}>Selecione a forma de pagamento:</p>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
              {[
                { key: 'pix', label: '⚡ PIX' },
                { key: 'boleto', label: '📄 Boleto' },
                { key: 'cartao', label: '💳 Cartão' },
                { key: 'dinheiro', label: '💵 Dinheiro' },
                { key: 'transferencia', label: '🏦 Transferência' },
              ].map(m => (
                <button key={m.key} onClick={() => handlePay(showPayModal, m.key)} style={{
                  padding: '12px', borderRadius: 10, border: '1.5px solid #E0E4F0',
                  background: 'white', cursor: 'pointer', fontSize: 13, fontWeight: 600,
                  transition: 'all 0.15s',
                }}>
                  {m.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

    </div>
  )
}
