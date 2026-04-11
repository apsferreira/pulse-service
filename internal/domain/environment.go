package domain

import "time"

// Environment representa um ambiente de projeto no Pulse. BKL-1075.
// Cada projeto tem até 4 ambientes: production, staging, development, test.
// Cada ambiente tem sua própria API key — flags podem estar ativas em staging e off em prod.
type Environment struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`      // production|staging|development|test
	APIKey    string    `json:"api_key"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnvironmentCreateRequest payload para POST /api/v1/projects/:slug/environments.
type EnvironmentCreateRequest struct {
	Name string `json:"name" validate:"required,oneof=production staging development test"`
}

// EnvironmentListResponse agrupa ambientes de um projeto.
type EnvironmentListResponse struct {
	Environments []Environment `json:"environments"`
	Total        int           `json:"total"`
}
