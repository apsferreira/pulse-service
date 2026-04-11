package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthScoreRepository persiste e recupera health scores por tenant.
type HealthScoreRepository struct {
	db *pgxpool.Pool
}

func NewHealthScoreRepository(db *pgxpool.Pool) *HealthScoreRepository {
	return &HealthScoreRepository{db: db}
}

// GetByTenantID retorna o health score atual do tenant.
func (r *HealthScoreRepository) GetByTenantID(ctx context.Context, tenantID string) (*domain.TenantHealthScore, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, score, login_score, content_score, adoption_score, subscription_score,
		       last_login_at, last_content_at, calculated_at, created_at, updated_at
		FROM tenant_health_scores WHERE tenant_id = $1
	`, tenantID)

	return scanHealthScore(row)
}

// Upsert cria ou atualiza o health score do tenant.
func (r *HealthScoreRepository) Upsert(ctx context.Context, hs *domain.TenantHealthScore) error {
	hs.ID = uuid.New().String()
	now := time.Now().UTC()
	hs.CalculatedAt = now
	hs.UpdatedAt = now

	_, err := r.db.Exec(ctx, `
		INSERT INTO tenant_health_scores
		  (id, tenant_id, score, login_score, content_score, adoption_score, subscription_score,
		   last_login_at, last_content_at, calculated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $10)
		ON CONFLICT (tenant_id) DO UPDATE SET
		  score = EXCLUDED.score,
		  login_score = EXCLUDED.login_score,
		  content_score = EXCLUDED.content_score,
		  adoption_score = EXCLUDED.adoption_score,
		  subscription_score = EXCLUDED.subscription_score,
		  last_login_at = EXCLUDED.last_login_at,
		  last_content_at = EXCLUDED.last_content_at,
		  calculated_at = EXCLUDED.calculated_at,
		  updated_at = EXCLUDED.updated_at
	`,
		hs.ID, hs.TenantID, hs.Score, hs.LoginScore, hs.ContentScore,
		hs.AdoptionScore, hs.SubscriptionScore,
		hs.LastLoginAt, hs.LastContentAt, now,
	)
	return err
}

// ListAll retorna todos os health scores para o cron de cálculo.
func (r *HealthScoreRepository) ListDistinctTenants(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT properties->>'tenant_id' AS tenant_id
		FROM events
		WHERE properties->>'tenant_id' IS NOT NULL
		  AND created_at >= NOW() - INTERVAL '90 days'
	`)
	if err != nil {
		return nil, fmt.Errorf("health_score_repository: ListDistinctTenants: %w", err)
	}
	defer rows.Close()

	var tenants []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err == nil && tid != "" {
			tenants = append(tenants, tid)
		}
	}
	return tenants, nil
}

// CountEventsLast30Days conta eventos de um tipo para um tenant nos últimos 30 dias.
func (r *HealthScoreRepository) CountEventsLast30Days(ctx context.Context, tenantID string, eventNames ...string) (int, error) {
	if len(eventNames) == 0 {
		return 0, nil
	}

	// Build $2, $3, ... placeholders for event names
	args := make([]interface{}, len(eventNames)+2)
	args[0] = tenantID
	args[1] = time.Now().UTC().Add(-30 * 24 * time.Hour)
	placeholders := ""
	for i, name := range eventNames {
		args[i+2] = name
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+3)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM events
		WHERE properties->>'tenant_id' = $1
		  AND created_at >= $2
		  AND event_type IN (%s)
	`, placeholders)

	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("health_score_repository: CountEventsLast30Days: %w", err)
	}
	return count, nil
}

// LastEventAt retorna o timestamp do último evento de um tipo para um tenant.
func (r *HealthScoreRepository) LastEventAt(ctx context.Context, tenantID string, eventNames ...string) (*time.Time, error) {
	if len(eventNames) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(eventNames)+1)
	args[0] = tenantID
	placeholders := ""
	for i, name := range eventNames {
		args[i+1] = name
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		SELECT MAX(created_at) FROM events
		WHERE properties->>'tenant_id' = $1
		  AND event_type IN (%s)
	`, placeholders)

	var ts *time.Time
	err := r.db.QueryRow(ctx, query, args...).Scan(&ts)
	if err != nil {
		return nil, fmt.Errorf("health_score_repository: LastEventAt: %w", err)
	}
	return ts, nil
}

// ListAtRisk retorna health scores com score <= threshold, ordenados por score ASC.
// Usado pelo ChurnAlertScheduler (BKL-814) para alertar o admin IIT.
func (r *HealthScoreRepository) ListAtRisk(ctx context.Context, threshold int) ([]*domain.TenantHealthScore, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, score, login_score, content_score, adoption_score, subscription_score,
		       last_login_at, last_content_at, calculated_at, created_at, updated_at
		FROM tenant_health_scores
		WHERE score <= $1
		ORDER BY score ASC
	`, threshold)
	if err != nil {
		return nil, fmt.Errorf("health_score_repository: ListAtRisk: %w", err)
	}
	defer rows.Close()

	var results []*domain.TenantHealthScore
	for rows.Next() {
		hs, err := scanHealthScoreRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, hs)
	}
	return results, rows.Err()
}

// ListAll retorna todos os health scores ordenados por score ASC (BKL-1083: visão admin).
func (r *HealthScoreRepository) ListAll(ctx context.Context) ([]*domain.TenantHealthScore, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, score, login_score, content_score, adoption_score, subscription_score,
		       last_login_at, last_content_at, calculated_at, created_at, updated_at
		FROM tenant_health_scores
		ORDER BY score ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("health_score_repository: ListAll: %w", err)
	}
	defer rows.Close()

	var results []*domain.TenantHealthScore
	for rows.Next() {
		hs, err := scanHealthScoreRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, hs)
	}
	return results, rows.Err()
}

func scanHealthScoreRows(rows pgx.Rows) (*domain.TenantHealthScore, error) {
	hs := &domain.TenantHealthScore{}
	err := rows.Scan(
		&hs.ID, &hs.TenantID, &hs.Score,
		&hs.LoginScore, &hs.ContentScore, &hs.AdoptionScore, &hs.SubscriptionScore,
		&hs.LastLoginAt, &hs.LastContentAt,
		&hs.CalculatedAt, &hs.CreatedAt, &hs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("health_score_repository: scan rows: %w", err)
	}
	return hs, nil
}

func scanHealthScore(row pgx.Row) (*domain.TenantHealthScore, error) {
	hs := &domain.TenantHealthScore{}
	err := row.Scan(
		&hs.ID, &hs.TenantID, &hs.Score,
		&hs.LoginScore, &hs.ContentScore, &hs.AdoptionScore, &hs.SubscriptionScore,
		&hs.LastLoginAt, &hs.LastContentAt,
		&hs.CalculatedAt, &hs.CreatedAt, &hs.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("health_score_repository: scan: %w", err)
	}
	return hs, nil
}
