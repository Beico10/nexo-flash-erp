// app/api/auth/login/route.ts
// Proxy de login — o Next.js faz a chamada para o backend Go,
// pega o refresh_token do header Set-Cookie e o repassa ao browser.
// Isso garante que o cookie seja HttpOnly no domínio do frontend.

import { NextRequest, NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8002'

export async function POST(req: NextRequest) {
  const body = await req.json()

  const backendRes = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  const data = await backendRes.json()

  if (!backendRes.ok) {
    return NextResponse.json(data, { status: backendRes.status })
  }

  // Criar resposta repassando o Set-Cookie do backend
  const response = NextResponse.json(data, { status: 200 })

  // Repassar o cookie de refresh_token do backend para o browser
  const setCookie = backendRes.headers.get('set-cookie')
  if (setCookie) {
    response.headers.set('Set-Cookie', setCookie)
  }

  return response
}
