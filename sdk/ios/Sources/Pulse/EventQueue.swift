import Foundation

/// Modelo interno de um evento de tracking.
internal struct PulseEvent: Codable {
    let projectKey: String
    let eventName: String
    let userId: String?
    let properties: [String: AnyCodable]
    let timestamp: TimeInterval

    init(
        projectKey: String,
        eventName: String,
        userId: String?,
        properties: [String: Any]
    ) {
        self.projectKey = projectKey
        self.eventName = eventName
        self.userId = userId
        self.properties = properties.mapValues { AnyCodable($0) }
        self.timestamp = Date().timeIntervalSince1970 * 1000 // ms
    }

    enum CodingKeys: String, CodingKey {
        case projectKey = "project_key"
        case eventName = "event_name"
        case userId = "user_id"
        case properties
        case timestamp
    }
}

/// Fila de eventos com batch automático, flush periódico e persistência em UserDefaults.
/// Thread-safe via DispatchQueue serial dedicada.
internal final class EventQueue {

    private let config: PulseConfig
    private let session: URLSession
    private let queue = DispatchQueue(label: "br.com.institutoitinerante.pulse.queue", qos: .utility)
    private var buffer: [PulseEvent] = []
    private var flushTimer: Timer?
    private let userDefaultsKey = "br.com.institutoitinerante.pulse.pending_events"

    init(config: PulseConfig, session: URLSession = .shared) {
        self.config = config
        self.session = session
        restorePendingEvents()
        scheduleTimer()
    }

    deinit {
        flushTimer?.invalidate()
    }

    func enqueue(_ event: PulseEvent) {
        queue.async { [weak self] in
            guard let self else { return }
            self.buffer.append(event)
            self.debugLog("Evento enfileirado: \(event.eventName) (fila: \(self.buffer.count))")
            if self.buffer.count >= self.config.batchSize {
                self.flushLocked()
            }
        }
    }

    func flush() {
        queue.async { [weak self] in
            self?.flushLocked()
        }
    }

    // MARK: - Privado (sempre chamado dentro de queue)

    private func flushLocked() {
        guard !buffer.isEmpty else { return }
        let batch = buffer
        buffer.removeAll()
        sendBatch(batch)
    }

    private func scheduleTimer() {
        DispatchQueue.main.async { [weak self] in
            guard let self else { return }
            self.flushTimer = Timer.scheduledTimer(
                withTimeInterval: self.config.flushInterval,
                repeats: true
            ) { [weak self] _ in
                self?.flush()
            }
        }
    }

    private func sendBatch(_ batch: [PulseEvent]) {
        guard let url = URL(string: "\(config.apiUrl)/events/batch") else { return }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let payload = BatchPayload(events: batch)
        do {
            request.httpBody = try JSONEncoder().encode(payload)
        } catch {
            debugLog("Erro ao serializar batch: \(error)")
            return
        }

        debugLog("Enviando batch com \(batch.count) evento(s)")

        let task = session.dataTask(with: request) { [weak self] _, response, error in
            guard let self else { return }

            if let error = error {
                self.debugLog("Falha ao enviar batch: \(error.localizedDescription)")
                self.persistPendingEvents(batch)
                return
            }

            if let http = response as? HTTPURLResponse, !(200...299).contains(http.statusCode) {
                self.debugLog("Batch rejeitado (HTTP \(http.statusCode)), persistindo \(batch.count) evento(s)")
                self.persistPendingEvents(batch)
                return
            }

            self.debugLog("Batch enviado com sucesso")
            self.clearPersistedEvents()
        }
        task.resume()
    }

    // MARK: - Persistência offline

    private func persistPendingEvents(_ events: [PulseEvent]) {
        queue.async { [weak self] in
            guard let self else { return }
            do {
                let data = try JSONEncoder().encode(events)
                UserDefaults.standard.set(data, forKey: self.userDefaultsKey)
            } catch {
                self.debugLog("Erro ao persistir eventos: \(error)")
            }
        }
    }

    private func restorePendingEvents() {
        // Chamado em init — ainda sem lock de queue necessário
        guard let data = UserDefaults.standard.data(forKey: userDefaultsKey) else { return }
        do {
            let events = try JSONDecoder().decode([PulseEvent].self, from: data)
            buffer.append(contentsOf: events)
            debugLog("\(events.count) evento(s) restaurado(s) do armazenamento local")
            clearPersistedEvents()
        } catch {
            debugLog("Erro ao restaurar eventos persistidos: \(error)")
        }
    }

    private func clearPersistedEvents() {
        UserDefaults.standard.removeObject(forKey: userDefaultsKey)
    }

    private func debugLog(_ message: String) {
        if config.debug {
            print("[Pulse] \(message)")
        }
    }
}

// MARK: - Tipos auxiliares de codificação

private struct BatchPayload: Encodable {
    let events: [PulseEvent]
}

/// Wrapper Codable para valores heterogêneos em propriedades de eventos.
internal struct AnyCodable: Codable {
    let value: Any

    init(_ value: Any) {
        self.value = value
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch value {
        case let v as Bool:   try container.encode(v)
        case let v as Int:    try container.encode(v)
        case let v as Double: try container.encode(v)
        case let v as Float:  try container.encode(Double(v))
        case let v as String: try container.encode(v)
        case let v as [String: Any]:
            try container.encode(v.mapValues { AnyCodable($0) })
        case let v as [Any]:
            try container.encode(v.map { AnyCodable($0) })
        default:
            try container.encode(String(describing: value))
        }
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if let v = try? container.decode(Bool.self)   { value = v; return }
        if let v = try? container.decode(Int.self)    { value = v; return }
        if let v = try? container.decode(Double.self) { value = v; return }
        if let v = try? container.decode(String.self) { value = v; return }
        value = NSNull()
    }
}
