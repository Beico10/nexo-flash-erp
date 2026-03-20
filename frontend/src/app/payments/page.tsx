'use client'
import { CreditCard, QrCode, FileText } from 'lucide-react'

export default function PaymentsPage() {
  return (
    <div className="space-y-5 animate-fade-in" data-testid="payments-page">
      <div className="grid lg:grid-cols-3 gap-4">
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-green-50 rounded-xl flex items-center justify-center"><QrCode size={20} className="text-green-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">PIX</p><p className="text-xs text-slate-400">Cobranca instantanea</p></div>
        </div>
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-blue-50 rounded-xl flex items-center justify-center"><FileText size={20} className="text-blue-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">Boleto</p><p className="text-xs text-slate-400">Geracao automatica</p></div>
        </div>
        <div className="card p-5 flex items-center gap-4">
          <div className="w-12 h-12 bg-violet-50 rounded-xl flex items-center justify-center"><CreditCard size={20} className="text-violet-600" /></div>
          <div><p className="text-lg font-bold text-slate-800">Split</p><p className="text-xs text-slate-400">Pagamento dividido</p></div>
        </div>
      </div>
      <div className="card p-8 text-center">
        <CreditCard size={48} className="text-slate-200 mx-auto mb-3" />
        <p className="text-sm text-slate-500">Modulo BaaS - Integracao bancaria</p>
        <p className="text-xs text-slate-400 mt-1">PIX, Boleto e Split de pagamentos serao ativados com gateway bancario</p>
      </div>
    </div>
  )
}
