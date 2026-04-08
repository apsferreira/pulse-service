package domain

import "time"

// FlagTargeting define regras de targeting por lista de IDs.
type FlagTargeting struct {
	UserIDs   []string `json:"user_ids,omitempty"`
	TenantIDs []string `json:"tenant_ids,omitempty"`
}

// FeatureFlag represents a feature flag stored in the database.
type FeatureFlag struct {
	ID                string        `json:"id"`
	Key               string        `json:"key"`
	Description       string        `json:"description"`
	Enabled           bool          `json:"enabled"`
	RolloutPercentage int           `json:"rollout_percentage"` // 0-100: % do tráfego que vê a flag ativa
	Targeting         FlagTargeting `json:"targeting"`          // lista de user_ids/tenant_ids com acesso direto
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
}

// UpdateFlagRequest is the payload for PUT /api/v1/flags/:key.
type UpdateFlagRequest struct {
	Description       *string        `json:"description"`
	Enabled           *bool          `json:"enabled"`
	RolloutPercentage *int           `json:"rollout_percentage"`
	Targeting         *FlagTargeting `json:"targeting"`
}

// ErrorResponse is used for JSON error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}
