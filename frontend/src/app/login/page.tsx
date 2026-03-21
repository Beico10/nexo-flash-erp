'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Head from 'next/head'

export default function LoginPage() {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [mounted, setMounted] = useState(false)

  useEffect(() => { setMounted(true) }, [])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error || 'E-mail ou senha incorretos')
        return
      }

      // Salvar token e redirecionar
      if (typeof window !== 'undefined') {
        localStorage.setItem('nexo_token', data.access_token)
        if (data.tenant_slug) {
          localStorage.setItem('nexo_tenant', data.tenant_slug)
        }
      }

      router.push('/dashboard')
    } catch {
      setError('Erro de conexão. Tente novamente.')
    } finally {
      setLoading(false)
    }
  }

  const handleExplore = () => {
    // Entra no sistema sem login — modo demonstração
    if (typeof window !== 'undefined') {
      localStorage.setItem('nexo_token', 'demo-token')
      localStorage.setItem('nexo_tenant', 'demo')
      localStorage.setItem('nexo_demo_mode', 'true')
      sessionStorage.setItem('access_token', 'demo-token')
      window.location.href = '/dashboard'
    }
  }

  const handleTrial = () => {
    // Por enquanto, entra em modo demo também
    if (typeof window !== 'undefined') {
      localStorage.setItem('nexo_token', 'demo-token')
      localStorage.setItem('nexo_tenant', 'demo')
      localStorage.setItem('nexo_demo_mode', 'true')
      sessionStorage.setItem('access_token', 'demo-token')
      window.location.href = '/dashboard'
    }
  }

  if (!mounted) return null

  return (
    <>
      <Head>
        <link href='https://fonts.googleapis.com/css2?family=Montserrat:wght@800;900&display=swap' rel='stylesheet' />
      </Head>
      <div style={{
        display: 'flex',
        minHeight: '100vh',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      }}>

      {/* ── LADO ESQUERDO — azul ── */}
      <div style={{
        width: '46%',
        background: '#0A3D8F',
        padding: '40px 36px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
      }}>
        <div>

          {/* Logo */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 36 }}>
            <div style={{
              width: 36, height: 36, borderRadius: 8,
              background: 'rgba(255,255,255,0.15)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              flexShrink: 0,
            }}>
              <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                <rect x="2" y="2" width="6" height="6" rx="1.5" fill="white" opacity="0.9"/>
                <rect x="10" y="2" width="6" height="6" rx="1.5" fill="white" opacity="0.55"/>
                <rect x="2" y="10" width="6" height="6" rx="1.5" fill="white" opacity="0.55"/>
                <rect x="10" y="10" width="6" height="6" rx="1.5" fill="white" opacity="0.9"/>
              </svg>
            </div>
            <div>
              <div style={{ fontSize: 15, fontWeight: 500, color: 'white', letterSpacing: -0.3 }}>
                Gestão Para Todos
              </div>
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginTop: 1 }}>
                ERP inteligente
              </div>
            </div>
          </div>

          {/* Headline + Badge lado a lado */}
          <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start', marginBottom: 28 }}>

            {/* Headline */}
            <div style={{ flex: 1 }}>
              <p style={{ fontSize: 26, fontWeight: 900, fontFamily: "'Montserrat', sans-serif", color: 'white', lineHeight: 1.3, margin: '0 0 12px' }}>
                Você trabalha.<br />A gente cuida<br />da gestão.
              </p>
              <p style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', lineHeight: 1.65, margin: 0 }}>
                Não pague imposto duas vezes.<br />
                Gerencie tudo em um só lugar.
              </p>
            </div>

            {/* Badge Reforma Tributária */}
            <div style={{
              width: 130, flexShrink: 0,
              background: 'rgba(255,255,255,0.1)',
              border: '1px solid rgba(255,255,255,0.2)',
              borderRadius: 12, padding: '14px 12px',
              textAlign: 'center',
            }}>
              <div style={{
                width: 10, height: 10, borderRadius: '50%',
                background: '#4ADE80', margin: '0 auto 10px',
                animation: 'dotPulse 2s infinite',
              }} />
              <p style={{ fontSize: 11, fontWeight: 500, color: 'white', margin: '0 0 6px', lineHeight: 1.35 }}>
                Pronto para a<br />Reforma Tributária
              </p>
              <p style={{ fontSize: 24, fontWeight: 500, color: '#4ADE80', margin: '0 0 6px' }}>
                2026
              </p>
              <p style={{ fontSize: 10, color: 'rgba(255,255,255,0.5)', margin: 0, lineHeight: 1.4 }}>
                IBS + CBS calculados automaticamente
              </p>
            </div>
          </div>

          {/* Benefícios */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
            {[
              'Dados criptografados — segurança bancária',
              'IBS/CBS 2026 calculado automaticamente',
              'IA assistente com aprovação humana',
              'Empresa pequena, média ou grande — nosso sistema atende',
            ].map((item, i) => (
              <div key={i} style={{
                display: 'flex', alignItems: 'center', gap: 10,
                background: 'rgba(255,255,255,0.07)',
                borderRadius: 8, padding: '9px 13px',
              }}>
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none" style={{ flexShrink: 0 }}>
                  <path d="M1 6l3.5 3.5L11 2" stroke="rgba(255,255,255,0.7)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.8)' }}>{item}</span>
              </div>
            ))}
          </div>
        </div>

        <p style={{ fontSize: 11, color: 'rgba(255,255,255,0.2)', margin: '24px 0 0' }}>
          © 2026 Gestão Para Todos · Reforma Tributária Brasil
        </p>
      </div>

      {/* ── LADO DIREITO — azul claro com pontos ── */}
      <div style={{
        flex: 1,
        background: '#F0F4FF',
        padding: '40px 44px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
      }}>

        {/* Padrão de pontos decorativos */}
        <svg
          style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}
          xmlns="http://www.w3.org/2000/svg"
        >
          <defs>
            <pattern id="dots" x="0" y="0" width="24" height="24" patternUnits="userSpaceOnUse">
              <circle cx="2" cy="2" r="1.2" fill="#0A3D8F" opacity="0.07"/>
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#dots)"/>
        </svg>

        <div style={{ maxWidth: 320, margin: '0 auto', width: '100%', position: 'relative' }}>

          {/* Saudação */}
          <div style={{ marginBottom: 24 }}>
            <p style={{ fontSize: 22, fontWeight: 500, color: '#1a1a2e', margin: '0 0 3px' }}>
              Bem-vindo ao
            </p>
            <p style={{ fontSize: 22, fontWeight: 500, color: '#0A3D8F', margin: '0 0 8px' }}>
              Gestão Para Todos
            </p>
            <p style={{ fontSize: 12, color: '#64748b', margin: 0 }}>
              Conheça o sistema sem precisar se cadastrar
            </p>
          </div>

          {/* BOTÃO PRINCIPAL — explorar sem login */}
          <button
            type="button"
            onClick={handleExplore}
            data-testid="explore-btn"
            style={{
              width: '100%',
              background: '#0A3D8F',
              borderRadius: 8, border: 'none',
              padding: '15px 18px', textAlign: 'center',
              cursor: 'pointer', marginBottom: 8,
              animation: 'pulseBtn 2s infinite',
              color: 'white',
              fontSize: 14,
              fontWeight: 500,
            }}
          >
            Explorar o sistema agora →
          </button>

          {/* Trial */}
          <button
            type="button"
            onClick={handleTrial}
            data-testid="trial-btn"
            style={{
              width: '100%',
              background: 'white',
              border: '1.5px solid #0A3D8F',
              borderRadius: 8, padding: '12px 16px',
              textAlign: 'center', cursor: 'pointer',
              marginBottom: 20,
              color: '#0A3D8F',
              fontSize: 13,
              fontWeight: 500,
            }}
          >
            Começar grátis por 7 dias
          </button>

          {/* Divisor */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 14 }}>
            <div style={{ flex: 1, height: 0.5, background: '#cbd5e1' }} />
            <span style={{ fontSize: 11, color: '#94a3b8' }}>já tenho conta</span>
            <div style={{ flex: 1, height: 0.5, background: '#cbd5e1' }} />
          </div>

          {/* Formulário de login */}
          <form onSubmit={handleLogin}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 9, marginBottom: 12 }}>
              <input
                type="email"
                placeholder="seu@email.com.br"
                value={email}
                onChange={e => setEmail(e.target.value)}
                style={{
                  border: '0.5px solid #cbd5e1',
                  borderRadius: 8, padding: '11px 14px',
                  fontSize: 13, color: '#1a1a2e',
                  background: 'white', outline: 'none',
                  width: '100%', boxSizing: 'border-box',
                }}
              />
              <input
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={e => setPassword(e.target.value)}
                style={{
                  border: '0.5px solid #cbd5e1',
                  borderRadius: 8, padding: '11px 14px',
                  fontSize: 13, color: '#1a1a2e',
                  background: 'white', outline: 'none',
                  width: '100%', boxSizing: 'border-box',
                }}
              />
              <div style={{ textAlign: 'right' }}>
                <a href="/recuperar-senha" style={{ fontSize: 11, color: '#0A3D8F', textDecoration: 'none' }}>
                  Esqueceu a senha?
                </a>
              </div>
            </div>

            {error && (
              <p style={{ fontSize: 12, color: '#ef4444', marginBottom: 10, textAlign: 'center' }}>
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              style={{
                width: '100%',
                border: '0.5px solid #cbd5e1',
                background: 'white',
                borderRadius: 8, padding: '11px',
                textAlign: 'center', fontSize: 13,
                color: '#475569', cursor: 'pointer',
                marginBottom: 20,
                opacity: loading ? 0.7 : 1,
              }}
            >
              {loading ? 'Entrando...' : 'Entrar na minha conta'}
            </button>
          </form>

          {/* Segurança */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 5 }}>
            <svg width="11" height="11" viewBox="0 0 11 11" fill="none">
              <path d="M5.5 1L2 2.8v2.7c0 2.3 1.5 4 3.5 4.5 2-.5 3.5-2.2 3.5-4.5V2.8L5.5 1z" stroke="#94a3b8" strokeWidth="0.9" fill="none"/>
            </svg>
            <span style={{ fontSize: 11, color: '#94a3b8' }}>
              Conexão segura · Dados criptografados · LGPD
            </span>
          </div>
        </div>

        {/* Animações CSS */}
        <style>{`
          @keyframes pulseBtn {
            0%   { box-shadow: 0 0 0 0 rgba(10,61,143,0.4); }
            60%  { box-shadow: 0 0 0 12px rgba(10,61,143,0); }
            100% { box-shadow: 0 0 0 0 rgba(10,61,143,0); }
          }
          @keyframes dotPulse {
            0%,100% { opacity: 1; transform: scale(1); }
            50%      { opacity: 0.6; transform: scale(0.85); }
          }
          @media (max-width: 768px) {
            div[style*="width: 46%"] {
              width: 100% !important;
            }
          }
        `}</style>
      </div>
    </div>
    </>
  )
}
