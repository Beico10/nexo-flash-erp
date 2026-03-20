'use client'
import { Factory, ClipboardList, BarChart3 } from 'lucide-react'

export default function IndustryPage() {
  return (
    <div className="space-y-5 animate-fade-in" data-testid="industry-page">
      <div className="grid lg:grid-cols-3 gap-4">
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-indigo-50 rounded-xl flex items-center justify-center"><Factory size={20} className="text-indigo-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">12</p><p className="text-xs text-slate-400">Ordens de producao</p></div>
        </div>
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-amber-50 rounded-xl flex items-center justify-center"><ClipboardList size={20} className="text-amber-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">89%</p><p className="text-xs text-slate-400">Eficiencia produtiva</p></div>
        </div>
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-emerald-50 rounded-xl flex items-center justify-center"><BarChart3 size={20} className="text-emerald-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">3.2t</p><p className="text-xs text-slate-400">Producao hoje</p></div>
        </div>
      </div>
      <div className="card p-8 text-center">
        <Factory size={48} className="text-slate-200 mx-auto mb-3" />
        <p className="text-sm text-slate-500">Modulo Industria - PCP em desenvolvimento</p>
        <p className="text-xs text-slate-400 mt-1">Planejamento e Controle de Producao sera ativado em breve</p>
      </div>
    </div>
  )
}
