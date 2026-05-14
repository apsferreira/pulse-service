package handler

// BKL-1053: Endpoint GET /metrics — Prometheus text exposition format.
// Expõe: events/s, flag evaluations/s, cache hit rate, latência P99, error rate.
// Configurar no Prometheus como scrape target (interval: 15s).
// Alertas recomendados no Grafana:
//   - pulse_http_latency_p99_ms > 200
//   - pulse_http_error_rate > 0.01
//   - Configurar via Grafana UI apontando para Prometheus como datasource.

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/pkg/metrics"
)

// MetricsHandler serve métricas para scraping pelo Prometheus.
type MetricsHandler struct {
	serviceToken string
}

func NewMetricsHandler(serviceToken string) *MetricsHandler {
	return &MetricsHandler{serviceToken: serviceToken}
}

// RegisterRoutes monta a rota /metrics.
func (h *MetricsHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/metrics", middleware.ServiceTokenMiddleware(h.serviceToken), h.Metrics)

	// Middleware de instrumentação: injeta latência e contadores em todas as rotas
	app.Use(h.Instrument)
}

// Metrics godoc
// GET /metrics — retorna métricas no formato Prometheus text exposition.
func (h *MetricsHandler) Metrics(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	return c.SendString(metrics.Collect(c.Context()))
}

// Instrument é um middleware que registra latência e erros para cada requisição.
func (h *MetricsHandler) Instrument(c *fiber.Ctx) error {
	start := time.Now()
	metrics.RecordRequest()

	err := c.Next()

	latencyMs := time.Since(start).Milliseconds()
	metrics.RecordLatency(latencyMs)

	if c.Response().StatusCode() >= 500 {
		metrics.RecordError()
	}

	return err
}
