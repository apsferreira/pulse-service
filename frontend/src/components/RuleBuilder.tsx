// BKL-1067: Visual Rule Builder para targeting de feature flags.
// Interface drag-and-drop para criar regras de targeting sem JSON manual.
// Suporta: AND/OR lógico, propriedades de usuário/tenant, operadores de comparação.
//
// Uso: <RuleBuilder rules={rules} onChange={setRules} />

import { useState } from 'react'
import { Plus, Trash2, ChevronDown } from 'lucide-react'

// -------- tipos --------

export type RuleOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'in_list'
  | 'greater_than'
  | 'less_than'

export type RuleLogic = 'AND' | 'OR'

export interface TargetingRule {
  id: string
  property: string   // ex: "plan", "country", "tenant_id", "user_id"
  operator: RuleOperator
  value: string      // valor comparado (string — convertido no servidor)
}

export interface RuleGroup {
  id: string
  logic: RuleLogic
  rules: TargetingRule[]
}

// -------- constantes --------

const PROPERTY_OPTIONS: { label: string; value: string }[] = [
  { label: 'Plano (plan)', value: 'plan' },
  { label: 'País (country)', value: 'country' },
  { label: 'Tenant ID', value: 'tenant_id' },
  { label: 'User ID', value: 'user_id' },
  { label: 'Email', value: 'email' },
  { label: 'Versão do app (app_version)', value: 'app_version' },
  { label: 'Plataforma (platform)', value: 'platform' },
  { label: 'Custom attribute', value: '__custom__' },
]

const OPERATOR_OPTIONS: { label: string; value: RuleOperator }[] = [
  { label: '= igual a', value: 'equals' },
  { label: '≠ diferente de', value: 'not_equals' },
  { label: 'contém', value: 'contains' },
  { label: 'não contém', value: 'not_contains' },
  { label: 'começa com', value: 'starts_with' },
  { label: 'está na lista (vírgulas)', value: 'in_list' },
  { label: '> maior que', value: 'greater_than' },
  { label: '< menor que', value: 'less_than' },
]

const PLAN_SUGGESTIONS = ['free', 'starter', 'growth', 'pro']
const PLATFORM_SUGGESTIONS = ['web', 'ios', 'android', 'rn']

// -------- utils --------

const genID = () => Math.random().toString(36).slice(2, 9)

function emptyRule(): TargetingRule {
  return { id: genID(), property: 'plan', operator: 'equals', value: '' }
}

function emptyGroup(): RuleGroup {
  return { id: genID(), logic: 'AND', rules: [emptyRule()] }
}

// -------- sub-componentes --------

interface RuleRowProps {
  rule: TargetingRule
  onChange: (updated: TargetingRule) => void
  onRemove: () => void
  isOnly: boolean
}

function RuleRow({ rule, onChange, onRemove, isOnly }: RuleRowProps) {
  const [customProp, setCustomProp] = useState(
    PROPERTY_OPTIONS.find(o => o.value === rule.property) ? '' : rule.property
  )

  const handlePropChange = (val: string) => {
    if (val === '__custom__') {
      onChange({ ...rule, property: customProp || '' })
    } else {
      onChange({ ...rule, property: val })
    }
  }

  const suggestions =
    rule.property === 'plan' ? PLAN_SUGGESTIONS :
    rule.property === 'platform' ? PLATFORM_SUGGESTIONS :
    []

  const selectedPropOption = PROPERTY_OPTIONS.find(o => o.value === rule.property)?.value ?? '__custom__'

  return (
    <div className="flex flex-wrap items-center gap-2 p-2 bg-gray-50 rounded-lg border border-gray-200">
      {/* Propriedade */}
      <div className="relative">
        <select
          value={selectedPropOption}
          onChange={e => handlePropChange(e.target.value)}
          className="appearance-none border border-gray-300 rounded px-3 py-1.5 pr-7 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white"
        >
          {PROPERTY_OPTIONS.map(opt => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <ChevronDown size={14} className="absolute right-2 top-2.5 pointer-events-none text-gray-400" />
      </div>

      {/* Input de propriedade custom */}
      {selectedPropOption === '__custom__' && (
        <input
          type="text"
          placeholder="nome do atributo"
          value={customProp}
          onChange={e => {
            setCustomProp(e.target.value)
            onChange({ ...rule, property: e.target.value })
          }}
          className="border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-36"
        />
      )}

      {/* Operador */}
      <div className="relative">
        <select
          value={rule.operator}
          onChange={e => onChange({ ...rule, operator: e.target.value as RuleOperator })}
          className="appearance-none border border-gray-300 rounded px-3 py-1.5 pr-7 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white"
        >
          {OPERATOR_OPTIONS.map(opt => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <ChevronDown size={14} className="absolute right-2 top-2.5 pointer-events-none text-gray-400" />
      </div>

      {/* Valor com sugestões */}
      <div className="relative">
        <input
          type="text"
          placeholder="valor"
          value={rule.value}
          onChange={e => onChange({ ...rule, value: e.target.value })}
          list={`suggestions-${rule.id}`}
          className="border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-36"
        />
        {suggestions.length > 0 && (
          <datalist id={`suggestions-${rule.id}`}>
            {suggestions.map(s => <option key={s} value={s} />)}
          </datalist>
        )}
      </div>

      {/* Remover regra */}
      {!isOnly && (
        <button
          type="button"
          onClick={onRemove}
          className="p-1 text-gray-400 hover:text-red-500 transition-colors"
          title="Remover regra"
        >
          <Trash2 size={15} />
        </button>
      )}
    </div>
  )
}

// -------- componente principal --------

interface RuleBuilderProps {
  groups: RuleGroup[]
  onChange: (groups: RuleGroup[]) => void
  disabled?: boolean
}

export default function RuleBuilder({ groups, onChange, disabled = false }: RuleBuilderProps) {
  const updateGroup = (idx: number, updated: RuleGroup) => {
    const next = [...groups]
    next[idx] = updated
    onChange(next)
  }

  const removeGroup = (idx: number) => {
    onChange(groups.filter((_, i) => i !== idx))
  }

  const addGroup = () => {
    onChange([...groups, emptyGroup()])
  }

  return (
    <div className={`space-y-3 ${disabled ? 'opacity-60 pointer-events-none' : ''}`}>
      {groups.map((group, gIdx) => (
        <div key={group.id} className="border border-indigo-200 rounded-xl p-3 bg-indigo-50/40">
          {/* Header do grupo */}
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <span className="text-xs font-semibold text-indigo-700 uppercase tracking-wide">
                {gIdx === 0 ? 'Mostrar quando' : 'OU quando'}
              </span>

              {/* Toggle AND/OR */}
              <div className="flex border border-gray-300 rounded overflow-hidden text-xs">
                {(['AND', 'OR'] as RuleLogic[]).map(logic => (
                  <button
                    key={logic}
                    type="button"
                    onClick={() => updateGroup(gIdx, { ...group, logic })}
                    className={`px-2 py-0.5 transition-colors ${
                      group.logic === logic
                        ? 'bg-indigo-600 text-white'
                        : 'bg-white text-gray-600 hover:bg-gray-50'
                    }`}
                  >
                    {logic}
                  </button>
                ))}
              </div>

              <span className="text-xs text-gray-500">todas as regras abaixo</span>
            </div>

            {/* Remover grupo */}
            {groups.length > 1 && (
              <button
                type="button"
                onClick={() => removeGroup(gIdx)}
                className="text-gray-400 hover:text-red-500 text-xs flex items-center gap-1"
              >
                <Trash2 size={13} /> remover grupo
              </button>
            )}
          </div>

          {/* Regras */}
          <div className="space-y-2">
            {group.rules.map((rule, rIdx) => (
              <div key={rule.id}>
                {rIdx > 0 && (
                  <div className="text-xs text-indigo-600 font-semibold px-2 py-0.5">
                    {group.logic}
                  </div>
                )}
                <RuleRow
                  rule={rule}
                  isOnly={group.rules.length === 1}
                  onChange={updated => {
                    const next = [...group.rules]
                    next[rIdx] = updated
                    updateGroup(gIdx, { ...group, rules: next })
                  }}
                  onRemove={() => {
                    const next = group.rules.filter((_, i) => i !== rIdx)
                    updateGroup(gIdx, { ...group, rules: next })
                  }}
                />
              </div>
            ))}

            {/* Adicionar regra */}
            <button
              type="button"
              onClick={() => updateGroup(gIdx, { ...group, rules: [...group.rules, emptyRule()] })}
              className="mt-1 flex items-center gap-1 text-xs text-indigo-600 hover:text-indigo-800 font-medium"
            >
              <Plus size={13} /> adicionar condição
            </button>
          </div>
        </div>
      ))}

      {/* Adicionar grupo OR */}
      <button
        type="button"
        onClick={addGroup}
        className="flex items-center gap-1.5 text-sm text-indigo-600 hover:text-indigo-800 font-medium border border-dashed border-indigo-300 rounded-lg px-3 py-2 w-full justify-center hover:border-indigo-500 transition-colors"
      >
        <Plus size={15} /> adicionar grupo OU
      </button>

      {/* Preview JSON (collapsible) */}
      {groups.some(g => g.rules.some(r => r.value)) && (
        <details className="text-xs">
          <summary className="cursor-pointer text-gray-500 hover:text-gray-700">
            ver JSON da regra
          </summary>
          <pre className="mt-1 bg-gray-100 rounded p-2 overflow-x-auto text-gray-700 text-[11px]">
            {JSON.stringify(groups, null, 2)}
          </pre>
        </details>
      )}
    </div>
  )
}

// -------- helpers de conversão --------

/** Converte RuleGroups para o formato aceito pelo backend (targeting.rules) */
export function ruleGroupsToBackend(groups: RuleGroup[]): object {
  return {
    version: 2,
    groups: groups.map(g => ({
      logic: g.logic,
      conditions: g.rules
        .filter(r => r.property && r.value)
        .map(r => ({
          attribute: r.property,
          operator: r.operator,
          value: r.operator === 'in_list'
            ? r.value.split(',').map(s => s.trim())
            : r.value,
        })),
    })).filter(g => g.conditions.length > 0),
  }
}

/** Converte targeting do backend para RuleGroups */
export function backendToRuleGroups(targeting: Record<string, unknown> | null | undefined): RuleGroup[] {
  if (!targeting || targeting.version !== 2) {
    return [emptyGroup()]
  }
  const groups = targeting.groups as Array<{ logic: RuleLogic; conditions: Array<{ attribute: string; operator: RuleOperator; value: string | string[] }> }> | undefined
  if (!groups?.length) return [emptyGroup()]

  return groups.map(g => ({
    id: genID(),
    logic: g.logic,
    rules: (g.conditions ?? []).map(c => ({
      id: genID(),
      property: c.attribute,
      operator: c.operator,
      value: Array.isArray(c.value) ? c.value.join(', ') : String(c.value),
    })),
  }))
}
