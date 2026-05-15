package domain

import "time"

// Account representa a conta de um owner no Pulse.
// O plano é definido por conta (email), não por projeto individual.
type Account struct {
	Email             string    `json:"email"`
	Plan              string    `json:"plan"`
	DataRetentionDays int       `json:"data_retention_days"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AccountUpdateRequest é o payload para atualização parcial de uma conta.
type AccountUpdateRequest struct {
	Plan *string `json:"plan,omitempty"`
}
