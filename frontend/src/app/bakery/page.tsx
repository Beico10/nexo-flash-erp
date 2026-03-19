'use client'
import { useState } from 'react'
import { Scale, Plus, Trash2, CheckCircle, CreditCard, Smartphone, Banknote } from 'lucide-react'

const products = [
  { id: '1', sku: 'PAO-001', name: 'Pão Francês',    plu: 'P001', type: 'weight', price: 8.50,  basket: true,  stock: 45.2 },
  { id: '2', sku: 'BOL-001', name: 'Bolo de Cenoura', plu: 'B001', type: 'unit',   price: 35.00, basket: false, stock: 8 },
  { id: '3', sku: 'CRO-001', name: 'Croissant',       plu: 'C001', type: 'unit',   price: 7.50,  basket: false, stock: 24 },
  { id: '4', sku: 'PAO-002', name: 'Pão de Queijo',   plu: 'P002', type: 'unit',   price: 4.00,  basket: true,  stock: 60 },
  { id: '5', sku: 'TOR-001', name: 'Torta Frango',    plu: 'T001', type: 'unit',   price: 12.00, basket: false, stock: 15 },
  { id: '6', sku: 'BRI-001', name: 'Brioche',         plu: 'B002', type: 'unit',   price: 9.00,  basket: false, stock: 18 },
]

interface CartItem { productId: string; name: string; quantity: number; unitPrice: number; isBasket: boolean }

export default function BakeryPage() {
  const [cart, setCart] = useState<CartItem[]>([])
  const [payment, setPayment] = useState<'pix' | 'cash' | 'credit'>('pix')
  const [scaleWeight, setScaleWeight] = useState<number | null>(null)
  const [success, setSuccess] = useState(false)

  const addToCart = (p: typeof products[0]) => {
    const qty = p.type === 'weight' ? (scaleWeight ?? 0.5) : 1
    setCart(c => {
      const existing = c.find(x => x.productId === p.id)
      if (existing) return c.map(x => x.productId === p.id ? { ...x, quantity: x.quantity + qty } : x)
      return [...c, { productId: p.id, name: p.name, quantity: qty, unitPrice: p.price, isBasket: p.basket }]
    })
  }

  const removeFromCart = (id: string) => setCart(c => c.filter(x => x.productId !== id))

  const subtotal = cart.reduce((s, i) => s + i.quantity * i.unitPrice, 0)

  const completeSale = () => {
    setSuccess(true)
    setCart([])
    setTimeout(() => setSuccess(false), 3000)
  }

  const simulateScale = () => {
    const w = parseFloat((Math.random() * 2 + 0.1).toFixed(3))
    setScaleWeight(w)
  }

  return (
    <div className="flex gap-5 h-[calc(100vh-10rem)] animate-fade-in">

      {/* Products grid */}
      <div className="flex-1 overflow-y-auto space-y-4">
        {/* Scale widget */}
        <div className="card p-4 flex items-center gap-4">
          <div className="w-10 h-10 bg-nexo-50 rounded-xl flex items-center justify-center">
            <Scale size={18} className="text-nexo-500" />
          </div>
          <div className="flex-1">
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wide">Balança</p>
            <p className="text-xl font-display font-700 text-slate-900">
              {scaleWeight ? `${scaleWeight.toFixed(3)} kg` : '— kg'}
            </p>
          </div>
          <button onClick={simulateScale} className="btn-secondary text-xs">
            Ler Balança
          </button>
          {scaleWeight && (
            <div className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 border border-emerald-200 rounded-xl">
              <div className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse" />
              <span className="text-xs font-semibold text-emerald-700">Peso estável</span>
            </div>
          )}
        </div>

        {/* Product cards */}
        <div className="grid grid-cols-3 gap-3">
          {products.map((p) => (
            <button
              key={p.id}
              onClick={() => addToCart(p)}
              className="card-hover p-4 text-left group"
            >
              <div className="flex items-start justify-between mb-2">
                <span className="font-mono text-[10px] text-slate-400">{p.plu}</span>
                {p.basket && (
                  <span className="text-[10px] font-bold px-1.5 py-0.5 bg-emerald-100 text-emerald-700 rounded-full">
                    Cesta Básica
                  </span>
                )}
              </div>
              <p className="font-semibold text-slate-800 text-sm">{p.name}</p>
              <p className="text-xs text-slate-400 mt-0.5">
                {p.type === 'weight' ? 'por kg' : 'unidade'} · estoque: {p.stock}
              </p>
              <p className="text-lg font-display font-700 text-nexo-600 mt-2">
                R$ {p.price.toLocaleString('pt-BR', {minimumFractionDigits: 2})}
              </p>
              <div className="mt-2 flex items-center gap-1 text-nexo-500 opacity-0 group-hover:opacity-100 transition-opacity">
                <Plus size={13} />
                <span className="text-xs font-medium">Adicionar</span>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Cart / Checkout */}
      <div className="w-80 flex flex-col gap-4 flex-shrink-0">
        <div className="card flex-1 flex flex-col overflow-hidden">
          <div className="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
            <h2 className="font-semibold text-slate-800 text-sm">Carrinho</h2>
            <span className="text-xs text-slate-400">{cart.length} iten{cart.length !== 1 ? 's' : ''}</span>
          </div>

          {/* Cart items */}
          <div className="flex-1 overflow-y-auto p-3 space-y-2">
            {cart.length === 0 && (
              <div className="py-8 text-center">
                <p className="text-sm text-slate-300">Nenhum item</p>
              </div>
            )}
            {cart.map((item) => (
              <div key={item.productId} className="flex items-center gap-2 p-2.5 bg-slate-50 rounded-xl">
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-semibold text-slate-700 truncate">{item.name}</p>
                  <p className="text-xs text-slate-400">{item.quantity.toFixed(3)} × R$ {item.unitPrice.toFixed(2)}</p>
                </div>
                <p className="text-sm font-bold text-slate-800 flex-shrink-0">
                  R$ {(item.quantity * item.unitPrice).toFixed(2)}
                </p>
                <button onClick={() => removeFromCart(item.productId)} className="text-red-400 hover:text-red-600">
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
          </div>

          {/* Total */}
          <div className="p-4 border-t border-slate-100 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-500">Subtotal</span>
              <span className="text-lg font-display font-700 text-slate-900">
                R$ {subtotal.toLocaleString('pt-BR', {minimumFractionDigits: 2})}
              </span>
            </div>

            {/* Payment method */}
            <div className="grid grid-cols-3 gap-1.5">
              {([['pix', 'PIX', Smartphone], ['cash', 'Dinheiro', Banknote], ['credit', 'Cartão', CreditCard]] as const).map(([key, label, Icon]) => (
                <button
                  key={key}
                  onClick={() => setPayment(key)}
                  className={`flex flex-col items-center gap-1 py-2 rounded-xl border text-xs font-medium transition-all ${
                    payment === key
                      ? 'bg-nexo-500 text-white border-nexo-500 shadow-nexo-sm'
                      : 'bg-white text-slate-500 border-slate-200 hover:border-nexo-300'
                  }`}
                >
                  <Icon size={14} />
                  {label}
                </button>
              ))}
            </div>

            <button
              onClick={completeSale}
              disabled={cart.length === 0}
              className="btn-primary w-full justify-center py-3 text-base"
            >
              <CheckCircle size={16} />
              Finalizar Venda
            </button>
          </div>
        </div>

        {/* Success message */}
        {success && (
          <div className="card p-4 bg-emerald-50 border border-emerald-200 flex items-center gap-3 animate-slide-up">
            <CheckCircle size={20} className="text-emerald-500" />
            <div>
              <p className="text-sm font-semibold text-emerald-700">Venda finalizada!</p>
              <p className="text-xs text-emerald-600">IBS/CBS calculados automaticamente</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
