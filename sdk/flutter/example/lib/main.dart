import 'package:flutter/material.dart';
import 'package:pulse_sdk/pulse.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Pulse.init(
    projectKey: 'librinho',
    debug: true,
    // userId: 'user-123', // opcional: se já estiver logado ao abrir o app
  );

  runApp(const PulseExampleApp());
}

class PulseExampleApp extends StatelessWidget {
  const PulseExampleApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pulse SDK — Demo',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF6366F1)),
        useMaterial3: true,
      ),
      home: const _DemoPage(),
    );
  }
}

class _DemoPage extends StatefulWidget {
  const _DemoPage();

  @override
  State<_DemoPage> createState() => _DemoPageState();
}

class _DemoPageState extends State<_DemoPage> with WidgetsBindingObserver {
  String _flagStatus = 'Não verificado';
  String _lastEvent = 'Nenhum';
  bool _identified = false;

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
    // Persiste e envia eventos quando o app vai para segundo plano.
    if (state == AppLifecycleState.paused) {
      Pulse.instance.flush();
    }
  }

  void _simulateLogin() {
    Pulse.instance.identify('demo-user-001');
    Pulse.instance.track('login_completed', properties: {
      'method': 'email',
      'app': 'librinho',
    });
    setState(() {
      _identified = true;
      _lastEvent = 'login_completed';
    });
  }

  void _simulateLogout() {
    Pulse.instance.track('logout_requested');
    Pulse.instance.reset();
    setState(() {
      _identified = false;
      _lastEvent = 'logout_requested';
    });
  }

  void _trackBookScanned() {
    Pulse.instance.track('book_scanned', properties: {
      'isbn': '978-85-359-0277-5',
      'source': 'camera',
    });
    setState(() => _lastEvent = 'book_scanned');
  }

  void _trackReadingStarted() {
    Pulse.instance.track('reading_started', properties: {
      'book_id': 'book-abc-123',
      'title': 'O Senhor dos Anéis',
    });
    setState(() => _lastEvent = 'reading_started');
  }

  void _trackReadingFinished() {
    Pulse.instance.track('reading_finished', properties: {
      'book_id': 'book-abc-123',
      'pages_read': 1178,
      'days_to_finish': 14,
    });
    setState(() => _lastEvent = 'reading_finished');
  }

  Future<void> _checkFlag() async {
    final enabled = await Pulse.instance.isEnabled(
      'new-reader-ui',
      defaultValue: false,
    );
    setState(() => _flagStatus = enabled ? 'Ativa' : 'Inativa');
  }

  Future<void> _forceFlush() async {
    await Pulse.instance.flush();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Eventos enviados (flush manual).')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Pulse SDK — Demo'),
        backgroundColor: theme.colorScheme.inversePrimary,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _SectionCard(
            title: 'Identidade',
            children: [
              Text('Status: ${_identified ? "Identificado" : "Anônimo"}'),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: FilledButton(
                      onPressed: _identified ? null : _simulateLogin,
                      child: const Text('Login'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: OutlinedButton(
                      onPressed: _identified ? _simulateLogout : null,
                      child: const Text('Logout'),
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),
          _SectionCard(
            title: 'Eventos — Librinho',
            children: [
              Text('Último evento: $_lastEvent'),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  ElevatedButton(
                    onPressed: _trackBookScanned,
                    child: const Text('book_scanned'),
                  ),
                  ElevatedButton(
                    onPressed: _trackReadingStarted,
                    child: const Text('reading_started'),
                  ),
                  ElevatedButton(
                    onPressed: _trackReadingFinished,
                    child: const Text('reading_finished'),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),
          _SectionCard(
            title: 'Feature Flags',
            children: [
              Text('new-reader-ui: $_flagStatus'),
              const SizedBox(height: 8),
              FilledButton(
                onPressed: _checkFlag,
                child: const Text('Verificar flag'),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _SectionCard(
            title: 'Controle da Fila',
            children: [
              OutlinedButton.icon(
                onPressed: _forceFlush,
                icon: const Icon(Icons.send),
                label: const Text('Flush manual'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _SectionCard extends StatelessWidget {
  final String title;
  final List<Widget> children;

  const _SectionCard({required this.title, required this.children});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 12),
            ...children,
          ],
        ),
      ),
    );
  }
}
