package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// PageViewHandler gerencia o endpoint de ingestão e analytics de page views.
type PageViewHandler struct {
	repo         *repository.PageViewRepository
	serviceToken string
}

// NewPageViewHandler cria um PageViewHandler.
func NewPageViewHandler(repo *repository.PageViewRepository, serviceToken string) *PageViewHandler {
	return &PageViewHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de page view no Fiber app.
func (h *PageViewHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Endpoint público — SDK envia diretamente sem autenticação
	api.Post("/events/page_view", h.TrackPageView)

	// Endpoints de analytics — requerem service token
	analytics := api.Group("/analytics", middleware.ServiceTokenMiddleware(h.serviceToken))
	analytics.Get("/top-pages", h.TopPages)
	analytics.Get("/user-journey", h.UserJourney)
}

// TrackPageView godoc
// POST /api/v1/events/page_view
// Recebe um page view do SDK e persiste na tabela page_views.
func (h *PageViewHandler) TrackPageView(c *fiber.Ctx) error {
	var req domain.TrackPageViewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if req.PagePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "page_path is required"})
	}
	if len(req.PagePath) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "page_path too long (max 500)"})
	}
	if len(req.SessionID) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "session_id too long (max 100)"})
	}
	if len(req.Product) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "product too long (max 50)"})
	}
	if len(req.PageTitle) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "page_title too long (max 255)"})
	}
	if len(req.Referrer) > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "referrer too long (max 500)"})
	}

	id, err := h.repo.Insert(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to persist page view"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// TopPages godoc
// GET /api/v1/analytics/top-pages?product=nitro&from=2026-01-01&to=2026-12-31&limit=10
func (h *PageViewHandler) TopPages(c *fiber.Ctx) error {
	product := c.Query("product")
	if product == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "product is required"})
	}
	if !safeProductRe.MatchString(product) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid product parameter"})
	}

	from, to, err := parseDateRangeFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	items, err := h.repo.TopPages(c.Context(), product, from, to, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to fetch top pages"})
	}

	return c.JSON(domain.TopPagesResponse{
		Product: product,
		From:    from.Format(time.RFC3339),
		To:      to.Format(time.RFC3339),
		Items:   items,
	})
}

// UserJourney godoc
// GET /api/v1/analytics/user-journey?user_id=...&product=...&limit=20
func (h *PageViewHandler) UserJourney(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "user_id is required"})
	}

	product := c.Query("product", "")
	if product != "" && !safeProductRe.MatchString(product) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid product parameter"})
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}

	items, err := h.repo.UserJourney(c.Context(), userID, product, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to fetch user journey"})
	}

	return c.JSON(domain.UserJourneyResponse{
		UserID:  userID,
		Product: product,
		Items:   items,
	})
}
