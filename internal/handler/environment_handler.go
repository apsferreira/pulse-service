package handler

// BKL-1075: Handler de ambientes por projeto.
// CRUD de environments + regeneração de API key por ambiente.

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type EnvironmentHandler struct {
	repo         *repository.EnvironmentRepository
	serviceToken string
}

func NewEnvironmentHandler(repo *repository.EnvironmentRepository, serviceToken string) *EnvironmentHandler {
	return &EnvironmentHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta rotas de ambientes abaixo de /api/v1/projects/:slug/environments.
// Todas as rotas são admin (requerem X-Service-Token).
func (h *EnvironmentHandler) RegisterRoutes(app *fiber.App) {
	auth := middleware.ServiceTokenMiddleware(h.serviceToken)

	api := app.Group("/api/v1/projects/:slug/environments", auth)
	api.Get("/", h.List)
	api.Post("/", h.Create)
	api.Post("/:env_id/regenerate-key", h.RegenerateKey)
}

// List godoc
// GET /api/v1/projects/:slug/environments
func (h *EnvironmentHandler) List(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "slug obrigatório"})
	}

	// Resolve project_id a partir do slug (reutiliza project repository inline)
	projectID, err := h.resolveProjectID(c, slug)
	if err != nil || projectID == "" {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "projeto não encontrado"})
	}

	envs, err := h.repo.ListByProject(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if envs == nil {
		envs = []domain.Environment{}
	}
	return c.JSON(domain.EnvironmentListResponse{
		Environments: envs,
		Total:        len(envs),
	})
}

// Create godoc
// POST /api/v1/projects/:slug/environments
func (h *EnvironmentHandler) Create(c *fiber.Ctx) error {
	slug := c.Params("slug")
	projectID, err := h.resolveProjectID(c, slug)
	if err != nil || projectID == "" {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "projeto não encontrado"})
	}

	var req domain.EnvironmentCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "name obrigatório: production|staging|development|test"})
	}

	env, err := h.repo.Create(c.Context(), projectID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(env)
}

// RegenerateKey godoc
// POST /api/v1/projects/:slug/environments/:env_id/regenerate-key
func (h *EnvironmentHandler) RegenerateKey(c *fiber.Ctx) error {
	envID := c.Params("env_id")
	if envID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "env_id obrigatório"})
	}

	env, err := h.repo.RegenerateAPIKey(c.Context(), envID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(env)
}

// resolveProjectID é um helper inline para obter project_id via slug.
// Evita ciclo de import buscando direto no DB via pgxpool disponível no repo.
func (h *EnvironmentHandler) resolveProjectID(c *fiber.Ctx, slug string) (string, error) {
	var id string
	err := h.repo.DB().QueryRow(c.Context(),
		`SELECT id FROM projects WHERE slug = $1 AND is_active = true`, slug,
	).Scan(&id)
	return id, err
}
