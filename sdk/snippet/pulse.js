/**
 * Pulse Snippet — tracking para qualquer site HTML.
 *
 * Uso:
 *   <script src="https://pulse.institutoitinerante.com.br/sdk/pulse.js"
 *           data-pulse-url="https://pulse.institutoitinerante.com.br"
 *           data-pulse-product="meu-site"></script>
 *
 * Após carregar, use window.pulse:
 *   pulse.track('page.viewed')
 *   pulse.identify('anon-id', 'user-id')
 *   pulse.isEnabled('minha-flag').then(enabled => { ... })
 */
;(function (w, d) {
  'use strict'

  var script = d.currentScript || (function () {
    var scripts = d.getElementsByTagName('script')
    return scripts[scripts.length - 1]
  })()

  var BASE_URL = (script && script.getAttribute('data-pulse-url')) || 'https://pulse.institutoitinerante.com.br'
  var PRODUCT  = (script && script.getAttribute('data-pulse-product')) || 'default'
  var ANON_KEY = 'pulse_anon_id'
  var CACHE_TTL = 5 * 60 * 1000 // 5 min

  var flagsCache = {}
  var cacheTime  = 0

  function getAnonId() {
    try {
      var id = localStorage.getItem(ANON_KEY)
      if (!id) {
        id = ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, function(c) {
          return (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16)
        })
        localStorage.setItem(ANON_KEY, id)
      }
      return id
    } catch (e) {
      return 'anon-' + Math.random().toString(36).slice(2)
    }
  }

  function post(path, payload) {
    try {
      navigator.sendBeacon
        ? navigator.sendBeacon(BASE_URL + path, new Blob([JSON.stringify(payload)], { type: 'application/json' }))
        : fetch(BASE_URL + path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload), keepalive: true })
    } catch (e) {}
  }

  var pulse = {
    /**
     * Envia um evento de tracking.
     * @param {string} eventName
     * @param {Object} [props]
     */
    track: function (eventName, props) {
      post('/api/v1/events/track', {
        event_name: eventName,
        product: PRODUCT,
        anonymous_id: getAnonId(),
        properties: Object.assign({
          url: w.location.pathname,
          referrer: d.referrer || undefined,
        }, props || {})
      })
    },

    /**
     * Associa ID anônimo a um usuário real após login.
     * @param {string} userId
     */
    identify: function (userId) {
      var anonId = getAnonId()
      post('/api/v1/events/identify', { anonymous_id: anonId, user_id: userId, product: PRODUCT })
      try { localStorage.removeItem(ANON_KEY) } catch (e) {}
    },

    /**
     * Verifica se uma feature flag está habilitada.
     * @param {string} flagKey
     * @param {boolean} [defaultValue=false]
     * @returns {Promise<boolean>}
     */
    isEnabled: function (flagKey, defaultValue) {
      defaultValue = defaultValue !== undefined ? defaultValue : false
      var now = Date.now()
      if (now - cacheTime < CACHE_TTL && Object.keys(flagsCache).length > 0) {
        return Promise.resolve(flagsCache[flagKey] !== undefined ? flagsCache[flagKey] : defaultValue)
      }
      return fetch(BASE_URL + '/flags', { method: 'GET' })
        .then(function (res) { return res.ok ? res.json() : { flags: {} } })
        .then(function (data) {
          flagsCache = data.flags || {}
          cacheTime = Date.now()
          return flagsCache[flagKey] !== undefined ? flagsCache[flagKey] : defaultValue
        })
        .catch(function () { return defaultValue })
    },

    /**
     * Rastreia visualização da página atual automaticamente.
     * Chamado automaticamente ao carregar o snippet.
     */
    pageview: function () {
      pulse.track('page.viewed', {
        title: d.title,
        url: w.location.pathname,
      })
    }
  }

  // Expõe globalmente
  w.pulse = pulse

  // Auto-pageview
  if (d.readyState === 'loading') {
    d.addEventListener('DOMContentLoaded', function () { pulse.pageview() })
  } else {
    pulse.pageview()
  }

  // SPA support: escuta pushState para rastrear navegações
  var _pushState = w.history.pushState
  w.history.pushState = function () {
    _pushState.apply(w.history, arguments)
    pulse.pageview()
  }
  w.addEventListener('popstate', function () { pulse.pageview() })

})(window, document)
