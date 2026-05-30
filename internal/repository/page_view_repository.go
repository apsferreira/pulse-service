package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PageViewRepository persiste e consulta page views.
type PageViewRepository struct {
	db *pgxpool.Pool
}

// NewPageViewRepository cria um novo PageViewRepository.
func NewPageViewRepository(db *pgxpool.Pool) *PageViewRepository {
	return &PageViewRepository{db: db}
}

// Insert persiste um page view e retorna o ID gerado.
func (r *PageViewRepository) Insert(ctx context.Context, req *domain.TrackPageViewRequest) (string, error) {
	id := uuid.New().String()

	_, err := r.db.Exec(ctx,
		`INSERT INTO page_views
		   (id, session_id, user_id, product, page_title, page_path, time_on_page_ms, referrer, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		id,
		nullString(req.SessionID),
		req.UserID,
		nullString(req.Product),
		nullString(req.PageTitle),
		nullString(req.PagePath),
		req.TimeOnPageMs,
		nullString(req.Referrer),
	)
	if err != nil {
		return "", fmt.Errorf("page_view_repository: Insert: %w", err)
	}
	return id, nil
}

// TopPages retorna as N páginas mais visitadas de um produto em um período.
func (r *PageViewRepository) TopPages(ctx context.Context, product string, from, to time.Time, limit int) ([]domain.TopPageItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT page_path,
		        COUNT(*) AS total_views,
		        COALESCE(AVG(time_on_page_ms), 0) AS avg_time_on_page_ms
		 FROM page_views
		 WHERE product = $1
		   AND created_at >= $2
		   AND created_at <= $3
		   AND page_path IS NOT NULL
		 GROUP BY page_path
		 ORDER BY total_views DESC
		 LIMIT $4`,
		product, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("page_view_repository: TopPages: %w", err)
	}
	defer rows.Close()

	var items []domain.TopPageItem
	for rows.Next() {
		var item domain.TopPageItem
		if err := rows.Scan(&item.PagePath, &item.TotalViews, &item.AvgTimeOnPageMs); err != nil {
			return nil, fmt.Errorf("page_view_repository: TopPages scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("page_view_repository: TopPages rows: %w", err)
	}
	return items, nil
}

// UserJourney retorna as últimas páginas visitadas por um usuário em um produto.
func (r *PageViewRepository) UserJourney(ctx context.Context, userID, product string, limit int) ([]domain.UserJourneyItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, session_id, product, page_title, page_path, time_on_page_ms, referrer, created_at
		 FROM page_views
		 WHERE user_id = $1::uuid
		   AND product = $2
		 ORDER BY created_at DESC
		 LIMIT $3`,
		userID, product, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("page_view_repository: UserJourney: %w", err)
	}
	defer rows.Close()

	var items []domain.UserJourneyItem
	for rows.Next() {
		var item domain.UserJourneyItem
		if err := rows.Scan(
			&item.ID,
			&item.SessionID,
			&item.Product,
			&item.PageTitle,
			&item.PagePath,
			&item.TimeOnPageMs,
			&item.Referrer,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("page_view_repository: UserJourney scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("page_view_repository: UserJourney rows: %w", err)
	}
	return items, nil
}

// nullString converte string vazia em nil para evitar gravar strings vazias no banco.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
