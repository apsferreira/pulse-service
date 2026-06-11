package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// ProjectHandler gerencia endpoints CRUD de projetos. BKL-099.
type ProjectHandler struct {
	repo        *repository.ProjectRepository
	accountRepo *repository.AccountRepository
	serviceToken string
}

func NewProjectHandler(repo *repository.ProjectRepository, accountRepo *repository.AccountRepository, serviceToken string) *ProjectHandler {
	return &ProjectHandler{repo: repo, accountRepo: accountRepo, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de projetos.
func (h *ProjectHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/projects", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/", h.ListProjects)
	api.Post("/", h.CreateProject)
	api.Get("/:slug", h.GetProject)
	api.Put("/:slug", h.UpdateProject)
	api.Post("/:slug/regenerate-key", h.RegenerateAPIKey)
}

// ListProjects GET /api/v1/projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	projects, err := h.repo.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to list projects"})
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	return c.JSON(domain.ProjectListResponse{Projects: projects, Total: len(projects)})
}

// projectPlanLimit retorna o número máximo de projetos permitidos para um plano.
// -1 significa ilimitado.
var projectPlanLimit = map[string]int{
	"free":       1,
	"starter":    3,
	"growth":     10,
	"pro":        -1,
	"enterprise": -1,
	"internal":   -1,
}

// projectPlanUpgrade aponta qual plano fazer upgrade para desbloquear mais projetos.
var projectPlanUpgrade = map[string]string{
	"free":    "starter",
	"starter": "growth",
	"growth":  "pro",
}

// CreateProject POST /api/v1/projects
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req domain.ProjectCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	if req.Name == "" || req.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "name e slug são obrigatórios"})
	}
	if req.OwnerEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "owner_email é obrigatório"})
	}

	// billing: enforce limite de projetos por plano da conta
	currentPlan, planErr := h.accountRepo.GetPlanByEmail(c.Context(), req.OwnerEmail)
	if planErr != nil {
		log.Printf("[project_handler] CreateProject: GetPlanByEmail slug=%s: %v", req.Slug, planErr)
		currentPlan = "free"
	}
	count, countErr := h.repo.Count(c.Context(), req.OwnerEmail)
	if countErr == nil {
		limit, ok := projectPlanLimit[currentPlan]
		if ok && limit >= 0 && count >= limit {
			upgradeTo := projectPlanUpgrade[currentPlan]
			if upgradeTo == "" {
				upgradeTo = "pro"
			}
			return upgradeGate(c, currentPlan, upgradeTo)
		}
	}

	project, err := h.repo.Create(c.Context(), req)
	if err != nil {
		log.Printf("[project_handler] CreateProject error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to create project"})
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

// GetProject GET /api/v1/projects/:slug
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.repo.GetBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to get project"})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "project not found"})
	}
	return c.JSON(project)
}

// UpdateProject PUT /api/v1/projects/:slug
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var req domain.ProjectUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	project, err := h.repo.Update(c.Context(), slug, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to update project"})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "project not found"})
	}
	return c.JSON(project)
}

// RegenerateAPIKey POST /api/v1/projects/:slug/regenerate-key
func (h *ProjectHandler) RegenerateAPIKey(c *fiber.Ctx) error {
	slug := c.Params("slug")
	project, err := h.repo.RegenerateAPIKey(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to regenerate api key"})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "project not found"})
	}
	return c.JSON(project)
}
