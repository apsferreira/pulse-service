package repository

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

type MonitorRepository struct {
	db *pgxpool.Pool
}

func NewMonitorRepository(db *pgxpool.Pool) *MonitorRepository {
	return &MonitorRepository{db: db}
}

// GetAll returns all monitors.
func (r *MonitorRepository) GetAll(ctx context.Context) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, url, method, interval_seconds, timeout_seconds, is_active, created_at, updated_at
		 FROM monitors
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor_repository: GetAll: %w", err)
	}
	defer rows.Close()

	var monitors []domain.Monitor
	for rows.Next() {
		var m domain.Monitor
		if err := rows.Scan(&m.ID, &m.Name, &m.URL, &m.Method, &m.IntervalSeconds, &m.TimeoutSeconds, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("monitor_repository: GetAll scan: %w", err)
		}
		monitors = append(monitors, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("monitor_repository: GetAll rows: %w", err)
	}
	return monitors, nil
}

// GetAllActive returns only active monitors.
func (r *MonitorRepository) GetAllActive(ctx context.Context) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, url, method, interval_seconds, timeout_seconds, is_active, created_at, updated_at
		 FROM monitors
		 WHERE is_active = true
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor_repository: GetAllActive: %w", err)
	}
	defer rows.Close()

	var monitors []domain.Monitor
	for rows.Next() {
		var m domain.Monitor
		if err := rows.Scan(&m.ID, &m.Name, &m.URL, &m.Method, &m.IntervalSeconds, &m.TimeoutSeconds, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("monitor_repository: GetAllActive scan: %w", err)
		}
		monitors = append(monitors, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("monitor_repository: GetAllActive rows: %w", err)
	}
	return monitors, nil
}

// GetByID returns a single monitor by UUID.
func (r *MonitorRepository) GetByID(ctx context.Context, id string) (*domain.Monitor, error) {
	var m domain.Monitor
	err := r.db.QueryRow(ctx,
		`SELECT id, name, url, method, interval_seconds, timeout_seconds, is_active, created_at, updated_at
		 FROM monitors WHERE id = $1`,
		id,
	).Scan(&m.ID, &m.Name, &m.URL, &m.Method, &m.IntervalSeconds, &m.TimeoutSeconds, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("monitor_repository: GetByID: %w", err)
	}
	return &m, nil
}

// Create inserts a new monitor.
func (r *MonitorRepository) Create(ctx context.Context, m *domain.Monitor) error {
	m.ID = uuid.New().String()
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO monitors (id, name, url, method, interval_seconds, timeout_seconds, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.Name, m.URL, m.Method, m.IntervalSeconds, m.TimeoutSeconds, m.IsActive, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("monitor_repository: Create: %w", err)
	}
	return nil
}

// Update applies partial updates to a monitor.
func (r *MonitorRepository) Update(ctx context.Context, id string, req domain.UpdateMonitorRequest) (*domain.Monitor, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.URL != nil {
		existing.URL = *req.URL
	}
	if req.Method != nil {
		existing.Method = *req.Method
	}
	if req.IntervalSeconds != nil {
		existing.IntervalSeconds = *req.IntervalSeconds
	}
	if req.TimeoutSeconds != nil {
		existing.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.Exec(ctx,
		`UPDATE monitors
		 SET name = $1, url = $2, method = $3, interval_seconds = $4, timeout_seconds = $5, is_active = $6, updated_at = $7
		 WHERE id = $8`,
		existing.Name, existing.URL, existing.Method, existing.IntervalSeconds, existing.TimeoutSeconds, existing.IsActive, existing.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor_repository: Update: %w", err)
	}
	return existing, nil
}

// Delete removes a monitor by ID. Cascades to monitor_checks.
func (r *MonitorRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("monitor_repository: Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("monitor_repository: Delete: monitor not found")
	}
	return nil
}

// InsertCheck persists the result of a monitor check.
func (r *MonitorRepository) InsertCheck(ctx context.Context, check *domain.MonitorCheck) error {
	check.ID = uuid.New().String()
	check.CheckedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO monitor_checks (id, monitor_id, status, status_code, response_time_ms, error_message, checked_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		check.ID, check.MonitorID, check.Status, check.StatusCode, check.ResponseTimeMs, check.ErrorMessage, check.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("monitor_repository: InsertCheck: %w", err)
	}
	return nil
}

// GetLastCheck returns the most recent check for a monitor.
func (r *MonitorRepository) GetLastCheck(ctx context.Context, monitorID string) (*domain.MonitorCheck, error) {
	var c domain.MonitorCheck
	err := r.db.QueryRow(ctx,
		`SELECT id, monitor_id, status, status_code, response_time_ms, error_message, checked_at
		 FROM monitor_checks
		 WHERE monitor_id = $1
		 ORDER BY checked_at DESC
		 LIMIT 1`,
		monitorID,
	).Scan(&c.ID, &c.MonitorID, &c.Status, &c.StatusCode, &c.ResponseTimeMs, &c.ErrorMessage, &c.CheckedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("monitor_repository: GetLastCheck: %w", err)
	}
	return &c, nil
}

// GetHistoryLast24h returns all checks for a monitor in the last 24 hours.
func (r *MonitorRepository) GetHistoryLast24h(ctx context.Context, monitorID string) ([]domain.MonitorCheck, error) {
	since := time.Now().UTC().Add(-24 * time.Hour)
	rows, err := r.db.Query(ctx,
		`SELECT id, monitor_id, status, status_code, response_time_ms, error_message, checked_at
		 FROM monitor_checks
		 WHERE monitor_id = $1 AND checked_at >= $2
		 ORDER BY checked_at DESC`,
		monitorID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor_repository: GetHistoryLast24h: %w", err)
	}
	defer rows.Close()

	var checks []domain.MonitorCheck
	for rows.Next() {
		var c domain.MonitorCheck
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.Status, &c.StatusCode, &c.ResponseTimeMs, &c.ErrorMessage, &c.CheckedAt); err != nil {
			return nil, fmt.Errorf("monitor_repository: GetHistoryLast24h scan: %w", err)
		}
		checks = append(checks, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("monitor_repository: GetHistoryLast24h rows: %w", err)
	}
	return checks, nil
}

// GetUptimeStats returns uptime % and avg response time for the last 24h for a given monitor.
func (r *MonitorRepository) GetUptimeStats(ctx context.Context, monitorID string) (uptime float64, avgResponseMs float64, err error) {
	since := time.Now().UTC().Add(-24 * time.Hour)

	var total, up int
	var avgResp *float64

	err = r.db.QueryRow(ctx,
		`SELECT
		    COUNT(*) AS total,
		    COUNT(*) FILTER (WHERE status = 'up') AS up_count,
		    AVG(response_time_ms) FILTER (WHERE status = 'up') AS avg_ms
		 FROM monitor_checks
		 WHERE monitor_id = $1 AND checked_at >= $2`,
		monitorID, since,
	).Scan(&total, &up, &avgResp)
	if err != nil {
		return 0, 0, fmt.Errorf("monitor_repository: GetUptimeStats: %w", err)
	}

	if total > 0 {
		uptime = float64(up) / float64(total) * 100
	}
	if avgResp != nil {
		avgResponseMs = *avgResp
	}
	return uptime, avgResponseMs, nil
}
