'use client'
import { useState, useEffect } from 'react'
import { isDemoMode, promptLogin, getBusinessType } from '@/lib/demo'
import { Plus, CheckCircle2, AlertTriangle, Clock, X, DollarSign, Calendar, Building2 } from 'lucide-react'

interface Payable { id: string; description: string; supplier_name: string; category: string; amount: number; amount_paid: number; due_date: string; paid_at?: string; payment_method?: string; installment: number; total_installments: number; recurrence: string; status: string; is_overdue: boolean; days_until_due: number }
interface Summary { total_pending: number; total_overdue: number; total_paid_month: number; count_pending: number; count_overdue: number }

const STATUS_LABEL: Record<string, string> = { pending: 'Pendente', paid: 'Pago', overdue: 'Vencido', cancelled: 'Cancelado' }
const STATUS_COLOR: Record<string, string> = { pending: '#1565C0', paid: '#2E7D32', overdue: '#B71C1C', cancelled: '#757575' }
const CATEGORY_LABEL: Record<string, string> = { aluguel: 'Aluguel', fornecedor: 'Fornecedor', imposto: 'Imposto', folha: 'Folha', servico: 'Servico', outros: 'Outros' }

const COMMON_PAYABLES = (nicho: string): Payable[] => {
  const fornecedorItems: Record<string, Payable[]> = {
    mechanic: [
      { id: '2', description: 'Compra Pecas - AutoPecas Silva', supplier_name: 'AutoPecas Silva Ltda', category: 'fornecedor', amount: 8750, amount_paid: 8750, due_date: '2026-03-15', installment: 2, total_installments: 3, recurrence: 'none', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Transferencia' },
      { id: '7', description: 'Pecas Motor - Distribuidora Norte', supplier_name: 'Distribuidora Norte', category: 'fornecedor', amount: 2100, amount_paid: 0, due_date: '2026-04-10', installment: 3, total_installments: 5, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 12 },
    ],
    bakery: [
      { id: '2', description: 'Farinha Trigo 50kg - Moinho Estrela', supplier_name: 'Moinho Estrela', category: 'fornecedor', amount: 3200, amount_paid: 3200, due_date: '2026-03-15', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Boleto' },
      { id: '7', description: 'Manteiga + Ovos - Distribuidora Fresca', supplier_name: 'Distribuidora Fresca', category: 'fornecedor', amount: 1840, amount_paid: 0, due_date: '2026-04-05', installment: 1, total_installments: 1, recurrence: 'weekly', status: 'pending', is_overdue: false, days_until_due: 7 },
    ],
    aesthetics: [
      { id: '2', description: 'Produtos capilares - L\'Oreal Pro', supplier_name: 'L\'Oreal Profissionnel', category: 'fornecedor', amount: 2800, amount_paid: 2800, due_date: '2026-03-15', installment: 1, total_installments: 2, recurrence: 'none', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Boleto' },
      { id: '7', description: 'Esmaltes + Produtos - Cimed Beauty', supplier_name: 'Cimed Beauty', category: 'fornecedor', amount: 980, amount_paid: 0, due_date: '2026-04-10', installment: 2, total_installments: 3, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 12 },
    ],
    logistics: [
      { id: '2', description: 'Diesel - Posto Petrolandia', supplier_name: 'Posto Petrolandia', category: 'fornecedor', amount: 12400, amount_paid: 12400, due_date: '2026-03-15', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Cartao Frota' },
      { id: '7', description: 'Pneus - Bridgestone (4 un)', supplier_name: 'Bridgestone Pneus', category: 'fornecedor', amount: 5600, amount_paid: 0, due_date: '2026-04-10', installment: 2, total_installments: 4, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 12 },
    ],
    industry: [
      { id: '2', description: 'Resina EP-40 200kg - Quimica Brasil', supplier_name: 'Quimica Brasil', category: 'fornecedor', amount: 18600, amount_paid: 18600, due_date: '2026-03-15', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Boleto' },
      { id: '7', description: 'PVC Rigido 500kg - Plasticos Sul', supplier_name: 'Plasticos Sul', category: 'fornecedor', amount: 9800, amount_paid: 0, due_date: '2026-04-10', installment: 2, total_installments: 3, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 12 },
    ],
    shoes: [
      { id: '2', description: 'Couro bovino 50m² - Curtume Gaucho', supplier_name: 'Curtume Gaucho', category: 'fornecedor', amount: 7200, amount_paid: 7200, due_date: '2026-03-15', installment: 2, total_installments: 4, recurrence: 'none', status: 'paid', is_overdue: false, days_until_due: -14, paid_at: '2026-03-14', payment_method: 'Boleto' },
      { id: '7', description: 'Solados - Vulcabrás (lote)', supplier_name: 'Vulcabras Calçados', category: 'fornecedor', amount: 4300, amount_paid: 0, due_date: '2026-04-10', installment: 3, total_installments: 5, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 12 },
    ],
  }
  const aluguelValores: Record<string, number> = { mechanic: 4800, bakery: 3200, aesthetics: 2400, logistics: 8500, industry: 12000, shoes: 3600 }
  const folhaValores: Record<string, number> = { mechanic: 12400, bakery: 8600, aesthetics: 9200, logistics: 18500, industry: 28000, shoes: 11200 }
  const impostoValores: Record<string, number> = { mechanic: 3200, bakery: 2100, aesthetics: 1800, logistics: 5400, industry: 9800, shoes: 2900 }
  const aluguel = aluguelValores[nicho] || 4800
  const folha = folhaValores[nicho] || 12400
  const imposto = impostoValores[nicho] || 3200
  return [
    { id: '1', description: 'Aluguel Marco 2026', supplier_name: 'Imobiliaria Central', category: 'aluguel', amount: aluguel, amount_paid: aluguel, due_date: '2026-03-05', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -24, paid_at: '2026-03-05', payment_method: 'PIX' },
    ...(fornecedorItems[nicho] || fornecedorItems.mechanic),
    { id: '3', description: 'Folha de Pagamento Marco', supplier_name: 'Funcionarios', category: 'folha', amount: folha, amount_paid: 0, due_date: '2026-03-31', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'pending', is_overdue: false, days_until_due: 2 },
    { id: '4', description: 'Energia Eletrica Marco', supplier_name: 'CPFL Energia', category: 'servico', amount: 1840, amount_paid: 1840, due_date: '2026-03-20', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -9, paid_at: '2026-03-19', payment_method: 'Boleto' },
    { id: '5', description: 'SIMPLES Nacional Marco', supplier_name: 'Receita Federal', category: 'imposto', amount: imposto, amount_paid: 0, due_date: '2026-03-20', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'overdue', is_overdue: true, days_until_due: -9 },
    { id: '6', description: 'Sistema ERP Nexo', supplier_name: 'Gestao Para Todos', category: 'servico', amount: 297, amount_paid: 0, due_date: '2026-04-01', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'pending', is_overdue: false, days_until_due: 3 },
    { id: '8', description: 'Internet e Telefone', supplier_name: 'Vivo Empresas', category: 'servico', amount: 380, amount_paid: 0, due_date: '2026-03-30', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'pending', is_overdue: false, days_until_due: 0 },
    { id: '9', description: 'Seguro Equipamentos', supplier_name: 'Porto Seguro', category: 'servico', amount: 620, amount_paid: 620, due_date: '2026-03-10', installment: 1, total_installments: 12, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -19, paid_at: '2026-03-10', payment_method: 'Debito Automatico' },
    { id: '10', description: 'Contador Honorarios Marco', supplier_name: 'Escritorio Contabil Lima', category: 'servico', amount: 850, amount_paid: 850, due_date: '2026-03-01', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'paid', is_overdue: false, days_until_due: -28, paid_at: '2026-03-01', payment_method: 'PIX' },
  ]
}

export default function PayablesPage() {
  const [payables, setPayables] = useState<Payable[]>([])
  const [summary, setSummary] = useState<Summary | null>(null)
  const [loading, setLoading] = useState(true)
  const [showPayModal, setShowPayModal] = useState<string | null>(null)
  const [filterStatus, setFilterStatus] = useState('')

  useEffect(() => {
    if (isDemoMode()) {
      const nicho = getBusinessType()
      const items = COMMON_PAYABLES(nicho)
      const pending = items.filter(p => p.status === 'pending' || p.status === 'overdue')
      const overdue = items.filter(p => p.is_overdue)
      const paid = items.filter(p => p.status === 'paid')
      setPayables(items)
      setSummary({
        total_pending: pending.reduce((s, p) => s + p.amount, 0),
        total_overdue: overdue.reduce((s, p) => s + p.amount, 0),
        total_paid_month: paid.reduce((s, p) => s + p.amount, 0),
        count_pending: pending.length,
        count_overdue: overdue.length,
      })
      setLoading(false)
      return
    }
    const t = localStorage.getItem('nexo_token') || ''
    const headers = { Authorization: `Bearer ${t}` }
    const qs = filterStatus ? `?status=${filterStatus}` : ''
    Promise.all([
      fetch(`/api/v1/payables${qs}`, { headers }).then(r => r.json()),
      fetch('/api/v1/payables/summary', { headers }).then(r => r.json()),
    ]).then(([list, sum]) => { setPayables(list.payables || []); setSummary(sum) }).finally(() => setLoading(false))
  }, [filterStatus])

  const fmtCurrency = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const fmtDate = (d: string) => new Date(d + 'T12:00:00').toLocaleDateString('pt-BR')

  const filtered = filterStatus ? payables.filter(p => p.status === filterStatus) : payables

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#0A3D8F' }}>Contas a Pagar</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>Controle seus compromissos financeiros</p>
        </div>
        <button onClick={() => promptLogin()} style={{ display: 'flex', alignItems: 'center', gap: 8, background: 'linear-gradient(135deg, #0A3D8F, #1565C0)', color: 'white', padding: '10px 20px', borderRadius: 10, border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 14 }}>
          <Plus size={16} /> Nova Conta
        </button>
      </div>

      {summary && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 24 }}>
          {[
            { label: 'A Pagar', value: fmtCurrency(summary.total_pending), count: `${summary.count_pending} contas`, color: '#1565C0', bg: '#E3F2FD', icon: <Clock size={20} /> },
            { label: 'Vencidas', value: fmtCurrency(summary.total_overdue), count: `${summary.count_overdue} contas`, color: '#B71C1C', bg: '#FFEBEE', icon: <AlertTriangle size={20} /> },
            { label: 'Pago este mes', value: fmtCurrency(summary.total_paid_month), count: 'mes atual', color: '#2E7D32', bg: '#E8F5E9', icon: <CheckCircle2 size={20} /> },
          ].map((card, i) => (
            <div key={i} style={{ background: card.bg, borderRadius: 14, padding: 20, border: `1.5px solid ${card.color}22` }}>
              <div className="flex items-center justify-between mb-2">
                <span style={{ fontSize: 12, fontWeight: 700, color: card.color, textTransform: 'uppercase' }}>{card.label}</span>
                <span style={{ color: card.color }}>{card.icon}</span>
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: card.color }}>{card.value}</div>
              <div style={{ fontSize: 11, color: '#757575', marginTop: 4 }}>{card.count}</div>
            </div>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        {['', 'pending', 'overdue', 'paid'].map(s => (
          <button key={s} onClick={() => setFilterStatus(s)} style={{ padding: '6px 14px', borderRadius: 20, border: '1.5px solid', borderColor: filterStatus === s ? '#1565C0' : '#E0E4F0', background: filterStatus === s ? '#E3F2FD' : 'white', color: filterStatus === s ? '#1565C0' : '#757575', fontSize: 12, fontWeight: 600, cursor: 'pointer' }}>
            {s === '' ? 'Todas' : STATUS_LABEL[s]}
          </button>
        ))}
      </div>

      {loading ? <div style={{ textAlign: 'center', padding: 40 }}>Carregando...</div> : filtered.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
          <DollarSign size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
          <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhuma conta encontrada</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {filtered.map(p => (
            <div key={p.id} style={{ background: 'white', borderRadius: 12, padding: '14px 16px', border: `1.5px solid ${p.is_overdue ? '#EF9A9A' : '#E0E4F0'}`, display: 'flex', alignItems: 'center', gap: 14 }}>
              <div style={{ width: 8, height: 8, borderRadius: '50%', flexShrink: 0, background: STATUS_COLOR[p.status] }} />
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#212121' }}>{p.description}</span>
                  {p.total_installments > 1 && <span style={{ fontSize: 10, background: '#E3F2FD', color: '#1565C0', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>{p.installment}/{p.total_installments}</span>}
                  {p.recurrence !== 'none' && <span style={{ fontSize: 10, background: '#F3E5F5', color: '#7B1FA2', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>Recorrente</span>}
                </div>
                <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 12, color: '#757575' }}>
                  {p.supplier_name && <span><Building2 size={11} style={{ display: 'inline', marginRight: 3 }} />{p.supplier_name}</span>}
                  <span>{CATEGORY_LABEL[p.category] || p.category}</span>
                  <span><Calendar size={11} style={{ display: 'inline', marginRight: 3 }} />{fmtDate(p.due_date)}</span>
                </div>
              </div>
              <div style={{ textAlign: 'right', flexShrink: 0 }}>
                <div style={{ fontSize: 16, fontWeight: 700, color: p.is_overdue ? '#B71C1C' : '#212121' }}>{fmtCurrency(p.amount)}</div>
                <div style={{ fontSize: 11, fontWeight: 700, marginTop: 2, color: STATUS_COLOR[p.status] }}>
                  {STATUS_LABEL[p.status]}
                  {p.is_overdue && p.days_until_due < 0 && ` ha ${Math.abs(p.days_until_due)} dias`}
                  {p.days_until_due >= 0 && p.days_until_due <= 3 && p.status === 'pending' && ` - vence em ${p.days_until_due === 0 ? 'hoje' : `${p.days_until_due}d`}`}
                </div>
              </div>
              {p.status !== 'paid' && p.status !== 'cancelled' && (
                <button onClick={() => isDemoMode() ? promptLogin() : setShowPayModal(p.id)} style={{ background: '#E8F5E9', color: '#2E7D32', border: '1.5px solid #A5D6A7', borderRadius: 8, padding: '7px 14px', fontSize: 12, fontWeight: 700, cursor: 'pointer', flexShrink: 0 }}>
                  Pagar
                </button>
              )}
              {p.status === 'paid' && p.paid_at && (
                <div style={{ fontSize: 11, color: '#2E7D32', flexShrink: 0, textAlign: 'right' }}>
                  <CheckCircle2 size={14} style={{ display: 'inline', marginRight: 3 }} />
                  {fmtDate(p.paid_at)}<br />
                  <span>{p.payment_method}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showPayModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 340 }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>Registrar Pagamento</h3>
              <button onClick={() => setShowPayModal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <p style={{ fontSize: 13, color: '#757575', marginBottom: 16 }}>Selecione a forma de pagamento:</p>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
              {['PIX', 'Boleto', 'Cartao', 'Dinheiro', 'Transferencia'].map(m => (
                <button key={m} onClick={() => promptLogin()} style={{ padding: '12px', borderRadius: 10, border: '1.5px solid #E0E4F0', background: 'white', cursor: 'pointer', fontSize: 13, fontWeight: 600 }}>
                  {m}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
