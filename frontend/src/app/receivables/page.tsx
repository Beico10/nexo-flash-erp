'use client'
import { useState, useEffect } from 'react'
import { Plus, CheckCircle2, AlertTriangle, Clock, TrendingUp, X, DollarSign, Phone, User } from 'lucide-react'

interface Receivable {
  id: string
  description: string
  customer_name: string
  customer_phone: string
  category: string
  amount: number
  amount_received: number
  due_date: string
  received_at?: string
  payment_method?: string
  installment: number
  total_installments: number
  recurrence: string
  status: string
  is_overdue: boolean
  days_until_due: number
}

interface Summary {
  total_pending: number
  total_overdue: number
  total_received_month: number
  count_pending: number
  count_overdue: number
  overdue_rate: number
  next_due?: Receivable
}

const STATUS_LABEL: Record<string, string> = {
  pending: 'A Receber', received: 'Recebido',
  overdue: 'Vencido', cancelled: 'Cancelado', partial: 'Parcial'
}
const STATUS_COLOR: Record<string, string> = {
  pending: '#1565C0', received: '#2E7D32',
  overdue: '#B71C1C', cancelled: '#757575', partial: '#E65100'
}
const CATEGORY_LABEL: Record<string, string> = {
  servico: '🔧 Serviço', produto: '📦 Produto',
  mensalidade: '🔄 Mensalidade', contrato: '📋 Contrato', outros: '💼 Outros'
}

export default function ReceivablesPage() {
  const [items, setItems] = useState<Receivable[]>([])
  const [summary, setSummary] = useState<Summary | null>(null)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [showReceiveModal, setShowReceiveModal] = useState<string | null>(null)
  const [filterStatus, setFilterStatus] = useState('')
  const [token, setToken] = useState('')

  // Form state
  const [form, setForm] = useState({
    description: '', customer_name: '', customer_phone: '',
    category: 'servico', amount: '', due_date: '',
    installments: '1', recurrence: 'none', notes: ''
  })

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
        fetch(`/api/v1/receivables${qs}`, { headers }),
        fetch('/api/v1/receivables/summary', { headers }),
      ])
      if (listRes.ok) setItems((await listRes.json()).receivables || [])
      if (summaryRes.ok) setSummary(await summaryRes.json())
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    const res = await fetch('/api/v1/receivables', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        ...form,
        amount: parseFloat(form.amount),
        installments: parseInt(form.installments),
      }),
    })
    if (res.ok) {
      setShowForm(false)
      setForm({ description: '', customer_name: '', customer_phone: '', category: 'servico', amount: '', due_date: '', installments: '1', recurrence: 'none', notes: '' })
      fetchData(token)
    }
  }

  const handleReceive = async (id: string, method: string) => {
    const res = await fetch(`/api/v1/receivables/${id}/receive`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ payment_method: method }),
    })
    if (res.ok) { setShowReceiveModal(null); fetchData(token) }
  }

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const fmtDate = (d: string) => new Date(d + 'T12:00:00').toLocaleDateString('pt-BR')

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#1B5E20' }}>Contas a Receber</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>Gerencie seus recebimentos</p>
        </div>
        <button onClick={() => setShowForm(true)} style={{
          display: 'flex', alignItems: 'center', gap: 8,
          background: 'linear-gradient(135deg, #1B5E20, #2E7D32)',
          color: 'white', padding: '10px 20px', borderRadius: 10,
          border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 14,
        }}>
          <Plus size={16} /> Nova Conta
        </button>
      </div>

      {/* Cards */}
      {summary && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14, marginBottom: 24 }}>
          {[
            { label: 'A Receber', value: fmt(summary.total_pending), sub: `${summary.count_pending} contas`, color: '#1565C0', bg: '#E3F2FD', icon: <Clock size={18} /> },
            { label: 'Vencido', value: fmt(summary.total_overdue), sub: `${summary.count_overdue} contas`, color: '#B71C1C', bg: '#FFEBEE', icon: <AlertTriangle size={18} /> },
            { label: 'Recebido/mês', value: fmt(summary.total_received_month), sub: 'mês atual', color: '#2E7D32', bg: '#E8F5E9', icon: <CheckCircle2 size={18} /> },
            { label: 'Inadimplência', value: `${summary.overdue_rate.toFixed(1)}%`, sub: 'do total', color: summary.overdue_rate > 20 ? '#B71C1C' : '#E65100', bg: summary.overdue_rate > 20 ? '#FFEBEE' : '#FFF8E1', icon: <TrendingUp size={18} /> },
          ].map((c, i) => (
            <div key={i} style={{ background: c.bg, borderRadius: 14, padding: 16, border: `1.5px solid ${c.color}22` }}>
              <div className="flex items-center justify-between mb-2">
                <span style={{ fontSize: 11, fontWeight: 700, color: c.color, textTransform: 'uppercase' }}>{c.label}</span>
                <span style={{ color: c.color }}>{c.icon}</span>
              </div>
              <div style={{ fontSize: 20, fontWeight: 700, color: c.color }}>{c.value}</div>
              <div style={{ fontSize: 11, color: '#757575', marginTop: 3 }}>{c.sub}</div>
            </div>
          ))}
        </div>
      )}

      {/* Filtros */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        {['', 'pending', 'overdue', 'received'].map(s => (
          <button key={s} onClick={() => setFilterStatus(s)} style={{
            padding: '6px 14px', borderRadius: 20, border: '1.5px solid',
            borderColor: filterStatus === s ? '#2E7D32' : '#E0E4F0',
            background: filterStatus === s ? '#E8F5E9' : 'white',
            color: filterStatus === s ? '#2E7D32' : '#757575',
            fontSize: 12, fontWeight: 600, cursor: 'pointer',
          }}>
            {s === '' ? 'Todas' : STATUS_LABEL[s]}
          </button>
        ))}
      </div>

      {/* Lista */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#757575' }}>Carregando...</div>
      ) : items.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
          <DollarSign size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
          <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhuma conta encontrada</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {items.map(item => (
            <div key={item.id} style={{
              background: 'white', borderRadius: 12, padding: '14px 16px',
              border: `1.5px solid ${item.is_overdue ? '#EF9A9A' : '#E0E4F0'}`,
              display: 'flex', alignItems: 'center', gap: 14,
              boxShadow: '0 1px 4px rgba(0,0,0,0.05)',
            }}>
              <div style={{ width: 8, height: 8, borderRadius: '50%', flexShrink: 0, background: STATUS_COLOR[item.status] }} />

              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#212121' }}>{item.description}</span>
                  {item.total_installments > 1 && (
                    <span style={{ fontSize: 10, background: '#E8F5E9', color: '#2E7D32', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>
                      {item.installment}/{item.total_installments}
                    </span>
                  )}
                </div>
                <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 12, color: '#757575' }}>
                  {item.customer_name && <span><User size={11} style={{ display: 'inline', marginRight: 3 }} />{item.customer_name}</span>}
                  {item.customer_phone && <span><Phone size={11} style={{ display: 'inline', marginRight: 3 }} />{item.customer_phone}</span>}
                  <span>{CATEGORY_LABEL[item.category] || item.category}</span>
                  <span>Vence: {fmtDate(item.due_date)}</span>
                </div>
              </div>

              <div style={{ textAlign: 'right', flexShrink: 0 }}>
                <div style={{ fontSize: 16, fontWeight: 700, color: item.is_overdue ? '#B71C1C' : '#212121' }}>
                  {fmt(item.amount)}
                </div>
                <div style={{ fontSize: 11, fontWeight: 700, marginTop: 2, color: STATUS_COLOR[item.status] }}>
                  {STATUS_LABEL[item.status]}
                  {item.is_overdue && item.days_until_due < 0 && ` há ${Math.abs(item.days_until_due)} dias`}
                </div>
              </div>

              {item.status !== 'received' && item.status !== 'cancelled' && (
                <button onClick={() => setShowReceiveModal(item.id)} style={{
                  background: '#E8F5E9', color: '#2E7D32',
                  border: '1.5px solid #A5D6A7', borderRadius: 8,
                  padding: '7px 14px', fontSize: 12, fontWeight: 700,
                  cursor: 'pointer', flexShrink: 0,
                }}>
                  ✓ Receber
                </button>
              )}
              {item.status === 'received' && item.received_at && (
                <div style={{ fontSize: 11, color: '#2E7D32', flexShrink: 0, textAlign: 'right' }}>
                  <CheckCircle2 size={14} style={{ display: 'inline', marginRight: 3 }} />
                  {fmtDate(item.received_at)}<br />
                  <span style={{ textTransform: 'capitalize' }}>{item.payment_method}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Modal receber */}
      {showReceiveModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 340 }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>Registrar Recebimento</h3>
              <button onClick={() => setShowReceiveModal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
              {[
                { key: 'pix', label: '⚡ PIX' },
                { key: 'dinheiro', label: '💵 Dinheiro' },
                { key: 'cartao', label: '💳 Cartão' },
                { key: 'boleto', label: '📄 Boleto' },
                { key: 'transferencia', label: '🏦 Transferência' },
              ].map(m => (
                <button key={m.key} onClick={() => handleReceive(showReceiveModal, m.key)} style={{
                  padding: '12px', borderRadius: 10, border: '1.5px solid #E0E4F0',
                  background: 'white', cursor: 'pointer', fontSize: 13, fontWeight: 600,
                }}>
                  {m.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Modal nova conta */}
      {showForm && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 440, maxHeight: '90vh', overflowY: 'auto' }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>Nova Conta a Receber</h3>
              <button onClick={() => setShowForm(false)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {[
                { label: 'Descrição', key: 'description', type: 'text', required: true },
                { label: 'Cliente', key: 'customer_name', type: 'text' },
                { label: 'Telefone (WhatsApp)', key: 'customer_phone', type: 'text' },
                { label: 'Valor (R$)', key: 'amount', type: 'number', required: true },
                { label: 'Vencimento', key: 'due_date', type: 'date', required: true },
                { label: 'Parcelas', key: 'installments', type: 'number' },
              ].map(f => (
                <div key={f.key}>
                  <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>{f.label}</label>
                  <input
                    type={f.type}
                    required={f.required}
                    value={(form as any)[f.key]}
                    onChange={e => setForm(prev => ({ ...prev, [f.key]: e.target.value }))}
                    style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }}
                  />
                </div>
              ))}
              <div>
                <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Categoria</label>
                <select value={form.category} onChange={e => setForm(prev => ({ ...prev, category: e.target.value }))}
                  style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }}>
                  {Object.entries(CATEGORY_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
              </div>
              <button type="submit" style={{
                width: '100%', padding: 13, borderRadius: 10,
                background: 'linear-gradient(135deg, #1B5E20, #2E7D32)',
                color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer', marginTop: 8,
              }}>
                Criar Conta
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
