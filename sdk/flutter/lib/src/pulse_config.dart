/// Configuração imutável do SDK Pulse.
///
/// Criada uma única vez em [Pulse.init] e propagada para todos os componentes
/// internos. Não deve ser instanciada diretamente pelo consumidor.
class PulseConfig {
  /// Chave de projeto — identifica o tenant no Pulse.
  final String projectKey;

  /// ID do usuário autenticado. Pode ser nulo até [Pulse.identify] ser chamado.
  final String? userId;

  /// Habilita logs de diagnóstico no console. Use apenas em desenvolvimento.
  final bool debug;

  /// URL base da API Pulse (sem barra final).
  final String apiUrl;

  const PulseConfig({
    required this.projectKey,
    this.userId,
    this.debug = false,
    this.apiUrl = 'https://pulse.institutoitinerante.com.br/api/v1',
  });

  PulseConfig copyWith({String? userId}) {
    return PulseConfig(
      projectKey: projectKey,
      userId: userId ?? this.userId,
      debug: debug,
      apiUrl: apiUrl,
    );
  }

  /// URL base sem o prefixo /api/v1, usado pelo endpoint público de flags
  /// que está montado em /flags (fora do grupo /api/v1).
  String get baseUrl {
    // Remove o sufixo /api/v1 se presente
    if (apiUrl.endsWith('/api/v1')) {
      return apiUrl.substring(0, apiUrl.length - 7);
    }
    return apiUrl;
  }
}
