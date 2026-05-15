import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import 'pulse_config.dart';

/// Um evento pendente de envio ao Pulse.
class _PendingEvent {
  final String eventName;
  final String? userId;
  final Map<String, dynamic> properties;
  final DateTime timestamp;

  const _PendingEvent({
    required this.eventName,
    this.userId,
    required this.properties,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() => {
        'event_name': eventName,
        if (userId != null) 'user_id': userId,
        'properties': properties,
        'timestamp': timestamp.toUtc().toIso8601String(),
      };

  factory _PendingEvent.fromJson(Map<String, dynamic> json) {
    return _PendingEvent(
      eventName: json['event_name'] as String,
      userId: json['user_id'] as String?,
      properties: (json['properties'] as Map<String, dynamic>? ?? {}),
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }
}

/// Fila de eventos em memória com batch flush automático.
///
/// Comportamento:
/// - Acumula eventos em memória.
/// - Envia batch quando atingir [_batchSize] eventos OU quando o timer de
///   [_flushInterval] disparar.
/// - Persiste eventos pendentes em SharedPreferences para sobreviver a
///   reinicializações (chave: [_prefsKey]).
/// - Fire-and-forget: falhas de rede descartam o batch sem retry.
class EventQueue {
  static const String _prefsKey = 'pulse_pending_events';
  static const int _batchSize = 10;
  static const Duration _flushInterval = Duration(seconds: 30);

  final PulseConfig _config;
  final http.Client _httpClient;

  final List<_PendingEvent> _queue = [];
  Timer? _flushTimer;
  bool _initialized = false;

  EventQueue(this._config, {http.Client? httpClient})
      : _httpClient = httpClient ?? http.Client();

  /// Inicializa a fila, carregando eventos não enviados da sessão anterior
  /// e agendando o timer de flush periódico.
  Future<void> initialize() async {
    if (_initialized) return;
    _initialized = true;

    await _loadPersistedEvents();
    _scheduleTimer();

    // Tenta enviar imediatamente se havia eventos da sessão anterior.
    if (_queue.isNotEmpty) {
      _log('Recuperados ${_queue.length} eventos persistidos — tentando enviar.');
      await flush();
    }
  }

  /// Adiciona um evento à fila. Dispara flush imediato se atingir [_batchSize].
  void enqueue(String eventName, String? userId, Map<String, dynamic> properties) {
    _queue.add(_PendingEvent(
      eventName: eventName,
      userId: userId,
      properties: properties,
      timestamp: DateTime.now(),
    ));
    _log('Evento enfileirado: $eventName (fila: ${_queue.length})');

    if (_queue.length >= _batchSize) {
      _log('Batch size atingido — flush imediato.');
      flush();
    }
  }

  /// Envia todos os eventos pendentes para a API Pulse.
  ///
  /// Usa o endpoint de batch [POST /api/v1/events/batch] quando há múltiplos
  /// eventos; cai para [POST /api/v1/events/track] para evento único.
  Future<void> flush() async {
    if (_queue.isEmpty) return;

    final batch = List<_PendingEvent>.from(_queue);
    _queue.clear();

    // Remove persistência imediatamente — se falhar, os eventos são descartados
    // (fire-and-forget com persistência de sessão, não retry infinito).
    await _clearPersistedEvents();

    try {
      await _sendBatch(batch);
      _log('Batch de ${batch.length} eventos enviado com sucesso.');
    } catch (e) {
      _log('Falha ao enviar batch (descartando ${batch.length} eventos): $e');
      // Fire-and-forget: não recoloca na fila para evitar acúmulo infinito.
    }
  }

  /// Atualiza o userId em eventos ainda na fila (para identify após enqueue).
  void updateUserId(String userId) {
    for (var i = 0; i < _queue.length; i++) {
      if (_queue[i].userId == null) {
        _queue[i] = _PendingEvent(
          eventName: _queue[i].eventName,
          userId: userId,
          properties: _queue[i].properties,
          timestamp: _queue[i].timestamp,
        );
      }
    }
  }

  void dispose() {
    _flushTimer?.cancel();
    _httpClient.close();
  }

  // ---------------------------------------------------------------------------
  // Implementação interna
  // ---------------------------------------------------------------------------

  Future<void> _sendBatch(List<_PendingEvent> events) async {
    final url = Uri.parse('${_config.apiUrl}/events/batch');

    // Monta payload no formato que o BatchIngestEvents espera.
    final payload = {
      'project_key': _config.projectKey,
      'events': events.map((e) => e.toJson()).toList(),
    };

    final response = await _httpClient
        .post(
          url,
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode(payload),
        )
        .timeout(const Duration(seconds: 10));

    if (response.statusCode >= 400) {
      throw Exception(
          'API retornou ${response.statusCode}: ${response.body}');
    }
  }

  void _scheduleTimer() {
    _flushTimer?.cancel();
    _flushTimer = Timer.periodic(_flushInterval, (_) {
      _log('Timer de flush disparado.');
      flush();
    });
  }

  Future<void> _loadPersistedEvents() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final raw = prefs.getString(_prefsKey);
      if (raw == null || raw.isEmpty) return;

      final List<dynamic> decoded = jsonDecode(raw) as List<dynamic>;
      for (final item in decoded) {
        try {
          _queue.add(_PendingEvent.fromJson(item as Map<String, dynamic>));
        } catch (_) {
          // Ignora eventos malformados — não bloqueia inicialização.
        }
      }
    } catch (e) {
      _log('Erro ao carregar eventos persistidos: $e');
    }
  }

  Future<void> _persistCurrentQueue() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final encoded = jsonEncode(_queue.map((e) => e.toJson()).toList());
      await prefs.setString(_prefsKey, encoded);
    } catch (e) {
      _log('Erro ao persistir fila: $e');
    }
  }

  Future<void> _clearPersistedEvents() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(_prefsKey);
    } catch (_) {}
  }

  /// Persiste a fila atual — deve ser chamado antes do app ser suspenso/fechado.
  Future<void> persistBeforeClose() async {
    if (_queue.isEmpty) return;
    await _persistCurrentQueue();
  }

  void _log(String message) {
    if (_config.debug) {
      // ignore: avoid_print
      print('[Pulse/Queue] $message');
    }
  }
}
