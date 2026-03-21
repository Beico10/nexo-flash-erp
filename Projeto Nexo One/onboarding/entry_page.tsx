'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'

const NICHOS = [
  { key: 'mechanic',   label: 'Mecânica',   icon: '🔧', desc: 'OS, peças, aprovação WhatsApp' },
  { key: 'bakery',     label: 'Padaria',    icon: '🍞', desc: 'PDV, produção, validade' },
  { key: 'industry',   label: 'Indústria',  icon: '🏭', desc: 'Pedidos, estoque, NF-e' },
  { key: 'logistics',  label: 'Logística',  icon: '🚛', desc: 'Rotas, CT-e, despacho' },
  { key: 'aesthetics', label: 'Estética',   icon: '💇', desc: 'Agenda, clientes, produtos' },
  { key: 'shoes',      label: 'Calçados',   icon: '👟', desc: 'Grade, estoque, vendas' },
]

export default function EntryPage() {
  const router = useRouter()
  const [selected, setSelected] = useState('')
  const [entering, setEntering] = useState(false)

  const handleSelect = (nicho: string) => {
    setSelected(nicho)
    setEntering(true)

    if (typeof window !== 'undefined') {
      localStorage.setItem('nexo_business_type', nicho)
      localStorage.setItem('nexo_token', 'demo-token')
      localStorage.setItem('nexo_tenant', 'demo')
      localStorage.setItem('nexo_demo_mode', 'true')
      localStorage.removeItem('nexo_dash_configured')
    }

    setTimeout(() => router.push('/dashboard'), 400)
  }

  return (
    <div style={{
      display: 'flex', minHeight: '100vh',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    }}>

      {/* LADO ESQUERDO — azul com benefícios */}
      <div style={{ width: '44%', background: '#0A3D8F', padding: '36px 30px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
        <div>
          {/* Logo */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 32 }}>
            <div style={{ width: 32, height: 32, background: 'rgba(255,255,255,0.15)', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <svg width="16" height="16" viewBox="0 0 64 64">
                <line x1="20" y1="20" x2="32" y2="28" stroke="white" strokeWidth="2.5"/>
                <line x1="32" y1="28" x2="44" y2="20" stroke="white" strokeWidth="2.5"/>
                <line x1="16" y1="34" x2="32" y2="28" stroke="white" strokeWidth="2.5"/>
                <line x1="32" y1="28" x2="48" y2="34" stroke="white" strokeWidth="2.5"/>
                <line x1="24" y1="36" x2="32" y2="44" stroke="white" strokeWidth="2.5"/>
                <line x1="40" y1="36" x2="32" y2="44" stroke="white" strokeWidth="2.5"/>
                <circle cx="32" cy="28" r="5" fill="white"/>
                <circle cx="32" cy="44" r="4" fill="white"/>
                <circle cx="20" cy="20" r="3" fill="white"/>
                <circle cx="44" cy="20" r="3" fill="white"/>
              </svg>
            </div>
            <div>
              <div style={{ fontSize: 14, fontWeight: 700, color: 'white' }}>Gestão Para Todos</div>
              <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)', marginTop: 1 }}>ERP inteligente</div>
            </div>
          </div>

          {/* Headline + Badge */}
          <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start', marginBottom: 24 }}>
            <div style={{ flex: 1 }}>
              <p style={{ fontSize: 26, fontWeight: 800, color: 'white', lineHeight: 1.2, margin: '0 0 12px', letterSpacing: -0.5 }}>
                Você trabalha.<br />A gente cuida<br />da gestão.
              </p>
              <p style={{ fontSize: 12, color: 'rgba(255,255,255,0.55)', lineHeight: 1.65, margin: 0 }}>
                Não pague imposto duas vezes.<br />
                Gerencie tudo em um só lugar.
              </p>
            </div>
            <div style={{ width: 100, flexShrink: 0, background: 'rgba(255,255,255,0.1)', border: '1px solid rgba(255,255,255,0.2)', borderRadius: 12, padding: '12px 8px', textAlign: 'center' }}>
              <div style={{ width: 8, height: 8, borderRadius: '50%', background: '#4ADE80', margin: '0 auto 8px' }} />
              <p style={{ fontSize: 9, fontWeight: 600, color: 'rgba(255,255,255,0.85)', margin: '0 0 6px', lineHeight: 1.3 }}>Pronto para a<br />Reforma</p>
              <p style={{ fontSize: 22, fontWeight: 800, color: '#4ADE80', margin: '0 0 4px', lineHeight: 1 }}>2026</p>
              <p style={{ fontSize: 8, color: 'rgba(255,255,255,0.4)', margin: 0, lineHeight: 1.4 }}>IBS+CBS automático</p>
            </div>
          </div>

          {/* Cards de benefício */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
            {[
              'Dados criptografados — segurança bancária',
              'IBS/CBS 2026 calculado automaticamente',
              'IA assistente com aprovação humana',
              'Empresa pequena, média ou grande — atendemos',
            ].map((item, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 9, background: 'rgba(255,255,255,0.07)', borderRadius: 8, padding: '8px 12px' }}>
                <svg width="11" height="11" viewBox="0 0 12 12" fill="none" style={{ flexShrink: 0 }}>
                  <path d="M1 6l3.5 3.5L11 2" stroke="rgba(255,255,255,0.7)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.8)' }}>{item}</span>
              </div>
            ))}
          </div>
        </div>

        <p style={{ fontSize: 10, color: 'rgba(255,255,255,0.2)', margin: '20px 0 0' }}>
          © 2026 Gestão Para Todos · Reforma Tributária Brasil
        </p>
      </div>

      {/* LADO DIREITO — escolha do nicho */}
      <div style={{ flex: 1, background: '#F0F4FF', padding: '36px 32px', display: 'flex', flexDirection: 'column', justifyContent: 'center', position: 'relative', overflow: 'hidden' }}>

        {/* Pontos decorativos */}
        <svg style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}>
          <defs>
            <pattern id="dots" x="0" y="0" width="24" height="24" patternUnits="userSpaceOnUse">
              <circle cx="2" cy="2" r="1.1" fill="#0A3D8F" opacity="0.06"/>
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#dots)"/>
        </svg>

        <div style={{ position: 'relative', maxWidth: 380, margin: '0 auto', width: '100%' }}>
          <p style={{ fontSize: 22, fontWeight: 800, color: '#0A3D8F', margin: '0 0 6px', letterSpacing: -0.5 }}>
            Qual é o seu negócio?
          </p>
          <p style={{ fontSize: 13, color: '#64748b', margin: '0 0 24px' }}>
            Explore o sistema do jeito que funciona para você — sem cadastro
          </p>

          {/* Grade de nichos */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 10, marginBottom: 20 }}>
            {NICHOS.map(nicho => (
              <button
                key={nicho.key}
                onClick={() => handleSelect(nicho.key)}
                disabled={entering}
                style={{
                  background: selected === nicho.key ? '#E3F2FD' : 'white',
                  border: selected === nicho.key ? '2px solid #0A3D8F' : '1px solid #e0e4f0',
                  borderRadius: 12, padding: '16px 8px',
                  cursor: 'pointer', textAlign: 'center',
                  transition: 'all 0.15s',
                  transform: selected === nicho.key ? 'scale(0.97)' : 'scale(1)',
                }}
              >
                <div style={{ fontSize: 28, marginBottom: 6 }}>{nicho.icon}</div>
                <p style={{ fontSize: 12, fontWeight: 700, color: selected === nicho.key ? '#0A3D8F' : '#1C1917', margin: '0 0 3px' }}>
                  {nicho.label}
                </p>
                <p style={{ fontSize: 10, color: '#94a3b8', margin: 0, lineHeight: 1.3 }}>
                  {nicho.desc}
                </p>
              </button>
            ))}
          </div>

          {/* Entrada sem cadastro */}
          <p style={{ fontSize: 11, color: '#94a3b8', textAlign: 'center', margin: '0 0 16px' }}>
            Clique no seu negócio para entrar direto · Sem email · Sem senha
          </p>

          {/* Divisor */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
            <div style={{ flex: 1, height: 0.5, background: '#cbd5e1' }} />
            <span style={{ fontSize: 11, color: '#94a3b8' }}>já tenho conta</span>
            <div style={{ flex: 1, height: 0.5, background: '#cbd5e1' }} />
          </div>

          {/* Login existente — discreto */}
          <button
            onClick={() => router.push('/login')}
            style={{ width: '100%', background: 'white', border: '0.5px solid #cbd5e1', borderRadius: 9, padding: '10px', fontSize: 13, color: '#475569', cursor: 'pointer' }}
          >
            Entrar na minha conta
          </button>

          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 5, marginTop: 16 }}>
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
              <path d="M5.5 1L2 2.8v2.7c0 2.3 1.5 4 3.5 4.5 2-.5 3.5-2.2 3.5-4.5V2.8L5.5 1z" stroke="#94a3b8" strokeWidth="0.9" fill="none"/>
            </svg>
            <span style={{ fontSize: 10, color: '#94a3b8' }}>Conexão segura · Dados criptografados · LGPD</span>
          </div>
        </div>

        <style>{`
          @keyframes pulseBtn {
            0%,100% { box-shadow: 0 0 0 0 rgba(10,61,143,0.4); }
            60%      { box-shadow: 0 0 0 10px rgba(10,61,143,0); }
          }
        `}</style>
      </div>
    </div>
  )
}
