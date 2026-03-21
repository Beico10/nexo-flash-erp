'use client'
import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'

const NICHO_LABELS: Record<string, string> = {
  mechanic: 'Mecânica', bakery: 'Padaria', industry: 'Indústria',
  logistics: 'Logística', aesthetics: 'Estética', shoes: 'Calçados',
}

const NICHO_CONCERNS: Record<string, string[]> = {
  mechanic:   ['OS atrasada ou aguardando aprovação', 'Peça em falta no estoque', 'Quanto faturei hoje', 'Conta vencendo'],
  bakery:     ['Produto vencendo', 'Caixa do dia', 'Produção do dia', 'Ingrediente em falta'],
  industry:   ['Pedido atrasado', 'Matéria-prima em falta', 'Faturamento do mês', 'Custo de produção'],
  logistics:  ['Entrega atrasada', 'Rota do dia', 'CT-e pendente', 'Faturamento de fretes'],
  aesthetics: ['Agendamento do dia', 'Produto vencendo', 'Cliente sem retorno', 'Conta vencendo'],
  shoes:      ['Produto em falta na grade', 'Venda do dia', 'Estoque baixo', 'Conta vencendo'],
}

// Dados demo por dia da semana
const WEEK_DATA = [
  { day: 'Segunda', value: 3200, economy: 480, isBest: false, isToday: false },
  { day: 'Terça',   value: 4100, economy: 615, isBest: true,  isToday: false },
  { day: 'Quarta',  value: 2800, economy: 420, isBest: false, isToday: false },
  { day: 'Quinta',  value: 5200, economy: 780, isBest: false, isToday: false },
  { day: 'Sexta',   value: 4800, economy: 720, isBest: false, isToday: false },
  { day: 'Sábado',  value: 3100, economy: 465, isBest: false, isToday: true  },
  { day: 'Domingo', value: 0,    economy: 0,   isBest: false, isToday: false },
]

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

// Co-Piloto de boas-vindas — aparece UMA VEZ após escolher nicho
function CoPilotoOnboarding({ businessType, onComplete }: { businessType: string, onComplete: () => void }) {
  const [msg, setMsg] = useState('')
  const [typing, setTyping] = useState(true)
  const [selected, setSelected] = useState<string[]>([])
  const [done, setDone] = useState(false)
  const nichoPretty = NICHO_LABELS[businessType] || 'seu negócio'
  const concerns = NICHO_CONCERNS[businessType] || NICHO_CONCERNS.mechanic

  const fullMsg = `Olá! Sou seu Co-Piloto. Vi que você é de ${nichoPretty}. O que você mais precisa ver todo dia quando abre o sistema?`

  useEffect(() => {
    let i = 0
    const interval = setInterval(() => {
      setMsg(fullMsg.slice(0, i + 1))
      i++
      if (i >= fullMsg.length) { clearInterval(interval); setTyping(false) }
    }, 16)
    return () => clearInterval(interval)
  }, [])

  const toggle = (c: string) => setSelected(prev => prev.includes(c) ? prev.filter(x => x !== c) : [...prev, c])

  const handleConfirm = () => {
    setDone(true)
    localStorage.setItem('nexo_dash_configured', 'true')
    localStorage.setItem('nexo_dash_concerns', JSON.stringify(selected))
    // Fechar rapidamente para mostrar o dashboard personalizado
    setTimeout(onComplete, 800)
  }

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', padding: 20 }}>
      <div style={{ background: 'white', borderRadius: 20, padding: 26, width: '100%', maxWidth: 500, marginBottom: 12 }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
          <div style={{ width: 38, height: 38, background: '#0A3D8F', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, flexShrink: 0 }}>🤖</div>
          <div>
            <p style={{ fontSize: 13, fontWeight: 700, color: '#1C1917', margin: 0 }}>Co-Piloto IA</p>
            <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>Montando seu painel personalizado</p>
          </div>
          {/* Indicador de progresso */}
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 5 }}>
            {[0,1,2].map(i => <div key={i} style={{ width: 7, height: 7, borderRadius: '50%', background: i === 0 ? '#0A3D8F' : '#E0E4F0' }} />)}
          </div>
        </div>

        {/* Mensagem digitando */}
        <div style={{ background: '#F0F4FF', borderRadius: 12, padding: '12px 16px', marginBottom: 18, minHeight: 52 }}>
          <p style={{ fontSize: 13, color: '#1C1917', lineHeight: 1.6, margin: 0 }}>
            {msg}{typing && <span style={{ opacity: 0.5 }}>|</span>}
          </p>
        </div>

        {/* Opções do nicho */}
        {!typing && !done && (
          <>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14 }}>
              {concerns.map(c => (
                <button key={c} onClick={() => toggle(c)} style={{
                  display: 'flex', alignItems: 'center', gap: 10,
                  background: selected.includes(c) ? '#E3F2FD' : '#F8F7F4',
                  border: selected.includes(c) ? '1.5px solid #0A3D8F' : '1px solid #e0e4f0',
                  borderRadius: 10, padding: '10px 14px', cursor: 'pointer', textAlign: 'left',
                }}>
                  <div style={{ width: 17, height: 17, borderRadius: 4, border: '1.5px solid', borderColor: selected.includes(c) ? '#0A3D8F' : '#cbd5e1', background: selected.includes(c) ? '#0A3D8F' : 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    {selected.includes(c) && <svg width="9" height="9" viewBox="0 0 9 9"><path d="M1 4.5l2.5 2.5 4.5-4.5" stroke="white" strokeWidth="1.5" fill="none" strokeLinecap="round"/></svg>}
                  </div>
                  <span style={{ fontSize: 13, color: '#1C1917', fontWeight: selected.includes(c) ? 600 : 400 }}>{c}</span>
                </button>
              ))}
            </div>
            <button onClick={handleConfirm} style={{ width: '100%', padding: 12, borderRadius: 10, border: 'none', background: selected.length > 0 ? '#0A3D8F' : '#e0e4f0', color: selected.length > 0 ? 'white' : '#94a3b8', fontSize: 14, fontWeight: 700, cursor: 'pointer', marginBottom: 8 }}>
              {selected.length > 0 ? 'Montar meu painel →' : 'Selecione o que importa para você'}
            </button>
            <button onClick={() => { localStorage.setItem('nexo_dash_configured', 'true'); onComplete() }} style={{ width: '100%', background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer' }}>
              Pular por agora
            </button>
          </>
        )}

        {/* Confirmação */}
        {done && (
          <div style={{ textAlign: 'center', padding: 8 }}>
            <p style={{ fontSize: 28, margin: '0 0 8px' }}>🎉</p>
            <p style={{ fontSize: 14, color: '#16A34A', fontWeight: 700, margin: 0 }}>Painel montado para {nichoPretty}!</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const router = useRouter()
  const [showCoPiloto, setShowCoPiloto] = useState(false)
  const [userName] = useState('João')
  const [mounted, setMounted] = useState(false)
  const [businessType, setBusinessType] = useState('mechanic')

  useEffect(() => {
    setMounted(true)
    const bt = localStorage.getItem('nexo_business_type') || 'mechanic'
    setBusinessType(bt)

    // Mostrar Co-Piloto APENAS se nicho foi escolhido mas painel ainda não configurado
    const configured = localStorage.getItem('nexo_dash_configured')
    const hasNicho = localStorage.getItem('nexo_business_type')
    if (!configured && hasNicho) {
      setTimeout(() => setShowCoPiloto(true), 800)
    }
  }, [])

  const greeting = () => {
    const h = new Date().getHours()
    if (h < 12) return 'Bom dia'
    if (h < 18) return 'Boa tarde'
    return 'Boa noite'
  }

  const today = new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })

  if (!mounted) return null

  return (
    <div style={{ background: '#F8F7F4', minHeight: '100vh', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>

      {/* Co-Piloto onboarding */}
      {showCoPiloto && (
        <CoPilotoOnboarding
          businessType={businessType}
          onComplete={() => setShowCoPiloto(false)}
        />
      )}

      <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>

        {/* Saudação + alerta OS noturna */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
          <div>
            <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>
              {greeting()}, {userName}! 👋
            </p>
            <p style={{ fontSize: 13, color: '#94a3b8', margin: 0, textTransform: 'capitalize' }}>{today}</p>
            <span style={{ fontSize: 11, background: '#F0F4FF', color: '#0A3D8F', padding: '2px 10px', borderRadius: 100, fontWeight: 600, marginTop: 4, display: 'inline-block' }}>
              {NICHO_LABELS[businessType]}
            </span>
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

        {/* 3 cards principais */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>

          {/* Card economia fiscal */}
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

          {/* Card OS — label muda por nicho */}
          <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                {businessType === 'logistics' ? 'Entregas hoje' : businessType === 'bakery' ? 'Pedidos hoje' : businessType === 'aesthetics' ? 'Agendamentos' : 'Ordens de Serviço'}
              </span>
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
        </div>

        {/* Tabela faturamento por dia */}
        <div style={{ background: 'white', borderRadius: 16, padding: 20, border: '0.5px solid #e8e8e8' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>Faturamento por dia</p>
            <span style={{ fontSize: 11, color: '#94a3b8' }}>esta semana</span>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                {['Dia', 'Faturamento', 'Imposto economizado', ''].map(h => (
                  <th key={h} style={{ fontSize: 10, color: '#94a3b8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', padding: '0 0 10px', textAlign: h === 'Dia' || h === '' ? 'left' : 'right' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {WEEK_DATA.map((row, i) => (
                <tr key={i} style={{ background: row.isBest ? '#F0FDF4' : row.isToday ? '#F0F4FF' : 'transparent' }}>
                  <td style={{ padding: '10px 0', fontSize: 13, fontWeight: row.isToday ? 700 : 400, color: '#1C1917' }}>
                    {row.day}
                    {row.isToday && <span style={{ fontSize: 10, color: '#0A3D8F', marginLeft: 6, fontWeight: 600 }}>hoje</span>}
                  </td>
                  <td style={{ padding: '10px 8px', fontSize: 13, fontWeight: 700, color: '#1C1917', textAlign: 'right' }}>
                    {row.value > 0 ? fmt(row.value) : '—'}
                  </td>
                  <td style={{ padding: '10px 8px', fontSize: 13, color: '#16A34A', fontWeight: 600, textAlign: 'right' }}>
                    {row.economy > 0 ? '+' + fmt(row.economy) : '—'}
                  </td>
                  <td style={{ padding: '10px 0', textAlign: 'right' }}>
                    {row.isBest && <span style={{ fontSize: 10, background: '#DCFCE7', color: '#16A34A', padding: '2px 8px', borderRadius: 100, fontWeight: 700 }}>🏆 melhor dia</span>}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr style={{ borderTop: '0.5px solid #e8e8e8' }}>
                <td style={{ padding: '12px 0', fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Total</td>
                <td style={{ padding: '12px 8px', fontSize: 14, fontWeight: 800, color: '#0A3D8F', textAlign: 'right' }}>
                  {fmt(WEEK_DATA.reduce((s, r) => s + r.value, 0))}
                </td>
                <td style={{ padding: '12px 8px', fontSize: 13, fontWeight: 700, color: '#16A34A', textAlign: 'right' }}>
                  +{fmt(WEEK_DATA.reduce((s, r) => s + r.economy, 0))}
                </td>
                <td />
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>
  )
}
