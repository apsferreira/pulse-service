package handler

// BKL-1063: LLM Observability expandida para o Pulse.
// Métricas por modelo (GPT-4o, Claude Sonnet, Gemini), por feature, latência P50/P99,
// error rate por provider. Diferencial único vs todos os concorrentes.
//
// Endpoints:
//   GET /api/v1/llm/overview?product=...&from=...&to=...  — visão geral com P50/P99
//   GET /api/v1/llm/by-model?from=...&to=...              — agrupado por model+provider
//   GET /api/v1/llm/by-feature?product=...                — agrupado por feature
//   GET /api/v1/llm/errors?from=...&to=...                — error rate por provider

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LLMObservabilityHandler expõe métricas detalhadas de uso de LLM.
type LLMObservabilityHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewLLMObservabilityHandler(db *pgxpool.Pool, serviceToken string) *LLMObservabilityHandler {
	return &LLMObservabilityHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de LLM observability.
func (h *LLMObservabilityHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/llm", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/overview", h.Overview)
	api.Get("/by-model", h.ByModel)
	api.Get("/by-feature", h.ByFeature)
	api.Get("/errors", h.Errors)
}

// LLMOverview agrega métricas gerais de LLM.
type LLMOverview struct {
	Period      string       `json:"period"`
	TotalCalls  int          `json:"total_calls"`
	TotalErrors int          `json:"total_errors"`
	ErrorRate   float64      `json:"error_rate_pct"`
	TotalUSD    float64      `json:"total_cost_usd"`
	TotalBRL    float64      `json:"total_cost_brl"`
	AvgLatencyMs  *float64   `json:"avg_latency_ms,omitempty"`
	P50LatencyMs  *int       `json:"p50_latency_ms,omitempty"`
	P99LatencyMs  *int       `json:"p99_latency_ms,omitempty"`
	ByProvider  []ProviderStats `json:"by_provider"`
}

// ProviderStats agrega por provider.
type ProviderStats struct {
	Provider   string  `json:"provider"`
	Calls      int     `json:"calls"`
	Errors     int     `json:"errors"`
	ErrorRate  float64 `json:"error_rate_pct"`
	TotalUSD   float64 `json:"total_cost_usd"`
	AvgTokens  float64 `json:"avg_tokens_per_call"`
}

// ModelStats agrega por model+provider.
type ModelStats struct {
	Provider     string   `json:"provider"`
	Model        string   `json:"model"`
	Calls        int      `json:"calls"`
	Errors       int      `json:"errors"`
	TotalUSD     float64  `json:"total_cost_usd"`
	AvgLatencyMs *float64 `json:"avg_latency_ms,omitempty"`
	P50LatencyMs *int     `json:"p50_latency_ms,omitempty"`
	P99LatencyMs *int     `json:"p99_latency_ms,omitempty"`
}

// FeatureStats agrega por feature que usou LLM.
type FeatureStats struct {
	Feature   string  `json:"feature"`
	Calls     int     `json:"calls"`
	TotalUSD  float64 `json:"total_cost_usd"`
	AvgTokens float64 `json:"avg_tokens_per_call"`
}

// ErrorStats agrega erros por tipo e provider.
type ErrorStats struct {
	Provider  string `json:"provider"`
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
	LastSeen  string `json:"last_seen"`
}

// Overview GET /api/v1/llm/overview?product=...&from=...&to=...
func (h *LLMObservabilityHandler) Overview(c *fiber.Ctx) error {
	product := c.Query("product", "")
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Filtro de produto via placeholder paramétrico — nunca interpolação de input.
	// whereProduct é string literal fixa; o valor do usuário só entra como argumento $3.
	const productFilter = " AND product = $3"
	args := []interface{}{start, end}
	filterClause := ""
	if product != "" {
		filterClause = productFilter
		args = append(args, product)
	}

	// Métricas gerais
	var totalCalls, totalErrors int
	var totalUSD float64
	err = h.db.QueryRow(c.Context(),
		"SELECT COUNT(*), COUNT(*) FILTER (WHERE is_error), COALESCE(SUM(cost_usd), 0)"+
			" FROM api_costs WHERE created_at BETWEEN $1 AND $2"+filterClause,
		args...).Scan(&totalCalls, &totalErrors, &totalUSD)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha na query"})
	}

	errorRate := 0.0
	if totalCalls > 0 {
		errorRate = float64(totalErrors) / float64(totalCalls) * 100
	}

	const usdToBRL = 5.0
	overview := LLMOverview{
		Period:      fmt.Sprintf("%s a %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
		TotalCalls:  totalCalls,
		TotalErrors: totalErrors,
		ErrorRate:   round2(errorRate),
		TotalUSD:    round2(totalUSD),
		TotalBRL:    round2(totalUSD * usdToBRL),
	}

	// Latência P50/P99 (apenas se a coluna tiver dados)
	var avgLat *float64
	var p50, p99 *int
	_ = h.db.QueryRow(c.Context(),
		"SELECT AVG(latency_ms),"+
			" PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY latency_ms)::INT,"+
			" PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::INT"+
			" FROM api_costs WHERE created_at BETWEEN $1 AND $2 AND latency_ms IS NOT NULL"+filterClause,
		args...).Scan(&avgLat, &p50, &p99)
	overview.AvgLatencyMs = avgLat
	overview.P50LatencyMs = p50
	overview.P99LatencyMs = p99

	// Por provider
	pRows, _ := h.db.Query(c.Context(),
		"SELECT provider,"+
			" COUNT(*) AS calls,"+
			" COUNT(*) FILTER (WHERE is_error) AS errors,"+
			" COALESCE(SUM(cost_usd), 0),"+
			" COALESCE(AVG(tokens_in + tokens_out), 0)"+
			" FROM api_costs WHERE created_at BETWEEN $1 AND $2"+filterClause+
			" GROUP BY provider ORDER BY calls DESC",
		args...)
	if pRows != nil {
		defer pRows.Close()
		for pRows.Next() {
			var ps ProviderStats
			var avgTok float64
			if err := pRows.Scan(&ps.Provider, &ps.Calls, &ps.Errors, &ps.TotalUSD, &avgTok); err == nil {
				if ps.Calls > 0 {
					ps.ErrorRate = round2(float64(ps.Errors) / float64(ps.Calls) * 100)
				}
				ps.AvgTokens = round2(avgTok)
				ps.TotalUSD = round2(ps.TotalUSD)
				overview.ByProvider = append(overview.ByProvider, ps)
			}
		}
	}
	if overview.ByProvider == nil {
		overview.ByProvider = []ProviderStats{}
	}

	return c.JSON(overview)
}

// ByModel GET /api/v1/llm/by-model?from=...&to=...
func (h *LLMObservabilityHandler) ByModel(c *fiber.Ctx) error {
	start, end, err := parseDateRange(c.Query("from"), c.Query("to"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT
			provider,
			COALESCE(model, 'unknown') AS model,
			COUNT(*) AS calls,
			COUNT(*) FILTER (WHERE is_error) AS errors,
			COALESCE(SUM(cost_usd), 0),
			AVG(latency_ms),
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY latency_ms)::INT,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms)::INT
		FROM api_costs
		WHERE created_at BETWEEN $1 AND $2
		  AND model IS NOT NULL AND model != ''
		GROUP BY provider, model
		ORDER BY calls DESC
		LIMIT 50
	`, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha na query"})
	}
	defer rows.Close()

	var stats []ModelStats
	for rows.Next() {
		var s ModelStats
		if err := rows.Scan(
			&s.Provider, &s.Model, &s.Calls, &s.Errors, &s.TotalUSD,
			&s.AvgLatencyMs, &s.P50LatencyMs, &s.P99LatencyMs,
		); err == nil {
			s.TotalUSD = round2(s.TotalUSD)
			stats = append(stats, s)
		}
	}
	if stats == nil {
		stats = []ModelStats{}
	}

	return c.JSON(fiber.Map{"models": stats, "total": len(stats)})
}

// ByFeature GET /api/v1/llm/by-feature?product=...&from=...&to=...
func (h *LLMObservabilityHandler) ByFeature(c *fiber.Ctx) error {
	product := c.Query("product", "")
	start, end, _ := parseDateRange(c.Query("from"), c.Query("to"))

	const productFilter = " AND product = $3"
	args := []interface{}{start, end}
	filterClause := ""
	if product != "" {
		filterClause = productFilter
		args = append(args, product)
	}

	rows, err := h.db.Query(c.Context(),
		"SELECT feature, COUNT(*) AS calls,"+
			" COALESCE(SUM(cost_usd), 0),"+
			" COALESCE(AVG(tokens_in + tokens_out), 0)"+
			" FROM api_costs"+
			" WHERE created_at BETWEEN $1 AND $2 AND feature IS NOT NULL"+filterClause+
			" GROUP BY feature ORDER BY calls DESC LIMIT 30",
		args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha na query"})
	}
	defer rows.Close()

	var stats []FeatureStats
	for rows.Next() {
		var s FeatureStats
		if err := rows.Scan(&s.Feature, &s.Calls, &s.TotalUSD, &s.AvgTokens); err == nil {
			s.TotalUSD = round2(s.TotalUSD)
			s.AvgTokens = round2(s.AvgTokens)
			stats = append(stats, s)
		}
	}
	if stats == nil {
		stats = []FeatureStats{}
	}

	return c.JSON(fiber.Map{"features": stats, "total": len(stats)})
}

// Errors GET /api/v1/llm/errors?from=...&to=...
func (h *LLMObservabilityHandler) Errors(c *fiber.Ctx) error {
	start, end, _ := parseDateRange(c.Query("from"), c.Query("to"))

	rows, err := h.db.Query(c.Context(), `
		SELECT
			provider,
			COALESCE(error_type, 'unknown') AS error_type,
			COUNT(*) AS count,
			MAX(created_at)::TEXT AS last_seen
		FROM api_costs
		WHERE is_error = true AND created_at BETWEEN $1 AND $2
		GROUP BY provider, error_type
		ORDER BY count DESC
		LIMIT 30
	`, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha na query"})
	}
	defer rows.Close()

	var stats []ErrorStats
	for rows.Next() {
		var s ErrorStats
		if err := rows.Scan(&s.Provider, &s.ErrorType, &s.Count, &s.LastSeen); err == nil {
			stats = append(stats, s)
		}
	}
	if stats == nil {
		stats = []ErrorStats{}
	}

	totalErrors := 0
	for _, s := range stats {
		totalErrors += s.Count
	}

	return c.JSON(fiber.Map{
		"errors":       stats,
		"total_errors": totalErrors,
		"period":       fmt.Sprintf("%s a %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
	})
}

// parseDateRange e round2 já definidas em funnel_handler.go (mesmo pacote handler).
