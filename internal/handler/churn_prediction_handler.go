package handler

// BKL-1065: Churn prediction com regressão logística simplificada (sem libs de ML).
// Features coletadas dos últimos 30 dias do projeto.
// Pesos heurísticos: recência (0.4), frequência (0.3), adoção (0.3).
// Score = sigmoid(w1*recência + w2*frequência + w3*adoção)
// Risk: low < 0.33, 0.33 <= medium <= 0.66, high > 0.66

import (
	"context"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChurnPredictionHandler expõe o endpoint de churn risk score.
type ChurnPredictionHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

// NewChurnPredictionHandler cria o handler.
func NewChurnPredictionHandler(db *pgxpool.Pool, serviceToken string) *ChurnPredictionHandler {
	return &ChurnPredictionHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes monta a rota de churn prediction.
func (h *ChurnPredictionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/projects", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/:id/churn-risk", h.ChurnRisk)
}

// ChurnFactor representa um fator contribuinte para o score de churn.
type ChurnFactor struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`   // peso no modelo (0.0-1.0)
	Value  float64 `json:"value"`    // valor normalizado (0.0-1.0)
	Score  float64 `json:"score"`    // contribuição = weight * value
}

// ChurnRiskResponse é a resposta do endpoint de churn risk.
type ChurnRiskResponse struct {
	ProjectID string        `json:"project_id"`
	Score     float64       `json:"score"`   // 0.0-1.0; alto = maior risco de churn
	Risk      string        `json:"risk"`    // "low" | "medium" | "high"
	Factors   []ChurnFactor `json:"factors"`
	Period    string        `json:"period"`  // ex: "últimos 30 dias"
	ComputedAt time.Time    `json:"computed_at"`
}

// churnFeatures são os dados brutos coletados para o projeto.
type churnFeatures struct {
	eventCount          int
	sessionCount        int
	daysSinceLastEvent  int
	flagsEvaluated      int
	featureAdoptionRate float64 // proporção de flags distintas avaliadas / total de flags
}

// ChurnRisk godoc
// GET /api/v1/projects/:id/churn-risk
// Calcula score de churn risk usando regressão logística simplificada. BKL-1065.
func (h *ChurnPredictionHandler) ChurnRisk(c *fiber.Ctx) error {
	projectID := c.Params("id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project id obrigatório"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()

	features, err := h.collectFeatures(ctx, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao coletar features"})
	}

	// Normalização das features para valores 0.0-1.0
	// Recência: quanto mais recente, menor o risco (invertida)
	// 0 dias = 1.0 (ativo hoje), 30+ dias = 0.0 (sem acesso)
	recenciaNorm := 0.0
	if features.daysSinceLastEvent <= 30 {
		recenciaNorm = 1.0 - float64(features.daysSinceLastEvent)/30.0
	}

	// Frequência: normaliza por 100 eventos nos últimos 30 dias = 1.0
	frequenciaNorm := math.Min(float64(features.eventCount)/100.0, 1.0)

	// Adoção: já normalizada (proporção 0.0-1.0)
	adocaoNorm := features.featureAdoptionRate

	// Pesos do modelo (somam 1.0)
	const (
		wRecencia    = 0.4
		wFrequencia  = 0.3
		wAdocao      = 0.3
	)

	// Regressão logística: z = -(w1*x1 + w2*x2 + w3*x3)
	// Invertemos o sinal: z negativo porque valores altos das features = MENOS churn
	// Score de churn = P(churn) = sigmoid(-engagement)
	engagement := wRecencia*recenciaNorm + wFrequencia*frequenciaNorm + wAdocao*adocaoNorm

	// Bias heurístico: projetos com 0 atividade nos últimos 30 dias têm churn alto
	// sigmoid(0) = 0.5, queremos que engagement=0 → score=1.0 (churn total)
	// Usamos: score = 1 - sigmoid(4 * engagement - 2)
	// Isso mapeia engagement=0 → score=~0.88, engagement=0.5 → score=0.5, engagement=1 → score=~0.12
	z := 4.0*engagement - 2.0
	churnScore := 1.0 - sigmoid(z)
	churnScore = math.Round(churnScore*10000) / 10000

	riskLevel := "low"
	switch {
	case churnScore > 0.66:
		riskLevel = "high"
	case churnScore > 0.33:
		riskLevel = "medium"
	}

	factors := []ChurnFactor{
		{
			Name:   "recência",
			Weight: wRecencia,
			Value:  math.Round(recenciaNorm*10000) / 10000,
			Score:  math.Round(wRecencia*recenciaNorm*10000) / 10000,
		},
		{
			Name:   "frequência",
			Weight: wFrequencia,
			Value:  math.Round(frequenciaNorm*10000) / 10000,
			Score:  math.Round(wFrequencia*frequenciaNorm*10000) / 10000,
		},
		{
			Name:   "adoção",
			Weight: wAdocao,
			Value:  math.Round(adocaoNorm*10000) / 10000,
			Score:  math.Round(wAdocao*adocaoNorm*10000) / 10000,
		},
	}

	return c.JSON(ChurnRiskResponse{
		ProjectID:  projectID,
		Score:      churnScore,
		Risk:       riskLevel,
		Factors:    factors,
		Period:     "últimos 30 dias",
		ComputedAt: time.Now().UTC(),
	})
}

// collectFeatures coleta as features dos últimos 30 dias para o projeto. BKL-1065.
func (h *ChurnPredictionHandler) collectFeatures(ctx context.Context, projectID string) (*churnFeatures, error) {
	f := &churnFeatures{}
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)

	// Contagem de eventos nos últimos 30 dias
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM events
		 WHERE project_id = $1::uuid AND created_at >= $2`,
		projectID, since,
	).Scan(&f.eventCount)

	// Contagem de sessões únicas nos últimos 30 dias (via session_id distinto)
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT session_id) FROM events
		 WHERE project_id = $1::uuid AND created_at >= $2 AND session_id IS NOT NULL AND session_id != ''`,
		projectID, since,
	).Scan(&f.sessionCount)

	// Dias desde o último evento
	var lastEvent *time.Time
	_ = h.db.QueryRow(ctx,
		`SELECT MAX(created_at) FROM events WHERE project_id = $1::uuid`,
		projectID,
	).Scan(&lastEvent)

	if lastEvent != nil {
		diff := time.Since(*lastEvent)
		f.daysSinceLastEvent = int(diff.Hours() / 24)
		if f.daysSinceLastEvent < 0 {
			f.daysSinceLastEvent = 0
		}
	} else {
		// Sem nenhum evento — assume 30+ dias para maximizar risco
		f.daysSinceLastEvent = 30
	}

	// Flags avaliadas nos últimos 30 dias
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT flag_key) FROM flag_evaluations
		 WHERE project_id = $1 AND evaluated_at >= $2`,
		projectID, since,
	).Scan(&f.flagsEvaluated)

	// Total de flags do projeto para calcular adoption rate
	var totalFlags int
	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM feature_flags WHERE project_id = $1::uuid`,
		projectID,
	).Scan(&totalFlags)

	if totalFlags > 0 {
		f.featureAdoptionRate = float64(f.flagsEvaluated) / float64(totalFlags)
		if f.featureAdoptionRate > 1.0 {
			f.featureAdoptionRate = 1.0
		}
	}

	return f, nil
}

// sigmoid calcula a função logística σ(z) = 1 / (1 + e^(-z)).
func sigmoid(z float64) float64 {
	return 1.0 / (1.0 + math.Exp(-z))
}
