/// SDK Flutter para Pulse — feature flags e event tracking do IIT.
///
/// ## Uso básico
///
/// ```dart
/// import 'package:pulse_sdk/pulse.dart';
///
/// void main() async {
///   WidgetsFlutterBinding.ensureInitialized();
///   await Pulse.init(projectKey: 'librinho');
///   runApp(const MyApp());
/// }
/// ```
library pulse_sdk;

export 'src/pulse_client.dart' show Pulse;
export 'src/pulse_config.dart' show PulseConfig;
