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
// Se projectID for não-nil, filtra por project_id. BKL-323.
func (r *FlagRepository) GetAll(ctx context.Context, projectID *string) ([]domain.FeatureFlag, error) {
	var rows pgx.Rows
	var err error
	if projectID != nil {
		rows, err = r.db.Query(ctx,
			`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, COALESCE(variants, '[]'::jsonb), COALESCE(prerequisites, '[]'::jsonb), project_id, COALESCE(require_approval, false), starts_at, ends_at, created_at, updated_at
			 FROM feature_flags
			 WHERE project_id = $1::uuid
			 ORDER BY key ASC`,
			*projectID,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, COALESCE(variants, '[]'::jsonb), COALESCE(prerequisites, '[]'::jsonb), project_id, COALESCE(require_approval, false), starts_at, ends_at, created_at, updated_at
			 FROM feature_flags
			 ORDER BY key ASC`,
		)
	}
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
		`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, COALESCE(variants, '[]'::jsonb), COALESCE(prerequisites, '[]'::jsonb), project_id, COALESCE(require_approval, false), starts_at, ends_at, created_at, updated_at
		 FROM feature_flags WHERE key = $1`,
		key,
	)
	var f domain.FeatureFlag
	var targetingBytes []byte
	var variantsBytes []byte
	var prerequisitesBytes []byte
	err := row.Scan(&f.ID, &f.Key, &f.Description, &f.Enabled, &f.RolloutPercentage, &targetingBytes, &variantsBytes, &prerequisitesBytes, &f.ProjectID, &f.RequireApproval, &f.StartsAt, &f.EndsAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: GetByKey: %w", err)
	}
	if len(targetingBytes) > 0 {
		_ = json.Unmarshal(targetingBytes, &f.Targeting)
	}
	if len(variantsBytes) > 0 {
		_ = json.Unmarshal(variantsBytes, &f.Variants)
	}
	if len(prerequisitesBytes) > 0 {
		_ = json.Unmarshal(prerequisitesBytes, &f.Prerequisites)
	}
	return &f, nil
}

// GetByKeyWithPrompt busca uma flag pelo key incluindo o campo prompt_variant. BKL-1002.
// Usado pelo endpoint GET /api/v1/flags/:key/evaluate para expor a config de LLM.
func (r *FlagRepository) GetByKeyWithPrompt(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	f, err := r.GetByKey(ctx, key)
	if err != nil || f == nil {
		return f, err
	}

	// Busca prompt_variant separadamente para não quebrar scanFlag existente
	var promptBytes []byte
	_ = r.db.QueryRow(ctx,
		`SELECT prompt_variant FROM feature_flags WHERE key = $1`,
		key,
	).Scan(&promptBytes)

	if len(promptBytes) > 0 {
		var pv domain.PromptVariant
		if err := json.Unmarshal(promptBytes, &pv); err == nil {
			f.PromptVariant = &pv
		}
	}
	return f, nil
}

// GetByKeys returns multiple feature flags by their keys (used by evaluate-batch and variants). BKL-094.
func (r *FlagRepository) GetByKeys(ctx context.Context, keys []string) ([]domain.FeatureFlag, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, key, COALESCE(description, ''), enabled, rollout_percentage, targeting, COALESCE(variants, '[]'::jsonb), COALESCE(prerequisites, '[]'::jsonb), project_id, COALESCE(require_approval, false), starts_at, ends_at, created_at, updated_at
		 FROM feature_flags
		 WHERE key = ANY($1)`,
		keys,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: GetByKeys: %w", err)
	}
	defer rows.Close()

	var flags []domain.FeatureFlag
	for rows.Next() {
		f, err := scanFlag(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: GetByKeys scan: %w", err)
		}
		flags = append(flags, *f)
	}
	return flags, rows.Err()
}

// Create inserts a new feature flag.
func (r *FlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	flag.ID = uuid.New().String()
	now := time.Now().UTC()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	targetingJSON, _ := json.Marshal(flag.Targeting)
	variantsJSON, _ := json.Marshal(flag.Variants)
	prerequisitesJSON, _ := json.Marshal(flag.Prerequisites)

	_, err := r.db.Exec(ctx,
		`INSERT INTO feature_flags (id, key, description, enabled, rollout_percentage, targeting, variants, prerequisites, project_id, require_approval, starts_at, ends_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		flag.ID, flag.Key, flag.Description, flag.Enabled, flag.RolloutPercentage,
		targetingJSON, variantsJSON, prerequisitesJSON, flag.ProjectID,
		flag.RequireApproval, flag.StartsAt, flag.EndsAt, flag.CreatedAt, flag.UpdatedAt,
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
	if req.Variants != nil {
		existing.Variants = req.Variants
	}
	if req.StartsAt != nil {
		existing.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		existing.EndsAt = req.EndsAt
	}
	if req.Prerequisites != nil {
		existing.Prerequisites = req.Prerequisites
	}
	if req.RequireApproval != nil {
		existing.RequireApproval = *req.RequireApproval
	}
	existing.UpdatedAt = time.Now().UTC()

	targetingJSON, _ := json.Marshal(existing.Targeting)
	variantsJSON, _ := json.Marshal(existing.Variants)
	prerequisitesJSON, _ := json.Marshal(existing.Prerequisites)

	_, err = r.db.Exec(ctx,
		`UPDATE feature_flags
		 SET description = $1, enabled = $2, rollout_percentage = $3, targeting = $4, variants = $5, starts_at = $6, ends_at = $7, prerequisites = $8, require_approval = $9, updated_at = $10
		 WHERE key = $11`,
		existing.Description, existing.Enabled, existing.RolloutPercentage, targetingJSON, variantsJSON,
		existing.StartsAt, existing.EndsAt, prerequisitesJSON, existing.RequireApproval, existing.UpdatedAt, key,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: Update: %w", err)
	}
	return existing, nil
}

// ActivateScheduledFlags ativa flags cujo starts_at já passou e desativa flags cujo ends_at já passou.
// BKL-1040: executado periodicamente pelo flag_scheduler.
// Retorna (ativadas, desativadas, error).
func (r *FlagRepository) ActivateScheduledFlags(ctx context.Context) (int, int, error) {
	now := time.Now().UTC()

	// Ativa flags com starts_at no passado que ainda estão desabilitadas
	resActivate, err := r.db.Exec(ctx, `
		UPDATE feature_flags
		SET enabled = true, updated_at = $1
		WHERE starts_at IS NOT NULL
		  AND starts_at <= $1
		  AND enabled = false
		  AND (ends_at IS NULL OR ends_at > $1)
	`, now)
	if err != nil {
		return 0, 0, fmt.Errorf("flag_repository: ActivateScheduledFlags activate: %w", err)
	}

	// Desativa flags com ends_at no passado que ainda estão habilitadas
	resDeactivate, err := r.db.Exec(ctx, `
		UPDATE feature_flags
		SET enabled = false, updated_at = $1
		WHERE ends_at IS NOT NULL
		  AND ends_at <= $1
		  AND enabled = true
	`, now)
	if err != nil {
		return int(resActivate.RowsAffected()), 0, fmt.Errorf("flag_repository: ActivateScheduledFlags deactivate: %w", err)
	}

	return int(resActivate.RowsAffected()), int(resDeactivate.RowsAffected()), nil
}

// Delete removes a feature flag by key.
// CountByProject retorna a quantidade de feature flags ativas em um projeto.
// Usado para enforcement de limite de plano (ex: free = max 5 flags).
func (r *FlagRepository) CountByProject(ctx context.Context, projectID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM feature_flags WHERE project_id = $1::uuid`,
		projectID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repository: CountByProject: %w", err)
	}
	return count, nil
}

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

// scanFlag escaneia uma row de feature_flags incluindo variants, prerequisites, project_id, require_approval e scheduled fields.
// BKL-094, BKL-323, BKL-1039, BKL-1040, BKL-990.
func scanFlag(rows pgx.Rows) (*domain.FeatureFlag, error) {
	var f domain.FeatureFlag
	var targetingBytes []byte
	var variantsBytes []byte
	var prerequisitesBytes []byte
	if err := rows.Scan(&f.ID, &f.Key, &f.Description, &f.Enabled, &f.RolloutPercentage, &targetingBytes, &variantsBytes, &prerequisitesBytes, &f.ProjectID, &f.RequireApproval, &f.StartsAt, &f.EndsAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
		return nil, err
	}
	if len(targetingBytes) > 0 {
		_ = json.Unmarshal(targetingBytes, &f.Targeting)
	}
	if len(variantsBytes) > 0 {
		_ = json.Unmarshal(variantsBytes, &f.Variants)
	}
	if len(prerequisitesBytes) > 0 {
		_ = json.Unmarshal(prerequisitesBytes, &f.Prerequisites)
	}
	return &f, nil
}
