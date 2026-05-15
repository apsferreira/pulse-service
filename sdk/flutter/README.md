# pulse_sdk

SDK Flutter para o **Pulse** — plataforma de event tracking e feature flags do Instituto Itinerante.

## Instalação

Adicione ao `pubspec.yaml` do seu projeto:

```yaml
dependencies:
  pulse_sdk:
    git:
      url: https://github.com/apsferreira/pulse-sdk-flutter.git
      ref: main
```

Ou, para desenvolvimento local:

```yaml
dependencies:
  pulse_sdk:
    path: ../pulse-sdk-flutter
```

## Setup

Inicialize o SDK em `main()`, antes de `runApp()`:

```dart
import 'package:flutter/material.dart';
import 'package:pulse_sdk/pulse.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Pulse.init(
    projectKey: 'librinho',   // chave do projeto no Pulse
    debug: true,              // habilita logs no console (desabilite em produção)
  );

  runApp(const MyApp());
}
```

### Identificar usuário após login

```dart
void onLoginSuccess(String userId) {
  Pulse.instance.identify(userId);
}

void onLogout() {
  Pulse.instance.reset();
}
```

### Rastrear eventos

```dart
Pulse.instance.track('book_scanned', properties: {
  'isbn': '978-85-359-0277-5',
  'source': 'camera',
});
```

### Feature flags

```dart
final enabled = await Pulse.instance.isEnabled('new-reader-ui');

if (enabled) {
  // exibe nova UI
}
```

### Flush em segundo plano

Para não perder eventos quando o usuário minimiza o app, adicione o observer no widget raiz:

```dart
class _MyAppState extends State<MyApp> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.paused) {
      Pulse.instance.flush();
    }
  }
}
```

## Eventos recomendados

### Librinho

| Evento | Propriedades sugeridas |
|--------|------------------------|
| `book_scanned` | `isbn`, `source` (camera / manual) |
| `book_added` | `book_id`, `title`, `isbn` |
| `reading_started` | `book_id`, `title` |
| `reading_progress_updated` | `book_id`, `pages_read`, `total_pages`, `percent` |
| `reading_finished` | `book_id`, `pages_read`, `days_to_finish` |

```dart
// Exemplo completo — Librinho
Pulse.instance.track('reading_progress_updated', properties: {
  'book_id': 'book-abc-123',
  'pages_read': 120,
  'total_pages': 320,
  'percent': 37,
});
```

### Nitro (focus-hub)

| Evento | Propriedades sugeridas |
|--------|------------------------|
| `task_created` | `task_id`, `title`, `priority` |
| `task_completed` | `task_id`, `title`, `duration_minutes` |
| `focus_session_started` | `session_id`, `planned_minutes` |
| `focus_session_ended` | `session_id`, `actual_minutes`, `completed` |
| `daily_review_completed` | `tasks_done`, `tasks_pending`, `mood` |

```dart
// Exemplo completo — Nitro
Pulse.instance.track('focus_session_ended', properties: {
  'session_id': 'sess-xyz-789',
  'actual_minutes': 25,
  'completed': true,
});
```

### SocialMake mobile

| Evento | Propriedades sugeridas |
|--------|------------------------|
| `post_created` | `post_id`, `networks`, `has_media` |
| `post_scheduled` | `post_id`, `scheduled_at`, `networks` |
| `post_published` | `post_id`, `networks`, `engagement_score` |
| `account_connected` | `network` (linkedin / instagram / facebook) |
| `analytics_viewed` | `period`, `network` |

```dart
// Exemplo completo — SocialMake
Pulse.instance.track('post_scheduled', properties: {
  'post_id': 'post-001',
  'scheduled_at': '2026-05-20T09:00:00Z',
  'networks': ['linkedin', 'instagram'],
});
```

## Comportamento interno

### Fila de eventos (EventQueue)

- Acumula eventos em memória.
- Envia batch automaticamente ao atingir **10 eventos** ou a cada **30 segundos**.
- Persiste eventos pendentes em `SharedPreferences` (chave `pulse_pending_events`) para sobreviver a reinicializações.
- Ao inicializar, recupera e tenta enviar eventos da sessão anterior.
- Fire-and-forget: falhas de rede descartam o batch sem retry infinito.

### Cache de feature flags (FlagCache)

- Cache em memória com **TTL de 60 segundos**.
- `isEnabled()` retorna do cache se válido — resposta imediata.
- Se o cache expirou, faz fetch com **timeout de 200ms**.
- Em caso de falha ou timeout, retorna `defaultValue` (**fail-open**).

## API de referência

```dart
// Inicialização (main.dart)
await Pulse.init({
  required String projectKey,
  String? userId,
  bool debug = false,
  String apiUrl = 'https://pulse-api.institutoitinerante.com.br/api/v1',
});

// Identidade
Pulse.instance.identify(String userId);
Pulse.instance.reset();

// Eventos
Pulse.instance.track(String eventName, {Map<String, dynamic>? properties});

// Feature flags
Future<bool> Pulse.instance.isEnabled(String flagKey, {bool defaultValue = false});

// Controle da fila
Future<void> Pulse.instance.flush();
Future<void> Pulse.instance.persistBeforeClose();
```

## Projeto keys cadastrados no Pulse

| App | projectKey |
|-----|-----------|
| Librinho | `librinho` |
| Nitro | `nitro` |
| SocialMake | `socialmake` |
| Catraca Virtual | `catraca` |

## Licença

Propriedade do Instituto Itinerante. Uso interno.
