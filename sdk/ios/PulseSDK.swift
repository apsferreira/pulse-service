// PulseSDK for iOS — BKL-992
// Swift SDK for Pulse feature flags and event tracking.
// Supports: feature flags, event tracking, offline mode with local queue.
//
// Usage:
//   PulseSDK.shared.initialize(apiKey: "your-api-key", projectId: "proj-id")
//   let flagEnabled = await PulseSDK.shared.isEnabled("dark-mode")
//   PulseSDK.shared.track("button_clicked", properties: ["screen": "home"])

import Foundation

public final class PulseSDK {
    public static let shared = PulseSDK()

    private var apiKey: String = ""
    private var projectId: String = ""
    private var baseURL: String = "https://pulse-api.institutoitinerante.com.br"
    private var userId: String = ""
    private var flagCache: [String: Bool] = [:]
    private var offlineQueue: [[String: Any]] = []
    private let session = URLSession.shared
    private let encoder = JSONEncoder()

    private init() {}

    // MARK: — Initialization

    public func initialize(
        apiKey: String,
        projectId: String,
        baseURL: String? = nil,
        userId: String? = nil
    ) {
        self.apiKey = apiKey
        self.projectId = projectId
        if let url = baseURL { self.baseURL = url }
        self.userId = userId ?? UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
        Task { await fetchFlags() }
    }

    // MARK: — Feature Flags

    public func isEnabled(_ flagKey: String) async -> Bool {
        if let cached = flagCache[flagKey] { return cached }
        await fetchFlags()
        return flagCache[flagKey] ?? false
    }

    public func isEnabled(_ flagKey: String) -> Bool {
        flagCache[flagKey] ?? false
    }

    private func fetchFlags() async {
        guard !apiKey.isEmpty, let url = URL(string: "\(baseURL)/api/v1/flags?project_id=\(projectId)&user_id=\(userId)") else { return }
        var req = URLRequest(url: url)
        req.setValue("Bearer \(apiKey)", forHTTPHeaderField: "Authorization")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")

        do {
            let (data, _) = try await session.data(for: req)
            if let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
               let flags = json["flags"] as? [[String: Any]] {
                for flag in flags {
                    if let key = flag["key"] as? String, let enabled = flag["enabled"] as? Bool {
                        flagCache[key] = enabled
                    }
                }
            }
        } catch {
            // Silently fail — offline mode handles this
        }
    }

    // MARK: — Event Tracking

    public func track(_ eventName: String, properties: [String: String] = [:]) {
        let payload: [String: Any] = [
            "event_name": eventName,
            "user_id": userId,
            "project_id": projectId,
            "properties": properties,
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]

        Task {
            let success = await sendEvent(payload)
            if !success {
                offlineQueue.append(payload)
                persistOfflineQueue()
            }
        }
    }

    private func sendEvent(_ payload: [String: Any]) async -> Bool {
        guard !apiKey.isEmpty, let url = URL(string: "\(baseURL)/api/v1/events/track") else { return false }
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("Bearer \(apiKey)", forHTTPHeaderField: "Authorization")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")

        do {
            req.httpBody = try JSONSerialization.data(withJSONObject: payload)
            let (_, response) = try await session.data(for: req)
            if let httpResponse = response as? HTTPURLResponse {
                return httpResponse.statusCode == 200 || httpResponse.statusCode == 201
            }
        } catch {}
        return false
    }

    // MARK: — Offline Mode

    public func flushOfflineQueue() async {
        var remaining: [[String: Any]] = []
        for payload in offlineQueue {
            let success = await sendEvent(payload)
            if !success { remaining.append(payload) }
        }
        offlineQueue = remaining
        persistOfflineQueue()
    }

    private func persistOfflineQueue() {
        UserDefaults.standard.set(offlineQueue.map { $0.description }, forKey: "pulse_offline_queue")
    }

    // MARK: — User Identification

    public func identify(userId: String, traits: [String: String] = [:]) {
        self.userId = userId
        track("user.identified", properties: traits)
    }
}
