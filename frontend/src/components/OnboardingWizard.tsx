// BKL-1050: Onboarding wizard — guia passo-a-passo para novos projetos no Pulse.
// Exibido quando não há flags criadas. Persiste progresso em localStorage.
// 4 passos: instalar SDK → criar flag → rastrear evento → ver no dashboard.

import { useState } from 'react'
import { CheckCircle2, Circle, ChevronRight, X } from 'lucide-react'

const STORAGE_KEY = 'pulse_onboarding_dismissed'

const STEPS = [
  {
    id: 'install',
    title: 'Instalar o SDK',
    description: 'Adicione o @iit/pulse-sdk à sua aplicação.',
    code: `npm install @iit/pulse-sdk

import PulseClient from "@iit/pulse-sdk";

const pulse = new PulseClient({
  baseURL: "https://pulse.institutoitinerante.com.br",
});`,
  },
  {
    id: 'flag',
    title: 'Criar sua primeira flag',
    description: 'Use o botão "Nova flag" acima para criar uma feature flag.',
    code: `// Após criar a flag "minha-nova-feature", use:
const flags = await pulse.getFlags();

if (flags["minha-nova-feature"]) {
  // mostrar nova experiência
}`,
  },
  {
    id: 'event',
    title: 'Rastrear um evento',
    description: 'Envie um evento de negócio para medir conversões.',
    code: `// Rastreamento imediato
await pulse.trackEvent({
  event_name: "checkout_completed",
  user_id: currentUser.id,
  product: "meu-produto",
});

// Ou com queue automática (mais eficiente)
pulse.queue({ event_name: "page_view" });`,
  },
  {
    id: 'dashboard',
    title: 'Ver no dashboard',
    description: 'Acompanhe eventos em tempo real em Eventos e experimentos em A/B Testing.',
    code: null,
  },
]

interface Props {
  flagsCount: number
}

export default function OnboardingWizard({ flagsCount }: Props) {
  const [dismissed, setDismissed] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === 'true'
    } catch {
      return false
    }
  })
  const [activeStep, setActiveStep] = useState(0)

  if (dismissed || flagsCount > 0) return null

  // Passos concluídos (simplificado: flags criadas = passou do passo 1)
  const completedSteps = flagsCount > 0 ? [0, 1] : []

  const handleDismiss = () => {
    try {
      localStorage.setItem(STORAGE_KEY, 'true')
    } catch {
      // ignore
    }
    setDismissed(true)
  }

  const step = STEPS[activeStep]

  return (
    <div className="rounded-xl border border-violet-200 dark:border-violet-800 bg-gradient-to-br from-violet-50 to-indigo-50 dark:from-violet-950/30 dark:to-indigo-950/30 p-6 space-y-5">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">
            Primeiros passos com o Pulse
          </h2>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Complete os 4 passos abaixo para começar a usar feature flags e tracking de eventos.
          </p>
        </div>
        <button
          onClick={handleDismiss}
          className="p-1 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
          aria-label="Fechar wizard"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Step navigation */}
      <div className="flex flex-wrap gap-2">
        {STEPS.map((s, i) => {
          const done = completedSteps.includes(i)
          const active = i === activeStep
          return (
            <button
              key={s.id}
              onClick={() => setActiveStep(i)}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium border transition-colors ${
                active
                  ? 'bg-violet-600 text-white border-violet-600'
                  : done
                  ? 'bg-white dark:bg-slate-800 text-violet-600 dark:text-violet-400 border-violet-300 dark:border-violet-700'
                  : 'bg-white dark:bg-slate-800 text-slate-500 border-slate-200 dark:border-slate-700 hover:border-violet-300'
              }`}
            >
              {done ? (
                <CheckCircle2 className="w-3.5 h-3.5 text-green-500" />
              ) : (
                <Circle className="w-3.5 h-3.5" />
              )}
              {i + 1}. {s.title}
            </button>
          )
        })}
      </div>

      {/* Active step content */}
      <div className="space-y-3">
        <div className="space-y-1">
          <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200">
            Passo {activeStep + 1}: {step.title}
          </h3>
          <p className="text-sm text-slate-500 dark:text-slate-400">{step.description}</p>
        </div>

        {step.code && (
          <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono overflow-x-auto">
            {step.code.split('\n').map((line, i) => (
              <p key={i} className={line === '' ? 'h-3' : ''}>
                {line}
              </p>
            ))}
          </div>
        )}

        {step.id === 'dashboard' && (
          <div className="rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950/30 p-4 flex items-start gap-3">
            <CheckCircle2 className="w-5 h-5 text-green-500 shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-green-800 dark:text-green-300">Tudo pronto!</p>
              <p className="text-sm text-green-700 dark:text-green-400 mt-0.5">
                Acesse <strong>Eventos</strong> para ver o stream em tempo real e <strong>A/B Testing</strong> para experimentos com variantes.
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Step navigation buttons */}
      <div className="flex justify-between items-center pt-2">
        <button
          onClick={handleDismiss}
          className="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
        >
          Não mostrar novamente
        </button>
        {activeStep < STEPS.length - 1 && (
          <button
            onClick={() => setActiveStep((s) => s + 1)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-xs font-medium transition-colors"
          >
            Próximo passo
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        )}
      </div>
    </div>
  )
}
