package domain

import "time"

// Project representa um projeto registrado no Pulse. BKL-099.
type Project struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Slug              string     `json:"slug"`
	APIKey            string     `json:"api_key"`
	Description       string     `json:"description,omitempty"`
	OwnerEmail        string     `json:"owner_email"`
	Plan              string     `json:"plan"`                // vem de accounts via JOIN — free|starter|growth|pro|enterprise|internal
	DataRetentionDays int        `json:"data_retention_days"` // BKL-1037: 90=free, 180=starter/growth, 365=pro/enterprise/internal
	IsActive          bool       `json:"is_active"`
	SDKConnectedAt    *time.Time `json:"sdk_connected_at"` // null = SDK ainda não conectou
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
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
	OwnerEmail  string `json:"owner_email"` // obrigatório — deve existir em accounts
}

// ProjectUpdateRequest é o payload para PUT /api/v1/projects/:slug.
// Plan foi removido: plano é atualizado via AccountRepository.UpdatePlan, não por projeto.
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
