'use client'
import { useState } from 'react'
import { Eye, EyeOff, Zap, ArrowRight, Shield, Activity, Lock } from 'lucide-react'

export default function LoginPage() {
  const [form, setForm] = useState({ tenantSlug: '', email: '', password: '' })
  const [showPass, setShowPass] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [step, setStep] = useState<'slug' | 'credentials'>('slug')

  const handleSlugNext = (e: React.FormEvent) => {
    e.preventDefault()
    if (form.tenantSlug.trim()) setStep('credentials')
  }

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          tenant_slug: form.tenantSlug,
          email: form.email,
          password: form.password,
        }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Falha no login')
      // Salvar access_token em memória (nunca localStorage)
      sessionStorage.setItem('access_token', data.access_token)
      window.location.href = '/dashboard'
    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex" style={{ background: '#F2F4FA' }}>

      {/* Left panel — branding */}
      <div
        className="hidden lg:flex flex-col justify-between w-[480px] flex-shrink-0 p-12 relative overflow-hidden"
        style={{ background: 'linear-gradient(145deg, #0F2D8A 0%, #1A47C8 50%, #2255E0 100%)' }}
      >
        {/* Background decoration */}
        <div style={{
          position: 'absolute', inset: 0, opacity: 0.06,
          backgroundImage: 'radial-gradient(circle at 1px 1px, white 1px, transparent 0)',
          backgroundSize: '32px 32px',
        }} />
        <div style={{
          position: 'absolute', bottom: -100, right: -100,
          width: 400, height: 400,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(255,255,255,0.08) 0%, transparent 70%)',
        }} />
        <div style={{
          position: 'absolute', top: -60, left: -60,
          width: 240, height: 240,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(255,255,255,0.05) 0%, transparent 70%)',
        }} />

        {/* Logo */}
        <div className="relative">
          <div className="flex items-center gap-3">
            <div style={{
              width: 44, height: 44, borderRadius: 14,
              background: 'rgba(255,255,255,0.15)',
              border: '1px solid rgba(255,255,255,0.2)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              backdropFilter: 'blur(10px)',
            }}>
              <Zap size={22} className="text-white" strokeWidth={2.5} />
            </div>
            <div>
              <div style={{ fontFamily: 'var(--font-syne)', fontWeight: 800, fontSize: 24, color: '#fff', letterSpacing: '-0.02em' }}>
                NexoOne
              </div>
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase' }}>
                ERP Inteligente
              </div>
            </div>
          </div>
        </div>

        {/* Middle content */}
        <div className="relative space-y-8">
          <div>
            <p style={{ fontFamily: 'var(--font-syne)', fontSize: 34, fontWeight: 800, color: '#fff', lineHeight: 1.15, letterSpacing: '-0.03em' }}>
              Do TOTVS ao cafezinho — gestão para todos.
            </p>
            <p style={{ fontSize: 15, color: 'rgba(255,255,255,0.65)', marginTop: 16, lineHeight: 1.7 }}>
              ERP multi-nicho com motor fiscal IBS/CBS 2026, IA assistente com aprovação humana e zero burocracia.
            </p>
          </div>

          {/* Feature pills */}
          <div className="space-y-3">
            {[
              { icon: <Shield size={14} />, text: 'Dados isolados por empresa — Row Level Security' },
              { icon: <Activity size={14} />, text: 'Motor fiscal IBS/CBS 2026 automático' },
              { icon: <Lock size={14} />, text: 'IA assistente com aprovação humana obrigatória' },
            ].map((f, i) => (
              <div key={i} className="flex items-center gap-3" style={{
                background: 'rgba(255,255,255,0.08)',
                border: '1px solid rgba(255,255,255,0.12)',
                borderRadius: 12, padding: '10px 14px',
              }}>
                <div style={{ color: 'rgba(255,255,255,0.7)' }}>{f.icon}</div>
                <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.8)', fontWeight: 500 }}>{f.text}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="relative">
          <p style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', fontWeight: 500 }}>
            © 2026 Nexo One ERP · Reforma Tributária Brasil 2026
          </p>
        </div>
      </div>

      {/* Right panel — form */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-[400px] animate-fade-in">

          {/* Mobile logo */}
          <div className="lg:hidden flex items-center gap-2.5 mb-10 justify-center">
            <div style={{ width: 36, height: 36, borderRadius: 10, background: 'linear-gradient(135deg,#1A47C8,#0F2D8A)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Zap size={18} className="text-white" strokeWidth={2.5} />
            </div>
            <span style={{ fontFamily: 'var(--font-syne)', fontWeight: 800, fontSize: 22, color: '#0D1B4B', letterSpacing: '-0.02em' }}>
              Nexo<span style={{ color: '#1A47C8' }}>One</span>
            </span>
          </div>

          {/* Step indicator */}
          <div className="flex items-center gap-2 mb-8">
            {['Empresa', 'Acesso'].map((s, i) => {
              const active = (i === 0 && step === 'slug') || (i === 1 && step === 'credentials')
              const done = i === 0 && step === 'credentials'
              return (
                <div key={s} className="flex items-center gap-2">
                  <div style={{
                    width: 24, height: 24, borderRadius: '50%',
                    background: done ? '#047857' : active ? '#1A47C8' : 'rgba(26,51,120,0.1)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 11, fontWeight: 700,
                    color: (done || active) ? '#fff' : '#8892B8',
                    transition: 'all 0.3s',
                  }}>
                    {done ? '✓' : i + 1}
                  </div>
                  <span style={{ fontSize: 12, fontWeight: 600, color: active ? '#0D1B4B' : '#8892B8', transition: 'all 0.3s' }}>{s}</span>
                  {i < 1 && <div style={{ width: 24, height: 1, background: done ? '#047857' : 'rgba(26,51,120,0.12)', transition: 'all 0.3s' }} />}
                </div>
              )
            })}
          </div>

          {/* Heading */}
          <div className="mb-8">
            <h1 style={{ fontFamily: 'var(--font-syne)', fontSize: 28, fontWeight: 800, color: '#0D1B4B', letterSpacing: '-0.025em', lineHeight: 1.2 }}>
              {step === 'slug' ? 'Bem-vindo de volta' : `Olá, ${form.tenantSlug}`}
            </h1>
            <p style={{ fontSize: 14, color: '#8892B8', marginTop: 6 }}>
              {step === 'slug'
                ? 'Digite o identificador da sua empresa para continuar.'
                : 'Insira suas credenciais para acessar o sistema.'}
            </p>
          </div>

          {/* STEP 1: Tenant slug */}
          {step === 'slug' && (
            <form onSubmit={handleSlugNext} className="space-y-5 animate-slide-up">
              <div>
                <label className="label">Identificador da empresa</label>
                <div className="relative">
                  <input
                    autoFocus
                    value={form.tenantSlug}
                    onChange={e => setForm(f => ({ ...f, tenantSlug: e.target.value.toLowerCase().replace(/\s/g, '-') }))}
                    placeholder="mecanica-do-joao"
                    className="input pr-32"
                    style={{ paddingLeft: 14 }}
                  />
                  <span style={{
                    position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)',
                    fontSize: 11, color: '#8892B8', fontFamily: 'var(--font-jetbrains)',
                    background: 'rgba(26,51,120,0.05)', padding: '2px 8px', borderRadius: 6,
                    border: '1px solid rgba(26,51,120,0.1)',
                  }}>
                    .nexoone.com
                  </span>
                </div>
                <p style={{ fontSize: 11, color: '#B0B8D8', marginTop: 5 }}>
                  Fornecido no seu e-mail de cadastro
                </p>
              </div>
              <button type="submit" className="btn-primary w-full justify-center py-3" style={{ fontSize: 14 }}>
                Continuar
                <ArrowRight size={15} />
              </button>
            </form>
          )}

          {/* STEP 2: Email + Password */}
          {step === 'credentials' && (
            <form onSubmit={handleLogin} className="space-y-5 animate-slide-up">
              {/* Error */}
              {error && (
                <div style={{
                  padding: '12px 14px', borderRadius: 10,
                  background: '#FFF1F1', border: '1px solid rgba(197,48,48,0.2)',
                  fontSize: 13, color: '#C53030', fontWeight: 500,
                }}>
                  {error}
                </div>
              )}

              <div>
                <label className="label">E-mail</label>
                <input
                  type="email"
                  autoFocus
                  value={form.email}
                  onChange={e => setForm(f => ({ ...f, email: e.target.value }))}
                  placeholder="joao@mecanicadojoao.com.br"
                  className="input"
                  required
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="label" style={{ marginBottom: 0 }}>Senha</label>
                  <button type="button" style={{ fontSize: 12, color: '#1A47C8', fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer' }}>
                    Esqueci a senha
                  </button>
                </div>
                <div className="relative">
                  <input
                    type={showPass ? 'text' : 'password'}
                    value={form.password}
                    onChange={e => setForm(f => ({ ...f, password: e.target.value }))}
                    placeholder="••••••••"
                    className="input pr-11"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPass(v => !v)}
                    style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: '#8892B8' }}
                  >
                    {showPass ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="btn-primary w-full justify-center py-3"
                style={{ fontSize: 14, marginTop: 8 }}
              >
                {loading ? (
                  <div style={{ width: 16, height: 16, border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#fff', borderRadius: '50%', animation: 'spin 0.7s linear infinite' }} />
                ) : (
                  <>Entrar no sistema <ArrowRight size={15} /></>
                )}
              </button>

              <button
                type="button"
                onClick={() => { setStep('slug'); setError('') }}
                className="btn-ghost w-full justify-center"
                style={{ fontSize: 13 }}
              >
                ← Trocar empresa
              </button>
            </form>
          )}

          {/* Security note */}
          <div className="mt-8 flex items-center gap-2 justify-center">
            <Lock size={12} style={{ color: '#B0B8D8' }} />
            <p style={{ fontSize: 11, color: '#B0B8D8' }}>
              Conexão segura · Dados criptografados · LGPD compliant
            </p>
          </div>
        </div>
      </div>

      <style jsx global>{`
        @keyframes spin { to { transform: rotate(360deg); } }
      `}</style>
    </div>
  )
}
