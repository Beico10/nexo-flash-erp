'use client'
import { useState } from 'react'
import Sidebar from '@/components/layout/Sidebar'
import { Zap, Key, Building2, Webhook, Plus, Copy, Eye, EyeOff, Trash2, Check, AlertCircle, Globe, Shield, Database, FileText } from 'lucide-react'

type Tab = 'api-keys' | 'branches' | 'webhooks'

interface ApiKey {
  id: string
  name: string
  key: string
  scopes: string[]
  createdAt: string
  lastUsed: string | null
  status: 'active' | 'revoked'
}

interface Branch {
  id: string
  name: string
  cnpj: string
  city: string
  state: string
  status: 'active' | 'inactive'
}

interface WebhookConfig {
  id: string
  name: string
  url: string
  events: string[]
  status: 'active' | 'inactive'
  lastTriggered: string | null
}

const SCOPES = [
  { id: 'nfe:read', label: 'NF-e Leitura', icon: FileText },
  { id: 'nfe:write', label: 'NF-e Escrita', icon: FileText },
  { id: 'inventory:read', label: 'Estoque Leitura', icon: Database },
  { id: 'inventory:write', label: 'Estoque Escrita', icon: Database },
  { id: 'finance:read', label: 'Financeiro Leitura', icon: Shield },
  { id: 'finance:write', label: 'Financeiro Escrita', icon: Shield },
  { id: 'webhook:manage', label: 'Gerenciar Webhooks', icon: Webhook },
]

const WEBHOOK_EVENTS = [
  'nfe.emitted', 'nfe.cancelled', 'order.created', 'order.updated',
  'payment.received', 'payment.overdue', 'inventory.low', 'inventory.updated'
]

export default function EnterprisePage() {
  const [activeTab, setActiveTab] = useState<Tab>('api-keys')
  const [showCreateKey, setShowCreateKey] = useState(false)
  const [showCreateBranch, setShowCreateBranch] = useState(false)
  const [showCreateWebhook, setShowCreateWebhook] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set())

  // Form states
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyScopes, setNewKeyScopes] = useState<string[]>([])
  const [newBranch, setNewBranch] = useState({ name: '', cnpj: '', city: '', state: '' })
  const [newWebhook, setNewWebhook] = useState({ name: '', url: '', events: [] as string[] })

  // Demo data
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([
    { id: '1', name: 'Integração ERP', key: 'nxo_live_a1b2c3d4e5f6g7h8i9j0', scopes: ['nfe:read', 'nfe:write', 'inventory:read'], createdAt: '2024-01-15', lastUsed: '2024-03-20', status: 'active' },
    { id: '2', name: 'App Mobile', key: 'nxo_live_z9y8x7w6v5u4t3s2r1q0', scopes: ['inventory:read', 'finance:read'], createdAt: '2024-02-01', lastUsed: '2024-03-19', status: 'active' },
  ])

  const [branches, setBranches] = useState<Branch[]>([
    { id: '1', name: 'Matriz São Paulo', cnpj: '12.345.678/0001-90', city: 'São Paulo', state: 'SP', status: 'active' },
    { id: '2', name: 'Filial Rio de Janeiro', cnpj: '12.345.678/0002-71', city: 'Rio de Janeiro', state: 'RJ', status: 'active' },
    { id: '3', name: 'Filial Belo Horizonte', cnpj: '12.345.678/0003-52', city: 'Belo Horizonte', state: 'MG', status: 'inactive' },
  ])

  const [webhooks, setWebhooks] = useState<WebhookConfig[]>([
    { id: '1', name: 'Sistema Contábil', url: 'https://contabil.exemplo.com/webhook', events: ['nfe.emitted', 'nfe.cancelled'], status: 'active', lastTriggered: '2024-03-20 14:32' },
    { id: '2', name: 'Notificações Slack', url: 'https://hooks.slack.com/services/xxx', events: ['payment.received', 'inventory.low'], status: 'active', lastTriggered: '2024-03-19 09:15' },
  ])

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const toggleKeyVisibility = (id: string) => {
    const newVisible = new Set(visibleKeys)
    if (newVisible.has(id)) newVisible.delete(id)
    else newVisible.add(id)
    setVisibleKeys(newVisible)
  }

  const createApiKey = () => {
    if (!newKeyName || newKeyScopes.length === 0) return
    const newKey: ApiKey = {
      id: Date.now().toString(),
      name: newKeyName,
      key: `nxo_live_${Math.random().toString(36).substring(2, 22)}`,
      scopes: newKeyScopes,
      createdAt: new Date().toISOString().split('T')[0],
      lastUsed: null,
      status: 'active'
    }
    setApiKeys([newKey, ...apiKeys])
    setNewKeyName('')
    setNewKeyScopes([])
    setShowCreateKey(false)
  }

  const createBranch = () => {
    if (!newBranch.name || !newBranch.cnpj) return
    const branch: Branch = {
      id: Date.now().toString(),
      ...newBranch,
      status: 'active'
    }
    setBranches([...branches, branch])
    setNewBranch({ name: '', cnpj: '', city: '', state: '' })
    setShowCreateBranch(false)
  }

  const createWebhook = () => {
    if (!newWebhook.name || !newWebhook.url || newWebhook.events.length === 0) return
    const webhook: WebhookConfig = {
      id: Date.now().toString(),
      ...newWebhook,
      status: 'active',
      lastTriggered: null
    }
    setWebhooks([...webhooks, webhook])
    setNewWebhook({ name: '', url: '', events: [] })
    setShowCreateWebhook(false)
  }

  const revokeKey = (id: string) => {
    setApiKeys(apiKeys.map(k => k.id === id ? { ...k, status: 'revoked' as const } : k))
  }

  const tabs = [
    { id: 'api-keys' as Tab, label: 'API Keys', icon: Key, count: apiKeys.filter(k => k.status === 'active').length },
    { id: 'branches' as Tab, label: 'Filiais', icon: Building2, count: branches.filter(b => b.status === 'active').length },
    { id: 'webhooks' as Tab, label: 'Webhooks', icon: Webhook, count: webhooks.filter(w => w.status === 'active').length },
  ]

  return (
    <div className="flex min-h-screen bg-[#F8F9FC]">
      <Sidebar />
      <main className="flex-1 p-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg, #1A47C8 0%, #0F2D8A 100%)' }}>
              <Zap size={20} className="text-white" />
            </div>
            <div>
              <h1 style={{ fontFamily: 'var(--font-syne)', fontWeight: 700, fontSize: 24, color: '#0D1B4B' }}>Enterprise</h1>
              <p style={{ fontSize: 14, color: '#8892B8' }}>Gerencie integrações, filiais e configurações avançadas</p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6 p-1 bg-white rounded-xl border border-slate-200 w-fit">
          {tabs.map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all ${
                activeTab === tab.id
                  ? 'bg-[#1A47C8] text-white shadow-sm'
                  : 'text-slate-600 hover:bg-slate-50'
              }`}
            >
              <tab.icon size={16} />
              {tab.label}
              <span className={`px-2 py-0.5 rounded-full text-xs ${
                activeTab === tab.id ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-600'
              }`}>
                {tab.count}
              </span>
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="bg-white rounded-2xl border border-slate-200 shadow-sm">
          {/* API Keys Tab */}
          {activeTab === 'api-keys' && (
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-lg font-semibold text-slate-800">Chaves de API</h2>
                  <p className="text-sm text-slate-500">Gerencie suas chaves de acesso à API</p>
                </div>
                <button
                  onClick={() => setShowCreateKey(true)}
                  className="flex items-center gap-2 px-4 py-2.5 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] transition-colors"
                >
                  <Plus size={16} />
                  Nova Chave
                </button>
              </div>

              {/* Create Key Modal */}
              {showCreateKey && (
                <div className="mb-6 p-5 bg-slate-50 rounded-xl border border-slate-200">
                  <h3 className="font-semibold text-slate-800 mb-4">Criar Nova Chave</h3>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-1">Nome da Chave</label>
                      <input
                        type="text"
                        value={newKeyName}
                        onChange={e => setNewKeyName(e.target.value)}
                        placeholder="Ex: Integração ERP"
                        className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-2">Escopos de Acesso</label>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                        {SCOPES.map(scope => (
                          <label
                            key={scope.id}
                            className={`flex items-center gap-2 p-3 rounded-xl border cursor-pointer transition-all ${
                              newKeyScopes.includes(scope.id)
                                ? 'border-[#1A47C8] bg-[#1A47C8]/5'
                                : 'border-slate-200 hover:border-slate-300'
                            }`}
                          >
                            <input
                              type="checkbox"
                              checked={newKeyScopes.includes(scope.id)}
                              onChange={e => {
                                if (e.target.checked) setNewKeyScopes([...newKeyScopes, scope.id])
                                else setNewKeyScopes(newKeyScopes.filter(s => s !== scope.id))
                              }}
                              className="sr-only"
                            />
                            <scope.icon size={14} className={newKeyScopes.includes(scope.id) ? 'text-[#1A47C8]' : 'text-slate-400'} />
                            <span className={`text-xs font-medium ${newKeyScopes.includes(scope.id) ? 'text-[#1A47C8]' : 'text-slate-600'}`}>
                              {scope.label}
                            </span>
                          </label>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2 pt-2">
                      <button
                        onClick={createApiKey}
                        disabled={!newKeyName || newKeyScopes.length === 0}
                        className="px-4 py-2 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Criar Chave
                      </button>
                      <button
                        onClick={() => { setShowCreateKey(false); setNewKeyName(''); setNewKeyScopes([]) }}
                        className="px-4 py-2 bg-slate-100 text-slate-600 rounded-xl text-sm font-medium hover:bg-slate-200"
                      >
                        Cancelar
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Keys List */}
              <div className="space-y-3">
                {apiKeys.map(key => (
                  <div
                    key={key.id}
                    className={`p-4 rounded-xl border ${key.status === 'revoked' ? 'bg-slate-50 border-slate-200 opacity-60' : 'bg-white border-slate-200'}`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <span className="font-semibold text-slate-800">{key.name}</span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                            key.status === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
                          }`}>
                            {key.status === 'active' ? 'Ativa' : 'Revogada'}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 mb-3">
                          <code className="px-3 py-1.5 bg-slate-100 rounded-lg text-sm font-mono text-slate-600">
                            {visibleKeys.has(key.id) ? key.key : key.key.substring(0, 12) + '••••••••••••'}
                          </code>
                          <button
                            onClick={() => toggleKeyVisibility(key.id)}
                            className="p-1.5 text-slate-400 hover:text-slate-600"
                          >
                            {visibleKeys.has(key.id) ? <EyeOff size={14} /> : <Eye size={14} />}
                          </button>
                          <button
                            onClick={() => copyToClipboard(key.key, key.id)}
                            className="p-1.5 text-slate-400 hover:text-slate-600"
                          >
                            {copiedId === key.id ? <Check size={14} className="text-emerald-500" /> : <Copy size={14} />}
                          </button>
                        </div>
                        <div className="flex flex-wrap gap-1.5 mb-2">
                          {key.scopes.map(scope => (
                            <span key={scope} className="px-2 py-1 bg-[#1A47C8]/10 text-[#1A47C8] rounded-lg text-xs font-medium">
                              {scope}
                            </span>
                          ))}
                        </div>
                        <div className="flex gap-4 text-xs text-slate-500">
                          <span>Criada: {key.createdAt}</span>
                          <span>Último uso: {key.lastUsed || 'Nunca'}</span>
                        </div>
                      </div>
                      {key.status === 'active' && (
                        <button
                          onClick={() => revokeKey(key.id)}
                          className="p-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-lg"
                        >
                          <Trash2 size={16} />
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Branches Tab */}
          {activeTab === 'branches' && (
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-lg font-semibold text-slate-800">Filiais</h2>
                  <p className="text-sm text-slate-500">Gerencie as filiais da sua empresa</p>
                </div>
                <button
                  onClick={() => setShowCreateBranch(true)}
                  className="flex items-center gap-2 px-4 py-2.5 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] transition-colors"
                >
                  <Plus size={16} />
                  Nova Filial
                </button>
              </div>

              {/* Create Branch Modal */}
              {showCreateBranch && (
                <div className="mb-6 p-5 bg-slate-50 rounded-xl border border-slate-200">
                  <h3 className="font-semibold text-slate-800 mb-4">Adicionar Filial</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-1">Nome</label>
                      <input
                        type="text"
                        value={newBranch.name}
                        onChange={e => setNewBranch({ ...newBranch, name: e.target.value })}
                        placeholder="Filial Centro"
                        className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-1">CNPJ</label>
                      <input
                        type="text"
                        value={newBranch.cnpj}
                        onChange={e => setNewBranch({ ...newBranch, cnpj: e.target.value })}
                        placeholder="00.000.000/0000-00"
                        className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-1">Cidade</label>
                      <input
                        type="text"
                        value={newBranch.city}
                        onChange={e => setNewBranch({ ...newBranch, city: e.target.value })}
                        placeholder="São Paulo"
                        className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-1">Estado</label>
                      <input
                        type="text"
                        value={newBranch.state}
                        onChange={e => setNewBranch({ ...newBranch, state: e.target.value })}
                        placeholder="SP"
                        maxLength={2}
                        className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                      />
                    </div>
                  </div>
                  <div className="flex gap-2 pt-4">
                    <button
                      onClick={createBranch}
                      disabled={!newBranch.name || !newBranch.cnpj}
                      className="px-4 py-2 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Adicionar Filial
                    </button>
                    <button
                      onClick={() => { setShowCreateBranch(false); setNewBranch({ name: '', cnpj: '', city: '', state: '' }) }}
                      className="px-4 py-2 bg-slate-100 text-slate-600 rounded-xl text-sm font-medium hover:bg-slate-200"
                    >
                      Cancelar
                    </button>
                  </div>
                </div>
              )}

              {/* Branches List */}
              <div className="overflow-hidden rounded-xl border border-slate-200">
                <table className="w-full">
                  <thead className="bg-slate-50">
                    <tr>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">Filial</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">CNPJ</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">Localização</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">Status</th>
                      <th className="text-right px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">Ações</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {branches.map(branch => (
                      <tr key={branch.id} className="hover:bg-slate-50">
                        <td className="px-4 py-4">
                          <div className="flex items-center gap-3">
                            <div className="w-9 h-9 rounded-xl bg-[#1A47C8]/10 flex items-center justify-center">
                              <Building2 size={16} className="text-[#1A47C8]" />
                            </div>
                            <span className="font-medium text-slate-800">{branch.name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-4 text-sm text-slate-600 font-mono">{branch.cnpj}</td>
                        <td className="px-4 py-4 text-sm text-slate-600">{branch.city}, {branch.state}</td>
                        <td className="px-4 py-4">
                          <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                            branch.status === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'
                          }`}>
                            {branch.status === 'active' ? 'Ativa' : 'Inativa'}
                          </span>
                        </td>
                        <td className="px-4 py-4 text-right">
                          <button className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg">
                            <Trash2 size={14} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Webhooks Tab */}
          {activeTab === 'webhooks' && (
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-lg font-semibold text-slate-800">Webhooks</h2>
                  <p className="text-sm text-slate-500">Configure endpoints para receber eventos em tempo real</p>
                </div>
                <button
                  onClick={() => setShowCreateWebhook(true)}
                  className="flex items-center gap-2 px-4 py-2.5 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] transition-colors"
                >
                  <Plus size={16} />
                  Novo Webhook
                </button>
              </div>

              {/* Create Webhook Modal */}
              {showCreateWebhook && (
                <div className="mb-6 p-5 bg-slate-50 rounded-xl border border-slate-200">
                  <h3 className="font-semibold text-slate-800 mb-4">Configurar Webhook</h3>
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">Nome</label>
                        <input
                          type="text"
                          value={newWebhook.name}
                          onChange={e => setNewWebhook({ ...newWebhook, name: e.target.value })}
                          placeholder="Minha Integração"
                          className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">URL do Endpoint</label>
                        <input
                          type="url"
                          value={newWebhook.url}
                          onChange={e => setNewWebhook({ ...newWebhook, url: e.target.value })}
                          placeholder="https://api.exemplo.com/webhook"
                          className="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-[#1A47C8] focus:ring-2 focus:ring-[#1A47C8]/20 outline-none"
                        />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-2">Eventos</label>
                      <div className="flex flex-wrap gap-2">
                        {WEBHOOK_EVENTS.map(event => (
                          <label
                            key={event}
                            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border cursor-pointer transition-all text-sm ${
                              newWebhook.events.includes(event)
                                ? 'border-[#1A47C8] bg-[#1A47C8]/5 text-[#1A47C8]'
                                : 'border-slate-200 hover:border-slate-300 text-slate-600'
                            }`}
                          >
                            <input
                              type="checkbox"
                              checked={newWebhook.events.includes(event)}
                              onChange={e => {
                                if (e.target.checked) setNewWebhook({ ...newWebhook, events: [...newWebhook.events, event] })
                                else setNewWebhook({ ...newWebhook, events: newWebhook.events.filter(ev => ev !== event) })
                              }}
                              className="sr-only"
                            />
                            {event}
                          </label>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2 pt-2">
                      <button
                        onClick={createWebhook}
                        disabled={!newWebhook.name || !newWebhook.url || newWebhook.events.length === 0}
                        className="px-4 py-2 bg-[#1A47C8] text-white rounded-xl text-sm font-medium hover:bg-[#0F2D8A] disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Criar Webhook
                      </button>
                      <button
                        onClick={() => { setShowCreateWebhook(false); setNewWebhook({ name: '', url: '', events: [] }) }}
                        className="px-4 py-2 bg-slate-100 text-slate-600 rounded-xl text-sm font-medium hover:bg-slate-200"
                      >
                        Cancelar
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Webhooks List */}
              <div className="space-y-3">
                {webhooks.map(webhook => (
                  <div key={webhook.id} className="p-4 rounded-xl border border-slate-200 bg-white">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <Globe size={16} className="text-[#1A47C8]" />
                          <span className="font-semibold text-slate-800">{webhook.name}</span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                            webhook.status === 'active' ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'
                          }`}>
                            {webhook.status === 'active' ? 'Ativo' : 'Inativo'}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 mb-3">
                          <code className="px-3 py-1.5 bg-slate-100 rounded-lg text-sm font-mono text-slate-600 truncate max-w-md">
                            {webhook.url}
                          </code>
                          <button
                            onClick={() => copyToClipboard(webhook.url, `wh-${webhook.id}`)}
                            className="p-1.5 text-slate-400 hover:text-slate-600"
                          >
                            {copiedId === `wh-${webhook.id}` ? <Check size={14} className="text-emerald-500" /> : <Copy size={14} />}
                          </button>
                        </div>
                        <div className="flex flex-wrap gap-1.5 mb-2">
                          {webhook.events.map(event => (
                            <span key={event} className="px-2 py-1 bg-amber-100 text-amber-700 rounded-lg text-xs font-medium">
                              {event}
                            </span>
                          ))}
                        </div>
                        <div className="text-xs text-slate-500">
                          Último disparo: {webhook.lastTriggered || 'Nunca'}
                        </div>
                      </div>
                      <button className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg">
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              {/* Info Box */}
              <div className="mt-6 p-4 bg-blue-50 rounded-xl border border-blue-200 flex items-start gap-3">
                <AlertCircle size={18} className="text-blue-600 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-blue-800">Como funcionam os webhooks?</p>
                  <p className="text-sm text-blue-700 mt-1">
                    Quando um evento selecionado ocorre, enviamos uma requisição POST para a URL configurada com os dados do evento em formato JSON.
                    Certifique-se de que seu endpoint responde com status 200 em até 30 segundos.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
