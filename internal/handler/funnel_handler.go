package handler

// BKL-1058: Funis de conversão server-side para o Pulse.
// API: POST /api/v1/funnels/analyze — encadeia N eventos em sequência com janela de tempo.
// Permite visualizar taxa de conversão por etapa (ex: page_view → signup → payment).

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FunnelHandler calcula funis de conversão baseados em eventos rastreados.
type FunnelHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewFunnelHandler(db *pgxpool.Pool, serviceToken string) *FunnelHandler {
	return &FunnelHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de funis.
func (h *FunnelHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/funnels", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/analyze", h.Analyze)
}

// FunnelRequest é o payload da análise de funil.
type FunnelRequest struct {
	ProjectID   string   `json:"project_id"`   // obrigatório
	Steps       []string `json:"steps"`         // eventos em ordem (ex: ["page_view","signup","payment"])
	WindowHours int      `json:"window_hours"`  // janela máxima entre primeiro e último evento (default: 24)
	BreakdownBy string   `json:"breakdown_by"`  // propriedade para segmentar (ex: "plan", "country") — opcional
	StartDate   string   `json:"start_date"`    // YYYY-MM-DD (default: 30 dias atrás)
	EndDate     string   `json:"end_date"`      // YYYY-MM-DD (default: hoje)
}

// FunnelStep representa uma etapa do funil com contagem e taxa de conversão.
type FunnelStep struct {
	Event          string  `json:"event"`
	Users          int     `json:"users"`           // usuários únicos que realizaram este evento
	ConversionRate float64 `json:"conversion_rate"` // % em relação à etapa anterior (100% para a primeira)
	DropoffRate    float64 `json:"dropoff_rate"`    // % que saiu nessa etapa
}

// FunnelResponse é a resposta da análise de funil.
type FunnelResponse struct {
	ProjectID   string             `json:"project_id"`
	Steps       []FunnelStep       `json:"steps"`
	TotalUsers  int                `json:"total_users"`        // usuários que iniciaram o funil
	Completed   int                `json:"completed"`          // chegaram ao final
	OverallRate float64            `json:"overall_rate"`       // % início → fim
	WindowHours int                `json:"window_hours"`
	Period      string             `json:"period"`
	Breakdown   []FunnelBreakdown  `json:"breakdown,omitempty"`
}

// FunnelBreakdown representa o funil segmentado por um valor de propriedade.
type FunnelBreakdown struct {
	Value string       `json:"value"`
	Steps []FunnelStep `json:"steps"`
}

// Analyze POST /api/v1/funnels/analyze
func (h *FunnelHandler) Analyze(c *fiber.Ctx) error {
	var req FunnelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}
	if len(req.Steps) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "funil requer ao menos 2 steps"})
	}
	if len(req.Steps) > 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "máximo de 10 steps por funil"})
	}
	if req.WindowHours <= 0 {
		req.WindowHours = 24
	}

	// Parse dates
	start, end, err := parseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Calcula funil principal
	steps, err := h.computeFunnel(c.Context(), req.ProjectID, req.Steps, req.WindowHours, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao calcular funil"})
	}

	totalUsers := 0
	completed := 0
	if len(steps) > 0 {
		totalUsers = steps[0].Users
		completed = steps[len(steps)-1].Users
	}
	overallRate := 0.0
	if totalUsers > 0 {
		overallRate = float64(completed) / float64(totalUsers) * 100
	}

	resp := FunnelResponse{
		ProjectID:   req.ProjectID,
		Steps:       steps,
		TotalUsers:  totalUsers,
		Completed:   completed,
		OverallRate: round2(overallRate),
		WindowHours: req.WindowHours,
		Period:      fmt.Sprintf("%s a %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
	}

	// Breakdown por propriedade (opcional)
	if req.BreakdownBy != "" {
		breakdown, err := h.computeBreakdown(c.Context(), req.ProjectID, req.Steps, req.WindowHours, start, end, req.BreakdownBy)
		if err == nil {
			resp.Breakdown = breakdown
		}
	}

	return c.JSON(resp)
}

// computeFunnel executa a query CTE de funil step a step.
// Para cada step N, conta usuários únicos que fizeram o step N
// dentro da janela de tempo APÓS terem feito o step N-1.
func (h *FunnelHandler) computeFunnel(ctx context.Context, projectID string, steps []string, windowHours int, start, end time.Time) ([]FunnelStep, error) {
	counts := make([]int, len(steps))

	// Step 0: todos os usuários que fizeram o primeiro evento no período
	err := h.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM events
		WHERE properties->>'project_id' = $1
		  AND event_type = $2
		  AND created_at BETWEEN $3 AND $4
		  AND user_id IS NOT NULL AND user_id != ''
	`, projectID, steps[0], start, end).Scan(&counts[0])
	if err != nil {
		return nil, fmt.Errorf("funnel: step 0: %w", err)
	}

	// Steps subsequentes: usuários que fizeram step[i] após step[i-1] dentro da janela
	for i := 1; i < len(steps); i++ {
		err := h.db.QueryRow(ctx, `
			SELECT COUNT(DISTINCT e2.user_id)
			FROM events e1
			JOIN events e2 ON e1.user_id = e2.user_id
			WHERE e1.properties->>'project_id' = $1
			  AND e1.event_type = $2
			  AND e1.created_at BETWEEN $3 AND $4
			  AND e1.user_id IS NOT NULL AND e1.user_id != ''
			  AND e2.properties->>'project_id' = $1
			  AND e2.event_type = $5
			  AND e2.created_at > e1.created_at
			  AND e2.created_at <= e1.created_at + ($6 || ' hours')::INTERVAL
		`, projectID, steps[i-1], start, end, steps[i], windowHours).Scan(&counts[i])
		if err != nil {
			return nil, fmt.Errorf("funnel: step %d: %w", i, err)
		}
	}

	// Monta resultado
	result := make([]FunnelStep, len(steps))
	for i, event := range steps {
		convRate := 100.0
		dropRate := 0.0
		if i > 0 && counts[i-1] > 0 {
			convRate = float64(counts[i]) / float64(counts[i-1]) * 100
			dropRate = 100 - convRate
		}
		result[i] = FunnelStep{
			Event:          event,
			Users:          counts[i],
			ConversionRate: round2(convRate),
			DropoffRate:    round2(dropRate),
		}
	}
	return result, nil
}

// computeBreakdown calcula o funil segmentado por um valor de propriedade.
func (h *FunnelHandler) computeBreakdown(ctx context.Context, projectID string, steps []string, windowHours int, start, end time.Time, breakdownBy string) ([]FunnelBreakdown, error) {
	// Pega os top 10 valores distintos da propriedade
	rows, err := h.db.Query(ctx, `
		SELECT DISTINCT properties->>$1 AS val
		FROM events
		WHERE properties->>'project_id' = $2
		  AND event_type = $3
		  AND created_at BETWEEN $4 AND $5
		  AND properties->>$1 IS NOT NULL
		LIMIT 10
	`, breakdownBy, projectID, steps[0], start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err == nil && v != "" {
			values = append(values, v)
		}
	}

	var breakdown []FunnelBreakdown
	for _, val := range values {
		// Para cada valor, filtra eventos onde a propriedade tem aquele valor
		// Simplificado: contar step 0 e step final apenas para manter a query leve
		var firstCount, lastCount int
		_ = h.db.QueryRow(ctx, `
			SELECT COUNT(DISTINCT user_id) FROM events
			WHERE properties->>'project_id' = $1 AND event_type = $2
			  AND created_at BETWEEN $3 AND $4
			  AND properties->>$5 = $6
		`, projectID, steps[0], start, end, breakdownBy, val).Scan(&firstCount)

		_ = h.db.QueryRow(ctx, `
			SELECT COUNT(DISTINCT e2.user_id)
			FROM events e1 JOIN events e2 ON e1.user_id = e2.user_id
			WHERE e1.properties->>'project_id' = $1 AND e1.event_type = $2
			  AND e1.created_at BETWEEN $3 AND $4
			  AND e1.properties->>$5 = $6
			  AND e2.event_type = $7
			  AND e2.created_at > e1.created_at
			  AND e2.created_at <= e1.created_at + ($8 || ' hours')::INTERVAL
		`, projectID, steps[0], start, end, breakdownBy, val, steps[len(steps)-1], windowHours).Scan(&lastCount)

		rate := 0.0
		if firstCount > 0 {
			rate = float64(lastCount) / float64(firstCount) * 100
		}

		breakdown = append(breakdown, FunnelBreakdown{
			Value: val,
			Steps: []FunnelStep{
				{Event: steps[0], Users: firstCount, ConversionRate: 100, DropoffRate: 0},
				{Event: steps[len(steps)-1], Users: lastCount, ConversionRate: round2(rate), DropoffRate: round2(100 - rate)},
			},
		})
	}
	return breakdown, nil
}

// parseDateRange valida e converte as datas do request.
func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	end := time.Now().UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
	start := end.AddDate(0, 0, -30)

	if startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("start_date inválido: use YYYY-MM-DD")
		}
		start = t.UTC()
	}
	if endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("end_date inválido: use YYYY-MM-DD")
		}
		end = t.UTC().Add(24*time.Hour - time.Second)
	}
	return start, end, nil
}

// round2 arredonda para 2 casas decimais.
func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
