// BKL-1053: Métricas do Pulse em formato Prometheus text exposition.
// Implementação manual sem dependência externa para evitar adicionar prometheus/client_golang.
// Coleta: events/s, flag evaluations/s, cache hit rate, latência P99 estimada, error rate.
//
// Alertas documentados (configurar no Grafana):
//   - latência > 200ms
//   - error rate > 1%
//   - queue depth (redis list) > 1000

package metrics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Counters globais incrementados pelos handlers durante o ciclo de vida do processo.
var (
	eventsIngested     atomic.Int64 // total de eventos ingeridos (POST /api/v1/events)
	batchIngested      atomic.Int64 // total de eventos ingeridos via batch
	flagEvaluations    atomic.Int64 // total de avaliações de flags
	requestsTotal      atomic.Int64 // total de requisições HTTP
	requestErrors      atomic.Int64 // total de requisições com status >= 500
	cacheHits          atomic.Int64 // hits no cache Redis de flags
	cacheMisses        atomic.Int64 // misses no cache Redis de flags

	// Latência simplificada: somamos latências e contamos requisições para calcular média
	latencySumMs   atomic.Int64 // soma das latências em ms
	latencyCount   atomic.Int64 // número de requisições medidas

	// P99 estimado via bucket de latências (100ms bucket increments, 0-2000ms)
	mu             sync.Mutex
	latencyBuckets [21]int64 // 0-100, 100-200, ..., 1900-2000, >2000

	startTime = time.Now()
)

// RecordEvent registra uma ingestão de evento.
func RecordEvent() {
	eventsIngested.Add(1)
	requestsTotal.Add(1)
}

// RecordBatchEvent registra múltiplas ingestões em batch.
func RecordBatchEvent(n int) {
	batchIngested.Add(int64(n))
	requestsTotal.Add(1)
}

// RecordFlagEvaluation registra uma avaliação de flag.
func RecordFlagEvaluation() {
	flagEvaluations.Add(1)
	requestsTotal.Add(1)
}

// RecordRequest registra uma requisição HTTP genérica.
func RecordRequest() {
	requestsTotal.Add(1)
}

// RecordError registra um erro HTTP (status >= 500).
func RecordError() {
	requestErrors.Add(1)
}

// RecordCacheHit registra um hit no cache de flags.
func RecordCacheHit() {
	cacheHits.Add(1)
}

// RecordCacheMiss registra um miss no cache de flags.
func RecordCacheMiss() {
	cacheMisses.Add(1)
}

// RecordLatency registra a latência de uma requisição em ms.
func RecordLatency(ms int64) {
	latencySumMs.Add(ms)
	latencyCount.Add(1)

	// Popula bucket para estimativa de P99
	bucketIdx := int(ms / 100)
	if bucketIdx > 20 {
		bucketIdx = 20
	}
	mu.Lock()
	latencyBuckets[bucketIdx]++
	mu.Unlock()
}

// p99EstimateMs calcula uma estimativa do P99 via histograma de buckets.
func p99EstimateMs() float64 {
	mu.Lock()
	buckets := latencyBuckets
	mu.Unlock()

	var total int64
	for _, c := range buckets {
		total += c
	}
	if total == 0 {
		return 0
	}

	p99Threshold := int64(float64(total) * 0.99)
	var cumulative int64
	for i, c := range buckets {
		cumulative += c
		if cumulative >= p99Threshold {
			// Ponto médio do bucket
			return float64(i*100+50)
		}
	}
	return 2000
}

// uptimeSeconds retorna o uptime do processo em segundos.
func uptimeSeconds() float64 {
	return time.Since(startTime).Seconds()
}

// Collect retorna as métricas em formato Prometheus text exposition.
// Chamado pelo handler GET /metrics.
func Collect(_ context.Context) string {
	uptime := uptimeSeconds()

	totalReqs := requestsTotal.Load()
	totalErrors := requestErrors.Load()
	hits := cacheHits.Load()
	misses := cacheMisses.Load()
	totalCacheOps := hits + misses

	var cacheHitRate float64
	if totalCacheOps > 0 {
		cacheHitRate = float64(hits) / float64(totalCacheOps)
	}

	var errorRate float64
	if totalReqs > 0 {
		errorRate = float64(totalErrors) / float64(totalReqs)
	}

	var avgLatencyMs float64
	if lc := latencyCount.Load(); lc > 0 {
		avgLatencyMs = float64(latencySumMs.Load()) / float64(lc)
	}

	p99 := p99EstimateMs()

	return fmt.Sprintf(`# HELP pulse_uptime_seconds Uptime do processo em segundos
# TYPE pulse_uptime_seconds gauge
pulse_uptime_seconds %f

# HELP pulse_events_total Total de eventos ingeridos (POST /api/v1/events)
# TYPE pulse_events_total counter
pulse_events_total %d

# HELP pulse_batch_events_total Total de eventos ingeridos via batch
# TYPE pulse_batch_events_total counter
pulse_batch_events_total %d

# HELP pulse_flag_evaluations_total Total de avaliações de feature flags
# TYPE pulse_flag_evaluations_total counter
pulse_flag_evaluations_total %d

# HELP pulse_http_requests_total Total de requisições HTTP recebidas
# TYPE pulse_http_requests_total counter
pulse_http_requests_total %d

# HELP pulse_http_errors_total Total de respostas HTTP com status >= 500
# TYPE pulse_http_errors_total counter
pulse_http_errors_total %d

# HELP pulse_http_error_rate Taxa de erro HTTP (errors/requests)
# TYPE pulse_http_error_rate gauge
pulse_http_error_rate %f

# HELP pulse_flag_cache_hits_total Total de cache hits no Redis de flags
# TYPE pulse_flag_cache_hits_total counter
pulse_flag_cache_hits_total %d

# HELP pulse_flag_cache_misses_total Total de cache misses no Redis de flags
# TYPE pulse_flag_cache_misses_total counter
pulse_flag_cache_misses_total %d

# HELP pulse_flag_cache_hit_rate Taxa de cache hit de flags (0.0-1.0)
# TYPE pulse_flag_cache_hit_rate gauge
pulse_flag_cache_hit_rate %f

# HELP pulse_http_latency_avg_ms Latência média das requisições em ms
# TYPE pulse_http_latency_avg_ms gauge
pulse_http_latency_avg_ms %f

# HELP pulse_http_latency_p99_ms Latência P99 estimada das requisições em ms
# TYPE pulse_http_latency_p99_ms gauge
pulse_http_latency_p99_ms %f
`,
		uptime,
		eventsIngested.Load(),
		batchIngested.Load(),
		flagEvaluations.Load(),
		totalReqs,
		totalErrors,
		errorRate,
		hits,
		misses,
		cacheHitRate,
		avgLatencyMs,
		p99,
	)
}
