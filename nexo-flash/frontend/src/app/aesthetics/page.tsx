'use client'
import { useState } from 'react'
import { Plus, ChevronLeft, ChevronRight, Clock, Scissors, User, DollarSign, AlertCircle } from 'lucide-react'

const professionals = [
  { id: '1', name: 'Ana Beatriz', color: '#1A47C8', initials: 'AB' },
  { id: '2', name: 'Carla Souza', color: '#7C3AED', initials: 'CS' },
  { id: '3', name: 'Daniela Lima', color: '#047857', initials: 'DL' },
]

const appointments = [
  { id: '1', profId: '1', customerName: 'Maria Santos',    service: 'Coloração completa', start: 9,  duration: 2,   price: 280, status: 'confirmed' },
  { id: '2', profId: '1', customerName: 'Joana Pereira',   service: 'Corte + escova',     start: 11.5, duration: 1.5, price: 120, status: 'scheduled' },
  { id: '3', profId: '2', customerName: 'Fernanda Costa',  service: 'Manicure + pedicure',start: 9,  duration: 1.5, price: 85,  status: 'in_progress' },
  { id: '4', profId: '2', customerName: 'Paula Rodrigues', service: 'Design de sobrancelha',start:11, duration: 1,   price: 60,  status: 'scheduled' },
  { id: '5', profId: '3', customerName: 'Letícia Alves',   service: 'Hidratação',         start: 10, duration: 1.5, price: 95,  status: 'confirmed' },
  { id: '6', profId: '3', customerName: 'Camila Torres',   service: 'Limpeza de pele',    start: 13, duration: 1,   price: 110, status: 'scheduled' },
]

const statusStyle: Record<string, { bg: string; border: string; label: string }> = {
  scheduled:   { bg: 'rgba(26,71,200,0.06)',  border: 'rgba(26,71,200,0.2)',  label: 'Agendado' },
  confirmed:   { bg: 'rgba(4,120,87,0.06)',   border: 'rgba(4,120,87,0.2)',   label: 'Confirmado' },
  in_progress: { bg: 'rgba(180,83,9,0.06)',   border: 'rgba(180,83,9,0.2)',   label: 'Em andamento' },
  done:        { bg: 'rgba(107,114,128,0.06)',border: 'rgba(107,114,128,0.2)',label: 'Concluído' },
}

const hours = [8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18]
const CELL_H = 56 // px per hour

export default function AestheticsPage() {
  const [showModal, setShowModal] = useState(false)
  const [selectedDate] = useState(new Date())

  const totalRevenue = appointments.reduce((s, a) => s + a.price, 0)

  return (
    <div className="space-y-5 animate-fade-in" style={{height:'calc(100vh - 10rem)',display:'flex',flexDirection:'column'}}>

      {/* Top bar */}
      <div className="flex items-center gap-4 flex-shrink-0">
        <div className="card flex items-center gap-3 px-4 py-2.5">
          <button className="w-7 h-7 rounded-lg flex items-center justify-center hover:bg-slate-50 transition-colors"><ChevronLeft size={14} /></button>
          <div className="text-center px-2">
            <p style={{fontSize:14,fontWeight:700,color:'#0D1B4B'}}>
              {selectedDate.toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })}
            </p>
          </div>
          <button className="w-7 h-7 rounded-lg flex items-center justify-center hover:bg-slate-50 transition-colors"><ChevronRight size={14} /></button>
        </div>

        {/* Quick stats */}
        <div className="card px-4 py-2.5 flex items-center gap-3">
          <Scissors size={14} style={{color:'#1A47C8'}} />
          <span style={{fontSize:13,fontWeight:600,color:'#0D1B4B'}}>{appointments.length} agendamentos</span>
        </div>
        <div className="card px-4 py-2.5 flex items-center gap-3">
          <DollarSign size={14} style={{color:'#047857'}} />
          <span style={{fontSize:13,fontWeight:600,color:'#047857'}}>R$ {totalRevenue.toLocaleString('pt-BR')} previsto</span>
        </div>

        <button onClick={() => setShowModal(true)} className="btn-primary ml-auto"><Plus size={14} />Novo Agendamento</button>
      </div>

      {/* Calendar grid */}
      <div className="card overflow-hidden flex-1 min-h-0">
        <div className="flex h-full overflow-hidden">

          {/* Time column */}
          <div style={{width:56,flexShrink:0,borderRight:'1px solid rgba(26,51,120,0.07)',paddingTop:48}}>
            {hours.map(h => (
              <div key={h} style={{height:CELL_H,display:'flex',alignItems:'flex-start',justifyContent:'center',paddingTop:6}}>
                <span style={{fontSize:10,fontWeight:600,color:'#B0B8D8',fontFamily:'var(--font-jetbrains)'}}>{h}:00</span>
              </div>
            ))}
          </div>

          {/* Professional columns */}
          <div className="flex flex-1 overflow-x-auto overflow-y-auto">
            {professionals.map((prof) => {
              const profApts = appointments.filter(a => a.profId === prof.id)
              return (
                <div key={prof.id} style={{flex:1,minWidth:200,borderRight:'1px solid rgba(26,51,120,0.05)',position:'relative'}}>
                  {/* Header */}
                  <div style={{height:48,borderBottom:'1px solid rgba(26,51,120,0.07)',display:'flex',alignItems:'center',gap:8,padding:'0 12px',background:'#FAFBFF',position:'sticky',top:0,zIndex:10}}>
                    <div style={{width:28,height:28,borderRadius:8,background:prof.color,display:'flex',alignItems:'center',justifyContent:'center'}}>
                      <span style={{fontSize:10,fontWeight:700,color:'#fff'}}>{prof.initials}</span>
                    </div>
                    <div>
                      <p style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}}>{prof.name}</p>
                      <p style={{fontSize:10,color:'#8892B8'}}>{profApts.length} hoje · R$ {profApts.reduce((s,a)=>s+a.price,0)}</p>
                    </div>
                  </div>

                  {/* Hour slots */}
                  <div style={{position:'relative'}}>
                    {hours.map(h => (
                      <div key={h} style={{height:CELL_H,borderBottom:'1px dashed rgba(26,51,120,0.05)',cursor:'pointer'}}
                        className="hover:bg-blue-50/30 transition-colors" />
                    ))}

                    {/* Appointments */}
                    {profApts.map(apt => {
                      const topPct = (apt.start - hours[0]) * CELL_H
                      const heightPct = apt.duration * CELL_H - 4
                      const s = statusStyle[apt.status]
                      return (
                        <div key={apt.id} style={{
                          position:'absolute', top:topPct+2, left:6, right:6,
                          height:heightPct, background:s.bg,
                          border:`1px solid ${s.border}`, borderRadius:10,
                          padding:'6px 8px', cursor:'pointer', overflow:'hidden',
                          borderLeft:`3px solid ${prof.color}`,
                          transition:'all 0.15s ease',
                        }}
                        className="hover:shadow-md group"
                        >
                          <p style={{fontSize:11,fontWeight:700,color:'#0D1B4B',lineHeight:1.2}}>{apt.customerName}</p>
                          <p style={{fontSize:10,color:'#8892B8',marginTop:2,lineHeight:1.2}}>{apt.service}</p>
                          {heightPct > 50 && (
                            <div className="flex items-center justify-between mt-1.5">
                              <div className="flex items-center gap-1">
                                <Clock size={9} style={{color:'#8892B8'}} />
                                <span style={{fontSize:9,color:'#8892B8'}}>{apt.duration}h</span>
                              </div>
                              <span style={{fontSize:10,fontWeight:700,color:prof.color}}>R$ {apt.price}</span>
                            </div>
                          )}
                        </div>
                      )
                    })}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* New appointment modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" style={{background:'rgba(13,27,75,0.45)',backdropFilter:'blur(4px)'}}>
          <div className="animate-scale-in" style={{background:'#fff',borderRadius:20,boxShadow:'0 24px 48px rgba(13,27,75,0.2)',width:'100%',maxWidth:480}}>
            <div style={{padding:'20px 24px',borderBottom:'1px solid rgba(26,51,120,0.08)'}}>
              <p style={{fontFamily:'var(--font-syne)',fontSize:17,fontWeight:800,color:'#0D1B4B'}}>Novo Agendamento</p>
              <p style={{fontSize:12,color:'#8892B8',marginTop:2}}>A trava de conflito verifica automaticamente sobreposição de horários</p>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="label">Profissional</label>
                <select className="input">
                  {professionals.map(p => <option key={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Data</label>
                  <input type="date" className="input" defaultValue="2026-03-19" />
                </div>
                <div>
                  <label className="label">Horário</label>
                  <input type="time" className="input" defaultValue="09:00" />
                </div>
              </div>
              <div>
                <label className="label">Serviço</label>
                <select className="input">
                  <option>Corte + escova</option>
                  <option>Coloração completa</option>
                  <option>Manicure + pedicure</option>
                  <option>Hidratação</option>
                </select>
              </div>
              <div>
                <label className="label">Nome do cliente</label>
                <input className="input" placeholder="Nome completo" />
              </div>
              <div>
                <label className="label">WhatsApp</label>
                <input className="input" placeholder="(11) 99999-9999" />
              </div>
              {/* Split toggle */}
              <div className="flex items-center justify-between px-4 py-3 rounded-xl" style={{background:'#F2F4FA',border:'1px solid rgba(26,51,120,0.08)'}}>
                <div>
                  <p style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}}>Split de Pagamento</p>
                  <p style={{fontSize:11,color:'#8892B8'}}>Dividir entre profissional e salão</p>
                </div>
                <div style={{width:36,height:20,borderRadius:10,background:'#1A47C8',cursor:'pointer',display:'flex',alignItems:'center',padding:2}}>
                  <div style={{width:16,height:16,borderRadius:'50%',background:'#fff',marginLeft:'auto',boxShadow:'0 1px 3px rgba(0,0,0,0.2)'}} />
                </div>
              </div>
              <div className="flex items-center gap-2 px-3 py-2.5 rounded-xl" style={{background:'#FFFBEB',border:'1px solid rgba(180,83,9,0.15)'}}>
                <AlertCircle size={13} style={{color:'#B45309',flexShrink:0}} />
                <p style={{fontSize:11,color:'#B45309'}}>Verificação de conflito automática ao salvar</p>
              </div>
            </div>
            <div className="flex gap-3 px-6 pb-6">
              <button onClick={() => setShowModal(false)} className="btn-ghost flex-1 justify-center">Cancelar</button>
              <button className="btn-primary flex-1 justify-center"><Plus size={14} />Agendar</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
