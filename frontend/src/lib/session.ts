'use client'

export function getToken(): string {
  if (typeof window === 'undefined') return ''
  return sessionStorage.getItem('access_token') || ''
}

export function requireAuth(): boolean {
  if (typeof window === 'undefined') return true
  const token = getToken()
  if (!token) {
    window.location.href = '/login'
    return false
  }
  return true
}
