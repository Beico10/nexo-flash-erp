'use client'
import { useState, useEffect } from 'react'
import { Brain, CheckCircle, XCircle, AlertTriangle, Loader2, ThumbsUp, ThumbsDown } from 'lucide-react'

interface Suggestion {
  ID: string
  Type: string
  TargetTable: string
  TargetID: string
  Reason: string
  Confidence: number
  CreatedByAI: string
  Status: string
}

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }
const apiFetch = (path: string, opts: RequestInit = {}) => fetch(path, { ...opts, headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${getToken()}`, ...(opts.headers as Record<string, string> || {}) } })

const typeLabels: Record<string, string> = {
  missing_labor_cost: 'Custo de mao de obra ausente',
  ncm_correction: 'Correcao de NCM',
  onboard_field: 'Importacao de NF-e',
}

export default function AIApprovalsPage() {
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!getToken()) { window.location.href = '/login'; return }
    apiFetch('/api/v1/ai/suggestions').then(r => { if (r.status === 401) { window.location.href = '/login'; return null }; return r.json() }).then(d => { if (d) { setSuggestions(d.data || []); setLoading(false) } }).catch(() => setLoading(false))
  }, [])

  async function handleAction(id: string, action: 'approve' | 'reject') {
    await apiFetch(`/api/v1/ai/suggestions/${id}/${action}`, { method: 'POST', body: JSON.stringify({ reason: action === 'reject' ? 'Rejeitado pelo usuario' : '' }) })
    setSuggestions(prev => prev.filter(s => s.ID !== id))
  }

  if (loading) return <div className="flex items-center justify-center h-64"><Loader2 size={32} className="text-nexo-500 animate-spin" /></div>

  return (
    <div className="space-y-5 animate-fade-in" data-testid="ai-approvals-page">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-slate-800">Aprovacoes IA</h2>
          <p className="text-sm text-slate-400">Todas as sugestoes da IA precisam de aprovacao humana</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-violet-50 border border-violet-100">
          <Brain size={14} className="text-violet-500" />
          <span className="text-xs font-semibold text-violet-600">{suggestions.length} pendente{suggestions.length !== 1 ? 's' : ''}</span>
        </div>
      </div>

      {suggestions.length === 0 ? (
        <div className="card p-12 text-center">
          <CheckCircle size={40} className="text-emerald-400 mx-auto mb-3" />
          <p className="text-sm text-slate-500">Nenhuma sugestao pendente!</p>
        </div>
      ) : (
        <div className="space-y-3">
          {suggestions.map(s => (
            <div key={s.ID} className="card p-5 hover:shadow-md transition-shadow" data-testid={`suggestion-${s.ID}`}>
              <div className="flex items-start gap-4">
                <div className="w-10 h-10 rounded-xl bg-violet-50 flex items-center justify-center flex-shrink-0">
                  <Brain size={18} className="text-violet-500" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-semibold text-slate-800">{typeLabels[s.Type] || s.Type}</span>
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-50 text-amber-600 border border-amber-100">
                      {Math.round(s.Confidence * 100)}% confianca
                    </span>
                    <span className="text-[10px] text-slate-400">{s.CreatedByAI}</span>
                  </div>
                  <p className="text-sm text-slate-600 mb-2">{s.Reason}</p>
                  <div className="flex items-center gap-2 text-xs text-slate-400">
                    <AlertTriangle size={11} />
                    <span>Ref: {s.TargetTable} / {s.TargetID}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button onClick={() => handleAction(s.ID, 'approve')} className="flex items-center gap-1.5 px-3 py-2 rounded-xl bg-emerald-50 text-emerald-600 hover:bg-emerald-100 transition-colors text-xs font-semibold" data-testid={`approve-${s.ID}`}>
                    <ThumbsUp size={13} /> Aprovar
                  </button>
                  <button onClick={() => handleAction(s.ID, 'reject')} className="flex items-center gap-1.5 px-3 py-2 rounded-xl bg-red-50 text-red-500 hover:bg-red-100 transition-colors text-xs font-semibold" data-testid={`reject-${s.ID}`}>
                    <ThumbsDown size={13} /> Rejeitar
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
