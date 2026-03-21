'use client'
import { useState, useEffect } from 'react'
import { Plus, Search, AlertTriangle, Package, TrendingDown, DollarSign, X, ArrowDown, ArrowUp, RotateCcw } from 'lucide-react'

interface Product {
  id: string; code: string; barcode: string; name: string
  category: string; unit: string; quantity: number
  min_quantity: number; cost_price: number; sale_price: number
  total_value: number; location: string; is_active: boolean
  is_low_stock: boolean; is_out_of_stock: boolean
  extra: Record<string, any>
}

interface Summary {
  total_products: number; total_value: number
  low_stock_count: number; out_of_stock_count: number
}

interface NichoConfig {
  name: string; icon: string; unit_default: string
  categories: string[]; extra_fields: string[]
  show_expiry: boolean; disabled?: boolean
}

const MOVEMENT_LABEL: Record<string, string> = {
  in: '↑ Entrada', out: '↓ Saída', adjust: '⚖️ Ajuste', loss: '🗑️ Perda'
}

const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })

export default function InventoryPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [summary, setSummary] = useState<Summary | null>(null)
  const [config, setConfig] = useState<NichoConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [filterCategory, setFilterCategory] = useState('')
  const [filterLow, setFilterLow] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [movModal, setMovModal] = useState<{product: Product, type: string} | null>(null)
  const [movQty, setMovQty] = useState('')
  const [movCost, setMovCost] = useState('')
  const [movNotes, setMovNotes] = useState('')
  const token = typeof window !== 'undefined' ? localStorage.getItem('nexo_token') || '' : ''

  const [form, setForm] = useState({
    code: '', name: '', category: '', unit: '',
    quantity: '0', min_quantity: '0', cost_price: '0', sale_price: '0',
    location: '', ncm: '', description: ''
  })

  useEffect(() => { fetchAll() }, [search, filterCategory, filterLow])

  const fetchAll = async () => {
    setLoading(true)
    const h = { Authorization: `Bearer ${token}` }
    try {
      const qs = new URLSearchParams()
      if (search) qs.set('search', search)
      if (filterCategory) qs.set('category', filterCategory)
      if (filterLow) qs.set('low_stock', 'true')

      const [cfgRes, listRes, sumRes] = await Promise.all([
        fetch('/api/v1/inventory/config', { headers: h }),
        fetch(`/api/v1/inventory/products?${qs}`, { headers: h }),
        fetch('/api/v1/inventory/products/summary', { headers: h }),
      ])
      if (cfgRes.ok) setConfig(await cfgRes.json())
      if (listRes.ok) setProducts((await listRes.json()).products || [])
      if (sumRes.ok) setSummary(await sumRes.json())
    } finally { setLoading(false) }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    const res = await fetch('/api/v1/inventory/products', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        ...form,
        quantity: parseFloat(form.quantity),
        min_quantity: parseFloat(form.min_quantity),
        cost_price: parseFloat(form.cost_price),
        sale_price: parseFloat(form.sale_price),
      }),
    })
    if (res.ok) { setShowForm(false); fetchAll() }
  }

  const handleMovement = async () => {
    if (!movModal) return
    const endpoint = movModal.type === 'adjust'
      ? `/api/v1/inventory/products/${movModal.product.id}/adjust`
      : movModal.type === 'in'
        ? `/api/v1/inventory/products/${movModal.product.id}/add`
        : `/api/v1/inventory/products/${movModal.product.id}/remove`

    const body = movModal.type === 'adjust'
      ? { new_quantity: parseFloat(movQty), notes: movNotes }
      : { quantity: parseFloat(movQty), unit_cost: parseFloat(movCost || '0'), notes: movNotes }

    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    })
    if (res.ok) { setMovModal(null); setMovQty(''); setMovCost(''); setMovNotes(''); fetchAll() }
  }

  if (config?.disabled) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: '#9E9E9E' }}>
        <Package size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
        <p style={{ fontSize: 16, fontWeight: 600 }}>Estoque não disponível para Logística</p>
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', maxWidth: 1100, margin: '0 auto' }}>

      {/* Header camaleão */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: '#212121' }}>
            {config?.icon} {config?.name || 'Estoque'}
          </h1>
          <p style={{ fontSize: 13, color: '#757575', marginTop: 2 }}>
            Custo Médio Ponderado automático • Alertas de mínimo via WhatsApp
          </p>
        </div>
        <button onClick={() => setShowForm(true)} style={{
          display: 'flex', alignItems: 'center', gap: 8,
          background: 'linear-gradient(135deg, #4A148C, #7B1FA2)',
          color: 'white', padding: '10px 20px', borderRadius: 10,
          border: 'none', cursor: 'pointer', fontWeight: 700, fontSize: 14,
        }}>
          <Plus size={16} /> Novo Produto
        </button>
      </div>

      {/* Cards resumo */}
      {summary && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14, marginBottom: 24 }}>
          {[
            { label: 'Produtos', value: summary.total_products, sub: 'cadastrados', color: '#1565C0', bg: '#E3F2FD', icon: <Package size={18} /> },
            { label: 'Valor Total', value: fmt(summary.total_value), sub: 'em estoque', color: '#2E7D32', bg: '#E8F5E9', icon: <DollarSign size={18} /> },
            { label: 'Estoque Baixo', value: summary.low_stock_count, sub: 'abaixo do mínimo', color: '#E65100', bg: '#FFF3E0', icon: <AlertTriangle size={18} /> },
            { label: 'Sem Estoque', value: summary.out_of_stock_count, sub: 'zerados', color: '#B71C1C', bg: '#FFEBEE', icon: <TrendingDown size={18} /> },
          ].map((c, i) => (
            <div key={i} style={{ background: c.bg, borderRadius: 14, padding: 16, border: `1.5px solid ${c.color}22` }}>
              <div className="flex items-center justify-between mb-2">
                <span style={{ fontSize: 11, fontWeight: 700, color: c.color, textTransform: 'uppercase' }}>{c.label}</span>
                <span style={{ color: c.color }}>{c.icon}</span>
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: c.color }}>{c.value}</div>
              <div style={{ fontSize: 11, color: '#757575', marginTop: 3 }}>{c.sub}</div>
            </div>
          ))}
        </div>
      )}

      {/* Filtros */}
      <div style={{ display: 'flex', gap: 10, marginBottom: 16, flexWrap: 'wrap' }}>
        <div style={{ position: 'relative', flex: 1, minWidth: 200 }}>
          <Search size={14} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: '#9E9E9E' }} />
          <input value={search} onChange={e => setSearch(e.target.value)}
            placeholder="Buscar produto..." style={{
              width: '100%', padding: '9px 12px 9px 34px',
              border: '1.5px solid #E0E4F0', borderRadius: 10, fontSize: 13, outline: 'none',
            }} />
        </div>
        <select value={filterCategory} onChange={e => setFilterCategory(e.target.value)} style={{
          padding: '9px 14px', border: '1.5px solid #E0E4F0', borderRadius: 10, fontSize: 13, outline: 'none',
        }}>
          <option value="">Todas categorias</option>
          {config?.categories.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
        <button onClick={() => setFilterLow(v => !v)} style={{
          padding: '9px 16px', borderRadius: 10, border: '1.5px solid',
          borderColor: filterLow ? '#E65100' : '#E0E4F0',
          background: filterLow ? '#FFF3E0' : 'white',
          color: filterLow ? '#E65100' : '#757575',
          fontSize: 12, fontWeight: 600, cursor: 'pointer',
        }}>
          ⚠️ Estoque Baixo
        </button>
      </div>

      {/* Lista */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#757575' }}>Carregando...</div>
      ) : products.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#BDBDBD' }}>
          <Package size={48} style={{ marginBottom: 12, opacity: 0.3 }} />
          <p style={{ fontSize: 15, fontWeight: 600 }}>Nenhum produto encontrado</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {products.map(p => (
            <div key={p.id} style={{
              background: 'white', borderRadius: 12, padding: '14px 16px',
              border: `1.5px solid ${p.is_out_of_stock ? '#EF9A9A' : p.is_low_stock ? '#FFCC02' : '#E0E4F0'}`,
              display: 'flex', alignItems: 'center', gap: 14,
            }}>
              {/* Indicador */}
              <div style={{
                width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
                background: p.is_out_of_stock ? '#B71C1C' : p.is_low_stock ? '#E65100' : '#2E7D32',
              }} />

              {/* Info */}
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#212121' }}>{p.name}</span>
                  {p.code && <span style={{ fontSize: 10, background: '#F5F7FF', color: '#5C6BC0', padding: '1px 6px', borderRadius: 100, fontWeight: 700 }}>{p.code}</span>}
                  {p.is_out_of_stock && <span style={{ fontSize: 10, background: '#FFEBEE', color: '#B71C1C', padding: '1px 8px', borderRadius: 100, fontWeight: 700 }}>SEM ESTOQUE</span>}
                  {p.is_low_stock && !p.is_out_of_stock && <span style={{ fontSize: 10, background: '#FFF3E0', color: '#E65100', padding: '1px 8px', borderRadius: 100, fontWeight: 700 }}>⚠️ BAIXO</span>}
                </div>
                <div style={{ display: 'flex', gap: 12, marginTop: 4, fontSize: 12, color: '#757575' }}>
                  <span>{p.category}</span>
                  {p.location && <span>📍 {p.location}</span>}
                  {p.extra?.application && <span>🚗 {p.extra.application}</span>}
                  {p.extra?.brand && <span>{p.extra.brand}</span>}
                </div>
              </div>

              {/* Quantidade */}
              <div style={{ textAlign: 'center', minWidth: 80 }}>
                <div style={{ fontSize: 20, fontWeight: 700, color: p.is_out_of_stock ? '#B71C1C' : '#212121' }}>
                  {p.quantity}
                </div>
                <div style={{ fontSize: 11, color: '#9E9E9E' }}>{p.unit} {p.min_quantity > 0 && `(mín: ${p.min_quantity})`}</div>
              </div>

              {/* CMP */}
              <div style={{ textAlign: 'right', minWidth: 100 }}>
                <div style={{ fontSize: 13, fontWeight: 700, color: '#212121' }}>{fmt(p.cost_price)}</div>
                <div style={{ fontSize: 11, color: '#9E9E9E' }}>CMP unitário</div>
                <div style={{ fontSize: 11, color: '#2E7D32', fontWeight: 600 }}>{fmt(p.total_value)} total</div>
              </div>

              {/* Ações */}
              <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                <button onClick={() => setMovModal({ product: p, type: 'in' })} title="Entrada" style={{ background: '#E8F5E9', border: '1px solid #A5D6A7', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>↑</button>
                <button onClick={() => setMovModal({ product: p, type: 'out' })} title="Saída" style={{ background: '#FFEBEE', border: '1px solid #EF9A9A', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>↓</button>
                <button onClick={() => setMovModal({ product: p, type: 'adjust' })} title="Ajuste" style={{ background: '#F5F7FF', border: '1px solid #C5CAE9', borderRadius: 8, padding: '6px 10px', cursor: 'pointer', fontSize: 16 }}>⚖️</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal movimentação */}
      {movModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100 }}>
          <div style={{ background: 'white', borderRadius: 16, padding: 28, width: 380 }}>
            <div className="flex items-center justify-between mb-4">
              <h3 style={{ fontSize: 16, fontWeight: 700 }}>{MOVEMENT_LABEL[movModal.type]}</h3>
              <button onClick={() => setMovModal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer' }}><X size={18} /></button>
            </div>
            <p style={{ fontSize: 13, color: '#757575', marginBottom: 16 }}>{movModal.product.name}</p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>
                  {movModal.type === 'adjust' ? 'Nova Quantidade' : 'Quantidade'}
                </label>
                <input type="number" value={movQty} onChange={e => setMovQty(e.target.value)}
                  placeholder="0" style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
              </div>
              {movModal.type === 'in' && (
                <div>
                  <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Custo Unitário (R$)</label>
                  <input type="number" value={movCost} onChange={e => setMovCost(e.target.value)}
                    placeholder="0.00" style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
                  {movCost && movQty && (
                    <p style={{ fontSize: 11, color: '#2E7D32', marginTop: 4 }}>
                      Novo CMP será recalculado automaticamente
                    </p>
                  )}
                </div>
              )}
              <div>
                <label style={{ fontSize: 11, fontWeight: 700, color: '#757575', textTransform: 'uppercase', display: 'block', marginBottom: 4 }}>Observação</label>
                <input type="text" value={movNotes} onChange={e => setMovNotes(e.target.value)}
                  placeholder="Ex: NF-e 4521, OS #1042..." style={{ width: '100%', padding: '10px 12px', border: '1.5px solid #E0E4F0', borderRadius: 8, fontSize: 14, outline: 'none' }} />
              </div>
              <button onClick={handleMovement} style={{
                width: '100%', padding: 13, borderRadius: 10,
                background: movModal.type === 'in' ? '#2E7D32' : movModal.type === 'out' ? '#B71C1C' : '#4A148C',
                color: 'white', fontWeight: 700, fontSize: 14, border: 'none', cursor: 'pointer', marginTop: 4,
              }}>
                Confirmar {MOVEMENT_LABEL[movModal.type]}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
