package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// MABHandler expõe endpoints de Multi-Armed Bandits epsilon-greedy. BKL-1154.
type MABHandler struct {
	repo         *repository.MABRepository
	projectRepo  *repository.ProjectRepository
	serviceToken string
}

func NewMABHandler(repo *repository.MABRepository, serviceToken string) *MABHandler {
	return &MABHandler{repo: repo, serviceToken: serviceToken}
}

// WithProjectRepo injeta o ProjectRepository para enforcement de limites por plano.
func (h *MABHandler) WithProjectRepo(repo *repository.ProjectRepository) *MABHandler {
	h.projectRepo = repo
	return h
}

// RegisterRoutes monta as rotas MAB no app Fiber.
// Prefixo: /api/v1/projects/:projectId/mab (protegido por service token)
func (h *MABHandler) RegisterRoutes(app *fiber.App) {
	base := app.Group("/api/v1/projects/:projectId/mab", middleware.ServiceTokenMiddleware(h.serviceToken))
	base.Post("/", h.CreateExperiment)
	base.Get("/", h.ListExperiments)
	base.Get("/:flagKey", h.GetExperiment)
	base.Get("/:flagKey/select", h.SelectVariant)
	base.Post("/:flagKey/events", h.RecordEvent)
	base.Post("/:flagKey/recalculate", h.RecalculateWeights)
	base.Patch("/:flagKey/active", h.SetActive)
}

// mabPlanExperimentLimit define o número máximo de experimentos MAB por plano.
// -1 significa ilimitado.
var mabPlanExperimentLimit = map[string]int{
	"free":       0,
	"starter":    2,
	"growth":     -1,
	"pro":        -1,
	"enterprise": -1,
	"internal":   -1,
}

// mabPlanUpgrade aponta qual plano fazer upgrade para desbloquear experimentos.
var mabPlanUpgrade = map[string]string{
	"free":    "starter",
	"starter": "growth",
}

// CreateExperiment godoc
// POST /api/v1/projects/:projectId/mab
func (h *MABHandler) CreateExperiment(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	var req domain.CreateMABExperimentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}
	if req.FlagKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag_key é obrigatório"})
	}
	if len(req.Variants) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "mínimo de 2 variantes"})
	}
	if req.Epsilon < 0 || req.Epsilon > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "epsilon deve estar entre 0 e 1"})
	}

	// billing: enforce limite de experimentos por plano
	if h.projectRepo != nil {
		proj, perr := h.projectRepo.GetByID(c.Context(), projectID.String())
		if perr == nil && proj != nil {
			limit, ok := mabPlanExperimentLimit[proj.Plan]
			if ok && limit == 0 {
				return upgradeGate(c, proj.Plan, "starter")
			}
			if ok && limit > 0 {
				count, cerr := h.repo.CountExperiments(c.Context(), projectID)
				if cerr == nil && count >= limit {
					upgradeTo := mabPlanUpgrade[proj.Plan]
					if upgradeTo == "" {
						upgradeTo = "growth"
					}
					return upgradeGate(c, proj.Plan, upgradeTo)
				}
			}
		}
	}

	exp, err := h.repo.CreateExperiment(c.Context(), projectID, &req)
	if err != nil {
		log.Printf("mab_handler: CreateExperiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao criar experimento MAB"})
	}
	return c.Status(fiber.StatusCreated).JSON(exp)
}

// ListExperiments godoc
// GET /api/v1/projects/:projectId/mab
func (h *MABHandler) ListExperiments(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	experiments, err := h.repo.ListExperiments(c.Context(), projectID)
	if err != nil {
		log.Printf("mab_handler: ListExperiments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao listar experimentos MAB"})
	}
	if experiments == nil {
		experiments = []domain.MABExperiment{}
	}
	return c.JSON(fiber.Map{"experiments": experiments, "total": len(experiments)})
}

// GetExperiment godoc
// GET /api/v1/projects/:projectId/mab/:flagKey
func (h *MABHandler) GetExperiment(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	exp, err := h.repo.GetExperiment(c.Context(), projectID, c.Params("flagKey"))
	if err != nil {
		log.Printf("mab_handler: GetExperiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar experimento MAB"})
	}
	if exp == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "experimento não encontrado"})
	}
	return c.JSON(exp)
}

// SelectVariant godoc
// GET /api/v1/projects/:projectId/mab/:flagKey/select
func (h *MABHandler) SelectVariant(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	exp, err := h.repo.GetExperiment(c.Context(), projectID, c.Params("flagKey"))
	if err != nil {
		log.Printf("mab_handler: SelectVariant: get experiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar experimento MAB"})
	}
	if exp == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "experimento não encontrado"})
	}
	if !exp.IsActive {
		return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{Error: "experimento inativo"})
	}

	result, err := h.repo.SelectVariant(c.Context(), exp.ID, exp.Epsilon)
	if err != nil {
		log.Printf("mab_handler: SelectVariant: epsilon-greedy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao selecionar variante"})
	}
	return c.JSON(result)
}

// RecordEvent godoc
// POST /api/v1/projects/:projectId/mab/:flagKey/events
func (h *MABHandler) RecordEvent(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	var req domain.RecordMABEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}
	if req.VariantKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "variant_key é obrigatório"})
	}

	exp, err := h.repo.GetExperiment(c.Context(), projectID, c.Params("flagKey"))
	if err != nil {
		log.Printf("mab_handler: RecordEvent: get experiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar experimento MAB"})
	}
	if exp == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "experimento não encontrado"})
	}

	if err := h.repo.RecordEvent(c.Context(), exp.ID, req.VariantKey, req.IsConversion); err != nil {
		log.Printf("mab_handler: RecordEvent: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao registrar evento MAB"})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RecalculateWeights godoc
// POST /api/v1/projects/:projectId/mab/:flagKey/recalculate
func (h *MABHandler) RecalculateWeights(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	exp, err := h.repo.GetExperiment(c.Context(), projectID, c.Params("flagKey"))
	if err != nil {
		log.Printf("mab_handler: RecalculateWeights: get experiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar experimento MAB"})
	}
	if exp == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "experimento não encontrado"})
	}

	if err := h.repo.RecalculateWeights(c.Context(), exp.ID); err != nil {
		log.Printf("mab_handler: RecalculateWeights: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao recalcular pesos MAB"})
	}

	updated, err := h.repo.GetExperimentByID(c.Context(), exp.ID)
	if err != nil {
		log.Printf("mab_handler: RecalculateWeights: reload: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao recarregar experimento"})
	}
	return c.JSON(updated)
}

// SetActive godoc
// PATCH /api/v1/projects/:projectId/mab/:flagKey/active
func (h *MABHandler) SetActive(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	var req domain.SetMABActiveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}

	exp, err := h.repo.GetExperiment(c.Context(), projectID, c.Params("flagKey"))
	if err != nil {
		log.Printf("mab_handler: SetActive: get experiment: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar experimento MAB"})
	}
	if exp == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "experimento não encontrado"})
	}

	if err := h.repo.SetActive(c.Context(), exp.ID, req.IsActive); err != nil {
		log.Printf("mab_handler: SetActive: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao atualizar status do experimento MAB"})
	}
	return c.JSON(fiber.Map{"id": exp.ID, "is_active": req.IsActive})
}
