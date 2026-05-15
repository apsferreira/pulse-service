package br.com.institutoitinerante.pulse

/**
 * Configuração do SDK Pulse.
 *
 * @param projectKey  Chave do projeto no Pulse (obrigatório).
 * @param apiUrl      Base URL da API Pulse.
 * @param batchSize   Número máximo de eventos por batch antes do flush automático.
 * @param flushIntervalMs Intervalo de flush periódico em milissegundos.
 * @param debug       Habilita logs detalhados via android.util.Log.
 */
data class PulseConfig(
    val projectKey: String,
    val apiUrl: String = "https://pulse-api.institutoitinerante.com.br/api/v1",
    val batchSize: Int = 10,
    val flushIntervalMs: Long = 30_000,
    val debug: Boolean = false
)
