package br.com.institutoitinerante.pulse

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import okhttp3.Call
import okhttp3.Callback
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import org.json.JSONArray
import org.json.JSONObject
import java.io.IOException
import java.util.concurrent.Executors
import java.util.concurrent.ScheduledFuture
import java.util.concurrent.TimeUnit
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

private const val TAG = "PulseEventQueue"
private const val PREFS_NAME = "pulse_sdk_prefs"
private const val PREFS_KEY_PENDING = "pending_events"
private const val JSON_MEDIA_TYPE = "application/json; charset=utf-8"

/**
 * Fila de eventos com batch automático, flush periódico e persistência em SharedPreferences.
 * Toda I/O de rede ocorre em um executor de thread única — nunca bloqueia a main thread.
 */
internal class EventQueue(
    context: Context,
    private val config: PulseConfig,
    private val httpClient: OkHttpClient
) {
    private val appContext: Context = context.applicationContext
    private val prefs: SharedPreferences = appContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    private val lock = ReentrantLock()
    private val queue: ArrayDeque<PulseEvent> = ArrayDeque()
    private val executor = Executors.newSingleThreadScheduledExecutor { r ->
        Thread(r, "pulse-worker").also { it.isDaemon = true }
    }
    private var scheduledFlush: ScheduledFuture<*>? = null

    init {
        restorePendingEvents()
        schedulePeriodicFlush()
    }

    fun enqueue(event: PulseEvent) {
        lock.withLock {
            queue.addLast(event)
            debugLog("Evento enfileirado: ${event.eventName} (fila: ${queue.size})")
            if (queue.size >= config.batchSize) {
                flushAsync()
            }
        }
    }

    fun flush() {
        flushAsync()
    }

    fun shutdown() {
        scheduledFlush?.cancel(false)
        executor.shutdown()
    }

    private fun schedulePeriodicFlush() {
        scheduledFlush = executor.scheduleAtFixedRate(
            ::flushAsync,
            config.flushIntervalMs,
            config.flushIntervalMs,
            TimeUnit.MILLISECONDS
        )
    }

    private fun flushAsync() {
        executor.execute {
            val batch = lock.withLock {
                if (queue.isEmpty()) return@execute
                val snapshot = queue.toList()
                queue.clear()
                snapshot
            }
            sendBatch(batch)
        }
    }

    private fun sendBatch(batch: List<PulseEvent>) {
        val payload = buildBatchPayload(batch)
        val body = payload.toString().toRequestBody(JSON_MEDIA_TYPE.toMediaType())
        val request = Request.Builder()
            .url("${config.apiUrl}/events/batch")
            .post(body)
            .build()

        debugLog("Enviando batch com ${batch.size} evento(s)")

        httpClient.newCall(request).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                Log.w(TAG, "Falha ao enviar batch, persistindo ${batch.size} evento(s): ${e.message}")
                persistPendingEvents(batch)
            }

            override fun onResponse(call: Call, response: Response) {
                response.use {
                    if (!response.isSuccessful) {
                        Log.w(TAG, "Batch rejeitado (HTTP ${response.code}), persistindo ${batch.size} evento(s)")
                        persistPendingEvents(batch)
                    } else {
                        debugLog("Batch enviado com sucesso (HTTP ${response.code})")
                        clearPersistedEvents()
                    }
                }
            }
        })
    }

    private fun buildBatchPayload(batch: List<PulseEvent>): JSONObject {
        val eventsArray = JSONArray()
        batch.forEach { eventsArray.put(it.toJson()) }
        return JSONObject().apply {
            put("events", eventsArray)
        }
    }

    // --- Persistência offline ---

    private fun persistPendingEvents(events: List<PulseEvent>) {
        try {
            val array = JSONArray()
            events.forEach { array.put(it.toJson()) }
            prefs.edit().putString(PREFS_KEY_PENDING, array.toString()).apply()
        } catch (e: Exception) {
            Log.e(TAG, "Erro ao persistir eventos: ${e.message}")
        }
    }

    private fun restorePendingEvents() {
        val raw = prefs.getString(PREFS_KEY_PENDING, null) ?: return
        try {
            val array = JSONArray(raw)
            for (i in 0 until array.length()) {
                val obj = array.getJSONObject(i)
                val props = mutableMapOf<String, Any>()
                val propsObj = obj.optJSONObject("properties")
                propsObj?.keys()?.forEach { key -> props[key] = propsObj.get(key) }

                queue.addLast(
                    PulseEvent(
                        projectKey = obj.getString("project_key"),
                        eventName = obj.getString("event_name"),
                        userId = obj.optString("user_id").ifBlank { null },
                        properties = props,
                        timestampMs = obj.optLong("timestamp", System.currentTimeMillis())
                    )
                )
            }
            debugLog("${queue.size} evento(s) restaurado(s) do armazenamento local")
            clearPersistedEvents()
        } catch (e: Exception) {
            Log.e(TAG, "Erro ao restaurar eventos persistidos: ${e.message}")
        }
    }

    private fun clearPersistedEvents() {
        prefs.edit().remove(PREFS_KEY_PENDING).apply()
    }

    private fun debugLog(message: String) {
        if (config.debug) Log.d(TAG, message)
    }
}
