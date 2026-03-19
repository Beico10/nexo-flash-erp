'use client'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard, Wrench, ShoppingBag, Factory, Truck,
  Scissors, Wheat, Brain, CreditCard, Settings, ChevronRight,
  Zap, LogOut
} from 'lucide-react'
import clsx from 'clsx'

const navGroups = [
  {
    label: 'Geral',
    items: [
      { label: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
      { label: 'Aprovações IA', icon: Brain, href: '/ai-approvals', badge: '3' },
      { label: 'Pagamentos', icon: CreditCard, href: '/payments' },
    ]
  },
  {
    label: 'Módulos',
    items: [
      { label: 'Mecânica', icon: Wrench, href: '/mechanic' },
      { label: 'Padaria', icon: Wheat, href: '/bakery' },
      { label: 'Indústria', icon: Factory, href: '/industry' },
      { label: 'Logística', icon: Truck, href: '/logistics' },
      { label: 'Estética', icon: Scissors, href: '/aesthetics' },
      { label: 'Calçados', icon: ShoppingBag, href: '/shoes' },
    ]
  },
  {
    label: 'Sistema',
    items: [
      { label: 'Configurações', icon: Settings, href: '/settings' },
    ]
  }
]

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className="w-64 h-screen bg-white border-r border-slate-100 flex flex-col flex-shrink-0 overflow-y-auto">
      {/* Logo */}
      <div className="px-5 py-5 border-b border-slate-100">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 bg-nexo-gradient rounded-xl flex items-center justify-center shadow-nexo-md">
            <Zap size={16} className="text-white" strokeWidth={2.5} />
          </div>
          <div>
            <span className="font-display text-lg font-700 text-slate-900 leading-none">Nexo</span>
            <span className="font-display text-lg font-700 text-nexo-500 leading-none"> Flash</span>
          </div>
        </div>
        {/* Tenant info */}
        <div className="mt-3 px-3 py-2 bg-nexo-50 rounded-xl border border-nexo-100">
          <p className="text-xs font-medium text-nexo-700">Mecânica do João</p>
          <p className="text-xs text-nexo-400 mt-0.5">Plano Pro · mechanic</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-5">
        {navGroups.map((group) => (
          <div key={group.label}>
            <p className="text-[10px] font-bold text-slate-300 uppercase tracking-widest px-3 mb-1.5">
              {group.label}
            </p>
            <div className="space-y-0.5">
              {group.items.map((item) => {
                const active = pathname.startsWith(item.href)
                const Icon = item.icon
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={clsx(
                      active ? 'nav-item-active' : 'nav-item'
                    )}
                  >
                    <Icon size={16} strokeWidth={active ? 2.5 : 2} />
                    <span className="flex-1">{item.label}</span>
                    {item.badge && (
                      <span className="text-[10px] font-bold bg-nexo-500 text-white px-1.5 py-0.5 rounded-full">
                        {item.badge}
                      </span>
                    )}
                    {active && <ChevronRight size={14} className="text-nexo-400" />}
                  </Link>
                )
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div className="px-3 py-4 border-t border-slate-100">
        <div className="flex items-center gap-3 px-3 py-2.5">
          <div className="w-8 h-8 bg-nexo-gradient rounded-full flex items-center justify-center">
            <span className="text-xs font-bold text-white">AN</span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-slate-700 truncate">Antonio</p>
            <p className="text-xs text-slate-400 truncate">owner</p>
          </div>
          <button className="text-slate-400 hover:text-red-500 transition-colors">
            <LogOut size={15} />
          </button>
        </div>
      </div>
    </aside>
  )
}
