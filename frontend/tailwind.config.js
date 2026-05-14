/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        pulse: {
          brand: {
            primary: '#7C3AED',
            'primary-light': '#9D68F0',
            'primary-dark': '#5B21B6',
            secondary: '#A78BFA',
          },
          surface: {
            base: '#FFFFFF',
            subtle: '#F5F3FF',
            elevated: '#EDE9FE',
          },
          'surface-dark': {
            base: '#0D0B14',
            subtle: '#14111F',
            elevated: '#1E1A2E',
          },
          text: {
            primary: '#1C1033',
            secondary: '#4C3B6B',
            muted: '#8B7AAB',
            'on-brand': '#FFFFFF',
          },
          border: {
            default: '#DDD6FE',
            subtle: '#EDE9FE',
            brand: '#7C3AED',
          },
          status: {
            success: '#10B981',
            'success-bg': '#D1FAE5',
            warning: '#F59E0B',
            'warning-bg': '#FEF3C7',
            error: '#EF4444',
            'error-bg': '#FEE2E2',
            info: '#7C3AED',
            'info-bg': '#EDE9FE',
            loading: '#A78BFA',
            'loading-bg': '#F5F3FF',
          },
          toggle: {
            on: '#10B981',
            'on-track': '#D1FAE5',
            off: '#CBD5E1',
            'off-track': '#F1F5F9',
            loading: '#A78BFA',
          },
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'spin-slow': 'spin 2s linear infinite',
        'flag-flash-on': 'flagFlashOn 700ms ease-out',
        'flag-flash-off': 'flagFlashOff 700ms ease-out',
        'slide-down': 'slideDown 0.18s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        flagFlashOn: {
          '0%':   { backgroundColor: 'transparent' },
          '25%':  { backgroundColor: 'rgb(209 250 229 / 0.5)' },
          '100%': { backgroundColor: 'transparent' },
        },
        flagFlashOff: {
          '0%':   { backgroundColor: 'transparent' },
          '25%':  { backgroundColor: 'rgb(241 245 249 / 0.7)' },
          '100%': { backgroundColor: 'transparent' },
        },
        slideDown: {
          '0%':   { transform: 'translateY(-4px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
      boxShadow: {
        'soft': '0 2px 8px rgba(0, 0, 0, 0.08)',
        'soft-lg': '0 4px 16px rgba(0, 0, 0, 0.12)',
        'pulse': '0 2px 8px rgba(124, 58, 237, 0.15)',
        'flag-active': '0 0 0 3px rgba(16, 185, 129, 0.12)',
        'flag-focus': '0 0 0 3px rgba(124, 58, 237, 0.18)',
      },
    },
  },
  plugins: [],
}
