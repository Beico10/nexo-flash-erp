/** @type {import('next').NextConfig} */
const nextConfig = {
  // Proxy todas as chamadas /api/* para o backend Go (porta 8002)
  // Isso resolve o problema de CORS e conecta o frontend ao backend
  async rewrites() {
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8002'
    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ]
  },

  // Configurações de produção
  output: 'standalone',
  
  // Permite imagens do domínio do sistema
  images: {
    domains: ['localhost'],
  },

  // Desabilitar telemetria
  telemetry: false,
}

module.exports = nextConfig
