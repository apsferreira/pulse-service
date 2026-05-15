package repository

// BKL-insights: Repositório de insights acionáveis gerados pelo agente autônomo.
// Todas as queries usam prepared statements com $N — R1 obrigatório.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectInsightRepository persiste e consulta insights gerados pelo agente.
type ProjectInsightRepository struct {
	db *pgxpool.Pool
}

// NewProjectInsightRepository cria o repositório.
func NewProjectInsightRepository(db *pgxpool.Pool) *ProjectInsightRepository {
	return &ProjectInsightRepository{db: db}
}

// Insert persiste um novo insight. Retorna erro se já existir insight do mesmo
// tipo/projeto/flag criado nas últimas 24h (deduplicação).
func (r *ProjectInsightRepository) Insert(ctx context.Context, ins *domain.ProjectInsight) error {
	// Deduplicação: não inserir se já existe insight do mesmo tipo para o mesmo
	// projeto/flag nas últimas 24h.
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM project_insights
			WHERE project_id   = $1::uuid
			  AND insight_type = $2
			  AND ($3::varchar IS NULL OR flag_key = $3)
			  AND created_at  >= NOW() - INTERVAL '24 hours'
			  AND is_dismissed = false
		)
	`, ins.ProjectID, string(ins.InsightType), ins.FlagKey).Scan(&exists)
	if err != nil {
		return fmt.Errorf("project_insight: dedup check: %w", err)
	}
	if exists {
		return nil // insight duplicado — ignorar silenciosamente
	}

	metaJSON, err := json.Marshal(ins.Metadata)
	if err != nil {
		return fmt.Errorf("project_insight: marshal metadata: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO project_insights
			(project_id, insight_type, flag_key, message, severity, metadata, expires_at)
		VALUES
			($1::uuid, $2, $3, $4, $5, $6, $7)
	`,
		ins.ProjectID,
		string(ins.InsightType),
		ins.FlagKey,
		ins.Message,
		string(ins.Severity),
		metaJSON,
		ins.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("project_insight: Insert: %w", err)
	}
	return nil
}

// List retorna insights de um projeto. dismissed=false → apenas ativos/não expirados.
func (r *ProjectInsightRepository) List(ctx context.Context, projectID string, dismissed bool) ([]domain.ProjectInsight, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, insight_type, flag_key, message, severity,
		       metadata, is_dismissed, dismissed_at, dismissed_by, created_at, expires_at
		FROM project_insights
		WHERE project_id   = $1::uuid
		  AND is_dismissed = $2
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY
		  CASE severity
		    WHEN 'critical' THEN 1
		    WHEN 'warning'  THEN 2
		    ELSE 3
		  END,
		  created_at DESC
	`, projectID, dismissed)
	if err != nil {
		return nil, fmt.Errorf("project_insight: List: %w", err)
	}
	defer rows.Close()

	var items []domain.ProjectInsight
	for rows.Next() {
		ins, err := scanInsight(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, ins)
	}
	if items == nil {
		items = []domain.ProjectInsight{}
	}
	return items, rows.Err()
}

// Dismiss marca um insight como dispensado.
func (r *ProjectInsightRepository) Dismiss(ctx context.Context, id, dismissedBy string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE project_insights
		SET is_dismissed = true,
		    dismissed_at = NOW(),
		    dismissed_by = $2
		WHERE id = $1::uuid
		  AND is_dismissed = false
	`, id, dismissedBy)
	if err != nil {
		return fmt.Errorf("project_insight: Dismiss: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("project_insight: insight não encontrado ou já dispensado")
	}
	return nil
}

// ActiveProjects retorna IDs e slugs de todos os projetos ativos.
func (r *ProjectInsightRepository) ActiveProjects(ctx context.Context) ([]projectRef, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug FROM projects WHERE is_active = true
	`)
	if err != nil {
		return nil, fmt.Errorf("project_insight: ActiveProjects: %w", err)
	}
	defer rows.Close()

	var refs []projectRef
	for rows.Next() {
		var p projectRef
		if err := rows.Scan(&p.ID, &p.Slug); err != nil {
			return nil, err
		}
		refs = append(refs, p)
	}
	return refs, rows.Err()
}

// projectRef é um par id/slug interno ao repositório.
type projectRef struct {
	ID   string
	Slug string
}

// --- Queries analíticas para o agente ---

// DormantFlags retorna flags ativas sem avaliações nos últimos 30 dias.
func (r *ProjectInsightRepository) DormantFlags(ctx context.Context, projectID string) ([]flagRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.key
		FROM feature_flags f
		WHERE f.project_id = $1::uuid
		  AND f.enabled    = true
		  AND NOT EXISTS (
		      SELECT 1 FROM flag_evaluations fe
		      WHERE fe.flag_key   = f.key
		        AND fe.project_id = $1::uuid
		        AND fe.evaluated_at >= NOW() - INTERVAL '30 days'
		  )
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_insight: DormantFlags: %w", err)
	}
	defer rows.Close()
	return collectFlagRows(rows)
}

// FullRolloutFlags retorna flags com rollout=100 e enabled=true por mais de 14 dias.
func (r *ProjectInsightRepository) FullRolloutFlags(ctx context.Context, projectID string) ([]flagRowWithDays, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.key,
		       EXTRACT(DAY FROM NOW() - f.updated_at)::int AS days_active
		FROM feature_flags f
		WHERE f.project_id         = $1::uuid
		  AND f.enabled            = true
		  AND f.rollout_percentage = 100
		  AND f.updated_at        <= NOW() - INTERVAL '14 days'
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_insight: FullRolloutFlags: %w", err)
	}
	defer rows.Close()

	var out []flagRowWithDays
	for rows.Next() {
		var r flagRowWithDays
		if err := rows.Scan(&r.Key, &r.Days); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// EventVolumeStats retorna volume das últimas 24h e média diária dos últimos 7 dias para o projeto.
func (r *ProjectInsightRepository) EventVolumeStats(ctx context.Context, projectID string) (last24h int64, avg7d float64, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
		    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours') AS last_24h,
		    COALESCE(
		        COUNT(*) FILTER (
		            WHERE created_at >= NOW() - INTERVAL '7 days'
		              AND created_at <  NOW() - INTERVAL '24 hours'
		        )::numeric / 6.0,
		        0
		    ) AS avg_6_prior_days
		FROM events
		WHERE project_id = $1::uuid
		  AND created_at >= NOW() - INTERVAL '7 days'
	`, projectID).Scan(&last24h, &avg7d)
	if err != nil {
		return 0, 0, fmt.Errorf("project_insight: EventVolumeStats: %w", err)
	}
	return last24h, avg7d, nil
}

// NegativeRateFlags retorna flags onde >90% das avaliações retornaram false nos últimos 7 dias.
func (r *ProjectInsightRepository) NegativeRateFlags(ctx context.Context, projectID string) ([]flagRowWithPct, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
		    flag_key,
		    ROUND(
		        100.0 * COUNT(*) FILTER (WHERE result = false)
		        / NULLIF(COUNT(*), 0),
		        1
		    )::numeric AS negative_pct
		FROM flag_evaluations
		WHERE project_id  = $1::uuid
		  AND evaluated_at >= NOW() - INTERVAL '7 days'
		GROUP BY flag_key
		HAVING COUNT(*) >= 10   -- mínimo 10 avaliações para ser relevante
		   AND (
		       COUNT(*) FILTER (WHERE result = false)::numeric / NULLIF(COUNT(*), 0)
		   ) > 0.9
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_insight: NegativeRateFlags: %w", err)
	}
	defer rows.Close()

	var out []flagRowWithPct
	for rows.Next() {
		var r flagRowWithPct
		if err := rows.Scan(&r.Key, &r.Pct); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PartialRolloutFlags retorna flags com rollout 10-90%, ativas há mais de 7 dias,
// sem variantes A/B configuradas (variants IS NULL ou vazio).
func (r *ProjectInsightRepository) PartialRolloutFlags(ctx context.Context, projectID string) ([]flagRowWithPct, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.key,
		       f.rollout_percentage::numeric AS pct
		FROM feature_flags f
		WHERE f.project_id         = $1::uuid
		  AND f.enabled            = true
		  AND f.rollout_percentage BETWEEN 10 AND 90
		  AND f.updated_at        <= NOW() - INTERVAL '7 days'
		  AND (f.variants IS NULL OR f.variants::text = '[]' OR f.variants::text = 'null')
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_insight: PartialRolloutFlags: %w", err)
	}
	defer rows.Close()

	var out []flagRowWithPct
	for rows.Next() {
		var r flagRowWithPct
		if err := rows.Scan(&r.Key, &r.Pct); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- tipos auxiliares internos ---

type flagRow struct {
	Key string
}

type flagRowWithDays struct {
	Key  string
	Days int
}

type flagRowWithPct struct {
	Key string
	Pct float64
}

func collectFlagRows(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]flagRow, error) {
	var out []flagRow
	for rows.Next() {
		var r flagRow
		if err := rows.Scan(&r.Key); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// scanInsight escaneia uma linha completa de project_insights.
func scanInsight(rows interface {
	Scan(dest ...any) error
}) (domain.ProjectInsight, error) {
	var ins domain.ProjectInsight
	var metaRaw []byte
	var insightType string
	var severity string

	err := rows.Scan(
		&ins.ID,
		&ins.ProjectID,
		&insightType,
		&ins.FlagKey,
		&ins.Message,
		&severity,
		&metaRaw,
		&ins.IsDismissed,
		&ins.DismissedAt,
		&ins.DismissedBy,
		&ins.CreatedAt,
		&ins.ExpiresAt,
	)
	if err != nil {
		return ins, fmt.Errorf("project_insight: scan: %w", err)
	}

	ins.InsightType = domain.InsightType(insightType)
	ins.Severity = domain.InsightSeverity(severity)

	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &ins.Metadata); err != nil {
			// metadata corrompido não deve derrubar a listagem
			ins.Metadata = nil
		}
	}

	return ins, nil
}

// PurgeExpired deleta insights expirados. Chamado opcionalmente pelo agente.
func (r *ProjectInsightRepository) PurgeExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM project_insights
		WHERE expires_at IS NOT NULL
		  AND expires_at < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("project_insight: PurgeExpired: %w", err)
	}
	return tag.RowsAffected(), nil
}

// InsightsSince retorna a contagem de insights por tipo gerados após `since` para um projeto.
// Usado em testes e no endpoint de trigger manual para feedback.
func (r *ProjectInsightRepository) CountActive(ctx context.Context, projectID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM project_insights
		WHERE project_id   = $1::uuid
		  AND is_dismissed = false
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, projectID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("project_insight: CountActive: %w", err)
	}
	return n, nil
}

// InsertedAt é o tempo de escrita de um insight — usado no agente para logs.
func insightExpiresAt() *time.Time {
	t := time.Now().UTC().Add(7 * 24 * time.Hour)
	return &t
}
