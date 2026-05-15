package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LicenseRepository gerencia licenças self-hosted do Pulse.
type LicenseRepository struct {
	db *pgxpool.Pool
}

func NewLicenseRepository(db *pgxpool.Pool) *LicenseRepository {
	return &LicenseRepository{db: db}
}

// Create insere uma nova licença no banco.
func (r *LicenseRepository) Create(ctx context.Context, lic domain.LicenseKey) (*domain.LicenseKey, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO license_keys (id, key, plan, customer_email, customer_name, order_id, valid_from, valid_until, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), $6, NOW())
		RETURNING id, key, plan, customer_email, COALESCE(customer_name, ''), COALESCE(order_id, ''),
		          valid_from, valid_until, activated_at, COALESCE(activation_hostname, ''), revoked_at, created_at
	`,
		lic.Key,
		lic.Plan,
		lic.CustomerEmail,
		lic.CustomerName,
		lic.OrderID,
		lic.ValidUntil,
	)

	return scanLicense(row)
}

// GetByKey busca uma licença pela chave.
func (r *LicenseRepository) GetByKey(ctx context.Context, key string) (*domain.LicenseKey, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, key, plan, customer_email, COALESCE(customer_name, ''), COALESCE(order_id, ''),
		       valid_from, valid_until, activated_at, COALESCE(activation_hostname, ''), revoked_at, created_at
		FROM license_keys
		WHERE key = $1
	`, key)
	return scanLicense(row)
}

// Activate preenche activated_at e activation_hostname se ainda não ativada (idempotente).
func (r *LicenseRepository) Activate(ctx context.Context, key, hostname string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE license_keys
		SET activated_at = NOW(), activation_hostname = $2
		WHERE key = $1 AND activated_at IS NULL
	`, key, hostname)
	if err != nil {
		return fmt.Errorf("license_repository: Activate: %w", err)
	}
	return nil
}

// Revoke preenche revoked_at de uma licença.
func (r *LicenseRepository) Revoke(ctx context.Context, key string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE license_keys SET revoked_at = NOW() WHERE key = $1 AND revoked_at IS NULL
	`, key)
	if err != nil {
		return fmt.Errorf("license_repository: Revoke: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("license_repository: Revoke: licença não encontrada ou já revogada")
	}
	return nil
}

func scanLicense(row pgx.Row) (*domain.LicenseKey, error) {
	lic := &domain.LicenseKey{}
	err := row.Scan(
		&lic.ID,
		&lic.Key,
		&lic.Plan,
		&lic.CustomerEmail,
		&lic.CustomerName,
		&lic.OrderID,
		&lic.ValidFrom,
		&lic.ValidUntil,
		&lic.ActivatedAt,
		&lic.ActivationHostname,
		&lic.RevokedAt,
		&lic.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("license_repository: scan: %w", err)
	}
	return lic, nil
}
