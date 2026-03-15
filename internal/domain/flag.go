package domain

import "time"

type FeatureFlag struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsEnabled   bool      `json:"is_enabled" db:"is_enabled"`
	Product     *string   `json:"product,omitempty" db:"product"`
	RolloutPct  int       `json:"rollout_pct" db:"rollout_pct"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type FlagOverride struct {
	ID        string  `json:"id" db:"id"`
	FlagID    string  `json:"flag_id" db:"flag_id"`
	UserID    *string `json:"user_id,omitempty" db:"user_id"`
	TenantID  *string `json:"tenant_id,omitempty" db:"tenant_id"`
	IsEnabled bool    `json:"is_enabled" db:"is_enabled"`
}

type FlagEvaluation struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
	Source    string `json:"source"` // "override", "rollout", "default"
}

type FlagsResponse struct {
	Flags map[string]bool `json:"flags"`
}

type CreateFlagRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description"`
	IsEnabled   bool    `json:"is_enabled"`
	Product     *string `json:"product"`
	RolloutPct  *int    `json:"rollout_pct"`
}

type UpdateFlagRequest struct {
	Description *string `json:"description"`
	IsEnabled   *bool   `json:"is_enabled"`
	RolloutPct  *int    `json:"rollout_pct"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
