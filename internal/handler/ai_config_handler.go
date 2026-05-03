package handler

// BKL-1064: AI Configs — Feature flags para prompts e parâmetros de LLM.
// Permite que times de IA façam deploy de prompt/modelo sem redeploy de código.
// Rollback instantâneo via versioning. Diferencial único do Pulse.
//
// Endpoints:
//   GET    /api/v1/ai-configs?project_id=...      — lista configs do projeto
//   POST   /api/v1/ai-configs                     — cria nova config
//   GET    /api/v1/ai-configs/:key?project_id=... — retorna config por key (SDK usa isso)
//   PUT    /api/v1/ai-configs/:key                — atualiza config (cria versão)
//   DELETE /api/v1/ai-configs/:key                — desativa config
//   GET    /api/v1/ai-configs/:key/versions       — histórico de versões

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AIConfigHandler gerencia as AI Configs do Pulse.
type AIConfigHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewAIConfigHandler(db *pgxpool.Pool, serviceToken string) *AIConfigHandler {
	return &AIConfigHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de AI Configs.
func (h *AIConfigHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/ai-configs", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Get("/", h.List)
	api.Post("/", h.Create)
	api.Get("/:key", h.GetByKey)
	api.Put("/:key", h.Update)
	api.Delete("/:key", h.Deactivate)
	api.Get("/:key/versions", h.ListVersions)
}

// AIConfig representa uma configuração de IA.
type AIConfig struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"project_id"`
	Key                string    `json:"key"`
	Description        string    `json:"description,omitempty"`
	Model              string    `json:"model"`
	Temperature        float64   `json:"temperature"`
	MaxTokens          int       `json:"max_tokens"`
	SystemPrompt       string    `json:"system_prompt,omitempty"`
	UserPromptTemplate string    `json:"user_prompt_template,omitempty"`
	IsActive           bool      `json:"is_active"`
	Version            int       `json:"version"`
	CreatedBy          string    `json:"created_by,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AIConfigRequest é o payload para criar/atualizar uma config.
type AIConfigRequest struct {
	ProjectID          string  `json:"project_id"`
	Key                string  `json:"key"`
	Description        string  `json:"description,omitempty"`
	Model              string  `json:"model,omitempty"`
	Temperature        *float64 `json:"temperature,omitempty"`
	MaxTokens          *int    `json:"max_tokens,omitempty"`
	SystemPrompt       string  `json:"system_prompt,omitempty"`
	UserPromptTemplate string  `json:"user_prompt_template,omitempty"`
	ChangedBy          string  `json:"changed_by,omitempty"`
}

// List godoc
//
//	@Summary		Lista AI Configs do projeto
//	@Description	Retorna configs de IA (prompt, modelo, temperatura) — feature flags para LLMs
//	@Tags			ai-configs
//	@Produce		json
//	@Security		ServiceToken
//	@Param			project_id	query		string	true	"UUID do projeto"
//	@Success		200			{array}		AIConfig
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/ai-configs [get]
func (h *AIConfigHandler) List(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT id, project_id, key, COALESCE(description,''), model, temperature, max_tokens,
		       COALESCE(system_prompt,''), COALESCE(user_prompt_template,''),
		       is_active, version, COALESCE(created_by,''), created_at, updated_at
		FROM ai_configs WHERE project_id = $1 ORDER BY key
	`, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar configs"})
	}
	defer rows.Close()

	var configs []AIConfig
	for rows.Next() {
		var cfg AIConfig
		if err := rows.Scan(
			&cfg.ID, &cfg.ProjectID, &cfg.Key, &cfg.Description,
			&cfg.Model, &cfg.Temperature, &cfg.MaxTokens,
			&cfg.SystemPrompt, &cfg.UserPromptTemplate,
			&cfg.IsActive, &cfg.Version, &cfg.CreatedBy,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		); err != nil {
			continue
		}
		configs = append(configs, cfg)
	}
	if configs == nil {
		configs = []AIConfig{}
	}

	return c.JSON(fiber.Map{"configs": configs, "total": len(configs)})
}

// GetByKey godoc
//
//	@Summary		Obtém AI Config por chave
//	@Description	Retorna a config ativa para uma key — usado pelo SDK em runtime
//	@Tags			ai-configs
//	@Produce		json
//	@Security		ServiceToken
//	@Param			key			path		string	true	"Chave da AI Config"
//	@Param			project_id	query		string	true	"UUID do projeto"
//	@Success		200			{object}	AIConfig
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/ai-configs/{key} [get]
func (h *AIConfigHandler) GetByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	var cfg AIConfig
	err := h.db.QueryRow(c.Context(), `
		SELECT id, project_id, key, COALESCE(description,''), model, temperature, max_tokens,
		       COALESCE(system_prompt,''), COALESCE(user_prompt_template,''),
		       is_active, version, COALESCE(created_by,''), created_at, updated_at
		FROM ai_configs WHERE project_id = $1 AND key = $2 AND is_active = true
	`, projectID, key).Scan(
		&cfg.ID, &cfg.ProjectID, &cfg.Key, &cfg.Description,
		&cfg.Model, &cfg.Temperature, &cfg.MaxTokens,
		&cfg.SystemPrompt, &cfg.UserPromptTemplate,
		&cfg.IsActive, &cfg.Version, &cfg.CreatedBy,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "config não encontrada"})
	}

	return c.JSON(cfg)
}

// Create godoc
//
//	@Summary		Cria AI Config
//	@Description	Cria nova configuração de IA com prompt, modelo, temperatura e max_tokens
//	@Tags			ai-configs
//	@Accept			json
//	@Produce		json
//	@Security		ServiceToken
//	@Param			body	body		AIConfigRequest	true	"Dados da AI Config"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/ai-configs [post]
func (h *AIConfigHandler) Create(c *fiber.Ctx) error {
	var req AIConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.ProjectID == "" || req.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id e key são obrigatórios"})
	}

	model := req.Model
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	temperature := 0.7
	if req.Temperature != nil {
		temperature = *req.Temperature
	}
	maxTokens := 1024
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	var id string
	err := h.db.QueryRow(c.Context(), `
		INSERT INTO ai_configs (project_id, key, description, model, temperature, max_tokens, system_prompt, user_prompt_template, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, req.ProjectID, req.Key, req.Description, model, temperature, maxTokens,
		req.SystemPrompt, req.UserPromptTemplate, req.ChangedBy).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "key já existe para este projeto"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      id,
		"key":     req.Key,
		"model":   model,
		"version": 1,
		"message": "AI Config criada com sucesso",
	})
}

// Update godoc
//
//	@Summary		Atualiza AI Config
//	@Description	Atualiza prompt/modelo e salva versão anterior no histórico para rollback
//	@Tags			ai-configs
//	@Accept			json
//	@Produce		json
//	@Security		ServiceToken
//	@Param			key			path		string				true	"Chave da AI Config"
//	@Param			project_id	query		string				true	"UUID do projeto"
//	@Param			body		body		map[string]interface{}	true	"Campos a atualizar"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/ai-configs/{key} [put]
func (h *AIConfigHandler) Update(c *fiber.Ctx) error {
	key := c.Params("key")
	var req AIConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	// Lê versão atual para salvar no histórico
	var cfg AIConfig
	err := h.db.QueryRow(c.Context(), `
		SELECT id, model, temperature, max_tokens, COALESCE(system_prompt,''), COALESCE(user_prompt_template,''), version
		FROM ai_configs WHERE project_id = $1 AND key = $2
	`, req.ProjectID, key).Scan(
		&cfg.ID, &cfg.Model, &cfg.Temperature, &cfg.MaxTokens,
		&cfg.SystemPrompt, &cfg.UserPromptTemplate, &cfg.Version,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "config não encontrada"})
	}

	// Salva versão atual no histórico
	_, _ = h.db.Exec(c.Context(), `
		INSERT INTO ai_config_versions (ai_config_id, version, model, temperature, max_tokens, system_prompt, user_prompt_template, changed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, cfg.ID, cfg.Version, cfg.Model, cfg.Temperature, cfg.MaxTokens,
		cfg.SystemPrompt, cfg.UserPromptTemplate, req.ChangedBy)

	// Aplica updates (apenas campos enviados)
	newModel := cfg.Model
	if req.Model != "" {
		newModel = req.Model
	}
	newTemp := cfg.Temperature
	if req.Temperature != nil {
		newTemp = *req.Temperature
	}
	newMaxTokens := cfg.MaxTokens
	if req.MaxTokens != nil {
		newMaxTokens = *req.MaxTokens
	}
	newSystemPrompt := cfg.SystemPrompt
	if req.SystemPrompt != "" {
		newSystemPrompt = req.SystemPrompt
	}
	newUserTemplate := cfg.UserPromptTemplate
	if req.UserPromptTemplate != "" {
		newUserTemplate = req.UserPromptTemplate
	}

	_, err = h.db.Exec(c.Context(), `
		UPDATE ai_configs SET
			model = $1, temperature = $2, max_tokens = $3,
			system_prompt = $4, user_prompt_template = $5,
			version = version + 1, updated_at = NOW()
		WHERE id = $6
	`, newModel, newTemp, newMaxTokens, newSystemPrompt, newUserTemplate, cfg.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao atualizar config"})
	}

	return c.JSON(fiber.Map{
		"key":     key,
		"version": cfg.Version + 1,
		"message": "AI Config atualizada — versão anterior salva no histórico",
	})
}

// Deactivate godoc
//
//	@Summary		Desativa AI Config
//	@Description	Marca a config como inativa — rollback instantâneo sem redeploy
//	@Tags			ai-configs
//	@Produce		json
//	@Security		ServiceToken
//	@Param			key			path		string	true	"Chave da AI Config"
//	@Param			project_id	query		string	true	"UUID do projeto"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/ai-configs/{key} [delete]
func (h *AIConfigHandler) Deactivate(c *fiber.Ctx) error {
	key := c.Params("key")
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	result, err := h.db.Exec(c.Context(), `
		UPDATE ai_configs SET is_active = false, updated_at = NOW()
		WHERE project_id = $1 AND key = $2
	`, projectID, key)
	if err != nil || result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "config não encontrada"})
	}

	return c.JSON(fiber.Map{"message": "AI Config desativada"})
}

// ListVersions godoc
//
//	@Summary		Histórico de versões da AI Config
//	@Description	Retorna todas as versões anteriores da config — útil para auditoria e rollback
//	@Tags			ai-configs
//	@Produce		json
//	@Security		ServiceToken
//	@Param			key			path		string	true	"Chave da AI Config"
//	@Param			project_id	query		string	true	"UUID do projeto"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		401			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/ai-configs/{key}/versions [get]
func (h *AIConfigHandler) ListVersions(c *fiber.Ctx) error {
	key := c.Params("key")
	projectID := c.Query("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT v.id, v.version, v.model, v.temperature, v.max_tokens,
		       COALESCE(v.system_prompt,''), COALESCE(v.user_prompt_template,''),
		       COALESCE(v.changed_by,''), v.changed_at
		FROM ai_config_versions v
		JOIN ai_configs c ON c.id = v.ai_config_id
		WHERE c.project_id = $1 AND c.key = $2
		ORDER BY v.version DESC
		LIMIT 20
	`, projectID, key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao listar versões"})
	}
	defer rows.Close()

	type VersionRow struct {
		ID                 int       `json:"id"`
		Version            int       `json:"version"`
		Model              string    `json:"model"`
		Temperature        float64   `json:"temperature"`
		MaxTokens          int       `json:"max_tokens"`
		SystemPrompt       string    `json:"system_prompt,omitempty"`
		UserPromptTemplate string    `json:"user_prompt_template,omitempty"`
		ChangedBy          string    `json:"changed_by,omitempty"`
		ChangedAt          time.Time `json:"changed_at"`
	}

	var versions []VersionRow
	for rows.Next() {
		var v VersionRow
		if err := rows.Scan(
			&v.ID, &v.Version, &v.Model, &v.Temperature, &v.MaxTokens,
			&v.SystemPrompt, &v.UserPromptTemplate, &v.ChangedBy, &v.ChangedAt,
		); err != nil {
			continue
		}
		versions = append(versions, v)
	}
	if versions == nil {
		versions = []VersionRow{}
	}

	return c.JSON(fiber.Map{"versions": versions, "total": len(versions)})
}
