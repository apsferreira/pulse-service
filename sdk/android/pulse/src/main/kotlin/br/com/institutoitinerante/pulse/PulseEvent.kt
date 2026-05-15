package br.com.institutoitinerante.pulse

import org.json.JSONObject

/**
 * Representa um evento a ser enviado ao Pulse.
 */
internal data class PulseEvent(
    val projectKey: String,
    val eventName: String,
    val userId: String?,
    val properties: Map<String, Any>,
    val timestampMs: Long = System.currentTimeMillis()
) {
    fun toJson(): JSONObject = JSONObject().apply {
        put("project_key", projectKey)
        put("event_name", eventName)
        if (userId != null) put("user_id", userId)
        put("timestamp", timestampMs)
        put("properties", JSONObject(properties))
    }
}
