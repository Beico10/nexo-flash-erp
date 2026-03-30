'use client'
import { isDemoMode, promptLogin } from '@/lib/demo'
import { useState, useEffect } from 'react'
import { FileText, Send, AlertCircle, CheckCircle2, Clock, XCircle } from 'lucide-react'

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

const docTypes = [
  { code: 'nfe', label: 'NF-e (Nota Fiscal Eletronica)', model: '55' },
  { code: 'nfce', label: 'NFC-e (Nota Fiscal Consumidor)', model: '65' },
  { code: 'cte', label: 'CT-e (Conhecimento de Transporte)', model: '57' },
]

const statusConfig: Record<string, { label: string; color: string; icon: any }> = {
  draft: { label: 'Rascunho', color: 'bg-gray-100 text-gray-600', icon: Clock },
  authorized: { label: 'Autorizada', color: 'bg-green-100 text-green-700', icon: CheckCircle2 },
  rejected: { label: 'Rejeitada', color: 'bg-red-100 text-red-700', icon: XCircle },
  cancelled: { label: 'Cancelada', color: 'bg-orange-100 text-orange-700', icon: XCircle },
}

interface Document { id: string; type: string; number: string; series: string; access_key: string; recipient_name: string; recipient_cnpj: string; total: number; status: string; issued_at: string }

export default function NFePage() {
  const [docs, setDocs] = useState<Document[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ type: 'nfe', recipient_name: '', recipient_cnpj: '', description: '', total: '' })
  const [emitting, setEmitting] = useState(false)

  useEffect(() => {
    const token = getToken()
    if (isDemoMode()) { setDocs([{ id: '1', type: 'nfe', number: '000001', series: '1', access_key: '35260311111111000100550010000000011000000010', recipient_name: 'Cliente Exemplo Ltda', recipient_cnpj: '11.111.111/0001-00', total: 1250.00, status: 'authorized', issued_at: new Date().toISOString() },{ id: '2', type: 'nfce', number: '000002', series: '1', access_key: '35260311111111000100650010000000021000000020', recipient_name: 'Consumidor Final', recipient_cnpj: '000.000.000-00', total: 89.90, status: 'authorized', issued_at: new Date().toISOString() },{ id: '3', type: 'nfe', number: '000003', series: '1', access_key: '35260311111111000100550010000000031000000030', recipient_name: 'Distribuidora ABC', recipient_cnpj: '22.222.222/0001-00', total: 4500.00, status: 'draft', issued_at: new Date().toISOString() }]); setLoading(false); return } if (!token) { promptLogin(); setLoading(false); return }
    fetch('/api/v1/nfe/documents', { headers: { Authorization: `Bearer ${token}` } })
      .then(r => { if (r.status === 401) { window.location.href = '/login'; return null }; return r.json() })
      .then(d => { if (d) { setDocs(d.documents || []); setLoading(false) } })
      .catch(() => setLoading(false))
  }, [])

  const emit = async () => {
    setEmitting(true)
    const token = getToken()
    const res = await fetch('/api/v1/nfe/emit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ ...form, total: parseFloat(form.total) || 0 }),
    })
    if (res.ok) {
      const data = await res.json()
      setDocs(prev => [data.document, ...prev])
      setShowForm(false)
      setForm({ type: 'nfe', recipient_name: '', recipient_cnpj: '', description: '', total: '' })
    }
    setEmitting(false)
  }

  if (loading) return <div className="flex items-center justify-center min-h-[400px]"><div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600" /></div>

  return (
    <div className="max-w-5xl mx-auto py-6 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900" data-testid="nfe-title">Emissao de Documentos Fiscais</h1>
          <p className="text-sm text-gray-500">NF-e, NFC-e e CT-e</p>
        </div>
        <button data-testid="new-nfe-btn" onClick={() => setShowForm(!showForm)} className="px-4 py-2 bg-blue-600 text-white text-sm font-semibold rounded-lg hover:bg-blue-700 flex items-center gap-2">
          <FileText size={16} /> Nova Emissao
        </button>
      </div>

      {showForm && (
        <div className="bg-white rounded-xl border p-6 space-y-4">
          <h3 className="font-semibold">Emitir Documento Fiscal</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">Tipo</label>
              <select data-testid="nfe-type" value={form.type} onChange={e => setForm({ ...form, type: e.target.value })} className="w-full px-3 py-2 border rounded-lg text-sm">
                {docTypes.map(d => <option key={d.code} value={d.code}>{d.label}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">Valor Total (R$)</label>
              <input data-testid="nfe-total" type="number" value={form.total} onChange={e => setForm({ ...form, total: e.target.value })} className="w-full px-3 py-2 border rounded-lg text-sm font-mono" placeholder="0.00" />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">Destinatario</label>
              <input value={form.recipient_name} onChange={e => setForm({ ...form, recipient_name: e.target.value })} className="w-full px-3 py-2 border rounded-lg text-sm" placeholder="Nome/Razao Social" />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">CNPJ/CPF</label>
              <input value={form.recipient_cnpj} onChange={e => setForm({ ...form, recipient_cnpj: e.target.value })} className="w-full px-3 py-2 border rounded-lg text-sm font-mono" placeholder="00.000.000/0000-00" />
            </div>
          </div>
          <div>
            <label className="block text-xs font-semibold text-gray-500 mb-1">Descricao</label>
            <input value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} className="w-full px-3 py-2 border rounded-lg text-sm" placeholder="Descricao do documento" />
          </div>
          <div className="flex justify-end gap-2">
            <button onClick={() => setShowForm(false)} className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700">Cancelar</button>
            <button data-testid="emit-btn" onClick={emit} disabled={emitting} className="px-6 py-2 bg-green-600 text-white text-sm font-semibold rounded-lg hover:bg-green-700 disabled:opacity-50 flex items-center gap-2">
              <Send size={14} /> {emitting ? 'Emitindo...' : 'Emitir'}
            </button>
          </div>
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 flex gap-2">
            <AlertCircle size={16} className="text-amber-500 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-amber-700">Ambiente de homologacao. Os documentos sao simulados e nao tem validade fiscal.</p>
          </div>
        </div>
      )}

      <div className="bg-white rounded-xl border">
        <div className="px-5 py-3 border-b"><h2 className="font-semibold text-gray-900">Documentos Emitidos</h2></div>
        {docs.length === 0 ? (
          <div className="p-8 text-center text-gray-400"><FileText size={32} className="mx-auto mb-2 opacity-50" /><p className="text-sm">Nenhum documento emitido</p></div>
        ) : (
          <div className="divide-y">
            {docs.map(d => {
              const cfg = statusConfig[d.status] || statusConfig.draft
              const Icon = cfg.icon
              return (
                <div key={d.id} data-testid={`doc-${d.id}`} className="px-5 py-4 flex items-center gap-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-mono bg-gray-100 px-2 py-0.5 rounded">{d.type.toUpperCase()}</span>
                      <span className="font-mono text-xs text-gray-400">Serie {d.series} / Nr {d.number}</span>
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium flex items-center gap-1 ${cfg.color}`}><Icon size={11} />{cfg.label}</span>
                    </div>
                    <p className="text-sm font-medium mt-1">{d.recipient_name}</p>
                    <p className="text-xs text-gray-400 font-mono mt-0.5">{d.access_key}</p>
                  </div>
                  <p className="font-mono font-bold">R$ {d.total.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}</p>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

