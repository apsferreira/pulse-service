import Foundation

/// Configuração do SDK Pulse.
public struct PulseConfig {
    /// Chave do projeto no Pulse (obrigatório).
    public let projectKey: String

    /// Base URL da API Pulse.
    public let apiUrl: String

    /// Número máximo de eventos por batch antes do flush automático.
    public let batchSize: Int

    /// Intervalo de flush periódico em segundos.
    public let flushInterval: TimeInterval

    /// Habilita logs detalhados no console.
    public let debug: Bool

    public init(
        projectKey: String,
        apiUrl: String = "https://pulse-api.institutoitinerante.com.br/api/v1",
        batchSize: Int = 10,
        flushInterval: TimeInterval = 30,
        debug: Bool = false
    ) {
        self.projectKey = projectKey
        self.apiUrl = apiUrl
        self.batchSize = batchSize
        self.flushInterval = flushInterval
        self.debug = debug
    }
}
