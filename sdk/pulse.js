// pulse-service JS SDK — minimal Fase 1
// Usage: import { PulseSDK } from './pulse'
// const pulse = new PulseSDK({ apiUrl: 'https://pulse-api.institutoitinerante.com.br', product: 'libri' })
// const flags = await pulse.getFlags({ userId: '...' })
// await pulse.track('book_added', { bookId: '...' })  // no-op in Fase 1

class PulseSDK {
  constructor({ apiUrl, product, userId = null }) {
    this.apiUrl = apiUrl
    this.product = product
    this.userId = userId
    this._flags = null
  }

  async getFlags(overrides = {}) {
    const uid = overrides.userId ?? this.userId ?? ''
    try {
      const url = new URL(`${this.apiUrl}/api/v1/flags/evaluate`)
      url.searchParams.set('product', this.product)
      if (uid) url.searchParams.set('user_id', uid)
      const resp = await fetch(url.toString())
      if (!resp.ok) return {}
      const data = await resp.json()
      this._flags = data.flags ?? {}
      return this._flags
    } catch {
      return {}
    }
  }

  isEnabled(flagName) {
    if (!this._flags) return false
    return !!this._flags[flagName]
  }

  // Fase 1: no-op. Fase 2 will persist events.
  track(_eventName, _properties = {}) {
    return Promise.resolve()
  }
}

export { PulseSDK }
export default PulseSDK
