package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// UserIdentityHandler gerencia endpoints de mapeamento cross-produto. BKL-763.
type UserIdentityHandler struct {
	repo         *repository.UserIdentityRepository
	serviceToken string
}

func NewUserIdentityHandler(repo *repository.UserIdentityRepository, serviceToken string) *UserIdentityHandler {
	return &UserIdentityHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de user-identities.
func (h *UserIdentityHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/user-identities", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/", h.Upsert)
	api.Get("/:canonical_user_id", h.ListByCanonical)
}

// Upsert POST /api/v1/user-identities
// Registra ou atualiza o mapeamento de identidade para um produto.
func (h *UserIdentityHandler) Upsert(c *fiber.Ctx) error {
	var req domain.UserIdentityUpsertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if req.CanonicalUserID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "canonical_user_id é obrigatório"})
	}
	if req.Product == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "product é obrigatório"})
	}
	if len(req.Product) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "product deve ter no máximo 50 caracteres"})
	}
	if req.ProductUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "product_user_id é obrigatório"})
	}

	if err := h.repo.Upsert(c.Context(), req.CanonicalUserID, req.Product, req.ProductUserID, req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to upsert user identity"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
}

// ListByCanonical GET /api/v1/user-identities/:canonical_user_id
// Retorna todos os mapeamentos de produto para um usuário canônico.
func (h *UserIdentityHandler) ListByCanonical(c *fiber.Ctx) error {
	canonicalID, err := uuid.Parse(c.Params("canonical_user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "canonical_user_id inválido"})
	}

	identities, err := h.repo.FindByCanonical(c.Context(), canonicalID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to list user identities"})
	}
	if identities == nil {
		identities = []domain.UserIdentity{}
	}
	return c.JSON(domain.UserIdentityListResponse{
		CanonicalUserID: canonicalID,
		Identities:      identities,
		Total:           len(identities),
	})
}
