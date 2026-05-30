// BKL-1061: Pricing page pública do Pulse.
// BKL-1085: ROI Calculator interativo.
// Rota: /pricing (sem autenticação — página pública)
import { useState, useEffect } from 'react'
import { Check, Zap } from 'lucide-react'
import { fetchPulsePlans, type PlanResponse } from '../services/billingApi'

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatPriceCents(cents: number): string {
  if (cents === 0) return 'R$ 0'
  return (cents / 100).toLocaleString('pt-BR', {
    style: 'currency',
    currency: 'BRL',
    maximumFractionDigits: 0,
  })
}


// ── Plan card (dados vindos do checkout-service) ───────────────────────────────

function PlanCard({ plan }: { plan: PlanResponse }) {
  const isFree = plan.price_monthly_cents === 0
  const ctaHref = isFree
    ? '/register'
    : `/checkout?plan=${plan.slug}`
  const ctaLabel = isFree ? 'Começar grátis' : `Assinar ${plan.name}`

  return (
    <div
      className={`rounded-2xl border p-6 flex flex-col gap-4 ${
        plan.highlighted
          ? 'border-violet-500 bg-violet-950/30 shadow-lg shadow-violet-500/10'
          : 'border-slate-800 bg-slate-900'
      }`}
    >
      {plan.highlighted && (
        <span className="text-xs font-semibold text-violet-400 uppercase tracking-wider">
          Mais popular
        </span>
      )}

      <div>
        <h2 className="text-xl font-bold">{plan.name}</h2>
        <div className="flex items-baseline gap-1 mt-1">
          <span className="text-3xl font-bold">{formatPriceCents(plan.price_monthly_cents)}</span>
          {!isFree && <span className="text-slate-400 text-sm">/mês</span>}
        </div>
      </div>

      <div className="space-y-2 flex-1">
        {plan.features.map((feat) => (
          <div key={feat} className="flex items-center gap-2 text-sm">
            <Check className="w-4 h-4 text-violet-400 flex-shrink-0" />
            <span className="text-slate-300">{feat}</span>
          </div>
        ))}
      </div>

      <a
        href={ctaHref}
        className={`block text-center py-2.5 px-4 rounded-xl font-semibold text-sm transition-colors ${
          plan.highlighted
            ? 'bg-violet-600 hover:bg-violet-500 text-white'
            : 'bg-slate-800 hover:bg-slate-700 text-slate-200'
        }`}
      >
        {ctaLabel}
      </a>
    </div>
  )
}

// ── Skeleton de loading ────────────────────────────────────────────────────────

function PlanCardSkeleton() {
  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6 flex flex-col gap-4 animate-pulse">
      <div className="h-6 w-24 bg-slate-700 rounded" />
      <div className="h-10 w-32 bg-slate-700 rounded" />
      <div className="space-y-2 flex-1">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-4 bg-slate-700 rounded w-full" />
        ))}
      </div>
      <div className="h-10 bg-slate-700 rounded-xl" />
    </div>
  )
}

// ── ROI Calculator ─────────────────────────────────────────────────────────────

const USER_STEPS = [1_000, 10_000, 50_000, 100_000, 500_000, 1_000_000]

function formatUsers(n: number): string {
  if (n >= 1_000_000) return '1M'
  if (n >= 1_000) return `${n / 1_000}k`
  return String(n)
}

function formatBRL(n: number): string {
  return n.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL', maximumFractionDigits: 0 })
}

interface ROIOutputs {
  hoursSaved: number
  timeSavingsValue: number
  conversionGain: number
  extraRevenue: number | null
  recommendedPlan: { name: string; href: string }
}

function calcROI(flags: number, experiments: number, users: number): ROIOutputs {
  const hoursSaved = flags * 0.5 + experiments * 2
  const timeSavingsValue = hoursSaved * 150
  const conversionGain = Math.min(experiments * 3, 25)
  const extraRevenue = users >= 10_000 ? users * 0.02 * 50 : null

  let recommendedPlan: { name: string; href: string }
  if (flags <= 5 && users <= 10_000) {
    recommendedPlan = { name: 'Free', href: '/register' }
  } else if (flags <= 20 && users <= 500_000) {
    recommendedPlan = { name: 'Starter', href: '/checkout?plan=pulse-starter' }
  } else if (flags <= 50) {
    recommendedPlan = { name: 'Growth', href: '/checkout?plan=pulse-growth' }
  } else {
    recommendedPlan = { name: 'Scale', href: 'mailto:comercial@institutoitinerante.com.br' }
  }

  return { hoursSaved, timeSavingsValue, conversionGain, extraRevenue, recommendedPlan }
}

function ROICalculator() {
  const [flags, setFlags] = useState(10)
  const [experiments, setExperiments] = useState(3)
  const [userStepIndex, setUserStepIndex] = useState(1)

  const users = USER_STEPS[userStepIndex]
  const { hoursSaved, timeSavingsValue, conversionGain, extraRevenue, recommendedPlan } = calcROI(
    flags,
    experiments,
    users,
  )

  return (
    <section
      aria-labelledby="roi-calc-title"
      className="rounded-2xl border border-violet-700 bg-slate-900 p-8 mb-16"
    >
      <div className="text-center mb-8">
        <h2 id="roi-calc-title" className="text-2xl font-bold mb-2">
          Calcule seu ROI com o Pulse
        </h2>
        <p className="text-slate-400 text-sm">
          Estimativas baseadas em métricas de produto reais
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-10">
        <div className="space-y-3">
          <label htmlFor="slider-flags" className="block text-sm font-medium text-slate-300">
            Feature flags gerenciadas
            <span className="ml-2 text-violet-400 font-bold">{flags}</span>
          </label>
          <input
            id="slider-flags"
            type="range"
            min={1}
            max={100}
            step={1}
            value={flags}
            onChange={(e) => setFlags(Number(e.target.value))}
            className="w-full h-2 rounded-full appearance-none cursor-pointer bg-slate-700 accent-violet-500"
            aria-label="Quantidade de feature flags gerenciadas"
          />
          <div className="flex justify-between text-xs text-slate-500">
            <span>1</span>
            <span>100</span>
          </div>
        </div>

        <div className="space-y-3">
          <label htmlFor="slider-experiments" className="block text-sm font-medium text-slate-300">
            Experimentos A/B por mês
            <span className="ml-2 text-violet-400 font-bold">{experiments}</span>
          </label>
          <input
            id="slider-experiments"
            type="range"
            min={1}
            max={20}
            step={1}
            value={experiments}
            onChange={(e) => setExperiments(Number(e.target.value))}
            className="w-full h-2 rounded-full appearance-none cursor-pointer bg-slate-700 accent-violet-500"
            aria-label="Quantidade de experimentos A/B por mês"
          />
          <div className="flex justify-between text-xs text-slate-500">
            <span>1</span>
            <span>20</span>
          </div>
        </div>

        <div className="space-y-3">
          <label htmlFor="slider-users" className="block text-sm font-medium text-slate-300">
            Usuários únicos
            <span className="ml-2 text-violet-400 font-bold">{formatUsers(users)}</span>
          </label>
          <input
            id="slider-users"
            type="range"
            min={0}
            max={USER_STEPS.length - 1}
            step={1}
            value={userStepIndex}
            onChange={(e) => setUserStepIndex(Number(e.target.value))}
            className="w-full h-2 rounded-full appearance-none cursor-pointer bg-slate-700 accent-violet-500"
            aria-label="Quantidade de usuários únicos"
          />
          <div className="flex justify-between text-xs text-slate-500">
            <span>1k</span>
            <span>1M</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="rounded-xl bg-slate-800 border border-slate-700 p-4 text-center">
          <p className="text-xs text-slate-400 mb-1">Horas economizadas/semana</p>
          <p className="text-3xl font-bold text-violet-400">{hoursSaved.toFixed(1)}</p>
          <p className="text-xs text-slate-500 mt-1">horas de dev sênior</p>
        </div>

        <div className="rounded-xl bg-slate-800 border border-slate-700 p-4 text-center">
          <p className="text-xs text-slate-400 mb-1">Valor do tempo economizado</p>
          <p className="text-3xl font-bold text-emerald-400">{formatBRL(timeSavingsValue)}</p>
          <p className="text-xs text-slate-500 mt-1">por semana (R$150/h)</p>
        </div>

        <div className="rounded-xl bg-slate-800 border border-slate-700 p-4 text-center">
          <p className="text-xs text-slate-400 mb-1">Aumento de conversão est.</p>
          <p className="text-3xl font-bold text-violet-400">+{conversionGain}%</p>
          <p className="text-xs text-slate-500 mt-1">média por A/B test</p>
        </div>

        <div className="rounded-xl bg-slate-800 border border-slate-700 p-4 text-center">
          <p className="text-xs text-slate-400 mb-1">Receita adicional est.</p>
          {extraRevenue !== null ? (
            <>
              <p className="text-3xl font-bold text-emerald-400">{formatBRL(extraRevenue)}</p>
              <p className="text-xs text-slate-500 mt-1">ARPU R$50 hipotético</p>
            </>
          ) : (
            <>
              <p className="text-2xl font-bold text-slate-500">—</p>
              <p className="text-xs text-slate-500 mt-1">disponível a partir de 10k users</p>
            </>
          )}
        </div>
      </div>

      <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-6 border-t border-slate-700">
        <div className="flex items-center gap-3">
          <span className="text-slate-400 text-sm">Plano recomendado para você:</span>
          <span className="inline-block bg-violet-600/20 border border-violet-600 text-violet-300 text-sm font-semibold px-3 py-1 rounded-full">
            {recommendedPlan.name}
          </span>
        </div>
        <a
          href={recommendedPlan.href}
          className="bg-violet-600 hover:bg-violet-500 text-white font-semibold py-2.5 px-6 rounded-xl transition-colors text-sm whitespace-nowrap"
        >
          Começar com {recommendedPlan.name}
        </a>
      </div>
    </section>
  )
}

// ── Página principal ───────────────────────────────────────────────────────────

type FetchState =
  | { status: 'loading' }
  | { status: 'error' }
  | { status: 'success'; plans: PlanResponse[] }

export default function Pricing() {
  const [fetchState, setFetchState] = useState<FetchState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false
    fetchPulsePlans()
      .then((plans) => {
        if (!cancelled) setFetchState({ status: 'success', plans })
      })
      .catch(() => {
        if (!cancelled) setFetchState({ status: 'error' })
      })
    return () => { cancelled = true }
  }, [])

  return (
    <div className="min-h-screen bg-slate-950 text-white">
      <header className="border-b border-slate-800 px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center gap-2">
          <div className="w-7 h-7 rounded-lg bg-violet-600 flex items-center justify-center">
            <Zap className="w-4 h-4 text-white" aria-hidden="true" />
          </div>
          <span className="font-bold text-lg">Pulse</span>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-16">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold tracking-tight mb-4">
            Pricing simples e transparente
          </h1>
          <p className="text-slate-400 text-lg max-w-xl mx-auto">
            Feature flags, event tracking e A/B testing num só produto.
            Preços em R$, sem surpresas.
          </p>
        </div>

        {/* Cards de planos */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
          {fetchState.status === 'loading' && (
            <>
              <PlanCardSkeleton />
              <PlanCardSkeleton />
              <PlanCardSkeleton />
              <PlanCardSkeleton />
            </>
          )}

          {fetchState.status === 'error' && (
            <div
              role="alert"
              className="col-span-full text-center py-12 text-slate-400"
            >
              Não foi possível carregar os planos no momento. Tente novamente em instantes ou{' '}
              <a
                href="mailto:comercial@institutoitinerante.com.br"
                className="underline hover:text-white transition-colors"
              >
                fale com a nossa equipe
              </a>
              .
            </div>
          )}

          {fetchState.status === 'success' &&
            fetchState.plans.map((plan) => <PlanCard key={plan.slug} plan={plan} />)}
        </div>

        <ROICalculator />

        {/* FAQ */}
        <div className="max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold text-center mb-8">Perguntas frequentes</h2>
          <div className="space-y-6">
            {[
              {
                q: 'Posso usar o Free indefinidamente?',
                a: 'Sim. Sem limite de tempo nem cartão de crédito. Quando ultrapassar 10k eventos/mês, sugerimos upgrade para Starter.',
              },
              {
                q: 'O que acontece se eu ultrapassar o limite de eventos?',
                a: 'Você paga overage proporcional ao excedente. O serviço não para.',
              },
              {
                q: 'Vocês oferecem desconto anual?',
                a: 'Sim — 2 meses grátis no plano anual (Growth e Scale). Entre em contato.',
              },
              {
                q: 'O SDK funciona com React, Next.js e React Native?',
                a: 'Sim. Temos pacotes para JS/TS, React (hooks), Next.js (SSR guard) e React Native/Expo.',
              },
              {
                q: 'Onde posso hospedar o Pulse?',
                a: 'Managed (nosso homelab, melhor opção) ou self-hosted via Docker Compose (imagem pública disponível).',
              },
            ].map(({ q, a }) => (
              <div key={q} className="border-b border-slate-800 pb-4">
                <h3 className="font-semibold text-slate-200 mb-2">{q}</h3>
                <p className="text-slate-400 text-sm">{a}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="text-center mt-16 p-8 rounded-2xl bg-violet-950/30 border border-violet-800">
          <h2 className="text-2xl font-bold mb-3">Comece hoje mesmo</h2>
          <p className="text-slate-400 mb-6">
            14 dias de trial no Growth. Sem cartão de crédito até você decidir continuar.
          </p>
          <a
            href="/register?trial=growth"
            className="inline-block bg-violet-600 hover:bg-violet-500 text-white font-semibold py-3 px-8 rounded-xl transition-colors"
          >
            Começar trial de 14 dias
          </a>
        </div>
      </main>

      <footer className="border-t border-slate-800 px-6 py-8 text-center text-slate-500 text-sm">
        © 2026 Instituto Itinerante — Pulse v1.0
        &nbsp;·&nbsp;
        <a href="/docs" className="hover:text-slate-300 transition-colors">Documentação</a>
        &nbsp;·&nbsp;
        <a href="/sla" className="hover:text-slate-300 transition-colors">SLA</a>
        &nbsp;·&nbsp;
        <a href="https://suporte.institutoitinerante.com.br?produto=pulse" target="_blank" rel="noopener noreferrer" className="hover:text-slate-300 transition-colors">Suporte</a>
      </footer>
    </div>
  )
}

