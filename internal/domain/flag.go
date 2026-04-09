package domain

import (
	"time"

	"github.com/google/uuid"
)

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

// FeatureFlag represents a feature flag stored in the database.
type FeatureFlag struct {
	ID                string        `json:"id"`
	Key               string        `json:"key"`
	Description       string        `json:"description"`
	Enabled           bool          `json:"enabled"`
	RolloutPercentage int           `json:"rollout_percentage"` // 0-100: % do tráfego que vê a flag ativa
	Targeting         FlagTargeting `json:"targeting"`          // lista de user_ids/tenant_ids com acesso direto
	Variants          []FlagVariant `json:"variants,omitempty"` // BKL-094: variantes A/B (vazio = on/off normal)
	ProjectID         *uuid.UUID    `json:"project_id,omitempty"` // BKL-323: scope por projeto (nullable)
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// FlagsMapResponse is the public response for GET /flags — a simple key:enabled map.
// Quando user_id é fornecido, a avaliação considera rollout e targeting.
type FlagsMapResponse struct {
	Flags map[string]bool `json:"flags"`
}

// CreateFlagRequest is the payload for POST /api/v1/flags.
type CreateFlagRequest struct {
	Key               string        `json:"key"`
	Description       string        `json:"description"`
	Enabled           bool          `json:"enabled"`
	RolloutPercentage int           `json:"rollout_percentage,omitempty"`
	Targeting         FlagTargeting `json:"targeting,omitempty"`
	Variants          []FlagVariant `json:"variants,omitempty"` // BKL-094
	ProjectID         *uuid.UUID    `json:"project_id,omitempty"` // BKL-323
}

// UpdateFlagRequest is the payload for PUT /api/v1/flags/:key.
type UpdateFlagRequest struct {
	Description       *string        `json:"description"`
	Enabled           *bool          `json:"enabled"`
	RolloutPercentage *int           `json:"rollout_percentage"`
	Targeting         *FlagTargeting `json:"targeting"`
	Variants          []FlagVariant  `json:"variants"` // BKL-094 (nil = não alterar, [] = remover variantes)
}

// VariantResponse é retornada pelo endpoint GET /flags/variants. BKL-094.
type VariantResponse struct {
	Variants map[string]string `json:"variants"` // flag_key → variant_key (ex: "pricing-test" → "challenger")
}

// ErrorResponse is used for JSON error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}
