'use client'
import { useState } from 'react'
import { Plus, TrendingUp, Package, ShoppingBag, Filter } from 'lucide-react'

const colors = ['Preto', 'Marrom', 'Nude', 'Branco', 'Vermelho']
const sizes  = ['33','34','35','36','37','38','39','40','41','42']

const generateGrid = (base: number) =>
  colors.reduce((acc, c) => ({
    ...acc,
    [c]: sizes.reduce((a, s) => ({
      ...a,
      [s]: { stock: Math.floor(Math.random() * 20), price: base + (Math.random() * 50 - 25) }
    }), {})
  }), {} as Record<string, Record<string, { stock: number; price: number }>>)

const models = [
  { id: '1', code: 'SANDAL001', name: 'Sandália Tiras',    brand: 'Própria', grid: generateGrid(189), basePrice: 189 },
  { id: '2', code: 'ANKLE001',  name: 'Ankle Boot Couro', brand: 'Importada', grid: generateGrid(349), basePrice: 349 },
]

const stockColor = (qty: number) => {
  if (qty === 0) return { bg: '#FFF1F1', text: '#C53030', border: 'rgba(197,48,48,0.2)' }
  if (qty <= 3)  return { bg: '#FFFBEB', text: '#B45309', border: 'rgba(180,83,9,0.2)' }
  return { bg: '#ECFDF5', text: '#047857', border: 'rgba(4,120,87,0.15)' }
}

export default function ShoesPage() {
  const [activeModel, setActiveModel] = useState(models[0])
  const [showModal, setShowModal] = useState(false)

  const totalStock = Object.values(activeModel.grid).flatMap(c => Object.values(c)).reduce((s, cell: any) => s + cell.stock, 0)
  const totalValue = Object.values(activeModel.grid).flatMap(c => Object.values(c)).reduce((s, cell: any) => s + (cell.stock * cell.price), 0)
  const skusEmpty = Object.values(activeModel.grid).flatMap(c => Object.values(c)).filter((cell: any) => cell.stock === 0).length

  return (
    <div className="space-y-5 animate-fade-in">

      {/* Model selector */}
      <div className="flex items-center gap-3">
        <div className="flex gap-2">
          {models.map(m => (
            <button key={m.id} onClick={() => setActiveModel(m)}
              className={activeModel.id === m.id ? 'btn-primary' : 'btn-ghost border'}
              style={activeModel.id !== m.id ? {border:'1px solid rgba(26,51,120,0.12)'} : {}}>
              <ShoppingBag size={13} />
              {m.name}
            </button>
          ))}
        </div>
        <button className="btn-ghost ml-2"><Filter size={13} />Filtros</button>
        <button onClick={() => setShowModal(true)} className="btn-primary ml-auto"><Plus size={14} />Novo Modelo</button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 stagger">
        {[
          { label: 'Total em estoque', value: `${totalStock} pares`, icon: <Package size={16} style={{color:'#1A47C8'}} />, bg: '#EEF3FF' },
          { label: 'Valor do estoque',  value: `R$ ${Math.round(totalValue).toLocaleString('pt-BR')}`, icon: <TrendingUp size={16} style={{color:'#047857'}} />, bg: '#ECFDF5' },
          { label: 'SKUs ativos',       value: `${colors.length * sizes.length}`, icon: <ShoppingBag size={16} style={{color:'#7C3AED'}} />, bg: '#F5F3FF' },
          { label: 'SKUs zerados',      value: `${skusEmpty}`, icon: <Package size={16} style={{color:'#C53030'}} />, bg: '#FFF1F1' },
        ].map(s => (
          <div key={s.label} className="card stat-card animate-slide-up">
            <div className="flex items-center justify-between">
              <p className="section-label">{s.label}</p>
              <div className="w-8 h-8 rounded-xl flex items-center justify-center" style={{background:s.bg}}>{s.icon}</div>
            </div>
            <p className="metric-lg">{s.value}</p>
          </div>
        ))}
      </div>

      {/* Grid matrix */}
      <div className="card overflow-hidden">
        <div style={{padding:'16px 20px',borderBottom:'1px solid rgba(26,51,120,0.07)',display:'flex',alignItems:'center',justifyContent:'space-between'}}>
          <div>
            <p style={{fontSize:14,fontWeight:700,color:'#0D1B4B'}}>{activeModel.name}</p>
            <p style={{fontSize:11,color:'#8892B8',marginTop:2}}>Código: {activeModel.code} · {activeModel.brand} · Preço base: R$ {activeModel.basePrice}</p>
          </div>
          <div className="flex items-center gap-3">
            {[{label:'Sem estoque',bg:'#FFF1F1',text:'#C53030'},{label:'Estoque baixo',bg:'#FFFBEB',text:'#B45309'},{label:'OK',bg:'#ECFDF5',text:'#047857'}].map(l => (
              <div key={l.label} className="flex items-center gap-1.5">
                <div style={{width:8,height:8,borderRadius:2,background:l.bg,border:`1px solid ${l.text}33`}} />
                <span style={{fontSize:10,color:'#8892B8',fontWeight:500}}>{l.label}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="overflow-x-auto">
          <table style={{width:'100%',borderCollapse:'collapse'}}>
            <thead>
              <tr style={{background:'#FAFBFF'}}>
                <th style={{padding:'10px 16px',textAlign:'left',fontSize:11,fontWeight:700,color:'#8892B8',textTransform:'uppercase',letterSpacing:'0.07em',borderBottom:'1px solid rgba(26,51,120,0.07)',width:100}}>Cor</th>
                {sizes.map(s => (
                  <th key={s} style={{padding:'10px 8px',textAlign:'center',fontSize:11,fontWeight:700,color:'#8892B8',textTransform:'uppercase',letterSpacing:'0.07em',borderBottom:'1px solid rgba(26,51,120,0.07)',minWidth:56}}>
                    {s}
                  </th>
                ))}
                <th style={{padding:'10px 16px',textAlign:'right',fontSize:11,fontWeight:700,color:'#8892B8',textTransform:'uppercase',letterSpacing:'0.07em',borderBottom:'1px solid rgba(26,51,120,0.07)'}}>Total</th>
              </tr>
            </thead>
            <tbody>
              {colors.map((color, ci) => {
                const rowTotal = sizes.reduce((s, sz) => s + (activeModel.grid[color]?.[sz]?.stock ?? 0), 0)
                return (
                  <tr key={color} style={{borderBottom:'1px solid rgba(26,51,120,0.04)'}} className="hover:bg-slate-50/50 transition-colors">
                    <td style={{padding:'10px 16px'}}>
                      <div className="flex items-center gap-2.5">
                        <div style={{width:12,height:12,borderRadius:3,background:
                          color==='Preto'?'#1a1a1a':color==='Marrom'?'#7c4a2a':color==='Nude'?'#d4a98a':color==='Branco'?'#f0ede8':'#c53030',
                          border:'1px solid rgba(0,0,0,0.1)'
                        }} />
                        <span style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}}>{color}</span>
                      </div>
                    </td>
                    {sizes.map(sz => {
                      const cell = activeModel.grid[color]?.[sz]
                      const stock = cell?.stock ?? 0
                      const sc = stockColor(stock)
                      return (
                        <td key={sz} style={{padding:'6px 4px',textAlign:'center'}}>
                          <div style={{display:'inline-flex',alignItems:'center',justifyContent:'center',width:40,height:32,borderRadius:8,background:sc.bg,border:`1px solid ${sc.border}`,cursor:'pointer',transition:'all 0.15s'}}
                            className="hover:scale-105">
                            <span style={{fontSize:12,fontWeight:700,color:sc.text,fontFamily:'var(--font-jetbrains)'}}>{stock}</span>
                          </div>
                        </td>
                      )
                    })}
                    <td style={{padding:'10px 16px',textAlign:'right'}}>
                      <span style={{fontSize:13,fontWeight:700,color:rowTotal > 0 ? '#0D1B4B' : '#C53030',fontFamily:'var(--font-jetbrains)'}}>{rowTotal}</span>
                    </td>
                  </tr>
                )
              })}
              {/* Column totals */}
              <tr style={{background:'#FAFBFF',borderTop:'2px solid rgba(26,51,120,0.1)'}}>
                <td style={{padding:'10px 16px'}}><span style={{fontSize:11,fontWeight:700,color:'#8892B8',textTransform:'uppercase',letterSpacing:'0.05em'}}>Total</span></td>
                {sizes.map(sz => {
                  const colTotal = colors.reduce((s, c) => s + (activeModel.grid[c]?.[sz]?.stock ?? 0), 0)
                  return (
                    <td key={sz} style={{padding:'10px 4px',textAlign:'center'}}>
                      <span style={{fontSize:12,fontWeight:700,color:'#1A47C8',fontFamily:'var(--font-jetbrains)'}}>{colTotal}</span>
                    </td>
                  )
                })}
                <td style={{padding:'10px 16px',textAlign:'right'}}>
                  <span style={{fontSize:13,fontWeight:800,color:'#0D1B4B',fontFamily:'var(--font-jetbrains)'}}>{totalStock}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
