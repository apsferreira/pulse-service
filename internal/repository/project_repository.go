package repository

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

// ProjectRepository gerencia projetos do Pulse. BKL-099.
type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pulse_" + hex.EncodeToString(b), nil
}

// Create insere um novo projeto com API key gerada automaticamente.
func (r *ProjectRepository) Create(ctx context.Context, req domain.ProjectCreateRequest) (*domain.Project, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("repository: Create project: generate api key: %w", err)
	}

	p := &domain.Project{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        req.Slug,
		APIKey:      apiKey,
		Description: req.Description,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO projects (id, name, slug, api_key, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, p.ID, p.Name, p.Slug, p.APIKey, p.Description, p.IsActive, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository: Create project: %w", err)
	}
	return p, nil
}

// GetBySlug retorna um projeto pelo slug.
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), is_active, created_at, updated_at
		FROM projects WHERE slug = $1
	`, slug)
	return scanProject(row)
}

// GetByAPIKey retorna um projeto pela API key (usado para autenticação do SDK).
func (r *ProjectRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), is_active, created_at, updated_at
		FROM projects WHERE api_key = $1 AND is_active = true
	`, apiKey)
	return scanProject(row)
}

// List retorna todos os projetos.
func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), is_active, created_at, updated_at
		FROM projects ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository: List projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, *p)
	}
	return projects, rows.Err()
}

// Update atualiza campos de um projeto.
func (r *ProjectRepository) Update(ctx context.Context, slug string, req domain.ProjectUpdateRequest) (*domain.Project, error) {
	existing, err := r.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.Exec(ctx, `
		UPDATE projects SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE slug = $5
	`, existing.Name, existing.Description, existing.IsActive, existing.UpdatedAt, slug)
	if err != nil {
		return nil, fmt.Errorf("repository: Update project: %w", err)
	}
	return existing, nil
}

// RegenerateAPIKey gera uma nova API key para um projeto.
func (r *ProjectRepository) RegenerateAPIKey(ctx context.Context, slug string) (*domain.Project, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = r.db.Exec(ctx, `
		UPDATE projects SET api_key = $1, updated_at = $2 WHERE slug = $3
	`, apiKey, now, slug)
	if err != nil {
		return nil, fmt.Errorf("repository: RegenerateAPIKey: %w", err)
	}
	return r.GetBySlug(ctx, slug)
}

func scanProject(row pgx.Row) (*domain.Project, error) {
	p := &domain.Project{}
	err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: scanProject: %w", err)
	}
	return p, nil
}
