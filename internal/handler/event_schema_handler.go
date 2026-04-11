package handler

// BKL-1080: Endpoints do schema registry de eventos.
// Permite definir schemas para event_types e validar eventos ingeridos.
//
// Endpoints:
//   GET    /api/v1/event-schemas?project_id=<uuid>                 — listar schemas
//   GET    /api/v1/event-schemas/:event_name?project_id=<uuid>     — obter schema
//   POST   /api/v1/event-schemas?project_id=<uuid>                 — criar/atualizar schema
//   DELETE /api/v1/event-schemas/:event_name?project_id=<uuid>     — remover schema
//   POST   /api/v1/event-schemas/:event_name/validate?project_id=<uuid> — validar payload

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type EventSchemaHandler struct {
	repo         *repository.EventSchemaRepository
	serviceToken string
}

func NewEventSchemaHandler(repo *repository.EventSchemaRepository, serviceToken string) *EventSchemaHandler {
	return &EventSchemaHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes monta as rotas do schema registry.
func (h *EventSchemaHandler) RegisterRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1", middleware.ServiceTokenMiddleware(h.serviceToken))

	v1.Get("/event-schemas", h.List)
	v1.Get("/event-schemas/:event_name", h.Get)
	v1.Post("/event-schemas", h.Create)
	v1.Delete("/event-schemas/:event_name", h.Delete)
	v1.Post("/event-schemas/:event_name/validate", h.Validate)
}

// List godoc
// GET /api/v1/event-schemas?project_id=<uuid>
func (h *EventSchemaHandler) List(c *fiber.Ctx) error {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	schemas, err := h.repo.List(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if schemas == nil {
		schemas = []domain.EventSchema{}
	}
	return c.JSON(schemas)
}

// Get godoc
// GET /api/v1/event-schemas/:event_name?project_id=<uuid>
func (h *EventSchemaHandler) Get(c *fiber.Ctx) error {
	eventName := c.Params("event_name")
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	schema, err := h.repo.GetByEventName(c.Context(), projectID, eventName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if schema == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "schema não encontrado"})
	}
	return c.JSON(schema)
}

// Create godoc
// POST /api/v1/event-schemas?project_id=<uuid>
// Body: CreateEventSchemaRequest
func (h *EventSchemaHandler) Create(c *fiber.Ctx) error {
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	var req domain.CreateEventSchemaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}
	req.EventName = strings.TrimSpace(req.EventName)
	if req.EventName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "event_name é obrigatório"})
	}
	if len(req.EventName) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "event_name deve ter no máximo 100 caracteres"})
	}

	// Valida tipos de propriedades
	validTypes := map[domain.EventPropertyType]bool{
		domain.PropertyTypeString:  true,
		domain.PropertyTypeNumber:  true,
		domain.PropertyTypeBoolean: true,
		domain.PropertyTypeObject:  true,
		domain.PropertyTypeArray:   true,
	}
	for _, p := range req.Properties {
		if !validTypes[p.Type] {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
				Error: "tipo inválido para propriedade '" + p.Name + "': use string, number, boolean, object ou array",
			})
		}
	}

	schema, err := h.repo.Create(c.Context(), projectID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(schema)
}

// Delete godoc
// DELETE /api/v1/event-schemas/:event_name?project_id=<uuid>
func (h *EventSchemaHandler) Delete(c *fiber.Ctx) error {
	eventName := c.Params("event_name")
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	if err := h.repo.Delete(c.Context(), projectID, eventName); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Validate godoc
// POST /api/v1/event-schemas/:event_name/validate?project_id=<uuid>
// Body: {"properties": {...}}
// Response: 200 se válido, 422 com erros se inválido, 404 se schema não existe
func (h *EventSchemaHandler) Validate(c *fiber.Ctx) error {
	eventName := c.Params("event_name")
	projectID := strings.TrimSpace(c.Query("project_id"))
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id é obrigatório"})
	}

	schema, err := h.repo.GetByEventName(c.Context(), projectID, eventName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	if schema == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "schema não encontrado — nenhuma validação aplicada"})
	}

	var body struct {
		Properties map[string]any `json:"properties"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}

	errs := domain.ValidateAgainstSchema(*schema, body.Properties)
	if len(errs) > 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"valid":  false,
			"errors": errs,
		})
	}
	return c.JSON(fiber.Map{"valid": true})
}
