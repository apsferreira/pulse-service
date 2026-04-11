// BKL-1061: Pricing page pública do Pulse.
// Rota: /pricing (sem autenticação — página pública)
import { Check, X, Zap } from 'lucide-react'

interface Plan {
  name: string
  price: string
  priceNote?: string
  description: string
  events: string
  flags: string
  projects: string
  abTesting: string | false
  history: string
  sla: string
  support: string
  extras: string[]
  cta: string
  ctaHref: string
  highlighted?: boolean
}

const plans: Plan[] = [
  {
    name: 'Free',
    price: 'R$ 0',
    description: 'Para protótipos e MVPs. Sem cartão de crédito.',
    events: '10.000/mês',
    flags: 'até 5',
    projects: '1',
    abTesting: false,
    history: '7 dias',
    sla: 'Best effort',
    support: 'Community',
    extras: [],
    cta: 'Começar grátis',
    ctaHref: '/register',
  },
  {
    name: 'Starter',
    price: 'R$ 399',
    priceNote: '/mês',
    description: 'Para validação com primeiros usuários reais.',
    events: '500.000/mês',
    flags: 'Ilimitadas',
    projects: 'até 3',
    abTesting: 'Básico (2 variantes)',
    history: '30 dias',
    sla: '99,5% uptime',
    support: 'Email 24h',
    extras: ['Rollout percentual', 'Ambientes prod/staging/dev'],
    cta: 'Assinar Starter',
    ctaHref: '/checkout?plan=starter',
  },
  {
    name: 'Growth',
    price: 'R$ 999',
    priceNote: '/mês',
    description: 'Para startups em crescimento acelerado.',
    events: '2.000.000/mês',
    flags: 'Ilimitadas',
    projects: 'até 10',
    abTesting: 'Avançado + chi-squared',
    history: '90 dias',
    sla: '99,5% uptime',
    support: 'Email prioritário',
    extras: ['CSAT / NPS tracking', 'Audit log de flags', 'Overrides por usuário/tenant'],
    cta: 'Assinar Growth',
    ctaHref: '/checkout?plan=growth',
    highlighted: true,
  },
  {
    name: 'Pro',
    price: 'R$ 2.499',
    priceNote: '/mês',
    description: 'Para empresas com necessidades de escala e compliance.',
    events: 'Ilimitados',
    flags: 'Ilimitadas',
    projects: 'Ilimitados',
    abTesting: 'Avançado + ilimitado',
    history: '365 dias',
    sla: '99,9% uptime',
    support: 'Email + Slack canal dedicado',
    extras: ['Retention analysis', 'Webhooks customizados', 'API priority SLA 4h', 'Audit log compliance LGPD'],
    cta: 'Falar com vendas',
    ctaHref: 'mailto:comercial@institutoitinerante.com.br',
  },
]

const featureRow = (label: string, value: string | false) => (
  <div key={label} className="flex items-center gap-2 text-sm">
    {value === false ? (
      <X className="w-4 h-4 text-slate-300 dark:text-slate-600 flex-shrink-0" />
    ) : (
      <Check className="w-4 h-4 text-emerald-500 flex-shrink-0" />
    )}
    <span className={value === false ? 'text-slate-400 dark:text-slate-500' : 'text-slate-700 dark:text-slate-300'}>
      {value === false ? label : `${label}: ${value}`}
    </span>
  </div>
)

export default function Pricing() {
  return (
    <div className="min-h-screen bg-slate-950 text-white">
      {/* Header */}
      <header className="border-b border-slate-800 px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center gap-2">
          <div className="w-7 h-7 rounded-lg bg-violet-600 flex items-center justify-center">
            <Zap className="w-4 h-4 text-white" />
          </div>
          <span className="font-bold text-lg">Pulse</span>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-16">
        {/* Hero */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold tracking-tight mb-4">
            Pricing simples e transparente
          </h1>
          <p className="text-slate-400 text-lg max-w-xl mx-auto">
            Feature flags, event tracking e A/B testing num só produto.
            Preços em R$, sem surpresas.
          </p>
        </div>

        {/* Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
          {plans.map((plan) => (
            <div
              key={plan.name}
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
                  <span className="text-3xl font-bold">{plan.price}</span>
                  {plan.priceNote && (
                    <span className="text-slate-400 text-sm">{plan.priceNote}</span>
                  )}
                </div>
                <p className="text-slate-400 text-sm mt-2">{plan.description}</p>
              </div>

              <div className="space-y-2 flex-1">
                {featureRow('Eventos/mês', plan.events)}
                {featureRow('Feature flags', plan.flags)}
                {featureRow('Projetos', plan.projects)}
                {featureRow('A/B testing', plan.abTesting)}
                {featureRow('Histórico', plan.history)}
                {featureRow('SLA', plan.sla)}
                {featureRow('Suporte', plan.support)}
                {plan.extras.map((extra) => (
                  <div key={extra} className="flex items-center gap-2 text-sm">
                    <Check className="w-4 h-4 text-violet-400 flex-shrink-0" />
                    <span className="text-slate-300">{extra}</span>
                  </div>
                ))}
              </div>

              <a
                href={plan.ctaHref}
                className={`block text-center py-2.5 px-4 rounded-xl font-semibold text-sm transition-colors ${
                  plan.highlighted
                    ? 'bg-violet-600 hover:bg-violet-500 text-white'
                    : 'bg-slate-800 hover:bg-slate-700 text-slate-200'
                }`}
              >
                {plan.cta}
              </a>
            </div>
          ))}
        </div>

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
                a: 'Você paga overage: Starter R$0,20/1k eventos, Growth R$0,10/1k. O serviço não para.',
              },
              {
                q: 'Vocês oferecem desconto anual?',
                a: 'Sim — 2 meses grátis no plano anual (Growth e Pro). Entre em contato.',
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

        {/* CTA final */}
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
        <a href="mailto:suporte@institutoitinerante.com.br" className="hover:text-slate-300 transition-colors">Suporte</a>
      </footer>
    </div>
  )
}
