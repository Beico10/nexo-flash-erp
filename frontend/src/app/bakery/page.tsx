'use client'
import { useState, useEffect } from 'react'
import { Wheat, ShoppingCart, AlertTriangle, Loader2, Scale } from 'lucide-react'

interface Product {
  ID: string; Name: string; SKU: string; SaleType: string; UnitPrice: number; NCMCode: string; IsBasketItem: boolean; CurrentStock: number; MinStock: number; ScaleCode: string
}

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }
const apiFetch = (path: string) => fetch(path, { headers: { Authorization: `Bearer ${getToken()}` } })

export default function BakeryPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!getToken()) { window.location.href = '/login'; return }
    apiFetch('/api/v1/bakery/products').then(r => { if (r.status === 401) { window.location.href = '/login'; return null }; return r.json() }).then(d => { if (d) { setProducts(d.data || []); setLoading(false) } }).catch(() => setLoading(false))
  }, [])

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const lowStock = products.filter(p => p.CurrentStock <= p.MinStock)

  if (loading) return <div className="flex items-center justify-center h-64"><Loader2 size={32} className="text-nexo-500 animate-spin" /></div>

  return (
    <div className="space-y-5 animate-fade-in" data-testid="bakery-page">
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-amber-50 rounded-xl flex items-center justify-center"><Wheat size={18} className="text-amber-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{products.length}</p><p className="text-xs text-slate-400">Produtos ativos</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center"><ShoppingCart size={18} className="text-emerald-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">{products.filter(p => p.IsBasketItem).length}</p><p className="text-xs text-slate-400">Cesta basica (IBS zero)</p></div>
        </div>
        <div className="card p-4 flex items-center gap-3">
          <div className="w-10 h-10 bg-red-50 rounded-xl flex items-center justify-center"><AlertTriangle size={18} className="text-red-500" /></div>
          <div><p className="text-lg font-bold text-slate-800">{lowStock.length}</p><p className="text-xs text-slate-400">Estoque baixo</p></div>
        </div>
      </div>

      {/* Products Table */}
      <div className="card overflow-hidden">
        <table className="w-full" data-testid="products-table">
          <thead className="bg-slate-50 border-b border-slate-100">
            <tr>
              <th className="table-header">SKU</th>
              <th className="table-header">Produto</th>
              <th className="table-header">Tipo Venda</th>
              <th className="table-header">Preco</th>
              <th className="table-header">NCM</th>
              <th className="table-header">Estoque</th>
              <th className="table-header">Balanca</th>
            </tr>
          </thead>
          <tbody>
            {products.map(p => (
              <tr key={p.ID} className="hover:bg-slate-50 transition-colors" data-testid={`product-${p.ID}`}>
                <td className="table-cell font-mono text-xs text-nexo-600">{p.SKU}</td>
                <td className="table-cell">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-slate-800">{p.Name}</span>
                    {p.IsBasketItem && <span className="text-[9px] font-bold px-1.5 py-0.5 rounded-full bg-emerald-100 text-emerald-700">CESTA</span>}
                  </div>
                </td>
                <td className="table-cell">
                  <span className={`text-xs font-medium px-2 py-1 rounded-lg ${p.SaleType === 'weight' ? 'bg-blue-50 text-blue-600' : 'bg-slate-50 text-slate-600'}`}>
                    {p.SaleType === 'weight' ? 'Peso (kg)' : 'Unidade'}
                  </span>
                </td>
                <td className="table-cell font-semibold text-slate-800">{fmt(p.UnitPrice)}{p.SaleType === 'weight' ? '/kg' : ''}</td>
                <td className="table-cell font-mono text-xs text-slate-500">{p.NCMCode}</td>
                <td className="table-cell">
                  <div className="flex items-center gap-2">
                    <span className={`font-semibold ${p.CurrentStock <= p.MinStock ? 'text-red-500' : 'text-slate-700'}`}>
                      {p.CurrentStock}{p.SaleType === 'weight' ? ' kg' : ' un'}
                    </span>
                    {p.CurrentStock <= p.MinStock && <AlertTriangle size={12} className="text-red-400" />}
                  </div>
                </td>
                <td className="table-cell">
                  <span className="flex items-center gap-1 text-xs text-slate-400"><Scale size={11} /> {p.ScaleCode}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
