package handler

// BKL-997: Funis de feature flags — correlação flag → conversão.
// Endpoint: GET /api/v1/projects/:project_id/funnel/:flag_key
//
// Busca usuários que receberam a flag (via flag_evaluations) e verifica
// quais completaram eventos-chave de conversão (purchase, signup, upgrade).
//
// Resposta:
//   {
//     "flag_key": "checkout-v2",
//     "enabled_users": 150,
//     "control_users": 120,
//     "conversion_enabled": 0.25,
//     "conversion_control": 0.18,
//     "lift": 38.89
//   }

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FlagFunnelHandler calcula correlação entre exposição a flags e eventos de conversão.
type FlagFunnelHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

// NewFlagFunnelHandler cria o handler.
func NewFlagFunnelHandler(db *pgxpool.Pool, serviceToken string) *FlagFunnelHandler {
	return &FlagFunnelHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de funil de flags.
func (h *FlagFunnelHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/projects", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/:project_id/funnel/:flag_key", h.GetFlagFunnel)
}

// FlagFunnelResponse é a resposta do endpoint de funil de flags.
type FlagFunnelResponse struct {
	FlagKey            string   `json:"flag_key"`
	ProjectID          string   `json:"project_id"`
	EnabledUsers       int64    `json:"enabled_users"`        // usuários que receberam flag=true
	ControlUsers       int64    `json:"control_users"`        // usuários que receberam flag=false
	ConversionEnabled  float64  `json:"conversion_enabled"`   // taxa de conversão no grupo enabled (0.0–1.0)
	ConversionControl  float64  `json:"conversion_control"`   // taxa de conversão no grupo control (0.0–1.0)
	Lift               float64  `json:"lift"`                 // lift relativo em %
	ConversionEvents   []string `json:"conversion_events"`    // eventos contados como conversão
	Period             string   `json:"period"`
}

// GetFlagFunnel godoc
// GET /api/v1/projects/:project_id/funnel/:flag_key
// Query params opcionais:
//   - events: lista separada por vírgula (default: "purchase,signup,upgrade")
//   - days: janela em dias (default: 30)
func (h *FlagFunnelHandler) GetFlagFunnel(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	flagKey := c.Params("flag_key")

	if projectID == "" || flagKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id e flag_key são obrigatórios",
		})
	}
	if !validKeyPattern.MatchString(flagKey) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "flag_key inválida",
		})
	}

	// Eventos de conversão: default ou via query param
	eventsParam := c.Query("events", "purchase,signup,upgrade")
	conversionEvents := splitAndTrim(eventsParam)
	if len(conversionEvents) == 0 {
		conversionEvents = []string{"purchase", "signup", "upgrade"}
	}
	if len(conversionEvents) > 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "máximo de 10 eventos de conversão",
		})
	}

	days := c.QueryInt("days", 30)
	if days <= 0 || days > 365 {
		days = 30
	}

	ctx := c.Context()

	// 1. Conta usuários expostos por grupo (enabled=true vs enabled=false)
	// Usa flag_evaluations que registra cada avaliação com result=true/false
	var enabledUsers, controlUsers int64
	err := h.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT CASE WHEN result = true  THEN user_id END),
			COUNT(DISTINCT CASE WHEN result = false THEN user_id END)
		FROM flag_evaluations
		WHERE flag_key   = $1
		  AND project_id = $2
		  AND user_id    IS NOT NULL
		  AND evaluated_at >= NOW() - ($3 || ' days')::INTERVAL
	`, flagKey, projectID, days).Scan(&enabledUsers, &controlUsers)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "erro ao buscar avaliações de flag",
		})
	}

	// 2. Conta conversões por grupo
	// Um usuário "converteu" se realizou qualquer um dos eventos de conversão
	// APÓS ter sido exposto à flag, dentro da janela.
	convPlaceholders := buildInPlaceholders(3, len(conversionEvents))
	args := buildConvArgs(flagKey, projectID, days, conversionEvents)

	var enabledConv, controlConv int64
	convQuery := fmt.Sprintf(`
		SELECT
			COUNT(DISTINCT CASE WHEN fe.result = true  THEN fe.user_id END),
			COUNT(DISTINCT CASE WHEN fe.result = false THEN fe.user_id END)
		FROM flag_evaluations fe
		JOIN events e ON e.user_id = fe.user_id
		WHERE fe.flag_key   = $1
		  AND fe.project_id = $2
		  AND fe.user_id    IS NOT NULL
		  AND fe.evaluated_at >= NOW() - ($3 || ' days')::INTERVAL
		  AND e.event_type  IN (%s)
		  AND e.created_at  >= fe.evaluated_at
		  AND e.created_at  <= fe.evaluated_at + ($3 || ' days')::INTERVAL
		  AND (e.properties->>'project_id' = $2 OR e.properties->>'project_id' IS NULL)
	`, convPlaceholders)

	if err := h.db.QueryRow(ctx, convQuery, args...).Scan(&enabledConv, &controlConv); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "erro ao calcular conversões",
		})
	}

	// 3. Calcula taxas e lift
	convEnabled := safeDiv(float64(enabledConv), float64(enabledUsers))
	convControl := safeDiv(float64(controlConv), float64(controlUsers))

	lift := 0.0
	if convControl > 0 {
		lift = (convEnabled - convControl) / convControl * 100
	}

	return c.JSON(FlagFunnelResponse{
		FlagKey:           flagKey,
		ProjectID:         projectID,
		EnabledUsers:      enabledUsers,
		ControlUsers:      controlUsers,
		ConversionEnabled: round2(convEnabled),
		ConversionControl: round2(convControl),
		Lift:              round2(lift),
		ConversionEvents:  conversionEvents,
		Period:            fmt.Sprintf("últimos %d dias", days),
	})
}

// buildInPlaceholders gera "$start, $start+1, ..." para N valores.
func buildInPlaceholders(start, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(parts, ", ")
}

// buildConvArgs monta o slice de args para a query de conversão.
// Posições: $1=flagKey, $2=projectID, $3=days, $4...$N=conversionEvents
func buildConvArgs(flagKey, projectID string, days int, events []string) []interface{} {
	args := []interface{}{flagKey, projectID, days}
	for _, e := range events {
		args = append(args, e)
	}
	return args
}

// safeDiv divide dois float64, retornando 0 se o divisor for zero.
func safeDiv(num, den float64) float64 {
	if den == 0 {
		return 0
	}
	return num / den
}

// splitAndTrim divide uma string por vírgula e remove espaços.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
