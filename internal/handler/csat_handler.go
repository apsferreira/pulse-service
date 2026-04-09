package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// CSATHandler gerencia as rotas do modulo CSAT/NPS.
type CSATHandler struct {
	repo         *repository.CSATRepository
	serviceToken string
}

// NewCSATHandler cria um novo CSATHandler.
func NewCSATHandler(repo *repository.CSATRepository, serviceToken string) *CSATHandler {
	return &CSATHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes registra todas as rotas CSAT no app Fiber.
// Rotas /api/v1/csat/* requerem autenticacao admin.
// A rota /public/csat/:survey_id/respond e publica (widget embed).
func (h *CSATHandler) RegisterRoutes(app *fiber.App) {
	// Rotas admin — requerem X-Service-Token ou JWT
	api := app.Group("/api/v1/csat", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/surveys", h.CreateSurvey)
	api.Get("/surveys", h.ListSurveys)
	api.Put("/surveys/:id", h.UpdateSurvey)
	api.Get("/surveys/:id/nps", h.GetNPS)
	api.Get("/surveys/:id/responses", h.ListResponses)

	// Rota publica — sem autenticacao (widget embed externo)
	app.Post("/public/csat/:survey_id/respond", h.SubmitResponse)
}

// CreateSurvey godoc
// POST /api/v1/csat/surveys
// Cria uma nova survey de satisfacao.
func (h *CSATHandler) CreateSurvey(c *fiber.Ctx) error {
	var req domain.CreateSurveyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "corpo da requisicao invalido"})
	}
	if req.Product == "" || req.Question == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "product e question sao obrigatorios",
		})
	}

	surveyType := req.Type
	if surveyType == "" {
		surveyType = "nps"
	}
	scale := req.Scale
	if scale <= 0 {
		scale = 10
	}

	survey := &domain.CSATSurvey{
		Product:  req.Product,
		Question: req.Question,
		Scale:    scale,
		Type:     surveyType,
		Active:   true,
	}

	if err := h.repo.CreateSurvey(c.Context(), survey); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao criar survey"})
	}
	return c.Status(fiber.StatusCreated).JSON(survey)
}

// ListSurveys godoc
// GET /api/v1/csat/surveys?product=xxx
// Lista surveys, com filtro opcional por produto.
func (h *CSATHandler) ListSurveys(c *fiber.Ctx) error {
	product := c.Query("product")

	surveys, err := h.repo.ListSurveys(c.Context(), product)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao listar surveys"})
	}
	return c.JSON(fiber.Map{
		"surveys": surveys,
		"count":   len(surveys),
	})
}

// UpdateSurvey godoc
// PUT /api/v1/csat/surveys/:id
// Atualiza dados de uma survey existente.
func (h *CSATHandler) UpdateSurvey(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "id invalido"})
	}

	existing, err := h.repo.GetSurvey(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "survey nao encontrada"})
	}

	var req domain.CreateSurveyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "corpo da requisicao invalido"})
	}

	if req.Product != "" {
		existing.Product = req.Product
	}
	if req.Question != "" {
		existing.Question = req.Question
	}
	if req.Scale > 0 {
		existing.Scale = req.Scale
	}
	if req.Type != "" {
		existing.Type = req.Type
	}

	// Permite desativar survey via campo active no body
	type updateBody struct {
		Active *bool `json:"active"`
	}
	var body updateBody
	if err := c.BodyParser(&body); err == nil && body.Active != nil {
		existing.Active = *body.Active
	}

	if err := h.repo.UpdateSurvey(c.Context(), existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao atualizar survey"})
	}
	return c.JSON(existing)
}

// SubmitResponse godoc
// POST /public/csat/:survey_id/respond
// Endpoint publico para submissao de resposta via widget embed.
// Nao requer autenticacao.
func (h *CSATHandler) SubmitResponse(c *fiber.Ctx) error {
	surveyID, err := uuid.Parse(c.Params("survey_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "survey_id invalido"})
	}

	// Verifica se a survey existe e esta ativa
	survey, err := h.repo.GetSurvey(c.Context(), surveyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "survey nao encontrada"})
	}
	if !survey.Active {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(domain.ErrorResponse{Error: "survey inativa"})
	}

	var req domain.SubmitResponseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "corpo da requisicao invalido"})
	}
	if req.Score < 0 || req.Score > int(survey.Scale) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error: "score fora do intervalo valido para esta survey",
		})
	}

	resp := &domain.CSATResponse{
		SurveyID:   surveyID,
		Score:      req.Score,
		Comment:    req.Comment,
		Respondent: req.Respondent,
		Metadata:   req.Metadata,
	}

	if err := h.repo.SubmitResponse(c.Context(), resp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao registrar resposta"})
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetNPS godoc
// GET /api/v1/csat/surveys/:id/nps?days=30
// Retorna o calculo NPS para a survey no periodo informado.
func (h *CSATHandler) GetNPS(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "id invalido"})
	}

	from, to := parsePeriod(c)

	result, err := h.repo.GetNPS(c.Context(), id, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao calcular NPS"})
	}
	return c.JSON(result)
}

// ListResponses godoc
// GET /api/v1/csat/surveys/:id/responses?days=30&limit=100
// Lista respostas individuais de uma survey no periodo.
func (h *CSATHandler) ListResponses(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "id invalido"})
	}

	from, to := parsePeriod(c)

	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	responses, err := h.repo.ListResponses(c.Context(), id, from, to, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao listar respostas"})
	}
	return c.JSON(fiber.Map{
		"survey_id": id,
		"responses": responses,
		"count":     len(responses),
		"period": fiber.Map{
			"from": from.Format(time.RFC3339),
			"to":   to.Format(time.RFC3339),
		},
	})
}
