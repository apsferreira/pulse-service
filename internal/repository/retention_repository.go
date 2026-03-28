package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetentionRepository struct {
	db *pgxpool.Pool
}

func NewRetentionRepository(db *pgxpool.Pool) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// GetRetention returns DAU/WAU/MAU for each date in the range, optionally filtered by product.
func (r *RetentionRepository) GetRetention(ctx context.Context, product string, startDate, endDate time.Time) ([]domain.ActiveUsersItem, error) {
	query := `
		WITH date_series AS (
			SELECT generate_series($1::date, $2::date, '1 day'::interval)::date AS d
		)
		SELECT
			ds.d::text AS date,
			COUNT(DISTINCT CASE WHEN e.created_at::date = ds.d THEN e.user_id END) AS dau,
			COUNT(DISTINCT CASE WHEN e.created_at::date BETWEEN ds.d - 6 AND ds.d THEN e.user_id END) AS wau,
			COUNT(DISTINCT CASE WHEN e.created_at::date BETWEEN ds.d - 29 AND ds.d THEN e.user_id END) AS mau
		FROM date_series ds
		LEFT JOIN events e ON e.created_at::date BETWEEN ds.d - 29 AND ds.d
			AND e.user_id IS NOT NULL`

	args := []interface{}{startDate, endDate}
	argIdx := 3

	if product != "" {
		query += fmt.Sprintf(" AND e.product = $%d", argIdx)
		args = append(args, product)
		argIdx++
	}

	query += `
		GROUP BY ds.d
		ORDER BY ds.d ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("retention_repository: GetRetention: %w", err)
	}
	defer rows.Close()

	var items []domain.ActiveUsersItem
	for rows.Next() {
		var item domain.ActiveUsersItem
		if err := rows.Scan(&item.Date, &item.DAU, &item.WAU, &item.MAU); err != nil {
			return nil, fmt.Errorf("retention_repository: GetRetention scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("retention_repository: GetRetention rows: %w", err)
	}
	return items, nil
}

// GetCohorts returns cohort analysis grouped by signup week or month.
func (r *RetentionRepository) GetCohorts(ctx context.Context, product, granularity string) ([]domain.CohortItem, error) {
	truncFunc := "week"
	if granularity == "month" {
		truncFunc = "month"
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "user_id IS NOT NULL")
	if product != "" {
		conditions = append(conditions, fmt.Sprintf("product = $%d", argIdx))
		args = append(args, product)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		WITH first_seen AS (
			SELECT
				user_id,
				date_trunc('%s', MIN(created_at)) AS cohort
			FROM events
			%s
			GROUP BY user_id
		),
		activity AS (
			SELECT
				fs.cohort,
				date_trunc('%s', e.created_at) AS activity_period,
				COUNT(DISTINCT e.user_id) AS active_users
			FROM events e
			JOIN first_seen fs ON e.user_id = fs.user_id
			%s
			GROUP BY fs.cohort, date_trunc('%s', e.created_at)
		),
		cohort_sizes AS (
			SELECT cohort, COUNT(*) AS size
			FROM first_seen
			GROUP BY cohort
		)
		SELECT
			cs.cohort::date::text AS cohort_label,
			cs.size AS cohort_size,
			EXTRACT(EPOCH FROM (a.activity_period - cs.cohort)) /
				CASE '%s' WHEN 'week' THEN 604800 ELSE 2592000 END AS period_offset,
			a.active_users
		FROM cohort_sizes cs
		JOIN activity a ON a.cohort = cs.cohort
		ORDER BY cs.cohort ASC, period_offset ASC`,
		truncFunc, whereClause, truncFunc, whereClause, truncFunc, truncFunc,
	)

	// args are duplicated for both CTEs
	fullArgs := append(args, args...)

	rows, err := r.db.Query(ctx, query, fullArgs...)
	if err != nil {
		return nil, fmt.Errorf("retention_repository: GetCohorts: %w", err)
	}
	defer rows.Close()

	// Build cohort map
	type cohortKey struct {
		Label string
		Size  int
	}
	cohortMap := make(map[string]*domain.CohortItem)
	var order []string

	for rows.Next() {
		var label string
		var size int
		var periodOffset float64
		var activeUsers int

		if err := rows.Scan(&label, &size, &periodOffset, &activeUsers); err != nil {
			return nil, fmt.Errorf("retention_repository: GetCohorts scan: %w", err)
		}

		period := int(periodOffset)
		if period < 0 {
			continue
		}

		ci, ok := cohortMap[label]
		if !ok {
			ci = &domain.CohortItem{
				CohortLabel: label,
				CohortSize:  size,
			}
			cohortMap[label] = ci
			order = append(order, label)
		}

		retentionPct := 0.0
		if ci.CohortSize > 0 {
			retentionPct = float64(activeUsers) / float64(ci.CohortSize) * 100
		}

		ci.Buckets = append(ci.Buckets, domain.CohortBucket{
			Period:       period,
			ActiveUsers:  activeUsers,
			RetentionPct: retentionPct,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("retention_repository: GetCohorts rows: %w", err)
	}

	var result []domain.CohortItem
	for _, label := range order {
		result = append(result, *cohortMap[label])
	}
	return result, nil
}

// GetActiveUsers returns distinct user counts per day for DAU, per week for WAU, per month for MAU.
func (r *RetentionRepository) GetActiveUsers(ctx context.Context, product string, startDate, endDate time.Time) (dau, wau, mau []domain.ActiveUsersCountItem, err error) {
	var conditions []string
	var args []interface{}
	argIdx := 3 // $1=start, $2=end

	conditions = append(conditions, "user_id IS NOT NULL")
	conditions = append(conditions, "created_at >= $1")
	conditions = append(conditions, "created_at <= $2")
	args = append(args, startDate, endDate)

	if product != "" {
		conditions = append(conditions, fmt.Sprintf("product = $%d", argIdx))
		args = append(args, product)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// DAU — distinct users per day
	dauQuery := fmt.Sprintf(`
		SELECT created_at::date::text AS period, COUNT(DISTINCT user_id) AS cnt
		FROM events %s
		GROUP BY created_at::date ORDER BY created_at::date ASC`, whereClause)

	// WAU — distinct users per ISO week
	wauQuery := fmt.Sprintf(`
		SELECT date_trunc('week', created_at)::date::text AS period, COUNT(DISTINCT user_id) AS cnt
		FROM events %s
		GROUP BY date_trunc('week', created_at) ORDER BY date_trunc('week', created_at) ASC`, whereClause)

	// MAU — distinct users per month
	mauQuery := fmt.Sprintf(`
		SELECT date_trunc('month', created_at)::date::text AS period, COUNT(DISTINCT user_id) AS cnt
		FROM events %s
		GROUP BY date_trunc('month', created_at) ORDER BY date_trunc('month', created_at) ASC`, whereClause)

	// Execute all three
	dau, err = r.queryCountItems(ctx, dauQuery, args)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("retention_repository: GetActiveUsers DAU: %w", err)
	}

	wau, err = r.queryCountItems(ctx, wauQuery, args)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("retention_repository: GetActiveUsers WAU: %w", err)
	}

	mau, err = r.queryCountItems(ctx, mauQuery, args)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("retention_repository: GetActiveUsers MAU: %w", err)
	}

	return dau, wau, mau, nil
}

func (r *RetentionRepository) queryCountItems(ctx context.Context, query string, args []interface{}) ([]domain.ActiveUsersCountItem, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ActiveUsersCountItem
	for rows.Next() {
		var item domain.ActiveUsersCountItem
		if err := rows.Scan(&item.Period, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
