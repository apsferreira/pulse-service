package repository

// BKL-1074: Repositório de audit log de avaliações de feature flags.

import (
	"context"
	"fmt"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlagEvaluationRepository struct {
	db *pgxpool.Pool
}

func NewFlagEvaluationRepository(db *pgxpool.Pool) *FlagEvaluationRepository {
	return &FlagEvaluationRepository{db: db}
}

// Record persiste uma nova avaliação de flag no audit log.
// Executado de forma assíncrona — não bloqueia o path de avaliação.
func (r *FlagEvaluationRepository) Record(ctx context.Context, eval *domain.FlagEvaluation) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO flag_evaluations (flag_key, project_id, user_id, tenant_id, result, variant, reason, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, eval.FlagKey, eval.ProjectID, eval.UserID, eval.TenantID, eval.Result, eval.Variant, eval.Reason)
	if err != nil {
		return fmt.Errorf("flag_evaluation: Record: %w", err)
	}
	return nil
}

// List retorna avaliações filtradas por flag_key e/ou project_id, com paginação.
func (r *FlagEvaluationRepository) List(ctx context.Context, flagKey, projectID string, limit, offset int) ([]domain.FlagEvaluation, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	// Constrói query com filtros opcionais
	args := []any{}
	where := "WHERE 1=1"
	argIdx := 1

	if flagKey != "" {
		where += fmt.Sprintf(" AND flag_key = $%d", argIdx)
		args = append(args, flagKey)
		argIdx++
	}
	if projectID != "" {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, projectID)
		argIdx++
	}

	// Count
	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM flag_evaluations "+where,
		countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("flag_evaluation: List count: %w", err)
	}

	// Data
	listArgs := append(args, limit, offset)
	rows, err := r.db.Query(ctx,
		"SELECT id, flag_key, project_id, user_id, tenant_id, result, variant, reason, evaluated_at "+
			"FROM flag_evaluations "+where+
			fmt.Sprintf(" ORDER BY evaluated_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1),
		listArgs...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("flag_evaluation: List query: %w", err)
	}
	defer rows.Close()

	var items []domain.FlagEvaluation
	for rows.Next() {
		var e domain.FlagEvaluation
		if err := rows.Scan(&e.ID, &e.FlagKey, &e.ProjectID, &e.UserID, &e.TenantID,
			&e.Result, &e.Variant, &e.Reason, &e.EvaluatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	if items == nil {
		items = []domain.FlagEvaluation{}
	}
	return items, total, rows.Err()
}
