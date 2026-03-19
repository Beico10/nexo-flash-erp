'use client'

import { useState, useEffect } from 'react'
import { Check, X, Zap, Building2, Rocket, Crown, Star } from 'lucide-react'

interface PlanFeatures {
  fiscal_2026: boolean
  baas_pix: boolean
  baas_boleto: boolean
  baas_split: boolean
  whatsapp: boolean
  ai_copilot: boolean
  ai_concierge: boolean
  roteirizador: boolean
  multi_pdv: boolean
  api_access: boolean
  priority_support: boolean
  custom_reports: boolean
}

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
  features: PlanFeatures
  is_featured: boolean
}

const planIcons: Record<string, React.ReactNode> = {
  micro: <Zap className="w-6 h-6" />,
  starter: <Star className="w-6 h-6" />,
  pro: <Rocket className="w-6 h-6" />,
  business: <Building2 className="w-6 h-6" />,
  enterprise: <Crown className="w-6 h-6" />,
}

const featureLabels: Record<string, string> = {
  fiscal_2026: 'Motor Fiscal IBS/CBS 2026',
  baas_pix: 'Recebimento via PIX',
  baas_boleto: 'Emissão de Boletos',
  baas_split: 'Split de Pagamento',
  whatsapp: 'Notificações WhatsApp',
  ai_copilot: 'IA Co-Piloto',
  ai_concierge: 'IA Concierge (Onboarding)',
  roteirizador: 'Roteirizador Inteligente',
  multi_pdv: 'Múltiplos PDVs',
  api_access: 'Acesso à API',
  priority_support: 'Suporte Prioritário',
  custom_reports: 'Relatórios Personalizados',
}

export default function PricingPage() {
  const [plans, setPlans] = useState<Plan[]>([])
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly')
  const [couponCode, setCouponCode] = useState('')
  const [couponValid, setCouponValid] = useState<boolean | null>(null)
  const [couponDiscount, setCouponDiscount] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchPlans()
  }, [])

  const fetchPlans = async () => {
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/billing/plans`)
      const data = await res.json()
      setPlans(data.plans || [])
    } catch (error) {
      console.error('Erro ao carregar planos:', error)
    } finally {
      setLoading(false)
    }
  }

  const validateCoupon = async () => {
    if (!couponCode) return
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/billing/coupon/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: couponCode, plan_code: 'pro' }),
      })
      const data = await res.json()
      setCouponValid(data.valid)
      if (data.valid) {
        setCouponDiscount(data.discount_value)
      }
    } catch {
      setCouponValid(false)
    }
  }

  const getPrice = (plan: Plan) => {
    const basePrice = billingCycle === 'yearly' ? plan.price_yearly : plan.price_monthly
    if (couponValid && couponDiscount > 0) {
      return basePrice * (1 - couponDiscount / 100)
    }
    return basePrice
  }

  const formatPrice = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value)
  }

  const formatLimit = (value: number | null) => {
    if (value === null) return 'Ilimitado'
    return value.toLocaleString('pt-BR')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-50 to-white py-16 px-4">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Planos que crescem com você
          </h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Do autônomo à indústria — mesmo sistema, mesmo código, mesmo suporte.
            <br />
            <strong className="text-blue-600">7 dias grátis</strong> para testar qualquer plano.
          </p>
        </div>

        {/* Billing Toggle */}
        <div className="flex justify-center mb-8">
          <div className="bg-gray-100 p-1 rounded-lg inline-flex">
            <button
              onClick={() => setBillingCycle('monthly')}
              className={`px-6 py-2 rounded-md text-sm font-medium transition-all ${
                billingCycle === 'monthly'
                  ? 'bg-white text-gray-900 shadow'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Mensal
            </button>
            <button
              onClick={() => setBillingCycle('yearly')}
              className={`px-6 py-2 rounded-md text-sm font-medium transition-all ${
                billingCycle === 'yearly'
                  ? 'bg-white text-gray-900 shadow'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Anual <span className="text-green-600 ml-1">-17%</span>
            </button>
          </div>
        </div>

        {/* Coupon Input */}
        <div className="flex justify-center mb-12">
          <div className="flex items-center gap-2">
            <input
              type="text"
              placeholder="Cupom de desconto"
              value={couponCode}
              onChange={(e) => setCouponCode(e.target.value.toUpperCase())}
              className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              onClick={validateCoupon}
              className="px-4 py-2 bg-gray-800 text-white rounded-lg hover:bg-gray-700 transition-colors"
            >
              Aplicar
            </button>
            {couponValid === true && (
              <span className="text-green-600 font-medium">✓ {couponDiscount}% OFF</span>
            )}
            {couponValid === false && (
              <span className="text-red-600 font-medium">Cupom inválido</span>
            )}
          </div>
        </div>

        {/* Plans Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
          {plans.map((plan) => (
            <div
              key={plan.id}
              data-testid={`plan-card-${plan.code}`}
              className={`relative bg-white rounded-2xl shadow-lg overflow-hidden transition-transform hover:scale-105 ${
                plan.is_featured ? 'ring-2 ring-blue-600 scale-105' : ''
              }`}
            >
              {/* Featured Badge */}
              {plan.is_featured && (
                <div className="absolute top-0 right-0 bg-blue-600 text-white px-3 py-1 text-xs font-bold rounded-bl-lg">
                  MAIS POPULAR
                </div>
              )}

              <div className="p-6">
                {/* Plan Header */}
                <div className="flex items-center gap-3 mb-4">
                  <div className={`p-2 rounded-lg ${plan.is_featured ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-600'}`}>
                    {planIcons[plan.code]}
                  </div>
                  <div>
                    <h3 className="text-xl font-bold text-gray-900">{plan.name}</h3>
                    <p className="text-sm text-gray-500">{plan.description}</p>
                  </div>
                </div>

                {/* Price */}
                <div className="mb-6">
                  <div className="flex items-baseline gap-1">
                    <span className="text-4xl font-bold text-gray-900">
                      {formatPrice(getPrice(plan))}
                    </span>
                    <span className="text-gray-500">/{billingCycle === 'yearly' ? 'ano' : 'mês'}</span>
                  </div>
                  {couponValid && couponDiscount > 0 && (
                    <p className="text-sm text-green-600 mt-1">
                      De {formatPrice(billingCycle === 'yearly' ? plan.price_yearly : plan.price_monthly)}
                    </p>
                  )}
                  {plan.setup_fee > 0 && (
                    <p className="text-sm text-gray-500 mt-1">
                      + {formatPrice(plan.setup_fee)} setup
                    </p>
                  )}
                </div>

                {/* CTA Button */}
                <button
                  data-testid={`plan-cta-${plan.code}`}
                  className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
                    plan.is_featured
                      ? 'bg-blue-600 text-white hover:bg-blue-700'
                      : 'bg-gray-100 text-gray-900 hover:bg-gray-200'
                  }`}
                >
                  Começar 7 dias grátis
                </button>

                {/* Limits */}
                <div className="mt-6 pt-6 border-t border-gray-100">
                  <p className="text-sm font-medium text-gray-900 mb-3">Limites:</p>
                  <ul className="space-y-2 text-sm text-gray-600">
                    <li className="flex justify-between">
                      <span>Usuários</span>
                      <span className="font-medium">{formatLimit(plan.max_users)}</span>
                    </li>
                    <li className="flex justify-between">
                      <span>Transações/mês</span>
                      <span className="font-medium">{formatLimit(plan.max_transactions)}</span>
                    </li>
                    <li className="flex justify-between">
                      <span>Produtos</span>
                      <span className="font-medium">{formatLimit(plan.max_products)}</span>
                    </li>
                    <li className="flex justify-between">
                      <span>Notas/mês</span>
                      <span className="font-medium">{formatLimit(plan.max_invoices)}</span>
                    </li>
                  </ul>
                </div>

                {/* Features */}
                <div className="mt-6 pt-6 border-t border-gray-100">
                  <p className="text-sm font-medium text-gray-900 mb-3">Recursos:</p>
                  <ul className="space-y-2">
                    {Object.entries(plan.features).map(([key, enabled]) => (
                      <li key={key} className="flex items-center gap-2 text-sm">
                        {enabled ? (
                          <Check className="w-4 h-4 text-green-500 flex-shrink-0" />
                        ) : (
                          <X className="w-4 h-4 text-gray-300 flex-shrink-0" />
                        )}
                        <span className={enabled ? 'text-gray-700' : 'text-gray-400'}>
                          {featureLabels[key] || key}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* FAQ / Trust */}
        <div className="mt-16 text-center">
          <p className="text-gray-600">
            Sem fidelidade. Cancele quando quiser. <br />
            Upgrade ou downgrade em 1 clique, sem burocracia.
          </p>
          <p className="mt-4 text-sm text-gray-500">
            Dúvidas? Fale com nosso suporte via WhatsApp.
          </p>
        </div>
      </div>
    </div>
  )
}
