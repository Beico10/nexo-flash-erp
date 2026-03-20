'use client'
import { useState, useEffect } from 'react'
import { ShoppingBag, TrendingUp, Package, Award } from 'lucide-react'

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

interface Grid { id: string; model_code: string; model_name: string; brand: string; colors: string[]; sizes: string[]; total_skus: number; in_stock_skus: number; total_units: number; total_value: number }
interface Commission { seller_name: string; base_rate: number; monthly_target: number; total_sales: number; total_commission: number; met_achieved: boolean }

export default function ShoesPage() {
  const [grids, setGrids] = useState<Grid[]>([])
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = getToken()
    if (!token) { window.location.href = '/login'; return }
    const h = { Authorization: `Bearer ${token}` }
    Promise.all([
      fetch('/api/v1/shoes/grids', { headers: h }).then(r => r.json()),
      fetch('/api/v1/shoes/commissions', { headers: h }).then(r => r.json()),
    ]).then(([g, c]) => {
      setGrids(g.grids || [])
      setCommissions(c.commissions || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  if (loading) return <div className="flex items-center justify-center min-h-[400px]"><div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" /></div>

  const totalUnits = grids.reduce((s, g) => s + g.total_units, 0)
  const totalValue = grids.reduce((s, g) => s + g.total_value, 0)

  return (
    <div className="max-w-6xl mx-auto py-6 px-4 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900" data-testid="shoes-title">Calcados - Gestao de Grades</h1>
        <p className="text-sm text-gray-500">Colecoes, estoque por grade e comissoes</p>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Modelos</p><p className="text-2xl font-bold">{grids.length}</p></div>
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Total SKUs</p><p className="text-2xl font-bold">{grids.reduce((s, g) => s + g.total_skus, 0)}</p></div>
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Unidades</p><p className="text-2xl font-bold">{totalUnits}</p></div>
        <div className="bg-white rounded-xl border p-4"><p className="text-xs text-gray-500">Valor Estoque</p><p className="text-2xl font-bold text-green-600">R$ {totalValue.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}</p></div>
      </div>

      <div className="bg-white rounded-xl border">
        <div className="px-5 py-3 border-b"><h2 className="font-semibold text-gray-900">Grades de Produto</h2></div>
        <div className="divide-y">
          {grids.map(g => (
            <div key={g.id} data-testid={`grid-${g.id}`} className="px-5 py-4">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs bg-gray-100 px-2 py-0.5 rounded">{g.model_code}</span>
                    <span className="text-xs text-gray-400">{g.brand}</span>
                  </div>
                  <p className="font-medium text-gray-900 mt-1">{g.model_name}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm"><span className="font-bold">{g.in_stock_skus}</span>/<span className="text-gray-400">{g.total_skus}</span> SKUs em estoque</p>
                  <p className="text-xs text-gray-400">{g.total_units} unidades</p>
                </div>
              </div>
              <div className="flex gap-2 mt-2 flex-wrap">
                {g.colors.map(c => <span key={c} className="text-xs bg-gray-50 border rounded px-2 py-0.5">{c}</span>)}
                <span className="text-xs text-gray-300">|</span>
                <span className="text-xs text-gray-500">{g.sizes.join(', ')}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="bg-white rounded-xl border">
        <div className="px-5 py-3 border-b flex items-center gap-2"><Award size={16} /><h2 className="font-semibold text-gray-900">Comissoes - Vendedores</h2></div>
        <div className="divide-y">
          {commissions.map((c, i) => (
            <div key={i} className="px-5 py-4 flex items-center justify-between">
              <div>
                <p className="font-medium">{c.seller_name}</p>
                <p className="text-xs text-gray-400">Base {c.base_rate}% | Meta R$ {c.monthly_target.toLocaleString('pt-BR')}</p>
              </div>
              <div className="text-right">
                <p className="font-mono text-sm">R$ {c.total_sales.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}</p>
                <div className="flex items-center gap-2 justify-end">
                  {c.met_achieved && <TrendingUp size={12} className="text-green-500" />}
                  <span className={`text-xs font-bold ${c.met_achieved ? 'text-green-600' : 'text-gray-500'}`}>Comissao: R$ {c.total_commission.toFixed(2)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
