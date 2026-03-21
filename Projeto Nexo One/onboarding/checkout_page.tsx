'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

const PLANS = [
  {
    id: 'starter', name: 'Starter', price: 147, priceYearly: 117,
    desc: 'Para começar com o essencial',
    features: ['Módulo do nicho', 'Motor IBS/CBS 2026', 'WhatsApp de aprovação', 'Suporte por email'],
  },
  {
    id: 'pro', name: 'Pro', price: 297, priceYearly: 237,
    desc: 'O mais escolhido', popular: true,
    features: ['Tudo do Starter', 'Estoque inteligente', 'DRE + Fluxo de Caixa', 'Co-Piloto IA', 'Roteirizador'],
  },
  {
    id: 'business', name: 'Business', price: 497, priceYearly: 397,
    desc: 'Para operações maiores',
    features: ['Tudo do Pro', 'Despacho em Lote XML/EDI', 'API Access', 'Suporte prioritário'],
  },
]

export default function CheckoutPage() {
  const router = useRouter()
  const [cycle, setCycle] = useState<'monthly' | 'yearly'>('monthly')
  const [selectedPlan, setSelectedPlan] = useState('pro')
  const [step, setStep] = useState<'plan' | 'company' | 'payment'>('plan')
  const [loading, setLoading] = useState(false)
  const [nichoPretty, setNichoPretty] = useState('seu negócio')
  const [form, setForm] = useState({
    company_name: '', cnpj: '', phone: '',
    address: '', city: '', state: '',
  })

  const NICHO_LABELS: Record<string, string> = {
    mechanic: 'Mecânica', bakery: 'Padaria', industry: 'Indústria',
    logistics: 'Logística', aesthetics: 'Estética', shoes: 'Calçados',
  }

  useEffect(() => {
    const bt = localStorage.getItem('nexo_business_type') || 'mechanic'
    setNichoPretty(NICHO_LABELS[bt] || 'seu negócio')
  }, [])

  const plan = PLANS.find(p => p.id === selectedPlan)!
  const price = cycle === 'yearly' ? plan.priceYearly : plan.price
  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  const formatCNPJ = (value: string) => {
    const n = value.replace(/\D/g, '').slice(0, 14)
    if (n.length <= 2) return n
    if (n.length <= 5) return `${n.slice(0,2)}.${n.slice(2)}`
    if (n.length <= 8) return `${n.slice(0,2)}.${n.slice(2,5)}.${n.slice(5)}`
    if (n.length <= 12) return `${n.slice(0,2)}.${n.slice(2,5)}.${n.slice(5,8)}/${n.slice(8)}`
    return `${n.slice(0,2)}.${n.slice(2,5)}.${n.slice(5,8)}/${n.slice(8,12)}-${n.slice(12)}`
  }

  return (
    <div style={{
      minHeight: '100vh', background: '#F8F7F4',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    }}>

      {/* Topbar */}
      <div style={{ background: 'white', borderBottom: '0.5px solid #e8e8e8', padding: '14px 24px', display: 'flex', alignItems: 'center', gap: 12 }}>
        <div style={{ width: 28, height: 28, background: '#0A3D8F', borderRadius: 7, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <svg width="14" height="14" viewBox="0 0 64 64">
            <circle cx="32" cy="28" r="5" fill="white"/>
            <circle cx="32" cy="44" r="4" fill="white"/>
            <line x1="20" y1="20" x2="32" y2="28" stroke="white" strokeWidth="3"/>
            <line x1="32" y1="28" x2="44" y2="20" stroke="white" strokeWidth="3"/>
          </svg>
        </div>
        <span style={{ fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Gestão Para Todos</span>
        <div style={{ flex: 1 }} />
        <div style={{ display: 'flex', gap: 8 }}>
          {['Escolher plano', 'Dados da empresa', 'Pagamento'].map((s, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              {i > 0 && <span style={{ color: '#cbd5e1', fontSize: 12 }}>›</span>}
              <span style={{ fontSize: 12, fontWeight: step === ['plan','company','payment'][i] ? 700 : 400, color: step === ['plan','company','payment'][i] ? '#0A3D8F' : '#94a3b8' }}>
                {s}
              </span>
            </div>
          ))}
        </div>
      </div>

      <div style={{ maxWidth: 900, margin: '32px auto', padding: '0 24px' }}>

        {/* STEP 1: Escolher plano */}
        {step === 'plan' && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: 32 }}>
              <p style={{ fontSize: 24, fontWeight: 800, color: '#1C1917', margin: '0 0 8px', letterSpacing: -0.5 }}>
                Escolha seu plano
              </p>
              <p style={{ fontSize: 14, color: '#64748b', margin: '0 0 20px' }}>
                Você testou {nichoPretty} por 7 dias. Continue com o plano ideal para o seu negócio.
              </p>

              {/* Toggle ciclo */}
              <div style={{ display: 'inline-flex', background: '#F0F4FF', borderRadius: 10, padding: 4, gap: 4 }}>
                {(['monthly', 'yearly'] as const).map(c => (
                  <button key={c} onClick={() => setCycle(c)} style={{
                    padding: '7px 18px', borderRadius: 8, border: 'none', cursor: 'pointer',
                    background: cycle === c ? 'white' : 'transparent',
                    color: cycle === c ? '#0A3D8F' : '#94a3b8',
                    fontSize: 13, fontWeight: 600,
                  }}>
                    {c === 'monthly' ? 'Mensal' : 'Anual'}
                    {c === 'yearly' && <span style={{ fontSize: 10, background: '#DCFCE7', color: '#16A34A', padding: '1px 6px', borderRadius: 100, marginLeft: 6, fontWeight: 700 }}>-20%</span>}
                  </button>
                ))}
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 24 }}>
              {PLANS.map(p => {
                const pPrice = cycle === 'yearly' ? p.priceYearly : p.price
                const isSelected = selectedPlan === p.id
                return (
                  <div key={p.id} onClick={() => setSelectedPlan(p.id)} style={{
                    background: 'white', borderRadius: 16, padding: 24,
                    border: isSelected ? '2px solid #0A3D8F' : '0.5px solid #e8e8e8',
                    cursor: 'pointer', position: 'relative',
                  }}>
                    {p.popular && (
                      <div style={{ position: 'absolute', top: -10, left: '50%', transform: 'translateX(-50%)', background: '#0A3D8F', color: 'white', fontSize: 10, fontWeight: 700, padding: '3px 12px', borderRadius: 100, whiteSpace: 'nowrap' }}>
                        Mais escolhido
                      </div>
                    )}
                    <p style={{ fontSize: 16, fontWeight: 800, color: '#1C1917', margin: '0 0 4px' }}>{p.name}</p>
                    <p style={{ fontSize: 11, color: '#94a3b8', margin: '0 0 16px' }}>{p.desc}</p>
                    <p style={{ fontSize: 30, fontWeight: 800, color: '#0A3D8F', margin: '0 0 2px', letterSpacing: -1 }}>
                      {fmt(pPrice)}
                    </p>
                    <p style={{ fontSize: 11, color: '#94a3b8', margin: '0 0 20px' }}>por mês</p>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
                      {p.features.map((f, i) => (
                        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
                          <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M1 6l3.5 3.5L10 2" stroke="#16A34A" strokeWidth="1.5" strokeLinecap="round"/></svg>
                          <span style={{ fontSize: 12, color: '#475569' }}>{f}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>

            <button onClick={() => setStep('company')} style={{ display: 'block', margin: '0 auto', background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 10, padding: '13px 40px', fontSize: 14, fontWeight: 700, cursor: 'pointer' }}>
              Continuar com o plano {plan.name} →
            </button>
          </div>
        )}

        {/* STEP 2: Dados da empresa */}
        {step === 'company' && (
          <div style={{ maxWidth: 520, margin: '0 auto' }}>
            <div style={{ marginBottom: 28 }}>
              <button onClick={() => setStep('plan')} style={{ background: 'none', border: 'none', color: '#0A3D8F', fontSize: 13, cursor: 'pointer', marginBottom: 12, padding: 0 }}>← Voltar</button>
              <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 6px', letterSpacing: -0.5 }}>Dados da empresa</p>
              <p style={{ fontSize: 13, color: '#64748b', margin: 0 }}>Necessário para emissão de notas e cobrança. Seus dados estão protegidos.</p>
            </div>

            <div style={{ background: 'white', borderRadius: 16, padding: 24, border: '0.5px solid #e8e8e8' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                {[
                  { key: 'company_name', label: 'Razão Social ou Nome Fantasia *', placeholder: 'Mecânica do João Ltda' },
                  { key: 'cnpj', label: 'CNPJ *', placeholder: '00.000.000/0001-00', format: formatCNPJ },
                  { key: 'phone', label: 'Telefone comercial', placeholder: '(11) 3333-4444' },
                  { key: 'address', label: 'Endereço', placeholder: 'Rua das Flores, 123' },
                  { key: 'city', label: 'Cidade', placeholder: 'São Paulo' },
                  { key: 'state', label: 'Estado (UF)', placeholder: 'SP' },
                ].map(field => (
                  <div key={field.key}>
                    <label style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 5 }}>
                      {field.label}
                    </label>
                    <input
                      placeholder={field.placeholder}
                      value={(form as any)[field.key]}
                      onChange={e => setForm(p => ({
                        ...p,
                        [field.key]: field.format ? field.format(e.target.value) : e.target.value
                      }))}
                      style={{ width: '100%', padding: '11px 13px', border: '0.5px solid #cbd5e1', borderRadius: 9, fontSize: 13, outline: 'none', boxSizing: 'border-box', color: '#1C1917' }}
                    />
                  </div>
                ))}
              </div>
            </div>

            <button
              onClick={() => setStep('payment')}
              disabled={!form.company_name || !form.cnpj}
              style={{
                width: '100%', marginTop: 16, padding: 14, borderRadius: 10, border: 'none',
                background: form.company_name && form.cnpj ? '#0A3D8F' : '#e0e4f0',
                color: form.company_name && form.cnpj ? 'white' : '#94a3b8',
                fontSize: 14, fontWeight: 700, cursor: 'pointer',
              }}
            >
              Ir para o pagamento →
            </button>
          </div>
        )}

        {/* STEP 3: Pagamento */}
        {step === 'payment' && (
          <div style={{ maxWidth: 520, margin: '0 auto' }}>
            <div style={{ marginBottom: 24 }}>
              <button onClick={() => setStep('company')} style={{ background: 'none', border: 'none', color: '#0A3D8F', fontSize: 13, cursor: 'pointer', marginBottom: 12, padding: 0 }}>← Voltar</button>
              <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 6px', letterSpacing: -0.5 }}>Pagamento</p>
            </div>

            {/* Resumo */}
            <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 12, padding: 16, marginBottom: 20 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <span style={{ fontSize: 14, fontWeight: 700, color: '#1C1917' }}>Plano {plan.name} · {cycle === 'yearly' ? 'Anual' : 'Mensal'}</span>
                <span style={{ fontSize: 16, fontWeight: 800, color: '#0A3D8F' }}>{fmt(price)}/mês</span>
              </div>
              <p style={{ fontSize: 12, color: '#64748b', margin: 0 }}>
                {form.company_name} · {form.cnpj}
              </p>
            </div>

            {/* Aviso NÃO pede cartão aqui — direciona para meio seguro */}
            <div style={{ background: 'white', border: '0.5px solid #e8e8e8', borderRadius: 12, padding: 20, marginBottom: 16, textAlign: 'center' }}>
              <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: '0 0 8px' }}>Escolha como pagar</p>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <button style={{ width: '100%', background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 10, padding: 12, fontSize: 13, fontWeight: 700, cursor: 'pointer' }}>
                  Pagar com PIX — aprovação imediata
                </button>
                <button style={{ width: '100%', background: 'white', color: '#0A3D8F', border: '1.5px solid #0A3D8F', borderRadius: 10, padding: 12, fontSize: 13, fontWeight: 700, cursor: 'pointer' }}>
                  Cartão de crédito
                </button>
                <button style={{ width: '100%', background: 'white', color: '#475569', border: '0.5px solid #e0e4f0', borderRadius: 10, padding: 12, fontSize: 13, cursor: 'pointer' }}>
                  Boleto bancário (3 dias úteis)
                </button>
              </div>
              <p style={{ fontSize: 10, color: '#94a3b8', marginTop: 12 }}>
                Pagamento processado com segurança · SSL · LGPD
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
