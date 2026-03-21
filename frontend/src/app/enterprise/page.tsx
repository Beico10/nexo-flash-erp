'use client'
import { useState, useEffect } from 'react'
import { Key, Plus, Trash2, Building2, Webhook, Copy, CheckCircle2, X, Eye, EyeOff, Globe } from 'lucide-react'

interface APIKey {
  id: string; name: string; key_prefix: string
  scopes: string[]; rate_limit: number; is_active: boolean
  last_used_at?: string; expires_at?: string; created_at: string
}

interface Subsidiary {
  id: string; name: string; cnpj: string
  city: string; state: string; is_headquarter: boolean; is_active: boolean
}

interface WebhookItem {
  id: string; url: string; events: string[]
  is_active: boolean; created_at: string
}

const SCOPES = [
  { key: 'read', label: '📖 Leitura', desc: 'Consultar dados' },
  { key: 'write', label: '✏️ Escrita', desc: 'Criar e editar' },
  { key: 'finance', label: '💰 Financeiro', desc: 'Contas e DRE' },
  { key: 'inventory', label: '📦 Estoque', desc: 'Produtos e movimentos' },
  { key: 'orders', label: '🔧 Pedidos', desc: 'OS e vendas' },
  { key: 'webhook', label: '🔔 Webhooks', desc: 'Gerenciar webhooks' },
]

const WEBHOOK_EVENTS = [
  'os.created', 'os.completed', 'payment.received',
  'stock.low', 'stock.out', 'invoice.issued',
  'customer.created', 'trial.started', 'trial.converted'
]

export default function EnterprisePage() {
  const [tab, setTab] = useState<'apikeys' | 'subsidiaries' | 'webhooks'>('apikeys')
  const [apiKeys, setApiKeys] = useState<APIKey[]>([])
  const [subsidiaries, setSubsidiaries] = useState<Subsidiary[]>([])
  const [webhooks, setWebhooks] = useState<WebhookItem[]>([])
  const [loading, setLoading] = useState(false)
  const [showKeyForm, setShowKeyForm] = useState(false)
  const [showSubForm, setShowSubForm] = useState(false)
  const [showWebhookForm, setShowWebhookForm] = useState(false)
  const [newKey, setNewKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [keyForm, setKeyForm] = useState({ name: '', scopes: ['read'], rate_limit: '60', expires_at: '' })
  const [subForm, setSubForm] = useState({ name: '', cnpj: '', city: '', state: '' })
  const [webhookForm, setWebhookForm] = useState({ url: '', events: ['os.created', 'payment.received'] })
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  useEffect(() => { fetchData() }, [tab])

  const fetchData = async () => {
    setLoading(true)
    const h = { Authorization: `Bearer ${token}` }
    try {
      if (tab === 'apikeys') {
        const res = await fetch('/api/v1/enterprise/api-keys', { headers: h })
        if (res.ok) setApiKeys((await res.json()).api_keys || [])
      } else if (tab === 'subsidiaries') {
        const res = await fetch('/api/v1/enterprise/subsidiaries', { headers: h })
        if (res.ok) setSubsidiaries((await res.json()).subsidiaries || [])
      } else {
        const res = await fetch('/api/v1/enterprise/webhooks', { headers: h })
        if (res.ok) setWebhooks((await res.json()).webhooks || [])
      }
    } finally { setLoading(false) }
  }

  const createAPIKey = async (e: React.FormEvent) => {
    e.preventDefault()
    const res = await fetch('/api/v1/enterprise/api-keys', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ ...keyForm, rate_limit: parseInt(keyForm.rate_limit) }),
    })
    if (res.ok) {
      const data = await res.json()
      setNewKey(data.key)
      setShowKeyForm(false)
      fetchData()
    }
  }

  const revokeKey = async (id: string) => {
    if (!confirm('Revogar esta API Key? Integrações que usam ela vão parar de funcionar.')) return
    const res = await fetch(`/api/v1/enterprise/api-keys/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) fetchData()
  }

  const createSubsidiary = async (e: React.FormEvent) => {
    e.preventDefault()
    const res = await fetch('/api/v1/enterprise/subsidiaries', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(subForm),
    })
    if (res.ok) { setShowSubForm(false); fetchData() }
  }

  const createWebhook = async (e: React.FormEvent) => {
    e.preventDefault()
    const res = await fetch('/api/v1/enterprise/webhooks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(webhookForm),
    })
    if (res.ok) { setShowWebhookForm(false); fetchData() }
  }

  const copyKey = () => {
    if (newKey) {
      navigator.clipboard.writeText(newKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const toggleScope = (scope: string) => {
    setKeyForm(prev => ({
      ...prev,
      scopes: prev.scopes.includes(scope)
        ? prev.scopes.filter(s => s !== scope)
        : [...prev.scopes, scope]
    }))
  }

  const toggleEvent = (event: string) => {
    setWebhookForm(prev => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter(e => e !== event)
        : [...prev.events, event]
    }))
  }

  return (
    <div style={{ padding: '24px', maxWidth: 900, margin: '0 auto' }}>

      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>⚡ Enterprise</h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>API Keys • Filiais • Webhooks</p>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', background: '#F0F2F8', borderRadius: 10, padding: 4, gap: 4, marginBottom: 24 }}>
        {[
          { key: 'apikeys', label: '🔑 API Keys', icon: Key },
          { key: 'subsidiaries', label: '🏢 Filiais', icon: Building2 },
          { key: 'webhooks', label: '🔔 Webhooks', icon: Webhook },
        ].map(t => (
          <button key={t.key} onClick={() => setTab(t.key as any)} style={{
            flex: 1, padding: '8px', borderRadius: 8, border: 'none', cursor: 'pointer',
            fontWeight: 600, fontSize: 13,
            background: tab === t.key ? 'white' : 'transparent',
            color: tab === t.key ? '#0A3D8F' : '#757575',
            boxShadow: tab === t.key ? '0 1px 4px rgba(0,0,0,0.1)' : 'none',
          }}>
            {t.label}
          </button>
        ))}
      </div>

      {/* Modal chave criada */}
      {newKey && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 200 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 32, width: 480 }}>
            <div style={{ textAlign: 'center', marginBottom: 20 }}>
              <div style={{ fontSize: 40, marginBottom: 8 }}>🔑</div>
              <h3 style={{ fontSize: 18, fontWeight: 700 }}>API Key criada!</h3>
              <p style={{ fontSize: 13, color: '#B71C1C', marginTop: 4, fontWeight: 600 }}>
                ⚠️ Copie agora — esta chave não será mostrada novamente
              </p>
            </div>
            <div style={{ background: '#F5F7FF', border: '1.5px solid #C5CAE9', borderRadius: 10, padding: 12, fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all', marginBottom: 16 }}>
              {newKey}
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={copyKey} style={{
                flex: 1, padding: 12, borderRadius: 10, border: 'none', cursor: 'pointer',
                background: copied ? '#2E7D32' : '#1565C0', color: 'white', fontWeight: 700,
                display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6,
              }}>
                {copied ? <><CheckCircle2 size={16} /> Copiado!</> : <><Copy size={16} /> Copiar Chave</>}
              </button>
              <button onClick={() => setNewKey(null)} style={{
                flex: 1, padding: 12, borderRadius: 10, border: '1.5px solid #E0E4F0', cursor: 'pointer', fontWeight: 700,
              }}>
                Fechar
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TAB: API Keys */}
      {tab === 'apikeys' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p style={{ fontSize: 13, color: '#757575' }}>Use API Keys para integrar com marketplaces, apps e ERPs</p>
            <button onClick={() => setShowKeyForm(true)} style={{
              display: 'flex', alignItems: 'center', gap: 6,
              background: '#1565C0', color: 'white', padding: '9px 16px', borderRadius: 10,
              border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 13,
            }}>
              <Plus size={14} /> Nova API Key
            </button>
          </div>

          {/* Docs da API */}
          <div style={{ background: '#E3F2FD', border: '1.5px solid #90CAF9', borderRadius: 12, padding: 14, marginBottom: 20 }}>
            <p style={{ fontSize: 13, fontWeight: 700, color: '#1565C0', marginBottom: 4 }}>📚 Como usar a API</p>
            <p style={{ fontSize: 12, color: '#424242', fontFamily: 'monospace' }}>
              Authorization: Bearer nxo_sua_chave_aqui
            </p>
            <p style={{ fontSize: 11, color: '#757575', marginTop: 4 }}>
              Base URL: https://seudominio.nexoone.com/api/v1 • Rate limit: 60 req/min por padrão
            </p>
          </div>

          {apiKeys.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
              <Key size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
              <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhuma API Key criada</p>
              <p style={{ fontSize: 13, marginTop: 4 }}>Crie uma chave para integrar com sistemas externos</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {apiKeys.map(k => (
                <div key={k.id} style={{ background: 'white', borderRadius: 12, padding: '14px 16px', border: `1.5px solid ${k.is_active ? '#E0E4F0' : '#EF9A9A'}`, display: 'flex', alignItems: 'center', gap: 12 }}>
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: k.is_active ? '#2E7D32' : '#B71C1C', flexShrink: 0 }} />
                  <div style={{ flex: 1 }}>
                    <p style={{ fontSize: 14, fontWeight: 600 }}>{k.name}</p>
                    <p style={{ fontSize: 12, color: '#9E9E9E', fontFamily: 'monospace' }}>{k.key_prefix}</p>
                    <div style={{ display: 'flex', gap: 4, marginTop: 6, flexWrap: 'wrap' }}>
                      {k.scopes.map(s => (
                        <span key={s} style={{ fontSize: 10, background: '#E3F2FD', color: '#1565C0', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>{s}</span>
                      ))}
                    </div>
                  </div>
                  <div style={{ textAlign: 'right', fontSize: 11, color: '#9E9E9E' }}>
                    <p>{k.rate_limit} req/min</p>
                    {k.last_used_at && <p>Usado: {new Date(k.last_used_at).toLocaleDateString('pt-BR')}</p>}
                  </div>
                  <button onClick={() => revokeKey(k.id)} style={{ background: '#FFEBEE', border: '1px solid #EF9A9A', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', color: '#B71C1C' }}>
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Form nova API Key */}
          {showKeyForm && (
            <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
              <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 480, maxHeight: '90vh', overflowY: 'auto' }}>
                <div className="flex items-center justify-between mb-4">
                  <h3 style={{ fontSize: 16, fontWeight: 700 }}>Nova API Key</h3>
                  <button onClick={() => setShowKeyForm(false)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
                </div>
                <form onSubmit={createAPIKey} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                  <div>
                    <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Nome da Integração</label>
                    <input required value={keyForm.name} onChange={e => setKeyForm(p => ({ ...p, name: e.target.value }))}
                      placeholder="Ex: Integração Shopify, App Mobile..." style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                  </div>
                  <div>
                    <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 8 }}>Permissões</label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
                      {SCOPES.map(s => (
                        <button key={s.key} type="button" onClick={() => toggleScope(s.key)} style={{
                          padding: '8px 10px', borderRadius: 8, border: '1.5px solid',
                          borderColor: keyForm.scopes.includes(s.key) ? '#1565C0' : '#E0E4F0',
                          background: keyForm.scopes.includes(s.key) ? '#E3F2FD' : 'white',
                          cursor: 'pointer', textAlign: 'left', fontSize: 12,
                        }}>
                          <p style={{ fontWeight: 600, color: keyForm.scopes.includes(s.key) ? '#1565C0' : '#424242' }}>{s.label}</p>
                          <p style={{ fontSize: 10, color: '#9E9E9E' }}>{s.desc}</p>
                        </button>
                      ))}
                    </div>
                  </div>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                    <div>
                      <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Rate Limit (req/min)</label>
                      <input type="number" value={keyForm.rate_limit} onChange={e => setKeyForm(p => ({ ...p, rate_limit: e.target.value }))}
                        style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                    </div>
                    <div>
                      <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Expira em (opcional)</label>
                      <input type="date" value={keyForm.expires_at} onChange={e => setKeyForm(p => ({ ...p, expires_at: e.target.value }))}
                        style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                    </div>
                  </div>
                  <button type="submit" style={{ width: '100%', padding: 13, borderRadius: 10, background: '#1565C0', color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer' }}>
                    🔑 Gerar API Key
                  </button>
                </form>
              </div>
            </div>
          )}
        </div>
      )}

      {/* TAB: Filiais */}
      {tab === 'subsidiaries' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p style={{ fontSize: 13, color: '#757575' }}>Gerencie suas filiais e veja resultados consolidados</p>
            <button onClick={() => setShowSubForm(true)} style={{
              display: 'flex', alignItems: 'center', gap: 6,
              background: '#2E7D32', color: 'white', padding: '9px 16px', borderRadius: 10,
              border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 13,
            }}>
              <Plus size={14} /> Nova Filial
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
            {subsidiaries.map(s => (
              <div key={s.id} style={{ background: 'white', borderRadius: 12, padding: 16, border: `1.5px solid ${s.is_headquarter ? '#FFC107' : '#E0E4F0'}` }}>
                <div className="flex items-center justify-between mb-2">
                  <span style={{ fontSize: 20 }}>🏢</span>
                  {s.is_headquarter && <span style={{ fontSize: 10, background: '#FFF8E1', color: '#E65100', padding: '2px 8px', borderRadius: 100, fontWeight: 700 }}>MATRIZ</span>}
                </div>
                <p style={{ fontSize: 14, fontWeight: 700 }}>{s.name}</p>
                <p style={{ fontSize: 12, color: '#757575', marginTop: 4 }}>{s.city} — {s.state}</p>
                {s.cnpj && <p style={{ fontSize: 11, color: '#9E9E9E', fontFamily: 'monospace' }}>{s.cnpj}</p>}
              </div>
            ))}
          </div>

          {showSubForm && (
            <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
              <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 400 }}>
                <div className="flex items-center justify-between mb-4">
                  <h3 style={{ fontSize: 16, fontWeight: 700 }}>Nova Filial</h3>
                  <button onClick={() => setShowSubForm(false)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
                </div>
                <form onSubmit={createSubsidiary} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {[
                    { label: 'Nome', key: 'name', placeholder: 'Ex: Filial Santo André' },
                    { label: 'CNPJ', key: 'cnpj', placeholder: '00.000.000/0000-00' },
                    { label: 'Cidade', key: 'city', placeholder: 'São Paulo' },
                    { label: 'Estado (UF)', key: 'state', placeholder: 'SP' },
                  ].map(f => (
                    <div key={f.key}>
                      <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>{f.label}</label>
                      <input required={f.key === 'name'} value={(subForm as any)[f.key]} placeholder={f.placeholder}
                        onChange={e => setSubForm(p => ({ ...p, [f.key]: e.target.value }))}
                        style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                    </div>
                  ))}
                  <button type="submit" style={{ width: '100%', padding: 13, borderRadius: 10, background: '#2E7D32', color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer', marginTop: 4 }}>
                    Criar Filial
                  </button>
                </form>
              </div>
            </div>
          )}
        </div>
      )}

      {/* TAB: Webhooks */}
      {tab === 'webhooks' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p style={{ fontSize: 13, color: '#757575' }}>Receba notificações em tempo real no seu sistema</p>
            <button onClick={() => setShowWebhookForm(true)} style={{
              display: 'flex', alignItems: 'center', gap: 6,
              background: '#6A1B9A', color: 'white', padding: '9px 16px', borderRadius: 10,
              border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 13,
            }}>
              <Plus size={14} /> Novo Webhook
            </button>
          </div>

          {webhooks.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
              <Globe size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
              <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhum webhook configurado</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {webhooks.map(w => (
                <div key={w.id} style={{ background: 'white', borderRadius: 12, padding: '14px 16px', border: '1.5px solid #E0E4F0' }}>
                  <p style={{ fontSize: 13, fontWeight: 600, fontFamily: 'monospace' }}>{w.url}</p>
                  <div style={{ display: 'flex', gap: 4, marginTop: 6, flexWrap: 'wrap' }}>
                    {w.events.map(e => (
                      <span key={e} style={{ fontSize: 10, background: '#F3E5F5', color: '#6A1B9A', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>{e}</span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}

          {showWebhookForm && (
            <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
              <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 480, maxHeight: '90vh', overflowY: 'auto' }}>
                <div className="flex items-center justify-between mb-4">
                  <h3 style={{ fontSize: 16, fontWeight: 700 }}>Novo Webhook</h3>
                  <button onClick={() => setShowWebhookForm(false)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
                </div>
                <form onSubmit={createWebhook} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                  <div>
                    <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>URL do Endpoint</label>
                    <input required type="url" value={webhookForm.url} onChange={e => setWebhookForm(p => ({ ...p, url: e.target.value }))}
                      placeholder="https://meusite.com/webhooks/nexo" style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                  </div>
                  <div>
                    <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 8 }}>Eventos</label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 4 }}>
                      {WEBHOOK_EVENTS.map(ev => (
                        <button key={ev} type="button" onClick={() => toggleEvent(ev)} style={{
                          padding: '6px 8px', borderRadius: 6, border: '1.5px solid',
                          borderColor: webhookForm.events.includes(ev) ? '#6A1B9A' : '#E0E4F0',
                          background: webhookForm.events.includes(ev) ? '#F3E5F5' : 'white',
                          cursor: 'pointer', textAlign: 'left', fontSize: 11,
                          color: webhookForm.events.includes(ev) ? '#6A1B9A' : '#424242', fontWeight: 600,
                        }}>
                          {ev}
                        </button>
                      ))}
                    </div>
                  </div>
                  <button type="submit" style={{ width: '100%', padding: 13, borderRadius: 10, background: '#6A1B9A', color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer' }}>
                    🔔 Criar Webhook
                  </button>
                </form>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
