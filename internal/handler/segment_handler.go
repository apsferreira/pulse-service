package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/jackc/pgx/v5"
)

// SegmentHandler gerencia endpoints CRUD e avaliação de segmentos avançados. BKL-1153.
type SegmentHandler struct {
	repo         *repository.SegmentRepository
	serviceToken string
}

func NewSegmentHandler(repo *repository.SegmentRepository, serviceToken string) *SegmentHandler {
	return &SegmentHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas de segmentos no app Fiber.
// Prefixo: /api/v1/projects/:projectId/segments (protegido por service token)
// Rotas de flag→segment: /api/v1/projects/:projectId/flags/:flagId/segments
func (h *SegmentHandler) RegisterRoutes(app *fiber.App) {
	base := app.Group("/api/v1/projects/:projectId/segments", middleware.ServiceTokenMiddleware(h.serviceToken))
	base.Get("/", h.ListSegments)
	base.Post("/", h.CreateSegment)
	base.Get("/:segmentId", h.GetSegment)
	base.Put("/:segmentId", h.UpdateSegment)
	base.Delete("/:segmentId", h.DeleteSegment)
	base.Post("/:segmentId/evaluate", h.EvaluateSegment)

	flags := app.Group("/api/v1/projects/:projectId/flags/:flagId/segments", middleware.ServiceTokenMiddleware(h.serviceToken))
	flags.Get("/", h.ListFlagSegments)
	flags.Post("/", h.AssignSegmentToFlag)
	flags.Delete("/:segmentId", h.UnassignSegmentFromFlag)
}

// ListSegments godoc
// GET /api/v1/projects/:projectId/segments
func (h *SegmentHandler) ListSegments(c *fiber.Ctx) error {
	projectID, err := parseUUID(c, "projectId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	segments, err := h.repo.ListByProject(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao listar segmentos"})
	}
	if segments == nil {
		segments = []domain.Segment{}
	}
	return c.JSON(fiber.Map{"segments": segments, "total": len(segments)})
}

// CreateSegment godoc
// POST /api/v1/projects/:projectId/segments
func (h *SegmentHandler) CreateSegment(c *fiber.Ctx) error {
	projectID, err := parseUUID(c, "projectId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "projectId inválido"})
	}

	var req domain.CreateSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "name é obrigatório"})
	}
	if len(req.Rules) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "rules deve ter ao menos um grupo"})
	}

	seg, err := h.repo.Create(c.Context(), projectID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao criar segmento"})
	}
	return c.Status(fiber.StatusCreated).JSON(seg)
}

// GetSegment godoc
// GET /api/v1/projects/:projectId/segments/:segmentId
func (h *SegmentHandler) GetSegment(c *fiber.Ctx) error {
	segmentID, err := parseUUID(c, "segmentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segmentId inválido"})
	}

	seg, err := h.repo.GetByID(c.Context(), segmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar segmento"})
	}
	if seg == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "segmento não encontrado"})
	}
	return c.JSON(seg)
}

// UpdateSegment godoc
// PUT /api/v1/projects/:projectId/segments/:segmentId
func (h *SegmentHandler) UpdateSegment(c *fiber.Ctx) error {
	segmentID, err := parseUUID(c, "segmentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segmentId inválido"})
	}

	var req domain.UpdateSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}

	seg, err := h.repo.Update(c.Context(), segmentID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao atualizar segmento"})
	}
	if seg == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "segmento não encontrado"})
	}
	return c.JSON(seg)
}

// DeleteSegment godoc
// DELETE /api/v1/projects/:projectId/segments/:segmentId
func (h *SegmentHandler) DeleteSegment(c *fiber.Ctx) error {
	segmentID, err := parseUUID(c, "segmentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segmentId inválido"})
	}

	if err := h.repo.Delete(c.Context(), segmentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "segmento não encontrado"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao deletar segmento"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// EvaluateSegment godoc
// POST /api/v1/projects/:projectId/segments/:segmentId/evaluate
// Body: { "context": {"user.plan": "premium", "user.country": "BR"} }
// Response: { "match": true }
func (h *SegmentHandler) EvaluateSegment(c *fiber.Ctx) error {
	segmentID, err := parseUUID(c, "segmentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segmentId inválido"})
	}

	var req domain.EvaluateSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}
	if req.Context == nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "context é obrigatório"})
	}

	seg, err := h.repo.GetByID(c.Context(), segmentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao buscar segmento"})
	}
	if seg == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "segmento não encontrado"})
	}

	match := domain.EvaluateSegment(seg, req.Context)
	return c.JSON(fiber.Map{"match": match})
}

// ListFlagSegments godoc
// GET /api/v1/projects/:projectId/flags/:flagId/segments
func (h *SegmentHandler) ListFlagSegments(c *fiber.Ctx) error {
	flagID, err := parseUUID(c, "flagId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flagId inválido"})
	}

	segments, err := h.repo.ListByFlag(c.Context(), flagID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao listar segmentos da flag"})
	}
	if segments == nil {
		segments = []domain.Segment{}
	}
	return c.JSON(fiber.Map{"segments": segments, "total": len(segments)})
}

// AssignSegmentToFlag godoc
// POST /api/v1/projects/:projectId/flags/:flagId/segments
// Body: { "segment_id": "uuid", "rollout_percentage": 100 }
func (h *SegmentHandler) AssignSegmentToFlag(c *fiber.Ctx) error {
	flagID, err := parseUUID(c, "flagId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flagId inválido"})
	}

	var req domain.AssignSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}
	if req.SegmentID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segment_id é obrigatório"})
	}
	if req.RolloutPercentage < 0 || req.RolloutPercentage > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "rollout_percentage deve ser entre 0 e 100"})
	}

	if err := h.repo.AssignToFlag(c.Context(), flagID, req.SegmentID, req.RolloutPercentage); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao associar segmento à flag"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"flag_id":            flagID,
		"segment_id":         req.SegmentID,
		"rollout_percentage": req.RolloutPercentage,
	})
}

// UnassignSegmentFromFlag godoc
// DELETE /api/v1/projects/:projectId/flags/:flagId/segments/:segmentId
func (h *SegmentHandler) UnassignSegmentFromFlag(c *fiber.Ctx) error {
	flagID, err := parseUUID(c, "flagId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flagId inválido"})
	}
	segmentID, err := parseUUID(c, "segmentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "segmentId inválido"})
	}

	if err := h.repo.UnassignFromFlag(c.Context(), flagID, segmentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "associação não encontrada"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "falha ao remover segmento da flag"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// parseUUID extrai e valida um parâmetro de rota como UUID.
func parseUUID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(param))
}
