'use client'
import { useState } from 'react'
import { Brain, CheckCircle, XCircle, Eye, AlertTriangle, Package, Wrench, FileText, TrendingUp } from 'lucide-react'

const suggestions = [
  {
    id: '1',
    type: 'missing_labor_cost',
    typeLabel: 'Mão de obra faltante',
    targetTable: 'mechanic_os',
    targetRef: 'OS-2026-001842',
    reason: 'OS contém 3 peças (pastilha, disco, fluido) mas nenhum item de mão de obra foi registrado. Instalação deve ser cobrada separadamente.',
    confidence: 0.94,
    urgency: 'high',
    suggestedData: { descricao: 'Troca de pastilha e disco de freio dianteiro', horas: 1.5, valor_hora: 120.00, total: 180.00 },
    createdAt: '2026-03-19 09:12',
    agent: 'co-pilot-v1',
  },
  {
    id: '2',
    type: 'ncm_correction',
    typeLabel: 'Correção de NCM',
    targetTable: 'products',
    targetRef: 'Pastilha de Freio Dianteira',
    reason: 'NCM cadastrado (84714900) parece incorreto para este produto. Sugestão baseada na descrição e CFOP da NF-e de compra.',
    confidence: 0.87,
    urgency: 'medium',
    suggestedData: { ncm_atual: '84714900', ncm_sugerido: '87083000', descricao_ncm: 'Partes e acessórios de veículos' },
    createdAt: '2026-03-19 08:55',
    agent: 'concierge-v1',
  },
  {
    id: '3',
    type: 'onboard_field',
    typeLabel: 'Produto detectado em NF-e',
    targetTable: 'products',
    targetRef: 'NF-e 001.234 — Auto Peças Silva',
    reason: '12 produtos detectados no XML da NF-e de compra. Nenhum deles consta no catálogo atual.',
    confidence: 0.92,
    urgency: 'low',
    suggestedData: { total_produtos: 12, fornecedor: 'Auto Peças Silva LTDA', cnpj: '12.345.678/0001-90' },
    createdAt: '2026-03-19 08:30',
    agent: 'concierge-v1',
  },
]

const typeIcon: Record<string, React.ReactNode> = {
  missing_labor_cost: <Wrench size={14} />,
  ncm_correction:    <FileText size={14} />,
  onboard_field:     <Package size={14} />,
  price_anomaly:     <TrendingUp size={14} />,
}

const urgencyConfig = {
  high:   { label: 'Alta',  cls: 'bg-red-100 text-red-600 border-red-200' },
  medium: { label: 'Média', cls: 'bg-amber-100 text-amber-600 border-amber-200' },
  low:    { label: 'Baixa', cls: 'bg-nexo-100 text-nexo-600 border-nexo-200' },
}

export default function AIApprovalsPage() {
  const [items, setItems] = useState(suggestions)
  const [selected, setSelected] = useState<string | null>(null)

  const approve = (id: string) => setItems(i => i.filter(x => x.id !== id))
  const reject  = (id: string) => setItems(i => i.filter(x => x.id !== id))

  return (
    <div className="space-y-5 animate-fade-in">

      {/* Header info */}
      <div className="card p-5 bg-nexo-gradient text-white border-0">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-white/20 rounded-xl flex items-center justify-center">
            <Brain size={20} className="text-white" />
          </div>
          <div>
            <h2 className="font-display font-700 text-lg">Painel de Aprovação — IA Human-in-the-Loop</h2>
            <p className="text-white/70 text-sm">Nenhuma IA altera dados sem sua aprovação. Cada sugestão abaixo exige um clique seu para ser aplicada.</p>
          </div>
          <div className="ml-auto text-right">
            <p className="text-3xl font-display font-700">{items.length}</p>
            <p className="text-white/70 text-xs">pendentes</p>
          </div>
        </div>
      </div>

      {items.length === 0 && (
        <div className="card py-16 text-center">
          <CheckCircle size={40} className="text-emerald-300 mx-auto mb-3" />
          <p className="font-medium text-slate-500">Tudo em dia! Nenhuma sugestão pendente.</p>
        </div>
      )}

      {/* Suggestions */}
      <div className="space-y-3">
        {items.map((s) => {
          const urg = urgencyConfig[s.urgency as keyof typeof urgencyConfig]
          const isOpen = selected === s.id
          return (
            <div key={s.id} className={`card overflow-hidden transition-all duration-200 ${isOpen ? 'ring-2 ring-nexo-300' : ''}`}>
              <div className="p-5">
                <div className="flex items-start gap-4">
                  {/* Icon */}
                  <div className="w-9 h-9 bg-nexo-50 rounded-xl flex items-center justify-center text-nexo-500 flex-shrink-0">
                    {typeIcon[s.type] ?? <Brain size={14} />}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-semibold text-slate-800 text-sm">{s.typeLabel}</span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border ${urg.cls}`}>
                        Urgência {urg.label}
                      </span>
                      <span className="text-[10px] font-medium px-2 py-0.5 rounded-full bg-slate-100 text-slate-500">
                        Confiança {Math.round(s.confidence * 100)}%
                      </span>
                      <span className="text-[10px] text-slate-400 ml-auto">{s.createdAt} · {s.agent}</span>
                    </div>
                    <p className="text-xs text-nexo-600 font-medium mt-0.5">{s.targetRef}</p>
                    <p className="text-sm text-slate-500 mt-1.5 leading-relaxed">{s.reason}</p>

                    {/* Suggested data preview */}
                    {isOpen && (
                      <div className="mt-3 p-3 bg-slate-50 rounded-xl border border-slate-200 animate-fade-in">
                        <p className="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Dados que serão aplicados:</p>
                        <div className="space-y-1">
                          {Object.entries(s.suggestedData).map(([k, v]) => (
                            <div key={k} className="flex items-center gap-2">
                              <span className="text-xs font-mono text-slate-400 w-32 flex-shrink-0">{k}</span>
                              <span className="text-xs font-medium text-slate-700">
                                {typeof v === 'number' && k.includes('valor') || k.includes('total')
                                  ? `R$ ${(v as number).toLocaleString('pt-BR', {minimumFractionDigits: 2})}`
                                  : String(v)
                                }
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-2 mt-4 pt-4 border-t border-slate-50">
                  <button
                    onClick={() => approve(s.id)}
                    className="btn-primary"
                  >
                    <CheckCircle size={14} />
                    Aprovar e Aplicar
                  </button>
                  <button
                    onClick={() => reject(s.id)}
                    className="btn-danger"
                  >
                    <XCircle size={14} />
                    Rejeitar
                  </button>
                  <button
                    onClick={() => setSelected(isOpen ? null : s.id)}
                    className="btn-ghost"
                  >
                    <Eye size={14} />
                    {isOpen ? 'Ocultar' : 'Ver detalhes'}
                  </button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
