// pulse-service JS SDK — Feature Flags
// Usage: import { PulseSDK } from './pulse'
// const pulse = new PulseSDK({ apiUrl: 'https://pulse-api.institutoitinerante.com.br' })
// const flags = await pulse.getFlags()
// pulse.isEnabled('my-feature')

class PulseSDK {
  constructor({ apiUrl }) {
    this.apiUrl = apiUrl
    this._flags = null
  }

  async getFlags() {
    try {
      const resp = await fetch(`${this.apiUrl}/flags`)
      if (!resp.ok) return {}
      const data = await resp.json()
      this._flags = data.flags ?? {}
      return this._flags
    } catch {
      return {}
    }
  }

  isEnabled(flagKey) {
    if (!this._flags) return false
    return !!this._flags[flagKey]
  }

  // Fase 1: no-op. Fase 2 will persist events.
  track(_eventName, _properties = {}) {
    return Promise.resolve()
  }
}

export { PulseSDK }
export default PulseSDK
