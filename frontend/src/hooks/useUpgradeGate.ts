import { useState } from 'react'

// ── Types ─────────────────────────────────────────────────────────────────────

interface GateState {
  isOpen: boolean
  currentPlan: string
  requiredPlan: string
  feature: string
  upgradeUrl: string
  message: string
}

interface TriggerGatePayload {
  plan: string
  upgrade_url: string
  message: string
  feature?: string
  required_plan?: string
}

// ── Hook ──────────────────────────────────────────────────────────────────────

export function useUpgradeGate() {
  const [gateState, setGateState] = useState<GateState>({
    isOpen: false,
    currentPlan: 'free',
    requiredPlan: 'starter',
    feature: '',
    upgradeUrl: '',
    message: '',
  })

  const triggerGate = (payload: TriggerGatePayload) => {
    setGateState({
      isOpen: true,
      currentPlan: payload.plan,
      requiredPlan: payload.required_plan ?? 'starter',
      feature: payload.feature ?? 'usar este recurso',
      upgradeUrl: payload.upgrade_url,
      message: payload.message,
    })
  }

  const closeGate = () => setGateState((s) => ({ ...s, isOpen: false }))

  return { gateState, triggerGate, closeGate }
}
