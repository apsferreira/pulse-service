package handler

// BKL-769: North Star Metrics por produto IIT — dados reais da tabela events.
//
// STD-013: Protegido com ServiceTokenMiddleware — métricas de negócio dos 6
// produtos IIT (DAU, churn, conversão) não devem ser públicas.
//
// Endpoints:
//   GET /api/v1/north-star   — retorna métricas de conversão dos 6 produtos IIT

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/institutoitinerante/pulse-service/internal/middleware"
)

// dbPool é a dependência mínima de banco para o NorthStarHandler.
// Interface permite injeção de mock nos testes sem expor pgx internamente.
type dbPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// NorthStarHandler retorna métricas de conversão reais por produto IIT.
type NorthStarHandler struct {
	db           dbPool
	serviceToken string
}

func NewNorthStarHandler(db *pgxpool.Pool, serviceToken string) *NorthStarHandler {
	return &NorthStarHandler{db: db, serviceToken: serviceToken}
}

func (h *NorthStarHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/north-star", h.Get)
}

// NorthStarMetric representa a métrica principal de um produto com série histórica (8 semanas).
type NorthStarMetric struct {
	Product       string    `json:"product"`
	MetricName    string    `json:"metric_name"`
	Unit          string    `json:"unit"`
	CurrentValue  int64     `json:"current_value"`  // semana corrente (ou mês, conforme produto)
	PreviousValue int64     `json:"previous_value"` // semana/mês anterior
	Sparkline     []int64   `json:"sparkline"`      // 8 períodos mais recentes (crescente, índice 0 = mais antigo)
	UpdatedAt     time.Time `json:"updated_at"`
}

// Get godoc
// GET /api/v1/north-star
// Retorna métricas de conversão reais dos 6 produtos IIT.
// Sem autenticação — dados agregados, sem PII.
func (h *NorthStarHandler) Get(c *fiber.Ctx) error {
	now := time.Now().UTC()

	results := []NorthStarMetric{
		h.socialMake(c, now),
		h.nitro(c, now),
		h.libri(c, now),
		h.meuIngresso(c, now),
		h.brio(c, now),
		h.pulse(c, now),
	}

	return c.JSON(fiber.Map{
		"metrics":      results,
		"generated_at": now.Format(time.RFC3339),
	})
}

// ─── Métricas por produto ────────────────────────────────────────────────────

func (h *NorthStarHandler) socialMake(c *fiber.Ctx, now time.Time) NorthStarMetric {
	cur, prev, spark := h.weeklyCount(c, now, "socialmake", "post.published", "")
	return NorthStarMetric{
		Product:       "socialmake",
		MetricName:    "Posts publicados",
		Unit:          "por semana",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

func (h *NorthStarHandler) nitro(c *fiber.Ctx, now time.Time) NorthStarMetric {
	cur, prev, spark := h.weeklyCount(c, now, "nitro", "session.completed", "")
	return NorthStarMetric{
		Product:       "nitro",
		MetricName:    "Sessões de foco completadas",
		Unit:          "por semana",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

func (h *NorthStarHandler) libri(c *fiber.Ctx, now time.Time) NorthStarMetric {
	// Libri: book.status_changed com properties contendo {"to_status":"read"}
	// Usa jsonb @> operador com $3 para evitar concatenação de string — R1.
	filter := mustJSON(map[string]string{"to_status": "read"})
	cur, prev, spark := h.monthlyCountJSONB(c, now, "libri", "book.status_changed", filter)
	return NorthStarMetric{
		Product:       "libri",
		MetricName:    "Livros concluídos",
		Unit:          "no mês",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

func (h *NorthStarHandler) meuIngresso(c *fiber.Ctx, now time.Time) NorthStarMetric {
	cur, prev, spark := h.weeklyCount(c, now, "meuingresso", "ticket.purchased", "")
	return NorthStarMetric{
		Product:       "meuingresso",
		MetricName:    "Ingressos vendidos",
		Unit:          "por evento",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

func (h *NorthStarHandler) brio(c *fiber.Ctx, now time.Time) NorthStarMetric {
	cur, prev, spark := h.brioResolutionRate(c, now)
	return NorthStarMetric{
		Product:       "brio",
		MetricName:    "Conversas resolvidas pelo agente",
		Unit:          "% do total",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

func (h *NorthStarHandler) pulse(c *fiber.Ctx, now time.Time) NorthStarMetric {
	cur, prev, spark := h.pulseDistinctProjects(c, now)
	return NorthStarMetric{
		Product:       "pulse",
		MetricName:    "Projetos com flag tracked",
		Unit:          "ativos",
		CurrentValue:  cur,
		PreviousValue: prev,
		Sparkline:     spark,
		UpdatedAt:     now,
	}
}

// ─── Helpers de query ────────────────────────────────────────────────────────

// weeklyCount retorna contagem de eventos para: semana corrente, anterior e sparkline de 8 semanas.
// jsonbFilter: se não vazio, aplica AND properties @> $N::jsonb (prepared statement — R1).
func (h *NorthStarHandler) weeklyCount(c *fiber.Ctx, now time.Time, product, eventType, jsonbFilter string) (current, prev int64, sparkline []int64) {
	sparkline = make([]int64, 8)

	weekStart := now.Add(-7 * 24 * time.Hour)
	prevStart := now.Add(-14 * 24 * time.Hour)

	current = h.countEvents(c, product, eventType, weekStart, now, jsonbFilter)
	prev = h.countEvents(c, product, eventType, prevStart, weekStart, jsonbFilter)

	for i := 0; i < 8; i++ {
		wEnd := now.Add(time.Duration(-i) * 7 * 24 * time.Hour)
		wBegin := now.Add(time.Duration(-(i+1)) * 7 * 24 * time.Hour)
		sparkline[7-i] = h.countEvents(c, product, eventType, wBegin, wEnd, jsonbFilter)
	}
	return
}

// monthlyCountJSONB: semelhante ao weeklyCount mas usa janelas de 30 dias para current/prev.
// Sparkline ainda semanal (8 semanas) para visualização de tendência.
func (h *NorthStarHandler) monthlyCountJSONB(c *fiber.Ctx, now time.Time, product, eventType, jsonbFilter string) (current, prev int64, sparkline []int64) {
	sparkline = make([]int64, 8)

	monthStart := now.Add(-30 * 24 * time.Hour)
	prevStart := now.Add(-60 * 24 * time.Hour)

	current = h.countEvents(c, product, eventType, monthStart, now, jsonbFilter)
	prev = h.countEvents(c, product, eventType, prevStart, monthStart, jsonbFilter)

	for i := 0; i < 8; i++ {
		wEnd := now.Add(time.Duration(-i) * 7 * 24 * time.Hour)
		wBegin := now.Add(time.Duration(-(i+1)) * 7 * 24 * time.Hour)
		sparkline[7-i] = h.countEvents(c, product, eventType, wBegin, wEnd, jsonbFilter)
	}
	return
}

// countEvents executa COUNT com prepared statements ($1..$N). Nunca monta SQL com fmt.Sprintf.
// jsonbFilter (se não vazio) deve ser um JSON serializado — adicionado como parâmetro $5.
func (h *NorthStarHandler) countEvents(c *fiber.Ctx, product, eventType string, from, to time.Time, jsonbFilter string) int64 {
	var cnt int64
	if jsonbFilter != "" {
		_ = h.db.QueryRow(c.Context(),
			`SELECT COUNT(*) FROM events
			 WHERE product = $1
			   AND event_type = $2
			   AND created_at >= $3
			   AND created_at < $4
			   AND properties @> $5::jsonb`,
			product, eventType, from, to, jsonbFilter,
		).Scan(&cnt)
	} else {
		_ = h.db.QueryRow(c.Context(),
			`SELECT COUNT(*) FROM events
			 WHERE product = $1
			   AND event_type = $2
			   AND created_at >= $3
			   AND created_at < $4`,
			product, eventType, from, to,
		).Scan(&cnt)
	}
	return cnt
}

// pulseDistinctProjects conta projetos distintos com flag_evaluation nos últimos 7 dias.
func (h *NorthStarHandler) pulseDistinctProjects(c *fiber.Ctx, now time.Time) (current, prev int64, sparkline []int64) {
	sparkline = make([]int64, 8)

	weekStart := now.Add(-7 * 24 * time.Hour)
	prevStart := now.Add(-14 * 24 * time.Hour)

	countDistinct := func(from, to time.Time) int64 {
		var cnt int64
		_ = h.db.QueryRow(c.Context(),
			`SELECT COUNT(DISTINCT project_id) FROM events
			 WHERE event_type = $1
			   AND created_at >= $2
			   AND created_at < $3
			   AND project_id IS NOT NULL`,
			"flag_evaluation", from, to,
		).Scan(&cnt)
		return cnt
	}

	current = countDistinct(weekStart, now)
	prev = countDistinct(prevStart, weekStart)

	for i := 0; i < 8; i++ {
		wEnd := now.Add(time.Duration(-i) * 7 * 24 * time.Hour)
		wBegin := now.Add(time.Duration(-(i+1)) * 7 * 24 * time.Hour)
		sparkline[7-i] = countDistinct(wBegin, wEnd)
	}
	return
}

// brioResolutionRate calcula % de conversas resolvidas pelo agente nos últimos 7 dias.
// Retorna valor inteiro 0–100. Sparkline em pontos percentuais por semana.
//
// BKL-1015: o denominador NÃO pode contar todos os eventos do produto 'brio'
// (page_view, button_click, etc.). Deve contar apenas eventos que representam
// início de conversa: conversation.started e session.initiated.
func (h *NorthStarHandler) brioResolutionRate(c *fiber.Ctx, now time.Time) (current, prev int64, sparkline []int64) {
	sparkline = make([]int64, 8)

	weekStart := now.Add(-7 * 24 * time.Hour)
	prevStart := now.Add(-14 * 24 * time.Hour)

	// Conversa-start: eventos que representam início de conversa (denominador correto).
	// BKL-1015: antigamente contava QUALQUER evento do produto 'brio' no denominador,
	// incluindo page_view, button_click etc. — decisões de produto baseadas em taxa
	// matematicamente incorreta.
	conversationStarts := []string{"conversation.started", "session.initiated"}
	rate := func(from, to time.Time) int64 {
		var total, resolved int64
		_ = h.db.QueryRow(c.Context(),
			`SELECT COUNT(*) FROM events
			 WHERE product = $1 AND event_type = ANY($2::text[])
			   AND created_at >= $3 AND created_at < $4`,
			"brio", conversationStarts, from, to,
		).Scan(&total)
		_ = h.db.QueryRow(c.Context(),
			`SELECT COUNT(*) FROM events
			 WHERE product = $1 AND event_type = $2
			   AND created_at >= $3 AND created_at < $4`,
			"brio", "conversation.resolved_by_agent", from, to,
		).Scan(&resolved)
		if total == 0 {
			return 0
		}
		return (resolved * 100) / total
	}

	current = rate(weekStart, now)
	prev = rate(prevStart, weekStart)

	for i := 0; i < 8; i++ {
		wEnd := now.Add(time.Duration(-i) * 7 * 24 * time.Hour)
		wBegin := now.Add(time.Duration(-(i+1)) * 7 * 24 * time.Hour)
		sparkline[7-i] = rate(wBegin, wEnd)
	}
	return
}

// mustJSON serializa v para JSON string. Entra em panic apenas em tempo de compilação
// com valores estáticos — nunca chamado com input de usuário.
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		// Valor estático — nunca deve falhar
		panic(fmt.Sprintf("north_star_handler: mustJSON: %v", err))
	}
	return string(b)
}
