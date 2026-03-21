'use client'
import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'

// ── TIPOS ─────────────────────────────────────────────────────────────────────

interface DashCard {
  id: string
  type: 'economy' | 'revenue' | 'os' | 'stock' | 'payables' | 'cashflow'
  enabled: boolean
}

interface OSItem {
  id: string
  vehicle: string
  client: string
  service: string
  status: string
  value: number
  arrivedAt?: string
}

const DEFAULT_CARDS: DashCard[] = [
  { id: 'economy',  type: 'economy',  enabled: true },
  { id: 'revenue',  type: 'revenue',  enabled: true },
  { id: 'os',       type: 'os',       enabled: true },
  { id: 'stock',    type: 'stock',    enabled: false },
  { id: 'payables', type: 'payables', enabled: false },
  { id: 'cashflow', type: 'cashflow', enabled: false },
]

const CARD_KEYWORDS: Record<string, string[]> = {
  economy:  ['imposto', 'economiz', 'crédito', 'ibs', 'cbs', 'fiscal'],
  revenue:  ['faturei', 'faturamento', 'receita', 'dinheiro', 'vendas'],
  os:       ['os', 'ordem', 'serviço', 'agendamento', 'carro', 'veículo'],
  stock:    ['estoque', 'peça', 'produto', 'material'],
  payables: ['conta', 'boleto', 'vencendo', 'pagar', 'despesa'],
  cashflow: ['caixa', 'saldo', 'fluxo', 'dinheiro em caixa'],
}

function parseCardsFromAnswer(answer: string): DashCard[] {
  const lower = answer.toLowerCase()
  return DEFAULT_CARDS.map(card => ({
    ...card,
    enabled: CARD_KEYWORDS[card.type]?.some(kw => lower.includes(kw)) ?? card.enabled,
  }))
}

// ── COMPONENTE PRINCIPAL ──────────────────────────────────────────────────────

export default function DashboardPage() {
  const router = useRouter()
  const [cards, setCards] = useState<DashCard[]>(DEFAULT_CARDS)
  const [showOnboarding, setShowOnboarding] = useState(false)
  const [onboardingStep, setOnboardingStep] = useState(0)
  const [userAnswer, setUserAnswer] = useState('')
  const [agentTyping, setAgentTyping] = useState(false)
  const [agentMsg, setAgentMsg] = useState('')
  const [nightOS, setNightOS] = useState<OSItem[]>([])
  const [userName, setUserName] = useState('João')
  const inputRef = useRef<HTMLInputElement>(null)
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  useEffect(() => {
    // Verificar se é primeira vez
    const configured = localStorage.getItem('nexo_dash_configured')
    if (!configured) {
      setTimeout(() => {
        setShowOnboarding(true)
        typeMessage('Olá! Sou seu assistente de gestão. Para montar seu painel do jeito que você precisa, me diz: o que você mais quer ver todo dia quando abre o sistema?')
      }, 800)
    } else {
      const saved = localStorage.getItem('nexo_dash_cards')
      if (saved) setCards(JSON.parse(saved))
    }

    // Buscar OS da noite
    fetchNightOS()

    // Nome do usuário
    const name = localStorage.getItem('nexo_user_name')
    if (name) setUserName(name)
  }, [])

  const typeMessage = (msg: string) => {
    setAgentTyping(true)
    setAgentMsg('')
    let i = 0
    const interval = setInterval(() => {
      setAgentMsg(msg.slice(0, i + 1))
      i++
      if (i >= msg.length) {
        clearInterval(interval)
        setAgentTyping(false)
        setTimeout(() => inputRef.current?.focus(), 100)
      }
    }, 18)
  }

  const fetchNightOS = async () => {
    // Demo: simular OS que chegou durante a noite
    setNightOS([
      { id: '1', vehicle: 'Honda Civic', client: 'João Silva', service: 'Troca de óleo', status: 'open', value: 180, arrivedAt: '02:14' },
    ])
  }

  const handleOnboardingSubmit = async () => {
    if (!userAnswer.trim()) return
    const answer = userAnswer
    setUserAnswer('')
    setOnboardingStep(1)

    // Interpretar resposta
    setAgentTyping(true)
    await new Promise(r => setTimeout(r, 1200))

    const newCards = parseCardsFromAnswer(answer)

    // Tentar usar IA para interpretar (opcional)
    try {
      const res = await fetch('https://api.anthropic.com/v1/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: 'claude-sonnet-4-20250514',
          max_tokens: 200,
          messages: [{
            role: 'user',
            content: `O usuário disse: "${answer}". Quais destes cards ele quer ver no dashboard? Responda SOMENTE com JSON: {"economy":true/false,"revenue":true/false,"os":true/false,"stock":true/false,"payables":true/false,"cashflow":true/false}. economy=crédito fiscal IBS/CBS, revenue=faturamento, os=ordens de serviço, stock=estoque, payables=contas a pagar, cashflow=fluxo de caixa.`
          }]
        })
      })
      if (res.ok) {
        const data = await res.json()
        const text = data.content?.[0]?.text || ''
        const json = JSON.parse(text.replace(/```json|```/g, '').trim())
        const aiCards = DEFAULT_CARDS.map(c => ({ ...c, enabled: json[c.type] ?? c.enabled }))
        setCards(aiCards)
        localStorage.setItem('nexo_dash_cards', JSON.stringify(aiCards))
      } else {
        setCards(newCards)
        localStorage.setItem('nexo_dash_cards', JSON.stringify(newCards))
      }
    } catch {
      setCards(newCards)
      localStorage.setItem('nexo_dash_cards', JSON.stringify(newCards))
    }

    localStorage.setItem('nexo_dash_configured', 'true')
    setAgentTyping(false)
    typeMessage('Perfeito! Montei seu painel com o que você precisa. Você pode personalizar quando quiser. Estou aqui se precisar de mim! 😊')
    setTimeout(() => setShowOnboarding(false), 3000)
  }

  const greeting = () => {
    const h = new Date().getHours()
    if (h < 12) return 'Bom dia'
    if (h < 18) return 'Boa tarde'
    return 'Boa noite'
  }

  const today = new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })

  return (
    <div style={{ background: '#F8F7F4', minHeight: '100vh', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>

      {/* ── ONBOARDING OVERLAY ── */}
      {showOnboarding && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 100, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', padding: 24 }}>
          <div style={{ background: 'white', borderRadius: 20, padding: 28, width: '100%', maxWidth: 520, marginBottom: 20 }}>

            {/* Avatar do agente */}
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
              <div style={{ width: 40, height: 40, background: '#0A3D8F', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18 }}>🤖</div>
              <div>
                <p style={{ fontSize: 13, fontWeight: 600, color: '#1C1917', margin: 0 }}>Co-Piloto IA</p>
                <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>Assistente de gestão</p>
              </div>
            </div>

            {/* Mensagem do agente */}
            <div style={{ background: '#F0F4FF', borderRadius: 12, padding: '14px 16px', marginBottom: 20, minHeight: 60 }}>
              <p style={{ fontSize: 14, color: '#1C1917', lineHeight: 1.6, margin: 0 }}>
                {agentMsg}
                {agentTyping && <span style={{ animation: 'blink 1s infinite' }}>|</span>}
              </p>
            </div>

            {/* Sugestões rápidas */}
            {onboardingStep === 0 && !agentTyping && (
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 14 }}>
                {[
                  'Faturamento e OS do dia',
                  'Imposto economizado',
                  'Estoque e contas a pagar',
                ].map(s => (
                  <button key={s} onClick={() => setUserAnswer(s)} style={{
                    background: '#F0F4FF', border: '1px solid #dde3f0',
                    borderRadius: 100, padding: '6px 14px', fontSize: 12,
                    color: '#0A3D8F', cursor: 'pointer', fontWeight: 500,
                  }}>
                    {s}
                  </button>
                ))}
              </div>
            )}

            {/* Input */}
            {onboardingStep === 0 && (
              <div style={{ display: 'flex', gap: 10 }}>
                <input
                  ref={inputRef}
                  value={userAnswer}
                  onChange={e => setUserAnswer(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleOnboardingSubmit()}
                  placeholder="Digite o que quer ver todos os dias..."
                  style={{
                    flex: 1, border: '1px solid #e0e4f0', borderRadius: 10,
                    padding: '10px 14px', fontSize: 13, outline: 'none',
                    color: '#1C1917',
                  }}
                />
                <button onClick={handleOnboardingSubmit} style={{
                  background: '#0A3D8F', color: 'white', border: 'none',
                  borderRadius: 10, padding: '10px 18px', cursor: 'pointer',
                  fontSize: 14, fontWeight: 600,
                }}>→</button>
              </div>
            )}

            {/* Skip */}
            <button onClick={() => { setShowOnboarding(false); localStorage.setItem('nexo_dash_configured', 'true') }} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer', marginTop: 12, display: 'block', width: '100%', textAlign: 'center' }}>
              Pular por agora
            </button>
          </div>
        </div>
      )}

      {/* ── TOPBAR ── */}
      <div style={{ background: 'white', borderBottom: '0.5px solid #e8e8e8', padding: '12px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ width: 30, height: 30, background: '#0A3D8F', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <svg width="16" height="16" viewBox="0 0 64 64"><line x1="20" y1="20" x2="32" y2="28" stroke="white" stroke-width="2"/><line x1="32" y1="28" x2="44" y2="20" stroke="white" stroke-width="2"/><line x1="16" y1="34" x2="32" y2="28" stroke="white" stroke-width="2"/><line x1="32" y1="28" x2="48" y2="34" stroke="white" stroke-width="2"/><line x1="24" y1="36" x2="32" y2="44" stroke="white" stroke-width="2"/><line x1="40" y1="36" x2="32" y2="44" stroke="white" stroke-width="2"/><circle cx="32" cy="28" r="5" fill="white"/><circle cx="32" cy="44" r="4" fill="white"/><circle cx="20" cy="20" r="3" fill="white"/><circle cx="44" cy="20" r="3" fill="white"/></svg>
          </div>
          <span style={{ fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Gestão Para Todos</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#0A3D8F', fontWeight: 600 }}>
            ⚡ IBS/CBS 2026 ativo
          </div>
          <button onClick={() => { localStorage.removeItem('nexo_dash_configured'); setShowOnboarding(true); setOnboardingStep(0); typeMessage('Vamos reconfigurar seu painel! O que você quer ver todo dia?') }} style={{ background: 'none', border: '1px solid #e0e4f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#94a3b8', cursor: 'pointer' }}>
            ⚙️ Personalizar
          </button>
        </div>
      </div>

      {/* ── CONTEÚDO ── */}
      <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>

        {/* Saudação + alerta noturno */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
          <div>
            <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>
              {greeting()}, {userName}! 👋
            </p>
            <p style={{ fontSize: 13, color: '#94a3b8', margin: 0, textTransform: 'capitalize' }}>{today}</p>
          </div>

          {nightOS.length > 0 && (
            <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: 12, padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ fontSize: 20 }}>🌙</span>
              <div>
                <p style={{ fontSize: 12, fontWeight: 700, color: '#854D0E', margin: '0 0 2px' }}>
                  {nightOS.length} serviço{nightOS.length > 1 ? 's' : ''} chegou{nightOS.length > 1 ? 'ram' : ''} essa noite
                </p>
                <p style={{ fontSize: 11, color: '#A16207', margin: 0 }}>{nightOS[0].vehicle} — {nightOS[0].service}</p>
              </div>
            </div>
          )}
        </div>

        {/* Cards dinâmicos */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>
          {cards.filter(c => c.enabled).map(card => (
            <DashCardComponent key={card.id} type={card.type} token={token} />
          ))}
        </div>

        {/* Ação necessária */}
        {nightOS.length > 0 && (
          <div style={{ background: 'white', borderRadius: 16, padding: 20, border: '0.5px solid #e8e8e8', marginBottom: 20 }}>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: '0 0 14px' }}>Ação necessária agora</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {nightOS.map(os => (
                <div key={os.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 14px', background: '#FFFBEB', borderRadius: 10, border: '1px solid #FDE68A' }}>
                  <div>
                    <p style={{ fontSize: 13, fontWeight: 600, color: '#1C1917', margin: 0 }}>{os.vehicle} — {os.client}</p>
                    <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>{os.service} · chegou {os.arrivedAt}</p>
                  </div>
                  <button onClick={() => router.push('/mechanic')} style={{ background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 8, padding: '7px 14px', fontSize: 12, fontWeight: 600, cursor: 'pointer' }}>
                    Abrir OS
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Gráfico semanal */}
        <WeeklyChart token={token} />
      </div>

      <style>{`
        @keyframes blink { 0%,100%{opacity:1} 50%{opacity:0} }
        @keyframes pulseGreen {
          0%,100% { box-shadow: 0 0 0 0 rgba(22,163,74,0.3); }
          50%      { box-shadow: 0 0 0 8px rgba(22,163,74,0); }
        }
      `}</style>
    </div>
  )
}

// ── CARD DINÂMICO ─────────────────────────────────────────────────────────────

function DashCardComponent({ type, token }: { type: string; token: string }) {
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    fetchCardData()
  }, [type])

  const fetchCardData = async () => {
    const h = { Authorization: `Bearer ${token}` }
    try {
      switch (type) {
        case 'economy': {
          const res = await fetch('/api/v1/expenses/summary', { headers: h })
          if (res.ok) setData(await res.json())
          else setData({ total_credit: 847, month: 'Março' })
          break
        }
        case 'revenue': {
          const res = await fetch('/api/v1/finance/dre', { headers: h })
          if (res.ok) setData(await res.json())
          else setData({ gross_revenue: 8100, net_result: 1820, net_margin: 22.5 })
          break
        }
        case 'os': {
          const res = await fetch('/api/v1/mechanic/os', { headers: h })
          if (res.ok) {
            const json = await res.json()
            setData({ total: json.total || 5, open: 2, pending: 1, done: 2 })
          } else {
            setData({ total: 5, open: 2, pending: 1, done: 2 })
          }
          break
        }
        case 'stock': {
          const res = await fetch('/api/v1/inventory/products/summary', { headers: h })
          if (res.ok) setData(await res.json())
          else setData({ low_stock_count: 3, out_of_stock_count: 1, total_value: 12450 })
          break
        }
        case 'payables': {
          const res = await fetch('/api/v1/payables', { headers: h })
          if (res.ok) setData(await res.json())
          else setData({ overdue_count: 1, due_today: 2, total_pending: 3850 })
          break
        }
        case 'cashflow': {
          const res = await fetch('/api/v1/finance/cashflow?days=7', { headers: h })
          if (res.ok) setData(await res.json())
          else setData({ net_cash_flow: 4230, total_inflows: 8100, total_outflows: 3870 })
          break
        }
      }
    } catch {
      // Usar dados demo
    }
  }

  const fmt = (v: number) => v?.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' }) || 'R$ 0'

  switch (type) {
    case 'economy':
      return (
        <div style={{ background: 'linear-gradient(135deg, #0A3D8F, #1565C0)', borderRadius: 16, padding: '22px 20px', animation: 'pulseGreen 3s infinite' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: 'rgba(255,255,255,0.65)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Você economizou</span>
            <span style={{ background: 'rgba(74,222,128,0.2)', color: '#4ADE80', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>este mês</span>
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: 'white', margin: '0 0 4px', letterSpacing: -1 }}>{fmt(data?.total_credit || 847)}</p>
          <p style={{ fontSize: 12, color: 'rgba(255,255,255,0.55)', margin: '0 0 14px' }}>em crédito IBS/CBS abatido</p>
          <div style={{ background: 'rgba(255,255,255,0.1)', borderRadius: 8, padding: '8px 12px', fontSize: 11, color: 'rgba(255,255,255,0.8)' }}>
            💡 Sem o sistema, você pagaria esse valor extra
          </div>
        </div>
      )

    case 'revenue':
      return (
        <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Faturamento do mês</span>
            <span style={{ color: '#16A34A', fontSize: 12, fontWeight: 700 }}>↑ 12%</span>
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{fmt(data?.gross_revenue || 8100)}</p>
          <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 16px' }}>Meta: R$ 10.000 · 81% atingido</p>
          <div style={{ height: 6, background: '#F0F4FF', borderRadius: 3 }}>
            <div style={{ height: '100%', width: '81%', background: '#0A3D8F', borderRadius: 3 }} />
          </div>
        </div>
      )

    case 'os':
      return (
        <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Ordens de Serviço</span>
            <span style={{ background: '#FEF9C3', color: '#854D0E', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>hoje</span>
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{data?.total || 5}</p>
          <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 14px' }}>{data?.open || 2} abertas · {data?.pending || 1} aguardando</p>
          <div style={{ display: 'flex', gap: 6 }}>
            <div style={{ flex: 1, background: '#F0FDF4', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#16A34A' }}>{data?.open || 2} abertas</div>
            <div style={{ flex: 1, background: '#FEF9C3', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#854D0E' }}>{data?.pending || 1} aguard.</div>
            <div style={{ flex: 1, background: '#F0F4FF', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 11, fontWeight: 600, color: '#0A3D8F' }}>{data?.done || 2} concl.</div>
          </div>
        </div>
      )

    case 'stock':
      return (
        <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: data?.low_stock_count > 0 ? '1px solid #FDE68A' : '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Estoque</span>
            {(data?.low_stock_count || 0) > 0 && <span style={{ background: '#FEF9C3', color: '#854D0E', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>⚠️ {data.low_stock_count} baixo</span>}
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{fmt(data?.total_value || 12450)}</p>
          <p style={{ fontSize: 12, color: '#94a3b8', margin: 0 }}>valor total em estoque</p>
        </div>
      )

    case 'payables':
      return (
        <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: (data?.overdue_count || 0) > 0 ? '1px solid #FECDD3' : '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Contas a Pagar</span>
            {(data?.overdue_count || 0) > 0 && <span style={{ background: '#FFF1F2', color: '#BE123C', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>🔴 {data.overdue_count} vencida</span>}
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{fmt(data?.total_pending || 3850)}</p>
          <p style={{ fontSize: 12, color: '#94a3b8', margin: 0 }}>{data?.due_today || 2} vencem hoje</p>
        </div>
      )

    case 'cashflow':
      return (
        <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>Fluxo de Caixa</span>
            <span style={{ color: '#16A34A', fontSize: 10, fontWeight: 700, padding: '3px 8px' }}>7 dias</span>
          </div>
          <p style={{ fontSize: 32, fontWeight: 800, color: (data?.net_cash_flow || 4230) >= 0 ? '#16A34A' : '#DC2626', margin: '0 0 4px', letterSpacing: -1 }}>{fmt(data?.net_cash_flow || 4230)}</p>
          <p style={{ fontSize: 12, color: '#94a3b8', margin: 0 }}>saldo líquido projetado</p>
        </div>
      )

    default: return null
  }
}

// ── GRÁFICO SEMANAL ───────────────────────────────────────────────────────────

function WeeklyChart({ token }: { token: string }) {
  const days = ['Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb', 'Dom']
  const revenue = [3200, 4100, 2800, 5200, 4800, 3100, 1200]
  const economy = [480, 615, 420, 780, 720, 465, 180]
  const max = Math.max(...revenue)

  return (
    <div style={{ background: 'white', borderRadius: 16, padding: 20, border: '0.5px solid #e8e8e8' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20 }}>
        <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>Faturamento semanal</p>
        <div style={{ display: 'flex', gap: 14 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
            <div style={{ width: 10, height: 10, background: '#0A3D8F', borderRadius: 2 }} />
            <span style={{ fontSize: 11, color: '#94a3b8' }}>Receita</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
            <div style={{ width: 10, height: 10, background: '#4ADE80', borderRadius: 2 }} />
            <span style={{ fontSize: 11, color: '#94a3b8' }}>Imposto economizado</span>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 10, height: 120 }}>
        {days.map((day, i) => (
          <div key={day} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, height: '100%', justifyContent: 'flex-end' }}>
            <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: 2, justifyContent: 'flex-end' }}>
              <div style={{ width: '100%', height: Math.round(revenue[i] / max * 90), background: '#0A3D8F', borderRadius: '3px 3px 0 0', opacity: 0.85 }} />
              <div style={{ width: '100%', height: Math.round(economy[i] / max * 90), background: '#4ADE80', borderRadius: 0, opacity: 0.7 }} />
            </div>
            <span style={{ fontSize: 10, color: '#94a3b8' }}>{day}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
