'use client'
import { useState, useEffect } from 'react'
import { Check, Plus, X, Star, Zap, ArrowRight, Info } from 'lucide-react'

interface Module {
  id: string; name: string; description: string; icon: string
  price: number; price_yearly: number; category: string
  for_niches: string[]; features: string[]
  is_popular: boolean; trial_days: number
}

interface ModuleSubscription {
  module_id: string; status: string; price: number
  trial_ends_at?: string; renews_at: string
}

const CATEGORY_LABEL: Record<string, string> = {
  operacional: '⚙️ Operacional',
  fiscal: '📋 Fiscal',
  logistica: '🚛 Logística',
  ia: '🤖 Inteligência Artificial',
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export default function ModulesPage() {
  const [catalog, setCatalog] = useState<Module[]>([])
  const [subscriptions, setSubscriptions] = useState<ModuleSubscription[]>([])
  const [loading, setLoading] = useState(true)
  const [cycle, setCycle] = useState<'monthly' | 'yearly'>('monthly')
  const [subscribing, setSubscribing] = useState<string | null>(null)
  const [selected, setSelected] = useState<Module | null>(null)
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  useEffect(() => { fetchData() }, [])

  const fetchData = async () => {
    setLoading(true)
    const h = { Authorization: `Bearer ${token}` }
    try {
      const [catRes, subRes] = await Promise.all([
        fetch('/api/v1/modules/catalog', { headers: h }),
        fetch('/api/v1/modules/subscriptions', { headers: h }),
      ])
      if (catRes.ok) setCatalog((await catRes.json()).modules || [])
      if (subRes.ok) setSubscriptions((await subRes.json()).subscriptions || [])
    } finally { setLoading(false) }
  }

  const isSubscribed = (moduleId: string) =>
    subscriptions.some(s => s.module_id === moduleId && s.status !== 'cancelled')

  const handleSubscribe = async (moduleId: string) => {
    setSubscribing(moduleId)
    try {
      const res = await fetch('/api/v1/modules/subscribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ module_id: moduleId, cycle }),
      })
      if (res.ok) { await fetchData(); setSelected(null) }
    } finally { setSubscribing(null) }
  }

  const handleCancel = async (moduleId: string) => {
    if (!confirm('Cancelar este módulo? Você perderá acesso ao final do período.')) return
    await fetch(`/api/v1/modules/${moduleId}/cancel`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    })
    fetchData()
  }

  // Agrupar por categoria
  const byCategory = catalog.reduce((acc, m) => {
    if (!acc[m.category]) acc[m.category] = []
    acc[m.category].push(m)
    return acc
  }, {} as Record<string, Module[]>)

  // Calcular custo total dos módulos avulsos
  const totalAddon = subscriptions
    .filter(s => s.status === 'active' || s.status === 'trialing')
    .reduce((sum, s) => sum + s.price, 0)

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>🧩 Módulos</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>
            Contrate só o que precisa — encaixa em qualquer sistema
          </p>
        </div>
        {totalAddon > 0 && (
          <div style={{ background: '#E3F2FD', padding: '10px 16px', borderRadius: 10, textAlign: 'right' }}>
            <p style={{ fontSize: 11, color: '#757575' }}>Módulos avulsos</p>
            <p style={{ fontSize: 18, fontWeight: 700, color: '#1565C0' }}>{fmt(totalAddon)}/mês</p>
          </div>
        )}
      </div>

      {/* Banner tema principal */}
      <div style={{ background: 'linear-gradient(135deg, #0A3D8F, #1565C0)', borderRadius: 16, padding: '20px 24px', marginBottom: 28, display: 'flex', gap: 16, alignItems: 'center' }}>
        <div style={{ fontSize: 32 }}>⚡</div>
        <div>
          <p style={{ fontSize: 15, fontWeight: 700, color: 'white', marginBottom: 4 }}>
            Você paga imposto duas vezes sem saber?
          </p>
          <p style={{ fontSize: 13, color: 'rgba(255,255,255,0.8)' }}>
            O Motor Fiscal IBS/CBS 2026 calcula automaticamente o crédito das suas compras e abate na venda. Experimente 30 dias grátis.
          </p>
        </div>
        <button onClick={() => handleSubscribe('fiscal_ibs_cbs')} style={{
          background: 'white', color: '#0A3D8F', padding: '10px 20px',
          borderRadius: 10, border: 'none', cursor: 'pointer',
          fontWeight: 700, fontSize: 13, flexShrink: 0,
          display: 'flex', alignItems: 'center', gap: 6,
        }}>
          Ativar grátis <ArrowRight size={14} />
        </button>
      </div>

      {/* Toggle mensal/anual */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 24 }}>
        <span style={{ fontSize: 13, color: cycle === 'monthly' ? '#212121' : '#9E9E9E', fontWeight: cycle === 'monthly' ? 700 : 400 }}>Mensal</span>
        <div onClick={() => setCycle(c => c === 'monthly' ? 'yearly' : 'monthly')} style={{
          width: 44, height: 24, borderRadius: 12, cursor: 'pointer',
          background: cycle === 'yearly' ? '#1565C0' : '#E0E4F0',
          position: 'relative', transition: 'background 0.2s',
        }}>
          <div style={{
            position: 'absolute', top: 3, left: cycle === 'yearly' ? 23 : 3,
            width: 18, height: 18, borderRadius: '50%', background: 'white',
            transition: 'left 0.2s', boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
          }} />
        </div>
        <span style={{ fontSize: 13, color: cycle === 'yearly' ? '#212121' : '#9E9E9E', fontWeight: cycle === 'yearly' ? 700 : 400 }}>
          Anual <span style={{ background: '#E8F5E9', color: '#2E7D32', padding: '1px 6px', borderRadius: 100, fontSize: 10, fontWeight: 700 }}>-20%</span>
        </span>
      </div>

      {/* Catálogo por categoria */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#757575' }}>Carregando módulos...</div>
      ) : (
        Object.entries(byCategory).map(([category, modules]) => (
          <div key={category} style={{ marginBottom: 32 }}>
            <p style={{ fontSize: 13, fontWeight: 700, color: '#424242', marginBottom: 14, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
              {CATEGORY_LABEL[category] || category}
            </p>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
              {modules.map(m => {
                const subscribed = isSubscribed(m.id)
                const price = cycle === 'yearly' ? m.price_yearly : m.price
                return (
                  <div key={m.id} style={{
                    background: 'white', borderRadius: 16, padding: 20,
                    border: `1.5px solid ${subscribed ? '#A5D6A7' : m.is_popular ? '#90CAF9' : '#E0E4F0'}`,
                    position: 'relative', display: 'flex', flexDirection: 'column',
                    boxShadow: m.is_popular ? '0 4px 16px rgba(21,101,192,0.1)' : 'none',
                  }}>
                    {/* Badge popular */}
                    {m.is_popular && !subscribed && (
                      <div style={{ position: 'absolute', top: -10, right: 16, background: '#1565C0', color: 'white', fontSize: 10, fontWeight: 700, padding: '2px 10px', borderRadius: 100, display: 'flex', alignItems: 'center', gap: 3 }}>
                        <Star size={9} fill="white" /> POPULAR
                      </div>
                    )}
                    {subscribed && (
                      <div style={{ position: 'absolute', top: -10, right: 16, background: '#2E7D32', color: 'white', fontSize: 10, fontWeight: 700, padding: '2px 10px', borderRadius: 100, display: 'flex', alignItems: 'center', gap: 3 }}>
                        <Check size={9} /> ATIVO
                      </div>
                    )}

                    {/* Header */}
                    <div className="flex items-center gap-2 mb-3">
                      <span style={{ fontSize: 28 }}>{m.icon}</span>
                      <div>
                        <p style={{ fontSize: 14, fontWeight: 700, color: '#212121' }}>{m.name}</p>
                        <p style={{ fontSize: 11, color: '#757575', marginTop: 1 }}>{m.trial_days} dias grátis</p>
                      </div>
                    </div>

                    {/* Descrição */}
                    <p style={{ fontSize: 12, color: '#616161', lineHeight: 1.5, marginBottom: 14, flex: 1 }}>
                      {m.description}
                    </p>

                    {/* Features (top 3) */}
                    <div style={{ marginBottom: 16 }}>
                      {m.features.slice(0, 3).map((f, i) => (
                        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 4 }}>
                          <Check size={11} style={{ color: '#2E7D32', flexShrink: 0 }} />
                          <span style={{ fontSize: 11, color: '#424242' }}>{f}</span>
                        </div>
                      ))}
                      {m.features.length > 3 && (
                        <button onClick={() => setSelected(m)} style={{ fontSize: 11, color: '#1565C0', background: 'none', border: 'none', cursor: 'pointer', padding: 0, marginTop: 2 }}>
                          +{m.features.length - 3} mais funcionalidades →
                        </button>
                      )}
                    </div>

                    {/* Preço e botão */}
                    <div style={{ borderTop: '1px solid #F0F2F8', paddingTop: 14 }}>
                      <div className="flex items-center justify-between mb-10">
                        <div>
                          <span style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>
                            {fmt(price)}
                          </span>
                          <span style={{ fontSize: 11, color: '#9E9E9E' }}>/mês</span>
                          {cycle === 'yearly' && (
                            <p style={{ fontSize: 10, color: '#2E7D32', fontWeight: 700 }}>
                              vs {fmt(m.price)}/mês no mensal
                            </p>
                          )}
                        </div>
                      </div>

                      {subscribed ? (
                        <button onClick={() => handleCancel(m.id)} style={{
                          width: '100%', padding: '9px', borderRadius: 10,
                          background: '#F5F5F5', border: '1px solid #E0E0E0',
                          color: '#757575', fontSize: 12, fontWeight: 600, cursor: 'pointer',
                        }}>
                          Cancelar módulo
                        </button>
                      ) : (
                        <button onClick={() => handleSubscribe(m.id)} disabled={subscribing === m.id} style={{
                          width: '100%', padding: '10px', borderRadius: 10, border: 'none',
                          background: subscribing === m.id ? '#90CAF9' : 'linear-gradient(135deg, #0A3D8F, #1565C0)',
                          color: 'white', fontSize: 13, fontWeight: 700, cursor: 'pointer',
                          display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
                        }}>
                          {subscribing === m.id ? 'Ativando...' : <><Zap size={14} /> Experimentar grátis</>}
                        </button>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ))
      )}

      {/* Modal detalhes do módulo */}
      {selected && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 20, padding: 32, width: 480, maxHeight: '90vh', overflowY: 'auto' }}>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <span style={{ fontSize: 32 }}>{selected.icon}</span>
                <h3 style={{ fontSize: 18, fontWeight: 700 }}>{selected.name}</h3>
              </div>
              <button onClick={() => setSelected(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            <p style={{ fontSize: 14, color: '#616161', lineHeight: 1.6, marginBottom: 20 }}>{selected.description}</p>

            <p style={{ fontSize: 12, fontWeight: 700, color: '#424242', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 10 }}>
              Todas as funcionalidades
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 24 }}>
              {selected.features.map((f, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div style={{ width: 20, height: 20, borderRadius: '50%', background: '#E8F5E9', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <Check size={12} style={{ color: '#2E7D32' }} />
                  </div>
                  <span style={{ fontSize: 13, color: '#424242' }}>{f}</span>
                </div>
              ))}
            </div>

            <div style={{ background: '#F5F7FF', borderRadius: 12, padding: 16, marginBottom: 20, display: 'flex', gap: 16 }}>
              <div style={{ textAlign: 'center', flex: 1 }}>
                <p style={{ fontSize: 22, fontWeight: 700, color: '#0A3D8F' }}>{fmt(selected.price)}</p>
                <p style={{ fontSize: 11, color: '#757575' }}>por mês</p>
              </div>
              <div style={{ textAlign: 'center', flex: 1 }}>
                <p style={{ fontSize: 22, fontWeight: 700, color: '#2E7D32' }}>{fmt(selected.price_yearly)}</p>
                <p style={{ fontSize: 11, color: '#757575' }}>por mês (anual)</p>
              </div>
              <div style={{ textAlign: 'center', flex: 1 }}>
                <p style={{ fontSize: 22, fontWeight: 700, color: '#E65100' }}>{selected.trial_days}</p>
                <p style={{ fontSize: 11, color: '#757575' }}>dias grátis</p>
              </div>
            </div>

            <button onClick={() => handleSubscribe(selected.id)} style={{
              width: '100%', padding: 14, borderRadius: 12,
              background: 'linear-gradient(135deg, #0A3D8F, #1565C0)',
              color: 'white', fontWeight: 700, fontSize: 15, border: 'none', cursor: 'pointer',
            }}>
              <Zap size={16} style={{ display: 'inline', marginRight: 6 }} />
              Experimentar {selected.trial_days} dias grátis
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
