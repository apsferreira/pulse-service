/**
 * Landing Page pública do Pulse
 * Posicionamento: concorrente direto do LaunchDarkly / VWO / Avo
 * Rota: /landing (não requer autenticação)
 */

const FEATURES = [
  {
    icon: '🚩',
    title: 'Feature Flags',
    desc: 'Rollout gradual, targeting por usuário/tenant, múltiplos ambientes. Tudo sem redeploy.',
  },
  {
    icon: '🧪',
    title: 'A/B Testing',
    desc: 'Crie variantes, atribua tráfego percentual, meça significância estatística. Bayesiano + frequentista.',
  },
  {
    icon: '📊',
    title: 'Event Tracking & Funis',
    desc: 'Registre eventos customizados, monte funis de conversão e visualize onde os usuários abandonam o fluxo.',
  },
  {
    icon: '🎥',
    title: 'Session Replay',
    desc: 'Assista gravações de sessões reais de usuários. Veja cliques, scroll e interações — sem instalar nada extra.',
  },
  {
    icon: '📈',
    title: 'Retenção & Cohorts',
    desc: 'Analise retenção por coorte de usuários. Descubra o que separa os que ficam dos que abandonam o produto.',
  },
  {
    icon: '🖥️',
    title: 'Dashboard Unificado',
    desc: 'Flags, experimentos, session replay e cohorts em um único painel. Sem alternar entre ferramentas.',
  },
  {
    icon: '🔒',
    title: 'LGPD Nativo',
    desc: 'Dados no Brasil. Self-hosted disponível para compliance total. Zero transferência para fora do país.',
  },
  {
    icon: '⚡',
    title: 'SDKs Prontos',
    desc: 'CDN script tag, @pulse/react com hooks, @pulse/react-native para Expo. Integre em minutos.',
  },
  {
    icon: '💰',
    title: 'Preço em R$',
    desc: 'Sem surpresa de câmbio. 40-60% mais barato que LaunchDarkly e VWO para startups brasileiras.',
  },
]

const PLANS = [
  {
    name: 'Free',
    price: 'R$ 0',
    period: '/mês',
    highlight: false,
    cta: 'Começar grátis',
    items: [
      'Até 5k MAU',
      '10 feature flags',
      '1 ambiente',
      '7 dias retenção',
      'Rollout percentual',
      'Suporte comunidade',
    ],
  },
  {
    name: 'Starter',
    price: 'R$ 299',
    period: '/mês',
    highlight: false,
    cta: 'Iniciar trial',
    items: [
      'Até 25k MAU',
      '50 feature flags',
      '3 ambientes',
      '30 dias retenção',
      'A/B testing',
      '2 seats',
      'Email support 72h',
    ],
  },
  {
    name: 'Growth',
    price: 'R$ 799',
    period: '/mês',
    highlight: true,
    badge: 'Mais popular',
    cta: 'Iniciar trial',
    items: [
      'Até 100k MAU',
      '500 feature flags',
      '5 ambientes',
      '90 dias retenção',
      'Cohorts e targeting',
      'Webhooks',
      '5 seats',
      'SSO (SAML)',
    ],
  },
  {
    name: 'Pro',
    price: 'R$ 1.999',
    period: '/mês',
    highlight: false,
    cta: 'Falar com vendas',
    items: [
      'Até 500k MAU',
      '5.000 feature flags',
      'Ambientes ilimitados',
      '180 dias retenção',
      'Análise Bayesiana',
      '20 seats',
      'RBAC avançado',
      'SLA 4h',
    ],
  },
  {
    name: 'Self-Hosted',
    price: 'R$ 24k',
    period: '/ano',
    highlight: false,
    cta: 'Falar com vendas',
    items: [
      'MAU ilimitado',
      'Flags ilimitadas',
      'Deploy on-prem/VPC/K8s',
      'Dados 100% no cliente',
      'Suporte 24/5',
      'LGPD compliant',
    ],
  },
]

const COMPARE = [
  { feature: 'Feature flags com rollout', pulse: true, ld: true, vwo: false },
  { feature: 'A/B testing', pulse: true, ld: true, vwo: true },
  { feature: 'Event tracking nativo', pulse: true, ld: false, vwo: false },
  { feature: 'Analytics de produto', pulse: true, ld: false, vwo: true },
  { feature: 'Session Replay', pulse: true, ld: false, vwo: true },
  { feature: 'Heatmaps', pulse: false, ld: false, vwo: true },
  { feature: 'Funil de conversão', pulse: true, ld: false, vwo: true },
  { feature: 'Retenção & Cohorts', pulse: true, ld: false, vwo: false },
  { feature: 'Dashboard unificado', pulse: true, ld: false, vwo: false },
  { feature: 'Self-hosted', pulse: true, ld: false, vwo: false },
  { feature: 'Pricing em BRL', pulse: true, ld: false, vwo: false },
  { feature: 'SDK React / React Native', pulse: true, ld: true, vwo: false },
  { feature: 'LGPD nativo (dados no BR)', pulse: true, ld: false, vwo: false },
  { feature: 'Plano gratuito', pulse: true, ld: true, vwo: true },
  { feature: 'Suporte PT-BR', pulse: true, ld: false, vwo: false },
]

function CheckIcon() {
  return (
    <svg className="w-5 h-5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  )
}

function XIcon() {
  return (
    <svg className="w-5 h-5 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
    </svg>
  )
}

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* ── Nav ─────────────────────────────────────────── */}
      <nav className="border-b border-white/10 px-6 py-4 flex items-center justify-between max-w-7xl mx-auto">
        <div className="flex items-center gap-2">
          <span className="text-2xl font-black tracking-tight">Pulse</span>
          <span className="text-xs bg-emerald-500/20 text-emerald-400 px-2 py-0.5 rounded-full font-medium">beta</span>
        </div>
        <div className="hidden md:flex items-center gap-8 text-sm text-gray-400">
          <a href="#features" className="hover:text-white transition-colors">Features</a>
          <a href="#compare" className="hover:text-white transition-colors">Comparativo</a>
          <a href="#pricing" className="hover:text-white transition-colors">Preços</a>
          <a href="#sdk" className="hover:text-white transition-colors">SDKs</a>
        </div>
        <div className="flex items-center gap-3">
          <a href="/login" className="text-sm text-gray-400 hover:text-white transition-colors">Entrar</a>
          <a
            href="#pricing"
            className="text-sm bg-emerald-500 hover:bg-emerald-400 text-black font-semibold px-4 py-2 rounded-lg transition-colors"
          >
            Começar grátis
          </a>
        </div>
      </nav>

      {/* ── Hero ────────────────────────────────────────── */}
      <section className="max-w-5xl mx-auto px-6 pt-24 pb-20 text-center">
        <div className="inline-flex items-center gap-2 bg-emerald-500/10 border border-emerald-500/20 rounded-full px-4 py-1.5 text-sm text-emerald-400 mb-8">
          <span>🇧🇷</span>
          <span>Alternativa brasileira ao LaunchDarkly e VWO</span>
        </div>

        <h1 className="text-5xl md:text-7xl font-black leading-tight mb-6">
          Product analytics<br />
          <span className="text-emerald-400">completo</span> sem<br />
          pagar em dólar
        </h1>

        <p className="text-xl text-gray-400 max-w-2xl mx-auto mb-10">
          Pulse reúne feature flags, A/B testing, session replay, heatmaps, funis de conversão
          e cohorts em uma plataforma só. Self-hosted ou cloud, pricing em BRL, dados no Brasil.
          40-60% mais barato que LaunchDarkly e VWO.
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="#pricing"
            className="bg-emerald-500 hover:bg-emerald-400 text-black font-bold px-8 py-4 rounded-xl text-lg transition-colors"
          >
            Começar grátis
          </a>
          <a
            href="#sdk"
            className="border border-white/20 hover:border-white/40 text-white font-semibold px-8 py-4 rounded-xl text-lg transition-colors"
          >
            Ver SDKs
          </a>
        </div>

        <p className="text-sm text-gray-500 mt-6">
          Sem cartão de crédito · Plano Free com 5k MAU · Self-hosted disponível
        </p>
      </section>

      {/* ── Features ────────────────────────────────────── */}
      <section id="features" className="max-w-6xl mx-auto px-6 py-20">
        <h2 className="text-3xl font-bold text-center mb-4">Tudo que você precisa, em um lugar</h2>
        <p className="text-gray-400 text-center mb-16 max-w-xl mx-auto">
          Não é só feature flags. É feature flags + A/B testing + session replay + heatmaps + funis + cohorts — tudo integrado, sem juntar cinco ferramentas.
        </p>
        <div className="grid sm:grid-cols-2 lg:grid-cols-5 gap-6">
          {FEATURES.map((f) => (
            <div
              key={f.title}
              className="bg-gray-900 border border-white/10 rounded-2xl p-6 hover:border-emerald-500/30 transition-colors"
            >
              <div className="text-3xl mb-4">{f.icon}</div>
              <h3 className="font-bold text-lg mb-2">{f.title}</h3>
              <p className="text-gray-400 text-sm leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* ── SDK ─────────────────────────────────────────── */}
      <section id="sdk" className="max-w-5xl mx-auto px-6 py-20">
        <h2 className="text-3xl font-bold text-center mb-4">Integre em minutos</h2>
        <p className="text-gray-400 text-center mb-16">
          Três SDKs para cobrir todas as plataformas que você usa.
        </p>

        <div className="grid md:grid-cols-3 gap-6 mb-12">
          {/* CDN */}
          <div className="bg-gray-900 border border-white/10 rounded-2xl p-6">
            <div className="text-xs text-emerald-400 font-semibold mb-2 uppercase tracking-wide">HTML / CDN</div>
            <h3 className="font-bold mb-3">pulse.js</h3>
            <p className="text-gray-400 text-sm mb-4">Adicione uma tag script no head. Funciona em qualquer site.</p>
            <pre className="bg-black/50 rounded-lg p-3 text-xs text-gray-300 overflow-x-auto">
{`<script src="https://pulse-api.
  institutoitinerante.com.br/sdk/
  pulse.js"></script>
<script>
  pulse.init({
    projectKey: 'YOUR_KEY',
    apiUrl: 'https://pulse-api...',
  })
  pulse.track('page_viewed')
</script>`}
            </pre>
          </div>

          {/* React */}
          <div className="bg-gray-900 border border-white/10 rounded-2xl p-6">
            <div className="text-xs text-emerald-400 font-semibold mb-2 uppercase tracking-wide">React</div>
            <h3 className="font-bold mb-3">@pulse/react</h3>
            <p className="text-gray-400 text-sm mb-4">Hooks nativos para React. Provider de contexto incluso.</p>
            <pre className="bg-black/50 rounded-lg p-3 text-xs text-gray-300 overflow-x-auto">
{`import { PulseProvider, useFlag,
  useExperiment } from '@pulse/react'

function App() {
  const newUI = useFlag('new-ui')
  const { variant } =
    useExperiment('pricing-test')
  return newUI ? <NewUI /> : <OldUI />
}`}
            </pre>
          </div>

          {/* React Native */}
          <div className="bg-gray-900 border border-white/10 rounded-2xl p-6">
            <div className="text-xs text-emerald-400 font-semibold mb-2 uppercase tracking-wide">React Native / Expo</div>
            <h3 className="font-bold mb-3">@pulse/react-native</h3>
            <p className="text-gray-400 text-sm mb-4">API idêntica ao SDK React, com AsyncStorage e sem DOM.</p>
            <pre className="bg-black/50 rounded-lg p-3 text-xs text-gray-300 overflow-x-auto">
{`import { PulseProvider, useFlag,
  useTrack } from '@pulse/react-native'

function HomeScreen() {
  const flag = useFlag('new-onboarding')
  const track = useTrack()
  return <Button
    onPress={() => track('cta_pressed')}
  />`}
            </pre>
          </div>
        </div>
      </section>

      {/* ── Compare ─────────────────────────────────────── */}
      <section id="compare" className="max-w-4xl mx-auto px-6 py-20">
        <h2 className="text-3xl font-bold text-center mb-4">Pulse vs LaunchDarkly vs VWO</h2>
        <p className="text-gray-400 text-center mb-16">
          Feature flags, A/B testing e analytics juntos — sem pagar em dólar.
        </p>

        <div className="bg-gray-900 border border-white/10 rounded-2xl overflow-hidden">
          <div className="grid grid-cols-4 px-6 py-4 border-b border-white/10 text-sm font-semibold">
            <div className="col-span-1 text-gray-400">Feature</div>
            <div className="text-center text-emerald-400">Pulse</div>
            <div className="text-center text-gray-400">LaunchDarkly</div>
            <div className="text-center text-gray-400">VWO</div>
          </div>
          {COMPARE.map((row, i) => (
            <div
              key={row.feature}
              className={`grid grid-cols-4 px-6 py-3 text-sm ${i % 2 === 0 ? 'bg-white/[0.02]' : ''}`}
            >
              <div className="col-span-1 text-gray-300">{row.feature}</div>
              <div className="flex justify-center">{row.pulse ? <CheckIcon /> : <XIcon />}</div>
              <div className="flex justify-center">{row.ld ? <CheckIcon /> : <XIcon />}</div>
              <div className="flex justify-center">{row.vwo ? <CheckIcon /> : <XIcon />}</div>
            </div>
          ))}
        </div>

        <div className="mt-8 grid md:grid-cols-3 gap-4 text-center text-sm">
          <div className="bg-gray-900 border border-white/10 rounded-xl p-4">
            <div className="text-gray-400 mb-1">LaunchDarkly Growth (10k MAU)</div>
            <div className="text-2xl font-bold text-red-400">~R$520/mês</div>
            <div className="text-gray-500 text-xs">USD $100 × câmbio 5,20</div>
          </div>
          <div className="bg-gray-900 border border-white/10 rounded-xl p-4">
            <div className="text-gray-400 mb-1">VWO Starter (10k MAU)</div>
            <div className="text-2xl font-bold text-red-400">~R$1.030/mês</div>
            <div className="text-gray-500 text-xs">USD $198 × câmbio 5,20</div>
          </div>
          <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4">
            <div className="text-gray-400 mb-1">Pulse Starter (25k MAU)</div>
            <div className="text-2xl font-bold text-emerald-400">R$299/mês</div>
            <div className="text-gray-500 text-xs">Billing em BRL · sem câmbio</div>
          </div>
        </div>
      </section>

      {/* ── Pricing ─────────────────────────────────────── */}
      <section id="pricing" className="max-w-7xl mx-auto px-6 py-20">
        <h2 className="text-3xl font-bold text-center mb-4">Planos e preços</h2>
        <p className="text-gray-400 text-center mb-16">
          Transparente, sem cobrança por MAU acima do limite, sem surpresas de câmbio.
        </p>

        <div className="grid md:grid-cols-5 gap-4">
          {PLANS.map((plan) => (
            <div
              key={plan.name}
              className={`relative rounded-2xl p-6 flex flex-col ${
                plan.highlight
                  ? 'bg-emerald-500/10 border-2 border-emerald-500'
                  : 'bg-gray-900 border border-white/10'
              }`}
            >
              {plan.badge && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-emerald-500 text-black text-xs font-bold px-3 py-1 rounded-full">
                  {plan.badge}
                </div>
              )}
              <div className="mb-4">
                <div className="text-sm text-gray-400 mb-1">{plan.name}</div>
                <div className="flex items-end gap-1">
                  <span className="text-2xl font-black">{plan.price}</span>
                  <span className="text-gray-400 text-sm pb-0.5">{plan.period}</span>
                </div>
              </div>

              <ul className="space-y-2 flex-1 mb-6">
                {plan.items.map((item) => (
                  <li key={item} className="flex items-center gap-2 text-sm text-gray-300">
                    <CheckIcon />
                    {item}
                  </li>
                ))}
              </ul>

              <a
                href="#"
                className={`text-center text-sm font-semibold py-2.5 rounded-lg transition-colors ${
                  plan.highlight
                    ? 'bg-emerald-500 hover:bg-emerald-400 text-black'
                    : 'border border-white/20 hover:border-white/40 text-white'
                }`}
              >
                {plan.cta}
              </a>
            </div>
          ))}
        </div>
      </section>

      {/* ── CTA final ───────────────────────────────────── */}
      <section className="max-w-3xl mx-auto px-6 py-20 text-center">
        <h2 className="text-4xl font-black mb-4">
          Pronto para parar de pagar em dólar?
        </h2>
        <p className="text-gray-400 mb-8">
          Feature flags + A/B testing + session replay + heatmaps + cohorts. Self-hosted ou cloud. Dados no Brasil.
        </p>
        <a
          href="#pricing"
          className="inline-block bg-emerald-500 hover:bg-emerald-400 text-black font-bold px-10 py-4 rounded-xl text-lg transition-colors"
        >
          Começar grátis agora
        </a>
        <p className="text-gray-500 text-sm mt-4">Sem cartão · Self-hosted disponível · Suporte em PT-BR</p>
      </section>

      {/* ── Footer ──────────────────────────────────────── */}
      <footer className="border-t border-white/10 px-6 py-8 text-center text-sm text-gray-500">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-4">
          <span className="font-bold text-white">Pulse</span>
          <span>© 2026 Instituto Itinerante de Tecnologia · Salvador, BA · LGPD Compliant</span>
          <div className="flex gap-6">
            <a href="#" className="hover:text-white transition-colors">Privacidade</a>
            <a href="#" className="hover:text-white transition-colors">Termos</a>
            <a href="#" className="hover:text-white transition-colors">Documentação</a>
          </div>
        </div>
      </footer>
    </div>
  )
}
