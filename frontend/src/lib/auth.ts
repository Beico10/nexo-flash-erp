'use client'
// lib/auth.ts — gerenciamento de tokens JWT no frontend.
//
// SEGURANÇA:
// - access_token: armazenado em memória (sessionStorage) — NUNCA localStorage
//   (localStorage é vulnerável a XSS; sessionStorage é limpo ao fechar a aba)
// - refresh_token: armazenado em cookie HttpOnly pelo servidor
//   (inacessível para JavaScript — protegido contra XSS)
//
// Fluxo de refresh automático:
//   1. Toda request inclui access_token no header
//   2. Se resposta for 401, chama /auth/refresh automaticamente
//   3. Se refresh falhar, redireciona para /login

import { useCallback, useEffect, useRef, useState } from 'react'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export interface AuthUser {
  user_id:       string
  tenant_id:     string
  tenant_slug:   string
  business_type: string
  role:          string
  permissions:   string[]
  expires_at:    string
}

// getAccessToken lê o access_token da memória (sessionStorage).
export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null
  return sessionStorage.getItem('access_token')
}

// setAccessToken salva o access_token na memória.
export function setAccessToken(token: string): void {
  if (typeof window !== 'undefined') {
    sessionStorage.setItem('access_token', token)
  }
}

// clearAccessToken remove o access_token da memória.
export function clearAccessToken(): void {
  if (typeof window !== 'undefined') {
    sessionStorage.removeItem('access_token')
  }
}

// apiFetch é o wrapper de fetch que:
//   1. Adiciona Bearer token automaticamente
//   2. Renova o token se receber 401
//   3. Redireciona para /login se não conseguir renovar
export async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const token = getAccessToken()
  const headers = new Headers(options.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  headers.set('Content-Type', 'application/json')

  const response = await fetch(`${API_URL}${path}`, { ...options, headers })

  // Token expirou — tentar refresh automático
  if (response.status === 401) {
    const refreshed = await tryRefresh()
    if (refreshed) {
      // Retry com novo token
      const newToken = getAccessToken()
      if (newToken) headers.set('Authorization', `Bearer ${newToken}`)
      return fetch(`${API_URL}${path}`, { ...options, headers })
    }
    // Refresh falhou — redirecionar para login
    clearAccessToken()
    window.location.href = '/login'
    return response
  }

  return response
}

// tryRefresh tenta renovar o access_token usando o refresh_token do cookie.
async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${API_URL}/auth/refresh`, {
      method: 'POST',
      credentials: 'include', // inclui o cookie refresh_token
    })
    if (!res.ok) return false
    const data = await res.json()
    setAccessToken(data.access_token)
    return true
  } catch {
    return false
  }
}

// useAuth hook para componentes React.
export function useAuth() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Buscar dados do usuário autenticado
  const fetchMe = useCallback(async () => {
    try {
      const res = await apiFetch('/auth/me')
      if (res.ok) {
        const data = await res.json()
        setUser(data)
        scheduleRefresh(data.expires_at)
      } else {
        setUser(null)
      }
    } catch {
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  // Agendar refresh automático 1 minuto antes do token expirar
  const scheduleRefresh = (expiresAt: string) => {
    if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current)
    const expiresMs = new Date(expiresAt).getTime()
    const refreshIn = expiresMs - Date.now() - 60_000 // 1min antes
    if (refreshIn > 0) {
      refreshTimerRef.current = setTimeout(async () => {
        const ok = await tryRefresh()
        if (ok) fetchMe()
      }, refreshIn)
    }
  }

  useEffect(() => {
    fetchMe()
    return () => {
      if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current)
    }
  }, [fetchMe])

  const logout = async () => {
    await fetch(`${API_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    })
    clearAccessToken()
    window.location.href = '/login'
  }

  const hasPermission = (permission: string): boolean => {
    if (!user) return false
    return user.permissions.includes('*') || user.permissions.includes(permission)
  }

  return { user, loading, logout, hasPermission }
}
