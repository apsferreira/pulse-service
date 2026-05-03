// Package telemetry configura OpenTelemetry tracing e Prometheus metrics para o pulse-service.
//
// Variáveis de ambiente:
//
//	OTEL_EXPORTER_OTLP_ENDPOINT  ex: obs-otel-collector:4317 (gRPC: host:port, sem scheme)
//	OTEL_SERVICE_NAME             ex: pulse-service
//
// Quando OTEL_EXPORTER_OTLP_ENDPOINT não está definido, o tracer no-op é usado
// para que o serviço suba normalmente em dev sem coletor rodando.
package telemetry

import (
	"context"
	"log"
	"os"
	"strings"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	otelfiber "github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Init configura o TracerProvider global do OpenTelemetry.
// Retorna uma função de shutdown que deve ser diferida pelo chamador.
func Init(defaultServiceName string) func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	if endpoint == "" {
		log.Println("[telemetry] OTEL_EXPORTER_OTLP_ENDPOINT não definido — tracing desabilitado")
		return func() {}
	}

	// Exportador gRPC requer host:port — remove http:// ou https:// se presente
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		log.Printf("[telemetry] Falha ao criar resource: %v — tracing desabilitado", err)
		return func() {}
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Printf("[telemetry] Falha no exportador OTLP: %v — tracing desabilitado", err)
		return func() {}
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	log.Printf("[telemetry] Tracing configurado — service=%s endpoint=%s", serviceName, endpoint)

	return func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Printf("[telemetry] Erro no shutdown: %v", err)
		}
	}
}

// RegisterFiber adiciona middleware OTel tracing e endpoint Prometheus /metrics ao app Fiber.
// Deve ser chamado antes de registrar as rotas da aplicação.
// Nota: o pulse-service já tem um MetricsHandler próprio em /metrics com contadores customizados.
// OTel tracing é adicionado aqui via middleware, sem sobrescrever o endpoint existente.
func RegisterFiber(app *fiber.App, serviceName string) {
	// Middleware de tracing distribuído OTel — instrumenta todas as rotas automaticamente
	app.Use(otelfiber.Middleware())

	// Prometheus /otel-metrics exposto separado do /metrics customizado do pulse-service
	// para não conflitar com o MetricsHandler existente (BKL-1053)
	prom := fiberprometheus.New(serviceName)
	prom.RegisterAt(app, "/otel-metrics")
	app.Use(prom.Middleware)
}
