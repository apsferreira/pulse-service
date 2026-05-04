package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SegmentRepository gerencia segmentos de targeting avançado para feature flags. BKL-1153.
type SegmentRepository struct {
	db *pgxpool.Pool
}

func NewSegmentRepository(db *pgxpool.Pool) *SegmentRepository {
	return &SegmentRepository{db: db}
}

// Create insere um novo segmento associado ao projeto.
func (r *SegmentRepository) Create(ctx context.Context, projectID uuid.UUID, req *domain.CreateSegmentRequest) (*domain.Segment, error) {
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: Create: marshal rules: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`INSERT INTO flag_segments (project_id, name, description, rules)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, project_id, name, description, rules, created_at, updated_at`,
		projectID, req.Name, req.Description, rulesJSON,
	)

	return scanSegment(row)
}

// Update atualiza campos não-nulos de um segmento.
func (r *SegmentRepository) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSegmentRequest) (*domain.Segment, error) {
	seg, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if seg == nil {
		return nil, nil
	}

	if req.Name != nil {
		seg.Name = *req.Name
	}
	if req.Description != nil {
		seg.Description = *req.Description
	}
	if req.Rules != nil {
		seg.Rules = req.Rules
	}

	rulesJSON, err := json.Marshal(seg.Rules)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: Update: marshal rules: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`UPDATE flag_segments
		 SET name = $2, description = $3, rules = $4, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, project_id, name, description, rules, created_at, updated_at`,
		id, seg.Name, seg.Description, rulesJSON,
	)

	return scanSegment(row)
}

// Delete remove um segmento pelo ID.
func (r *SegmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM flag_segments WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("segment_repository: Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetByID busca um segmento pelo ID.
func (r *SegmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Segment, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, name, description, rules, created_at, updated_at
		 FROM flag_segments
		 WHERE id = $1`,
		id,
	)
	seg, err := scanSegment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return seg, nil
}

// ListByProject retorna todos os segmentos de um projeto ordenados por nome.
func (r *SegmentRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Segment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, name, description, rules, created_at, updated_at
		 FROM flag_segments
		 WHERE project_id = $1
		 ORDER BY name ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: ListByProject: %w", err)
	}
	defer rows.Close()

	var segments []domain.Segment
	for rows.Next() {
		seg, err := scanSegmentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("segment_repository: ListByProject scan: %w", err)
		}
		segments = append(segments, *seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("segment_repository: ListByProject rows: %w", err)
	}
	return segments, nil
}

// AssignToFlag associa um segmento a uma flag com percentual de rollout.
func (r *SegmentRepository) AssignToFlag(ctx context.Context, flagID, segmentID uuid.UUID, rolloutPct int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO flag_segment_assignments (flag_id, segment_id, rollout_percentage)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (flag_id, segment_id) DO UPDATE SET rollout_percentage = EXCLUDED.rollout_percentage`,
		flagID, segmentID, rolloutPct,
	)
	if err != nil {
		return fmt.Errorf("segment_repository: AssignToFlag: %w", err)
	}
	return nil
}

// UnassignFromFlag remove a associação entre uma flag e um segmento.
func (r *SegmentRepository) UnassignFromFlag(ctx context.Context, flagID, segmentID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM flag_segment_assignments WHERE flag_id = $1 AND segment_id = $2`,
		flagID, segmentID,
	)
	if err != nil {
		return fmt.Errorf("segment_repository: UnassignFromFlag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListByFlag retorna todos os segmentos associados a uma flag específica.
func (r *SegmentRepository) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]domain.Segment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT fs.id, fs.project_id, fs.name, fs.description, fs.rules, fs.created_at, fs.updated_at
		 FROM flag_segments fs
		 INNER JOIN flag_segment_assignments fsa ON fsa.segment_id = fs.id
		 WHERE fsa.flag_id = $1
		 ORDER BY fs.name ASC`,
		flagID,
	)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: ListByFlag: %w", err)
	}
	defer rows.Close()

	var segments []domain.Segment
	for rows.Next() {
		seg, err := scanSegmentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("segment_repository: ListByFlag scan: %w", err)
		}
		segments = append(segments, *seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("segment_repository: ListByFlag rows: %w", err)
	}
	return segments, nil
}

// scanSegment lê um pgx.Row (QueryRow) e devolve um Segment.
func scanSegment(row pgx.Row) (*domain.Segment, error) {
	var seg domain.Segment
	var rulesJSON []byte
	err := row.Scan(
		&seg.ID,
		&seg.ProjectID,
		&seg.Name,
		&seg.Description,
		&rulesJSON,
		&seg.CreatedAt,
		&seg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: scan: %w", err)
	}
	if err := json.Unmarshal(rulesJSON, &seg.Rules); err != nil {
		return nil, fmt.Errorf("segment_repository: unmarshal rules: %w", err)
	}
	if seg.Rules == nil {
		seg.Rules = []domain.RuleGroup{}
	}
	return &seg, nil
}

// scanSegmentRow lê pgx.Rows (Query) e devolve um Segment.
func scanSegmentRow(rows pgx.Rows) (*domain.Segment, error) {
	var seg domain.Segment
	var rulesJSON []byte
	err := rows.Scan(
		&seg.ID,
		&seg.ProjectID,
		&seg.Name,
		&seg.Description,
		&rulesJSON,
		&seg.CreatedAt,
		&seg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("segment_repository: scanRow: %w", err)
	}
	if err := json.Unmarshal(rulesJSON, &seg.Rules); err != nil {
		return nil, fmt.Errorf("segment_repository: unmarshal rules: %w", err)
	}
	if seg.Rules == nil {
		seg.Rules = []domain.RuleGroup{}
	}
	return &seg, nil
}
