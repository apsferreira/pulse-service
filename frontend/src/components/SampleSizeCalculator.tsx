// BKL-1045: Sample size calculator para experimentos
// Cálculo Z-test two-proportion com power estatística

import { useState } from 'react'
import { Calculator, HelpCircle } from 'lucide-react'

interface CalculationResult {
  samplePerVariant: number
  totalSample: number
  estimatedDays: number
  exposureRate: number
}

const Z_SCORES = {
  '80': 0.842, // 80% power
  '90': 1.282, // 90% power
}

const SIGNIFICANCE_LEVELS = {
  '80': 1.282, // 80% (two-sided, z = 1.282)
  '90': 1.645, // 90% (two-sided, z = 1.645)
  '95': 1.96,  // 95% (two-sided, z = 1.96)
}

function calculateSampleSize(
  baselineConversion: number,
  expectedUplift: number,
  significanceLevel: string,
  power: string,
  exposureRate: number = 100
): CalculationResult {
  // Converter percentagens para proporções
  const p1 = baselineConversion / 100
  const p2 = p1 * (1 + expectedUplift / 100)

  // Clamp p2 para [0, 1]
  const adjustedP2 = Math.min(Math.max(p2, 0), 1)
  const pooledP = (p1 + adjustedP2) / 2

  // Z-scores
  const z_alpha = SIGNIFICANCE_LEVELS[significanceLevel as keyof typeof SIGNIFICANCE_LEVELS] || 1.96
  const z_beta = Z_SCORES[power as keyof typeof Z_SCORES] || 0.842

  // Fórmula: n = (z_alpha + z_beta)^2 * p(1-p) * 2 / (p1 - p2)^2
  const numerator = Math.pow(z_alpha + z_beta, 2) * pooledP * (1 - pooledP) * 2
  const denominator = Math.pow(p1 - adjustedP2, 2)
  const samplePerVariant = Math.ceil(numerator / denominator)
  const totalSample = samplePerVariant * 2

  // Estimar dias (baseado em exposureRate)
  const dailyExposures = (totalSample * exposureRate) / 100
  const estimatedDays = dailyExposures > 0 ? Math.ceil(totalSample / dailyExposures) : 0

  return {
    samplePerVariant,
    totalSample,
    estimatedDays,
    exposureRate,
  }
}

export default function SampleSizeCalculator() {
  const [baseline, setBaseline] = useState('5')
  const [uplift, setUplift] = useState('10')
  const [significance, setSignificance] = useState('95')
  const [power, setPower] = useState('80')
  const [exposureRate, setExposureRate] = useState('100')

  const baselineNum = parseFloat(baseline) || 0
  const upliftNum = parseFloat(uplift) || 0
  const exposureNum = parseFloat(exposureRate) || 100

  const isValid = baselineNum > 0 && baselineNum <= 100 && upliftNum > 0 && exposureNum > 0

  const result = isValid ? calculateSampleSize(baselineNum, upliftNum, significance, power, exposureNum) : null

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
      <div className="flex items-center gap-2 mb-4">
        <Calculator className="w-4 h-4 text-violet-500" />
        <h2 className="text-base font-semibold text-slate-800 dark:text-slate-100">
          Calculadora de Tamanho de Amostra
        </h2>
      </div>

      <div className="space-y-4">
        {/* Inputs Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {/* Baseline Conversion Rate */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Taxa de conversão baseline (%)
              <span className="text-xs text-slate-500 dark:text-slate-400 font-normal ml-1">
                (ex: 5%)
              </span>
            </label>
            <input
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={baseline}
              onChange={(e) => setBaseline(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
              placeholder="5"
            />
          </div>

          {/* Expected Uplift */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Aumento esperado (%)
              <span className="text-xs text-slate-500 dark:text-slate-400 font-normal ml-1">
                (ex: 10%)
              </span>
            </label>
            <input
              type="number"
              min="0"
              step="0.1"
              value={uplift}
              onChange={(e) => setUplift(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
              placeholder="10"
            />
          </div>

          {/* Significance Level */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Nível de significância
            </label>
            <select
              value={significance}
              onChange={(e) => setSignificance(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            >
              <option value="80">80% (menos rigoroso)</option>
              <option value="90">90%</option>
              <option value="95">95% (padrão)</option>
            </select>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {significance === '95' && 'Risco de falso positivo: 5%'}
              {significance === '90' && 'Risco de falso positivo: 10%'}
              {significance === '80' && 'Risco de falso positivo: 20%'}
            </p>
          </div>

          {/* Power */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Poder estatístico (sensibilidade)
            </label>
            <select
              value={power}
              onChange={(e) => setPower(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            >
              <option value="80">80% (menos sensível)</option>
              <option value="90">90% (padrão)</option>
            </select>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
              {power === '90' && 'Risco de falso negativo: 10%'}
              {power === '80' && 'Risco de falso negativo: 20%'}
            </p>
          </div>
        </div>

        {/* Exposure Rate */}
        <div>
          <div className="flex items-center gap-1 mb-1.5">
            <label className="text-sm font-medium text-slate-700 dark:text-slate-300">
              Taxa de exposição diária
            </label>
            <div className="group relative">
              <HelpCircle className="w-3.5 h-3.5 text-slate-400 cursor-help" />
              <div className="absolute -top-8 left-full ml-2 bg-slate-800 text-white text-xs rounded px-2 py-1 whitespace-nowrap opacity-0 group-hover:opacity-100 pointer-events-none z-20">
                Quantos usuários entram no experimento por dia
              </div>
            </div>
          </div>
          <input
            type="number"
            min="1"
            step="100"
            value={exposureRate}
            onChange={(e) => setExposureRate(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            placeholder="1000"
          />
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
            Usuários por dia (será usado para estimar duração)
          </p>
        </div>
      </div>

      {/* Results */}
      {result && (
        <div className="mt-6 space-y-3">
          {/* Key Metric */}
          <div className="bg-violet-50 dark:bg-violet-950/40 rounded-lg border border-violet-200 dark:border-violet-800 p-4">
            <p className="text-xs text-violet-600 dark:text-violet-400 font-medium mb-1">
              Tamanho de amostra necessário (por variante)
            </p>
            <p className="text-3xl font-bold text-violet-700 dark:text-violet-300">
              {result.samplePerVariant.toLocaleString()}
            </p>
            <p className="text-xs text-violet-600 dark:text-violet-400 mt-1">
              Total: {result.totalSample.toLocaleString()} usuários (controle + variante)
            </p>
          </div>

          {/* Duration */}
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
              <p className="text-xs text-slate-500 dark:text-slate-400 mb-1">Duração estimada</p>
              <p className="text-xl font-bold text-slate-900 dark:text-slate-100">
                {result.estimatedDays === 1 ? '1 dia' : `${result.estimatedDays} dias`}
              </p>
              <p className="text-[10px] text-slate-400 mt-0.5">
                a {result.exposureRate.toLocaleString()} usuários/dia
              </p>
            </div>

            <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
              <p className="text-xs text-slate-500 dark:text-slate-400 mb-1">Semanas</p>
              <p className="text-xl font-bold text-slate-900 dark:text-slate-100">
                {(result.estimatedDays / 7).toFixed(1)}
              </p>
              <p className="text-[10px] text-slate-400 mt-0.5">
                aprox.
              </p>
            </div>
          </div>

          {/* Assumptions */}
          <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3 space-y-2 text-xs text-slate-600 dark:text-slate-400">
            <p className="font-medium text-slate-700 dark:text-slate-300">Parâmetros:</p>
            <ul className="space-y-1 text-[11px]">
              <li>
                • Baseline: {baselineNum.toFixed(2)}% → Esperado: {(baselineNum * (1 + upliftNum / 100)).toFixed(2)}%
              </li>
              <li>• Teste bilateral (Z = {(SIGNIFICANCE_LEVELS[significance as keyof typeof SIGNIFICANCE_LEVELS] || 1.96).toFixed(3)})</li>
              <li>• Poder: {power}% (detectar efeito {upliftNum}% do baseline com {power}% de confiança)</li>
            </ul>
          </div>

          {/* Recommendation */}
          {result.estimatedDays > 60 && (
            <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-3 text-xs text-amber-700 dark:text-amber-300">
              <p className="font-medium mb-1">Dica: Experimento longo detectado</p>
              <p>
                Consider aumentar o uplift esperado ou taxa de exposição para reduzir duração. Ou reduza o nível de significância para 90%.
              </p>
            </div>
          )}
        </div>
      )}

      {!isValid && (
        <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-200 dark:border-red-800 text-xs text-red-600 dark:text-red-400">
          Preencha todos os campos com valores válidos
        </div>
      )}
    </div>
  )
}
