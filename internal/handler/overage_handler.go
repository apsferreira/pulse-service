package handler

// BKL-1071: Cobrança por overage de eventos no Pulse.
// Lógica: eventos acima da quota do plano custam R$0,02 por 1.000 eventos.
// O quota middleware continua retornando HTTP 429 em plano free quando ultrapassar o limite.
// Para planos pagos (starter/growth), eventos acima da quota são registrados para cobrança posterior.
//
// Endpoints (admin/interno, requer service token):
//   POST /api/v1/overage/calculate   — calcula e registra overage do mês corrente para todos os projetos
//   GET  /api/v1/overage             — lista overages pendentes de faturamento
//   GET  /api/v1/overage/:id         — detalha um overage específico
//   POST /api/v1/overage/:id/waive   — cancela/isenta cobrança (admin)

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	overagePricePer1k = 0.02 // R$ por 1.000 eventos acima da quota
)

// OverageHandler gerencia o ciclo de vida das cobranças por overage.
type OverageHandler struct {
	db           *pgxpool.Pool
	rdb          *redis.Client
	serviceToken string
}

func NewOverageHandler(db *pgxpool.Pool, rdb *redis.Client, serviceToken string) *OverageHandler {
	return &OverageHandler{db: db, rdb: rdb, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de overage (apenas service token — admin/interno).
func (h *OverageHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/overage", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/calculate", h.Calculate)
	api.Get("/", h.List)
	api.Get("/:id", h.GetByID)
	api.Post("/:id/waive", h.Waive)
}

// OverageSummary representa um registro de overage de um projeto.
type OverageSummary struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	ProjectName     string    `json:"project_name"`
	Plan            string    `json:"plan"`
	PeriodYear      int       `json:"period_year"`
	PeriodMonth     int       `json:"period_month"`
	QuotaLimit      int64     `json:"quota_limit"`
	EventsConsumed  int64     `json:"events_consumed"`
	OverageEvents   int64     `json:"overage_events"`
	AmountBRL       float64   `json:"amount_brl"`
	Status          string    `json:"status"`
	ExternalChargeID string   `json:"external_charge_id,omitempty"`
	InvoicedAt      *time.Time `json:"invoiced_at,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Calculate POST /api/v1/overage/calculate
// Varre todos os projetos pagos, calcula overage do mês atual a partir do Redis, e persiste.
// Idempotente: UPSERT na constraint project+período.
func (h *OverageHandler) Calculate(c *fiber.Ctx) error {
	now := time.Now().UTC()
	year, month := now.Year(), int(now.Month())

	// Busca todos os projetos com planos que permitem overage (starter e growth)
	rows, err := h.db.Query(c.Context(), `
		SELECT id, COALESCE(plan, 'free'), COALESCE(name, id::text)
		FROM projects
		WHERE is_active = true
		  AND COALESCE(plan, 'free') IN ('starter', 'growth')
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar projetos"})
	}
	defer rows.Close()

	type projectRow struct {
		ID   string
		Plan string
		Name string
	}
	var projects []projectRow
	for rows.Next() {
		var p projectRow
		if err := rows.Scan(&p.ID, &p.Plan, &p.Name); err == nil {
			projects = append(projects, p)
		}
	}
	rows.Close()

	registered := 0
	ctx := c.Context()

	for _, proj := range projects {
		limit := quotaLimitByPlan(proj.Plan)
		consumed := h.getConsumedFromRedis(ctx, proj.ID, year, month)

		if consumed <= limit {
			continue // sem overage
		}

		overageEvents := consumed - limit
		amountBRL := calculateOverageAmount(overageEvents)

		// UPSERT — idempotente por período
		_, err := h.db.Exec(ctx, `
			INSERT INTO overage_billing
			    (project_id, period_year, period_month, plan, quota_limit, events_consumed,
			     overage_events, price_per_1k, amount_brl, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending')
			ON CONFLICT (project_id, period_year, period_month)
			DO UPDATE SET
			    events_consumed = EXCLUDED.events_consumed,
			    overage_events  = EXCLUDED.overage_events,
			    amount_brl      = EXCLUDED.amount_brl,
			    updated_at      = NOW()
			WHERE overage_billing.status = 'pending'
		`, proj.ID, year, month, proj.Plan, limit, consumed, overageEvents, overagePricePer1k, amountBRL)
		if err != nil {
			continue
		}
		registered++
	}

	return c.JSON(fiber.Map{
		"period":     fmt.Sprintf("%d-%02d", year, month),
		"projects":   len(projects),
		"registered": registered,
	})
}

// List GET /api/v1/overage?status=pending&year=2026&month=4
func (h *OverageHandler) List(c *fiber.Ctx) error {
	status := c.Query("status", "pending")
	year := c.QueryInt("year", time.Now().Year())
	month := c.QueryInt("month", int(time.Now().Month()))

	rows, err := h.db.Query(c.Context(), `
		SELECT ob.id, ob.project_id, COALESCE(p.name, ob.project_id::text),
		       ob.plan, ob.period_year, ob.period_month,
		       ob.quota_limit, ob.events_consumed, ob.overage_events,
		       ob.amount_brl::float8, ob.status,
		       COALESCE(ob.external_charge_id, ''),
		       ob.invoiced_at, ob.paid_at, ob.created_at
		FROM overage_billing ob
		LEFT JOIN projects p ON p.id = ob.project_id
		WHERE ob.status = $1
		  AND ob.period_year = $2
		  AND ob.period_month = $3
		ORDER BY ob.amount_brl DESC
	`, status, year, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar overages"})
	}
	defer rows.Close()

	var list []OverageSummary
	for rows.Next() {
		var o OverageSummary
		var extChargeID string
		if err := rows.Scan(
			&o.ID, &o.ProjectID, &o.ProjectName,
			&o.Plan, &o.PeriodYear, &o.PeriodMonth,
			&o.QuotaLimit, &o.EventsConsumed, &o.OverageEvents,
			&o.AmountBRL, &o.Status, &extChargeID,
			&o.InvoicedAt, &o.PaidAt, &o.CreatedAt,
		); err == nil {
			if extChargeID != "" {
				o.ExternalChargeID = extChargeID
			}
			list = append(list, o)
		}
	}
	if list == nil {
		list = []OverageSummary{}
	}

	totalBRL := 0.0
	for _, o := range list {
		totalBRL += o.AmountBRL
	}

	return c.JSON(fiber.Map{
		"period":    fmt.Sprintf("%d-%02d", year, month),
		"status":    status,
		"count":     len(list),
		"total_brl": round2(totalBRL),
		"items":     list,
	})
}

// GetByID GET /api/v1/overage/:id
func (h *OverageHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var o OverageSummary
	var extChargeID string
	err := h.db.QueryRow(c.Context(), `
		SELECT ob.id, ob.project_id, COALESCE(p.name, ob.project_id::text),
		       ob.plan, ob.period_year, ob.period_month,
		       ob.quota_limit, ob.events_consumed, ob.overage_events,
		       ob.amount_brl::float8, ob.status,
		       COALESCE(ob.external_charge_id, ''),
		       ob.invoiced_at, ob.paid_at, ob.created_at
		FROM overage_billing ob
		LEFT JOIN projects p ON p.id = ob.project_id
		WHERE ob.id = $1
	`, id).Scan(
		&o.ID, &o.ProjectID, &o.ProjectName,
		&o.Plan, &o.PeriodYear, &o.PeriodMonth,
		&o.QuotaLimit, &o.EventsConsumed, &o.OverageEvents,
		&o.AmountBRL, &o.Status, &extChargeID,
		&o.InvoicedAt, &o.PaidAt, &o.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "overage não encontrado"})
	}
	if extChargeID != "" {
		o.ExternalChargeID = extChargeID
	}
	return c.JSON(o)
}

// Waive POST /api/v1/overage/:id/waive — cancela/isenta a cobrança
func (h *OverageHandler) Waive(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.db.Exec(c.Context(), `
		UPDATE overage_billing SET status = 'waived', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, id)
	if err != nil || result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "overage não encontrado ou não está pendente"})
	}
	return c.JSON(fiber.Map{"message": "overage isento"})
}

// getConsumedFromRedis lê o contador de eventos do mês da chave Redis do quota middleware.
func (h *OverageHandler) getConsumedFromRedis(ctx context.Context, projectID string, year, month int) int64 {
	if h.rdb == nil {
		return 0
	}
	key := fmt.Sprintf("quota:pulse:project:%s:%d-%02d", projectID, year, month)
	val, err := h.rdb.Get(ctx, key).Int64()
	if err != nil {
		return 0
	}
	return val
}

// quotaLimitByPlan retorna o limite mensal de eventos para o plano.
// Duplica a lógica do quota middleware para isolamento — evita acoplamento circular.
func quotaLimitByPlan(plan string) int64 {
	switch plan {
	case "starter":
		return 500_000
	case "growth":
		return 2_000_000
	case "pro":
		return -1 // ilimitado
	default:
		return 10_000
	}
}

// calculateOverageAmount calcula o valor em BRL do overage.
// Preço: R$0,02 por 1.000 eventos acima da quota (arredondado para cima em grupos de 1k).
func calculateOverageAmount(overageEvents int64) float64 {
	// Arredonda para cima em blocos de 1.000
	blocks := (overageEvents + 999) / 1000
	return round2(float64(blocks) * overagePricePer1k)
}
