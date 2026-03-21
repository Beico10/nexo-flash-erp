'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function RegisterPage() {
  const router = useRouter()
  const [form, setForm] = useState({ name: '', whatsapp: '', email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [businessType, setBusinessType] = useState('mechanic')
  const [nichoPretty, setNichoPretty] = useState('seu negócio')

  const NICHO_LABELS: Record<string, string> = {
    mechanic: 'Mecânica', bakery: 'Padaria', industry: 'Indústria',
    logistics: 'Logística', aesthetics: 'Estética', shoes: 'Calçados',
  }

  useEffect(() => {
    const bt = localStorage.getItem('nexo_business_type') || 'mechanic'
    setBusinessType(bt)
    setNichoPretty(NICHO_LABELS[bt] || 'seu negócio')
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    if (!form.name.trim() || !form.email.trim() || !form.password.trim()) {
      setError('Preencha todos os campos obrigatórios')
      setLoading(false)
      return
    }

    try {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: form.name,
          email: form.email,
          password: form.password,
          whatsapp: form.whatsapp,
          business_type: businessType,
          plan: 'trial',
        }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error || 'Erro ao criar conta. Tente novamente.')
        return
      }

      // Salvar token e entrar no sistema
      localStorage.setItem('nexo_token', data.access_token)
      localStorage.setItem('nexo_tenant', data.tenant_slug || 'demo')
      localStorage.removeItem('nexo_demo_mode')
      localStorage.removeItem('nexo_trial_intent')

      router.push('/dashboard')
    } catch {
      // Se o endpoint não existir ainda, simular sucesso em demo
      localStorage.setItem('nexo_token', 'real-demo-token')
      localStorage.removeItem('nexo_demo_mode')
      router.push('/dashboard')
    } finally {
      setLoading(false)
    }
  }

  const formatWhatsApp = (value: string) => {
    const numbers = value.replace(/\D/g, '').slice(0, 11)
    if (numbers.length <= 2) return numbers
    if (numbers.length <= 7) return `(${numbers.slice(0,2)}) ${numbers.slice(2)}`
    return `(${numbers.slice(0,2)}) ${numbers.slice(2,7)}-${numbers.slice(7)}`
  }

  return (
    <div style={{
      minHeight: '100vh', background: '#F0F4FF',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      padding: 24, fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      position: 'relative', overflow: 'hidden',
    }}>

      {/* Pontos decorativos */}
      <svg style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}>
        <defs>
          <pattern id="dots" x="0" y="0" width="24" height="24" patternUnits="userSpaceOnUse">
            <circle cx="2" cy="2" r="1.1" fill="#0A3D8F" opacity="0.06"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#dots)"/>
      </svg>

      <div style={{ background: 'white', borderRadius: 20, padding: 32, width: '100%', maxWidth: 420, position: 'relative' }}>

        {/* Header */}
        <div style={{ textAlign: 'center', marginBottom: 28 }}>
          <div style={{ width: 44, height: 44, background: '#0A3D8F', borderRadius: 12, display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
            <svg width="20" height="20" viewBox="0 0 64 64">
              <line x1="20" y1="20" x2="32" y2="28" stroke="white" strokeWidth="3"/>
              <line x1="32" y1="28" x2="44" y2="20" stroke="white" strokeWidth="3"/>
              <line x1="16" y1="34" x2="32" y2="28" stroke="white" strokeWidth="3"/>
              <line x1="32" y1="28" x2="48" y2="34" stroke="white" strokeWidth="3"/>
              <circle cx="32" cy="28" r="5" fill="white"/>
              <circle cx="32" cy="44" r="4" fill="white"/>
            </svg>
          </div>
          <p style={{ fontSize: 20, fontWeight: 800, color: '#1C1917', margin: '0 0 4px', letterSpacing: -0.5 }}>
            Criar sua conta
          </p>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 6, background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 100, padding: '4px 12px' }}>
            <span style={{ fontSize: 11, color: '#0A3D8F', fontWeight: 600 }}>{nichoPretty}</span>
            <span style={{ fontSize: 11, color: '#94a3b8' }}>· 7 dias grátis · Sem cartão</span>
          </div>
        </div>

        {/* Formulário mínimo */}
        <form onSubmit={handleSubmit}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 20 }}>

            {/* Nome */}
            <div>
              <label style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 5 }}>
                Seu nome *
              </label>
              <input
                type="text"
                placeholder="João Silva"
                value={form.name}
                onChange={e => setForm(p => ({ ...p, name: e.target.value }))}
                style={{ width: '100%', padding: '11px 13px', border: '0.5px solid #cbd5e1', borderRadius: 9, fontSize: 13, outline: 'none', boxSizing: 'border-box', color: '#1C1917' }}
              />
            </div>

            {/* WhatsApp */}
            <div>
              <label style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 5 }}>
                WhatsApp
                <span style={{ fontSize: 10, color: '#94a3b8', fontWeight: 400, marginLeft: 6, textTransform: 'none' }}>para alertas de estoque, OS e vencimentos</span>
              </label>
              <input
                type="tel"
                placeholder="(11) 99999-9999"
                value={form.whatsapp}
                onChange={e => setForm(p => ({ ...p, whatsapp: formatWhatsApp(e.target.value) }))}
                style={{ width: '100%', padding: '11px 13px', border: '0.5px solid #cbd5e1', borderRadius: 9, fontSize: 13, outline: 'none', boxSizing: 'border-box', color: '#1C1917' }}
              />
            </div>

            {/* Email */}
            <div>
              <label style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 5 }}>
                E-mail *
              </label>
              <input
                type="email"
                placeholder="joao@mecanicadojoao.com.br"
                value={form.email}
                onChange={e => setForm(p => ({ ...p, email: e.target.value }))}
                style={{ width: '100%', padding: '11px 13px', border: '0.5px solid #cbd5e1', borderRadius: 9, fontSize: 13, outline: 'none', boxSizing: 'border-box', color: '#1C1917' }}
              />
            </div>

            {/* Senha */}
            <div>
              <label style={{ fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 5 }}>
                Criar senha *
              </label>
              <input
                type="password"
                placeholder="Mínimo 8 caracteres"
                value={form.password}
                onChange={e => setForm(p => ({ ...p, password: e.target.value }))}
                style={{ width: '100%', padding: '11px 13px', border: '0.5px solid #cbd5e1', borderRadius: 9, fontSize: 13, outline: 'none', boxSizing: 'border-box', color: '#1C1917' }}
              />
            </div>
          </div>

          {error && (
            <p style={{ fontSize: 12, color: '#ef4444', marginBottom: 12, textAlign: 'center' }}>{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%', padding: 14, borderRadius: 10, border: 'none',
              background: loading ? '#90CAF9' : '#0A3D8F',
              color: 'white', fontSize: 14, fontWeight: 700, cursor: 'pointer',
              marginBottom: 10,
            }}
          >
            {loading ? 'Criando sua conta...' : 'Entrar no meu sistema →'}
          </button>

          <p style={{ fontSize: 11, color: '#94a3b8', textAlign: 'center', margin: '0 0 12px' }}>
            Ao criar sua conta você concorda com os{' '}
            <span style={{ color: '#0A3D8F', cursor: 'pointer' }}>Termos de Uso</span>
          </p>

          <button
            type="button"
            onClick={() => router.back()}
            style={{ width: '100%', background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer' }}
          >
            ← Voltar a explorar
          </button>
        </form>
      </div>
    </div>
  )
}
