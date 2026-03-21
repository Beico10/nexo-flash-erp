'use client'
import { useState, useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'

// ── MODAL LOGIN SUAVE ─────────────────────────────────────────────────────────

function LoginModal({ onClose }: { onClose: () => void }) {
  const router = useRouter()
  return (
    <div style={{
      position: 'fixed', inset: 0, zIndex: 999,
      background: 'rgba(0,0,0,0.45)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      padding: 24,
    }}
      onClick={onClose}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          background: 'white', borderRadius: 20, padding: 32,
          width: '100%', maxWidth: 420,
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
        }}
      >
        {/* Ícone */}
        <div style={{ textAlign: 'center', marginBottom: 20 }}>
          <div style={{ width: 56, height: 56, background: '#F0F4FF', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M12 2C9.24 2 7 4.24 7 7s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zm0 12c-5.33 0-8 2.67-8 4v2h16v-2c0-1.33-2.67-4-8-4z" fill="#0A3D8F" opacity="0.7"/>
            </svg>
          </div>
          <p style={{ fontSize: 18, fontWeight: 800, color: '#1C1917', margin: '0 0 8px', letterSpacing: -0.5 }}>
            Faça login para continuar
          </p>
          <p style={{ fontSize: 13, color: '#64748b', lineHeight: 1.5, margin: 0 }}>
            Você está explorando o sistema.<br />
            Para usar este recurso, crie sua conta gratuita.
          </p>
        </div>

        {/* Botões */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <button
            onClick={() => router.push('/login?trial=true')}
            style={{
              width: '100%', padding: '14px', borderRadius: 10, border: 'none',
              background: '#0A3D8F', color: 'white',
              fontSize: 14, fontWeight: 700, cursor: 'pointer',
            }}
          >
            Criar conta grátis — 7 dias sem cartão
          </button>
          <button
            onClick={() => router.push('/login')}
            style={{
              width: '100%', padding: '12px', borderRadius: 10,
              border: '1px solid #e0e4f0', background: 'white',
              fontSize: 13, fontWeight: 600, color: '#475569', cursor: 'pointer',
            }}
          >
            Já tenho conta — Entrar
          </button>
          <button
            onClick={onClose}
            style={{
              background: 'none', border: 'none', color: '#94a3b8',
              fontSize: 12, cursor: 'pointer', padding: '6px',
            }}
          >
            Continuar explorando sem login
          </button>
        </div>
      </div>
    </div>
  )
}

// ── SIDEBAR ───────────────────────────────────────────────────────────────────

const SIDEBAR_GROUPS = [
  {
    label: 'PRINCIPAL',
    items: [
      { href: '/dashboard', label: 'Dashboard' },
    ],
  },
  {
    label: 'FINANCEIRO',
    items: [
      { href: '/finance', label: 'Visão Geral' },
      { href: '/payables', label: 'Contas a Pagar' },
      { href: '/receivables', label: 'Contas a Receber' },
      { href: '/expenses', label: 'Despesas' },
    ],
  },
  {
    label: 'OPERACIONAL',
    items: [
      { href: '/mechanic', label: 'Mecânica', nicho: 'mechanic' },
      { href: '/bakery', label: 'Padaria', nicho: 'bakery' },
      { href: '/industry', label: 'Indústria', nicho: 'industry' },
      { href: '/logistics', label: 'Logística', nicho: 'logistics' },
      { href: '/aesthetics', label: 'Estética', nicho: 'aesthetics' },
      { href: '/shoes', label: 'Calçados', nicho: 'shoes' },
      { href: '/inventory', label: 'Estoque' },
      { href: '/dispatch', label: 'Despacho em Lote' },
    ],
  },
  {
    label: 'FISCAL',
    items: [
      { href: '/tax-engine', label: 'Motor IBS/CBS 2026' },
      { href: '/nfe', label: 'Emissão NF-e' },
      { href: '/simulator', label: 'Simulador Fiscal' },
    ],
  },
  {
    label: 'INTELIGÊNCIA',
    items: [
      { href: '/copilot', label: 'Co-Piloto IA' },
      { href: '/ai-approvals', label: 'Aprovações IA' },
    ],
  },
  {
    label: 'SISTEMA',
    items: [
      { href: '/modules', label: 'Módulos' },
      { href: '/payments', label: 'Pagamentos' },
      { href: '/settings', label: 'Configurações' },
      { href: '/enterprise', label: 'Enterprise' },
    ],
  },
]

// Rotas que requerem login real
const PROTECTED_ROUTES = [
  '/payables', '/receivables', '/finance', '/inventory',
  '/dispatch', '/nfe', '/enterprise', '/payments', '/settings',
]

function Sidebar({ businessType }: { businessType: string }) {
  const router = useRouter()
  const pathname = usePathname()
  const [collapsed, setCollapsed] = useState<string[]>([])

  const toggleGroup = (label: string) => {
    setCollapsed(prev =>
      prev.includes(label)
        ? prev.filter(l => l !== label)
        : [...prev, label]
    )
  }

  return (
    <div style={{
      width: 200, flexShrink: 0, background: 'white',
      borderRight: '0.5px solid #e8e8e8',
      height: '100vh', overflowY: 'auto',
      display: 'flex', flexDirection: 'column',
      position: 'sticky', top: 0,
    }}>
      {/* Logo */}
      <div style={{ padding: '16px 16px 12px', borderBottom: '0.5px solid #f0f0f0' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 28, height: 28, background: '#0A3D8F', borderRadius: 7, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
            <svg width="14" height="14" viewBox="0 0 64 64">
              <line x1="20" y1="20" x2="32" y2="28" stroke="white" strokeWidth="3"/>
              <line x1="32" y1="28" x2="44" y2="20" stroke="white" strokeWidth="3"/>
              <line x1="16" y1="34" x2="32" y2="28" stroke="white" strokeWidth="3"/>
              <line x1="32" y1="28" x2="48" y2="34" stroke="white" strokeWidth="3"/>
              <line x1="24" y1="36" x2="32" y2="44" stroke="white" strokeWidth="3"/>
              <line x1="40" y1="36" x2="32" y2="44" stroke="white" strokeWidth="3"/>
              <circle cx="32" cy="28" r="5" fill="white"/>
              <circle cx="32" cy="44" r="4" fill="white"/>
              <circle cx="20" cy="20" r="3" fill="white"/>
              <circle cx="44" cy="20" r="3" fill="white"/>
            </svg>
          </div>
          <div>
            <div style={{ fontSize: 12, fontWeight: 700, color: '#1C1917', lineHeight: 1.2 }}>Gestão Para</div>
            <div style={{ fontSize: 12, fontWeight: 700, color: '#0A3D8F', lineHeight: 1.2 }}>Todos</div>
          </div>
        </div>
      </div>

      {/* Tenant */}
      <div style={{ padding: '10px 14px', borderBottom: '0.5px solid #f0f0f0' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 28, height: 28, background: '#E3F2FD', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 13, fontWeight: 700, color: '#0A3D8F', flexShrink: 0 }}>M</div>
          <div>
            <div style={{ fontSize: 12, fontWeight: 600, color: '#1C1917' }}>Mecânica do João</div>
            <div style={{ fontSize: 10, color: '#94a3b8' }}>Plano Pro</div>
          </div>
          <div style={{ width: 6, height: 6, borderRadius: '50%', background: '#16A34A', marginLeft: 'auto', flexShrink: 0 }} />
        </div>
      </div>

      {/* Grupos */}
      <div style={{ flex: 1, padding: '8px 0', overflowY: 'auto' }}>
        {SIDEBAR_GROUPS.map(group => {
          const isCollapsed = collapsed.includes(group.label)

          // Filtrar itens por nicho
          const items = group.items.filter(item => {
            if (!item.nicho) return true
            return item.nicho === businessType
          })

          if (items.length === 0) return null

          return (
            <div key={group.label} style={{ marginBottom: 4 }}>
              {/* Header do grupo */}
              <button
                onClick={() => toggleGroup(group.label)}
                style={{
                  width: '100%', background: 'none', border: 'none',
                  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                  padding: '6px 14px 4px', cursor: 'pointer',
                }}
              >
                <span style={{ fontSize: 9, fontWeight: 700, color: '#94a3b8', letterSpacing: '0.1em' }}>
                  {group.label}
                </span>
                <span style={{ fontSize: 10, color: '#cbd5e1', transform: isCollapsed ? 'rotate(-90deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }}>▾</span>
              </button>

              {/* Itens */}
              {!isCollapsed && items.map(item => {
                const isActive = pathname === item.href || pathname?.startsWith(item.href + '/')
                return (
                  <button
                    key={item.href}
                    onClick={() => router.push(item.href)}
                    style={{
                      width: '100%', background: isActive ? '#F0F4FF' : 'none',
                      border: 'none', textAlign: 'left',
                      padding: '7px 14px 7px 20px', cursor: 'pointer',
                      borderLeft: isActive ? '2px solid #0A3D8F' : '2px solid transparent',
                    }}
                  >
                    <span style={{ fontSize: 13, color: isActive ? '#0A3D8F' : '#475569', fontWeight: isActive ? 600 : 400 }}>
                      {item.label}
                    </span>
                  </button>
                )
              })}
            </div>
          )
        })}
      </div>

      {/* Footer */}
      <div style={{ padding: '12px 14px', borderTop: '0.5px solid #f0f0f0' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{ width: 28, height: 28, background: '#0A3D8F', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 700, color: 'white', flexShrink: 0 }}>AN</div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 12, fontWeight: 600, color: '#1C1917', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>Antonio</div>
            <div style={{ fontSize: 10, color: '#94a3b8' }}>Proprietário</div>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── LAYOUT SHELL ──────────────────────────────────────────────────────────────

export default function LayoutShell({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const [showLoginModal, setShowLoginModal] = useState(false)
  const [mounted, setMounted] = useState(false)

  const isLoginPage = pathname === '/login' || pathname === '/cadastro' || pathname === '/' || pathname === '/checkout'

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!mounted) return
    if (isLoginPage) return

    const token = localStorage.getItem('nexo_token') || ''
    const isDemoMode = localStorage.getItem('nexo_demo_mode') === 'true'
    const isDemoToken = token === 'demo-token' || token.startsWith('demo')

    // Demo mode ou demo token: navega LIVREMENTE por todas as telas
    // O modal só aparece quando o usuário chama showLoginPrompt() em ações reais
    if (isDemoMode || isDemoToken) return

    // Sem token nenhum: redirecionar para entrada
    if (!token) {
      router.push('/')
      return
    }
  }, [pathname, mounted])

  const businessType = mounted
    ? localStorage.getItem('nexo_business_type') || 'mechanic'
    : 'mechanic'

  if (!mounted) return null
  if (isLoginPage) return <>{children}</>

  return (
    <div style={{ display: 'flex', minHeight: '100vh', background: '#F8F7F4' }}>
      {/* Modal login suave */}
      {showLoginModal && (
        <LoginModal onClose={() => setShowLoginModal(false)} />
      )}

      {/* Sidebar */}
      <Sidebar businessType={businessType} />

      {/* Conteúdo */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {/* Topbar com botão voltar */}
        <div style={{ background: 'white', borderBottom: '0.5px solid #e8e8e8', padding: '10px 20px', display: 'flex', alignItems: 'center', gap: 12, position: 'sticky', top: 0, zIndex: 10 }}>
          {pathname !== '/dashboard' && (
            <button
              onClick={() => router.back()}
              style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 12, color: '#0A3D8F', cursor: 'pointer', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 4, flexShrink: 0 }}
            >
              ← Voltar
            </button>
          )}
          <div style={{ flex: 1 }} />
          <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#0A3D8F', fontWeight: 600 }}>
            ⚡ IBS/CBS 2026 ativo
          </div>
        </div>

        {/* Página */}
        {children}
      </div>
    </div>
  )
}
