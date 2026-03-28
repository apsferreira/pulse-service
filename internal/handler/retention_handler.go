package handler

import (
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

var safeProductRe = regexp.MustCompile(`^[a-z0-9_-]{0,50}$`)

type RetentionHandler struct {
	svc          *service.RetentionService
	serviceToken string
}

func NewRetentionHandler(svc *service.RetentionService, serviceToken string) *RetentionHandler {
	return &RetentionHandler{svc: svc, serviceToken: serviceToken}
}

// RegisterRoutes mounts retention metric routes onto the given Fiber app.
func (h *RetentionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/metrics", middleware.ServiceTokenMiddleware(h.serviceToken))

	api.Get("/retention", h.GetRetention)
	api.Get("/cohorts", h.GetCohorts)
	api.Get("/active-users", h.GetActiveUsers)
}

func parseDate(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t, err = time.Parse("2006-01-02", raw)
	}
	return t, err
}

func sanitizeProduct(c *fiber.Ctx) (string, error) {
	product := c.Query("product")
	if product != "" && !safeProductRe.MatchString(product) {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid product parameter")
	}
	return product, nil
}

// GetRetention godoc
// GET /api/v1/metrics/retention?product=libri&start_date=2026-01-01&end_date=2026-03-28
func (h *RetentionHandler) GetRetention(c *fiber.Ctx) error {
	product, err := sanitizeProduct(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	result, err := h.svc.GetRetention(c.Context(), product, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(result)
}

// GetCohorts godoc
// GET /api/v1/metrics/cohorts?product=libri&granularity=week
func (h *RetentionHandler) GetCohorts(c *fiber.Ctx) error {
	product, err := sanitizeProduct(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	granularity := c.Query("granularity", "week")
	if granularity != "week" && granularity != "month" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "granularity must be 'week' or 'month'"})
	}

	result, err := h.svc.GetCohorts(c.Context(), product, granularity)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(result)
}

// GetActiveUsers godoc
// GET /api/v1/metrics/active-users?product=libri&start_date=2026-01-01&end_date=2026-03-28
func (h *RetentionHandler) GetActiveUsers(c *fiber.Ctx) error {
	product, err := sanitizeProduct(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	result, err := h.svc.GetActiveUsers(c.Context(), product, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(result)
}

func parseDateRange(c *fiber.Ctx) (time.Time, time.Time, error) {
	sdRaw := c.Query("start_date")
	edRaw := c.Query("end_date")

	if sdRaw == "" || edRaw == "" {
		// Default: last 30 days
		now := time.Now().UTC()
		return now.AddDate(0, 0, -30), now, nil
	}

	startDate, err := parseDate(sdRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid start_date format, use RFC3339 or YYYY-MM-DD")
	}

	endDate, err := parseDate(edRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid end_date format, use RFC3339 or YYYY-MM-DD")
	}

	// If end_date is just a date, set to end of day
	if len(edRaw) == 10 {
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	return startDate, endDate, nil
}
