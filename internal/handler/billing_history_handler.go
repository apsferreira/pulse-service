package handler

// billing_history_handler.go — histórico de mudanças de plano por projeto
// Rota: GET /api/v1/billing/history (protegida por X-Service-Token)

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingHistoryHandler expõe o histórico da tabela plan_changes.
type BillingHistoryHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewBillingHistoryHandler(db *pgxpool.Pool, serviceToken string) *BillingHistoryHandler {
	return &BillingHistoryHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra a rota de histórico de billing.
func (h *BillingHistoryHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/billing/history",
		middleware.ServiceTokenMiddleware(h.serviceToken),
		h.GetBillingHistory,
	)
}

// billingHistoryItem representa uma entrada de mudança de plano.
type billingHistoryItem struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	ProjectName     *string    `json:"project_name"`
	ProjectSlug     *string    `json:"project_slug"`
	FromPlan        string     `json:"from_plan"`
	ToPlan          string     `json:"to_plan"`
	CheckoutOrderID *string    `json:"checkout_order_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

// GetBillingHistory godoc
// GET /api/v1/billing/history
func (h *BillingHistoryHandler) GetBillingHistory(c *fiber.Ctx) error {
	rows, err := h.db.Query(c.Context(),
		`SELECT pc.id, pc.project_id, p.name AS project_name, p.slug AS project_slug,
		        pc.from_plan, pc.to_plan, pc.checkout_order_id, pc.created_at
		 FROM plan_changes pc
		 LEFT JOIN projects p ON p.id = pc.project_id
		 ORDER BY pc.created_at DESC
		 LIMIT 50`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	history := make([]billingHistoryItem, 0)
	for rows.Next() {
		var item billingHistoryItem
		if err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&item.ProjectName,
			&item.ProjectSlug,
			&item.FromPlan,
			&item.ToPlan,
			&item.CheckoutOrderID,
			&item.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"history": history})
}
