'use client'
import { useState } from 'react'
import { Factory, Layers, Package, Plus, ChevronRight, AlertTriangle, CheckCircle, Clock } from 'lucide-react'

const bom = [
  { id:'1', component: 'Tecido Oxford 100% Algodão', unit: 'metros', qty: 2.5, scrap: 0.05, netQty: 2.625, unitCost: 18.5, totalCost: 48.56, ncm: '52081900', available: 145.2 },
  { id:'2', component: 'Linha de costura nº 60',      unit: 'metros', qty: 45,  scrap: 0.02, netQty: 45.9,  unitCost: 0.08, totalCost: 3.67,  ncm: '54011090', available: 3200 },
  { id:'3', component: 'Botão de madre-pérola',        unit: 'unid',  qty: 8,   scrap: 0.03, netQty: 8.24,  unitCost: 0.45, totalCost: 3.71,  ncm: '96062900', available: 840 },
  { id:'4', component: 'Entretela termocolante',        unit: 'metros', qty: 0.3, scrap: 0.08, netQty: 0.324, unitCost: 12.0, totalCost: 3.89,  ncm: '56031290', available: 22.5 },
  { id:'5', component: 'Etiqueta bordada',              unit: 'unid',  qty: 1,   scrap: 0,    netQty: 1,     unitCost: 1.20, totalCost: 1.20,  ncm: '58071000', available: 0 },
]

const orders = [
  { id:'OP-2026-0042', product: 'Camisa Social Slim', qty: 50, status: 'in_progress', start: '2026-03-18', end: '2026-03-22', produced: 32 },
  { id:'OP-2026-0041', product: 'Blusa Feminina Básica', qty: 100, status: 'released',  start: '2026-03-19', end: '2026-03-25', produced: 0 },
  { id:'OP-2026-0040', product: 'Calça Chino Masculina', qty: 30, status: 'done',       start: '2026-03-15', end: '2026-03-17', produced: 30 },
]

const statusConfig: Record<string, { label: string; cls: string; icon: React.ReactNode }> = {
  planned:     { label: 'Planejada',    cls: 'badge-open',    icon: <Clock size={10} /> },
  released:    { label: 'Liberada',     cls: 'badge-pending', icon: <ChevronRight size={10} /> },
  in_progress: { label: 'Em produção',  cls: 'badge-approved',icon: <Factory size={10} /> },
  done:        { label: 'Concluída',    cls: 'badge-done',    icon: <CheckCircle size={10} /> },
}

export default function IndustryPage() {
  const [activeTab, setActiveTab] = useState<'pcp' | 'bom'>('pcp')
  const totalCost = bom.reduce((s, b) => s + b.totalCost, 0)
  const missingItems = bom.filter(b => b.available < b.netQty * 50)

  return (
    <div className="space-y-5 animate-fade-in">

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 stagger">
        {[
          { label: 'OPs abertas',      value: '2',               icon: <Factory size={16} style={{color:'#1A47C8'}} />, bg:'#EEF3FF' },
          { label: 'Custo da BOM',     value: `R$ ${totalCost.toFixed(2)}`, icon: <Layers size={16} style={{color:'#7C3AED'}} />, bg:'#F5F3FF' },
          { label: 'Insumos em falta', value: `${missingItems.length}`,     icon: <AlertTriangle size={16} style={{color:'#C53030'}} />, bg:'#FFF1F1' },
          { label: 'Produzidos hoje',  value: '32 peças',        icon: <CheckCircle size={16} style={{color:'#047857'}} />, bg:'#ECFDF5' },
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

      {/* Tabs */}
      <div className="flex items-center gap-2">
        {[{key:'pcp',label:'Ordens de Produção'},{key:'bom',label:'Ficha Técnica (BOM)'}].map(t => (
          <button key={t.key} onClick={() => setActiveTab(t.key as 'pcp'|'bom')}
            className={activeTab === t.key ? 'btn-primary' : 'btn-ghost'} style={activeTab !== t.key ? {border:'1px solid rgba(26,51,120,0.12)'} : {}}>
            {t.label}
          </button>
        ))}
        <button className="btn-primary ml-auto"><Plus size={14} />{activeTab === 'pcp' ? 'Nova OP' : 'Novo Produto'}</button>
      </div>

      {/* PCP Tab */}
      {activeTab === 'pcp' && (
        <div className="card overflow-hidden animate-fade-in">
          <table className="w-full">
            <thead style={{background:'#FAFBFF',borderBottom:'1px solid rgba(26,51,120,0.07)'}}>
              <tr>
                <th className="table-header">Ordem</th>
                <th className="table-header">Produto</th>
                <th className="table-header">Qty planejada</th>
                <th className="table-header">Progresso</th>
                <th className="table-header">Período</th>
                <th className="table-header">Status</th>
              </tr>
            </thead>
            <tbody>
              {orders.map(op => {
                const s = statusConfig[op.status]
                const pct = op.qty > 0 ? (op.produced / op.qty) * 100 : 0
                return (
                  <tr key={op.id} className="table-row-hover">
                    <td className="table-cell">
                      <span style={{fontFamily:'var(--font-jetbrains)',fontSize:12,color:'#1A47C8',fontWeight:600}}>{op.id}</span>
                    </td>
                    <td className="table-cell">
                      <p style={{fontWeight:600,fontSize:13,color:'#0D1B4B'}}>{op.product}</p>
                    </td>
                    <td className="table-cell">
                      <span style={{fontFamily:'var(--font-jetbrains)',fontSize:13,fontWeight:600,color:'#0D1B4B'}}>{op.qty}</span>
                      <span style={{fontSize:11,color:'#8892B8',marginLeft:4}}>peças</span>
                    </td>
                    <td className="table-cell" style={{minWidth:180}}>
                      <div className="flex items-center gap-3">
                        <div style={{flex:1,height:6,borderRadius:10,background:'rgba(26,51,120,0.08)',overflow:'hidden'}}>
                          <div style={{height:'100%',borderRadius:10,width:`${pct}%`,background: pct===100 ? 'linear-gradient(90deg,#047857,#059669)' : 'linear-gradient(90deg,#1A47C8,#4B79F5)',transition:'width 0.5s ease'}} />
                        </div>
                        <span style={{fontSize:11,fontWeight:700,color:'#0D1B4B',minWidth:36}}>{op.produced}/{op.qty}</span>
                      </div>
                    </td>
                    <td className="table-cell">
                      <p style={{fontSize:11,color:'#4A5680'}}>{op.start}</p>
                      <p style={{fontSize:11,color:'#8892B8'}}>→ {op.end}</p>
                    </td>
                    <td className="table-cell"><span className={s.cls}>{s.icon}{s.label}</span></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* BOM Tab */}
      {activeTab === 'bom' && (
        <div className="card overflow-hidden animate-fade-in">
          <div style={{padding:'14px 20px',borderBottom:'1px solid rgba(26,51,120,0.07)',display:'flex',alignItems:'center',justifyContent:'space-between'}}>
            <div>
              <p style={{fontSize:14,fontWeight:700,color:'#0D1B4B'}}>Camisa Social Slim — Ficha Técnica v3</p>
              <p style={{fontSize:11,color:'#8892B8',marginTop:2}}>Custo total por unidade: <strong style={{color:'#1A47C8'}}>R$ {totalCost.toFixed(2)}</strong></p>
            </div>
            {missingItems.length > 0 && (
              <div className="flex items-center gap-2 px-3 py-2 rounded-xl" style={{background:'#FFF1F1',border:'1px solid rgba(197,48,48,0.2)'}}>
                <AlertTriangle size={13} style={{color:'#C53030'}} />
                <span style={{fontSize:12,fontWeight:600,color:'#C53030'}}>{missingItems.length} insumo(s) insuficientes para a OP</span>
              </div>
            )}
          </div>
          <table className="w-full">
            <thead style={{background:'#FAFBFF',borderBottom:'1px solid rgba(26,51,120,0.07)'}}>
              <tr>
                <th className="table-header">Componente</th>
                <th className="table-header">NCM</th>
                <th className="table-header text-right">Qtd líquida</th>
                <th className="table-header text-right">Perda</th>
                <th className="table-header text-right">Qtd bruta</th>
                <th className="table-header text-right">Custo unit.</th>
                <th className="table-header text-right">Custo total</th>
                <th className="table-header text-center">Disponível</th>
              </tr>
            </thead>
            <tbody>
              {bom.map(item => {
                const ok = item.available >= item.netQty * 50
                return (
                  <tr key={item.id} className="table-row-hover" style={!ok ? {background:'rgba(197,48,48,0.02)'} : {}}>
                    <td className="table-cell">
                      <div className="flex items-center gap-2">
                        {!ok && <AlertTriangle size={12} style={{color:'#C53030',flexShrink:0}} />}
                        <span style={{fontSize:13,fontWeight:500,color:'#0D1B4B'}}>{item.component}</span>
                      </div>
                    </td>
                    <td className="table-cell"><span style={{fontFamily:'var(--font-jetbrains)',fontSize:11,color:'#8892B8'}}>{item.ncm}</span></td>
                    <td className="table-cell text-right" style={{fontFamily:'var(--font-jetbrains)',fontSize:12}}>{item.qty} {item.unit}</td>
                    <td className="table-cell text-right">
                      <span style={{fontSize:11,color: item.scrap > 0 ? '#B45309' : '#8892B8',fontWeight:item.scrap > 0 ? 600 : 400}}>
                        {(item.scrap * 100).toFixed(0)}%
                      </span>
                    </td>
                    <td className="table-cell text-right" style={{fontFamily:'var(--font-jetbrains)',fontSize:12,fontWeight:600,color:'#0D1B4B'}}>{item.netQty.toFixed(3)}</td>
                    <td className="table-cell text-right" style={{fontFamily:'var(--font-jetbrains)',fontSize:12}}>R$ {item.unitCost.toFixed(2)}</td>
                    <td className="table-cell text-right" style={{fontFamily:'var(--font-jetbrains)',fontSize:12,fontWeight:700,color:'#0D1B4B'}}>R$ {item.totalCost.toFixed(2)}</td>
                    <td className="table-cell text-center">
                      <span className={ok ? 'badge-approved' : 'badge-danger'} style={{fontSize:10}}>
                        {ok ? <CheckCircle size={9} /> : <AlertTriangle size={9} />}
                        {item.available} {item.unit}
                      </span>
                    </td>
                  </tr>
                )
              })}
              <tr style={{background:'#FAFBFF',borderTop:'2px solid rgba(26,51,120,0.1)'}}>
                <td colSpan={6} className="table-cell" style={{fontWeight:700,color:'#0D1B4B'}}>Custo total por unidade</td>
                <td className="table-cell text-right" style={{fontFamily:'var(--font-jetbrains)',fontSize:14,fontWeight:800,color:'#1A47C8'}}>R$ {totalCost.toFixed(2)}</td>
                <td />
              </tr>
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
