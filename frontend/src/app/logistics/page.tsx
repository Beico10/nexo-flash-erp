'use client'
import { useState } from 'react'
import { Truck, MapPin, TrendingUp, TrendingDown, Calculator, Plus, Eye, FileText, AlertTriangle } from 'lucide-react'

const routes = [
  { id: '1', origin: 'São Paulo, SP', dest: 'Campinas, SP', distance: 98, vehicle: 'VUC', shipper: 'Atacadão SP', status: 'planned', gross: 420, cost: 280, profit: 140, margin: 33.3, driver: 'José Souza' },
  { id: '2', origin: 'São Paulo, SP', dest: 'Santos, SP',   distance: 72, vehicle: 'Truck', shipper: 'Porto Sul', status: 'in_transit', gross: 850, cost: 540, profit: 310, margin: 36.5, driver: 'Carlos Lima' },
  { id: '3', origin: 'Campinas, SP', dest: 'Ribeirão Preto, SP', distance: 235, vehicle: 'Carreta', shipper: 'Atacadão SP', status: 'done', gross: 2100, cost: 1450, profit: 650, margin: 31.0, driver: 'Pedro Rocha' },
  { id: '4', origin: 'São Paulo, SP', dest: 'Sorocaba, SP', distance: 96, vehicle: 'Van', shipper: 'Geral', status: 'planned', gross: 380, cost: 290, profit: 90, margin: 23.7, driver: 'Ana Costa' },
]

const vehicleBadge: Record<string, string> = {
  VUC: 'bg-purple-100 text-purple-700', Truck: 'bg-blue-100 text-blue-700',
  Carreta: 'bg-nexo-50 text-nexo-700', Van: 'bg-teal-100 text-teal-700',
}

const statusConfig: Record<string, { label: string; cls: string }> = {
  planned:    { label: 'Planejada',   cls: 'badge-open' },
  in_transit: { label: 'Em trânsito', cls: 'badge-approved' },
  done:       { label: 'Entregue',    cls: 'badge-done' },
}

export default function LogisticsPage() {
  const [showDRE, setShowDRE] = useState<string | null>(null)
  const dreRoute = routes.find(r => r.id === showDRE)

  return (
    <div className="space-y-5 animate-fade-in">

      {/* Top stats */}
      <div className="grid grid-cols-4 gap-4 stagger">
        {[
          { label: 'Rotas hoje', value: '4', sub: '2 em trânsito', icon: <Truck size={17} style={{color:'#1A47C8'}} />, bg: '#EEF3FF' },
          { label: 'Receita bruta', value: 'R$ 3.750', sub: '+8% vs ontem', icon: <TrendingUp size={17} style={{color:'#047857'}} />, bg: '#ECFDF5' },
          { label: 'Lucro estimado', value: 'R$ 1.190', sub: '31,7% margem', icon: <Calculator size={17} style={{color:'#B45309'}} />, bg: '#FFFBEB' },
          { label: 'CT-e emitidos', value: '12', sub: 'Este mês', icon: <FileText size={17} style={{color:'#7C3AED'}} />, bg: '#F5F3FF' },
        ].map(s => (
          <div key={s.label} className="card stat-card animate-slide-up">
            <div className="flex items-center justify-between">
              <p className="section-label">{s.label}</p>
              <div className="w-8 h-8 rounded-xl flex items-center justify-center" style={{background:s.bg}}>{s.icon}</div>
            </div>
            <div>
              <p className="metric-lg">{s.value}</p>
              <p style={{fontSize:11,color:'#8892B8',marginTop:3}}>{s.sub}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-3">
        <h2 style={{fontSize:14,fontWeight:700,color:'#0D1B4B',flex:1}}>Rotas e Contratos</h2>
        <button className="btn-secondary text-xs">Calcular Frete</button>
        <button className="btn-primary"><Plus size={14} />Nova Rota</button>
      </div>

      {/* Routes table */}
      <div className="card overflow-hidden">
        <table className="w-full">
          <thead style={{background:'#FAFBFF',borderBottom:'1px solid rgba(26,51,120,0.07)'}}>
            <tr>
              <th className="table-header">Origem → Destino</th>
              <th className="table-header">Embarcador</th>
              <th className="table-header">Veículo</th>
              <th className="table-header">Motorista</th>
              <th className="table-header">Distância</th>
              <th className="table-header">Status</th>
              <th className="table-header text-right">Margem</th>
              <th className="table-header text-center">DRE</th>
            </tr>
          </thead>
          <tbody>
            {routes.map((r) => {
              const s = statusConfig[r.status]
              const profitable = r.profit >= 0
              return (
                <tr key={r.id} className="table-row-hover group">
                  <td className="table-cell">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rounded-lg flex items-center justify-center" style={{background:'rgba(26,71,200,0.06)'}}>
                        <MapPin size={11} style={{color:'#1A47C8'}} />
                      </div>
                      <div>
                        <p style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}}>{r.origin}</p>
                        <p style={{fontSize:11,color:'#8892B8'}}>→ {r.dest}</p>
                      </div>
                    </div>
                  </td>
                  <td className="table-cell">
                    <span style={{fontSize:12,fontWeight:500,color:r.shipper === 'Geral' ? '#8892B8' : '#0D1B4B'}}>
                      {r.shipper}
                    </span>
                    {r.shipper !== 'Geral' && (
                      <span style={{fontSize:10,marginLeft:4,background:'#EEF3FF',color:'#1A47C8',padding:'1px 6px',borderRadius:10,fontWeight:600}}>específico</span>
                    )}
                  </td>
                  <td className="table-cell">
                    <span className={`text-xs font-semibold px-2 py-1 rounded-lg ${vehicleBadge[r.vehicle]}`}>{r.vehicle}</span>
                  </td>
                  <td className="table-cell" style={{fontSize:12}}>{r.driver}</td>
                  <td className="table-cell">
                    <span style={{fontFamily:'var(--font-jetbrains)',fontSize:12,color:'#4A5680'}}>{r.distance} km</span>
                  </td>
                  <td className="table-cell"><span className={s.cls}>{s.label}</span></td>
                  <td className="table-cell text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      {profitable ? <TrendingUp size={12} style={{color:'#047857'}} /> : <TrendingDown size={12} style={{color:'#C53030'}} />}
                      <span style={{fontWeight:700,fontSize:13,color: profitable ? '#047857' : '#C53030'}}>{r.margin.toFixed(1)}%</span>
                    </div>
                  </td>
                  <td className="table-cell text-center">
                    <button onClick={() => setShowDRE(r.id)} className="btn-ghost py-1.5 px-3 text-xs opacity-0 group-hover:opacity-100 transition-opacity">
                      <Eye size={13} />DRE
                    </button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {/* DRE Modal */}
      {showDRE && dreRoute && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" style={{background:'rgba(13,27,75,0.45)',backdropFilter:'blur(4px)'}}>
          <div className="animate-scale-in" style={{background:'#fff',borderRadius:20,boxShadow:'0 24px 48px rgba(13,27,75,0.2)',width:'100%',maxWidth:460,border:'1px solid rgba(26,51,120,0.1)'}}>
            {/* Header */}
            <div className="card-premium p-6">
              <p style={{fontSize:11,fontWeight:700,color:'rgba(255,255,255,0.6)',textTransform:'uppercase',letterSpacing:'0.08em'}}>DRE da Viagem</p>
              <p style={{fontFamily:'var(--font-syne)',fontSize:18,fontWeight:800,color:'#fff',marginTop:4}}>{dreRoute.origin} → {dreRoute.dest}</p>
              <div className="flex items-center gap-3 mt-3">
                <span style={{fontSize:11,background:'rgba(255,255,255,0.15)',color:'rgba(255,255,255,0.9)',padding:'3px 10px',borderRadius:20,fontWeight:600}}>{dreRoute.vehicle}</span>
                <span style={{fontSize:11,color:'rgba(255,255,255,0.7)'}}>{dreRoute.distance} km · {dreRoute.driver}</span>
              </div>
            </div>

            {/* DRE lines */}
            <div className="p-6 space-y-3">
              {[
                { label: 'Receita bruta de frete', value: dreRoute.gross, type: 'revenue' },
                { label: 'Combustível estimado',   value: -(dreRoute.cost * 0.45), type: 'cost' },
                { label: 'Custo do motorista',      value: -(dreRoute.cost * 0.35), type: 'cost' },
                { label: 'Pedágios',                value: -(dreRoute.cost * 0.20), type: 'cost' },
              ].map((line, i) => (
                <div key={i} className="flex items-center justify-between py-2" style={{borderBottom:'1px solid rgba(26,51,120,0.06)'}}>
                  <span style={{fontSize:13,color:'#4A5680'}}>{line.label}</span>
                  <span style={{fontSize:14,fontWeight:600,color: line.type === 'revenue' ? '#047857' : '#C53030', fontFamily:'var(--font-jetbrains)'}}>
                    {line.type === 'revenue' ? '+' : ''}R$ {Math.abs(line.value).toLocaleString('pt-BR',{minimumFractionDigits:2})}
                  </span>
                </div>
              ))}

              {/* Result */}
              <div className="flex items-center justify-between py-3 px-4 rounded-xl mt-2" style={{background: dreRoute.profit >= 0 ? '#ECFDF5' : '#FFF1F1', border: `1px solid ${dreRoute.profit >= 0 ? 'rgba(4,120,87,0.15)' : 'rgba(197,48,48,0.15)'}`}}>
                <div>
                  <p style={{fontSize:12,fontWeight:700,color: dreRoute.profit >= 0 ? '#047857' : '#C53030',textTransform:'uppercase',letterSpacing:'0.05em'}}>Resultado estimado</p>
                  <p style={{fontSize:11,color:'#8892B8',marginTop:1}}>Margem: {dreRoute.margin.toFixed(1)}%</p>
                </div>
                <div className="text-right">
                  <p style={{fontFamily:'var(--font-syne)',fontSize:24,fontWeight:800,color: dreRoute.profit >= 0 ? '#047857' : '#C53030'}}>
                    R$ {dreRoute.profit.toLocaleString('pt-BR')}
                  </p>
                  {dreRoute.profit < 0 && (
                    <div className="flex items-center gap-1 justify-end mt-1">
                      <AlertTriangle size={11} style={{color:'#C53030'}} />
                      <span style={{fontSize:10,color:'#C53030',fontWeight:600}}>Revisar antes de partir</span>
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="flex gap-3 px-6 pb-6">
              <button onClick={() => setShowDRE(null)} className="btn-ghost flex-1 justify-center">Fechar</button>
              <button className="btn-primary flex-1 justify-center"><FileText size={14} />Emitir CT-e</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
