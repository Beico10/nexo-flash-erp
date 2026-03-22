'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

const NICHO_CONFIG: Record<string, any> = {
  mechanic:   { label: 'Mecanica',  emoji: '🔧', suffix: 'sua oficina esta movimentada hoje.', alertIcon: '🌙', alertTitle: '1 servico chegou essa noite', alertSub: 'Honda Civic - troca de oleo · 02:14', alertBtn: 'Abrir OS', alertRoute: '/mechanic', c1Label: 'Ordens de Servico', c1Val: '5', c1Sub: '2 abertas · 1 aguard. aprovacao', c1a: '2 abertas', c1b: '1 aguard.', c1c: '2 concl.', c2Label: 'Faturamento do mes', c2Val: 'R$ 8.100', c2Sub: 'Meta: R$ 10.000 · 81% atingido', c3Label: 'Voce economizou', c3Val: 'R$ 847', c3Sub: 'em credito IBS/CBS abatido', weekLabel: 'Faturamento por dia', weekCol2: 'Imposto economizado' },
  bakery:     { label: 'Padaria',   emoji: '🍞', suffix: 'a padaria esta a todo vapor hoje.', alertIcon: '🌅', alertTitle: '2 pedidos chegaram essa manha', alertSub: 'Pao frances (50un) + Bolo de chocolate · 06:30', alertBtn: 'Ver pedidos', alertRoute: '/bakery', c1Label: 'Producao do dia', c1Val: '18', c1Sub: '6 itens em forno · 3 prontos', c1a: '6 em forno', c1b: '3 prontos', c1c: '9 a produzir', c2Label: 'Caixa do dia', c2Val: 'R$ 1.240', c2Sub: 'Meta: R$ 1.500 · 83% atingido', c3Label: 'Ingredientes em falta', c3Val: '3', c3Sub: 'Farinha, Fermento, Manteiga', weekLabel: 'Producao por dia', weekCol2: 'Receita do dia' },
  aesthetics: { label: 'Estetica',  emoji: '💇', suffix: 'a agenda de hoje esta cheia.', alertIcon: '📅', alertTitle: '3 agendamentos para hoje', alertSub: 'Primeiro as 09:00 - Corte + Escova · Ana Beatriz', alertBtn: 'Ver agenda', alertRoute: '/aesthetics', c1Label: 'Agendamentos hoje', c1Val: '8', c1Sub: '3 confirmados · 2 a confirmar', c1a: '3 confirm.', c1b: '2 aguard.', c1c: '3 concl.', c2Label: 'Receita prevista', c2Val: 'R$ 2.340', c2Sub: '4 profissionais ativas hoje', c3Label: 'Clientes sem retorno', c3Val: '5', c3Sub: 'Ultimo atendimento ha +30 dias', weekLabel: 'Agendamentos por dia', weekCol2: 'Receita do dia' },
  logistics:  { label: 'Logistica', emoji: '🚛', suffix: 'as rotas de hoje ja estao definidas.', alertIcon: '⚠️', alertTitle: '1 entrega em atraso', alertSub: 'Rota SP-03 - Saida prevista 07:00 · ainda no patio', alertBtn: 'Ver rota', alertRoute: '/logistics', c1Label: 'Entregas hoje', c1Val: '12', c1Sub: '8 em rota · 3 entregues · 1 atrasada', c1a: '8 em rota', c1b: '1 atrasada', c1c: '3 entregues', c2Label: 'Faturamento fretes', c2Val: 'R$ 5.400', c2Sub: 'Meta: R$ 7.000 · 77% atingido', c3Label: 'CT-e pendentes', c3Val: '4', c3Sub: 'Aguardando emissao ou assinatura', weekLabel: 'Entregas por dia', weekCol2: 'Receita de fretes' },
  industry:   { label: 'Industria', emoji: '🏭', suffix: 'a producao esta em andamento.', alertIcon: '📦', alertTitle: '2 pedidos para entregar hoje', alertSub: 'Pedido #1042 e #1043 · saida as 14:00', alertBtn: 'Ver pedidos', alertRoute: '/industry', c1Label: 'Pedidos em aberto', c1Val: '7', c1Sub: '3 em producao · 2 aguard. material', c1a: '3 producao', c1b: '2 aguard.', c1c: '2 prontos', c2Label: 'Faturamento do mes', c2Val: 'R$ 42.000', c2Sub: 'Meta: R$ 50.000 · 84% atingido', c3Label: 'Materia-prima critica', c3Val: '2', c3Sub: 'Aco carbono · Resina EP-40', weekLabel: 'Producao por dia', weekCol2: 'Faturamento' },
  shoes:      { label: 'Calcados',  emoji: '👟', suffix: 'as vendas de hoje estao aquecidas.', alertIcon: '📉', alertTitle: '3 modelos em estoque critico', alertSub: 'Tenis Runner 42 · Sandalia 37 · Bota 39', alertBtn: 'Ver estoque', alertRoute: '/shoes', c1Label: 'Vendas hoje', c1Val: '23', c1Sub: 'R$ 3.450 em produtos vendidos', c1a: '14 a vista', c1b: '6 cartao', c1c: '3 pix', c2Label: 'Faturamento do mes', c2Val: 'R$ 28.700', c2Sub: 'Meta: R$ 35.000 · 82% atingido', c3Label: 'Itens em falta na grade', c3Val: '8', c3Sub: 'Numeros 37, 38 e 42 criticos', weekLabel: 'Vendas por dia', weekCol2: 'Receita' },
}

const WEEK_DATA: Record<string, any[]> = {
  mechanic:   [ { day: 'Segunda', v: 3200, v2: 480,  best: false, today: false }, { day: 'Terca',   v: 4100, v2: 615,  best: true,  today: false }, { day: 'Quarta',  v: 2800, v2: 420,  best: false, today: false }, { day: 'Quinta',  v: 5200, v2: 780,  best: false, today: false }, { day: 'Sexta',   v: 4800, v2: 720,  best: false, today: false }, { day: 'Sabado',  v: 3100, v2: 465,  best: false, today: true  }, { day: 'Domingo', v: 0,    v2: 0,    best: false, today: false } ],
  bakery:     [ { day: 'Segunda', v: 980,  v2: 1100, best: false, today: false }, { day: 'Terca',   v: 1050, v2: 1200, best: false, today: false }, { day: 'Quarta',  v: 1300, v2: 1450, best: true,  today: false }, { day: 'Quinta',  v: 870,  v2: 980,  best: false, today: false }, { day: 'Sexta',   v: 1100, v2: 1250, best: false, today: false }, { day: 'Sabado',  v: 1240, v2: 1380, best: false, today: true  }, { day: 'Domingo', v: 600,  v2: 680,  best: false, today: false } ],
  aesthetics: [ { day: 'Segunda', v: 6,    v2: 1800, best: false, today: false }, { day: 'Terca',   v: 8,    v2: 2100, best: false, today: false }, { day: 'Quarta',  v: 10,   v2: 2800, best: true,  today: false }, { day: 'Quinta',  v: 7,    v2: 1950, best: false, today: false }, { day: 'Sexta',   v: 9,    v2: 2400, best: false, today: false }, { day: 'Sabado',  v: 8,    v2: 2340, best: false, today: true  }, { day: 'Domingo', v: 0,    v2: 0,    best: false, today: false } ],
  logistics:  [ { day: 'Segunda', v: 14,   v2: 5200, best: false, today: false }, { day: 'Terca',   v: 18,   v2: 6800, best: true,  today: false }, { day: 'Quarta',  v: 11,   v2: 4100, best: false, today: false }, { day: 'Quinta',  v: 16,   v2: 6000, best: false, today: false }, { day: 'Sexta',   v: 20,   v2: 7500, best: false, today: false }, { day: 'Sabado',  v: 12,   v2: 5400, best: false, today: true  }, { day: 'Domingo', v: 0,    v2: 0,    best: false, today: false } ],
  industry:   [ { day: 'Segunda', v: 3,    v2: 8400, best: false, today: false }, { day: 'Terca',   v: 5,    v2: 14000,best: true,  today: false }, { day: 'Quarta',  v: 2,    v2: 5600, best: false, today: false }, { day: 'Quinta',  v: 4,    v2: 11200,best: false, today: false }, { day: 'Sexta',   v: 6,    v2: 16800,best: false, today: false }, { day: 'Sabado',  v: 1,    v2: 2800, best: false, today: true  }, { day: 'Domingo', v: 0,    v2: 0,    best: false, today: false } ],
  shoes:      [ { day: 'Segunda', v: 18,   v2: 2700, best: false, today: false }, { day: 'Terca',   v: 24,   v2: 3600, best: false, today: false }, { day: 'Quarta',  v: 31,   v2: 4650, best: true,  today: false }, { day: 'Quinta',  v: 20,   v2: 3000, best: false, today: false }, { day: 'Sexta',   v: 28,   v2: 4200, best: false, today: false }, { day: 'Sabado',  v: 23,   v2: 3450, best: false, today: true  }, { day: 'Domingo', v: 8,    v2: 1200, best: false, today: false } ],
}

const NICHO_CONCERNS: Record<string, string[]> = {
  mechanic:   ['OS atrasada ou aguardando aprovacao', 'Peca em falta no estoque', 'Quanto faturei hoje', 'Conta vencendo'],
  bakery:     ['Produto vencendo', 'Caixa do dia', 'Producao do dia', 'Ingrediente em falta'],
  industry:   ['Pedido atrasado', 'Materia-prima em falta', 'Faturamento do mes', 'Custo de producao'],
  logistics:  ['Entrega atrasada', 'Rota do dia', 'CT-e pendente', 'Faturamento de fretes'],
  aesthetics: ['Agendamento do dia', 'Produto vencendo', 'Cliente sem retorno', 'Conta vencendo'],
  shoes:      ['Produto em falta na grade', 'Venda do dia', 'Estoque baixo', 'Conta vencendo'],
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

function CoPiloto({ bt, onDone }: { bt: string; onDone: () => void }) {
  const cfg = NICHO_CONFIG[bt] || NICHO_CONFIG.mechanic
  const concerns = NICHO_CONCERNS[bt] || NICHO_CONCERNS.mechanic
  const [msg, setMsg] = useState('')
  const [typing, setTyping] = useState(true)
  const [sel, setSel] = useState<string[]>([])
  const [done, setDone] = useState(false)
  const full = 'Ola! Sou seu Co-Piloto. Vi que voce e de ' + cfg.label + ' ' + cfg.emoji + '. O que voce mais precisa ver todo dia quando abre o sistema?'

  useEffect(() => {
    let i = 0
    const t = setInterval(() => { setMsg(full.slice(0, i + 1)); i++; if (i >= full.length) { clearInterval(t); setTyping(false) } }, 16)
    return () => clearInterval(t)
  }, [])

  const toggle = (c: string) => setSel(p => p.includes(c) ? p.filter(x => x !== c) : [...p, c])

  const confirm = () => {
    setDone(true)
    localStorage.setItem('nexo_dash_configured', 'true')
    localStorage.setItem('nexo_dash_concerns', JSON.stringify(sel))
    setTimeout(onDone, 900)
  }

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', padding: 20 }}>
      <div style={{ background: 'white', borderRadius: 20, padding: 26, width: '100%', maxWidth: 500, marginBottom: 12, boxShadow: '0 20px 60px rgba(0,0,0,0.2)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
          <div style={{ width: 38, height: 38, background: '#0A3D8F', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18 }}>🤖</div>
          <div>
            <p style={{ fontSize: 13, fontWeight: 700, color: '#1C1917', margin: 0 }}>Co-Piloto IA</p>
            <p style={{ fontSize: 11, color: '#94a3b8', margin: 0 }}>Montando seu painel para {cfg.label}</p>
          </div>
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 5 }}>
            {[0,1,2].map(i => <div key={i} style={{ width: 7, height: 7, borderRadius: '50%', background: i === 0 ? '#0A3D8F' : '#E0E4F0' }} />)}
          </div>
        </div>
        <div style={{ background: '#F0F4FF', borderRadius: 12, padding: '12px 16px', marginBottom: 18, minHeight: 52 }}>
          <p style={{ fontSize: 13, color: '#1C1917', lineHeight: 1.6, margin: 0 }}>{msg}{typing && <span style={{ opacity: 0.5 }}>|</span>}</p>
        </div>
        {!typing && !done && (
          <>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14 }}>
              {concerns.map(c => (
                <button key={c} onClick={() => toggle(c)} style={{ display: 'flex', alignItems: 'center', gap: 10, background: sel.includes(c) ? '#E3F2FD' : '#F8F7F4', border: sel.includes(c) ? '1.5px solid #0A3D8F' : '1px solid #e0e4f0', borderRadius: 10, padding: '10px 14px', cursor: 'pointer', textAlign: 'left' }}>
                  <div style={{ width: 17, height: 17, borderRadius: 4, border: '1.5px solid', borderColor: sel.includes(c) ? '#0A3D8F' : '#cbd5e1', background: sel.includes(c) ? '#0A3D8F' : 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    {sel.includes(c) && <svg width="9" height="9" viewBox="0 0 9 9"><path d="M1 4.5l2.5 2.5 4.5-4.5" stroke="white" strokeWidth="1.5" fill="none" strokeLinecap="round"/></svg>}
                  </div>
                  <span style={{ fontSize: 13, color: '#1C1917', fontWeight: sel.includes(c) ? 600 : 400 }}>{c}</span>
                </button>
              ))}
            </div>
            <button onClick={confirm} style={{ width: '100%', padding: 12, borderRadius: 10, border: 'none', background: sel.length > 0 ? '#0A3D8F' : '#e0e4f0', color: sel.length > 0 ? 'white' : '#94a3b8', fontSize: 14, fontWeight: 700, cursor: 'pointer', marginBottom: 8 }}>
              {sel.length > 0 ? 'Montar meu painel de ' + cfg.label + ' →' : 'Selecione o que importa para voce'}
            </button>
            <button onClick={() => { localStorage.setItem('nexo_dash_configured', 'true'); onDone() }} style={{ width: '100%', background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer' }}>
              Pular por agora
            </button>
          </>
        )}
        {done && (
          <div style={{ textAlign: 'center', padding: 8 }}>
            <p style={{ fontSize: 28, margin: '0 0 8px' }}>🎉</p>
            <p style={{ fontSize: 14, color: '#16A34A', fontWeight: 700, margin: 0 }}>Painel montado para {cfg.label}!</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const router = useRouter()
  const [showCP, setShowCP] = useState(false)
  const [mounted, setMounted] = useState(false)
  const [bt, setBt] = useState('mechanic')

  useEffect(() => {
    setMounted(true)
    const b = localStorage.getItem('nexo_business_type') || 'mechanic'
    setBt(b)
    if (!localStorage.getItem('nexo_dash_configured')) setTimeout(() => setShowCP(true), 800)
  }, [])

  const greeting = () => { const h = new Date().getHours(); return h < 12 ? 'Bom dia' : h < 18 ? 'Boa tarde' : 'Boa noite' }
  const today = new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })

  if (!mounted) return null

  const cfg = NICHO_CONFIG[bt] || NICHO_CONFIG.mechanic
  const week = WEEK_DATA[bt] || WEEK_DATA.mechanic
  const isCurrency = ['mechanic', 'bakery', 'logistics', 'industry', 'shoes'].includes(bt)

  const fmtV = (v: number) => {
    if (bt === 'aesthetics') return v > 0 ? v + ' agend.' : '-'
    if (bt === 'logistics') return v > 0 ? v + ' entreg.' : '-'
    if (bt === 'industry') return v > 0 ? v + ' pedidos' : '-'
    if (bt === 'bakery') return v > 0 ? v + ' itens' : '-'
    if (bt === 'shoes') return v > 0 ? v + ' vendas' : '-'
    return v > 0 ? fmt(v) : '-'
  }

  return (
    <div style={{ background: '#F8F7F4', minHeight: '100vh', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' }}>
      {showCP && <CoPiloto bt={bt} onDone={() => setShowCP(false)} />}
      <div style={{ padding: 24, maxWidth: 1200, margin: '0 auto' }}>

        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
          <div>
            <p style={{ fontSize: 22, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>
              {greeting()}, bem-vindo! 👋
            </p>
            <p style={{ fontSize: 13, color: '#94a3b8', margin: '0 0 6px', textTransform: 'capitalize' }}>{today}</p>
            <span style={{ fontSize: 11, background: '#F0F4FF', color: '#0A3D8F', padding: '3px 12px', borderRadius: 100, fontWeight: 600, display: 'inline-block' }}>
              {cfg.emoji} {cfg.label} — {cfg.suffix}
            </span>
          </div>
          <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: 12, padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ fontSize: 18 }}>{cfg.alertIcon}</span>
            <div>
              <p style={{ fontSize: 12, fontWeight: 700, color: '#854D0E', margin: '0 0 2px' }}>{cfg.alertTitle}</p>
              <p style={{ fontSize: 11, color: '#A16207', margin: 0 }}>{cfg.alertSub}</p>
            </div>
            <button onClick={() => router.push(cfg.alertRoute)} style={{ background: '#0A3D8F', color: 'white', border: 'none', borderRadius: 8, padding: '6px 12px', fontSize: 11, fontWeight: 600, cursor: 'pointer', marginLeft: 8, whiteSpace: 'nowrap' }}>
              {cfg.alertBtn}
            </button>
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginBottom: 20 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{cfg.c1Label}</span>
              <span style={{ background: '#FEF9C3', color: '#854D0E', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>hoje</span>
            </div>
            <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{cfg.c1Val}</p>
            <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 14px' }}>{cfg.c1Sub}</p>
            <div style={{ display: 'flex', gap: 6 }}>
              {[cfg.c1a, cfg.c1b, cfg.c1c].map((d: string, i: number) => (
                <div key={i} style={{ flex: 1, background: '#F0F4FF', borderRadius: 6, padding: 6, textAlign: 'center', fontSize: 10, fontWeight: 600, color: '#0A3D8F' }}>{d}</div>
              ))}
            </div>
          </div>

          <div style={{ background: 'white', borderRadius: 16, padding: '22px 20px', border: '0.5px solid #e8e8e8' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{cfg.c2Label}</span>
              <span style={{ color: '#16A34A', fontSize: 12, fontWeight: 700 }}>↑ 12%</span>
            </div>
            <p style={{ fontSize: 32, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -1 }}>{cfg.c2Val}</p>
            <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 16px' }}>{cfg.c2Sub}</p>
            <div style={{ height: 6, background: '#F0F4FF', borderRadius: 3 }}>
              <div style={{ height: '100%', width: '82%', background: '#0A3D8F', borderRadius: 3 }} />
            </div>
          </div>

          <div style={{ background: bt === 'mechanic' ? 'linear-gradient(135deg,#0A3D8F,#1565C0)' : 'white', borderRadius: 16, padding: '22px 20px', border: bt === 'mechanic' ? 'none' : '1px solid #FEE2E2' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 14 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: bt === 'mechanic' ? 'rgba(255,255,255,0.65)' : '#94a3b8', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{cfg.c3Label}</span>
              {bt === 'mechanic' && <span style={{ background: 'rgba(74,222,128,0.2)', color: '#4ADE80', fontSize: 10, fontWeight: 700, padding: '3px 8px', borderRadius: 100 }}>este mes</span>}
            </div>
            <p style={{ fontSize: 32, fontWeight: 800, color: bt === 'mechanic' ? 'white' : '#DC2626', margin: '0 0 4px', letterSpacing: -1 }}>{cfg.c3Val}</p>
            <p style={{ fontSize: 12, color: bt === 'mechanic' ? 'rgba(255,255,255,0.55)' : '#94a3b8', margin: '0 0 14px' }}>{cfg.c3Sub}</p>
            {bt === 'mechanic' && (
              <div style={{ background: 'rgba(255,255,255,0.1)', borderRadius: 8, padding: '8px 12px', fontSize: 11, color: 'rgba(255,255,255,0.8)' }}>
                💡 Sem o sistema, voce pagaria esse valor extra
              </div>
            )}
          </div>
        </div>

        <div style={{ background: 'white', borderRadius: 16, padding: 20, border: '0.5px solid #e8e8e8', marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
            <p style={{ fontSize: 14, fontWeight: 700, color: '#1C1917', margin: 0 }}>{cfg.weekLabel}</p>
            <span style={{ fontSize: 11, color: '#94a3b8' }}>esta semana</span>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                {['Dia', cfg.weekLabel.split(' ')[0], cfg.weekCol2, ''].map((h: string, i: number) => (
                  <th key={i} style={{ fontSize: 10, color: '#94a3b8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', padding: '0 0 10px', textAlign: i === 0 || i === 3 ? 'left' : 'right' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {week.map((row: any, i: number) => (
                <tr key={i} style={{ background: row.best ? '#F0FDF4' : row.today ? '#F0F4FF' : 'transparent' }}>
                  <td style={{ padding: '10px 0', fontSize: 13, fontWeight: row.today ? 700 : 400, color: '#1C1917' }}>
                    {row.day}{row.today && <span style={{ fontSize: 10, color: '#0A3D8F', marginLeft: 6, fontWeight: 600 }}>hoje</span>}
                  </td>
                  <td style={{ padding: '10px 8px', fontSize: 13, fontWeight: 700, color: '#1C1917', textAlign: 'right' }}>{fmtV(row.v)}</td>
                  <td style={{ padding: '10px 8px', fontSize: 13, color: '#16A34A', fontWeight: 600, textAlign: 'right' }}>{row.v2 > 0 ? fmt(row.v2) : '-'}</td>
                  <td style={{ padding: '10px 0', textAlign: 'right' }}>
                    {row.best && <span style={{ fontSize: 10, background: '#DCFCE7', color: '#16A34A', padding: '2px 8px', borderRadius: 100, fontWeight: 700 }}>🏆 melhor dia</span>}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr style={{ borderTop: '0.5px solid #e8e8e8' }}>
                <td style={{ padding: '12px 0', fontSize: 13, fontWeight: 700, color: '#1C1917' }}>Total</td>
                <td style={{ padding: '12px 8px', fontSize: 14, fontWeight: 800, color: '#0A3D8F', textAlign: 'right' }}>{fmtV(week.reduce((s: number, r: any) => s + r.v, 0))}</td>
                <td style={{ padding: '12px 8px', fontSize: 13, fontWeight: 700, color: '#16A34A', textAlign: 'right' }}>{fmt(week.reduce((s: number, r: any) => s + r.v2, 0))}</td>
                <td />
              </tr>
            </tfoot>
          </table>
        </div>

        <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 12, padding: '12px 18px', display: 'flex', alignItems: 'center', gap: 10 }}>
          <span style={{ fontSize: 16 }}>⚡</span>
          <div>
            <p style={{ fontSize: 12, fontWeight: 700, color: '#0A3D8F', margin: 0 }}>Reforma Tributaria 2026 ativa</p>
            <p style={{ fontSize: 11, color: '#64748b', margin: 0 }}>IBS e CBS calculados automaticamente em todas as operacoes de {cfg.label}.</p>
          </div>
        </div>

      </div>
    </div>
  )
}