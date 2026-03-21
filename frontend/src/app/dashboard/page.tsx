'use client'
import { useState, useEffect, useRef } from 'react'
import { useRouter, usePathname } from 'next/navigation'

// ── CONFIGURAÇÃO POR NICHO ────────────────────────────────────────────────────

const NICHO_CONFIG: Record<string, {
  label: string
  icon: string
  concerns: string[]
  cards: string[]
}> = {
  mechanic: {
    label: 'Mecânica',
    icon: '🔧',
    concerns: [
      'OS atrasada ou aguardando aprovação',
      'Peça em falta no estoque',
      'Cliente sem retorno',
      'Conta vencendo hoje',
      'Quanto faturei hoje',
    ],
    cards: ['economy', 'revenue', 'os', 'stock', 'payables'],
  },
  bakery: {
    label: 'Padaria',
    icon: '🍞',
    concerns: [
      'Produto vencendo',
      'Caixa do dia',
      'Produção do dia',
      'Ingrediente em falta',
      'Melhor dia de venda',
    ],
    cards: ['economy', 'revenue', 'stock', 'cashflow', 'payables'],
  },
  industry: {
    label: 'Indústria',
    icon: '🏭',
    concerns: [
      'Matéria-prima em falta',
      'Pedido atrasado',
      'Custo de produção',
      'Faturamento do mês',
      'Contas a pagar',
    ],
    cards: ['economy', 'revenue', 'stock', 'payables', 'cashflow'],
  },
  logistics: {
    label: 'Logística',
    icon: '🚛',
    concerns: [
      'Entrega atrasada',
      'Rota do dia',
      'Motorista sem posição',
      'CT-e pendente',
      'Faturamento de fretes',
    ],
    cards: ['revenue', 'os', 'cashflow', 'payables'],
  },
  aesthetics: {
    label: 'Estética',
    icon: '💇',
    concerns: [
      'Agendamento do dia',
      'Produto vencendo',
      'Cliente sem retorno',
      'Faturamento da semana',
      'Conta vencendo',
    ],
    cards: ['economy', 'revenue', 'os', 'stock', 'payables'],
  },
  shoes: {
    label: 'Calçados',
    icon: '👟',
    concerns: [
      'Produto em falta na grade',
      'Venda do dia',
      'Estoque baixo',
      'Conta vencendo',
      'Melhor produto',
    ],
    cards: ['revenue', 'stock', 'payables', 'cashflow'],
  },
}

const CARD_MAP: Record<string, string[]> = {
  'OS atrasada ou aguardando aprovação': ['os'],
  'Peça em falta no estoque': ['stock'],
  'Cliente sem retorno': ['os'],
  'Conta vencendo hoje': ['payables'],
  'Quanto faturei hoje': ['revenue'],
  'Produto vencendo': ['stock'],
  'Caixa do dia': ['cashflow'],
  'Produção do dia': ['os'],
  'Ingrediente em falta': ['stock'],
  'Melhor dia de venda': ['revenue'],
  'Matéria-prima em falta': ['stock'],
  'Pedido atrasado': ['os'],
  'Custo de produção': ['economy'],
  'Faturamento do mês': ['revenue'],
  'Contas a pagar': ['payables'],
  'Entrega atrasada': ['os'],
  'Rota do dia': ['os'],
  'Motorista sem posição': ['os'],
  'CT-e pendente': ['os'],
  'Faturamento de fretes': ['revenue'],
  'Agendamento do dia': ['os'],
  'Cliente sem retorno ': ['os'],
  'Faturamento da semana': ['revenue'],
  'Conta vencendo': ['payables'],
  'Produto em falta na grade': ['stock'],
  'Venda do dia': ['revenue'],
  'Estoque baixo': ['stock'],
  'Melhor produto': ['revenue'],
}

// ── TIPOS ─────────────────────────────────────────────────────────────────────

interface DashCard {
  id: string
  type: string
  enabled: boolean
}

interface WeekDay {
  day: string
  value: number
  economy: number
  isBest: boolean
  isToday: boolean
}

// ── COMPONENTE ONBOARDING ─────────────────────────────────────────────────────

function OnboardingModal({ onComplete, businessType }: {
  onComplete: (cards: DashCard[]) => void
  businessType: string
}) {
  const [step, setStep] = useState(0)
  const [agentMsg, setAgentMsg] = useState('')
  const [typing, setTyping] = useState(false)
  const [selectedNicho, setSelectedNicho] = useState(businessType || '')
  const [selectedConcerns, setSelectedConcerns] = useState<string[]>([])
  const [customInput, setCustomInput] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const STEPS = [
    {
      msg: `Olá! Sou seu Co-Piloto de gestão. Vou montar seu painel personalizado agora. Primeiro: qual é o seu negócio?`,
      action: 'nicho',
    },
    {
      msg: `Perfeito! Agora me diz: o que você mais precisa acompanhar todo dia? Pode escolher mais de um.`,
      action: 'concerns',
    },
    {
      msg: `Montei seu painel do jeito que você precisa. Qualquer coisa é só me chamar! 😊`,
      action: 'done',
    },
  ]

  useEffect(() => {
    typeMessage(STEPS[0].msg)
  }, [])

  const typeMessage = (msg: string) => {
    setTyping(true)
    setAgentMsg('')
    let i = 0
    const interval = setInterval(() => {
      setAgentMsg(msg.slice(0, i + 1))
      i++
      if (i >= msg.length) {
        clearInterval(interval)
        setTyping(false)
      }
    }, 15)
  }

  const handleNichoSelect = (nicho: string) => {
    setSelectedNicho(nicho)
    localStorage.setItem('nexo_business_type', nicho)
    setTimeout(() => {
      setStep(1)
      typeMessage(STEPS[1].msg)
    }, 300)
  }

  const toggleConcern = (concern: string) => {
    setSelectedConcerns(prev =>
      prev.includes(concern)
        ? prev.filter(c => c !== concern)
        : [...prev, concern]
    )
  }

  const handleConfirm = () => {
    // Montar cards baseado nas preocupações selecionadas
    const cardTypes = new Set<string>(['economy']) // economy sempre ativo
    selectedConcerns.forEach(concern => {
      CARD_MAP[concern]?.forEach(card => cardTypes.add(card))
    })

    if (customInput.trim()) {
      const lower = customInput.toLowerCase()
      if (lower.includes('fatur') || lower.includes('receita') || lower.includes('dinheiro')) cardTypes.add('revenue')
      if (lower.includes('estoque') || lower.includes('peça') || lower.includes('produto')) cardTypes.add('stock')
      if (lower.includes('os') || lower.includes('ordem') || lower.includes('serviço')) cardTypes.add('os')
      if (lower.includes('conta') || lower.includes('pagar') || lower.includes('boleto')) cardTypes.add('payables')
      if (lower.includes('caixa') || lower.includes('fluxo') || lower.includes('saldo')) cardTypes.add('cashflow')
    }

    const cards: DashCard[] = [
      'economy', 'revenue', 'os', 'stock', 'payables', 'cashflow'
    ].map(type => ({
      id: type,
      type,
      enabled: cardTypes.has(type),
    }))

    setStep(2)
    typeMessage(STEPS[2].msg)
    setTimeout(() => onComplete(cards), 2500)
  }

  const nicho = NICHO_CONFIG[selectedNicho]

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.55)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', padding: 24 }}>
      <div style={{ background: 'white', borderRadius: 20, padding: 28, width: '100%', maxWidth: 540, marginBottom: 16 }}>

        {/* Header do agente */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18 }}>
          <div style={{ width: 42, height: 42, background: '#0A3D8F', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20 }}>🤖</div>
          <div>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>Co-Piloto IA</p>
            <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>Montando seu painel personalizado</p>
          </div>
          {/* Steps indicator */}
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 6 }}>
            {[0, 1, 2].map(i => (
              <div key={i} style={{ width: 8, height: 8, borderRadius: '50%', background: i <= step ? '#0A3D8F' : '#E0E4F0' }} />
            ))}
          </div>
        </div>

        {/* Mensagem do agente */}
        <div style={{ background: '#F0F4FF', borderRadius: 12, padding: '14px 16px', marginBottom: 20, minHeight: 56 }}>
          <p style={{ fontSize: 14, color: '#1C1917', lineHeight: 1.6, margin: 0 }}>
            {agentMsg}
            {typing && <span style={{ opacity: 0.5 }}>|</span>}
          </p>
        </div>

        {/* Etapa 1 — Escolher nicho */}
        {step === 0 && !typing && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 8 }}>
            {Object.entries(NICHO_CONFIG).map(([key, cfg]) => (
              <button key={key} onClick={() => handleNichoSelect(key)} style={{
                background: selectedNicho === key ? '#E3F2FD' : '#F8F7F4',
                border: selectedNicho === key ? '1.5px solid #0A3D8F' : '1px solid #e0e4f0',
                borderRadius: 10, padding: '12px 8px', cursor: 'pointer',
                textAlign: 'center',
              }}>
                <p style={{ fontSize: 22, margin: '0 0 4px' }}>{cfg.icon}</p>
                <p style={{ fontSize: 12, fontWeight: 600, color: selectedNicho === key ? '#0A3D8F' : '#1C1917', margin: 0 }}>{cfg.label}</p>
              </button>
            ))}
          </div>
        )}

        {/* Etapa 2 — Escolher preocupações */}
        {step === 1 && !typing && nicho && (
          <div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14 }}>
              {nicho.concerns.map(concern => (
                <button key={concern} onClick={() => toggleConcern(concern)} style={{
                  display: 'flex', alignItems: 'center', gap: 10,
                  background: selectedConcerns.includes(concern) ? '#E3F2FD' : '#F8F7F4',
                  border: selectedConcerns.includes(concern) ? '1.5px solid #0A3D8F' : '1px solid #e0e4f0',
                  borderRadius: 10, padding: '11px 14px', cursor: 'pointer', textAlign: 'left',
                }}>
                  <div style={{ width: 18, height: 18, borderRadius: 4, border: '1.5px solid', borderColor: selectedConcerns.includes(concern) ? '#0A3D8F' : '#cbd5e1', background: selectedConcerns.includes(concern) ? '#0A3D8F' : 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    {selectedConcerns.includes(concern) && <svg width="10" height="10" viewBox="0 0 10 10"><path d="M1 5l3 3 5-5" stroke="white" strokeWidth="1.5" fill="none" strokeLinecap="round"/></svg>}
                  </div>
                  <span style={{ fontSize: 13, color: '#1C1917', fontWeight: selectedConcerns.includes(concern) ? 600 : 400 }}>{concern}</span>
                </button>
              ))}
            </div>

            {/* Input livre */}
            <input
              ref={inputRef}
              value={customInput}
              onChange={e => setCustomInput(e.target.value)}
              placeholder="Tem algo mais? Escreva aqui..."
              style={{ width: '100%', border: '1px solid #e0e4f0', borderRadius: 10, padding: '10px 14px', fontSize: 13, outline: 'none', marginBottom: 12, boxSizing: 'border-box' }}
            />

            <button
              onClick={handleConfirm}
              disabled={selectedConcerns.length === 0 && !customInput.trim()}
              style={{
                width: '100%', padding: 13, borderRadius: 10, border: 'none',
                background: selectedConcerns.length > 0 || customInput.trim() ? '#0A3D8F' : '#e0e4f0',
                color: selectedConcerns.length > 0 || customInput.trim() ? 'white' : '#94a3b8',
                fontSize: 14, fontWeight: 700, cursor: 'pointer',
              }}
            >
              Montar meu painel →
            </button>
          </div>
        )}

        {/* Etapa 3 — Confirmação */}
        {step === 2 && (
          <div style={{ textAlign: 'center', padding: '8px 0' }}>
            <p style={{ fontSize: 32, margin: '0 0 8px' }}>🎉</p>
            <p style={{ fontSize: 14, color: '#16A34A', fontWeight: 600, margin: 0 }}>Painel personalizado criado!</p>
          </div>
        )}

        {step < 2 && (
          <button onClick={() => { localStorage.setItem('nexo_dash_configured', 'true'); window.location.reload() }} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer', marginTop: 12, display: 'block', width: '100%', textAlign: 'center' }}>
            Pular por agora
          </button>
        )}
      </div>
    </div>
  )
}

// ── BREADCRUMB ────────────────────────────────────────────────────────────────

function Breadcrumb() {
  const router = useRouter()
  const pathname = usePathname()

  const LABELS: Record<string, string> = {
    '/dashboard': 'Dashboard',
    '/finance': 'Financeiro',
    '/mechanic': 'Mecânica',
    '/bakery': 'Padaria',
    '/inventory': 'Estoque',
    '/payables': 'Contas a Pagar',
    '/receivables': 'Contas a Receber',
    '/dispatch': 'Despacho em Lote',
    '/logistics': 'Logística',
    '/aesthetics': 'Estética',
    '/shoes': 'Calçados',
    '/enterprise': 'Enterprise',
    '/modules': 'Módulos',
  }

  const parts = pathname.split('/').filter(Boolean)
  const crumbs = [{ label: 'Dashboard', path: '/dashboard' }]
  let current = ''
  parts.forEach(part => {
    current += '/' + part
    if (current !== '/dashboard' && LABELS[current]) {
      crumbs.push({ label: LABELS[current], path: current })
    }
  })

  if (crumbs.length <= 1) return null

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 16 }}>
      <button onClick={() => router.back()} style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 10px', fontSize: 12, color: '#0A3D8F', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 4 }}>
        ← Voltar
      </button>
      <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
        {crumbs.map((crumb, i) => (
          <span key={crumb.path} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            {i > 0 && <span style={{ color: '#cbd5e1', fontSize: 12 }}>›</span>}
            <span
              onClick={() => router.push(crumb.path)}
              style={{ fontSize: 12, color: i === crumbs.length - 1 ? '#1C1917' : '#0A3D8F', fontWeight: i === crumbs.length - 1 ? 600 : 400, cursor: i === crumbs.length - 1 ? 'default' : 'pointer' }}
            >
              {crumb.label}
            </span>
          </span>
        ))}
      </div>
    </div>
  )
}

// ── TABELA DE FATURAMENTO ─────────────────────────────────────────────────────

function RevenueTable({ data }: { data: WeekDay[] }) {
  const best = data.reduce((a, b) => a.value > b.value ? a : b)

  return (
    <div style={{ background: 'white', borderRadius: 16, padding: 20, border: '0.5px solid #e8e8e8' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>Faturamento por dia</p>
        <span style={{ fontSize: 11, color: '#94a3b8' }}>esta semana</span>
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr>
            {['Dia', 'Faturamento', 'Imposto Economizado', ''].map(h => (
              <th key={h} style={{ fontSize: 10, color: '#94a3b8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', padding: '0 0 10px', textAlign: h === 'Faturamento' || h === 'Imposto Economizado' ? 'right' : 'left' }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr key={i} style={{ background: row.isBest ? '#F0FDF4' : row.isToday ? '#F0F4FF' : 'transparent', borderRadius: 8 }}>
              <td style={{ padding: '10px 0', fontSize: 13, fontWeight: row.isToday ? 700 : 400, color: '#1C1917', borderRadius: '8px 0 0 8px' }}>
                {row.day}
                {row.isToday && <span style={{ fontSize: 10, color: '#0A3D8F', marginLeft: 6, fontWeight: 600 }}>hoje</span>}
              </td>
              <td style={{ padding: '10px 8px', fontSize: 13, fontWeight: 700, color: '#1C1917', textAlign: 'right' }}>
                {row.value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
              </td>
              <td style={{ padding: '10px 8px', fontSize: 13, color: '#16A34A', fontWeight: 600, textAlign: 'right' }}>
                +{row.economy.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
              </td>
              <td style={{ padding: '10px 0', textAlign: 'right', borderRadius: '0 8px 8px 0' }}>
                {row.isBest && <span style={{ fontSize: 10, background: '#DCFCE7', color: '#16A34A', padding: '2px 8px', borderRadius: 100, fontWeight: 700 }}>🏆 melhor dia</span>}
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr style={{ borderTop: '0.5px solid #e8e8e8' }}>
            <td style={{ padding: '12px 0', fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Total</td>
            <td style={{ padding: '12px 8px', fontSize: 14, fontWeight: 800, color: '#0A3D8F', textAlign: 'right' }}>
              {data.reduce((s, r) => s + r.value, 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
            </td>
            <td style={{ padding: '12px 8px', fontSize: 13, fontWeight: 700, color: '#16A34A', textAlign: 'right' }}>
              +{data.reduce((s, r) => s + r.economy, 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
            </td>
            <td />
          </tr>
        </tfoot>
      </table>
    </div>
  )
}

// ── DASHBOARD PRINCIPAL ───────────────────────────────────────────────────────

export default function DashboardPage() {
  const router = useRouter()
  const [showOnboarding, setShowOnboarding] = useState(false)
  const [cards, setCards] = useState<DashCard[]>([
    { id: 'economy', type: 'economy', enabled: true },
    { id: 'revenue', type: 'revenue', enabled: true },
    { id: 'os', type: 'os', enabled: true },
  ])
  const [userName, setUserName] = useState('João')
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''
  const businessType = typeof window !== 'undefined' ? localStorage.getItem('nexo_business_type') || 'mechanic' : 'mechanic'

  const weekData: WeekDay[] = [
    { day: 'Segunda', value: 3200, economy: 480, isBest: false, isToday: false },
    { day: 'Terça',   value: 4100, economy: 615, isBest: true,  isToday: false },
    { day: 'Quarta',  value: 2800, economy: 420, isBest: false, isToday: false },
    { day: 'Quinta',  value: 5200, economy: 780, isBest: false, isToday: false },
    { day: 'Sexta',   value: 4800, economy: 720, isBest: false, isToday: false },
    { day: 'Sábado',  value: 3100, economy: 465, isBest: false, isToday: true  },
    { day: 'Domingo', value: 0,    economy: 0,   isBest: false, isToday: false },
  ]

  useEffect(() => {
    const configured = localStorage.getItem('nexo_dash_configured')
    const nichoJaEscolhido = localStorage.getItem('nexo_business_type')
    // Se o nicho ja foi escolhido na tela de entrada, nao mostra onboarding de novo
    if (!configured && !nichoJaEscolhido) {
      setTimeout(() => setShowOnboarding(true), 600)
    } else if (!configured && nichoJaEscolhido) {
      // Nicho ja escolhido na entrada — usar cards padrao do nicho
      localStorage.setItem('nexo_dash_configured', 'true')
    } else {
      const saved = localStorage.getItem('nexo_dash_cards')
      if (saved) setCards(JSON.parse(saved))
    }
    const name = localStorage.getItem('nexo_user_name')
    if (name) setUserName(name)
  }, [])

  const handleOnboardingComplete = (newCards: DashCard[]) => {
    setCards(newCards)
    localStorage.setItem('nexo_dash_cards', JSON.stringify(newCards))
    localStorage.setItem('nexo_dash_configured', 'true')
    setShowOnboarding(false)
  }

  const greeting = () => {
    const h = new Date().getHours()
    if (h < 12) return 'Bom dia'
    if (h < 18) return 'Boa tarde'
    return 'Boa noite'
  }

  const today = new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })
  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

  return (
    <div style={{ background: '#F8F7F4', minHeight: '100vh', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>

      {/* Onboarding */}
      {showOnboarding && (
        <OnboardingModal
          onComplete={handleOnboardingComplete}
          businessType={businessType}
        />
      )}

      {/* Topbar */}
      <div style={{ background: 'white', borderBottom: '0.5px solid #e8e8e8', padding: '12px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', position: 'sticky', top: 0, zIndex: 10 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ width: 30, height: 30, background: '#0A3D8F', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <svg width="16" height="16" viewBox="0 0 64 64"><line x1="20" y1="20" x2="32" y2="28" stroke="white" strokeWidth="2"/><line x1="32" y1="28" x2="44" y2="20" stroke="white" strokeWidth="2"/><line x1="16" y1="34" x2="32" y2="28" stroke="white" strokeWidth="2"/><line x1="32" y1="28" x2="48" y2="34" stroke="white" strokeWidth="2"/><line x1="24" y1="36" x2="32" y2="44" stroke="white" strokeWidth="2"/><line x1="40" y1="36" x2="32" y2="44" stroke="white" strokeWidth="2"/><circle cx="32" cy="28" r="5" fill="white"/><circle cx="32" cy="44" r="4" fill="white"/><circle cx="20" cy="20" r="3" fill="white"/><circle cx="44" cy="20" r="3" fill="white"/></svg>
          </div>
          <span style={{ fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Gestão Para Todos</span>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#0A3D8F', fontWeight: 600 }}>⚡ IBS/CBS 2026 ativo</div>
          <button onClick={() => { localStorage.removeItem('nexo_dash_configured'); setShowOnboarding(true) }} style={{ background: 'none', border: '1px solid #e0e4f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#94a3b8', cursor: 'pointer' }}>
            ⚙️ Personalizar painel
          </button>
        </div>
      </div>

      <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>

        {/* Breadcrumb */}
        <Breadcrumb />

        {/* Saudação + alerta */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
          <div>
            <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>{greeting()}, {userName}! 👋</p>
            <p style={{ fontSize: 13, color: '#94a3b8', margin: 0, textTransform: 'capitalize' }}>{today}</p>
          </div>
          <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: 12, padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ fontSize: 18 }}>🌙</span>
            <div>
              <p style={{ fontSize: 12, fontWeight: 700, color: '#854D0E', margin: '0 0 2px' }}>1 serviço chegou essa noite</p>
              <p style={{ fontSize: 11, color: '#A16207', margin: 0 }}>Honda Civic — troca de óleo · 02:14</p>
            </div>
            <button onClick={() => router.push('/mechanic')} style={{ background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 8, padding: '6px 12px', fontSize: 11, fontWeight: 600, cursor: 'pointer', marginLeft: 8 }}>
              Abrir OS
            </button>
          </div>
        </div>

        {/* Cards dinâmicos */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>

          {/* Card economia fiscal — sempre primeiro */}
          <div style={{ background: 'linear-gradient(135deg, #0A3D8F, #1565C0)', borderRadius: 16, padding: '22px 20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: 'rgba(255,255,255,0.65)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Você economizou</span>
              <span style={{ background: 'rgba(74,222,128,0.2)', color: '#4ADE80', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>este mês</span>
            </div>
            <p style={{ fontSize: 32, fontWeight: 800, color: 'white', margin: '0 0 4px', letterSpacing: -1 }}>R$ 847</p>
            <p style={{ fontSize: 12, color: 'rgba(255,255,255,0.55)', margin: '0 0 14px' }}>em crédito IBS/CBS abatido</p>
            <div style={{ background: 'rgba(255,255,255,0.1)', borderRadius: 8, padding: '8px 12px', fontSize: 11, color: 'rgba(255,255,255,0.8)' }}>
              💡 Sem o sistema, você pagaria esse valor extra
            </div>
          </div>

          {/* Card faturamento */}
          {cards.find(c => c.type === 'revenue' && c.enabled) && (
            <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
                <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Faturamento do mês</span>
                <span style={{ color: '#16A34A', fontSize: 12, fontWeight: 700 }}>↑ 12%</span>
              </div>
              <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>R$ 8.100</p>
              <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 16px' }}>Meta: R$ 10.000 · 81% atingido</p>
              <div style={{ height: 6, background: '#F0F4FF', borderRadius: 3 }}>
                <div style={{ height: '100%', width: '81%', background: '#0A3D8F', borderRadius: 3 }} />
              </div>
            </div>
          )}

          {/* Card OS */}
          {cards.find(c => c.type === 'os' && c.enabled) && (
            <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
                <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Ordens de Serviço</span>
                <span style={{ background: '#FEF9C3', color: '#854D0E', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>hoje</span>
              </div>
              <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>5</p>
              <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 14px' }}>2 abertas · 1 aguardando aprovação</p>
              <div style={{ display: 'flex', gap: 6 }}>
                <div style={{ flex: 1, background: '#F0FDF4', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#16A34A' }}>2 abertas</div>
                <div style={{ flex: 1, background: '#FEF9C3', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#854D0E' }}>1 aguard.</div>
                <div style={{ flex: 1, background: '#F0F4FF', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#0A3D8F' }}>2 concl.</div>
              </div>
            </div>
          )}

          {/* Outros cards opcionais */}
          {cards.filter(c => c.enabled && !['economy','revenue','os'].includes(c.type)).map(card => (
            <div key={card.id} style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase' }}>
                {card.type === 'stock' ? '📦 Estoque' : card.type === 'payables' ? '📋 Contas a Pagar' : '💰 Fluxo de Caixa'}
              </span>
              <p style={{ fontSize: 28, fontWeight: 800, color: '#1C1917', margin: '12px 0 4px' }}>—</p>
              <p style={{ fontSize: 12, color: '#94a3b8', margin: 0 }}>carregando dados...</p>
            </div>
          ))}
        </div>

        {/* Tabela de faturamento por dia */}
        <RevenueTable data={weekData} />
      </div>

      <style>{`
        @keyframes pulseGreen {
          0%,100% { box-shadow: 0 0 0 0 rgba(22,163,74,0.3); }
          50%      { box-shadow: 0 0 0 8px rgba(22,163,74,0); }
        }
      `}</style>
    </div>
  )
}
