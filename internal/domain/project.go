package domain

import "time"

// Project representa um projeto registrado no Pulse. BKL-099.
type Project struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Slug              string    `json:"slug"`
	APIKey            string    `json:"api_key"`
	Description       string    `json:"description,omitempty"`
	Plan              string    `json:"plan"`                // BKL-1077: free|starter|growth|pro
	DataRetentionDays int       `json:"data_retention_days"` // BKL-1037: 90=free, 180=starter/growth, 365=pro
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ProjectCreateRequest é o payload para POST /api/v1/projects.
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

// ProjectUpdateRequest é o payload para PUT /api/v1/projects/:slug.
type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// ProjectListResponse agrupa projetos com total.
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
}
