import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'pulse_config.dart';

/// Entrada de cache para o mapa de feature flags.
class _CacheEntry {
  final Map<String, bool> flags;
  final DateTime fetchedAt;

  const _CacheEntry({required this.flags, required this.fetchedAt});

  bool isExpired(Duration ttl) =>
      DateTime.now().difference(fetchedAt) > ttl;
}

/// Cache em memória de feature flags com TTL de 60s.
///
/// Comportamento:
/// - [isEnabled] retorna do cache se disponível e não expirado.
/// - Caso o cache esteja vazio ou expirado, faz fetch com timeout de 200ms.
/// - Se o fetch falhar ou expirar o timeout, retorna [defaultValue] (fail-open).
/// - Endpoint usado: GET /flags?project_id=<projectKey> (público, sem autenticação).
class FlagCache {
  static const Duration _ttl = Duration(seconds: 60);
  static const Duration _fetchTimeout = Duration(milliseconds: 200);

  final PulseConfig _config;
  final http.Client _httpClient;

  _CacheEntry? _cache;
  // Evita múltiplos fetches simultâneos.
  Future<Map<String, bool>?>? _pendingFetch;

  FlagCache(this._config, {http.Client? httpClient})
      : _httpClient = httpClient ?? http.Client();

  /// Verifica se a flag [flagKey] está habilitada.
  ///
  /// Ordem de resolução:
  /// 1. Cache válido (não expirado) — resposta imediata.
  /// 2. Fetch da API com timeout de 200ms.
  /// 3. Se timeout/erro → [defaultValue].
  Future<bool> isEnabled(String flagKey, {bool defaultValue = false}) async {
    final cached = _cache;
    if (cached != null && !cached.isExpired(_ttl)) {
      _log('Cache hit para "$flagKey": ${cached.flags[flagKey]}');
      return cached.flags[flagKey] ?? defaultValue;
    }

    final flags = await _fetchWithTimeout();
    if (flags == null) {
      _log('Fetch falhou ou timeout — retornando defaultValue=$defaultValue para "$flagKey".');
      return defaultValue;
    }

    return flags[flagKey] ?? defaultValue;
  }

  /// Pré-aquece o cache buscando todas as flags. Útil chamar durante o init.
  Future<void> prefetch() async {
    await _fetchWithTimeout();
  }

  /// Invalida o cache forçando novo fetch na próxima chamada.
  void invalidate() {
    _cache = null;
    _log('Cache invalidado.');
  }

  void dispose() {
    _httpClient.close();
  }

  // ---------------------------------------------------------------------------
  // Implementação interna
  // ---------------------------------------------------------------------------

  Future<Map<String, bool>?> _fetchWithTimeout() async {
    // Reutiliza fetch em andamento para evitar N requisições paralelas.
    _pendingFetch ??= _doFetch().whenComplete(() => _pendingFetch = null);

    try {
      return await _pendingFetch!.timeout(_fetchTimeout);
    } on TimeoutException {
      _log('Timeout de 200ms ao buscar flags.');
      return null;
    }
  }

  Future<Map<String, bool>?> _doFetch() async {
    try {
      // O endpoint de flags é público e montado em /flags (fora de /api/v1).
      // Query param: project_id — aceita tanto UUID quanto project_key.
      final url = Uri.parse('${_config.baseUrl}/flags')
          .replace(queryParameters: {'project_id': _config.projectKey});

      _log('Buscando flags: $url');

      final response = await _httpClient
          .get(url, headers: {'Accept': 'application/json'})
          .timeout(const Duration(seconds: 5));

      if (response.statusCode != 200) {
        _log('API de flags retornou ${response.statusCode}.');
        return null;
      }

      final body = jsonDecode(response.body) as Map<String, dynamic>;
      // Resposta esperada: {"flags": {"flag-key": true, ...}}
      final rawFlags = body['flags'] as Map<String, dynamic>? ?? {};
      final flags = rawFlags.map(
        (k, v) => MapEntry(k, v == true),
      );

      _cache = _CacheEntry(flags: flags, fetchedAt: DateTime.now());
      _log('Flags carregadas: ${flags.length} entradas.');
      return flags;
    } catch (e) {
      _log('Erro ao buscar flags: $e');
      return null;
    }
  }

  void _log(String message) {
    if (_config.debug) {
      // ignore: avoid_print
      print('[Pulse/Flags] $message');
    }
  }
}
