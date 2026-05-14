package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

// Insert persists a new event.
// BKL-750: persiste environment (production|staging|development).
// BKL-1038: persiste project_id para isolamento multi-tenant.
func (r *EventRepository) Insert(ctx context.Context, e *domain.Event) error {
	e.ID = uuid.New().String()
	e.CreatedAt = time.Now().UTC()
	if e.Environment == "" {
		e.Environment = "production"
	}

	props, err := json.Marshal(e.Properties)
	if err != nil {
		return fmt.Errorf("event_repository: Insert marshal properties: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO events (id, product, event_type, user_id, session_id, properties, ip_address, user_agent, environment, project_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		e.ID, e.Product, e.EventType, e.UserID, e.SessionID, props, e.IPAddress, e.UserAgent, e.Environment, e.ProjectID, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("event_repository: Insert: %w", err)
	}
	return nil
}

// InsertBatch persiste múltiplos eventos em uma única operação usando pgx CopyFrom.
// BKL-1041: substitui N INSERTs individuais por um COPY — throughput ~10x maior.
// Máximo de 500 eventos por chamada (validado na camada de serviço).
func (r *EventRepository) InsertBatch(ctx context.Context, events []*domain.Event) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	now := time.Now().UTC()
	rows := make([][]interface{}, 0, len(events))

	for _, e := range events {
		e.ID = uuid.New().String()
		e.CreatedAt = now
		if e.Environment == "" {
			e.Environment = "production"
		}

		props, err := json.Marshal(e.Properties)
		if err != nil {
			props = []byte("{}")
		}

		rows = append(rows, []interface{}{
			e.ID, e.Product, e.EventType, e.UserID, e.SessionID,
			props, e.IPAddress, e.UserAgent, e.Environment, e.ProjectID, e.CreatedAt,
		})
	}

	cols := []string{"id", "product", "event_type", "user_id", "session_id", "properties", "ip_address", "user_agent", "environment", "project_id", "created_at"}
	n, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"events"},
		cols,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("event_repository: InsertBatch: %w", err)
	}
	return int(n), nil
}

// MergeAnonymousUser atualiza todos os eventos de um anonymous_id para um user_id real.
func (r *EventRepository) MergeAnonymousUser(ctx context.Context, anonymousID, userID string) (int, error) {
	result, err := r.db.Exec(ctx,
		`UPDATE events SET user_id = $1 WHERE user_id = $2`,
		userID, anonymousID,
	)
	if err != nil {
		return 0, fmt.Errorf("event_repository: MergeAnonymousUser: %w", err)
	}
	return int(result.RowsAffected()), nil
}

// GetSummaryLast24h returns counts grouped by product and event_type for the last 24 hours.
// If projectID is non-empty, filters to that project only.
func (r *EventRepository) GetSummaryLast24h(ctx context.Context, projectID string) ([]domain.EventSummaryItem, error) {
	since := time.Now().UTC().Add(-24 * time.Hour)

	var rows pgx.Rows
	var err error
	if projectID != "" {
		rows, err = r.db.Query(ctx,
			`SELECT product, event_type, COUNT(*) AS cnt
			 FROM events
			 WHERE created_at >= $1 AND project_id = $2
			 GROUP BY product, event_type
			 ORDER BY product ASC, cnt DESC`,
			since, projectID,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT product, event_type, COUNT(*) AS cnt
			 FROM events
			 WHERE created_at >= $1
			 GROUP BY product, event_type
			 ORDER BY product ASC, cnt DESC`,
			since,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("event_repository: GetSummaryLast24h: %w", err)
	}
	defer rows.Close()

	var items []domain.EventSummaryItem
	for rows.Next() {
		var item domain.EventSummaryItem
		if err := rows.Scan(&item.Product, &item.EventType, &item.Count); err != nil {
			return nil, fmt.Errorf("event_repository: GetSummaryLast24h scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("event_repository: GetSummaryLast24h rows: %w", err)
	}
	return items, nil
}

// CountByEventName returns event counts grouped by event_type, optionally filtered by event_name and date range.
// BKL-1026: queries estáticas por combinação de filtros — sem fmt.Sprintf em SQL.
func (r *EventRepository) CountByEventName(ctx context.Context, eventName string, startDate, endDate *time.Time) ([]domain.EventCountItem, error) {
	var rows pgx.Rows
	var err error

	const base = `SELECT event_type, COUNT(*) AS cnt FROM events`
	const tail = ` GROUP BY event_type ORDER BY cnt DESC`

	switch {
	case eventName != "" && startDate != nil && endDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE event_type = $1 AND created_at >= $2 AND created_at <= $3`+tail, eventName, *startDate, *endDate)
	case eventName != "" && startDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE event_type = $1 AND created_at >= $2`+tail, eventName, *startDate)
	case eventName != "" && endDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE event_type = $1 AND created_at <= $2`+tail, eventName, *endDate)
	case eventName != "":
		rows, err = r.db.Query(ctx, base+` WHERE event_type = $1`+tail, eventName)
	case startDate != nil && endDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE created_at >= $1 AND created_at <= $2`+tail, *startDate, *endDate)
	case startDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE created_at >= $1`+tail, *startDate)
	case endDate != nil:
		rows, err = r.db.Query(ctx, base+` WHERE created_at <= $1`+tail, *endDate)
	default:
		rows, err = r.db.Query(ctx, base+tail)
	}
	if err != nil {
		return nil, fmt.Errorf("event_repository: CountByEventName: %w", err)
	}
	defer rows.Close()

	var items []domain.EventCountItem
	for rows.Next() {
		var item domain.EventCountItem
		if err := rows.Scan(&item.EventName, &item.Count); err != nil {
			return nil, fmt.Errorf("event_repository: CountByEventName scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("event_repository: CountByEventName rows: %w", err)
	}
	return items, nil
}

// ListEvents returns events with optional filters and pagination.
// BKL-1026: queries com filtros estáticos para evitar gosec G201 (dynamic SQL injection risk).
func (r *EventRepository) ListEvents(ctx context.Context, eventName, userID string, limit, offset int) ([]domain.Event, int, error) {
	total, rows, err := r.listEventsQuery(ctx, eventName, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		var propsRaw []byte
		if err := rows.Scan(&e.ID, &e.Product, &e.EventType, &e.UserID, &e.SessionID, &propsRaw, &e.IPAddress, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("event_repository: ListEvents scan: %w", err)
		}
		if propsRaw != nil {
			if err := json.Unmarshal(propsRaw, &e.Properties); err != nil {
				e.Properties = make(map[string]interface{})
			}
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("event_repository: ListEvents rows: %w", err)
	}
	return events, total, nil
}

// listEventsQuery selects count + data rows using static parameterized queries (no dynamic SQL).
func (r *EventRepository) listEventsQuery(ctx context.Context, eventName, userID string, limit, offset int) (int, pgx.Rows, error) {
	const selectCols = `SELECT id, product, event_type, user_id, session_id, properties, ip_address::text, user_agent, created_at FROM events`
	var total int
	var rows pgx.Rows
	var err error

	switch {
	case eventName != "" && userID != "":
		if err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE event_type = $1 AND user_id::text = $2`, eventName, userID).Scan(&total); err != nil {
			return 0, nil, fmt.Errorf("event_repository: ListEvents count: %w", err)
		}
		rows, err = r.db.Query(ctx, selectCols+` WHERE event_type = $1 AND user_id::text = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`, eventName, userID, limit, offset)

	case eventName != "":
		if err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE event_type = $1`, eventName).Scan(&total); err != nil {
			return 0, nil, fmt.Errorf("event_repository: ListEvents count: %w", err)
		}
		rows, err = r.db.Query(ctx, selectCols+` WHERE event_type = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, eventName, limit, offset)

	case userID != "":
		if err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE user_id::text = $1`, userID).Scan(&total); err != nil {
			return 0, nil, fmt.Errorf("event_repository: ListEvents count: %w", err)
		}
		rows, err = r.db.Query(ctx, selectCols+` WHERE user_id::text = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)

	default:
		if err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events`).Scan(&total); err != nil {
			return 0, nil, fmt.Errorf("event_repository: ListEvents count: %w", err)
		}
		rows, err = r.db.Query(ctx, selectCols+` ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	}

	if err != nil {
		return 0, nil, fmt.Errorf("event_repository: ListEvents query: %w", err)
	}
	return total, rows, nil
}
