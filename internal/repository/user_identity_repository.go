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

// UserIdentityRepository gerencia mapeamentos cross-produto de identidade. BKL-763.
type UserIdentityRepository struct {
	db *pgxpool.Pool
}

func NewUserIdentityRepository(db *pgxpool.Pool) *UserIdentityRepository {
	return &UserIdentityRepository{db: db}
}

// Upsert insere ou atualiza o mapeamento de identidade para um produto.
// A constraint UNIQUE (product, product_user_id) garante idempotência.
func (r *UserIdentityRepository) Upsert(ctx context.Context, canonicalUserID uuid.UUID, product, productUserID, email string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_identities (canonical_user_id, product, product_user_id, email, created_at, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $5)
		ON CONFLICT (product, product_user_id)
		DO UPDATE SET
			canonical_user_id = EXCLUDED.canonical_user_id,
			email             = EXCLUDED.email,
			updated_at        = EXCLUDED.updated_at
	`, canonicalUserID, product, productUserID, email, now)
	if err != nil {
		return fmt.Errorf("repository: Upsert user_identity: %w", err)
	}
	return nil
}

// FindByCanonical retorna todos os mapeamentos de produto para um canonical_user_id.
func (r *UserIdentityRepository) FindByCanonical(ctx context.Context, canonicalUserID uuid.UUID) ([]domain.UserIdentity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, canonical_user_id, product, product_user_id, COALESCE(email, ''), created_at, updated_at
		FROM user_identities
		WHERE canonical_user_id = $1
		ORDER BY product ASC, created_at ASC
	`, canonicalUserID)
	if err != nil {
		return nil, fmt.Errorf("repository: FindByCanonical: %w", err)
	}
	defer rows.Close()

	var identities []domain.UserIdentity
	for rows.Next() {
		var ui domain.UserIdentity
		if err := rows.Scan(&ui.ID, &ui.CanonicalUserID, &ui.Product, &ui.ProductUserID, &ui.Email, &ui.CreatedAt, &ui.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: FindByCanonical scan: %w", err)
		}
		identities = append(identities, ui)
	}
	return identities, rows.Err()
}

// UpsertByProductUserID é conveniente para o fluxo de identify: recebe o product_user_id
// e o canonical_user_id como strings (sem uuid.UUID), usável pelo EventHandler sem importar uuid.
func (r *UserIdentityRepository) UpsertByProductUserID(ctx context.Context, product, productUserID, canonicalUserIDStr, email string) error {
	canonicalUserID, err := uuid.Parse(canonicalUserIDStr)
	if err != nil {
		// Se não for um UUID válido, gera um determinístico a partir do productUserID
		canonicalUserID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(product+":"+productUserID))
	}
	return r.Upsert(ctx, canonicalUserID, product, productUserID, email)
}

// FindByProduct retorna o mapeamento para um produto e product_user_id específicos.
// Retorna nil, nil quando não encontrado.
func (r *UserIdentityRepository) FindByProduct(ctx context.Context, product, productUserID string) (*domain.UserIdentity, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, canonical_user_id, product, product_user_id, COALESCE(email, ''), created_at, updated_at
		FROM user_identities
		WHERE product = $1 AND product_user_id = $2
	`, product, productUserID)

	var ui domain.UserIdentity
	err := row.Scan(&ui.ID, &ui.CanonicalUserID, &ui.Product, &ui.ProductUserID, &ui.Email, &ui.CreatedAt, &ui.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: FindByProduct: %w", err)
	}
	return &ui, nil
}
