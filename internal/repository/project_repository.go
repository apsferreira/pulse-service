package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

// retentionByPlan retorna o número padrão de dias de retenção por plano.
func retentionByPlan(plan string) int {
	switch plan {
	case "pro":
		return 365
	case "starter", "growth":
		return 180
	default:
		return 90
	}
}

// Create insere um novo projeto com API key gerada automaticamente.
func (r *ProjectRepository) Create(ctx context.Context, req domain.ProjectCreateRequest) (*domain.Project, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("repository: Create project: generate api key: %w", err)
	}

	plan := "free"
	p := &domain.Project{
		ID:                uuid.New().String(),
		Name:              req.Name,
		Slug:              req.Slug,
		APIKey:            apiKey,
		Description:       req.Description,
		Plan:              plan,
		DataRetentionDays: retentionByPlan(plan), // BKL-1037
		IsActive:          false, // ativado automaticamente na primeira requisição do SDK
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO projects (id, name, slug, api_key, description, plan, data_retention_days, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, p.ID, p.Name, p.Slug, p.APIKey, p.Description, p.Plan, p.DataRetentionDays, p.IsActive, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository: Create project: %w", err)
	}
	return p, nil
}

// GetBySlug retorna um projeto pelo slug.
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), COALESCE(plan, 'free'), COALESCE(data_retention_days, 90), is_active, created_at, updated_at
		FROM projects WHERE slug = $1
	`, slug)
	return scanProject(row)
}

// GetByAPIKey retorna um projeto pela API key, incluindo projetos ainda inativos (pendentes de SDK).
// O middleware ativa o projeto automaticamente na primeira requisição válida.
func (r *ProjectRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), COALESCE(plan, 'free'), COALESCE(data_retention_days, 90), is_active, created_at, updated_at
		FROM projects WHERE api_key = $1
	`, apiKey)
	return scanProject(row)
}

// ActivateByAPIKey marca o projeto como ativo (chamado na primeira requisição do SDK).
func (r *ProjectRepository) ActivateByAPIKey(ctx context.Context, apiKey string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET is_active = true, updated_at = NOW() WHERE api_key = $1 AND is_active = false`,
		apiKey,
	)
	return err
}

// List retorna todos os projetos.
func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), COALESCE(plan, 'free'), COALESCE(data_retention_days, 90), is_active, created_at, updated_at
		FROM projects ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository: List projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description, &p.Plan, &p.DataRetentionDays, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
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

// GetByID retorna um projeto pelo UUID (necessário para endpoints /api/v1/projects/:id). BKL-998.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, api_key, COALESCE(description, ''), COALESCE(plan, 'free'), COALESCE(data_retention_days, 90), is_active, created_at, updated_at
		FROM projects WHERE id = $1
	`, id)
	return scanProject(row)
}

// GetHealthConfig lê a configuração de health score de um projeto. BKL-998.
// Retorna a configuração default se a coluna ainda não tiver sido setada.
func (r *ProjectRepository) GetHealthConfig(ctx context.Context, projectID string) (domain.ProjectHealthConfig, error) {
	var raw []byte
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(health_config, '{}'::jsonb) FROM projects WHERE id = $1
	`, projectID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return defaultHealthConfig(), nil
		}
		return domain.ProjectHealthConfig{}, fmt.Errorf("project_repository: GetHealthConfig: %w", err)
	}
	var cfg domain.ProjectHealthConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return defaultHealthConfig(), nil
	}
	// Aplica defaults para campos zerados
	applyHealthConfigDefaults(&cfg)
	return cfg, nil
}

// UpdateHealthConfig persiste a configuração de health score de um projeto. BKL-998.
func (r *ProjectRepository) UpdateHealthConfig(ctx context.Context, projectID string, cfg domain.ProjectHealthConfig) error {
	applyHealthConfigDefaults(&cfg)
	raw, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("project_repository: UpdateHealthConfig marshal: %w", err)
	}
	result, err := r.db.Exec(ctx, `
		UPDATE projects SET health_config = $1, updated_at = NOW() WHERE id = $2
	`, raw, projectID)
	if err != nil {
		return fmt.Errorf("project_repository: UpdateHealthConfig: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("project_repository: UpdateHealthConfig: projeto não encontrado")
	}
	return nil
}

// defaultHealthConfig retorna a configuração padrão de health score.
func defaultHealthConfig() domain.ProjectHealthConfig {
	return domain.ProjectHealthConfig{
		Weights: domain.HealthConfigWeights{
			FlagsWithAudit:     0.3,
			ExperimentsRunning: 0.3,
			EventsTracked24h:   0.2,
			ErrorRate:          0.2,
		},
		Thresholds: domain.HealthConfigThresholds{
			Green:  80,
			Yellow: 60,
		},
	}
}

// applyHealthConfigDefaults preenche zeros com valores padrão.
func applyHealthConfigDefaults(cfg *domain.ProjectHealthConfig) {
	if cfg.Thresholds.Green == 0 {
		cfg.Thresholds.Green = 80
	}
	if cfg.Thresholds.Yellow == 0 {
		cfg.Thresholds.Yellow = 60
	}
	total := cfg.Weights.FlagsWithAudit + cfg.Weights.ExperimentsRunning +
		cfg.Weights.EventsTracked24h + cfg.Weights.ErrorRate
	if total == 0 {
		cfg.Weights = domain.HealthConfigWeights{
			FlagsWithAudit:     0.3,
			ExperimentsRunning: 0.3,
			EventsTracked24h:   0.2,
			ErrorRate:          0.2,
		}
	}
}

func scanProject(row pgx.Row) (*domain.Project, error) {
	p := &domain.Project{}
	err := row.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description, &p.Plan, &p.DataRetentionDays, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: scanProject: %w", err)
	}
	return p, nil
}
