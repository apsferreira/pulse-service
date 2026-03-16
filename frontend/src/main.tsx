import React from 'react'
import ReactDOM from 'react-dom/client'
import { Toaster } from 'react-hot-toast'
import App from './App.tsx'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
    <Toaster
      position="top-right"
      toastOptions={{
        duration: 4000,
        style: {
          background: 'var(--toast-bg, #fff)',
          color: 'var(--toast-color, #1C1033)',
          border: '1px solid var(--toast-border, #DDD6FE)',
          fontSize: '14px',
        },
        success: {
          iconTheme: {
            primary: '#7C3AED',
            secondary: '#fff',
          },
          style: {
            background: '#F5F3FF',
            color: '#3B0764',
            border: '1px solid #DDD6FE',
          },
        },
        error: {
          iconTheme: {
            primary: '#EF4444',
            secondary: '#fff',
          },
          style: {
            background: '#FEE2E2',
            color: '#991B1B',
            border: '1px solid #FECACA',
            fontWeight: '500',
          },
        },
      }}
    />
  </React.StrictMode>,
)
