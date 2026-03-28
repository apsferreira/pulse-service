package domain

import "time"

// FeatureFlag represents a feature flag stored in the database.
type FeatureFlag struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FlagsMapResponse is the public response for GET /flags — a simple key:enabled map.
type FlagsMapResponse struct {
	Flags map[string]bool `json:"flags"`
}

// CreateFlagRequest is the payload for POST /api/v1/flags.
type CreateFlagRequest struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// UpdateFlagRequest is the payload for PUT /api/v1/flags/:key.
type UpdateFlagRequest struct {
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

// ErrorResponse is used for JSON error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}
