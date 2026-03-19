'use client'
import { Bell, Search, HelpCircle } from 'lucide-react'
import { usePathname } from 'next/navigation'

const titles: Record<string, string> = {
  '/dashboard':    'Dashboard',
  '/mechanic':     'Mecânica — Ordens de Serviço',
  '/bakery':       'Padaria — PDV',
  '/industry':     'Indústria — PCP',
  '/logistics':    'Logística — CT-e e Rotas',
  '/aesthetics':   'Estética — Agenda',
  '/shoes':        'Calçados — Grades',
  '/ai-approvals': 'Aprovações de IA',
  '/payments':     'Pagamentos — PIX e Boleto',
  '/settings':     'Configurações',
}

export default function Header() {
  const pathname = usePathname()
  const title = Object.entries(titles).find(([k]) => pathname.startsWith(k))?.[1] ?? 'Nexo Flash'

  return (
    <header className="h-16 bg-white border-b border-slate-100 px-6 flex items-center gap-4 flex-shrink-0">
      {/* Title */}
      <div className="flex-1">
        <h1 className="text-base font-semibold text-slate-800">{title}</h1>
        <p className="text-xs text-slate-400">
          {new Date().toLocaleDateString('pt-BR', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
        </p>
      </div>

      {/* Search */}
      <div className="hidden md:flex items-center gap-2 px-3.5 py-2 bg-slate-50 border border-slate-200 rounded-xl w-64">
        <Search size={14} className="text-slate-400" />
        <input
          type="text"
          placeholder="Buscar OS, produto, cliente..."
          className="bg-transparent text-sm text-slate-600 placeholder-slate-400 outline-none flex-1"
        />
        <kbd className="text-[10px] text-slate-300 font-mono">⌘K</kbd>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1">
        <button className="w-9 h-9 flex items-center justify-center rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors relative">
          <Bell size={18} />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-nexo-500 rounded-full ring-2 ring-white" />
        </button>
        <button className="w-9 h-9 flex items-center justify-center rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors">
          <HelpCircle size={18} />
        </button>
      </div>

      {/* IBS/CBS badge */}
      <div className="hidden lg:flex items-center gap-1.5 px-3 py-1.5 bg-nexo-50 border border-nexo-100 rounded-xl">
        <div className="w-1.5 h-1.5 bg-nexo-500 rounded-full animate-pulse" />
        <span className="text-xs font-semibold text-nexo-600">IBS/CBS 2026</span>
      </div>
    </header>
  )
}
