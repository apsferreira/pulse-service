package repository

// BKL-1044: repositório para flag_metrics — vinculação flag → evento de negócio.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlagMetricRepository struct {
	db *pgxpool.Pool
}

func NewFlagMetricRepository(db *pgxpool.Pool) *FlagMetricRepository {
	return &FlagMetricRepository{db: db}
}

// ListByFlag retorna todas as métricas vinculadas a uma flag dentro de um projeto.
func (r *FlagMetricRepository) ListByFlag(ctx context.Context, flagKey, projectID string) ([]domain.FlagMetric, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, flag_key, project_id, metric_name, COALESCE(description,''), created_at, updated_at
		 FROM flag_metrics
		 WHERE flag_key = $1 AND project_id = $2::uuid
		 ORDER BY metric_name ASC`,
		flagKey, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("flag_metric_repository: ListByFlag: %w", err)
	}
	defer rows.Close()

	var metrics []domain.FlagMetric
	for rows.Next() {
		m, err := scanFlagMetric(rows)
		if err != nil {
			return nil, fmt.Errorf("flag_metric_repository: ListByFlag scan: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// Create vincula um evento de negócio a uma flag (cria um FlagMetric).
func (r *FlagMetricRepository) Create(ctx context.Context, flagKey, projectID string, req domain.CreateFlagMetricRequest) (*domain.FlagMetric, error) {
	m := domain.FlagMetric{
		ID:          uuid.New().String(),
		FlagKey:     flagKey,
		ProjectID:   projectID,
		MetricName:  req.MetricName,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO flag_metrics (id, flag_key, project_id, metric_name, description, created_at, updated_at)
		 VALUES ($1, $2, $3::uuid, $4, $5, $6, $7)
		 ON CONFLICT (flag_key, project_id, metric_name) DO NOTHING`,
		m.ID, m.FlagKey, m.ProjectID, m.MetricName, m.Description, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("flag_metric_repository: Create: %w", err)
	}
	return &m, nil
}

// Delete remove a vinculação flag → métrica dentro de um projeto.
func (r *FlagMetricRepository) Delete(ctx context.Context, flagKey, projectID, metricName string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM flag_metrics WHERE flag_key = $1 AND project_id = $2::uuid AND metric_name = $3`,
		flagKey, projectID, metricName,
	)
	if err != nil {
		return fmt.Errorf("flag_metric_repository: Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("flag_metric_repository: Delete: métrica não encontrada")
	}
	return nil
}

// GetLift calcula o lift de uma flag sobre um evento de negócio.
// JOIN entre flag_evaluations (quem viu a flag) e events (quem converteu).
// BKL-1044: lift = (treatment_rate - control_rate) / control_rate * 100
func (r *FlagMetricRepository) GetLift(ctx context.Context, flagKey, projectID, metricName string) (*domain.FlagMetricLift, error) {
	// Verifica que a vinculação existe
	row := r.db.QueryRow(ctx,
		`SELECT id FROM flag_metrics WHERE flag_key = $1 AND project_id = $2::uuid AND metric_name = $3`,
		flagKey, projectID, metricName,
	)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // vinculação não encontrada
		}
		return nil, fmt.Errorf("flag_metric_repository: GetLift check: %w", err)
	}

	// Conta exposições (flag_evaluations) por grupo controle/tratamento
	liftRow := r.db.QueryRow(ctx, `
		WITH evaluations AS (
			SELECT
				user_id,
				result AS treated
			FROM flag_evaluations
			WHERE flag_key = $1
			  AND project_id = $2::uuid
			  AND user_id IS NOT NULL
		),
		conversions AS (
			SELECT DISTINCT user_id
			FROM events
			WHERE project_id = $2::uuid
			  AND event_type = $3
			  AND user_id IS NOT NULL
		),
		grouped AS (
			SELECT
				e.treated,
				COUNT(DISTINCT e.user_id)                                          AS exposed,
				COUNT(DISTINCT c.user_id)                                          AS converted
			FROM evaluations e
			LEFT JOIN conversions c ON c.user_id = e.user_id
			GROUP BY e.treated
		)
		SELECT
			COALESCE(SUM(CASE WHEN NOT treated THEN exposed   END), 0) AS control_count,
			COALESCE(SUM(CASE WHEN treated      THEN exposed   END), 0) AS treatment_count,
			COALESCE(SUM(CASE WHEN NOT treated THEN converted END), 0) AS control_conv,
			COALESCE(SUM(CASE WHEN treated      THEN converted END), 0) AS treatment_conv
		FROM grouped
	`, flagKey, projectID, metricName)

	lift := &domain.FlagMetricLift{
		FlagKey:    flagKey,
		MetricName: metricName,
	}
	if err := liftRow.Scan(&lift.ControlCount, &lift.TreatmentCount, &lift.ControlConv, &lift.TreatmentConv); err != nil {
		return nil, fmt.Errorf("flag_metric_repository: GetLift query: %w", err)
	}

	// Calcula taxas e lift relativo
	if lift.ControlCount > 0 {
		lift.ControlRate = float64(lift.ControlConv) / float64(lift.ControlCount)
	}
	if lift.TreatmentCount > 0 {
		lift.TreatmentRate = float64(lift.TreatmentConv) / float64(lift.TreatmentCount)
	}
	if lift.ControlRate > 0 {
		lift.LiftPct = (lift.TreatmentRate - lift.ControlRate) / lift.ControlRate * 100
	}

	return lift, nil
}

func scanFlagMetric(rows pgx.Rows) (domain.FlagMetric, error) {
	var m domain.FlagMetric
	err := rows.Scan(&m.ID, &m.FlagKey, &m.ProjectID, &m.MetricName, &m.Description, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}
