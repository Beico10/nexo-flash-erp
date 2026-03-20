'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { 
  Plus, Search, Filter, Download, Camera, Receipt,
  TrendingUp, Calendar, Building2, Tag
} from 'lucide-react'

interface Expense {
  id: string
  source: string
  supplier_name: string
  total_amount: number
  category: string
  issue_date: string
  status: string
  nfe_type: string
  ibs_credit: number
  cbs_credit: number
}

interface Summary {
  category: string
  total: number
  count: number
}

const categoryIcons: Record<string, string> = {
  mercadorias: '📦',
  materiais: '🧱',
  pecas: '⚙️',
  combustivel: '⛽',
  alimentacao: '🍽️',
  manutencao: '🔧',
  equipamentos: '🛠️',
  aluguel: '🏠',
  energia: '⚡',
  agua: '💧',
  telefone: '📱',
  servicos: '👥',
  outros: '📋',
}

const categoryColors: Record<string, string> = {
  mercadorias: 'bg-blue-100 text-blue-700',
  materiais: 'bg-green-100 text-green-700',
  pecas: 'bg-purple-100 text-purple-700',
  combustivel: 'bg-yellow-100 text-yellow-700',
  alimentacao: 'bg-red-100 text-red-700',
  manutencao: 'bg-indigo-100 text-indigo-700',
  equipamentos: 'bg-pink-100 text-pink-700',
  outros: 'bg-gray-100 text-gray-700',
}

export default function ExpensesPage() {
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [summary, setSummary] = useState<Summary[]>([])
  const [totals, setTotals] = useState({ amount: 0, ibs_credit: 0, cbs_credit: 0, tax_credit: 0 })
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState({ category: '', search: '' })

  useEffect(() => {
    fetchExpenses()
    fetchSummary()
  }, [filter.category])

  const fetchExpenses = async () => {
    try {
      const token = sessionStorage.getItem('access_token')
      if (!token) { window.location.href = '/login'; return }
      const params = new URLSearchParams()
      if (filter.category) params.append('category', filter.category)
      
      const res = await fetch(`/api/v1/expenses?${params}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.status === 401) { window.location.href = '/login'; return }
      const data = await res.json()
      setExpenses(data.expenses || [])
    } catch (error) {
      console.error('Erro ao carregar despesas:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchSummary = async () => {
    try {
      const token = sessionStorage.getItem('access_token')
      if (!token) return
      const res = await fetch('/api/v1/expenses/summary', {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await res.json()
      setSummary(data.summary || [])
      setTotals(data.totals || { amount: 0, ibs_credit: 0, cbs_credit: 0, tax_credit: 0 })
    } catch (error) {
      console.error('Erro ao carregar resumo:', error)
    }
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
    }).format(value)
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('pt-BR')
  }

  const filteredExpenses = expenses.filter(e => {
    if (filter.search) {
      const search = filter.search.toLowerCase()
      return e.supplier_name.toLowerCase().includes(search)
    }
    return true
  })

  return (
    <div className="min-h-screen bg-gray-50 py-6 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Despesas</h1>
            <p className="text-gray-600">Controle seus gastos e abata no imposto</p>
          </div>
          <Link
            href="/dashboard/expenses/scan"
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Camera className="w-5 h-5" />
            Escanear Nota
          </Link>
        </div>

        {/* Cards de resumo */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Receipt className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total Despesas</p>
                <p className="text-xl font-bold text-gray-900">{formatCurrency(totals.amount)}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 rounded-lg">
                <TrendingUp className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Crédito IBS</p>
                <p className="text-xl font-bold text-green-600">{formatCurrency(totals.ibs_credit)}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-100 rounded-lg">
                <TrendingUp className="w-5 h-5 text-emerald-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Crédito CBS</p>
                <p className="text-xl font-bold text-emerald-600">{formatCurrency(totals.cbs_credit)}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Tag className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total a Abater</p>
                <p className="text-xl font-bold text-purple-600">{formatCurrency(totals.tax_credit)}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Filtros */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-4 mb-6">
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  placeholder="Buscar por fornecedor..."
                  value={filter.search}
                  onChange={(e) => setFilter({ ...filter, search: e.target.value })}
                  className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
            
            <select
              value={filter.category}
              onChange={(e) => setFilter({ ...filter, category: e.target.value })}
              className="px-4 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Todas as categorias</option>
              <option value="pecas">Peças</option>
              <option value="mercadorias">Mercadorias</option>
              <option value="materiais">Materiais</option>
              <option value="combustivel">Combustível</option>
              <option value="servicos">Serviços</option>
              <option value="outros">Outros</option>
            </select>

            <button className="flex items-center gap-2 px-4 py-2 border border-gray-200 rounded-lg hover:bg-gray-50">
              <Download className="w-4 h-4" />
              Exportar
            </button>
          </div>
        </div>

        {/* Lista de despesas */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
          {loading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : filteredExpenses.length === 0 ? (
            <div className="p-8 text-center">
              <Receipt className="w-12 h-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500 mb-4">Nenhuma despesa registrada</p>
              <Link
                href="/dashboard/expenses/scan"
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                <Camera className="w-4 h-4" />
                Escanear primeira nota
              </Link>
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {filteredExpenses.map((expense) => (
                <div key={expense.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="text-2xl">
                        {categoryIcons[expense.category] || '📋'}
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900">{expense.supplier_name}</h3>
                        <div className="flex items-center gap-3 text-sm text-gray-500">
                          <span className="flex items-center gap-1">
                            <Calendar className="w-4 h-4" />
                            {formatDate(expense.issue_date)}
                          </span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${categoryColors[expense.category] || 'bg-gray-100 text-gray-700'}`}>
                            {expense.category}
                          </span>
                          {expense.source === 'qrcode' && (
                            <span className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded-full text-xs font-medium">
                              QR Code
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <p className="font-bold text-gray-900">{formatCurrency(expense.total_amount)}</p>
                      {(expense.ibs_credit + expense.cbs_credit) > 0 && (
                        <p className="text-sm text-green-600">
                          +{formatCurrency(expense.ibs_credit + expense.cbs_credit)} crédito
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Info footer */}
        <div className="mt-6 p-4 bg-green-50 rounded-xl border border-green-200">
          <h3 className="font-medium text-green-900 mb-2">💡 Dica: Não perca mais notas!</h3>
          <p className="text-sm text-green-700">
            Escaneie o QR Code das notas fiscais assim que receber. O sistema calcula automaticamente 
            o crédito de imposto (IBS/CBS) que você pode abater na Reforma Tributária 2026.
          </p>
        </div>
      </div>
    </div>
  )
}
