package br.com.institutoitinerante.pulse

import android.content.Context
import android.util.Log
import okhttp3.Call
import okhttp3.Callback
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import org.json.JSONArray
import org.json.JSONObject
import java.io.IOException
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference

private const val TAG = "Pulse"

/**
 * SDK Pulse para Android — event tracking e feature flags.
 *
 * Uso básico:
 * ```kotlin
 * // Application.onCreate()
 * Pulse.init(context, projectKey = "meu-projeto", userId = "user-123")
 *
 * // Em qualquer ponto do app:
 * Pulse.track("book_opened", mapOf("book_id" to "abc"))
 * val enabled = Pulse.isEnabled("nova_feature")
 * ```
 */
object Pulse {

    private val instance: AtomicReference<PulseInstance?> = AtomicReference(null)

    /**
     * Inicializa o SDK. Deve ser chamado em [android.app.Application.onCreate].
     *
     * @param context    Contexto da aplicação (Application ou Activity).
     * @param projectKey Chave do projeto configurada no painel Pulse.
     * @param userId     Identificador opcional do usuário atual.
     * @param config     Configurações avançadas (opcional — usa defaults se omitido).
     */
    @JvmStatic
    @JvmOverloads
    fun init(
        context: Context,
        projectKey: String,
        userId: String? = null,
        config: PulseConfig = PulseConfig(projectKey = projectKey)
    ) {
        val effectiveConfig = config.copy(projectKey = projectKey)
        instance.set(PulseInstance(context, effectiveConfig, userId))
        if (effectiveConfig.debug) Log.d(TAG, "Pulse inicializado. project=$projectKey userId=$userId")
    }

    /**
     * Associa o usuário atual ao tracking. Persiste entre sessões.
     */
    @JvmStatic
    fun identify(userId: String) {
        requireInstance().identify(userId)
    }

    /**
     * Registra um evento de tracking.
     *
     * @param eventName  Nome do evento (ex: "book_opened", "focus_session_started").
     * @param properties Propriedades adicionais do evento.
     */
    @JvmStatic
    @JvmOverloads
    fun track(eventName: String, properties: Map<String, Any> = emptyMap()) {
        requireInstance().track(eventName, properties)
    }

    /**
     * Retorna se uma feature flag está habilitada.
     * Usa cache em memória — não bloqueia a thread principal.
     *
     * @param flagKey      Chave da flag configurada no painel Pulse.
     * @param defaultValue Valor padrão caso a flag não seja encontrada.
     */
    @JvmStatic
    @JvmOverloads
    fun isEnabled(flagKey: String, defaultValue: Boolean = false): Boolean {
        return requireInstance().isEnabled(flagKey, defaultValue)
    }

    /**
     * Força o envio imediato de todos os eventos em fila.
     * Útil antes de encerrar a sessão ou navegar para fora do app.
     */
    @JvmStatic
    fun flush() {
        requireInstance().flush()
    }

    /**
     * Retorna a instância interna (uso avançado / testes).
     */
    @JvmStatic
    fun getInstance(): PulseInstance? = instance.get()

    private fun requireInstance(): PulseInstance {
        return instance.get()
            ?: error("Pulse não foi inicializado. Chame Pulse.init() em Application.onCreate().")
    }
}

/**
 * Implementação interna — separada do singleton para facilitar testes.
 */
internal class PulseInstance(
    context: Context,
    private val config: PulseConfig,
    initialUserId: String?
) {
    private val appContext = context.applicationContext
    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(10, TimeUnit.SECONDS)
        .writeTimeout(10, TimeUnit.SECONDS)
        .build()

    private val eventQueue = EventQueue(appContext, config, httpClient)
    private val flagCache = mutableMapOf<String, Boolean>()
    private var userId: String? = null

    init {
        initialUserId?.let { identify(it) }
        fetchFlags()
    }

    fun identify(userId: String) {
        this.userId = userId
        if (config.debug) Log.d("Pulse", "identify: $userId")
    }

    fun track(eventName: String, properties: Map<String, Any>) {
        val event = PulseEvent(
            projectKey = config.projectKey,
            eventName = eventName,
            userId = userId,
            properties = properties
        )
        eventQueue.enqueue(event)
    }

    fun isEnabled(flagKey: String, defaultValue: Boolean): Boolean {
        return flagCache[flagKey] ?: defaultValue
    }

    fun flush() {
        eventQueue.flush()
    }

    // --- Feature Flags ---

    private fun fetchFlags() {
        val url = "${config.apiUrl}/flags?project_id=${config.projectKey}"
        val request = Request.Builder().url(url).get().build()

        httpClient.newCall(request).enqueue(object : Callback {
            override fun onFailure(call: Call, e: IOException) {
                Log.w("Pulse", "Falha ao buscar feature flags: ${e.message}")
            }

            override fun onResponse(call: Call, response: Response) {
                response.use {
                    if (!response.isSuccessful) return
                    try {
                        val body = response.body?.string() ?: return
                        parseAndCacheFlags(body)
                        if (config.debug) Log.d("Pulse", "Flags carregadas: ${flagCache.keys}")
                    } catch (e: Exception) {
                        Log.e("Pulse", "Erro ao parsear flags: ${e.message}")
                    }
                }
            }
        })
    }

    private fun parseAndCacheFlags(json: String) {
        // Suporte a dois formatos: array de objetos ou objeto com campo "flags"
        val root = JSONObject(json)
        val array: JSONArray = when {
            root.has("flags") -> root.getJSONArray("flags")
            else -> JSONArray(json)
        }
        flagCache.clear()
        for (i in 0 until array.length()) {
            val flag = array.getJSONObject(i)
            val key = flag.optString("key").ifBlank { continue }
            flagCache[key] = flag.optBoolean("enabled", false)
        }
    }
}
