'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'

const DEMO_EXPENSES = [
  { id: '1', desc: 'Compra de peças — Distribuidora AutoParts', value: 1250, date: '21/03', category: 'Estoque', status: 'pago' },
  { id: '2', desc: 'Aluguel do galpão — Março 2026', value: 2800, date: '20/03', category: 'Fixo', status: 'pago' },
  { id: '3', desc: 'Energia elétrica — fatura fev/26', value: 480, date: '18/03', category: 'Utilidades', status: 'pago' },
  { id: '4', desc: 'Óleo e lubrificantes — estoque', value: 320, date: '15/03', category: 'Estoque', status: 'pago' },
  { id: '5', desc: 'Contador — honorários março', value: 650, date: '01/03', category: 'Serviços', status: 'pago' },
]

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export default function ExpensesPage() {
  const router = useRouter()
  const [showModal, setShowModal] = useState(false)
  const total = DEMO_EXPENSES.reduce((s, e) => s + e.value, 0)

  return (
    <div style={{ padding: 24, maxWidth: 900, margin: '0 auto', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
        <div>
          <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>Despesas</p>
          <p style={{ fontSize: 13, color: '#94a3b8', margin: 0 }}>Gerencie seus gastos e custos operacionais</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          style={{ background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 10, padding: '10px 18px', fontSize: 13, fontWeight: 700, cursor: 'pointer' }}
        >
          + Nova Despesa
        </button>
      </div>

      {/* Cards resumo */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 24 }}>
        <div style={{ background: 'white', borderRadius: 14, padding: '18px 20px', border: '0.5px solid #e8e8e8' }}>
          <p style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 8px' }}>Total em março</p>
          <p style={{ fontSize: 26, fontWeight: 800, color: '#1C1917', margin: 0, letterSpacing: -0.5 }}>{fmt(total)}</p>
        </div>
        <div style={{ background: 'white', borderRadius: 14, padding: '18px 20px', border: '0.5px solid #e8e8e8' }}>
          <p style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 8px' }}>Maior gasto</p>
          <p style={{ fontSize: 26, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>R$ 2.800</p>
          <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>Aluguel</p>
        </div>
        <div style={{ background: '#F0FDF4', borderRadius: 14, padding: '18px 20px', border: '1px solid #DCFCE7' }}>
          <p style={{ fontSize: 11, fontWeight: 600, color: '#16A34A', textTransform: 'uppercase', letterSpacing: '0.06em', margin: '0 0 8px' }}>Crédito fiscal IBS/CBS</p>
          <p style={{ fontSize: 26, fontWeight: 800, color: '#16A34A', margin: '0 0 4px', letterSpacing: -0.5 }}>+R$ 312</p>
          <p style={{ fontSize: 11, color: '#16A34A', margin: 0 }}>abatido das compras</p>
        </div>
      </div>

      {/* Lista de despesas */}
      <div style={{ background: 'white', borderRadius: 16, border: '0.5px solid #e8e8e8', overflow: 'hidden' }}>
        <div style={{ padding: '16px 20px', borderBottom: '0.5px solid #f0f0f0' }}>
          <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>Lançamentos</p>
        </div>
        {DEMO_EXPENSES.map((exp, i) => (
          <div key={exp.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 20px', borderBottom: i < DEMO_EXPENSES.length - 1 ? '0.5px solid #f8f8f8' : 'none' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{ width: 36, height: 36, background: '#F0F4FF', borderRadius: 9, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 14, flexShrink: 0 }}>
                {exp.category === 'Estoque' ? '📦' : exp.category === 'Fixo' ? '🏠' : exp.category === 'Utilidades' ? '⚡' : '📋'}
              </div>
              <div>
                <p style={{ fontSize: 13, fontWeight: 600, color: '#1C1917', margin: '0 0 2px' }}>{exp.desc}</p>
                <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>{exp.date} · {exp.category}</p>
              </div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <p style={{ fontSize: 14, fontWeight: 700, color: '#DC2626', margin: '0 0 2px' }}>-{fmt(exp.value)}</p>
              <span style={{ fontSize: 10, background: '#F0FDF4', color: '#16A34A', padding: '2px 8px', borderRadius: 100, fontWeight: 600 }}>✓ pago</span>
            </div>
          </div>
        ))}
      </div>

      {/* Modal nova despesa — pede login */}
      {showModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)', zIndex: 999, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}
          onClick={() => setShowModal(false)}
        >
          <div onClick={e => e.stopPropagation()} style={{ background: 'white', borderRadius: 20, padding: 28, width: '100%', maxWidth: 380, textAlign: 'center' }}>
            <p style={{ fontSize: 18, fontWeight: 800, color: '#1C1917', margin: '0 0 8px' }}>Gostou do sistema?</p>
            <p style={{ fontSize: 13, color: '#64748b', lineHeight: 1.6, margin: '0 0 20px' }}>
              Para lançar despesas reais e salvar seus dados,<br />crie sua conta gratuita.
            </p>
            <button onClick={() => router.push('/cadastro')} style={{ width: '100%', padding: 13, borderRadius: 10, border: 'none', background: '#0A3D8F', color: 'white', fontSize: 14, fontWeight: 700, cursor: 'pointer', marginBottom: 8 }}>
              Criar conta grátis — 7 dias sem cartão
            </button>
            <button onClick={() => setShowModal(false)} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer' }}>
              Continuar explorando
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
