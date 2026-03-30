'use client'
import { isDemoMode, promptLogin, getBusinessType } from '@/lib/demo'
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

const DEMO_RECEIVABLES_BY_NICHO: Record<string, { items: Receivable[], summary: Summary }> = {
  mechanic: {
    items: [
      { id: '1', description: 'OS-001 Honda Civic - Troca de oleo', customer_name: 'Joao Silva', customer_phone: '5511999990001', category: 'servico', amount: 280, amount_received: 280, due_date: '2026-03-10', received_at: '2026-03-10', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -19 },
      { id: '2', description: 'OS-002 Toyota Corolla - Amortecedor', customer_name: 'Maria Santos', customer_phone: '5511999990002', category: 'servico', amount: 1850, amount_received: 0, due_date: '2026-03-28', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 0 },
      { id: '3', description: 'OS-003 Chevrolet Onix - Revisao', customer_name: 'Carlos Pereira', customer_phone: '5511999990003', category: 'servico', amount: 420, amount_received: 420, due_date: '2026-03-15', received_at: '2026-03-15', payment_method: 'Cartao', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -14 },
      { id: '4', description: 'OS-004 VW Gol - Correia dentada', customer_name: 'Ana Oliveira', customer_phone: '5511999990004', category: 'servico', amount: 980, amount_received: 0, due_date: '2026-03-20', installment: 1, total_installments: 1, recurrence: 'none', status: 'overdue', is_overdue: true, days_until_due: -9 },
      { id: '5', description: 'OS-005 Fiat Pulse - Diagnostico', customer_name: 'Roberto Lima', customer_phone: '5511999990005', category: 'servico', amount: 150, amount_received: 0, due_date: '2026-04-05', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 7 },
      { id: '6', description: 'OS-006 Honda HRV - Freios', customer_name: 'Fernanda Costa', customer_phone: '5511999990006', category: 'servico', amount: 760, amount_received: 760, due_date: '2026-03-22', received_at: '2026-03-22', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -7 },
    ],
    summary: { total_pending: 3080, total_overdue: 980, total_received_month: 4210, count_pending: 3, count_overdue: 1, overdue_rate: 14.3 },
  },
  bakery: {
    items: [
      { id: '1', description: 'Buffet Sabor & Arte - Pedido mensal', customer_name: 'Buffet Sabor & Arte', customer_phone: '5511988880001', category: 'contrato', amount: 2400, amount_received: 2400, due_date: '2026-03-05', received_at: '2026-03-05', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'received', is_overdue: false, days_until_due: -24 },
      { id: '2', description: 'Cafeteria Central - Fornecimento semanal', customer_name: 'Cafeteria Central', customer_phone: '5511988880002', category: 'contrato', amount: 680, amount_received: 0, due_date: '2026-03-28', installment: 1, total_installments: 1, recurrence: 'weekly', status: 'pending', is_overdue: false, days_until_due: 0 },
      { id: '3', description: 'Encomenda casamento - Bolo 5 andares', customer_name: 'Fernanda & Lucas', customer_phone: '5511988880003', category: 'produto', amount: 1800, amount_received: 900, due_date: '2026-03-15', received_at: '2026-03-10', payment_method: 'PIX', installment: 1, total_installments: 2, recurrence: 'none', status: 'partial', is_overdue: false, days_until_due: -14 },
      { id: '4', description: 'Restaurante Bom Sabor - Pao artesanal', customer_name: 'Rest. Bom Sabor', customer_phone: '5511988880004', category: 'contrato', amount: 950, amount_received: 0, due_date: '2026-03-18', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'overdue', is_overdue: true, days_until_due: -11 },
      { id: '5', description: 'Encomenda Pascoa - 30 ovos chocolate', customer_name: 'Patricia Alves', customer_phone: '5511988880005', category: 'produto', amount: 1350, amount_received: 0, due_date: '2026-04-10', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 11 },
      { id: '6', description: 'Mercearia Viva - Pao de queijo (semana)', customer_name: 'Mercearia Viva', customer_phone: '5511988880006', category: 'produto', amount: 420, amount_received: 420, due_date: '2026-03-22', received_at: '2026-03-22', payment_method: 'Dinheiro', installment: 1, total_installments: 1, recurrence: 'weekly', status: 'received', is_overdue: false, days_until_due: -7 },
    ],
    summary: { total_pending: 3380, total_overdue: 950, total_received_month: 5220, count_pending: 3, count_overdue: 1, overdue_rate: 12.8 },
  },
  aesthetics: {
    items: [
      { id: '1', description: 'Progressiva + Hidratacao', customer_name: 'Larissa Lima', customer_phone: '5511977770001', category: 'servico', amount: 380, amount_received: 380, due_date: '2026-03-12', received_at: '2026-03-12', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -17 },
      { id: '2', description: 'Coloracao completa + Luzes', customer_name: 'Mariana Oliveira', customer_phone: '5511977770002', category: 'servico', amount: 580, amount_received: 0, due_date: '2026-03-29', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 1 },
      { id: '3', description: 'Pacote mensal manicure/pedicure', customer_name: 'Patricia Mendes', customer_phone: '5511977770003', category: 'mensalidade', amount: 240, amount_received: 240, due_date: '2026-03-10', received_at: '2026-03-10', payment_method: 'Cartao', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'received', is_overdue: false, days_until_due: -19 },
      { id: '4', description: 'Design de sobrancelha + Henna', customer_name: 'Camila Santos', customer_phone: '5511977770004', category: 'servico', amount: 180, amount_received: 0, due_date: '2026-03-19', installment: 1, total_installments: 1, recurrence: 'none', status: 'overdue', is_overdue: true, days_until_due: -10 },
      { id: '5', description: 'Corte + Escova + Tratamento', customer_name: 'Fernanda Costa', customer_phone: '5511977770005', category: 'servico', amount: 220, amount_received: 0, due_date: '2026-04-03', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 5 },
      { id: '6', description: 'Pacote noiva - maquiagem + cabelo', customer_name: 'Juliana Reis', customer_phone: '5511977770006', category: 'servico', amount: 850, amount_received: 425, due_date: '2026-03-25', received_at: '2026-03-20', payment_method: 'PIX', installment: 1, total_installments: 2, recurrence: 'none', status: 'partial', is_overdue: false, days_until_due: -4 },
    ],
    summary: { total_pending: 1830, total_overdue: 180, total_received_month: 3860, count_pending: 3, count_overdue: 1, overdue_rate: 6.8 },
  },
  logistics: {
    items: [
      { id: '1', description: 'CT-e 000.241 - Frete SP > RJ', customer_name: 'Eletro Distribuidora', customer_phone: '5511966660001', category: 'servico', amount: 3200, amount_received: 3200, due_date: '2026-03-08', received_at: '2026-03-08', payment_method: 'Transferencia', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -21 },
      { id: '2', description: 'CT-e 000.242 - Frete SP > BH', customer_name: 'Atacadao Norte', customer_phone: '5511966660002', category: 'servico', amount: 4800, amount_received: 0, due_date: '2026-03-28', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 0 },
      { id: '3', description: 'Contrato mensal - Logistica e-commerce', customer_name: 'Loja Virtual Plus', customer_phone: '5511966660003', category: 'contrato', amount: 8500, amount_received: 8500, due_date: '2026-03-15', received_at: '2026-03-15', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'received', is_overdue: false, days_until_due: -14 },
      { id: '4', description: 'CT-e 000.238 - Frete Campinas > Goiania', customer_name: 'Industria Sao Paulo', customer_phone: '5511966660004', category: 'servico', amount: 5600, amount_received: 0, due_date: '2026-03-17', installment: 1, total_installments: 1, recurrence: 'none', status: 'overdue', is_overdue: true, days_until_due: -12 },
      { id: '5', description: 'CT-e 000.245 - Frete SP > Curitiba', customer_name: 'Moveis & Cia', customer_phone: '5511966660005', category: 'servico', amount: 2900, amount_received: 0, due_date: '2026-04-05', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 7 },
      { id: '6', description: 'Contrato semanal - Cargas refrigeradas', customer_name: 'Laticinio Fresco', customer_phone: '5511966660006', category: 'contrato', amount: 6200, amount_received: 6200, due_date: '2026-03-22', received_at: '2026-03-22', payment_method: 'Transferencia', installment: 1, total_installments: 1, recurrence: 'weekly', status: 'received', is_overdue: false, days_until_due: -7 },
    ],
    summary: { total_pending: 13300, total_overdue: 5600, total_received_month: 31400, count_pending: 3, count_overdue: 1, overdue_rate: 10.2 },
  },
  industry: {
    items: [
      { id: '1', description: 'PED-2024-1038 - Conexoes PVC lote 50', customer_name: 'Construtora Alfa', customer_phone: '5511955550001', category: 'produto', amount: 12400, amount_received: 12400, due_date: '2026-03-10', received_at: '2026-03-10', payment_method: 'Boleto', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -19 },
      { id: '2', description: 'PED-2024-1042 - Perfis aluminio 200un', customer_name: 'Metalurgica Omega', customer_phone: '5511955550002', category: 'produto', amount: 28700, amount_received: 0, due_date: '2026-03-31', installment: 1, total_installments: 2, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 2 },
      { id: '3', description: 'PED-2024-1039 - Caixas organizadoras', customer_name: 'Supermercado Rede', customer_phone: '5511955550003', category: 'produto', amount: 8900, amount_received: 8900, due_date: '2026-03-15', received_at: '2026-03-14', payment_method: 'Transferencia', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -14 },
      { id: '4', description: 'PED-2024-1040 - Tampas industriais', customer_name: 'Industria Beta', customer_phone: '5511955550004', category: 'produto', amount: 15600, amount_received: 0, due_date: '2026-03-18', installment: 1, total_installments: 1, recurrence: 'none', status: 'overdue', is_overdue: true, days_until_due: -11 },
      { id: '5', description: 'PED-2024-1044 - Tubos PVC DN100', customer_name: 'Hidraulica Plus', customer_phone: '5511955550005', category: 'produto', amount: 9800, amount_received: 0, due_date: '2026-04-08', installment: 1, total_installments: 1, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 10 },
      { id: '6', description: 'Contrato trimestral - Porcas e parafusos', customer_name: 'Ferragens Unidas', customer_phone: '5511955550006', category: 'contrato', amount: 22000, amount_received: 22000, due_date: '2026-03-22', received_at: '2026-03-22', payment_method: 'Boleto', installment: 1, total_installments: 3, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -7 },
    ],
    summary: { total_pending: 54100, total_overdue: 15600, total_received_month: 87400, count_pending: 3, count_overdue: 1, overdue_rate: 9.1 },
  },
  shoes: {
    items: [
      { id: '1', description: 'Venda Loja - Tenis Runner Pro (3 pares)', customer_name: 'Marcos Vieira', customer_phone: '5511944440001', category: 'produto', amount: 897, amount_received: 897, due_date: '2026-03-10', received_at: '2026-03-10', payment_method: 'Cartao', installment: 1, total_installments: 3, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -19 },
      { id: '2', description: 'Representante Sul - Pedido grade aberta', customer_name: 'Calçados Sul Ltda', customer_phone: '5511944440002', category: 'produto', amount: 14800, amount_received: 0, due_date: '2026-03-31', installment: 1, total_installments: 2, recurrence: 'none', status: 'pending', is_overdue: false, days_until_due: 2 },
      { id: '3', description: 'Sandalia Comfort - 15 pares', customer_name: 'Loja Moda Feminina', customer_phone: '5511944440003', category: 'produto', amount: 3600, amount_received: 3600, due_date: '2026-03-14', received_at: '2026-03-14', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -15 },
      { id: '4', description: 'Bota Social - 8 pares', customer_name: 'Boutique Elegance', customer_phone: '5511944440004', category: 'produto', amount: 4720, amount_received: 0, due_date: '2026-03-20', installment: 1, total_installments: 1, recurrence: 'none', status: 'overdue', is_overdue: true, days_until_due: -9 },
      { id: '5', description: 'Comissao representante - Marco', customer_name: 'Anderson Vendas', customer_phone: '5511944440005', category: 'outros', amount: 2100, amount_received: 0, due_date: '2026-04-05', installment: 1, total_installments: 1, recurrence: 'monthly', status: 'pending', is_overdue: false, days_until_due: 7 },
      { id: '6', description: 'Venda e-commerce - 22 pedidos', customer_name: 'Clientes Online', customer_phone: '', category: 'produto', amount: 6420, amount_received: 6420, due_date: '2026-03-22', received_at: '2026-03-22', payment_method: 'PIX', installment: 1, total_installments: 1, recurrence: 'none', status: 'received', is_overdue: false, days_until_due: -7 },
    ],
    summary: { total_pending: 21620, total_overdue: 4720, total_received_month: 32840, count_pending: 3, count_overdue: 1, overdue_rate: 11.4 },
  },
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
    if (isDemoMode()) {
      const nicho = getBusinessType()
      const demoData = DEMO_RECEIVABLES_BY_NICHO[nicho] || DEMO_RECEIVABLES_BY_NICHO.mechanic
      setItems(demoData.items)
      setSummary(demoData.summary)
      setLoading(false)
      return
    }
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
