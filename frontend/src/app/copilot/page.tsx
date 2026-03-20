'use client'
import { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, Sparkles, Loader2 } from 'lucide-react'

function getToken() { return typeof window !== 'undefined' ? sessionStorage.getItem('access_token') || '' : '' }

interface Message { role: 'user' | 'assistant'; text: string }

export default function CopilotPage() {
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', text: 'Ola! Sou o Co-Piloto do Nexo One. Posso ajudar com gestao financeira, fiscal (IBS/CBS 2026), estoque, agendamentos e mais. Como posso ajudar?' },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [sessionId] = useState(() => `copilot-${Date.now()}`)
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!getToken()) window.location.href = '/login'
  }, [])

  useEffect(() => { endRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [messages])

  const send = async () => {
    if (!input.trim() || loading) return
    const q = input.trim()
    setInput('')
    setMessages(prev => [...prev, { role: 'user', text: q }])
    setLoading(true)

    try {
      const token = getToken()
      const res = await fetch('/api/v1/copilot/suggest', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ question: q, session_id: sessionId }),
      })
      const data = await res.json()
      setMessages(prev => [...prev, { role: 'assistant', text: data.suggestion || data.error || 'Sem resposta' }])
    } catch {
      setMessages(prev => [...prev, { role: 'assistant', text: 'Erro ao conectar com o Co-Piloto. Tente novamente.' }])
    } finally { setLoading(false) }
  }

  return (
    <div className="max-w-3xl mx-auto py-6 px-4 flex flex-col h-[calc(100vh-80px)]">
      <div className="flex items-center gap-3 mb-4">
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
          <Sparkles size={20} className="text-white" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-gray-900" data-testid="copilot-title">IA Co-Piloto</h1>
          <p className="text-xs text-gray-500">Gemini 3 Flash - Sugestoes inteligentes para seu negocio</p>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto space-y-4 pb-4">
        {messages.map((m, i) => (
          <div key={i} className={`flex gap-3 ${m.role === 'user' ? 'justify-end' : ''}`}>
            {m.role === 'assistant' && (
              <div className="w-7 h-7 rounded-lg bg-indigo-100 flex items-center justify-center flex-shrink-0 mt-1">
                <Bot size={14} className="text-indigo-600" />
              </div>
            )}
            <div data-testid={`msg-${i}`} className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap ${
              m.role === 'user' ? 'bg-blue-600 text-white rounded-br-sm' : 'bg-gray-100 text-gray-800 rounded-bl-sm'
            }`}>
              {m.text}
            </div>
            {m.role === 'user' && (
              <div className="w-7 h-7 rounded-lg bg-blue-100 flex items-center justify-center flex-shrink-0 mt-1">
                <User size={14} className="text-blue-600" />
              </div>
            )}
          </div>
        ))}
        {loading && (
          <div className="flex gap-3">
            <div className="w-7 h-7 rounded-lg bg-indigo-100 flex items-center justify-center"><Bot size={14} className="text-indigo-600" /></div>
            <div className="bg-gray-100 rounded-2xl rounded-bl-sm px-4 py-3"><Loader2 size={16} className="animate-spin text-gray-400" /></div>
          </div>
        )}
        <div ref={endRef} />
      </div>

      <div className="border-t pt-4">
        <div className="flex gap-2">
          <input
            data-testid="copilot-input"
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && send()}
            placeholder="Pergunte ao Co-Piloto..."
            className="flex-1 px-4 py-3 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          />
          <button data-testid="copilot-send" onClick={send} disabled={loading || !input.trim()} className="px-4 py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 disabled:opacity-50">
            <Send size={18} />
          </button>
        </div>
        <div className="flex gap-2 mt-2 flex-wrap">
          {['Como esta meu faturamento?', 'Sugestoes para reduzir impostos', 'Analise meu estoque'].map(s => (
            <button key={s} onClick={() => { setInput(s); }} className="text-xs px-3 py-1.5 bg-gray-50 border rounded-full hover:bg-gray-100 text-gray-600">{s}</button>
          ))}
        </div>
      </div>
    </div>
  )
}
