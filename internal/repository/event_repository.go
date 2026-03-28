package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

// Insert persists a new event.
func (r *EventRepository) Insert(ctx context.Context, e *domain.Event) error {
	e.ID = uuid.New().String()
	e.CreatedAt = time.Now().UTC()

	props, err := json.Marshal(e.Properties)
	if err != nil {
		return fmt.Errorf("event_repository: Insert marshal properties: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO events (id, product, event_type, user_id, session_id, properties, ip_address, user_agent, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.ID, e.Product, e.EventType, e.UserID, e.SessionID, props, e.IPAddress, e.UserAgent, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("event_repository: Insert: %w", err)
	}
	return nil
}

// GetSummaryLast24h returns counts grouped by product and event_type for the last 24 hours.
func (r *EventRepository) GetSummaryLast24h(ctx context.Context) ([]domain.EventSummaryItem, error) {
	since := time.Now().UTC().Add(-24 * time.Hour)

	rows, err := r.db.Query(ctx,
		`SELECT product, event_type, COUNT(*) AS cnt
		 FROM events
		 WHERE created_at >= $1
		 GROUP BY product, event_type
		 ORDER BY product ASC, cnt DESC`,
		since,
	)
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
func (r *EventRepository) CountByEventName(ctx context.Context, eventName string, startDate, endDate *time.Time) ([]domain.EventCountItem, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if eventName != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, eventName)
		argIdx++
	}
	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *startDate)
		argIdx++
	}
	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *endDate)
		argIdx++
	}

	query := `SELECT event_type, COUNT(*) AS cnt FROM events`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " GROUP BY event_type ORDER BY cnt DESC"

	rows, err := r.db.Query(ctx, query, args...)
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
func (r *EventRepository) ListEvents(ctx context.Context, eventName, userID string, limit, offset int) ([]domain.Event, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if eventName != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, eventName)
		argIdx++
	}
	if userID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id::text = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM events" + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("event_repository: ListEvents count: %w", err)
	}

	// Fetch page
	dataQuery := fmt.Sprintf(
		"SELECT id, product, event_type, user_id, session_id, properties, ip_address, user_agent, created_at FROM events%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, argIdx, argIdx+1,
	)
	dataArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("event_repository: ListEvents query: %w", err)
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
