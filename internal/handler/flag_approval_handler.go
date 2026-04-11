package handler

// BKL-1079: Endpoints de approval workflow para feature flags.
// Flags em produção podem exigir aprovação antes de ativar.
// Fluxo: POST /flags/:key/approvals → revisor faz POST /approvals/:id/review
//
// Endpoints:
//   GET  /api/v1/flags/:key/approvals?project_id=<uuid>           — listar aprovações
//   GET  /api/v1/approvals/pending?project_id=<uuid>              — pendências do projeto
//   POST /api/v1/flags/:key/approvals?project_id=<uuid>           — solicitar aprovação
//   POST /api/v1/approvals/:id/review?project_id=<uuid>           — aprovar ou rejeitar

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type FlagApprovalHandler struct {
	repo         *repository.FlagApprovalRepository
	serviceToken string
}

func NewFlagApprovalHandler(repo *repository.FlagApprovalRepository, serviceToken string) *FlagApprovalHandler {
	return &FlagApprovalHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas de approval workflow.
func (h *FlagApprovalHandler) RegisterRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))

	v1.Get("/flags/:key/approvals", h.ListByFlag)
	v1.Post("/flags/:key/approvals", h.Create)
	v1.Get("/approvals/pending", h.ListPending)
	v1.Post("/approvals/:id/review", h.Review)
}

// ListByFlag godoc
// GET /api/v1/flags/:key/approvals?project_id=<uuid>
func (h *FlagApprovalHandler) ListByFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	approvals, err := h.repo.ListByFlag(c.Context(), key, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if approvals == nil {
		approvals = []domain.FlagApproval{}
	}
	return c.JSON(approvals)
}

// ListPending godoc
// GET /api/v1/approvals/pending?project_id=<uuid>
func (h *FlagApprovalHandler) ListPending(c *fiber.Ctx) error {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	approvals, err := h.repo.ListPending(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if approvals == nil {
		approvals = []domain.FlagApproval{}
	}
	return c.JSON(approvals)
}

// Create godoc
// POST /api/v1/flags/:key/approvals?project_id=<uuid>
// Body: CreateApprovalRequest
func (h *FlagApprovalHandler) Create(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "flag key inválida"})
	}
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	var req domain.CreateApprovalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}
	req.RequestedBy = strings.TrimSpace(req.RequestedBy)
	if req.RequestedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "requested_by é obrigatório"})
	}
	if len(req.Payload) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload (UpdateFlagRequest) é obrigatório"})
	}

	approval, err := h.repo.Create(c.Context(), key, projectID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(approval)
}

// Review godoc
// POST /api/v1/approvals/:id/review?project_id=<uuid>
// Body: ReviewApprovalRequest
func (h *FlagApprovalHandler) Review(c *fiber.Ctx) error {
	approvalID := c.Params("id")
	if approvalID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "approval id é obrigatório"})
	}
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	var req domain.ReviewApprovalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}
	req.ReviewedBy = strings.TrimSpace(req.ReviewedBy)
	if req.ReviewedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "reviewed_by é obrigatório"})
	}
	if req.Status != domain.ApprovalStatusApproved && req.Status != domain.ApprovalStatusRejected {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "status deve ser 'approved' ou 'rejected'",
		})
	}

	approval, err := h.repo.Review(c.Context(), approvalID, projectID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if approval == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error: "aprovação não encontrada ou já processada",
		})
	}
	return c.JSON(approval)
}
