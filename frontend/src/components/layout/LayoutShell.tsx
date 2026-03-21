'use client'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import Sidebar from '@/components/layout/Sidebar'
import Header from '@/components/layout/Header'

export default function LayoutShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const [isReady, setIsReady] = useState(false)
  
  const isPublicRoute = pathname.startsWith('/login') || pathname.startsWith('/pricing') || pathname.startsWith('/simulador-fiscal')

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = sessionStorage.getItem('access_token')
      const demoMode = localStorage.getItem('nexo_demo_mode') === 'true'
      
      // Se não é rota pública e não tem token nem demo mode, redireciona
      if (!isPublicRoute && !token && !demoMode) {
        router.push('/login')
        return
      }
      
      setIsReady(true)
    }
  }, [pathname, isPublicRoute, router])

  // Rotas públicas sempre sem sidebar
  if (isPublicRoute) {
    return <>{children}</>
  }

  // Aguarda verificação de auth
  if (!isReady) {
    return null
  }

  return (
    <div className="flex h-screen overflow-hidden bg-slate-50">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <Header />
        <main className="flex-1 overflow-y-auto p-6 bg-mesh">
          {children}
        </main>
      </div>
    </div>
  )
}
