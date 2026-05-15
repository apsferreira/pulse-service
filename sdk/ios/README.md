# Pulse SDK — iOS

SDK nativo para iOS/macOS do Pulse, plataforma de event tracking e feature flags do Instituto Itinerante.

- iOS 14+ / macOS 11+
- Swift 5.9+
- Zero dependências externas (URLSession nativo)
- Swift Package Manager

---

## Instalacao

### Swift Package Manager (Xcode)

1. No Xcode: **File > Add Package Dependencies...**
2. Cole a URL: `https://github.com/apsferreira/pulse-sdk-ios`
3. Selecione a versao: `1.0.0` ou `branch: main`

### Package.swift (para libs e frameworks)

```swift
dependencies: [
    .package(url: "https://github.com/apsferreira/pulse-sdk-ios", from: "1.0.0")
],
targets: [
    .target(name: "MeuApp", dependencies: ["Pulse"])
]
```

---

## Configuracao

### AppDelegate

```swift
import Pulse

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {

        Pulse.shared.configure(
            projectKey: "meu-projeto-key",
            userId: AuthService.currentUserId, // opcional
            config: PulseConfig(
                projectKey: "meu-projeto-key",
                batchSize: 10,
                flushInterval: 30,
                debug: true // so em debug build
            )
        )

        return true
    }

    func applicationWillTerminate(_ application: UIApplication) {
        Pulse.shared.flush()
    }
}
```

### SwiftUI — App struct

```swift
import Pulse
import SwiftUI

@main
struct MeuApp: App {
    init() {
        Pulse.shared.configure(
            projectKey: "meu-projeto-key",
            config: PulseConfig(
                projectKey: "meu-projeto-key",
                debug: ProcessInfo.processInfo.environment["PULSE_DEBUG"] != nil
            )
        )
    }

    var body: some Scene {
        WindowGroup { ContentView() }
    }
}
```

---

## Uso

### Identificar o usuario

```swift
// Apos login bem-sucedido:
Pulse.shared.identify(userId: "user-abc-123")
```

O `userId` e persistido em `UserDefaults` e restaurado automaticamente na proxima sessao.

### Registrar eventos

```swift
// Evento simples
Pulse.shared.track("app_opened")

// Evento com propriedades
Pulse.shared.track("book_opened", properties: [
    "book_id": "978-3-16-148410-0",
    "title": "O Pequeno Principe",
    "source": "search"
])

// Encadeamento (fluent)
Pulse.shared
    .track("checkout_started", properties: ["plan": "pro"])

// Forcar envio imediato (ex: antes de logout)
Pulse.shared.flush()
```

### Feature flags

```swift
// Verificacao com valor padrao false
if Pulse.shared.isEnabled("novo_leitor_pdf") {
    // mostra nova experiencia
}

// Com default explicito
let habilitado = Pulse.shared.isEnabled("modo_noturno_v2", defaultValue: true)

// Uso em SwiftUI
struct ContentView: View {
    var body: some View {
        Group {
            if Pulse.shared.isEnabled("redesign_home") {
                HomeViewV2()
            } else {
                HomeViewV1()
            }
        }
    }
}
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

- Eventos sao acumulados em uma fila e enviados em batch (default: 10 eventos ou 30s).
- Eventos nao entregues por falha de rede sao persistidos em `UserDefaults` e reenviados na proxima sessao.
- Feature flags sao carregadas assincronamente na inicializacao e mantidas em cache em memoria.
- Thread-safe via `DispatchQueue` serial dedicada por componente.
- `Timer` de flush agendado na main thread para compatibilidade com RunLoop do iOS.
- Zero dependencias externas — apenas URLSession e Foundation.
