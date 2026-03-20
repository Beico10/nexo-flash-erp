'use client'
import { useState, useEffect } from 'react'
import { Calculator, Search, ArrowRightLeft, ShieldCheck, TrendingUp, ChevronDown, Zap, ExternalLink, ShoppingBasket, Wine, ArrowDownLeft, ArrowUpRight } from 'lucide-react'

interface NCMItem {
  ncm_code: string
  description: string
  ibs_rate: number
  cbs_rate: number
  selective_rate: number
  is_basket: boolean
  basket_type?: string
}

interface TaxResult {
  ncm_code: string
  base_value: number
  ibs_rate: number
  cbs_rate: number
  selective_rate: number
  ibs_amount: number
  cbs_amount: number
  selective_amount: number
  total_tax: number
  cashback_amount: number
  is_basket_item: boolean
  legal_basis: string
  approval_status: string
}

export default function SimuladorFiscalPage() {
  const [ncmList, setNcmList] = useState<NCMItem[]>([])
  const [selectedNCM, setSelectedNCM] = useState('')
  const [baseValue, setBaseValue] = useState('1000')
  const [operation, setOperation] = useState<'debit_exit' | 'credit_entry'>('debit_exit')
  const [result, setResult] = useState<TaxResult | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [searchNCM, setSearchNCM] = useState('')
  const [showDropdown, setShowDropdown] = useState(false)

  useEffect(() => {
    fetch('/api/v1/tax/ncm-list')
      .then(r => r.json())
      .then(d => {
        setNcmList(d.data || [])
        if (d.data?.length > 0) setSelectedNCM(d.data[0].ncm_code)
      })
      .catch(() => {})
  }, [])

  async function simulate() {
    if (!selectedNCM || !baseValue) return
    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/v1/tax/simulate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ncm_code: selectedNCM,
          base_value: parseFloat(baseValue),
          operation,
        }),
      })
      const data = await res.json()
      if (!res.ok) {
        setError(data.error || 'Erro no calculo')
        setResult(null)
      } else {
        setResult(data)
      }
    } catch {
      setError('Falha de conexao com o servidor')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (selectedNCM && baseValue) simulate()
  }, [selectedNCM, baseValue, operation])

  const selected = ncmList.find(n => n.ncm_code === selectedNCM)
  const filteredNCM = ncmList.filter(n =>
    n.ncm_code.includes(searchNCM) || n.description.toLowerCase().includes(searchNCM.toLowerCase())
  )

  const fmt = (v: number) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
  const pct = (v: number) => `${(v * 100).toFixed(2)}%`

  return (
    <div className="min-h-screen" style={{ background: 'linear-gradient(135deg, #0B1433 0%, #0D1B4B 40%, #142460 100%)' }}>
      {/* Header */}
      <header className="px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg,#1A47C8 0%,#0F2D8A 100%)' }}>
            <Zap size={17} className="text-white" strokeWidth={2.5} />
          </div>
          <div>
            <span className="text-white font-extrabold text-lg tracking-tight">Nexo<span className="text-blue-400">One</span></span>
            <span className="text-blue-400/60 text-xs ml-2 font-semibold">SIMULADOR FISCAL</span>
          </div>
        </div>
        <a href="/login" className="text-sm text-blue-300 hover:text-white transition-colors flex items-center gap-1.5">
          Acessar ERP <ExternalLink size={13} />
        </a>
      </header>

      <div className="max-w-6xl mx-auto px-6 pb-16">
        {/* Hero */}
        <div className="text-center pt-8 pb-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full mb-5" style={{ background: 'rgba(26,71,200,0.2)', border: '1px solid rgba(26,71,200,0.3)' }}>
            <ShieldCheck size={13} className="text-blue-400" />
            <span className="text-xs font-semibold text-blue-300">Conforme LC 214/2025 — Reforma Tributaria 2026</span>
          </div>
          <h1 className="text-4xl sm:text-5xl font-extrabold text-white tracking-tight mb-3" data-testid="simulator-title">
            Simulador Fiscal <span className="text-blue-400">IBS/CBS</span>
          </h1>
          <p className="text-blue-200/70 text-base max-w-xl mx-auto">
            Calcule impostos da Reforma Tributaria 2026 em tempo real.
            Cesta Basica, Imposto Seletivo, Cashback e Transicao — tudo automatico.
          </p>
        </div>

        <div className="grid lg:grid-cols-5 gap-6">
          {/* Input Panel */}
          <div className="lg:col-span-2 space-y-4">
            <div className="rounded-2xl p-5 space-y-4" style={{ background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.08)', backdropFilter: 'blur(16px)' }}>
              <h3 className="text-sm font-bold text-white/80 uppercase tracking-wider flex items-center gap-2">
                <Calculator size={14} className="text-blue-400" /> Dados do Produto
              </h3>

              {/* NCM Selector */}
              <div className="relative">
                <label className="text-xs font-semibold text-blue-300/80 mb-1 block">Codigo NCM</label>
                <button
                  data-testid="ncm-selector"
                  onClick={() => setShowDropdown(!showDropdown)}
                  className="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl text-left transition-all"
                  style={{ background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.12)' }}
                >
                  <div>
                    <span className="text-white font-mono text-sm font-bold">{selectedNCM}</span>
                    {selected && <span className="text-blue-300/60 text-xs ml-2">{selected.description}</span>}
                  </div>
                  <ChevronDown size={14} className="text-blue-300/60" />
                </button>
                {showDropdown && (
                  <div className="absolute z-50 mt-1 w-full rounded-xl overflow-hidden shadow-2xl" style={{ background: '#141E45', border: '1px solid rgba(255,255,255,0.12)', maxHeight: 300 }}>
                    <div className="p-2">
                      <div className="flex items-center gap-2 px-3 py-2 rounded-lg" style={{ background: 'rgba(255,255,255,0.06)' }}>
                        <Search size={13} className="text-blue-300/60" />
                        <input
                          data-testid="ncm-search"
                          value={searchNCM}
                          onChange={e => setSearchNCM(e.target.value)}
                          placeholder="Buscar NCM..."
                          className="bg-transparent text-sm text-white outline-none flex-1 placeholder-blue-300/40"
                          autoFocus
                        />
                      </div>
                    </div>
                    <div className="overflow-y-auto" style={{ maxHeight: 230 }}>
                      {filteredNCM.map(n => (
                        <button
                          key={n.ncm_code}
                          data-testid={`ncm-option-${n.ncm_code}`}
                          onClick={() => { setSelectedNCM(n.ncm_code); setShowDropdown(false); setSearchNCM('') }}
                          className="w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-white/5 transition-colors"
                        >
                          <span className="text-white font-mono text-xs font-bold w-20">{n.ncm_code}</span>
                          <span className="text-blue-200/70 text-xs flex-1">{n.description}</span>
                          {n.is_basket && (
                            <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400">
                              {n.basket_type === 'zero' ? 'ZERO' : '-60%'}
                            </span>
                          )}
                          {n.selective_rate > 0 && (
                            <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-500/20 text-red-400">SELETIVO</span>
                          )}
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Value */}
              <div>
                <label className="text-xs font-semibold text-blue-300/80 mb-1 block">Valor Base (R$)</label>
                <input
                  data-testid="base-value-input"
                  type="number"
                  value={baseValue}
                  onChange={e => setBaseValue(e.target.value)}
                  className="w-full px-3.5 py-2.5 rounded-xl text-white text-lg font-bold outline-none transition-all"
                  style={{ background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.12)' }}
                  min="0"
                  step="0.01"
                />
              </div>

              {/* Operation toggle */}
              <div>
                <label className="text-xs font-semibold text-blue-300/80 mb-1.5 block">Tipo de Operacao</label>
                <div className="flex gap-2">
                  <button
                    data-testid="op-exit"
                    onClick={() => setOperation('debit_exit')}
                    className={`flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-xl text-xs font-bold transition-all ${
                      operation === 'debit_exit'
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/30'
                        : 'text-blue-300/60 hover:bg-white/5'
                    }`}
                    style={operation !== 'debit_exit' ? { border: '1px solid rgba(255,255,255,0.08)' } : {}}
                  >
                    <ArrowUpRight size={14} /> Venda (Saida)
                  </button>
                  <button
                    data-testid="op-entry"
                    onClick={() => setOperation('credit_entry')}
                    className={`flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-xl text-xs font-bold transition-all ${
                      operation === 'credit_entry'
                        ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-600/30'
                        : 'text-blue-300/60 hover:bg-white/5'
                    }`}
                    style={operation !== 'credit_entry' ? { border: '1px solid rgba(255,255,255,0.08)' } : {}}
                  >
                    <ArrowDownLeft size={14} /> Compra (Entrada)
                  </button>
                </div>
              </div>
            </div>

            {/* Legend */}
            <div className="rounded-2xl p-4 space-y-2" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.06)' }}>
              <h4 className="text-xs font-bold text-white/50 uppercase tracking-wider">Legenda</h4>
              <div className="grid grid-cols-2 gap-2 text-[11px]">
                <div className="flex items-center gap-2 text-blue-300/70"><div className="w-2 h-2 rounded-full bg-blue-500" /> IBS — Imposto sobre Bens e Servicos</div>
                <div className="flex items-center gap-2 text-blue-300/70"><div className="w-2 h-2 rounded-full bg-cyan-500" /> CBS — Contribuicao sobre Bens e Servicos</div>
                <div className="flex items-center gap-2 text-blue-300/70"><div className="w-2 h-2 rounded-full bg-red-500" /> IS — Imposto Seletivo</div>
                <div className="flex items-center gap-2 text-blue-300/70"><div className="w-2 h-2 rounded-full bg-emerald-500" /> Cashback Tributario</div>
              </div>
            </div>
          </div>

          {/* Results Panel */}
          <div className="lg:col-span-3 space-y-4">
            {error && (
              <div className="rounded-xl px-4 py-3 text-sm font-medium text-red-300" style={{ background: 'rgba(239,68,68,0.15)', border: '1px solid rgba(239,68,68,0.2)' }}>
                {error}
              </div>
            )}

            {result && (
              <>
                {/* Main Tax Card */}
                <div className="rounded-2xl overflow-hidden" style={{ background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.08)', backdropFilter: 'blur(16px)' }}>
                  {/* Header row */}
                  <div className="px-5 py-4 flex items-center justify-between" style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}>
                    <div>
                      <span className="text-white font-mono font-bold text-sm">{result.ncm_code}</span>
                      <span className="text-blue-300/50 text-xs ml-2">{selected?.description}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      {result.is_basket_item && (
                        <span className="flex items-center gap-1 text-[10px] font-bold px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-400">
                          <ShoppingBasket size={10} /> CESTA BASICA
                        </span>
                      )}
                      {result.selective_rate > 0 && (
                        <span className="flex items-center gap-1 text-[10px] font-bold px-2.5 py-1 rounded-full bg-red-500/20 text-red-400">
                          <Wine size={10} /> SELETIVO
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Tax breakdown */}
                  <div className="p-5 grid grid-cols-3 gap-4">
                    <TaxCard label="IBS" rate={pct(result.ibs_rate)} value={fmt(result.ibs_amount)} color="#3B82F6" />
                    <TaxCard label="CBS" rate={pct(result.cbs_rate)} value={fmt(result.cbs_amount)} color="#06B6D4" />
                    <TaxCard label="Seletivo" rate={pct(result.selective_rate)} value={fmt(result.selective_amount)} color="#EF4444" />
                  </div>

                  {/* Total row */}
                  <div className="px-5 py-4 flex items-center justify-between" style={{ background: 'rgba(255,255,255,0.03)', borderTop: '1px solid rgba(255,255,255,0.06)' }}>
                    <div>
                      <span className="text-blue-300/60 text-xs font-semibold uppercase">Total de Impostos</span>
                      <div className="flex items-center gap-3 mt-1">
                        <span className="text-white text-2xl font-extrabold" data-testid="total-tax">{fmt(result.total_tax)}</span>
                        <span className="text-blue-300/40 text-sm">sobre {fmt(result.base_value)}</span>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-blue-300/60 text-xs font-semibold uppercase">Carga Tributaria</span>
                      <div className="text-white text-2xl font-extrabold mt-1" data-testid="tax-percentage">
                        {result.base_value > 0 ? ((result.total_tax / result.base_value) * 100).toFixed(2) : '0.00'}%
                      </div>
                    </div>
                  </div>
                </div>

                {/* Cashback Card */}
                <div className="rounded-2xl p-5 flex items-center justify-between" style={{
                  background: result.cashback_amount > 0 ? 'rgba(16,185,129,0.1)' : 'rgba(239,68,68,0.08)',
                  border: `1px solid ${result.cashback_amount > 0 ? 'rgba(16,185,129,0.2)' : 'rgba(239,68,68,0.15)'}`,
                }}>
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${result.cashback_amount > 0 ? 'bg-emerald-500/20' : 'bg-red-500/15'}`}>
                      <ArrowRightLeft size={18} className={result.cashback_amount > 0 ? 'text-emerald-400' : 'text-red-400'} />
                    </div>
                    <div>
                      <span className="text-xs font-semibold text-white/60 uppercase">Cashback Tributario</span>
                      <p className="text-xs text-blue-300/50 mt-0.5">
                        {result.cashback_amount > 0 ? 'Credito a recuperar (compra)' : 'Debito a recolher (venda)'}
                      </p>
                    </div>
                  </div>
                  <span className={`text-xl font-extrabold ${result.cashback_amount > 0 ? 'text-emerald-400' : 'text-red-400'}`} data-testid="cashback-value">
                    {result.cashback_amount > 0 ? '+' : ''}{fmt(result.cashback_amount)}
                  </span>
                </div>

                {/* Legal Basis */}
                <div className="rounded-2xl p-4 flex items-start gap-3" style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.06)' }}>
                  <ShieldCheck size={16} className="text-blue-400 mt-0.5 flex-shrink-0" />
                  <div>
                    <span className="text-xs font-bold text-white/60 uppercase">Base Legal</span>
                    <p className="text-xs text-blue-300/60 mt-0.5" data-testid="legal-basis">{result.legal_basis}</p>
                  </div>
                </div>

                {/* Tax Bar Visualization */}
                <div className="rounded-2xl p-5" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.06)' }}>
                  <h4 className="text-xs font-bold text-white/50 uppercase tracking-wider mb-3 flex items-center gap-2">
                    <TrendingUp size={12} /> Composicao Tributaria
                  </h4>
                  {result.base_value > 0 && (
                    <div className="space-y-2">
                      <div className="h-8 rounded-lg overflow-hidden flex" style={{ background: 'rgba(255,255,255,0.05)' }}>
                        {result.ibs_amount > 0 && (
                          <div
                            className="h-full flex items-center justify-center text-[10px] font-bold text-white"
                            style={{ width: `${(result.ibs_amount / result.base_value) * 100}%`, background: '#3B82F6', minWidth: result.ibs_amount > 0 ? '40px' : 0 }}
                          >
                            IBS
                          </div>
                        )}
                        {result.cbs_amount > 0 && (
                          <div
                            className="h-full flex items-center justify-center text-[10px] font-bold text-white"
                            style={{ width: `${(result.cbs_amount / result.base_value) * 100}%`, background: '#06B6D4', minWidth: result.cbs_amount > 0 ? '40px' : 0 }}
                          >
                            CBS
                          </div>
                        )}
                        {result.selective_amount > 0 && (
                          <div
                            className="h-full flex items-center justify-center text-[10px] font-bold text-white"
                            style={{ width: `${(result.selective_amount / result.base_value) * 100}%`, background: '#EF4444', minWidth: '40px' }}
                          >
                            IS
                          </div>
                        )}
                        <div className="h-full flex-1 flex items-center justify-center text-[10px] font-bold text-white/30">
                          Valor Liquido
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </>
            )}

            {!result && !error && (
              <div className="rounded-2xl p-16 text-center" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.06)' }}>
                <Calculator size={48} className="text-blue-400/30 mx-auto mb-4" />
                <p className="text-blue-300/50 text-sm">Selecione um NCM e valor para calcular</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function TaxCard({ label, rate, value, color }: { label: string; rate: string; value: string; color: string }) {
  return (
    <div className="rounded-xl p-3" style={{ background: `${color}10`, border: `1px solid ${color}20` }}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-[10px] font-bold uppercase tracking-wider" style={{ color: `${color}CC` }}>{label}</span>
        <span className="text-[10px] font-bold" style={{ color: `${color}99` }}>{rate}</span>
      </div>
      <span className="text-lg font-extrabold text-white">{value}</span>
    </div>
  )
}
