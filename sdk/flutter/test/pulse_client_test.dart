import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:pulse_sdk/pulse.dart';

void main() {
  // Reseta o singleton entre testes.
  tearDown(() {
    // Acesso via reflexão não é necessário em Dart — usamos uma API de reset
    // interna exposta apenas para testes (ver nota no pulse_client.dart).
    // Para este conjunto de testes, inicializamos o SDK do zero em cada caso.
  });

  group('Pulse.init', () {
    test('lança ArgumentError se projectKey for vazio', () async {
      expect(
        () => Pulse.init(projectKey: ''),
        throwsArgumentError,
      );
    });

    test('lança StateError ao acessar instance antes do init', () {
      // Reset forçado via hack de test — ver comentário em pulse_client.dart
      // ignore: invalid_use_of_visible_for_testing_member
      Pulse.resetForTest();
      expect(() => Pulse.instance, throwsStateError);
    });
  });

  group('Pulse.track', () {
    late MockClient mockHttpClient;
    late List<http.Request> capturedRequests;

    setUp(() async {
      capturedRequests = [];
      mockHttpClient = MockClient((request) async {
        capturedRequests.add(request);
        return http.Response('{"ingested":1}', 201);
      });

      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: mockHttpClient,
      );
    });

    test('track com eventName vazio não enfileira nada', () {
      Pulse.instance.track('');
      // Nenhuma requisição deve ser feita imediatamente.
      expect(capturedRequests, isEmpty);
    });

    test('track enfileira evento e flush envia para API', () async {
      Pulse.instance.track('test_event', properties: {'key': 'value'});
      await Pulse.instance.flush();

      expect(capturedRequests, isNotEmpty);
      final body = jsonDecode(capturedRequests.last.body) as Map<String, dynamic>;
      expect(body['project_key'], equals('test-project'));

      final events = body['events'] as List<dynamic>;
      expect(events, hasLength(1));
      expect(events.first['event_name'], equals('test_event'));
      expect(events.first['properties']['key'], equals('value'));
    });

    test('track de 10 eventos dispara flush automático', () async {
      for (var i = 0; i < 10; i++) {
        Pulse.instance.track('event_$i');
      }
      // Dá tempo ao flush assíncrono.
      await Future<void>.delayed(const Duration(milliseconds: 100));

      expect(capturedRequests, isNotEmpty);
    });
  });

  group('Pulse.identify + reset', () {
    setUp(() async {
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((_) async => http.Response('{}', 201)),
      );
    });

    test('identify define userId nos eventos subsequentes', () async {
      final capturedRequests = <http.Request>[];
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((req) async {
          capturedRequests.add(req);
          return http.Response('{}', 201);
        }),
      );

      Pulse.instance.identify('user-42');
      Pulse.instance.track('page_viewed');
      await Pulse.instance.flush();

      final body = jsonDecode(capturedRequests.last.body) as Map<String, dynamic>;
      final events = body['events'] as List<dynamic>;
      expect(events.first['user_id'], equals('user-42'));
    });

    test('reset limpa userId — eventos futuros não têm user_id', () async {
      final capturedRequests = <http.Request>[];
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((req) async {
          capturedRequests.add(req);
          return http.Response('{}', 201);
        }),
      );

      Pulse.instance.identify('user-99');
      Pulse.instance.reset();
      Pulse.instance.track('anonymous_event');
      await Pulse.instance.flush();

      final body = jsonDecode(capturedRequests.last.body) as Map<String, dynamic>;
      final events = body['events'] as List<dynamic>;
      expect(events.first.containsKey('user_id'), isFalse);
    });
  });

  group('Pulse.isEnabled', () {
    test('retorna defaultValue quando API falha', () async {
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((_) async => http.Response('error', 500)),
      );

      final result = await Pulse.instance.isEnabled('some-flag', defaultValue: true);
      expect(result, isTrue);
    });

    test('retorna valor da flag quando API responde com sucesso', () async {
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((req) async {
          if (req.url.path.endsWith('/flags')) {
            return http.Response(
              jsonEncode({'flags': {'my-feature': true, 'other-flag': false}}),
              200,
            );
          }
          return http.Response('{}', 201);
        }),
      );

      // Aguarda o prefetch terminar.
      await Future<void>.delayed(const Duration(milliseconds: 50));

      final enabled = await Pulse.instance.isEnabled('my-feature');
      final disabled = await Pulse.instance.isEnabled('other-flag');

      expect(enabled, isTrue);
      expect(disabled, isFalse);
    });

    test('retorna defaultValue para flag inexistente', () async {
      Pulse.resetForTest();
      await Pulse.init(
        projectKey: 'test-project',
        debug: false,
        httpClient: MockClient((req) async {
          if (req.url.path.endsWith('/flags')) {
            return http.Response(jsonEncode({'flags': {}}), 200);
          }
          return http.Response('{}', 201);
        }),
      );

      await Future<void>.delayed(const Duration(milliseconds: 50));

      final result = await Pulse.instance.isEnabled('nonexistent', defaultValue: false);
      expect(result, isFalse);
    });
  });
}
