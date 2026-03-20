'use client'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { LayoutDashboard, Wrench, ShoppingBag, Factory, Truck, Scissors, Wheat, Brain, CreditCard, Settings, ChevronRight, Zap, LogOut, Activity, Calculator, Crown, Compass } from 'lucide-react'
import clsx from 'clsx'

const navGroups = [
  { label: 'Principal', items: [
    { label: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
    { label: 'Simulador Fiscal', icon: Calculator, href: '/simulador-fiscal' },
    { label: 'Aprovacoes IA', icon: Brain, href: '/ai-approvals', badge: '3' },
    { label: 'Pagamentos', icon: CreditCard, href: '/payments' },
  ]},
  { label: 'Módulos', items: [
    { label: 'Mecânica', icon: Wrench, href: '/mechanic' },
    { label: 'Padaria', icon: Wheat, href: '/bakery' },
    { label: 'Indústria', icon: Factory, href: '/industry' },
    { label: 'Logística', icon: Truck, href: '/logistics' },
    { label: 'Estética', icon: Scissors, href: '/aesthetics' },
    { label: 'Calçados', icon: ShoppingBag, href: '/shoes' },
  ]},
  { label: 'Sistema', items: [
    { label: 'Minha Assinatura', icon: Crown, href: '/dashboard/subscription' },
    { label: 'Onboarding', icon: Compass, href: '/onboarding' },
    { label: 'Configurações', icon: Settings, href: '/settings' },
  ]},
]

export default function Sidebar() {
  const pathname = usePathname()
  return (
    <aside className="w-[240px] h-screen flex flex-col flex-shrink-0 overflow-y-auto" style={{background:'#fff',borderRight:'1px solid rgba(26,51,120,0.08)'}}>
      {/* Logo */}
      <div className="px-5 pt-6 pb-5">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl flex items-center justify-center shadow-sm" style={{background:'linear-gradient(135deg,#1A47C8 0%,#0F2D8A 100%)'}}>
            <Zap size={17} className="text-white" strokeWidth={2.5} />
          </div>
          <div className="leading-none">
            <div style={{fontFamily:'var(--font-syne)',fontWeight:800,fontSize:18,letterSpacing:'-0.02em',color:'#0D1B4B'}}>
              Nexo<span style={{color:'#1A47C8'}}>One</span>
            </div>
            <div style={{fontSize:10,fontWeight:600,color:'#8892B8',letterSpacing:'0.05em',textTransform:'uppercase',marginTop:1}}>ERP Inteligente</div>
          </div>
        </div>

        {/* Tenant pill */}
        <div className="mt-4 px-3 py-2.5 rounded-xl" style={{background:'#F2F4FA',border:'1px solid rgba(26,51,120,0.08)'}}>
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 rounded-lg flex items-center justify-center" style={{background:'linear-gradient(135deg,#1A47C8,#0F2D8A)'}}>
              <Wrench size={11} className="text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <p style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}} className="truncate">Mecânica do João</p>
              <p style={{fontSize:10,color:'#8892B8',fontWeight:500}}>Plano Pro</p>
            </div>
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 pb-4 space-y-5 overflow-y-auto">
        {navGroups.map((g) => (
          <div key={g.label}>
            <p className="section-label px-3 mb-1.5">{g.label}</p>
            <div className="space-y-0.5">
              {g.items.map((item) => {
                const active = pathname.startsWith(item.href)
                const Icon = item.icon
                return (
                  <Link key={item.href} href={item.href} className={clsx(active ? 'nav-item-active' : 'nav-item')}>
                    <Icon size={15} strokeWidth={active ? 2.5 : 2} />
                    <span className="flex-1">{item.label}</span>
                    {item.badge && (
                      <span style={{fontSize:10,fontWeight:700,background:'#1A47C8',color:'#fff',padding:'2px 6px',borderRadius:20}}>{item.badge}</span>
                    )}
                    {active && <ChevronRight size={13} style={{color:'#1A47C8',opacity:0.6}} />}
                  </Link>
                )
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* IBS status */}
      <div className="mx-3 mb-3 px-3 py-2.5 rounded-xl" style={{background:'linear-gradient(135deg,rgba(26,71,200,0.06),rgba(26,71,200,0.02))',border:'1px solid rgba(26,71,200,0.12)'}}>
        <div className="flex items-center gap-2">
          <Activity size={13} style={{color:'#1A47C8'}} />
          <span style={{fontSize:11,fontWeight:600,color:'#1A47C8'}}>Motor IBS/CBS 2026</span>
          <div className="ml-auto w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
        </div>
        <p style={{fontSize:10,color:'#8892B8',marginTop:3}}>Alíquotas sincronizadas</p>
      </div>

      {/* User */}
      <div style={{borderTop:'1px solid rgba(26,51,120,0.07)'}} className="px-3 py-3">
        <div className="flex items-center gap-2.5 px-2 py-2 rounded-xl hover:bg-slate-50 transition-colors cursor-pointer group">
          <div className="w-8 h-8 rounded-xl flex items-center justify-center" style={{background:'linear-gradient(135deg,#1A47C8,#0F2D8A)'}}>
            <span style={{fontSize:11,fontWeight:700,color:'#fff'}}>AN</span>
          </div>
          <div className="flex-1 min-w-0">
            <p style={{fontSize:12,fontWeight:600,color:'#0D1B4B'}} className="truncate">Antonio</p>
            <p style={{fontSize:10,color:'#8892B8'}}>Proprietário</p>
          </div>
          <LogOut size={14} style={{color:'#8892B8'}} className="group-hover:text-red-400 transition-colors" />
        </div>
      </div>
    </aside>
  )
}
