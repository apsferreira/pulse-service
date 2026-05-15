import Foundation

/// Cache em memória de feature flags com carregamento assíncrono.
/// Thread-safe via DispatchQueue serial dedicada.
internal final class FlagCache {

    private let config: PulseConfig
    private let session: URLSession
    private let queue = DispatchQueue(label: "br.com.institutoitinerante.pulse.flags", qos: .utility)
    private var cache: [String: Bool] = [:]

    init(config: PulseConfig, session: URLSession = .shared) {
        self.config = config
        self.session = session
    }

    /// Carrega as flags do servidor de forma assíncrona.
    func load() {
        guard var components = URLComponents(string: "\(config.apiUrl)/flags") else { return }
        components.queryItems = [URLQueryItem(name: "project_id", value: config.projectKey)]
        guard let url = components.url else { return }

        let task = session.dataTask(with: url) { [weak self] data, response, error in
            guard let self else { return }

            if let error = error {
                self.debugLog("Falha ao carregar flags: \(error.localizedDescription)")
                return
            }

            guard
                let http = response as? HTTPURLResponse,
                (200...299).contains(http.statusCode),
                let data = data
            else {
                self.debugLog("Resposta inválida ao carregar flags")
                return
            }

            self.parseAndCache(data)
        }
        task.resume()
    }

    /// Retorna se uma flag está habilitada, usando `defaultValue` se não encontrada.
    func isEnabled(_ key: String, defaultValue: Bool) -> Bool {
        queue.sync { cache[key] ?? defaultValue }
    }

    // MARK: - Privado

    private func parseAndCache(_ data: Data) {
        // Suporte a dois formatos de resposta:
        // 1. Objeto: { "flags": [{ "key": "...", "enabled": true }] }
        // 2. Array direto: [{ "key": "...", "enabled": true }]
        struct FlagItem: Decodable {
            let key: String
            let enabled: Bool
        }
        struct FlagResponse: Decodable {
            let flags: [FlagItem]
        }

        let parsed: [FlagItem]
        if let wrapped = try? JSONDecoder().decode(FlagResponse.self, from: data) {
            parsed = wrapped.flags
        } else if let direct = try? JSONDecoder().decode([FlagItem].self, from: data) {
            parsed = direct
        } else {
            debugLog("Formato de flags desconhecido")
            return
        }

        queue.async { [weak self] in
            guard let self else { return }
            self.cache = Dictionary(uniqueKeysWithValues: parsed.map { ($0.key, $0.enabled) })
            self.debugLog("Flags carregadas: \(self.cache.keys.joined(separator: ", "))")
        }
    }

    private func debugLog(_ message: String) {
        if config.debug {
            print("[Pulse:Flags] \(message)")
        }
    }
}
