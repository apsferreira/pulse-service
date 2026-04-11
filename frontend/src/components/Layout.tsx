import React, { useState, createContext, useContext } from 'react'
import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Activity, Zap, Moon, Sun, HeartPulse, Users, FlaskConical, DollarSign, Star, ChevronDown } from 'lucide-react'
import { useProjects } from '../hooks/useProjects'
import { useProjectContext } from '../hooks/useProjectContext'

interface LayoutProps {
  children: React.ReactNode
}

// BKL-1076: Contexto global de projeto — consume em qualquer página para filtrar dados.
export interface ProjectCtxValue {
  projectId: string | null
  projectSlug: string | null
  setProjectId: (id: string | null) => void
}

export const ProjectCtx = createContext<ProjectCtxValue>({
  projectId: null,
  projectSlug: null,
  setProjectId: () => {},
})

function useTheme() {
  const [dark, setDark] = React.useState(() => {
    if (typeof window === 'undefined') return false
    const stored = localStorage.getItem('pulse-theme')
    if (stored) return stored === 'dark'
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  React.useEffect(() => {
    if (dark) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('pulse-theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('pulse-theme', 'light')
    }
  }, [dark])

  return { dark, toggle: () => setDark((d) => !d) }
}

export function useProject() {
  return useContext(ProjectCtx)
}

export default function Layout({ children }: LayoutProps) {
  const { dark, toggle } = useTheme()
  const [mobileOpen, setMobileOpen] = useState(false)
  const { projects } = useProjects()
  const { projectId, setProjectId } = useProjectContext()
  const [selectorOpen, setSelectorOpen] = useState(false)

  const selectedProject = projects.find((p) => p.id === projectId)
  const projectSlug = selectedProject?.slug ?? null

  const navItems = [
    { to: '/', label: 'Feature Flags', icon: LayoutDashboard, end: true },
    { to: '/uptime', label: 'Uptime', icon: HeartPulse },
    { to: '/events', label: 'Eventos', icon: Activity },
    { to: '/retention', label: 'Retencao', icon: Users }, // BKL-095
    { to: '/ab-testing', label: 'A/B Testing', icon: FlaskConical }, // BKL-124
    { to: '/costs', label: 'Custos', icon: DollarSign },
    { to: '/csat', label: 'CSAT', icon: Star },
  ]

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    [
      'flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
      isActive
        ? 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300'
        : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800',
    ].join(' ')

  return (
    <ProjectCtx.Provider value={{ projectId, projectSlug, setProjectId }}>
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100">
      {/* Top bar */}
      <header className="sticky top-0 z-40 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 shadow-soft">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-14">
            {/* Logo */}
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-lg bg-violet-600 flex items-center justify-center">
                <Zap className="w-4 h-4 text-white" />
              </div>
              <span className="font-bold text-lg tracking-tight">
                Pulse
                <span className="ml-1.5 text-xs font-medium text-violet-500 dark:text-violet-400 align-middle">
                  Admin
                </span>
              </span>
            </div>

            {/* Desktop nav */}
            <nav className="hidden md:flex items-center gap-1">
              {navItems.map(({ to, label, icon: Icon, end }) => (
                <NavLink key={to} to={to} end={end} className={navLinkClass}>
                  <Icon className="w-4 h-4" />
                  {label}
                </NavLink>
              ))}
            </nav>

            {/* Right actions */}
            <div className="flex items-center gap-2">
              {/* BKL-1076: Project selector */}
              {projects.length > 0 && (
                <div className="relative">
                  <button
                    onClick={() => setSelectorOpen((v) => !v)}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                  >
                    <span className="max-w-[140px] truncate">
                      {selectedProject?.name ?? 'Todos os projetos'}
                    </span>
                    <ChevronDown className="w-3.5 h-3.5 flex-shrink-0" />
                  </button>
                  {selectorOpen && (
                    <div
                      className="absolute right-0 mt-1 w-56 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 shadow-lg z-50 overflow-hidden"
                      onBlur={() => setSelectorOpen(false)}
                    >
                      <div className="p-1">
                        <button
                          onClick={() => { setProjectId(null); setSelectorOpen(false) }}
                          className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                            projectId === null
                              ? 'bg-violet-50 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300 font-medium'
                              : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
                          }`}
                        >
                          Todos os projetos
                        </button>
                        {projects.map((p) => (
                          <button
                            key={p.id}
                            onClick={() => { setProjectId(p.id); setSelectorOpen(false) }}
                            className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                              projectId === p.id
                                ? 'bg-violet-50 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300 font-medium'
                                : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800'
                            }`}
                          >
                            <span className="block truncate">{p.name}</span>
                            <span className="text-xs text-slate-400 dark:text-slate-500">{p.plan}</span>
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              <button
                onClick={toggle}
                className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors"
                aria-label="Alternar tema"
              >
                {dark ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
              </button>
              {/* Mobile hamburger */}
              <button
                onClick={() => setMobileOpen((v) => !v)}
                className="md:hidden p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                aria-label="Menu"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        {/* Mobile nav */}
        {mobileOpen && (
          <div className="md:hidden border-t border-slate-200 dark:border-slate-800 px-4 py-2 space-y-1">
            {navItems.map(({ to, label, icon: Icon, end }) => (
              <NavLink
                key={to}
                to={to}
                end={end}
                className={navLinkClass}
                onClick={() => setMobileOpen(false)}
              >
                <Icon className="w-4 h-4" />
                {label}
              </NavLink>
            ))}
          </div>
        )}
      </header>

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {children}
      </main>
    </div>
    </ProjectCtx.Provider>
  )
}
