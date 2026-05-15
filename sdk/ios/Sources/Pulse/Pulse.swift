import Foundation

/// SDK Pulse para iOS/macOS — event tracking e feature flags.
///
/// ## Uso básico
/// ```swift
/// // AppDelegate.application(_:didFinishLaunchingWithOptions:)
/// Pulse.shared.configure(projectKey: "meu-projeto", userId: "user-123")
///
/// // Em qualquer parte do app:
/// Pulse.shared.track("book_opened", properties: ["book_id": "abc"])
/// let enabled = Pulse.shared.isEnabled("nova_feature")
/// ```
public final class Pulse {

    // MARK: - Singleton

    /// Instância compartilhada do SDK.
    public static let shared = Pulse()

    // MARK: - Estado privado (thread-safe via queue interna)

    private var config: PulseConfig?
    private var eventQueue: EventQueue?
    private var flagCache: FlagCache?
    private let stateQueue = DispatchQueue(label: "br.com.institutoitinerante.pulse.state", qos: .utility)

    private let userDefaultsUserIdKey = "br.com.institutoitinerante.pulse.user_id"
    private var _userId: String?

    private var userId: String? {
        get { stateQueue.sync { _userId } }
        set { stateQueue.async { [weak self] in self?._userId = newValue } }
    }

    // MARK: - Init privado (singleton)

    private init() {}

    // MARK: - API pública

    /// Configura o SDK. Deve ser chamado em `AppDelegate.application(_:didFinishLaunchingWithOptions:)`.
    ///
    /// - Parameters:
    ///   - projectKey: Chave do projeto configurada no painel Pulse.
    ///   - userId: Identificador opcional do usuário (pode ser definido depois com `identify`).
    ///   - config: Configurações avançadas (opcional).
    public func configure(
        projectKey: String,
        userId: String? = nil,
        config: PulseConfig? = nil
    ) {
        let effectiveConfig = config ?? PulseConfig(projectKey: projectKey)

        stateQueue.async { [weak self] in
            guard let self else { return }

            self.config = effectiveConfig
            self.eventQueue = EventQueue(config: effectiveConfig)
            self.flagCache = FlagCache(config: effectiveConfig)

            // Restaura userId persistido, sobrescrito se fornecido agora
            if let uid = userId {
                self._userId = uid
                UserDefaults.standard.set(uid, forKey: self.userDefaultsUserIdKey)
            } else {
                self._userId = UserDefaults.standard.string(forKey: self.userDefaultsUserIdKey)
            }

            self.flagCache?.load()

            if effectiveConfig.debug {
                print("[Pulse] Configurado. project=\(projectKey) userId=\(self._userId ?? "nil")")
            }
        }
    }

    /// Associa o usuário atual ao tracking. Persiste entre sessões via UserDefaults.
    public func identify(userId: String) {
        stateQueue.async { [weak self] in
            guard let self else { return }
            self._userId = userId
            UserDefaults.standard.set(userId, forKey: self.userDefaultsUserIdKey)
            self.config.map { if $0.debug { print("[Pulse] identify: \(userId)") } }
        }
    }

    /// Registra um evento de tracking.
    ///
    /// - Parameters:
    ///   - eventName: Nome do evento (ex: `"book_opened"`, `"focus_session_started"`).
    ///   - properties: Propriedades adicionais do evento (valores: String, Int, Double, Bool).
    @discardableResult
    public func track(_ eventName: String, properties: [String: Any] = [:]) -> Self {
        stateQueue.async { [weak self] in
            guard let self, let queue = self.eventQueue, let cfg = self.config else {
                print("[Pulse] AVISO: track() chamado antes de configure(). Evento descartado: \(eventName)")
                return
            }
            let event = PulseEvent(
                projectKey: cfg.projectKey,
                eventName: eventName,
                userId: self._userId,
                properties: properties
            )
            queue.enqueue(event)
        }
        return self
    }

    /// Retorna se uma feature flag está habilitada.
    /// Usa cache em memória — não bloqueia a thread principal.
    ///
    /// - Parameters:
    ///   - flagKey: Chave da flag configurada no painel Pulse.
    ///   - defaultValue: Valor padrão caso a flag não seja encontrada.
    public func isEnabled(_ flagKey: String, defaultValue: Bool = false) -> Bool {
        return stateQueue.sync { flagCache?.isEnabled(flagKey, defaultValue: defaultValue) ?? defaultValue }
    }

    /// Força o envio imediato de todos os eventos em fila.
    /// Útil antes de `applicationWillTerminate` ou antes de fazer logout.
    public func flush() {
        stateQueue.async { [weak self] in
            self?.eventQueue?.flush()
        }
    }
}
