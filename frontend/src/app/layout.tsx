import type { Metadata } from 'next'
import './globals.css'
import LayoutShell from '@/components/layout/LayoutShell'

export const metadata: Metadata = {
  title: 'Nexo One ERP',
  description: 'ERP SaaS Multi-Nicho — Reforma Tributaria Brasil 2026',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR">
      <body>
        <LayoutShell>{children}</LayoutShell>
      </body>
    </html>
  )
}
