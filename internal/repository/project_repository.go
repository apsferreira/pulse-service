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
	case "pro", "enterprise", "internal":
		return 365
	case "starter", "growth":
		return 180
	default:
		return 90
	}
}

// Create insere um novo projeto com API key gerada automaticamente.
// O plano não é mais armazenado no projeto — vem de accounts via JOIN.
// owner_email é obrigatório: deve existir em accounts antes de chamar Create.
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
		OwnerEmail:  req.OwnerEmail,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO projects (id, name, slug, api_key, description, owner_email, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, p.ID, p.Name, p.Slug, p.APIKey, p.Description, p.OwnerEmail, p.IsActive, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository: Create project: %w", err)
	}

	// Busca o projeto recém-criado com o plano da conta via JOIN.
	return r.GetBySlug(ctx, p.Slug)
}

// GetBySlug retorna um projeto pelo slug com o plano vindo de accounts via JOIN.
func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, p.api_key, COALESCE(p.description, ''),
		       p.owner_email, COALESCE(a.plan, 'free'), COALESCE(a.data_retention_days, 90),
		       p.is_active, p.sdk_connected_at, p.created_at, p.updated_at
		FROM projects p
		JOIN accounts a ON a.email = p.owner_email
		WHERE p.slug = $1
	`, slug)
	return scanProject(row)
}

// GetByAPIKey retorna um projeto pela API key com o plano vindo de accounts via JOIN.
func (r *ProjectRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, p.api_key, COALESCE(p.description, ''),
		       p.owner_email, COALESCE(a.plan, 'free'), COALESCE(a.data_retention_days, 90),
		       p.is_active, p.sdk_connected_at, p.created_at, p.updated_at
		FROM projects p
		JOIN accounts a ON a.email = p.owner_email
		WHERE p.api_key = $1
	`, apiKey)
	return scanProject(row)
}

// MarkSDKConnected registra o timestamp da primeira chamada real do SDK (BKL-1117).
// Idempotente: só atualiza se sdk_connected_at ainda for NULL.
func (r *ProjectRepository) MarkSDKConnected(ctx context.Context, apiKey string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET sdk_connected_at = NOW() WHERE api_key = $1 AND sdk_connected_at IS NULL`,
		apiKey,
	)
	return err
}

// List retorna todos os projetos com plano vindo de accounts via JOIN.
func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, p.slug, p.api_key, COALESCE(p.description, ''),
		       p.owner_email, COALESCE(a.plan, 'free'), COALESCE(a.data_retention_days, 90),
		       p.is_active, p.sdk_connected_at, p.created_at, p.updated_at
		FROM projects p
		JOIN accounts a ON a.email = p.owner_email
		ORDER BY p.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("repository: List projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description,
			&p.OwnerEmail, &p.Plan, &p.DataRetentionDays,
			&p.IsActive, &p.SDKConnectedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, *p)
	}
	return projects, rows.Err()
}

// validPlans contém os planos aceitos pelo Pulse.
var validPlans = map[string]bool{
	"free":       true,
	"starter":    true,
	"growth":     true,
	"pro":        true,
	"enterprise": true,
	"internal":   true,
}

// Update atualiza nome, descrição e status de um projeto.
// Plan não é mais atualizado via projeto — use AccountRepository.UpdatePlan.
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
		UPDATE projects
		SET name = $1, description = $2, is_active = $3, updated_at = $4
		WHERE slug = $5
	`, existing.Name, existing.Description, existing.IsActive, existing.UpdatedAt, slug)
	if err != nil {
		return nil, fmt.Errorf("repository: Update project: %w", err)
	}
	return existing, nil
}

// UpdatePlan atualiza o plano da conta dona do projeto e registra audit em plan_changes.
// Chamado pelo webhook de confirmação de pagamento do checkout-service.
// O plano agora fica em accounts; o project_id é usado para resolver o owner_email.
func (r *ProjectRepository) UpdatePlan(ctx context.Context, projectID, newPlan, orderID string) error {
	if !validPlans[newPlan] {
		return fmt.Errorf("repository: UpdatePlan: plano inválido: %q", newPlan)
	}

	// Resolve owner_email e plano atual a partir do projeto.
	var ownerEmail, fromPlan string
	err := r.db.QueryRow(ctx, `
		SELECT p.owner_email, COALESCE(a.plan, 'free')
		FROM projects p
		JOIN accounts a ON a.email = p.owner_email
		WHERE p.id = $1
	`, projectID).Scan(&ownerEmail, &fromPlan)
	if err != nil {
		return fmt.Errorf("repository: UpdatePlan: buscar projeto: %w", err)
	}

	retention := retentionByPlan(newPlan)

	// Atualiza o plano na tabela accounts (fonte de verdade).
	_, err = r.db.Exec(ctx, `
		UPDATE accounts SET plan = $1, data_retention_days = $2, updated_at = NOW()
		WHERE email = $3
	`, newPlan, retention, ownerEmail)
	if err != nil {
		return fmt.Errorf("repository: UpdatePlan: atualizar account: %w", err)
	}

	// Audit trail: registra qual projeto disparou a mudança de plano.
	_, err = r.db.Exec(ctx, `
		INSERT INTO plan_changes (project_id, from_plan, to_plan, checkout_order_id)
		VALUES ($1, $2, $3, $4)
	`, projectID, fromPlan, newPlan, orderID)
	if err != nil {
		return fmt.Errorf("repository: UpdatePlan: registrar audit: %w", err)
	}

	return nil
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

// GetByID retorna um projeto pelo UUID com plano vindo de accounts via JOIN. BKL-998.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, p.api_key, COALESCE(p.description, ''),
		       p.owner_email, COALESCE(a.plan, 'free'), COALESCE(a.data_retention_days, 90),
		       p.is_active, p.sdk_connected_at, p.created_at, p.updated_at
		FROM projects p
		JOIN accounts a ON a.email = p.owner_email
		WHERE p.id = $1
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

// Count retorna o número de projetos ativos de uma conta (owner_email).
// Usado pelo gating de plano no ProjectHandler (criação de projetos).
func (r *ProjectRepository) Count(ctx context.Context, ownerEmail string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM projects WHERE owner_email = $1 AND is_active = true`,
		ownerEmail,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("project_repository: Count: %w", err)
	}
	return n, nil
}

func scanProject(row pgx.Row) (*domain.Project, error) {
	p := &domain.Project{}
	err := row.Scan(
		&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description,
		&p.OwnerEmail, &p.Plan, &p.DataRetentionDays,
		&p.IsActive, &p.SDKConnectedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: scanProject: %w", err)
	}
	return p, nil
}
