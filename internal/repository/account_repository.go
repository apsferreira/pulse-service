package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountRepository gerencia contas no Pulse (plano por owner_email).
type AccountRepository struct {
	db *pgxpool.Pool
}

// NewAccountRepository cria um novo AccountRepository.
func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

// retentionByPlan retorna o número padrão de dias de retenção por plano.
// Duplica a lógica de project_repository para isolamento de pacote.
func accountRetentionByPlan(plan string) int {
	switch plan {
	case "pro", "enterprise", "internal":
		return 365
	case "starter", "growth":
		return 180
	default:
		return 90
	}
}

// GetByEmail retorna uma conta pelo email.
// Retorna nil, nil se o email não existir.
func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var a domain.Account
	err := r.db.QueryRow(ctx,
		`SELECT email, plan, data_retention_days, created_at, updated_at
		 FROM accounts WHERE email = $1`,
		email,
	).Scan(&a.Email, &a.Plan, &a.DataRetentionDays, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("account_repository: GetByEmail: %w", err)
	}
	return &a, nil
}

// GetPlanByEmail retorna apenas o plano de uma conta.
// Usado no hot path de criação de projetos (evita scan de colunas desnecessárias).
// Retorna "free" se o email não existir.
func (r *AccountRepository) GetPlanByEmail(ctx context.Context, email string) (string, error) {
	var plan string
	err := r.db.QueryRow(ctx,
		`SELECT plan FROM accounts WHERE email = $1`,
		email,
	).Scan(&plan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "free", nil
		}
		return "free", fmt.Errorf("account_repository: GetPlanByEmail: %w", err)
	}
	return plan, nil
}

// GetSlugByEmail retorna o slug de uma conta pelo email.
// Retorna "", nil se o email não existir ou o slug ainda for NULL.
func (r *AccountRepository) GetSlugByEmail(ctx context.Context, email string) (string, error) {
	var slug *string
	err := r.db.QueryRow(ctx,
		`SELECT slug FROM accounts WHERE email = $1`,
		email,
	).Scan(&slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("account_repository: GetSlugByEmail: %w", err)
	}
	if slug == nil {
		return "", nil
	}
	return *slug, nil
}

// UpdatePlan atualiza o plano de uma conta e recalcula data_retention_days.
// Upsert: cria a conta se não existir.
func (r *AccountRepository) UpdatePlan(ctx context.Context, email, newPlan string) error {
	if !validPlans[newPlan] {
		return fmt.Errorf("account_repository: UpdatePlan: plano inválido: %q", newPlan)
	}
	retention := accountRetentionByPlan(newPlan)
	_, err := r.db.Exec(ctx,
		`INSERT INTO accounts (email, plan, data_retention_days, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (email) DO UPDATE
		   SET plan = EXCLUDED.plan,
		       data_retention_days = EXCLUDED.data_retention_days,
		       updated_at = NOW()`,
		email, newPlan, retention,
	)
	if err != nil {
		return fmt.Errorf("account_repository: UpdatePlan: %w", err)
	}
	return nil
}
