'use client'
import { Bell, Search, HelpCircle, ChevronDown } from 'lucide-react'
import { usePathname } from 'next/navigation'

const titles: Record<string, { title: string; sub: string }> = {
  '/dashboard':    { title: 'Visão Geral',        sub: 'Resumo do dia e métricas principais' },
  '/mechanic':     { title: 'Ordens de Serviço',  sub: 'Mecânica — gestão de OS e aprovações' },
  '/bakery':       { title: 'PDV Rápido',          sub: 'Padaria — caixa e integração com balanças' },
  '/industry':     { title: 'PCP & Produção',      sub: 'Indústria — fichas técnicas e ordens de produção' },
  '/logistics':    { title: 'Logística',           sub: 'CT-e, contratos e DRE da viagem' },
  '/aesthetics':   { title: 'Agenda',              sub: 'Estética — agendamentos e split de pagamento' },
  '/shoes':        { title: 'Grades de Produtos',  sub: 'Calçados — matriz cor/tamanho e comissões' },
  '/ai-approvals': { title: 'Aprovações IA',       sub: 'Human-in-the-Loop — sugestões pendentes' },
  '/payments':     { title: 'Pagamentos',          sub: 'PIX dinâmico, boleto híbrido e conciliação' },
  '/settings':     { title: 'Configurações',       sub: 'Conta, plano e preferências do sistema' },
}

export default function Header() {
  const pathname = usePathname()
  const meta = Object.entries(titles).find(([k]) => pathname.startsWith(k))?.[1] ?? { title: 'Nexo Flash', sub: '' }
  const today = new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long' })

  return (
    <header style={{height:64,background:'#fff',borderBottom:'1px solid rgba(26,51,120,0.07)'}} className="px-6 flex items-center gap-5 flex-shrink-0">
      {/* Breadcrumb + title */}
      <div className="flex-1">
        <div className="flex items-baseline gap-2">
          <h1 style={{fontSize:15,fontWeight:700,color:'#0D1B4B',letterSpacing:'-0.01em'}}>{meta.title}</h1>
          <span style={{fontSize:12,color:'#8892B8',fontWeight:400}}>—</span>
          <span style={{fontSize:12,color:'#8892B8'}} className="hidden md:block">{meta.sub}</span>
        </div>
        <p style={{fontSize:11,color:'#B0B8D8',fontWeight:500,marginTop:1,textTransform:'capitalize'}}>{today}</p>
      </div>

      {/* Search bar */}
      <div className="hidden lg:flex items-center gap-2.5 px-3.5 py-2 rounded-xl" style={{background:'#F2F4FA',border:'1px solid rgba(26,51,120,0.1)',width:260}}>
        <Search size={13} style={{color:'#8892B8'}} />
        <input placeholder="Buscar OS, produto, cliente..." style={{background:'transparent',border:'none',outline:'none',fontSize:13,color:'#4A5680',flex:1,fontFamily:'var(--font-plus-jakarta)'}} />
        <kbd style={{fontSize:10,color:'#B0B8D8',fontFamily:'var(--font-jetbrains)',background:'rgba(26,51,120,0.06)',padding:'2px 5px',borderRadius:5,border:'1px solid rgba(26,51,120,0.1)'}}>⌘K</kbd>
      </div>

      {/* Notif */}
      <button style={{width:36,height:36,borderRadius:10,display:'flex',alignItems:'center',justifyContent:'center',background:'#F2F4FA',border:'1px solid rgba(26,51,120,0.08)',position:'relative',cursor:'pointer',transition:'all 0.15s'}}>
        <Bell size={16} style={{color:'#4A5680'}} />
        <span style={{position:'absolute',top:8,right:8,width:7,height:7,background:'#1A47C8',borderRadius:'50%',border:'2px solid #fff'}} />
      </button>
      <button style={{width:36,height:36,borderRadius:10,display:'flex',alignItems:'center',justifyContent:'center',background:'#F2F4FA',border:'1px solid rgba(26,51,120,0.08)',cursor:'pointer'}}>
        <HelpCircle size={16} style={{color:'#4A5680'}} />
      </button>

      {/* IBS live badge */}
      <div className="hidden xl:flex items-center gap-2 px-3 py-1.5 rounded-xl" style={{background:'linear-gradient(135deg,rgba(26,71,200,0.06),rgba(26,71,200,0.02))',border:'1px solid rgba(26,71,200,0.14)'}}>
        <div style={{width:6,height:6,borderRadius:'50%',background:'#1A47C8'}} className="animate-pulse" />
        <span style={{fontSize:11,fontWeight:700,color:'#1A47C8'}}>IBS/CBS 2026</span>
        <span style={{fontSize:10,color:'#8892B8'}}>· Ativo</span>
      </div>
    </header>
  )
}
