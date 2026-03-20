'use client'

import { useState, useEffect, useCallback } from 'react'
import { Save, Check, Crown, Loader2, ToggleLeft, ToggleRight } from 'lucide-react'

interface Plan {
  id: string
  code: string
  name: string
  description: string
  price_monthly: number
  price_yearly: number
  setup_fee: number
  max_users: number | null
  max_transactions: number | null
  max_products: number | null
  max_invoices: number | null
  max_storage_mb: number | null
  features: Record<string, boolean>
  allowed_niches: string[]
  is_active: boolean
  is_featured: boolean
}

const NICHES = [
  { code: 'mechanic', label: 'Mecanica' },
  { code: 'bakery', label: 'Padaria' },
  { code: 'aesthetics', label: 'Estetica' },
  { code: 'logistics', label: 'Logistica' },
  { code: 'shoes', label: 'Calcados' },
  { code: 'industry', label: 'Industria' },
]

const FEATURES = [
  { key: 'fiscal_2026', label: 'Motor Fiscal 2026' },
  { key: 'baas_pix', label: 'PIX' },
  { key: 'baas_boleto', label: 'Boleto' },
  { key: 'baas_split', label: 'Split Payment' },
  { key: 'whatsapp', label: 'WhatsApp' },
  { key: 'ai_copilot', label: 'IA Co-Piloto' },
  { key: 'ai_concierge', label: 'IA Concierge' },
  { key: 'roteirizador', label: 'Roteirizador' },
  { key: 'multi_pdv', label: 'Multi-PDV' },
  { key: 'api_access', label: 'API Access' },
  { key: 'priority_support', label: 'Suporte Prioritario' },
  { key: 'custom_reports', label: 'Relatorios Custom' },
  { key: 'dedicated_support', label: 'Suporte Dedicado' },
  { key: 'sla_99_9', label: 'SLA 99.9%' },
]

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

export default function AdminPlansPage() {
  const [plans, setPlans] = useState<Plan[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState<string | null>(null)
  const [saved, setSaved] = useState<string | null>(null)

  const fetchPlans = useCallback(async () => {
    const token = getToken()
    if (!token) { window.location.href = '/login'; return }
    const res = await fetch('/api/v1/admin/plans', { headers: { Authorization: `Bearer ${token}` } })
    if (res.status === 401) { window.location.href = '/login'; return }
    const data = await res.json()
    setPlans(data.plans || [])
    setLoading(false)
  }, [])

  useEffect(() => { fetchPlans() }, [fetchPlans])

  const updateField = (code: string, field: string, value: any) => {
    setPlans(prev => prev.map(p => p.code === code ? { ...p, [field]: value } : p))
  }

  const toggleFeature = (code: string, featureKey: string) => {
    setPlans(prev => prev.map(p => {
      if (p.code !== code) return p
      return { ...p, features: { ...p.features, [featureKey]: !p.features[featureKey] } }
    }))
  }

  const toggleNiche = (code: string, niche: string) => {
    setPlans(prev => prev.map(p => {
      if (p.code !== code) return p
      const niches = p.allowed_niches.includes(niche)
        ? p.allowed_niches.filter(n => n !== niche)
        : [...p.allowed_niches, niche]
      return { ...p, allowed_niches: niches }
    }))
  }

  const savePlan = async (plan: Plan) => {
    setSaving(plan.code)
    const token = getToken()
    const res = await fetch('/api/v1/admin/plans', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(plan),
    })
    if (res.ok) {
      setSaved(plan.code)
      setTimeout(() => setSaved(null), 2000)
    }
    setSaving(null)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" />
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900" data-testid="admin-plans-title">Gestao de Planos</h1>
        <p className="text-sm text-gray-500 mt-1">Ajuste precos, limites, features e nichos de cada plano</p>
      </div>

      <div className="space-y-6">
        {plans.map(plan => (
          <div key={plan.code} data-testid={`plan-card-${plan.code}`} className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 bg-gray-50/50">
              <div className="flex items-center gap-3">
                {plan.is_featured && <Crown size={16} className="text-amber-500" />}
                <h2 className="text-lg font-bold text-gray-900">{plan.name}</h2>
                <span className="text-xs font-mono bg-gray-200 px-2 py-0.5 rounded">{plan.code}</span>
              </div>
              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <span className="text-gray-500">Ativo</span>
                  <button
                    onClick={() => updateField(plan.code, 'is_active', !plan.is_active)}
                    className={plan.is_active ? 'text-green-500' : 'text-gray-300'}
                  >
                    {plan.is_active ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </button>
                </label>
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <span className="text-gray-500">Destaque</span>
                  <button
                    onClick={() => updateField(plan.code, 'is_featured', !plan.is_featured)}
                    className={plan.is_featured ? 'text-amber-500' : 'text-gray-300'}
                  >
                    {plan.is_featured ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </button>
                </label>
                <button
                  data-testid={`save-plan-${plan.code}`}
                  onClick={() => savePlan(plan)}
                  disabled={saving === plan.code}
                  className="flex items-center gap-1.5 px-4 py-2 bg-blue-600 text-white text-sm font-semibold rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving === plan.code ? <Loader2 size={14} className="animate-spin" /> :
                   saved === plan.code ? <Check size={14} /> : <Save size={14} />}
                  {saved === plan.code ? 'Salvo!' : 'Salvar'}
                </button>
              </div>
            </div>

            <div className="p-6">
              {/* Nome e Descricao */}
              <div className="grid grid-cols-2 gap-4 mb-6">
                <div>
                  <label className="block text-xs font-semibold text-gray-500 mb-1.5">Nome do plano</label>
                  <input
                    data-testid={`plan-name-${plan.code}`}
                    type="text"
                    value={plan.name}
                    onChange={e => updateField(plan.code, 'name', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-gray-500 mb-1.5">Descricao</label>
                  <input
                    type="text"
                    value={plan.description}
                    onChange={e => updateField(plan.code, 'description', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              {/* Precos */}
              <div className="grid grid-cols-3 gap-4 mb-6">
                <div>
                  <label className="block text-xs font-semibold text-gray-500 mb-1.5">Preco Mensal (R$)</label>
                  <input
                    data-testid={`plan-price-monthly-${plan.code}`}
                    type="number"
                    value={plan.price_monthly}
                    onChange={e => updateField(plan.code, 'price_monthly', parseFloat(e.target.value) || 0)}
                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-gray-500 mb-1.5">Preco Anual (R$)</label>
                  <input
                    data-testid={`plan-price-yearly-${plan.code}`}
                    type="number"
                    value={plan.price_yearly}
                    onChange={e => updateField(plan.code, 'price_yearly', parseFloat(e.target.value) || 0)}
                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-gray-500 mb-1.5">Taxa de Setup (R$)</label>
                  <input
                    data-testid={`plan-setup-fee-${plan.code}`}
                    type="number"
                    value={plan.setup_fee}
                    onChange={e => updateField(plan.code, 'setup_fee', parseFloat(e.target.value) || 0)}
                    className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              {/* Limites */}
              <div className="mb-6">
                <h3 className="text-xs font-bold text-gray-400 uppercase tracking-wider mb-3">Limites</h3>
                <div className="grid grid-cols-5 gap-3">
                  {[
                    { key: 'max_users', label: 'Usuarios' },
                    { key: 'max_transactions', label: 'Transacoes/mes' },
                    { key: 'max_products', label: 'Produtos' },
                    { key: 'max_invoices', label: 'Notas/mes' },
                    { key: 'max_storage_mb', label: 'Storage (MB)' },
                  ].map(lim => (
                    <div key={lim.key}>
                      <label className="block text-xs text-gray-500 mb-1">{lim.label}</label>
                      <input
                        type="number"
                        value={(plan as any)[lim.key] ?? ''}
                        placeholder="Ilimitado"
                        onChange={e => {
                          const v = e.target.value === '' ? null : parseInt(e.target.value)
                          updateField(plan.code, lim.key, v)
                        }}
                        className="w-full px-2 py-1.5 border border-gray-200 rounded text-xs font-mono focus:ring-1 focus:ring-blue-500"
                      />
                    </div>
                  ))}
                </div>
              </div>

              {/* Nichos */}
              <div className="mb-6">
                <h3 className="text-xs font-bold text-gray-400 uppercase tracking-wider mb-3">Nichos Permitidos</h3>
                <div className="flex flex-wrap gap-2">
                  {NICHES.map(n => {
                    const active = plan.allowed_niches.includes(n.code)
                    return (
                      <button
                        key={n.code}
                        onClick={() => toggleNiche(plan.code, n.code)}
                        className={`px-3 py-1.5 text-xs font-medium rounded-full border transition-colors ${
                          active
                            ? 'bg-blue-50 border-blue-200 text-blue-700'
                            : 'bg-gray-50 border-gray-200 text-gray-400 hover:border-gray-300'
                        }`}
                      >
                        {n.label}
                      </button>
                    )
                  })}
                </div>
              </div>

              {/* Features */}
              <div>
                <h3 className="text-xs font-bold text-gray-400 uppercase tracking-wider mb-3">Features</h3>
                <div className="grid grid-cols-4 gap-2">
                  {FEATURES.map(f => {
                    const active = plan.features[f.key]
                    return (
                      <button
                        key={f.key}
                        onClick={() => toggleFeature(plan.code, f.key)}
                        className={`px-3 py-2 text-xs text-left rounded-lg border transition-colors ${
                          active
                            ? 'bg-green-50 border-green-200 text-green-700'
                            : 'bg-gray-50 border-gray-200 text-gray-400 hover:border-gray-300'
                        }`}
                      >
                        {active ? '+ ' : ''}{f.label}
                      </button>
                    )
                  })}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
