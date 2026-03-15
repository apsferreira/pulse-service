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

type FlagRepository struct {
	db *pgxpool.Pool
}

func NewFlagRepository(db *pgxpool.Pool) *FlagRepository {
	return &FlagRepository{db: db}
}

// GetAll returns all feature flags, optionally filtered by product.
// When product is nil, returns all flags. When product is set, returns flags
// for that product AND global flags (product IS NULL).
func (r *FlagRepository) GetAll(ctx context.Context, product *string) ([]domain.FeatureFlag, error) {
	var rows pgx.Rows
	var err error

	if product != nil {
		rows, err = r.db.Query(ctx,
			`SELECT id, name, COALESCE(description, ''), is_enabled, product, rollout_pct, created_at, updated_at
			 FROM feature_flags
			 WHERE product = $1 OR product IS NULL
			 ORDER BY name ASC`,
			*product,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id, name, COALESCE(description, ''), is_enabled, product, rollout_pct, created_at, updated_at
			 FROM feature_flags
			 ORDER BY name ASC`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: GetAll query: %w", err)
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		var f domain.FeatureFlag
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.IsEnabled, &f.Product, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: GetAll scan: %w", err)
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: GetAll rows: %w", err)
	}
	return flags, nil
}

// GetByID returns a single feature flag by its UUID.
func (r *FlagRepository) GetByID(ctx context.Context, id string) (*domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	err := r.db.QueryRow(ctx,
		`SELECT id, name, COALESCE(description, ''), is_enabled, product, rollout_pct, created_at, updated_at
		 FROM feature_flags WHERE id = $1`,
		id,
	).Scan(&f.ID, &f.Name, &f.Description, &f.IsEnabled, &f.Product, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: GetByID: %w", err)
	}
	return &f, nil
}

// GetByName returns a flag by name and product (product empty string = global/NULL).
func (r *FlagRepository) GetByName(ctx context.Context, name, product string) (*domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	var err error

	if product == "" {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, COALESCE(description, ''), is_enabled, product, rollout_pct, created_at, updated_at
			 FROM feature_flags WHERE name = $1 AND product IS NULL`,
			name,
		).Scan(&f.ID, &f.Name, &f.Description, &f.IsEnabled, &f.Product, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, COALESCE(description, ''), is_enabled, product, rollout_pct, created_at, updated_at
			 FROM feature_flags WHERE name = $1 AND product = $2`,
			name, product,
		).Scan(&f.ID, &f.Name, &f.Description, &f.IsEnabled, &f.Product, &f.RolloutPct, &f.CreatedAt, &f.UpdatedAt)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: GetByName: %w", err)
	}
	return &f, nil
}

// Create inserts a new feature flag into the database.
func (r *FlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	flag.ID = uuid.New().String()
	now := time.Now().UTC()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO feature_flags (id, name, description, is_enabled, product, rollout_pct, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		flag.ID, flag.Name, flag.Description, flag.IsEnabled, flag.Product, flag.RolloutPct, flag.CreatedAt, flag.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: Create: %w", err)
	}
	return nil
}

// Update applies partial updates to a feature flag and returns the updated record.
func (r *FlagRepository) Update(ctx context.Context, id string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.IsEnabled != nil {
		existing.IsEnabled = *req.IsEnabled
	}
	if req.RolloutPct != nil {
		existing.RolloutPct = *req.RolloutPct
	}
	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.Exec(ctx,
		`UPDATE feature_flags
		 SET description = $1, is_enabled = $2, rollout_pct = $3, updated_at = $4
		 WHERE id = $5`,
		existing.Description, existing.IsEnabled, existing.RolloutPct, existing.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: Update: %w", err)
	}
	return existing, nil
}

// Delete removes a feature flag by ID. Cascades to flag_overrides.
func (r *FlagRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM feature_flags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("repository: Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("repository: Delete: flag not found")
	}
	return nil
}

// GetOverridesForUser returns all flag overrides that match the given userID or tenantID.
func (r *FlagRepository) GetOverridesForUser(ctx context.Context, userID, tenantID string) ([]domain.FlagOverride, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, flag_id, user_id, tenant_id, is_enabled
		 FROM flag_overrides
		 WHERE ($1 != '' AND user_id = $1)
		    OR ($2 != '' AND tenant_id = $2)`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: GetOverridesForUser: %w", err)
	}
	defer rows.Close()

	var overrides []domain.FlagOverride
	for rows.Next() {
		var o domain.FlagOverride
		if err := rows.Scan(&o.ID, &o.FlagID, &o.UserID, &o.TenantID, &o.IsEnabled); err != nil {
			return nil, fmt.Errorf("repository: GetOverridesForUser scan: %w", err)
		}
		overrides = append(overrides, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: GetOverridesForUser rows: %w", err)
	}
	return overrides, nil
}
