/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Nexo Flash — Paleta Azul Corporativo
        nexo: {
          50:  '#EEF4FF',
          100: '#D9E8FF',
          200: '#B3D0FF',
          300: '#80B0FF',
          400: '#4D8EFF',
          500: '#1A6BFF', // primária
          600: '#0052E0',
          700: '#003DB3',
          800: '#002B80',
          900: '#001A52',
          950: '#000D2E',
        },
        // Tons neutros quentes (não cinza puro — mais sofisticado)
        slate: {
          50:  '#F7F8FC',
          100: '#EEF0F8',
          200: '#D8DCF0',
          300: '#B8BFE0',
          400: '#8892C8',
          500: '#5E6BAD',
          600: '#3D4D8F',
          700: '#2B3870',
          800: '#1C2554',
          900: '#101640',
          950: '#080B22',
        },
        success: '#10B981',
        warning: '#F59E0B',
        danger:  '#EF4444',
        info:    '#3B82F6',
      },
      fontFamily: {
        sans:    ['var(--font-plus-jakarta)', 'system-ui', 'sans-serif'],
        display: ['var(--font-syne)', 'system-ui', 'sans-serif'],
        mono:    ['var(--font-jetbrains)', 'monospace'],
      },
      borderRadius: {
        '4xl': '2rem',
      },
      boxShadow: {
        'nexo-sm':  '0 1px 3px 0 rgba(26, 107, 255, 0.08), 0 1px 2px -1px rgba(26, 107, 255, 0.05)',
        'nexo-md':  '0 4px 6px -1px rgba(26, 107, 255, 0.1), 0 2px 4px -2px rgba(26, 107, 255, 0.06)',
        'nexo-lg':  '0 10px 15px -3px rgba(26, 107, 255, 0.1), 0 4px 6px -4px rgba(26, 107, 255, 0.06)',
        'nexo-xl':  '0 20px 25px -5px rgba(26, 107, 255, 0.12), 0 8px 10px -6px rgba(26, 107, 255, 0.08)',
        'glow':     '0 0 20px rgba(26, 107, 255, 0.35)',
        'glow-lg':  '0 0 40px rgba(26, 107, 255, 0.25)',
      },
      backgroundImage: {
        'nexo-gradient': 'linear-gradient(135deg, #1A6BFF 0%, #003DB3 100%)',
        'nexo-dark':     'linear-gradient(135deg, #001A52 0%, #000D2E 100%)',
        'mesh':          'radial-gradient(at 40% 20%, hsla(220,100%,60%,0.08) 0px, transparent 50%), radial-gradient(at 80% 0%, hsla(220,100%,40%,0.06) 0px, transparent 50%)',
      },
      animation: {
        'fade-in':     'fadeIn 0.4s ease-out',
        'slide-up':    'slideUp 0.4s ease-out',
        'slide-right': 'slideRight 0.3s ease-out',
        'pulse-blue':  'pulseBlue 2s ease-in-out infinite',
        'spin-slow':   'spin 3s linear infinite',
      },
      keyframes: {
        fadeIn:    { from: { opacity: '0' }, to: { opacity: '1' } },
        slideUp:   { from: { opacity: '0', transform: 'translateY(12px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
        slideRight:{ from: { opacity: '0', transform: 'translateX(-12px)' }, to: { opacity: '1', transform: 'translateX(0)' } },
        pulseBlue: { '0%, 100%': { boxShadow: '0 0 0 0 rgba(26,107,255,0.4)' }, '50%': { boxShadow: '0 0 0 8px rgba(26,107,255,0)' } },
      },
    },
  },
  plugins: [],
}
