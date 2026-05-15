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
	IsActive         bool       `json:"is_active"`
	SDKConnectedAt   *time.Time `json:"sdk_connected_at"` // null = SDK ainda não conectou
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// HealthConfigWeights define os pesos de cada dimensão no cálculo do health score. BKL-998.
// Soma dos pesos deve ser 1.0. Valores inválidos são ignorados — defaults aplicados.
type HealthConfigWeights struct {
	FlagsWithAudit      float64 `json:"flags_with_audit"`
	ExperimentsRunning  float64 `json:"experiments_running"`
	EventsTracked24h    float64 `json:"events_tracked_24h"`
	ErrorRate           float64 `json:"error_rate"`
}

// HealthConfigThresholds define os limites de status (verde/amarelo). BKL-998.
type HealthConfigThresholds struct {
	Green  float64 `json:"green"`  // score >= green → "healthy"
	Yellow float64 `json:"yellow"` // score >= yellow && < green → "at_risk"; abaixo → "critical"
}

// ProjectHealthConfig é a configuração persistida para cálculo de health score por projeto. BKL-998.
type ProjectHealthConfig struct {
	Weights    HealthConfigWeights    `json:"weights"`
	Thresholds HealthConfigThresholds `json:"thresholds"`
}

// ProjectHealthScoreResponse é a resposta do GET /api/v1/projects/:id/health. BKL-998.
type ProjectHealthScoreResponse struct {
	ProjectID  string              `json:"project_id"`
	Score      float64             `json:"score"`
	Status     string              `json:"status"`
	Dimensions map[string]float64  `json:"dimensions"` // contribuição de cada dimensão
	Config     ProjectHealthConfig `json:"config"`
	CalculatedAt time.Time         `json:"calculated_at"`
}

// UpdateHealthConfigRequest é o payload de PUT /api/v1/projects/:id/health-config. BKL-998.
type UpdateHealthConfigRequest struct {
	Weights    HealthConfigWeights    `json:"weights"`
	Thresholds HealthConfigThresholds `json:"thresholds"`
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
	Plan        *string `json:"plan,omitempty"` // BKL-billing: free|starter|growth|pro|enterprise|internal
}

// ProjectListResponse agrupa projetos com total.
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
}
