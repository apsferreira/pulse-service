package repository

import (
	"context"
	"encoding/json"
	"fmt"
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
