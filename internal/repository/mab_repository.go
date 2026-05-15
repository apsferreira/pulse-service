package repository

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MABRepository gerencia estado dos experimentos Multi-Armed Bandit. BKL-1154.
type MABRepository struct {
	db *pgxpool.Pool
}

func NewMABRepository(db *pgxpool.Pool) *MABRepository {
	return &MABRepository{db: db}
}

// CreateExperiment cria um experimento MAB com suas variantes iniciais (peso = 1/N).
func (r *MABRepository) CreateExperiment(ctx context.Context, projectID uuid.UUID, req *domain.CreateMABExperimentRequest) (*domain.MABExperiment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: CreateExperiment: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var exp domain.MABExperiment
	err = tx.QueryRow(ctx,
		`INSERT INTO mab_experiments (project_id, flag_key, epsilon)
		 VALUES ($1, $2, $3)
		 RETURNING id, project_id, flag_key, epsilon, is_active, created_at, updated_at`,
		projectID, req.FlagKey, req.Epsilon,
	).Scan(&exp.ID, &exp.ProjectID, &exp.FlagKey, &exp.Epsilon, &exp.IsActive, &exp.CreatedAt, &exp.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: CreateExperiment: insert experiment: %w", err)
	}

	initialWeight := 1.0 / float64(len(req.Variants))
	for _, v := range req.Variants {
		var vs domain.MABVariantStats
		err = tx.QueryRow(ctx,
			`INSERT INTO mab_variant_stats (experiment_id, variant_key, weight)
			 VALUES ($1, $2, $3)
			 RETURNING id, experiment_id, variant_key, impressions, conversions, weight, updated_at`,
			exp.ID, v.VariantKey, initialWeight,
		).Scan(&vs.ID, &vs.ExperimentID, &vs.VariantKey, &vs.Impressions, &vs.Conversions, &vs.Weight, &vs.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("mab_repository: CreateExperiment: insert variant %q: %w", v.VariantKey, err)
		}
		exp.Variants = append(exp.Variants, vs)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("mab_repository: CreateExperiment: commit: %w", err)
	}
	return &exp, nil
}

// GetExperiment busca experimento por project_id + flag_key, incluindo variantes.
func (r *MABRepository) GetExperiment(ctx context.Context, projectID uuid.UUID, flagKey string) (*domain.MABExperiment, error) {
	var exp domain.MABExperiment
	err := r.db.QueryRow(ctx,
		`SELECT id, project_id, flag_key, epsilon, is_active, created_at, updated_at
		 FROM mab_experiments
		 WHERE project_id = $1 AND flag_key = $2`,
		projectID, flagKey,
	).Scan(&exp.ID, &exp.ProjectID, &exp.FlagKey, &exp.Epsilon, &exp.IsActive, &exp.CreatedAt, &exp.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("mab_repository: GetExperiment: %w", err)
	}

	variants, err := r.listVariants(ctx, r.db, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants
	return &exp, nil
}

// GetExperimentByID busca experimento por UUID, incluindo variantes.
func (r *MABRepository) GetExperimentByID(ctx context.Context, id uuid.UUID) (*domain.MABExperiment, error) {
	var exp domain.MABExperiment
	err := r.db.QueryRow(ctx,
		`SELECT id, project_id, flag_key, epsilon, is_active, created_at, updated_at
		 FROM mab_experiments
		 WHERE id = $1`,
		id,
	).Scan(&exp.ID, &exp.ProjectID, &exp.FlagKey, &exp.Epsilon, &exp.IsActive, &exp.CreatedAt, &exp.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("mab_repository: GetExperimentByID: %w", err)
	}

	variants, err := r.listVariants(ctx, r.db, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants
	return &exp, nil
}

// CountExperiments retorna o número de experimentos MAB ativos em um projeto.
// Usado pelo gating de plano no MABHandler.
func (r *MABRepository) CountExperiments(ctx context.Context, projectID uuid.UUID) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM mab_experiments WHERE project_id = $1 AND is_active = true`,
		projectID,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("mab_repository: CountExperiments: %w", err)
	}
	return n, nil
}

// ListExperiments lista todos os experimentos de um projeto, sem variantes.
func (r *MABRepository) ListExperiments(ctx context.Context, projectID uuid.UUID) ([]domain.MABExperiment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, flag_key, epsilon, is_active, created_at, updated_at
		 FROM mab_experiments
		 WHERE project_id = $1
		 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: ListExperiments: %w", err)
	}
	defer rows.Close()

	var experiments []domain.MABExperiment
	for rows.Next() {
		var exp domain.MABExperiment
		if err := rows.Scan(&exp.ID, &exp.ProjectID, &exp.FlagKey, &exp.Epsilon, &exp.IsActive, &exp.CreatedAt, &exp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("mab_repository: ListExperiments: scan: %w", err)
		}
		experiments = append(experiments, exp)
	}
	return experiments, rows.Err()
}

// RecordEvent incrementa impressions (sempre) e conversions (se isConversion=true) de uma variante.
func (r *MABRepository) RecordEvent(ctx context.Context, experimentID uuid.UUID, variantKey string, isConversion bool) error {
	convDelta := 0
	if isConversion {
		convDelta = 1
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE mab_variant_stats
		 SET impressions = impressions + 1,
		     conversions = conversions + $3,
		     updated_at  = now()
		 WHERE experiment_id = $1 AND variant_key = $2`,
		experimentID, variantKey, convDelta,
	)
	if err != nil {
		return fmt.Errorf("mab_repository: RecordEvent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mab_repository: RecordEvent: variant %q not found in experiment %s", variantKey, experimentID)
	}
	return nil
}

// RecalculateWeights recalcula os pesos de todas as variantes do experimento (normaliza taxa de conversão).
// Variantes sem impressões recebem o peso médio para não ficarem permanentemente em zero.
func (r *MABRepository) RecalculateWeights(ctx context.Context, experimentID uuid.UUID) error {
	rows, err := r.db.Query(ctx,
		`SELECT id, impressions, conversions
		 FROM mab_variant_stats
		 WHERE experiment_id = $1`,
		experimentID,
	)
	if err != nil {
		return fmt.Errorf("mab_repository: RecalculateWeights: query: %w", err)
	}
	defer rows.Close()

	type variantRow struct {
		id          uuid.UUID
		impressions int64
		conversions int64
	}
	var variants []variantRow
	for rows.Next() {
		var v variantRow
		if err := rows.Scan(&v.id, &v.impressions, &v.conversions); err != nil {
			return fmt.Errorf("mab_repository: RecalculateWeights: scan: %w", err)
		}
		variants = append(variants, v)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("mab_repository: RecalculateWeights: rows: %w", err)
	}
	if len(variants) == 0 {
		return nil
	}

	// Calcula taxa de conversão por variante; sem impressões recebe 0.
	rates := make([]float64, len(variants))
	var total float64
	for i, v := range variants {
		if v.impressions > 0 {
			rates[i] = float64(v.conversions) / float64(v.impressions)
		}
		total += rates[i]
	}

	// Normaliza; se total=0 distribui igualmente.
	weights := make([]float64, len(variants))
	if total == 0 {
		eq := 1.0 / float64(len(variants))
		for i := range weights {
			weights[i] = eq
		}
	} else {
		for i, r := range rates {
			weights[i] = r / total
		}
	}

	// Atualiza em batch dentro de uma transação.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mab_repository: RecalculateWeights: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for i, v := range variants {
		if _, err := tx.Exec(ctx,
			`UPDATE mab_variant_stats SET weight = $1, updated_at = now() WHERE id = $2`,
			weights[i], v.id,
		); err != nil {
			return fmt.Errorf("mab_repository: RecalculateWeights: update variant %s: %w", v.id, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("mab_repository: RecalculateWeights: commit: %w", err)
	}
	return nil
}

// SelectVariant implementa epsilon-greedy: sorteia aleatoriamente (explore) ou escolhe a melhor variante (exploit).
// Incrementa impressions da variante escolhida atomicamente dentro de uma transação.
func (r *MABRepository) SelectVariant(ctx context.Context, experimentID uuid.UUID, epsilon float64) (*domain.SelectVariantResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: SelectVariant: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx,
		`SELECT id, variant_key, weight
		 FROM mab_variant_stats
		 WHERE experiment_id = $1
		 ORDER BY variant_key
		 FOR UPDATE`,
		experimentID,
	)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: SelectVariant: query: %w", err)
	}

	type variantRow struct {
		id         uuid.UUID
		variantKey string
		weight     float64
	}
	var variants []variantRow
	for rows.Next() {
		var v variantRow
		if err := rows.Scan(&v.id, &v.variantKey, &v.weight); err != nil {
			rows.Close()
			return nil, fmt.Errorf("mab_repository: SelectVariant: scan: %w", err)
		}
		variants = append(variants, v)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mab_repository: SelectVariant: rows: %w", err)
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("mab_repository: SelectVariant: no variants for experiment %s", experimentID)
	}

	// Epsilon-greedy: explore vs exploit.
	explore := rand.Float64() < epsilon //nolint:gosec
	var chosen variantRow
	if explore {
		chosen = variants[rand.Intn(len(variants))] //nolint:gosec
	} else {
		chosen = variants[0]
		for _, v := range variants[1:] {
			if v.weight > chosen.weight {
				chosen = v
			}
		}
	}

	if _, err := tx.Exec(ctx,
		`UPDATE mab_variant_stats SET impressions = impressions + 1, updated_at = now() WHERE id = $1`,
		chosen.id,
	); err != nil {
		return nil, fmt.Errorf("mab_repository: SelectVariant: increment impressions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("mab_repository: SelectVariant: commit: %w", err)
	}

	return &domain.SelectVariantResponse{
		VariantKey: chosen.variantKey,
		IsExplore:  explore,
		Weight:     chosen.weight,
	}, nil
}

// SetActive ativa ou desativa um experimento MAB.
func (r *MABRepository) SetActive(ctx context.Context, experimentID uuid.UUID, active bool) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE mab_experiments SET is_active = $1, updated_at = now() WHERE id = $2`,
		active, experimentID,
	)
	if err != nil {
		return fmt.Errorf("mab_repository: SetActive: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mab_repository: SetActive: experiment %s not found", experimentID)
	}
	return nil
}

// listVariants é um helper interno que carrega variantes de um experimento via qualquer executor (pool ou tx).
func (r *MABRepository) listVariants(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, experimentID uuid.UUID) ([]domain.MABVariantStats, error) {
	rows, err := db.Query(ctx,
		`SELECT id, experiment_id, variant_key, impressions, conversions, weight, updated_at
		 FROM mab_variant_stats
		 WHERE experiment_id = $1
		 ORDER BY variant_key`,
		experimentID,
	)
	if err != nil {
		return nil, fmt.Errorf("mab_repository: listVariants: %w", err)
	}
	defer rows.Close()

	var variants []domain.MABVariantStats
	for rows.Next() {
		var vs domain.MABVariantStats
		if err := rows.Scan(&vs.ID, &vs.ExperimentID, &vs.VariantKey, &vs.Impressions, &vs.Conversions, &vs.Weight, &vs.UpdatedAt); err != nil {
			return nil, fmt.Errorf("mab_repository: listVariants: scan: %w", err)
		}
		if vs.Impressions > 0 {
			vs.ConvRate = float64(vs.Conversions) / float64(vs.Impressions)
		}
		variants = append(variants, vs)
	}
	return variants, rows.Err()
}
