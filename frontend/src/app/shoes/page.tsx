'use client'
import { ShoppingBag, Grid3x3 } from 'lucide-react'

export default function ShoesPage() {
  return (
    <div className="space-y-5 animate-fade-in" data-testid="shoes-page">
      <div className="grid lg:grid-cols-2 gap-4">
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-orange-50 rounded-xl flex items-center justify-center"><ShoppingBag size={20} className="text-orange-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">248</p><p className="text-xs text-slate-400">SKUs cadastrados</p></div>
        </div>
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-teal-50 rounded-xl flex items-center justify-center"><Grid3x3 size={20} className="text-teal-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">Grade</p><p className="text-xs text-slate-400">Tamanho x Cor x Modelo</p></div>
        </div>
      </div>
      <div className="card p-8 text-center">
        <ShoppingBag size={48} className="text-slate-200 mx-auto mb-3" />
        <p className="text-sm text-slate-500">Modulo Calcados - Grade em desenvolvimento</p>
        <p className="text-xs text-slate-400 mt-1">Controle por grade Tamanho x Cor sera ativado em breve</p>
      </div>
    </div>
  )
}
