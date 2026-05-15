package domain

import (
	"time"

	"github.com/google/uuid"
)

// FlagPrerequisite define uma dependência entre flags.
// BKL-1039: flag só é avaliada como ativa se a flag prerequisito estiver ativa
// (e com a variante específica, se required_variant for informado).
type FlagPrerequisite struct {
	FlagKey         string `json:"flag_key"`                    // chave da flag da qual esta depende
	RequiredVariant string `json:"required_variant,omitempty"`  // variante obrigatória (vazio = só precisa estar enabled)
}

// FlagTargeting define regras de targeting por lista de IDs.
type FlagTargeting struct {
	UserIDs   []string `json:"user_ids,omitempty"`
	TenantIDs []string `json:"tenant_ids,omitempty"`
}

// FlagVariant define uma variante de A/B test com peso percentual.
type FlagVariant struct {
	Key    string `json:"key"`    // ex: "control", "challenger", "variant_b"
	Weight int    `json:"weight"` // percentual (0-100); soma dos pesos deve ser 100
}

// PromptVariant define a configuração de LLM para A/B test de prompts. BKL-1002.
type PromptVariant struct {
	SystemPrompt string `json:"system_prompt,omitempty"` // prompt de sistema do LLM
	Model        string `json:"model,omitempty"`         // ex: "gemini-2.5-flash-preview-04-17"
	Temperature  *float64 `json:"temperature,omitempty"` // opcional: 0.0-2.0
	MaxTokens    *int     `json:"max_tokens,omitempty"`  // opcional: limite de tokens
}

// FeatureFlag represents a feature flag stored in the database.
type FeatureFlag struct {
	ID                string             `json:"id"`
	Key               string             `json:"key"`
	Description       string             `json:"description"`
	Enabled           bool               `json:"enabled"`
	RolloutPercentage int                `json:"rollout_percentage"` // 0-100: % do tráfego que vê a flag ativa
	Targeting         FlagTargeting      `json:"targeting"`          // lista de user_ids/tenant_ids com acesso direto
	Variants          []FlagVariant      `json:"variants,omitempty"` // BKL-094: variantes A/B (vazio = on/off normal)
	Prerequisites     []FlagPrerequisite `json:"prerequisites,omitempty"` // BKL-1039: flags das quais esta depende
	ProjectID         *uuid.UUID         `json:"project_id,omitempty"` // BKL-323: scope por projeto (nullable)
	RequireApproval   bool               `json:"require_approval"`     // BKL-990: mudanças requerem aprovação
	StartsAt          *time.Time         `json:"starts_at,omitempty"`  // BKL-1040: ativação agendada (nil = imediato)
	EndsAt            *time.Time         `json:"ends_at,omitempty"`    // BKL-1040: desativação agendada (nil = sem expiração)
	PromptVariant     *PromptVariant     `json:"prompt_variant,omitempty"` // BKL-1002: config de LLM para A/B test de prompts
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// FlagsMapResponse is the public response for GET /flags — a simple key:enabled map.
// Quando user_id é fornecido, a avaliação considera rollout e targeting.
type FlagsMapResponse struct {
	Flags map[string]bool `json:"flags"`
}

// CreateFlagRequest is the payload for POST /api/v1/flags.
type CreateFlagRequest struct {
	Key               string             `json:"key"`
	Description       string             `json:"description"`
	Enabled           bool               `json:"enabled"`
	RolloutPercentage int                `json:"rollout_percentage,omitempty"`
	Targeting         FlagTargeting      `json:"targeting,omitempty"`
	Variants          []FlagVariant      `json:"variants,omitempty"`      // BKL-094
	Prerequisites     []FlagPrerequisite `json:"prerequisites,omitempty"` // BKL-1039
	ProjectID         *uuid.UUID         `json:"project_id,omitempty"`    // BKL-323
	RequireApproval   bool               `json:"require_approval,omitempty"` // BKL-990
	StartsAt          *time.Time         `json:"starts_at,omitempty"`     // BKL-1040
	EndsAt            *time.Time         `json:"ends_at,omitempty"`       // BKL-1040
}

// UpdateFlagRequest is the payload for PUT /api/v1/flags/:key.
type UpdateFlagRequest struct {
	Description       *string            `json:"description"`
	Enabled           *bool              `json:"enabled"`
	RolloutPercentage *int               `json:"rollout_percentage"`
	Targeting         *FlagTargeting     `json:"targeting"`
	Variants          []FlagVariant      `json:"variants"`                    // BKL-094 (nil = não alterar, [] = remover)
	Prerequisites     []FlagPrerequisite `json:"prerequisites,omitempty"`     // BKL-1039
	RequireApproval   *bool              `json:"require_approval,omitempty"`  // BKL-990
	RequestedBy       string             `json:"requested_by,omitempty"`      // BKL-990: obrigatório quando flag.RequireApproval=true
	StartsAt          *time.Time         `json:"starts_at,omitempty"`         // BKL-1040
	EndsAt            *time.Time         `json:"ends_at,omitempty"`           // BKL-1040
}

// VariantResponse é retornada pelo endpoint GET /flags/variants. BKL-094.
type VariantResponse struct {
	Variants map[string]string `json:"variants"` // flag_key → variant_key (ex: "pricing-test" → "challenger")
}

// FlagMetric vincula uma feature flag a um evento de negócio para cálculo de lift. BKL-1044.
type FlagMetric struct {
	ID          string    `json:"id"`
	FlagKey     string    `json:"flag_key"`
	ProjectID   string    `json:"project_id"`
	MetricName  string    `json:"metric_name"`  // nome do event_type em events
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateFlagMetricRequest é o payload para POST /api/v1/flags/:key/metrics.
type CreateFlagMetricRequest struct {
	MetricName  string `json:"metric_name"`
	Description string `json:"description"`
}

// FlagMetricLift representa o lift calculado de uma flag/variante sobre uma métrica. BKL-1044.
type FlagMetricLift struct {
	FlagKey        string  `json:"flag_key"`
	MetricName     string  `json:"metric_name"`
	ControlCount   int64   `json:"control_count"`    // usuários expostos ao controle (flag=false)
	TreatmentCount int64   `json:"treatment_count"`  // usuários expostos ao tratamento (flag=true)
	ControlConv    int64   `json:"control_conv"`     // conversões no grupo controle
	TreatmentConv  int64   `json:"treatment_conv"`   // conversões no grupo tratamento
	ControlRate    float64 `json:"control_rate"`     // taxa controle (0.0-1.0)
	TreatmentRate  float64 `json:"treatment_rate"`   // taxa tratamento (0.0-1.0)
	LiftPct        float64 `json:"lift_pct"`         // lift relativo em %
}

// FlagApprovalStatus define os estados possíveis de aprovação de uma flag. BKL-1079.
type FlagApprovalStatus string

const (
	ApprovalStatusDraft          FlagApprovalStatus = "draft"
	ApprovalStatusPendingApproval FlagApprovalStatus = "pending_approval"
	ApprovalStatusApproved       FlagApprovalStatus = "approved"
	ApprovalStatusRejected       FlagApprovalStatus = "rejected"
)

// FlagApproval representa uma solicitação de aprovação para ativar/alterar uma flag. BKL-1079.
type FlagApproval struct {
	ID          string             `json:"id"`
	FlagKey     string             `json:"flag_key"`
	ProjectID   string             `json:"project_id"`
	RequestedBy string             `json:"requested_by"`
	ReviewedBy  *string            `json:"reviewed_by,omitempty"`
	Status      FlagApprovalStatus `json:"status"`
	Notes       string             `json:"notes"`
	Payload     map[string]any     `json:"payload"` // snapshot do UpdateFlagRequest
	CreatedAt   time.Time          `json:"created_at"`
	ReviewedAt  *time.Time         `json:"reviewed_at,omitempty"`
}

// CreateApprovalRequest é o payload para POST /api/v1/flags/:key/approvals.
type CreateApprovalRequest struct {
	RequestedBy string         `json:"requested_by"`  // email ou user_id do solicitante
	Notes       string         `json:"notes"`
	Payload     map[string]any `json:"payload"` // UpdateFlagRequest a ser aprovado
}

// ReviewApprovalRequest é o payload para POST /api/v1/flags/:key/approvals/:id/review.
type ReviewApprovalRequest struct {
	ReviewedBy string             `json:"reviewed_by"`
	Status     FlagApprovalStatus `json:"status"` // approved | rejected
	Notes      string             `json:"notes"`
}

// FlagPreviewResponse é a resposta de GET /api/v1/flags/:key/preview?user_id=xxx.
// Mostra exatamente em qual grupo o usuário cai e por quê.
type FlagPreviewResponse struct {
	FlagKey           string `json:"flag_key"`
	UserID            string `json:"user_id"`
	Enabled           bool   `json:"enabled"`       // resultado final da avaliação
	Bucket            *int   `json:"bucket"`        // 0-99; nil se não houver rollout
	RolloutPercentage int    `json:"rollout_percentage"`
	InRollout         *bool  `json:"in_rollout"`    // nil se não há rollout configurado
	Variant           string `json:"variant"`       // variante selecionada; "" se on/off
	Reason            string `json:"reason"`        // "disabled|targeting_user|targeting_tenant|rollout|rollout_miss|always_on"
	IsTargeted        bool   `json:"is_targeted"`   // true se user_id está em targeting.user_ids
	PreviewURL        string `json:"preview_url"`   // URL de teste para colar no app
}

// ErrorResponse is used for JSON error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}
