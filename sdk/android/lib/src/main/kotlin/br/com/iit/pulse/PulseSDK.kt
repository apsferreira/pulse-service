// PulseSDK for Android — BKL-992
// Kotlin SDK for Pulse feature flags and event tracking.
// Supports: feature flags, event tracking, offline mode with local queue.
//
// Usage:
//   PulseSDK.init(context, apiKey = "your-api-key", projectId = "proj-id")
//   val flagEnabled = PulseSDK.isEnabled("dark-mode")
//   PulseSDK.track("button_clicked", mapOf("screen" to "home"))

package br.com.iit.pulse

import android.content.Context
import android.provider.Settings
import kotlinx.coroutines.*
import org.json.JSONArray
import org.json.JSONObject
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL
import java.text.SimpleDateFormat
import java.util.*

object PulseSDK {
    private var apiKey: String = ""
    private var projectId: String = ""
    private var baseURL: String = ""
    private var userId: String = ""
    private val flagCache = mutableMapOf<String, Boolean>()
    private val offlineQueue = mutableListOf<Map<String, Any>>()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val isoFormat = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US).apply {
        timeZone = TimeZone.getTimeZone("UTC")
    }

    // MARK: — Initialization

    fun init(
        context: Context,
        apiKey: String,
        projectId: String,
        baseURL: String? = null,
        userId: String? = null
    ) {
        this.apiKey = apiKey
        this.projectId = projectId
        baseURL?.let { this.baseURL = it }
        this.userId = userId
            ?: Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
            ?: UUID.randomUUID().toString()

        scope.launch { fetchFlags() }
    }

    // MARK: — Feature Flags

    fun isEnabled(flagKey: String): Boolean = flagCache[flagKey] ?: false

    suspend fun isEnabledAsync(flagKey: String): Boolean {
        if (flagCache.containsKey(flagKey)) return flagCache[flagKey]!!
        fetchFlags()
        return flagCache[flagKey] ?: false
    }

    private suspend fun fetchFlags() {
        if (apiKey.isEmpty()) return
        try {
            val url = URL("$baseURL/api/v1/flags?project_id=$projectId&user_id=$userId")
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "GET"
            conn.setRequestProperty("Authorization", "Bearer $apiKey")
            conn.setRequestProperty("Content-Type", "application/json")

            if (conn.responseCode == 200) {
                val response = conn.inputStream.bufferedReader().readText()
                val json = JSONObject(response)
                val flags = json.optJSONArray("flags") ?: JSONArray()
                for (i in 0 until flags.length()) {
                    val flag = flags.getJSONObject(i)
                    flagCache[flag.getString("key")] = flag.getBoolean("enabled")
                }
            }
            conn.disconnect()
        } catch (_: Exception) {
            // Silently fail — offline mode handles this
        }
    }

    // MARK: — Event Tracking

    fun track(eventName: String, properties: Map<String, String> = emptyMap()) {
        val payload = mapOf(
            "event_name" to eventName,
            "user_id" to userId,
            "project_id" to projectId,
            "properties" to properties,
            "timestamp" to isoFormat.format(Date())
        )

        scope.launch {
            val success = sendEvent(payload)
            if (!success) {
                synchronized(offlineQueue) { offlineQueue.add(payload) }
            }
        }
    }

    private suspend fun sendEvent(payload: Map<String, Any>): Boolean {
        if (apiKey.isEmpty()) return false
        return try {
            val url = URL("$baseURL/api/v1/events/track")
            val conn = url.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            conn.setRequestProperty("Authorization", "Bearer $apiKey")
            conn.setRequestProperty("Content-Type", "application/json")
            conn.doOutput = true

            val json = JSONObject(payload).toString()
            OutputStreamWriter(conn.outputStream).use { it.write(json) }

            val success = conn.responseCode in 200..201
            conn.disconnect()
            success
        } catch (_: Exception) { false }
    }

    // MARK: — Offline Mode

    suspend fun flushOfflineQueue() {
        val remaining = mutableListOf<Map<String, Any>>()
        synchronized(offlineQueue) {
            offlineQueue.toList().also { offlineQueue.clear() }
        }.forEach { payload ->
            if (!sendEvent(payload)) remaining.add(payload)
        }
        synchronized(offlineQueue) { offlineQueue.addAll(remaining) }
    }

    // MARK: — User Identification

    fun identify(userId: String, traits: Map<String, String> = emptyMap()) {
        this.userId = userId
        track("user.identified", traits)
    }
}
