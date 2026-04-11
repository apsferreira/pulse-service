// BKL-864: Serviço de geração de insights narrativos semanais via Claude Haiku.
// Agrega métricas de retenção, CSAT e custo de IA da semana passada,
// chama o pkg/insight para gerar texto acionável e persiste no banco.
// Scheduler: toda segunda-feira às 08h (ou no startup se ainda não gerado esta semana).
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/institutoitinerante/pulse-service/internal/pkg/insight"
)

// WeeklyInsightService orquestra a geração do insight semanal.
type WeeklyInsightService struct {
	db        *pgxpool.Pool
	generator *insight.Generator
}

// NewWeeklyInsightService cria o serviço. generator pode ser nil (sem API key).
func NewWeeklyInsightService(db *pgxpool.Pool) *WeeklyInsightService {
	return &WeeklyInsightService{
		db:        db,
		generator: insight.New(),
	}
}

// Generate busca as métricas da semana passada, gera o insight e persiste.
// Idempotente: se já existe insight para a semana, ignora silenciosamente.
func (s *WeeklyInsightService) Generate(ctx context.Context) error {
	if s.generator == nil {
		log.Println("[weekly-insight] gerador não disponível (sem ANTHROPIC_API_KEY)")
		return nil
	}

	now := time.Now().UTC()
	weekStart := startOfLastWeek(now)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Idempotência: verificar se já gerado para esta semana
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM weekly_insights WHERE week_start = $1`,
		weekStart.Format("2006-01-02"),
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("weekly_insight: check existing: %w", err)
	}
	if count > 0 {
		log.Printf("[weekly-insight] insight já gerado para semana %s", weekStart.Format("2006-01-02"))
		return nil
	}

	// Coletar métricas da semana passada
	m, err := s.collectMetrics(ctx, weekStart, weekEnd)
	if err != nil {
		return fmt.Errorf("weekly_insight: collectMetrics: %w", err)
	}
	m.WeekStart = weekStart.Format("2006-01-02")
	m.WeekEnd = weekEnd.Format("2006-01-02")

	// Gerar narrativa
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	narrative, err := s.generator.Generate(callCtx, m)
	if err != nil {
		return fmt.Errorf("weekly_insight: generate: %w", err)
	}
	if narrative == "" {
		log.Println("[weekly-insight] narrativa vazia, pulando persistência")
		return nil
	}

	// Persistir
	metricsJSON, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("weekly_insight: marshal metrics: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO weekly_insights (week_start, week_end, narrative, metrics)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (week_start) DO UPDATE
		   SET narrative = EXCLUDED.narrative,
		       metrics   = EXCLUDED.metrics`,
		weekStart.Format("2006-01-02"),
		weekEnd.Format("2006-01-02"),
		narrative,
		metricsJSON,
	)
	if err != nil {
		return fmt.Errorf("weekly_insight: persist: %w", err)
	}

	log.Printf("[weekly-insight] insight gerado para semana %s: %q", m.WeekStart, truncate(narrative, 80))
	return nil
}

// GetLatest retorna o insight mais recente disponível.
func (s *WeeklyInsightService) GetLatest(ctx context.Context) (*insight.WeeklyInsightReport, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id, week_start, week_end, narrative, metrics, created_at
		 FROM weekly_insights
		 ORDER BY week_start DESC
		 LIMIT 1`,
	)

	var r insight.WeeklyInsightReport
	var metricsJSON []byte
	var weekStart, weekEnd string

	err := row.Scan(&r.ID, &weekStart, &weekEnd, &r.Narrative, &metricsJSON, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("weekly_insight: GetLatest: %w", err)
	}

	r.WeekStart, _ = time.Parse("2006-01-02", weekStart)
	r.WeekEnd, _ = time.Parse("2006-01-02", weekEnd)

	if err := json.Unmarshal(metricsJSON, &r.Metrics); err != nil {
		log.Printf("[weekly-insight] aviso: falha ao deserializar métricas: %v", err)
	}

	return &r, nil
}

// collectMetrics agrega as métricas relevantes da semana informada.
func (s *WeeklyInsightService) collectMetrics(ctx context.Context, weekStart, weekEnd time.Time) (insight.MetricsSnapshot, error) {
	m := insight.MetricsSnapshot{}
	prevStart := weekStart.AddDate(0, 0, -7)
	prevEnd := weekEnd.AddDate(0, 0, -7)

	// DAU médio da semana atual
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(active_count), 0)
		 FROM (
		   SELECT DATE(tracked_at) as d, COUNT(DISTINCT user_id) as active_count
		   FROM events
		   WHERE tracked_at >= $1 AND tracked_at < $2
		   GROUP BY d
		 ) t`,
		weekStart, weekEnd.AddDate(0, 0, 1),
	).Scan(&m.DAU)

	// DAU semana anterior
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(active_count), 0)
		 FROM (
		   SELECT DATE(tracked_at) as d, COUNT(DISTINCT user_id) as active_count
		   FROM events
		   WHERE tracked_at >= $1 AND tracked_at < $2
		   GROUP BY d
		 ) t`,
		prevStart, prevEnd.AddDate(0, 0, 1),
	).Scan(&m.PrevDAU)

	// CSAT médio da semana atual
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(score), 0)
		 FROM csat_responses
		 WHERE created_at >= $1 AND created_at < $2`,
		weekStart, weekEnd.AddDate(0, 0, 1),
	).Scan(&m.CSATScore)

	// CSAT semana anterior
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(score), 0)
		 FROM csat_responses
		 WHERE created_at >= $1 AND created_at < $2`,
		prevStart, prevEnd.AddDate(0, 0, 1),
	).Scan(&m.PrevCSATScore)

	// NPS semana atual (média das respostas NPS, convertida para escala -100..100)
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(
		   (COUNT(*) FILTER (WHERE score >= 9) - COUNT(*) FILTER (WHERE score <= 6))::numeric * 100
		   / NULLIF(COUNT(*), 0), 0
		 )
		 FROM csat_responses
		 WHERE survey_id IS NOT NULL
		   AND created_at >= $1 AND created_at < $2`,
		weekStart, weekEnd.AddDate(0, 0, 1),
	).Scan(&m.NPSScore)

	// Custo total IA em USD na semana (convertido para BRL 1 USD ≈ R$5.70)
	var costUSD float64
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(cost_usd), 0)
		 FROM api_costs
		 WHERE created_at >= $1 AND created_at < $2`,
		weekStart, weekEnd.AddDate(0, 0, 1),
	).Scan(&costUSD)
	m.APICostBRL = costUSD * 5.70

	var prevCostUSD float64
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(cost_usd), 0)
		 FROM api_costs
		 WHERE created_at >= $1 AND created_at < $2`,
		prevStart, prevEnd.AddDate(0, 0, 1),
	).Scan(&prevCostUSD)
	m.PrevAPICostBRL = prevCostUSD * 5.70

	// Churn estimado: proporção de usuários que tiveram eventos na semana anterior mas não nesta
	// Simplificado: (PrevDAU - DAU) / PrevDAU se positivo
	if m.PrevDAU > 0 && m.PrevDAU > m.DAU {
		m.ChurnRate = float64(m.PrevDAU-m.DAU) / float64(m.PrevDAU) * 100
	}
	// Churn semana anterior vs 2 semanas atrás (aproximação)
	var twoWeeksDAU int
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(AVG(active_count), 0)
		 FROM (
		   SELECT DATE(tracked_at) as d, COUNT(DISTINCT user_id) as active_count
		   FROM events
		   WHERE tracked_at >= $1 AND tracked_at < $2
		   GROUP BY d
		 ) t`,
		prevStart.AddDate(0, 0, -7), prevStart,
	).Scan(&twoWeeksDAU)
	if twoWeeksDAU > 0 && twoWeeksDAU > m.PrevDAU {
		m.PrevChurnRate = float64(twoWeeksDAU-m.PrevDAU) / float64(twoWeeksDAU) * 100
	}

	// Churn por produto
	rows, err := s.db.Query(ctx,
		`SELECT product, COUNT(DISTINCT user_id) as cnt
		 FROM events
		 WHERE tracked_at >= $1 AND tracked_at < $2
		   AND product IS NOT NULL
		 GROUP BY product`,
		weekStart, weekEnd.AddDate(0, 0, 1),
	)
	if err == nil {
		defer rows.Close()
		prevRows, _ := s.db.Query(ctx,
			`SELECT product, COUNT(DISTINCT user_id) as cnt
			 FROM events
			 WHERE tracked_at >= $1 AND tracked_at < $2
			   AND product IS NOT NULL
			 GROUP BY product`,
			prevStart, prevEnd.AddDate(0, 0, 1),
		)

		currByProduct := map[string]int{}
		for rows.Next() {
			var prod string
			var cnt int
			if err := rows.Scan(&prod, &cnt); err == nil {
				currByProduct[prod] = cnt
			}
		}

		prevByProduct := map[string]int{}
		if prevRows != nil {
			defer prevRows.Close()
			for prevRows.Next() {
				var prod string
				var cnt int
				if err := prevRows.Scan(&prod, &cnt); err == nil {
					prevByProduct[prod] = cnt
				}
			}
		}

		m.ProductChurn = map[string]float64{}
		for prod, prev := range prevByProduct {
			curr := currByProduct[prod]
			if prev > 0 && prev > curr {
				m.ProductChurn[prod] = float64(prev-curr) / float64(prev) * 100
			}
		}
	}

	return m, nil
}

// startOfLastWeek retorna a segunda-feira da semana passada (UTC).
func startOfLastWeek(now time.Time) time.Time {
	// Dia da semana: 0=Sunday, 1=Monday ... 6=Saturday
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7 // treat Sunday as 7
	}
	// Dias desde a última segunda-feira: wd-1 dias atrás é esta semana, mais 7 para a semana passada
	daysBack := wd - 1 + 7
	lastMonday := now.AddDate(0, 0, -daysBack)
	return time.Date(lastMonday.Year(), lastMonday.Month(), lastMonday.Day(), 0, 0, 0, 0, time.UTC)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
