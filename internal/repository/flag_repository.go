package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlagRepository struct {
	db *pgxpool.Pool
}

func NewFlagRepository(db *pgxpool.Pool) *FlagRepository {
	return &FlagRepository{db: db}
}

// GetAll returns all feature flags ordered by key.
func (r *FlagRepository) GetAll(ctx context.Context) ([]domain.FeatureFlag, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, created_at, updated_at
		 FROM feature_flags
		 ORDER BY key ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: GetAll: %w", err)
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		f, err := scanFlag(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: GetAll scan: %w", err)
		}
		flags = append(flags, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: GetAll rows: %w", err)
	}
	return flags, nil
}

// GetByKey returns a single feature flag by its unique key.
func (r *FlagRepository) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, created_at, updated_at
		 FROM feature_flags WHERE key = $1`,
		key,
	)
	var f domain.FeatureFlag
	var targetingBytes []byte
	err := row.Scan(&f.ID, &f.Key, &f.Description, &f.Enabled, &f.RolloutPercentage, &targetingBytes, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: GetByKey: %w", err)
	}
	if len(targetingBytes) > 0 {
		_ = json.Unmarshal(targetingBytes, &f.Targeting)
	}
	return &f, nil
}

// Create inserts a new feature flag.
func (r *FlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	flag.ID = uuid.New().String()
	now := time.Now().UTC()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	targetingJSON, _ := json.Marshal(flag.Targeting)

	_, err := r.db.Exec(ctx,
		`INSERT INTO feature_flags (id, key, description, enabled, rollout_percentage, targeting, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		flag.ID, flag.Key, flag.Description, flag.Enabled, flag.RolloutPercentage, targetingJSON, flag.CreatedAt, flag.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: Create: %w", err)
	}
	return nil
}

// Update applies partial updates to a feature flag identified by key.
func (r *FlagRepository) Update(ctx context.Context, key string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	existing, err := r.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.RolloutPercentage != nil {
		existing.RolloutPercentage = *req.RolloutPercentage
	}
	if req.Targeting != nil {
		existing.Targeting = *req.Targeting
	}
	existing.UpdatedAt = time.Now().UTC()

	targetingJSON, _ := json.Marshal(existing.Targeting)

	_, err = r.db.Exec(ctx,
		`UPDATE feature_flags
		 SET description = $1, enabled = $2, rollout_percentage = $3, targeting = $4, updated_at = $5
		 WHERE key = $6`,
		existing.Description, existing.Enabled, existing.RolloutPercentage, targetingJSON, existing.UpdatedAt, key,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: Update: %w", err)
	}
	return existing, nil
}

// Delete removes a feature flag by key.
func (r *FlagRepository) Delete(ctx context.Context, key string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM feature_flags WHERE key = $1`, key)
	if err != nil {
		return fmt.Errorf("repository: Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("repository: Delete: flag not found")
	}
	return nil
}

// scanFlag escaneia uma row de feature_flags com os novos campos.
func scanFlag(rows pgx.Rows) (*domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	var targetingBytes []byte
	if err := rows.Scan(&f.ID, &f.Key, &f.Description, &f.Enabled, &f.RolloutPercentage, &targetingBytes, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	if len(targetingBytes) > 0 {
		_ = json.Unmarshal(targetingBytes, &f.Targeting)
	}
	return &f, nil
}
