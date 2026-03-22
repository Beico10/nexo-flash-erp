'use client'
import { useState, useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'

// Rotas que não precisam do layout (sidebar/topbar)
const NO_LAYOUT_ROUTES = ['/', '/login', '/cadastro', '/checkout']

// Mapa de rotas do sidebar para rotas reais existentes no sistema
// Se a rota não existe, redireciona para o módulo mais próximo
const ROUTE_MAP: Record<string, string> = {
  '/copilot':    '/dashboard',
  '/tax-engine': '/dashboard',
  '/simulator':  '/dashboard',
  '/nfe':        '/dashboard',
  '/payments':   '/dashboard',
  '/enterprise': '/dashboard',
}

// Modal login suave — aparece APENAS quando usuário tenta SALVAR/CRIAR algo
function LoginModal({ onClose }: { onClose: () => void }) {
  const router = useRouter()
  return (
    <div
      style={{ position: 'fixed', inset: 0, zIndex: 999, background: 'rgba(0,0,0,0.45)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}
      onClick={onClose}
    >
      <div onClick={e => e.stopPropagation()} style={{ background: 'white', borderRadius: 20, padding: 32, width: '100%', maxWidth: 400, textAlign: 'center' }}>
        <div style={{ width: 52, height: 52, background: '#F0F4FF', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 16px' }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
            <path d="M12 2C9.24 2 7 4.24 7 7s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5zm0 12c-5.33 0-8 2.67-8 4v2h16v-2c0-1.33-2.67-4-8-4z" fill="#0A3D8F" opacity="0.7"/>
          </svg>
        </div>
        <p style={{ fontSize: 18, fontWeight: 800, color: '#1C1917', margin: '0 0 8px' }}>Gostou do sistema?</p>
        <p style={{ fontSize: 13, color: '#64748b', lineHeight: 1.6, margin: '0 0 22px' }}>
          Para salvar e usar este recurso de verdade,<br />crie sua conta gratuita. Leva menos de 1 minuto.
        </p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <button onClick={() => router.push('/cadastro')} style={{ width: '100%', padding: 13, borderRadius: 10, border: 'none', background: '#0A3D8F', color: 'white', fontSize: 14, fontWeight: 700, cursor: 'pointer' }}>
            Criar conta grátis — 7 dias sem cartão
          </button>
          <button onClick={() => router.push('/login')} style={{ width: '100%', padding: 11, borderRadius: 10, border: '1px solid #e0e4f0', background: 'white', color: '#475569', fontSize: 13, cursor: 'pointer' }}>
            Já tenho conta — Entrar
          </button>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: 12, cursor: 'pointer', padding: 6 }}>
            Continuar explorando sem login
          </button>
        </div>
      </div>
    </div>
  )
}

// Sidebar organizado em grupos
const SIDEBAR_GROUPS = [
  {
    label: 'PRINCIPAL',
    items: [{ href: '/dashboard', label: 'Dashboard' }],
  },
  {
    label: 'OPERACIONAL',
    items: [
      { href: '/mechanic',   label: 'Mecânica',   nicho: 'mechanic' },
      { href: '/bakery',     label: 'Padaria',     nicho: 'bakery' },
      { href: '/industry',   label: 'Indústria',   nicho: 'industry' },
      { href: '/logistics',  label: 'Logística',   nicho: 'logistics' },
      { href: '/aesthetics', label: 'Estética',    nicho: 'aesthetics' },
      { href: '/shoes',      label: 'Calçados',    nicho: 'shoes' },
    ],
  },
  {
    label: 'INTELIGÊNCIA',
    items: [
      { href: '/ai-approvals', label: 'Aprovações IA' },
    ],
  },
  {
    label: 'SISTEMA',
    items: [
      { href: '/settings', label: 'Configurações' },
    ],
  },
]

function Sidebar({ businessType }: { businessType: string }) {
  const router = useRouter()
  const pathname = usePathname()
  const [collapsed, setCollapsed] = useState<string[]>([])

  const toggleGroup = (label: string) => {
    setCollapsed(prev => prev.includes(label) ? prev.filter(l => l !== label) : [...prev, label])
  }

  const handleNav = (href: string) => {
    // Se a rota não existe, usa o fallback
    const target = ROUTE_MAP[href] || href
    router.push(target)
  }

  return (
    <div style={{ width: 200, flexShrink: 0, background: 'white', borderRight: '0.5px solid #e8e8e8', height: '100vh', overflowY: 'auto', display: 'flex', flexDirection: 'column', position: 'sticky', top: 0, minHeight: 0 }}>
      {/* Logo */}
      <div style={{ padding: '14px 14px 10px', borderBottom: '0.5px solid #f0f0f0' }}>
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
          <div style={{ width: 28, height: 28, background: '#E3F2FD', borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 700, color: '#0A3D8F', flexShrink: 0 }}>M</div>
          <div>
            <div style={{ fontSize: 12, fontWeight: 600, color: '#1C1917' }}>Mecânica do João</div>
            <div style={{ fontSize: 10, color: '#94a3b8' }}>Modo demonstração</div>
          </div>
          <div style={{ width: 6, height: 6, borderRadius: '50%', background: '#16A34A', marginLeft: 'auto', flexShrink: 0 }} />
        </div>
      </div>

      {/* Grupos */}
      <div style={{ flex: 1, padding: '8px 0', overflowY: 'auto' }}>
        {SIDEBAR_GROUPS.map(group => {
          const isCollapsed = collapsed.includes(group.label)
          // Filtrar por nicho — só mostra o nicho do usuário
          const items = group.items.filter((item: any) => !item.nicho || item.nicho === businessType)
          if (items.length === 0) return null

          return (
            <div key={group.label} style={{ marginBottom: 2 }}>
              <button onClick={() => toggleGroup(group.label)} style={{ width: '100%', background: 'none', border: 'none', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '6px 14px 4px', cursor: 'pointer' }}>
                <span style={{ fontSize: 9, fontWeight: 700, color: '#94a3b8', letterSpacing: '0.1em' }}>{group.label}</span>
                <span style={{ fontSize: 10, color: '#cbd5e1', transform: isCollapsed ? 'rotate(-90deg)' : 'rotate(0)', transition: 'transform 0.2s' }}>▾</span>
              </button>
              {!isCollapsed && items.map(item => {
                const isActive = pathname === item.href
                return (
                  <button key={item.href} onClick={() => handleNav(item.href)} style={{ width: '100%', background: isActive ? '#F0F4FF' : 'none', border: 'none', textAlign: 'left', padding: '7px 14px 7px 20px', cursor: 'pointer', borderLeft: isActive ? '2px solid #0A3D8F' : '2px solid transparent' }}>
                    <span style={{ fontSize: 13, color: isActive ? '#0A3D8F' : '#475569', fontWeight: isActive ? 600 : 400 }}>{item.label}</span>
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

// Layout Shell principal
export default function LayoutShell({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const [showLoginModal, setShowLoginModal] = useState(false)
  const [mounted, setMounted] = useState(false)

  // Expor função global para módulos chamarem quando usuário tenta ação real
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).showLoginPrompt = () => setShowLoginModal(true)
    }
  }, [])

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!mounted) return

    // Rotas sem layout — não faz nada
    if (NO_LAYOUT_ROUTES.includes(pathname || '')) return

    const token = localStorage.getItem('nexo_token') || ''
    const isDemoMode = localStorage.getItem('nexo_demo_mode') === 'true'
    const isDemoToken = token === 'demo-token' || token.startsWith('demo')

    // Demo mode: navega LIVREMENTE — nunca redireciona
    if (isDemoMode || isDemoToken) return

    // Sem nenhum token: volta para entrada
    if (!token) {
      router.push('/')
    }
  }, [pathname, mounted])

  if (!mounted) return null

  // Páginas sem layout (entrada, login, cadastro, checkout)
  if (NO_LAYOUT_ROUTES.includes(pathname || '')) {
    return <>{children}</>
  }

  const businessType = localStorage.getItem('nexo_business_type') || 'mechanic'

  return (
    <div style={{ display: 'flex', minHeight: '100vh', background: '#F8F7F4' }}>
      {showLoginModal && <LoginModal onClose={() => setShowLoginModal(false)} />}

      <Sidebar businessType={businessType} />

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {/* Topbar */}
        <div style={{ background: 'white', borderBottom: '0.5px solid #e8e8e8', padding: '10px 20px', display: 'flex', alignItems: 'center', gap: 12, position: 'sticky', top: 0, zIndex: 10 }}>
          {pathname !== '/dashboard' && (
            <button onClick={() => router.back()} style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 12, color: '#0A3D8F', cursor: 'pointer', fontWeight: 600 }}>
              ← Voltar
            </button>
          )}
          <div style={{ flex: 1 }} />
          <div style={{ background: '#F0F4FF', border: '1px solid #dde3f0', borderRadius: 8, padding: '5px 12px', fontSize: 11, color: '#0A3D8F', fontWeight: 600 }}>
            ⚡ IBS/CBS 2026 ativo
          </div>
        </div>

        {children}
      </div>
    </div>
  )
}

