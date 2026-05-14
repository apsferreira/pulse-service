package handler

// Dashboard Analytics Unificado — agrega métricas de todos os projetos cadastrados.
// Endpoint interno (service token) que consolida DAU/MAU, conversão, churn e MRR.
//
// Endpoints:
//   GET /api/v1/admin/dashboard         — overview consolidado de todos os projetos
//   GET /api/v1/admin/projects/:slug    — métricas detalhadas por projeto
//   GET /api/v1/admin/mrr               — MRR consolidado + breakdown por plano

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminDashboardHandler agrega métricas de todos os projetos.
type AdminDashboardHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewAdminDashboardHandler(db *pgxpool.Pool, serviceToken string) *AdminDashboardHandler {
	return &AdminDashboardHandler{db: db, serviceToken: serviceToken}
}

func (h *AdminDashboardHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/admin", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/dashboard", h.Overview)
	api.Get("/projects/:slug", h.ProjectDetail)
	api.Get("/mrr", h.MRR)
}

// ProjectMetrics representa métricas de um projeto no período.
type ProjectMetrics struct {
	ProjectID    string `json:"project_id"`
	ProjectName  string `json:"project_name"`
	Slug         string `json:"slug"`
	Plan         string `json:"plan"`
	DAU          int64  `json:"dau"`
	MAU          int64  `json:"mau"`
	EventsToday  int64  `json:"events_today"`
	EventsMonth  int64  `json:"events_month"`
	FlagsActive  int64  `json:"flags_active"`
	FlagsEnabled int64  `json:"flags_enabled"`
	QuotaUsedPct int    `json:"quota_used_pct"`
}

// Overview GET /api/v1/admin/dashboard
func (h *AdminDashboardHandler) Overview(c *fiber.Ctx) error {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	start30d := now.Add(-30 * 24 * time.Hour)

	projectRows, err := h.db.Query(c.Context(), `
		SELECT id, name, slug, COALESCE(plan, 'free')
		FROM projects WHERE is_active = true
		ORDER BY name ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar projetos"})
	}
	defer projectRows.Close()

	type projectRow struct {
		ID   string
		Name string
		Slug string
		Plan string
	}
	var projects []projectRow
	for projectRows.Next() {
		var p projectRow
		if err := projectRows.Scan(&p.ID, &p.Name, &p.Slug, &p.Plan); err == nil {
			projects = append(projects, p)
		}
	}
	projectRows.Close()

	var metrics []ProjectMetrics
	totalDAU, totalMAU, totalEventsToday, totalEventsMonth := int64(0), int64(0), int64(0), int64(0)

	for _, proj := range projects {
		m := ProjectMetrics{
			ProjectID:   proj.ID,
			ProjectName: proj.Name,
			Slug:        proj.Slug,
			Plan:        proj.Plan,
		}

		_ = h.db.QueryRow(c.Context(), `
			SELECT COUNT(DISTINCT properties->>'user_id')
			FROM events
			WHERE properties->>'project_id' = $1
			  AND created_at >= $2
			  AND properties->>'user_id' IS NOT NULL
		`, proj.ID, startOfDay).Scan(&m.DAU)

		_ = h.db.QueryRow(c.Context(), `
			SELECT COUNT(DISTINCT properties->>'user_id')
			FROM events
			WHERE properties->>'project_id' = $1
			  AND created_at >= $2
			  AND properties->>'user_id' IS NOT NULL
		`, proj.ID, start30d).Scan(&m.MAU)

		_ = h.db.QueryRow(c.Context(), `
			SELECT COUNT(*) FROM events
			WHERE properties->>'project_id' = $1 AND created_at >= $2
		`, proj.ID, startOfDay).Scan(&m.EventsToday)

		_ = h.db.QueryRow(c.Context(), `
			SELECT COUNT(*) FROM events
			WHERE properties->>'project_id' = $1 AND created_at >= $2
		`, proj.ID, start30d).Scan(&m.EventsMonth)

		_ = h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM feature_flags WHERE project_id = $1`, proj.ID).Scan(&m.FlagsActive)
		_ = h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM feature_flags WHERE project_id = $1 AND is_enabled = true`, proj.ID).Scan(&m.FlagsEnabled)

		quotaLimit := quotaLimitByPlan(proj.Plan)
		if quotaLimit > 0 && m.EventsMonth > 0 {
			pct := int(m.EventsMonth * 100 / quotaLimit)
			if pct > 100 {
				pct = 100
			}
			m.QuotaUsedPct = pct
		}

		totalDAU += m.DAU
		totalMAU += m.MAU
		totalEventsToday += m.EventsToday
		totalEventsMonth += m.EventsMonth

		metrics = append(metrics, m)
	}

	if metrics == nil {
		metrics = []ProjectMetrics{}
	}

	return c.JSON(fiber.Map{
		"period": fiber.Map{
			"from": start30d.Format(time.RFC3339),
			"to":   now.Format(time.RFC3339),
		},
		"totals": fiber.Map{
			"projects":     len(projects),
			"dau":          totalDAU,
			"mau":          totalMAU,
			"events_today": totalEventsToday,
			"events_month": totalEventsMonth,
		},
		"products": metrics,
	})
}

// ProjectDetail GET /api/v1/admin/projects/:slug
func (h *AdminDashboardHandler) ProjectDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")

	var projectID, projectName, plan string
	err := h.db.QueryRow(c.Context(), `
		SELECT id, name, COALESCE(plan, 'free')
		FROM projects WHERE slug = $1 AND is_active = true
	`, slug).Scan(&projectID, &projectName, &plan)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "projeto não encontrado"})
	}

	now := time.Now().UTC()
	start7d := now.Add(-7 * 24 * time.Hour)
	start30d := now.Add(-30 * 24 * time.Hour)

	rows, _ := h.db.Query(c.Context(), `
		SELECT DATE(created_at AT TIME ZONE 'UTC') as day, COUNT(*) as events
		FROM events
		WHERE properties->>'project_id' = $1 AND created_at >= $2
		GROUP BY day ORDER BY day ASC
	`, projectID, start7d)
	type DailyStat struct {
		Day    string `json:"day"`
		Events int64  `json:"events"`
	}
	var daily []DailyStat
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d DailyStat
			if err := rows.Scan(&d.Day, &d.Events); err == nil {
				daily = append(daily, d)
			}
		}
	}
	if daily == nil {
		daily = []DailyStat{}
	}

	topRows, _ := h.db.Query(c.Context(), `
		SELECT event, COUNT(*) as cnt
		FROM events
		WHERE properties->>'project_id' = $1 AND created_at >= $2
		GROUP BY event ORDER BY cnt DESC LIMIT 5
	`, projectID, start30d)
	type TopEvent struct {
		Event string `json:"event"`
		Count int64  `json:"count"`
	}
	var topEvents []TopEvent
	if topRows != nil {
		defer topRows.Close()
		for topRows.Next() {
			var e TopEvent
			if err := topRows.Scan(&e.Event, &e.Count); err == nil {
				topEvents = append(topEvents, e)
			}
		}
	}
	if topEvents == nil {
		topEvents = []TopEvent{}
	}

	var totalFlags, enabledFlags, experimentFlags int64
	_ = h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM feature_flags WHERE project_id = $1`, projectID).Scan(&totalFlags)
	_ = h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM feature_flags WHERE project_id = $1 AND is_enabled = true`, projectID).Scan(&enabledFlags)
	_ = h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM experiments WHERE project_id = $1 AND status = 'running'`, projectID).Scan(&experimentFlags)

	return c.JSON(fiber.Map{
		"project_id":   projectID,
		"name":         projectName,
		"slug":         slug,
		"plan":         plan,
		"daily_events": daily,
		"top_events":   topEvents,
		"flags": fiber.Map{
			"total":       totalFlags,
			"enabled":     enabledFlags,
			"experiments": experimentFlags,
		},
	})
}

// MRR GET /api/v1/admin/mrr
func (h *AdminDashboardHandler) MRR(c *fiber.Ctx) error {
	planPrices := map[string]float64{
		"free":    0,
		"starter": 49,
		"growth":  99,
		"pro":     197,
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT COALESCE(plan, 'free'), COUNT(*) as count
		FROM projects WHERE is_active = true
		GROUP BY plan
		ORDER BY plan
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao calcular MRR"})
	}
	defer rows.Close()

	type PlanCount struct {
		Plan  string  `json:"plan"`
		Count int64   `json:"count"`
		Price float64 `json:"price_brl"`
		MRR   float64 `json:"mrr_brl"`
	}
	var breakdown []PlanCount
	totalMRR := 0.0

	for rows.Next() {
		var pc PlanCount
		if err := rows.Scan(&pc.Plan, &pc.Count); err == nil {
			pc.Price = planPrices[pc.Plan]
			pc.MRR = pc.Price * float64(pc.Count)
			totalMRR += pc.MRR
			breakdown = append(breakdown, pc)
		}
	}

	if breakdown == nil {
		breakdown = []PlanCount{}
	}

	return c.JSON(fiber.Map{
		"mrr_brl":   round2(totalMRR),
		"arr_brl":   round2(totalMRR * 12),
		"currency":  "BRL",
		"breakdown": breakdown,
		"note":      fmt.Sprintf("Estimativa baseada em planos ativos em %s", time.Now().Format("2006-01-02")),
	})
}
