'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { 
  CreditCard, Calendar, TrendingUp, AlertTriangle, 
  Check, ArrowUpRight, Zap, Clock
} from 'lucide-react'

interface UsageStatus {
  metric: string
  current: number
  limit: number | null
  percent: number
  is_at_limit: boolean
}

interface Subscription {
  id: string
  plan_code: string
  plan_name: string
  status: string
  trial_ends_at: string | null
  current_period_end: string
  billing_cycle: string
  price: number
  discount_percent: number
  discount_reason: string
}

const metricLabels: Record<string, string> = {
  users: 'Usuários',
  transactions: 'Transações',
  products: 'Produtos',
  invoices: 'Notas Fiscais',
}

const statusLabels: Record<string, { label: string; color: string }> = {
  trialing: { label: 'Em teste', color: 'bg-blue-100 text-blue-800' },
  active: { label: 'Ativo', color: 'bg-green-100 text-green-800' },
  past_due: { label: 'Pagamento pendente', color: 'bg-yellow-100 text-yellow-800' },
  cancelled: { label: 'Cancelado', color: 'bg-red-100 text-red-800' },
  expired: { label: 'Expirado', color: 'bg-gray-100 text-gray-800' },
}

export default function SubscriptionPage() {
  const router = useRouter()
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [usage, setUsage] = useState<UsageStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [showUpgradeModal, setShowUpgradeModal] = useState(false)

  useEffect(() => {
    fetchSubscription()
  }, [])

  const fetchSubscription = async () => {
    try {
      const token = sessionStorage.getItem('access_token')
      if (!token) { window.location.href = '/login'; return }
      const res = await fetch('/api/v1/billing/subscription', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.status === 401) { window.location.href = '/login'; return }
      const data = await res.json()
      setSubscription(data.subscription)
      setUsage(data.usage || [])
    } catch (error) {
      console.error('Erro ao carregar assinatura:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
    })
  }

  const formatPrice = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value)
  }

  const getDaysRemaining = (dateStr: string) => {
    const diff = new Date(dateStr).getTime() - Date.now()
    return Math.ceil(diff / (1000 * 60 * 60 * 24))
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!subscription) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600 mb-4">Nenhuma assinatura encontrada</p>
          <button
            onClick={() => router.push('/pricing')}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Ver Planos
          </button>
        </div>
      </div>
    )
  }

  const isTrialing = subscription.status === 'trialing'
  const trialDaysRemaining = subscription.trial_ends_at 
    ? getDaysRemaining(subscription.trial_ends_at)
    : 0

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-2xl font-bold text-gray-900 mb-8">Minha Assinatura</h1>

        {/* Trial Alert */}
        {isTrialing && trialDaysRemaining > 0 && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6 flex items-start gap-3">
            <Clock className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div>
              <p className="font-medium text-blue-900">
                Período de teste: {trialDaysRemaining} dias restantes
              </p>
              <p className="text-sm text-blue-700 mt-1">
                Aproveite para explorar todos os recursos. Depois, escolha o plano ideal para você.
              </p>
              <button
                onClick={() => setShowUpgradeModal(true)}
                className="mt-3 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700"
              >
                Ativar plano agora
              </button>
            </div>
          </div>
        )}

        {/* Trial Expiring Alert */}
        {isTrialing && trialDaysRemaining <= 2 && trialDaysRemaining > 0 && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6 flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
            <div>
              <p className="font-medium text-yellow-900">
                Seu teste expira em {trialDaysRemaining} dia{trialDaysRemaining > 1 ? 's' : ''}!
              </p>
              <p className="text-sm text-yellow-700 mt-1">
                Ative agora para não perder seus dados.
              </p>
            </div>
          </div>
        )}

        {/* Subscription Card */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-6">
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-4">
                <div className="p-3 bg-blue-100 rounded-lg">
                  <Zap className="w-6 h-6 text-blue-600" />
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-900">{subscription.plan_name}</h2>
                  <span className={`inline-block px-2 py-1 text-xs font-medium rounded-full ${
                    statusLabels[subscription.status]?.color || 'bg-gray-100'
                  }`}>
                    {statusLabels[subscription.status]?.label || subscription.status}
                  </span>
                </div>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold text-gray-900">{formatPrice(subscription.price)}</p>
                <p className="text-sm text-gray-500">
                  por {subscription.billing_cycle === 'yearly' ? 'ano' : 'mês'}
                </p>
                {subscription.discount_percent > 0 && (
                  <p className="text-sm text-green-600">
                    {subscription.discount_percent}% OFF - {subscription.discount_reason}
                  </p>
                )}
              </div>
            </div>

            {/* Billing Info */}
            <div className="grid grid-cols-2 gap-4 pt-6 border-t border-gray-100">
              <div className="flex items-center gap-3">
                <Calendar className="w-5 h-5 text-gray-400" />
                <div>
                  <p className="text-sm text-gray-500">Próxima cobrança</p>
                  <p className="font-medium text-gray-900">
                    {formatDate(subscription.current_period_end)}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <CreditCard className="w-5 h-5 text-gray-400" />
                <div>
                  <p className="text-sm text-gray-500">Ciclo</p>
                  <p className="font-medium text-gray-900">
                    {subscription.billing_cycle === 'yearly' ? 'Anual' : 'Mensal'}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="bg-gray-50 px-6 py-4 flex gap-3">
            <button
              onClick={() => router.push('/pricing')}
              className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
            >
              <TrendingUp className="w-4 h-4" />
              Mudar plano
            </button>
            <button className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">
              <CreditCard className="w-4 h-4" />
              Atualizar pagamento
            </button>
          </div>
        </div>

        {/* Usage */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-6">Uso do Plano</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {usage.map((item) => (
              <div key={item.metric} className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">{metricLabels[item.metric] || item.metric}</span>
                  <span className="font-medium text-gray-900">
                    {item.current.toLocaleString()} / {item.limit ? item.limit.toLocaleString() : '∞'}
                  </span>
                </div>
                <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${
                      item.is_at_limit
                        ? 'bg-red-500'
                        : item.percent >= 80
                        ? 'bg-yellow-500'
                        : 'bg-blue-500'
                    }`}
                    style={{ width: `${Math.min(item.percent, 100)}%` }}
                  />
                </div>
                {item.is_at_limit && (
                  <p className="text-xs text-red-600 flex items-center gap-1">
                    <AlertTriangle className="w-3 h-3" />
                    Limite atingido - faça upgrade para continuar
                  </p>
                )}
                {item.percent >= 80 && !item.is_at_limit && (
                  <p className="text-xs text-yellow-600">
                    Você está usando {item.percent}% do limite
                  </p>
                )}
              </div>
            ))}
          </div>

          {usage.some(u => u.percent >= 80) && (
            <div className="mt-6 pt-6 border-t border-gray-100">
              <button
                onClick={() => router.push('/pricing')}
                className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
              >
                <ArrowUpRight className="w-4 h-4" />
                Ver planos com mais recursos
              </button>
            </div>
          )}
        </div>

        {/* Cancel Link */}
        <div className="mt-8 text-center">
          <button className="text-sm text-gray-500 hover:text-gray-700">
            Cancelar assinatura
          </button>
        </div>
      </div>
    </div>
  )
}
