# Pulse SDK — Android

SDK nativo para Android do Pulse, plataforma de event tracking e feature flags do Instituto Itinerante.

- minSdk: 24 (Android 7.0)
- Kotlin 1.9+
- Dependência única: OkHttp 4.x
- Zero reflexão, zero code generation

---

## Instalação

### Via GitHub Packages (recomendado)

No `settings.gradle.kts` do projeto raiz:

```kotlin
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
        maven {
            url = uri("https://maven.pkg.github.com/apsferreira/pulse-sdk-android")
            credentials {
                username = providers.gradleProperty("gpr.user").orNull ?: System.getenv("GITHUB_ACTOR")
                password = providers.gradleProperty("gpr.key").orNull ?: System.getenv("GITHUB_TOKEN")
            }
        }
    }
}
```

No `build.gradle.kts` do módulo app:

```kotlin
dependencies {
    implementation("br.com.institutoitinerante:pulse:1.0.0")
}
```

### Via JAR local (desenvolvimento)

```bash
./gradlew :pulse:assembleRelease
# AAR gerado em pulse/build/outputs/aar/pulse-release.aar
```

---

## Configuração

### Application.kt

```kotlin
class MyApplication : Application() {
    override fun onCreate() {
        super.onCreate()

        Pulse.init(
            context = this,
            projectKey = "meu-projeto-key",
            userId = "user-123",            // opcional — pode ser definido depois com identify()
            config = PulseConfig(
                projectKey = "meu-projeto-key",
                batchSize = 10,
                flushIntervalMs = 30_000,
                debug = BuildConfig.DEBUG
            )
        )
    }
}
```

Registre no `AndroidManifest.xml`:

```xml
<application
    android:name=".MyApplication"
    ...>
```

---

## Uso

### Identificar o usuário

```kotlin
// Após login bem-sucedido:
Pulse.identify(userId = "user-abc-123")
```

### Registrar eventos

```kotlin
// Evento simples
Pulse.track("app_opened")

// Evento com propriedades
Pulse.track(
    eventName = "book_opened",
    properties = mapOf(
        "book_id" to "978-3-16-148410-0",
        "title" to "O Pequeno Príncipe",
        "source" to "search"
    )
)

// Forçar envio imediato (ex: antes de fazer logout)
Pulse.flush()
```

### Feature flags

```kotlin
// Verificação com valor padrão
if (Pulse.isEnabled("novo_leitor_pdf")) {
    // mostra nova experiência
}

// Com default explícito
val habilitado = Pulse.isEnabled("modo_noturno_v2", defaultValue = true)
```

---

## Eventos recomendados

### Librinho (my-library)

| Evento | Propriedades sugeridas |
|--------|------------------------|
| `app_opened` | `source` (push, direct, link) |
| `book_scanned` | `isbn`, `method` (camera, manual) |
| `book_added_to_library` | `book_id`, `isbn`, `title` |
| `book_opened` | `book_id`, `title`, `format` |
| `book_reading_started` | `book_id`, `page_start` |
| `book_reading_finished` | `book_id`, `duration_seconds` |
| `book_rating_submitted` | `book_id`, `rating` (1-5) |
| `book_removed` | `book_id` |
| `search_performed` | `query`, `results_count` |
| `library_filtered` | `filter_type`, `value` |

### Nitro (focus-hub)

| Evento | Propriedades sugeridas |
|--------|------------------------|
| `app_opened` | `source` |
| `focus_session_started` | `duration_minutes`, `mode` (pomodoro, custom) |
| `focus_session_completed` | `duration_minutes`, `interrupted` (bool) |
| `focus_session_abandoned` | `elapsed_seconds`, `reason` |
| `task_created` | `task_id`, `has_due_date` (bool) |
| `task_completed` | `task_id`, `elapsed_days` |
| `task_deleted` | `task_id` |
| `habit_checked_in` | `habit_id`, `streak` |
| `habit_created` | `habit_id`, `frequency` |
| `goal_created` | `goal_id`, `deadline` |
| `goal_completed` | `goal_id`, `duration_days` |
| `notification_tapped` | `type`, `source` |

---

## Comportamento interno

- Eventos são acumulados em memória e enviados em batch (default: 10 eventos ou 30s).
- Eventos não entregues por falha de rede são persistidos em `SharedPreferences` e reenviados na próxima sessão.
- Feature flags são carregadas assincronamente na inicialização e mantidas em cache em memória.
- Toda I/O de rede roda em uma thread dedicada (`pulse-worker`) — nunca bloqueia a main thread.
