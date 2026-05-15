import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import 'event_queue.dart';
import 'flag_cache.dart';
import 'pulse_config.dart';

/// SDK principal do Pulse — event tracking e feature flags para o IIT.
///
/// ## Ciclo de vida
///
/// ```dart
/// // 1. Inicializar antes de runApp()
/// await Pulse.init(projectKey: 'librinho');
///
/// // 2. Identificar após login
/// Pulse.instance.identify(userId);
///
/// // 3. Rastrear eventos
/// Pulse.instance.track('book_scanned', properties: {'isbn': '978...'});
///
/// // 4. Verificar feature flags
/// final enabled = await Pulse.instance.isEnabled('new-reader-ui');
///
/// // 5. Limpar identidade após logout
/// Pulse.instance.reset();
/// ```
class Pulse {
  static Pulse? _instance;

  /// Retorna a instância singleton do SDK.
  ///
  /// Lança [StateError] se [init] não foi chamado antes.
  static Pulse get instance {
    if (_instance == null) {
      throw StateError(
        'Pulse.init() não foi chamado. '
        'Chame await Pulse.init(...) antes de usar Pulse.instance.',
      );
    }
    return _instance!;
  }

  // ---------------------------------------------------------------------------
  // Estado interno
  // ---------------------------------------------------------------------------

  PulseConfig _config;
  final EventQueue _queue;
  final FlagCache _flagCache;

  Pulse._(this._config, this._queue, this._flagCache);

  // ---------------------------------------------------------------------------
  // API pública estática
  // ---------------------------------------------------------------------------

  /// Inicializa o SDK Pulse.
  ///
  /// Deve ser chamado uma única vez em `main()`, antes de `runApp()`.
  ///
  /// ```dart
  /// void main() async {
  ///   WidgetsFlutterBinding.ensureInitialized();
  ///   await Pulse.init(projectKey: 'librinho', debug: true);
  ///   runApp(const MyApp());
  /// }
  /// ```
  ///
  /// Parâmetros:
  /// - [projectKey]: chave do projeto no Pulse (obrigatório).
  /// - [userId]: ID do usuário, se já autenticado ao iniciar o app.
  /// - [debug]: habilita logs no console.
  /// - [apiUrl]: URL base da API (não alterar em produção).
  /// - [httpClient]: client HTTP injetável (útil em testes).
  static Future<void> init({
    required String projectKey,
    String? userId,
    bool debug = false,
    String apiUrl = 'https://pulse.institutoitinerante.com.br/api/v1',
    http.Client? httpClient,
  }) async {
    if (projectKey.isEmpty) {
      throw ArgumentError('projectKey não pode ser vazio.');
    }

    final config = PulseConfig(
      projectKey: projectKey,
      userId: userId,
      debug: debug,
      apiUrl: apiUrl,
    );

    final client = httpClient ?? http.Client();
    final queue = EventQueue(config, httpClient: client);
    final flagCache = FlagCache(config, httpClient: client);

    _instance = Pulse._(config, queue, flagCache);

    // Inicializa a fila (carrega eventos persistidos e agenda timer).
    await queue.initialize();

    // Pré-aquece o cache de flags em background — não bloqueia o init.
    flagCache.prefetch().ignore();

    if (debug) {
      // ignore: avoid_print
      print('[Pulse] Inicializado. project_key=$projectKey, userId=$userId');
    }
  }

  // ---------------------------------------------------------------------------
  // API pública de instância
  // ---------------------------------------------------------------------------

  /// Identifica o usuário autenticado.
  ///
  /// Deve ser chamado logo após o login bem-sucedido. Atualiza o userId em
  /// todos os eventos futuros e nos eventos ainda pendentes na fila.
  ///
  /// ```dart
  /// onLoginSuccess(String userId) {
  ///   Pulse.instance.identify(userId);
  /// }
  /// ```
  void identify(String userId) {
    if (userId.isEmpty) return;
    _config = _config.copyWith(userId: userId);
    _queue.updateUserId(userId);
    _log('Usuário identificado: $userId');
  }

  /// Limpa a identidade do usuário.
  ///
  /// Deve ser chamado após logout. Eventos futuros serão enviados sem userId.
  void reset() {
    _config = _config.copyWith(userId: null);
    _flagCache.invalidate();
    _log('Identidade resetada.');
  }

  /// Rastreia um evento.
  ///
  /// O evento é adicionado à fila em memória e enviado em batch.
  /// A operação retorna imediatamente (não bloqueia a UI).
  ///
  /// ```dart
  /// Pulse.instance.track('book_scanned', properties: {
  ///   'isbn': '978-85-359-0277-5',
  ///   'source': 'camera',
  /// });
  /// ```
  void track(String eventName, {Map<String, dynamic>? properties}) {
    if (eventName.isEmpty) return;
    _queue.enqueue(
      eventName,
      _config.userId,
      properties ?? const {},
    );
  }

  /// Verifica se uma feature flag está ativa para o usuário atual.
  ///
  /// Retorna do cache em memória se disponível (TTL 60s).
  /// Se o cache estiver expirado, faz fetch com timeout de 200ms.
  /// Em caso de falha ou timeout, retorna [defaultValue] (fail-open).
  ///
  /// ```dart
  /// if (await Pulse.instance.isEnabled('new-reader-ui')) {
  ///   // mostra nova UI
  /// }
  /// ```
  Future<bool> isEnabled(String flagKey, {bool defaultValue = false}) {
    return _flagCache.isEnabled(flagKey, defaultValue: defaultValue);
  }

  /// Força o envio imediato de todos os eventos pendentes na fila.
  ///
  /// Útil chamar quando o app vai para segundo plano (AppLifecycleState.paused).
  ///
  /// ```dart
  /// @override
  /// void didChangeAppLifecycleState(AppLifecycleState state) {
  ///   if (state == AppLifecycleState.paused) {
  ///     Pulse.instance.flush();
  ///   }
  /// }
  /// ```
  Future<void> flush() => _queue.flush();

  /// Persiste eventos pendentes antes do app fechar (chamar em didChangeAppLifecycleState).
  Future<void> persistBeforeClose() => _queue.persistBeforeClose();

  // ---------------------------------------------------------------------------
  // Helpers privados
  // ---------------------------------------------------------------------------

  void _log(String message) {
    if (_config.debug) {
      // ignore: avoid_print
      print('[Pulse] $message');
    }
  }

  // ---------------------------------------------------------------------------
  // API de teste — não usar em produção
  // ---------------------------------------------------------------------------

  /// Reseta o singleton. Usar APENAS em testes unitários.
  // ignore: invalid_use_of_visible_for_testing_member
  @visibleForTesting
  static void resetForTest() {
    _instance?._queue.dispose();
    _instance?._flagCache.dispose();
    _instance = null;
  }
}
