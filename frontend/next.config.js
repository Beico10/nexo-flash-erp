/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',

  // Proxy todas as chamadas /api/* para o backend Go (porta 8001)
  // Isso resolve o problema de CORS e conecta o frontend ao backend
  async rewrites() {
    const backendUrl = process.env.API_URL || 'http://localhost:8080'
    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ]
  },

  // Permite imagens do domínio do sistema
  images: {
    domains: ['localhost'],
  },
}

module.exports = nextConfig
