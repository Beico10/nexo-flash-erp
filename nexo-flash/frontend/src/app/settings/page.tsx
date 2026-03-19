'use client'
import { useState } from 'react'
import { Settings, User, Building2, Shield, Bell, CreditCard, Save, Plus, Trash2, Eye, EyeOff } from 'lucide-react'

const tabs = [
  { key: 'empresa', label: 'Empresa', icon: Building2 },
  { key: 'usuarios', label: 'Usuários', icon: User },
  { key: 'seguranca', label: 'Segurança', icon: Shield },
  { key: 'notificacoes', label: 'Notificações', icon: Bell },
  { key: 'plano', label: 'Plano', icon: CreditCard },
]

const users = [
  { id: '1', name: 'Antonio (você)', email: 'antonio@mecanica.com', role: 'owner',    active: true },
  { id: '2', name: 'Carlos Mecânico', email: 'carlos@mecanica.com', role: 'operator', active: true },
  { id: '3', name: 'Maria Recepção',  email: 'maria@mecanica.com',  role: 'viewer',   active: true },
]

const roleLabels: Record<string, { label: string; cls: string }> = {
  owner:    { label: 'Proprietário', cls: 'badge-danger' },
  admin:    { label: 'Admin',        cls: 'badge-open' },
  operator: { label: 'Operador',     cls: 'badge-approved' },
  viewer:   { label: 'Visualizador', cls: 'badge-done' },
}

export default function SettingsPage() {
  const [tab, setTab] = useState('empresa')
  const [saved, setSaved] = useState(false)
  const [showKey, setShowKey] = useState(false)

  const save = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="flex gap-6 animate-fade-in" style={{ minHeight: 'calc(100vh - 10rem)' }}>

      {/* Sidebar de tabs */}
      <div className="w-52 flex-shrink-0">
        <div className="card p-2 space-y-0.5">
          {tabs.map(t => {
            const Icon = t.icon
            const active = tab === t.key
            return (
              <button key={t.key} onClick={() => setTab(t.key)}
                className={active ? 'nav-item-active w-full' : 'nav-item w-full'}>
                <Icon size={15} strokeWidth={active ? 2.5 : 2} />
                {t.label}
              </button>
            )
          })}
        </div>
      </div>

      {/* Conteúdo */}
      <div className="flex-1 space-y-5">

        {/* ── Empresa ── */}
        {tab === 'empresa' && (
          <div className="space-y-4 animate-fade-in">
            <div className="card p-6">
              <h2 style={{ fontSize: 15, fontWeight: 700, color: '#0D1B4B', marginBottom: 20 }}>
                Dados da Empresa
              </h2>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Nome da empresa</label>
                  <input className="input" defaultValue="Mecânica do João" />
                </div>
                <div>
                  <label className="label">Identificador (slug)</label>
                  <div className="relative">
                    <input className="input" defaultValue="mecanica-joao" style={{ paddingRight: 140 }} />
                    <span style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', fontSize: 11, color: '#8892B8', fontFamily: 'var(--font-jetbrains)', background: 'rgba(26,51,120,0.05)', padding: '2px 7px', borderRadius: 6 }}>
                      .nexoflash.com
                    </span>
                  </div>
                </div>
                <div>
                  <label className="label">CNPJ</label>
                  <input className="input" defaultValue="12.345.678/0001-90" />
                </div>
                <div>
                  <label className="label">Tipo de negócio</label>
                  <select className="input">
                    <option value="mechanic">Mecânica</option>
                    <option value="bakery">Padaria</option>
                    <option value="industry">Indústria</option>
                    <option value="logistics">Logística</option>
                    <option value="aesthetics">Estética</option>
                    <option value="shoes">Calçados</option>
                  </select>
                </div>
                <div>
                  <label className="label">Fuso horário</label>
                  <select className="input">
                    <option>America/Sao_Paulo</option>
                    <option>America/Manaus</option>
                    <option>America/Belem</option>
                    <option>America/Fortaleza</option>
                  </select>
                </div>
                <div>
                  <label className="label">URL base do sistema</label>
                  <input className="input" defaultValue="https://app.nexoflash.com.br" />
                </div>
              </div>
              <div className="flex justify-end mt-5">
                <button onClick={save} className={saved ? 'btn-success' : 'btn-primary'}>
                  <Save size={14} />{saved ? 'Salvo!' : 'Salvar alterações'}
                </button>
              </div>
            </div>

            {/* Módulos habilitados */}
            <div className="card p-6">
              <h2 style={{ fontSize: 15, fontWeight: 700, color: '#0D1B4B', marginBottom: 4 }}>Módulos Add-on</h2>
              <p style={{ fontSize: 12, color: '#8892B8', marginBottom: 20 }}>Habilite módulos extras além do nicho principal</p>
              <div className="space-y-3">
                {[
                  { label: 'Roteirizador Inteligente', desc: 'Otimização de rotas com DRE da viagem', active: true, price: 'Incluído no Pro' },
                  { label: 'BaaS PIX + Boleto', desc: 'PIX dinâmico e boleto híbrido', active: true, price: 'Incluído no Pro' },
                  { label: 'IA Concierge', desc: 'Onboarding via XML NF-e', active: true, price: 'Incluído no Pro' },
                  { label: 'Relatórios Fiscais SPED', desc: 'Exportação IBS/CBS para SEFAZ', active: false, price: '+R$ 49/mês' },
                ].map(m => (
                  <div key={m.label} className="flex items-center justify-between p-3 rounded-xl" style={{ background: '#FAFBFF', border: '1px solid rgba(26,51,120,0.08)' }}>
                    <div>
                      <p style={{ fontSize: 13, fontWeight: 600, color: '#0D1B4B' }}>{m.label}</p>
                      <p style={{ fontSize: 11, color: '#8892B8', marginTop: 2 }}>{m.desc}</p>
                    </div>
                    <div className="flex items-center gap-3">
                      <span style={{ fontSize: 11, color: m.active ? '#047857' : '#8892B8', fontWeight: 600 }}>{m.price}</span>
                      <div style={{ width: 36, height: 20, borderRadius: 10, background: m.active ? '#1A47C8' : 'rgba(26,51,120,0.12)', cursor: 'pointer', display: 'flex', alignItems: 'center', padding: 2 }}>
                        <div style={{ width: 16, height: 16, borderRadius: '50%', background: '#fff', marginLeft: m.active ? 'auto' : 0, boxShadow: '0 1px 3px rgba(0,0,0,0.2)', transition: 'all 0.2s' }} />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* ── Usuários ── */}
        {tab === 'usuarios' && (
          <div className="space-y-4 animate-fade-in">
            <div className="card overflow-hidden">
              <div style={{ padding: '16px 20px', borderBottom: '1px solid rgba(26,51,120,0.07)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div>
                  <p style={{ fontSize: 14, fontWeight: 700, color: '#0D1B4B' }}>Usuários do sistema</p>
                  <p style={{ fontSize: 11, color: '#8892B8', marginTop: 2 }}>Gerencie quem tem acesso ao Nexo Flash</p>
                </div>
                <button className="btn-primary"><Plus size={14} />Convidar usuário</button>
              </div>
              <table className="w-full">
                <thead style={{ background: '#FAFBFF', borderBottom: '1px solid rgba(26,51,120,0.07)' }}>
                  <tr>
                    <th className="table-header">Nome</th>
                    <th className="table-header">E-mail</th>
                    <th className="table-header">Perfil</th>
                    <th className="table-header text-center">Status</th>
                    <th className="table-header text-center">Ações</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map(u => {
                    const r = roleLabels[u.role]
                    return (
                      <tr key={u.id} className="table-row-hover">
                        <td className="table-cell">
                          <div className="flex items-center gap-2.5">
                            <div style={{ width: 28, height: 28, borderRadius: 8, background: 'linear-gradient(135deg,#1A47C8,#0F2D8A)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                              <span style={{ fontSize: 10, fontWeight: 700, color: '#fff' }}>{u.name.charAt(0)}</span>
                            </div>
                            <span style={{ fontSize: 13, fontWeight: 600, color: '#0D1B4B' }}>{u.name}</span>
                          </div>
                        </td>
                        <td className="table-cell" style={{ fontSize: 12 }}>{u.email}</td>
                        <td className="table-cell"><span className={r.cls}>{r.label}</span></td>
                        <td className="table-cell text-center">
                          <span className={u.active ? 'badge-approved' : 'badge-done'}>
                            {u.active ? 'Ativo' : 'Inativo'}
                          </span>
                        </td>
                        <td className="table-cell text-center">
                          {u.role !== 'owner' && (
                            <button className="btn-ghost py-1 px-2 text-xs">
                              <Trash2 size={12} />
                            </button>
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* ── Segurança ── */}
        {tab === 'seguranca' && (
          <div className="space-y-4 animate-fade-in">
            <div className="card p-6">
              <h2 style={{ fontSize: 15, fontWeight: 700, color: '#0D1B4B', marginBottom: 20 }}>Alterar senha</h2>
              <div className="space-y-4 max-w-sm">
                <div>
                  <label className="label">Senha atual</label>
                  <input type="password" className="input" placeholder="••••••••" />
                </div>
                <div>
                  <label className="label">Nova senha</label>
                  <input type="password" className="input" placeholder="••••••••" />
                </div>
                <div>
                  <label className="label">Confirmar nova senha</label>
                  <input type="password" className="input" placeholder="••••••••" />
                </div>
                <button className="btn-primary"><Save size={14} />Alterar senha</button>
              </div>
            </div>

            <div className="card p-6">
              <h2 style={{ fontSize: 15, fontWeight: 700, color: '#0D1B4B', marginBottom: 4 }}>Chave da API</h2>
              <p style={{ fontSize: 12, color: '#8892B8', marginBottom: 16 }}>Use para integrar o Nexo Flash com sistemas externos</p>
              <div className="flex items-center gap-3">
                <div className="input flex-1 font-mono text-xs" style={{ background: '#FAFBFF', cursor: 'default' }}>
                  {showKey ? 'nxf_live_k9x2m4p8q1r5t7w3y6z0a' : '••••••••••••••••••••••••••••••'}
                </div>
                <button onClick={() => setShowKey(v => !v)} className="btn-ghost py-2.5">
                  {showKey ? <EyeOff size={15} /> : <Eye size={15} />}
                </button>
                <button className="btn-secondary py-2.5 text-xs">Regenerar</button>
              </div>
              <p style={{ fontSize: 11, color: '#C53030', marginTop: 8, fontWeight: 500 }}>
                ⚠ Nunca compartilhe sua chave. Em caso de vazamento, regenere imediatamente.
              </p>
            </div>
          </div>
        )}

        {/* ── Plano ── */}
        {tab === 'plano' && (
          <div className="space-y-4 animate-fade-in">
            <div className="card-premium p-6">
              <p style={{ fontSize: 11, fontWeight: 700, color: 'rgba(255,255,255,0.6)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>Plano atual</p>
              <p style={{ fontFamily: 'var(--font-syne)', fontSize: 28, fontWeight: 800, color: '#fff', marginTop: 4 }}>Pro</p>
              <p style={{ fontSize: 13, color: 'rgba(255,255,255,0.7)', marginTop: 4 }}>Todos os módulos incluídos · R$ 197/mês</p>
              <div className="flex items-center gap-3 mt-4">
                <span style={{ fontSize: 11, background: 'rgba(255,255,255,0.15)', color: '#fff', padding: '3px 10px', borderRadius: 20, fontWeight: 600 }}>Renovação em 12/04/2026</span>
              </div>
            </div>
            <div className="card p-6">
              <h2 style={{ fontSize: 15, fontWeight: 700, color: '#0D1B4B', marginBottom: 16 }}>O que está incluído</h2>
              <div className="grid grid-cols-2 gap-3">
                {['6 módulos de negócio', 'Motor IBS/CBS 2026', 'IA Human-in-the-Loop', 'BaaS PIX + Boleto', 'Suporte prioritário', 'Updates automáticos', 'Backup diário', 'SSL + segurança'].map(f => (
                  <div key={f} className="flex items-center gap-2">
                    <div style={{ width: 16, height: 16, borderRadius: '50%', background: '#ECFDF5', border: '1px solid rgba(4,120,87,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                      <span style={{ fontSize: 8, color: '#047857', fontWeight: 900 }}>✓</span>
                    </div>
                    <span style={{ fontSize: 12, color: '#4A5680' }}>{f}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

      </div>
    </div>
  )
}
