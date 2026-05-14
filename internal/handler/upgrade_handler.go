package handler

// BKL-1086: Upgrade recommender inteligente para o Pulse.
// Avalia uso do tenant (flags ativas, eventos, projetos) e recomenda o plano ideal
// com base em comportamento, não em calendário.
// Endpoint: GET /api/v1/upgrade/recommend?project_id=...

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpgradeHandler retorna recomendações de upgrade baseadas em comportamento.
type UpgradeHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewUpgradeHandler(db *pgxpool.Pool, serviceToken string) *UpgradeHandler {
	return &UpgradeHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de upgrade recommender.
func (h *UpgradeHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/upgrade", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/recommend", h.Recommend)
}

// UpgradeRecommendation é a resposta da análise de upgrade.
type UpgradeRecommendation struct {
	CurrentPlan      string   `json:"current_plan"`
	RecommendedPlan  string   `json:"recommended_plan"`
	ShouldUpgrade    bool     `json:"should_upgrade"`
	Triggers         []string `json:"triggers"`           // motivos que acionaram a recomendação
	QuotaUsage       QuotaInfo `json:"quota_usage"`
	CallToAction     string   `json:"call_to_action"`
	UpgradeURL       string   `json:"upgrade_url,omitempty"`
}

type QuotaInfo struct {
	FlagsUsed      int `json:"flags_used"`
	FlagsLimit     int `json:"flags_limit"`     // -1 = ilimitado
	FlagsPct       int `json:"flags_pct"`       // % de uso (0-100, -1 se ilimitado)
	EventsLast30d  int `json:"events_last_30d"`
	EventsLimit    int `json:"events_limit"`    // -1 = ilimitado
	EventsPct      int `json:"events_pct"`
	ProjectsUsed   int `json:"projects_used"`
	ProjectsLimit  int `json:"projects_limit"`  // -1 = ilimitado
	ProjectsPct    int `json:"projects_pct"`
}

// planLimits define os limites por plano.
var planLimits = map[string]struct {
	Flags    int // -1 = ilimitado
	Events   int // por mês
	Projects int // -1 = ilimitado
}{
	"free":    {Flags: 5, Events: 10_000, Projects: 1},
	"starter": {Flags: -1, Events: 500_000, Projects: 3},
	"growth":  {Flags: -1, Events: 2_000_000, Projects: 10},
	"pro":     {Flags: -1, Events: -1, Projects: -1},
}

// planOrder para determinar próximo upgrade.
var planOrder = []string{"free", "starter", "growth", "pro"}

// Recommend GET /api/v1/upgrade/recommend?project_id=...
func (h *UpgradeHandler) Recommend(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	ctx := c.Context()

	// Busca plano atual do projeto
	currentPlan, err := h.getProjectPlan(ctx, projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "projeto não encontrado"})
	}

	// Coleta métricas de uso
	quota, err := h.collectQuota(ctx, projectID, currentPlan)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao coletar métricas"})
	}

	rec := buildRecommendation(currentPlan, quota)
	return c.JSON(rec)
}

// getProjectPlan retorna o plano atual do projeto.
func (h *UpgradeHandler) getProjectPlan(ctx context.Context, projectID string) (string, error) {
	var plan string
	err := h.db.QueryRow(ctx, `SELECT COALESCE(plan, 'free') FROM projects WHERE id = $1`, projectID).Scan(&plan)
	if err != nil {
		return "", fmt.Errorf("upgrade: getProjectPlan: %w", err)
	}
	return plan, nil
}

// collectQuota coleta as métricas de uso do projeto.
func (h *UpgradeHandler) collectQuota(ctx context.Context, projectID, plan string) (QuotaInfo, error) {
	limits := planLimits[plan]
	if limits.Flags == 0 {
		limits = planLimits["free"] // fallback para plano desconhecido
	}

	// Conta flags ativas no projeto
	var flagsUsed int
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM feature_flags WHERE project_id = $1`, projectID).Scan(&flagsUsed)

	// Conta eventos nos últimos 30 dias
	var eventsLast30d int
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM events
		WHERE properties->>'project_id' = $1 AND created_at >= $2
	`, projectID, since).Scan(&eventsLast30d)

	// Conta projetos do tenant (usa owner_email para correlacionar tenants)
	var projectsUsed int
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM projects p2
		WHERE p2.owner_email = (SELECT owner_email FROM projects WHERE id = $1)
		  AND p2.is_active = true
	`, projectID).Scan(&projectsUsed)

	quota := QuotaInfo{
		FlagsUsed:     flagsUsed,
		FlagsLimit:    limits.Flags,
		EventsLast30d: eventsLast30d,
		EventsLimit:   limits.Events,
		ProjectsUsed:  projectsUsed,
		ProjectsLimit: limits.Projects,
	}

	// Calcula percentuais
	if limits.Flags > 0 {
		quota.FlagsPct = intMin(flagsUsed*100/limits.Flags, 100)
	} else {
		quota.FlagsPct = -1
	}
	if limits.Events > 0 {
		quota.EventsPct = intMin(eventsLast30d*100/limits.Events, 100)
	} else {
		quota.EventsPct = -1
	}
	if limits.Projects > 0 {
		quota.ProjectsPct = intMin(projectsUsed*100/limits.Projects, 100)
	} else {
		quota.ProjectsPct = -1
	}

	return quota, nil
}

// buildRecommendation decide se deve recomendar upgrade e qual plano.
func buildRecommendation(currentPlan string, quota QuotaInfo) UpgradeRecommendation {
	var triggers []string

	// Avalia triggers de comportamento
	if quota.FlagsPct >= 80 {
		triggers = append(triggers, fmt.Sprintf("Você usou %d%% do limite de feature flags (%d/%d)", quota.FlagsPct, quota.FlagsUsed, quota.FlagsLimit))
	}
	if quota.EventsPct >= 80 {
		triggers = append(triggers, fmt.Sprintf("Você usou %d%% da quota de eventos (%d/%d este mês)", quota.EventsPct, quota.EventsLast30d, quota.EventsLimit))
	}
	if quota.ProjectsPct >= 80 {
		triggers = append(triggers, fmt.Sprintf("Você usou %d%% do limite de projetos (%d/%d)", quota.ProjectsPct, quota.ProjectsUsed, quota.ProjectsLimit))
	}
	if quota.FlagsUsed >= 3 && currentPlan == "free" {
		triggers = append(triggers, "Você já criou 3+ feature flags — você está pronto para o Starter")
	}

	shouldUpgrade := len(triggers) > 0

	// Determina plano recomendado
	recommendedPlan := currentPlan
	if shouldUpgrade {
		recommendedPlan = nextPlan(currentPlan)
	}

	// Monta CTA e URL
	cta, upgradeURL := buildCTA(currentPlan, recommendedPlan, shouldUpgrade)

	return UpgradeRecommendation{
		CurrentPlan:     currentPlan,
		RecommendedPlan: recommendedPlan,
		ShouldUpgrade:   shouldUpgrade,
		Triggers:        triggers,
		QuotaUsage:      quota,
		CallToAction:    cta,
		UpgradeURL:      upgradeURL,
	}
}

// nextPlan retorna o próximo plano na hierarquia.
func nextPlan(current string) string {
	for i, p := range planOrder {
		if p == current && i < len(planOrder)-1 {
			return planOrder[i+1]
		}
	}
	return current
}

// buildCTA gera o texto e URL do call-to-action.
func buildCTA(current, recommended string, shouldUpgrade bool) (string, string) {
	if !shouldUpgrade || current == "pro" {
		return "Você está bem! Continue aproveitando o " + capitalize(current) + ".", ""
	}

	planURLs := map[string]string{
		"starter": "/checkout?plan=starter",
		"growth":  "/checkout?plan=growth",
		"pro":     "/checkout?plan=pro",
	}

	cta := fmt.Sprintf("Faça upgrade para o %s e continue crescendo sem limites.", capitalize(recommended))
	url := planURLs[recommended]
	return cta, url
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
