package repository

// BKL-1075: EnvironmentRepository gerencia ambientes por projeto.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentRepository struct {
	db *pgxpool.Pool
}

func NewEnvironmentRepository(db *pgxpool.Pool) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

// DB expõe o pool para consultas ad-hoc (ex: resolveProjectID no handler).
func (r *EnvironmentRepository) DB() *pgxpool.Pool {
	return r.db
}

func generateEnvAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "env_" + hex.EncodeToString(b), nil
}

// Create cria um novo ambiente para o projeto com API key gerada automaticamente.
func (r *EnvironmentRepository) Create(ctx context.Context, projectID string, req domain.EnvironmentCreateRequest) (*domain.Environment, error) {
	apiKey, err := generateEnvAPIKey()
	if err != nil {
		return nil, fmt.Errorf("environment: Create: generate api key: %w", err)
	}

	env := &domain.Environment{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Name:      req.Name,
		APIKey:    apiKey,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO environments (id, project_id, name, api_key, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, env.ID, env.ProjectID, env.Name, env.APIKey, env.IsActive, env.CreatedAt, env.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("environment: Create: %w", err)
	}
	return env, nil
}

// ListByProject retorna todos os ambientes de um projeto.
func (r *EnvironmentRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Environment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_id, name, api_key, is_active, created_at, updated_at
		FROM environments
		WHERE project_id = $1
		ORDER BY CASE name
			WHEN 'production'  THEN 1
			WHEN 'staging'     THEN 2
			WHEN 'development' THEN 3
			ELSE 4
		END
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("environment: ListByProject: %w", err)
	}
	defer rows.Close()

	var envs []domain.Environment
	for rows.Next() {
		e := domain.Environment{}
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.APIKey, &e.IsActive, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

// GetByAPIKey resolve ambiente pela API key (para autenticação do SDK por ambiente).
func (r *EnvironmentRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Environment, error) {
	var e domain.Environment
	err := r.db.QueryRow(ctx, `
		SELECT id, project_id, name, api_key, is_active, created_at, updated_at
		FROM environments
		WHERE api_key = $1 AND is_active = true
	`, apiKey).Scan(&e.ID, &e.ProjectID, &e.Name, &e.APIKey, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("environment: GetByAPIKey: %w", err)
	}
	return &e, nil
}

// RegenerateAPIKey gera nova API key para o ambiente.
func (r *EnvironmentRepository) RegenerateAPIKey(ctx context.Context, envID string) (*domain.Environment, error) {
	apiKey, err := generateEnvAPIKey()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = r.db.Exec(ctx, `
		UPDATE environments SET api_key = $1, updated_at = $2 WHERE id = $3
	`, apiKey, now, envID)
	if err != nil {
		return nil, fmt.Errorf("environment: RegenerateAPIKey: %w", err)
	}

	var e domain.Environment
	err = r.db.QueryRow(ctx, `
		SELECT id, project_id, name, api_key, is_active, created_at, updated_at
		FROM environments WHERE id = $1
	`, envID).Scan(&e.ID, &e.ProjectID, &e.Name, &e.APIKey, &e.IsActive, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("environment: RegenerateAPIKey get: %w", err)
	}
	return &e, nil
}
