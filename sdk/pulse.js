/*!
 * Pulse SDK v2.0.0 — Feature Flags + A/B Testing + Event Tracking
 * Instituto Itinerante de Tecnologia
 *
 * Usage (CDN / script tag):
 *   <script src="https://pulse.institutoitinerante.com.br/sdk/pulse.js"></script>
 *   <script>
 *     pulse.init({ projectKey: 'YOUR_PROJECT_KEY', apiUrl: 'https://pulse.institutoitinerante.com.br' })
 *     pulse.identify('user_123', { name: 'João', plan: 'growth' })
 *     pulse.track('button_clicked', { button: 'signup' })
 *     pulse.isEnabled('new-checkout').then(enabled => { if (enabled) showNewCheckout() })
 *     pulse.getVariant('pricing-test').then(variant => console.log('variant:', variant))
 *   </script>
 */
;(function (root, factory) {
  if (typeof module === 'object' && module.exports) {
    module.exports = factory()       // CommonJS
  } else if (typeof define === 'function' && define.amd) {
    define(factory)                  // AMD
  } else {
    root.pulse = factory()           // Browser global
  }
}(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  'use strict'

  // ─── Internal state ────────────────────────────────────────────────────────

  var _config = {
    apiUrl: '',
    projectKey: '',
    debug: false,
    autoPageview: true,
    sessionTimeout: 30 * 60 * 1000,   // 30 min
    flushInterval: 2000,               // ms entre flushes do batch queue
    maxBatchSize: 20,
    timeout: 10000,
  }

  var _state = {
    initialized: false,
    userId: null,
    anonymousId: null,
    traits: {},
    flags: null,
    variants: null,
    sessionId: null,
    sessionStart: null,
    queue: [],                         // event queue para batch send
    flushTimer: null,
  }

  // ─── Utils ─────────────────────────────────────────────────────────────────

  function _uuid() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
      var r = (Math.random() * 16) | 0
      return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
    })
  }

  function _log() {
    if (_config.debug && typeof console !== 'undefined') {
      console.log.apply(console, ['[Pulse]'].concat(Array.prototype.slice.call(arguments)))
    }
  }

  function _getStorage(key) {
    try { return localStorage.getItem(key) } catch (e) { return null }
  }

  function _setStorage(key, value) {
    try { localStorage.setItem(key, value) } catch (e) {}
  }

  function _ensureAnonymousId() {
    if (_state.anonymousId) return _state.anonymousId
    var stored = _getStorage('_pulse_anon_id')
    if (stored) { _state.anonymousId = stored; return stored }
    var id = 'anon_' + _uuid()
    _state.anonymousId = id
    _setStorage('_pulse_anon_id', id)
    return id
  }

  function _ensureSession() {
    var now = Date.now()
    var storedSession = _getStorage('_pulse_session')
    if (storedSession) {
      try {
        var s = JSON.parse(storedSession)
        if (now - s.last < _config.sessionTimeout) {
          _state.sessionId = s.id
          s.last = now
          _setStorage('_pulse_session', JSON.stringify(s))
          return _state.sessionId
        }
      } catch (e) {}
    }
    var newId = 'sess_' + _uuid()
    _state.sessionId = newId
    _state.sessionStart = now
    _setStorage('_pulse_session', JSON.stringify({ id: newId, start: now, last: now }))
    return newId
  }

  function _send(path, payload) {
    var url = _config.apiUrl.replace(/\/+$/, '') + path
    var body = JSON.stringify(payload)

    // Beacon API: fire-and-forget, works on page unload
    if (typeof navigator !== 'undefined' && navigator.sendBeacon && path.indexOf('/events') !== -1) {
      try {
        var blob = new Blob([body], { type: 'application/json' })
        navigator.sendBeacon(url, blob)
        return Promise.resolve()
      } catch (e) {}
    }

    return fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Project-Key': _config.projectKey,
      },
      body: body,
      signal: (typeof AbortController !== 'undefined')
        ? (function () { var c = new AbortController(); setTimeout(function () { c.abort() }, _config.timeout); return c.signal })()
        : undefined,
    }).catch(function (e) { _log('send error', e) })
  }

  function _get(path) {
    var url = _config.apiUrl.replace(/\/+$/, '') + path
    return fetch(url, {
      method: 'GET',
      headers: {
        'X-Project-Key': _config.projectKey,
        Accept: 'application/json',
      },
    }).then(function (r) { return r.ok ? r.json() : Promise.reject(r.status) })
      .catch(function (e) { _log('get error', path, e); return null })
  }

  // ─── Batch queue ────────────────────────────────────────────────────────────

  function _enqueue(eventName, props) {
    var event = {
      event_name: eventName,
      user_id: _state.userId || _ensureAnonymousId(),
      session_id: _state.sessionId,
      project_key: _config.projectKey,
      url: typeof location !== 'undefined' ? location.href : undefined,
      referrer: typeof document !== 'undefined' ? document.referrer || undefined : undefined,
      user_agent: typeof navigator !== 'undefined' ? navigator.userAgent : undefined,
      timestamp: new Date().toISOString(),
      properties: Object.assign({}, _state.traits, props),
    }
    _state.queue.push(event)
    _log('enqueue', eventName, event)

    if (_state.queue.length >= _config.maxBatchSize) {
      _flush()
    }
  }

  function _flush() {
    if (_state.queue.length === 0) return
    var batch = _state.queue.splice(0, _config.maxBatchSize)
    _log('flush', batch.length, 'events')

    // Tenta batch endpoint, fallback para individual
    _send('/api/v1/events/batch', { events: batch }).catch(function () {
      batch.forEach(function (ev) {
        _send('/api/v1/events/track', ev)
      })
    })
  }

  function _startFlushTimer() {
    if (_state.flushTimer) return
    _state.flushTimer = setInterval(_flush, _config.flushInterval)

    // Flush antes de fechar a página
    if (typeof window !== 'undefined') {
      window.addEventListener('beforeunload', _flush)
      window.addEventListener('visibilitychange', function () {
        if (document.visibilityState === 'hidden') _flush()
      })
    }
  }

  // ─── Feature flags ──────────────────────────────────────────────────────────

  function _loadFlags() {
    var qs = '?project_id=' + encodeURIComponent(_config.projectKey)
    if (_state.userId) qs += '&user_id=' + encodeURIComponent(_state.userId)
    return _get('/flags' + qs).then(function (data) {
      if (data && data.flags) {
        _state.flags = data.flags
        _log('flags loaded', _state.flags)
      }
      return _state.flags
    })
  }

  function _loadVariants() {
    var qs = '?user_id=' + encodeURIComponent(_state.userId || _ensureAnonymousId())
    if (_config.projectKey) qs += '&project_id=' + encodeURIComponent(_config.projectKey)
    return _get('/flags/variants' + qs).then(function (data) {
      if (data && data.variants) {
        _state.variants = data.variants
        _log('variants loaded', _state.variants)
      }
      return _state.variants
    })
  }

  // ─── Public API ─────────────────────────────────────────────────────────────

  var pulse = {
    /**
     * Inicializa o SDK.
     * @param {Object} options
     * @param {string} options.projectKey - chave do projeto no Pulse
     * @param {string} options.apiUrl - URL base do pulse-service
     * @param {boolean} [options.debug] - habilita logs no console
     * @param {boolean} [options.autoPageview] - track automático de pageviews (default: true)
     */
    init: function (options) {
      if (_state.initialized) { _log('already initialized'); return this }

      Object.assign(_config, options)
      if (!_config.projectKey) throw new Error('[Pulse] projectKey is required')
      if (!_config.apiUrl) throw new Error('[Pulse] apiUrl is required')

      _ensureAnonymousId()
      _ensureSession()
      _startFlushTimer()

      // Carrega flags e variantes em background
      _loadFlags()
      _loadVariants()

      // Auto-pageview
      if (_config.autoPageview && typeof window !== 'undefined') {
        pulse.track('$pageview', {
          path: location.pathname,
          title: document.title,
          referrer: document.referrer || undefined,
        })

        // SPA: intercepta pushState/replaceState
        var _origPush = history.pushState
        var _origReplace = history.replaceState
        function _patchHistory(fn) {
          return function () {
            var result = fn.apply(this, arguments)
            pulse.track('$pageview', {
              path: location.pathname,
              title: document.title,
            })
            return result
          }
        }
        history.pushState = _patchHistory(_origPush)
        history.replaceState = _patchHistory(_origReplace)
      }

      _state.initialized = true
      _log('initialized', _config)
      return this
    },

    /**
     * Identifica o usuário autenticado e associa traits.
     * @param {string} userId - ID único do usuário
     * @param {Object} [traits] - atributos do usuário (name, email, plan, etc.)
     */
    identify: function (userId, traits) {
      if (!_state.initialized) { _log('not initialized'); return this }
      var prevAnon = _state.anonymousId

      _state.userId = String(userId)
      _state.traits = Object.assign({}, _state.traits, traits || {})
      _setStorage('_pulse_user_id', _state.userId)

      _log('identify', userId, traits)

      // Merge anon → user no backend
      _send('/api/v1/events/identify', {
        user_id: _state.userId,
        anonymous_id: prevAnon,
        traits: _state.traits,
        project_key: _config.projectKey,
      })

      // Re-carrega flags com contexto do usuário
      _loadFlags()
      _loadVariants()

      return this
    },

    /**
     * Registra um evento de comportamento.
     * @param {string} eventName - nome do evento (ex: 'button_clicked', 'checkout_started')
     * @param {Object} [properties] - propriedades do evento
     */
    track: function (eventName, properties) {
      if (!_state.initialized && eventName !== '$pageview') { _log('not initialized'); return this }
      if (!eventName) { _log('eventName is required'); return this }

      _ensureSession()
      _enqueue(eventName, properties || {})
      return this
    },

    /**
     * Registra visualização de página (chamado automaticamente se autoPageview=true).
     * @param {Object} [properties]
     */
    page: function (properties) {
      return this.track('$pageview', Object.assign({
        path: typeof location !== 'undefined' ? location.pathname : undefined,
        title: typeof document !== 'undefined' ? document.title : undefined,
      }, properties || {}))
    },

    /**
     * Verifica se uma feature flag está ativa para o usuário atual.
     * Usa cache em memória — carregado no init e re-carregado após identify.
     * @param {string} flagKey
     * @param {boolean} [defaultValue=false]
     * @returns {Promise<boolean>}
     */
    isEnabled: function (flagKey, defaultValue) {
      if (_state.flags !== null) {
        return Promise.resolve(_state.flags[flagKey] === true)
      }
      return _loadFlags().then(function (flags) {
        return flags ? flags[flagKey] === true : (defaultValue || false)
      })
    },

    /**
     * Retorna o valor de uma variante de experimento A/B.
     * @param {string} flagKey - chave da flag com variantes configuradas
     * @param {string} [defaultVariant='control']
     * @returns {Promise<string>} - variant key (ex: 'control', 'challenger')
     */
    getVariant: function (flagKey, defaultVariant) {
      var def = defaultVariant || 'control'
      if (_state.variants !== null) {
        return Promise.resolve(_state.variants[flagKey] || def)
      }
      return _loadVariants().then(function (variants) {
        return variants ? (variants[flagKey] || def) : def
      })
    },

    /**
     * Registra exposição a um experimento (impression).
     * Deve ser chamado quando o usuário vê a variante.
     * @param {string} flagKey
     * @param {string} variant
     */
    trackExperiment: function (flagKey, variant) {
      return this.track('$experiment_started', {
        experiment_name: flagKey,
        variant: variant,
      })
    },

    /**
     * Recarrega flags do servidor (útil após mudança de contexto).
     * @returns {Promise<Object>}
     */
    reloadFlags: function () {
      _state.flags = null
      _state.variants = null
      return Promise.all([_loadFlags(), _loadVariants()])
    },

    /**
     * Força envio imediato de todos os eventos na fila.
     */
    flush: function () {
      _flush()
      return this
    },

    /**
     * Reseta o estado (logout). Remove userId e session.
     */
    reset: function () {
      _flush()
      _state.userId = null
      _state.traits = {}
      _state.flags = null
      _state.variants = null
      _ensureAnonymousId()   // gera novo anon_id
      _setStorage('_pulse_anon_id', _state.anonymousId)
      _setStorage('_pulse_session', '')
      _ensureSession()
      return this
    },

    // ── Debug helpers ────────────────────────────────────────────────────────

    /** @internal */
    _debug: function () { return { config: _config, state: _state } },
  }

  return pulse
}))
