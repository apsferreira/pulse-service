package middleware

// BKL-1029: ValidateProjectKey valida o header X-Project-Key contra a tabela projects.
// Sem esse middleware, qualquer API key acessaria qualquer projeto (cross-tenant).
//
// Uso: aplique nas rotas públicas do SDK (ingest, evaluate, flags, variants).
// O project_id resolvido é injetado como Locals("project_id") para uso nos handlers.
//
// Rotas admin (X-Service-Token) não precisam deste middleware.

import (
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// projectKeyLookup é a interface mínima necessária para resolver a API key.
type projectKeyLookup interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Project, error)
	ActivateByAPIKey(ctx context.Context, apiKey string) error
}

// ValidateProjectKey retorna middleware que valida o header X-Project-Key.
// Em caso de chave ausente ou inválida, retorna 401.
// Injeta project_id como Locals("project_id") após validação.
func ValidateProjectKey(db *pgxpool.Pool) fiber.Handler {
	repo := newProjectKeyRepo(db)
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-Project-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error: "X-Project-Key ausente — envie a API key do projeto",
			})
		}

		project, err := repo.GetByAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error: "falha ao validar project key",
			})
		}
		if project == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error: "X-Project-Key inválida ou projeto inativo",
			})
		}
		// Ativa o projeto automaticamente na primeira requisição do SDK
		if !project.IsActive {
			if activateErr := repo.ActivateByAPIKey(c.Context(), apiKey); activateErr != nil {
				log.Printf("[project_key] failed to activate project %s: %v", project.ID, activateErr)
			} else {
				project.IsActive = true
			}
		}

		c.Locals("project_id", project.ID)
		c.Locals("project", project)
		c.Locals("project_plan", project.Plan) // BKL-1077: injetado para quota + rate limit
		return c.Next()
	}
}

// projectKeyRepo é um repositório mínimo inline para evitar ciclo de dependência.
type projectKeyRepo struct {
	db *pgxpool.Pool
}

func newProjectKeyRepo(db *pgxpool.Pool) *projectKeyRepo {
	return &projectKeyRepo{db: db}
}

func (r *projectKeyRepo) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, api_key, COALESCE(description, ''), COALESCE(plan, 'free'), COALESCE(data_retention_days, 90), is_active, created_at, updated_at
		 FROM projects WHERE api_key = $1`,
		apiKey,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.Description, &p.Plan, &p.DataRetentionDays, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectKeyRepo) ActivateByAPIKey(ctx context.Context, apiKey string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET is_active = true, updated_at = NOW() WHERE api_key = $1 AND is_active = false`,
		apiKey,
	)
	return err
}

// isNoRows verifica pgx.ErrNoRows.
func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
