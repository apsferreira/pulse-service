package handler

import (
	"errors"
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

// validKeyPattern allows only alphanumeric, underscores, hyphens and dots.
var validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,100}$`)

type FlagHandler struct {
	svc            service.FlagServicer
	flagRepo       *repository.FlagRepository    // BKL-1002: acesso direto para GetByKeyWithPrompt
	projectRepo    *repository.ProjectRepository // billing: gating por plano
	serviceToken   string
	sseBroker      *FlagSSEBroker  // BKL-991: notifica subscribers SSE após updates
	webhookHandler *WebhookHandler // BKL-1078: dispara webhook flag.toggled ao atualizar flag
}

func NewFlagHandler(svc service.FlagServicer, serviceToken string) *FlagHandler {
	return &FlagHandler{svc: svc, serviceToken: serviceToken}
}

// WithFlagRepo injeta o FlagRepository para operações que precisam de acesso direto. BKL-1002.
func (h *FlagHandler) WithFlagRepo(repo *repository.FlagRepository) *FlagHandler {
	h.flagRepo = repo
	return h
}

// WithProjectRepo injeta o ProjectRepository para enforcement de limites por plano.
func (h *FlagHandler) WithProjectRepo(repo *repository.ProjectRepository) *FlagHandler {
	h.projectRepo = repo
	return h
}

// WithSSEBroker injeta o broker SSE para notificações em tempo real. BKL-991.
func (h *FlagHandler) WithSSEBroker(broker *FlagSSEBroker) *FlagHandler {
	h.sseBroker = broker
	return h
}

// WithWebhookHandler injeta o WebhookHandler para disparar webhooks flag.toggled. BKL-1078.
func (h *FlagHandler) WithWebhookHandler(wh *WebhookHandler) *FlagHandler {
	h.webhookHandler = wh
	return h
}

// RegisterRoutes mounts all flag routes onto the given Fiber app.
func (h *FlagHandler) RegisterRoutes(app *fiber.App) {
	// Public endpoints — no auth required (consumed by SDKs)
	app.Get("/flags", h.GetFlagsMap)
	app.Post("/flags/evaluate-batch", h.EvaluateBatch) // BKL-093: batch evaluation
	app.Get("/flags/variants", h.GetVariants)           // BKL-094: A/B variant selection

	// Admin CRUD endpoints — require service token
	api := app.Group("/api/v1/flags", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/", h.CreateFlag)
	api.Get("/", h.ListFlags)
	api.Get("/:key/evaluate", h.EvaluateFlag) // BKL-1002: evaluate com prompt_variant
	api.Get("/:key", h.GetFlagByKey)
	api.Put("/:key", h.UpdateFlag)
	api.Delete("/:key", h.DeleteFlag)
}

// GetFlagsMap godoc
// GET /flags — public, returns {flags: {key: enabled}} for SDK consumption.
// Query params opcionais: user_id, tenant_id — quando fornecidos, avalia rollout e targeting.
// Query param opcional: project_id (UUID) — isola flags por tenant. BKL-323, BKL-1027, BKL-1028.
func (h *FlagHandler) GetFlagsMap(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	tenantID := c.Query("tenant_id")
	projectID := c.Query("project_id")

	// Se tem contexto de usuário, avalia com rollout + targeting (filtrado por project)
	if userID != "" || tenantID != "" {
		resp, err := h.svc.EvaluateFlags(c.Context(), userID, tenantID, projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to evaluate flags"})
		}
		return c.JSON(resp)
	}

	// Sem contexto — retorna mapa simples (cache Redis por project_id)
	resp, err := h.svc.GetFlagsMap(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to load flags"})
	}
	return c.JSON(resp)
}

// ListFlags godoc
// GET /api/v1/flags — returns all flags with full details.
// Query param opcional: project_id (UUID) — filtra flags pelo projeto. BKL-323.
func (h *FlagHandler) ListFlags(c *fiber.Ctx) error {
	projectID := c.Query("project_id")

	var (
		flags []domain.FeatureFlag
		err   error
	)
	if projectID != "" {
		flags, err = h.svc.ListByProject(c.Context(), projectID)
	} else {
		flags, err = h.svc.List(c.Context())
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to list flags"})
	}
	if flags == nil {
		flags = []domain.FeatureFlag{}
	}
	return c.JSON(flags)
}

// GetFlagByKey godoc
// GET /api/v1/flags/:key
func (h *FlagHandler) GetFlagByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	flag, err := h.svc.GetByKey(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to get flag"})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}
	return c.JSON(flag)
}

// freePlanFlagLimit é o número máximo de feature flags permitidas no plano free.
const freePlanFlagLimit = 5

// CreateFlag godoc
// POST /api/v1/flags
func (h *FlagHandler) CreateFlag(c *fiber.Ctx) error {
	var req domain.CreateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	if !validKeyPattern.MatchString(req.Key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "key must be 2-100 chars, alphanumeric/underscore/hyphen/dot only"})
	}

	// billing: enforce limite de flags por plano quando project_id informado
	if req.ProjectID != nil && h.flagRepo != nil && h.projectRepo != nil {
		projectIDStr := req.ProjectID.String()
		proj, err := h.projectRepo.GetByID(c.Context(), projectIDStr)
		if err == nil && proj != nil && proj.Plan == "free" {
			count, cerr := h.flagRepo.CountByProject(c.Context(), projectIDStr)
			if cerr == nil && count >= freePlanFlagLimit {
				return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
					"error":       "Limite de flags atingido no plano Free. Faça upgrade.",
					"upgrade_url": "/checkout?plan=starter",
				})
			}
		}
	}

	flag, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{Error: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(flag)
}

// UpdateFlag godoc
// PUT /api/v1/flags/:key
func (h *FlagHandler) UpdateFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	var req domain.UpdateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}

	flag, err := h.svc.Update(c.Context(), key, req)
	if err != nil {
		// BKL-990: flag requer aprovação — retorna 202 com o approval criado
		var approvalErr *service.ErrApprovalRequired
		if errors.As(err, &approvalErr) {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"message":  "mudança pendente de aprovação",
				"approval": approvalErr.Approval,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to update flag"})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}

	// BKL-991: notifica subscribers SSE sobre a atualização da flag
	if h.sseBroker != nil && flag.ProjectID != nil {
		h.sseBroker.Notify(flag.ProjectID.String(), FlagUpdateEvent{
			FlagKey:   flag.Key,
			Enabled:   flag.Enabled,
			UpdatedAt: flag.UpdatedAt,
		})
	}

	// BKL-1078: dispara webhook flag.toggled para integrações externas (Slack, Datadog, n8n, etc.)
	if h.webhookHandler != nil && flag.ProjectID != nil {
		go h.webhookHandler.Dispatch(c.Context(), flag.ProjectID.String(), "flag.toggled", map[string]interface{}{
			"flag_key": flag.Key,
			"enabled":  flag.Enabled,
		})
	}

	return c.JSON(flag)
}

// DeleteFlag godoc
// DELETE /api/v1/flags/:key
func (h *FlagHandler) DeleteFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	if err := h.svc.Delete(c.Context(), key); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to delete flag"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// EvaluateBatch godoc
// POST /flags/evaluate-batch — BKL-093: avalia N flags de uma vez para SDKs.
// Body: {"keys": ["flag-a", "flag-b"], "user_id": "...", "tenant_id": "...", "project_id": "..."}
// BKL-1027: project_id obrigatório para isolamento multi-tenant.
// Response: {"flags": {"flag-a": true, "flag-b": false}}
func (h *FlagHandler) EvaluateBatch(c *fiber.Ctx) error {
	var req struct {
		Keys      []string `json:"keys"`
		UserID    string   `json:"user_id"`
		TenantID  string   `json:"tenant_id"`
		ProjectID string   `json:"project_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid request body"})
	}
	if len(req.Keys) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "keys must not be empty"})
	}
	if len(req.Keys) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "max 50 keys per batch"})
	}

	// Evaluate all flags with context (filtrado por project_id)
	allFlags, err := h.svc.EvaluateFlags(c.Context(), req.UserID, req.TenantID, req.ProjectID)
	if err != nil {
		// Fail-open: return all false defaults
		result := make(map[string]bool, len(req.Keys))
		for _, k := range req.Keys {
			result[k] = false
		}
		return c.JSON(domain.FlagsMapResponse{Flags: result})
	}

	// Filter to only requested keys
	result := make(map[string]bool, len(req.Keys))
	for _, k := range req.Keys {
		result[k] = allFlags.Flags[k]
	}
	return c.JSON(domain.FlagsMapResponse{Flags: result})
}

// EvaluateFlag godoc
// GET /api/v1/flags/:key/evaluate
// Retorna a flag com campos de avaliação incluindo prompt_variant quando presente. BKL-1002.
// Query params opcionais: user_id, tenant_id — para avaliação de rollout/targeting.
// Response: {enabled, variant, prompt_variant}
func (h *FlagHandler) EvaluateFlag(c *fiber.Ctx) error {
	key := c.Params("key")
	if !validKeyPattern.MatchString(key) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "invalid flag key"})
	}

	// Busca flag com prompt_variant se o repo estiver disponível
	var flag *domain.FeatureFlag
	var err error
	if h.flagRepo != nil {
		flag, err = h.flagRepo.GetByKeyWithPrompt(c.Context(), key)
	} else {
		flag, err = h.svc.GetByKey(c.Context(), key)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to evaluate flag"})
	}
	if flag == nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{Error: "flag not found"})
	}

	// Determina variante para o user_id fornecido (se houver variantes configuradas)
	variant := ""
	userID := c.Query("user_id")
	if userID != "" && len(flag.Variants) > 0 {
		// Avaliação determinística de variante pelo user_id
		variantResp, verr := h.svc.GetVariants(c.Context(), userID, c.Query("tenant_id"), "")
		if verr == nil {
			variant = variantResp.Variants[key]
		}
	}

	resp := fiber.Map{
		"enabled": flag.Enabled,
		"variant": variant,
	}
	if flag.PromptVariant != nil {
		resp["prompt_variant"] = flag.PromptVariant
	}

	return c.JSON(resp)
}

// GetVariants godoc
// GET /flags/variants — public, returns A/B variant for each flag with variants configured.
// Query params: user_id (required for deterministic selection), tenant_id (optional), project_id (required for isolation)
// BKL-094, BKL-1027
// Response: {"variants": {"flag-key": "variant-name"}}
func (h *FlagHandler) GetVariants(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	tenantID := c.Query("tenant_id")
	projectID := c.Query("project_id")

	resp, err := h.svc.GetVariants(c.Context(), userID, tenantID, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to get variants"})
	}

	return c.JSON(resp)
}
