package domain

import "time"

// TenantHealthScore representa o health score calculado por tenant.
// Score 0-100 com pesos: login 30% + content 40% + adoption 20% + subscription 10%.
// BKL-801: fundação para churn prevention.
type TenantHealthScore struct {
	ID                string     `json:"id"`
	TenantID          string     `json:"tenant_id"`
	Score             int        `json:"score"`              // 0-100 (total)
	LoginScore        int        `json:"login_score"`        // 0-30
	ContentScore      int        `json:"content_score"`      // 0-40
	AdoptionScore     int        `json:"adoption_score"`     // 0-20
	SubscriptionScore int        `json:"subscription_score"` // 0-10
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LastContentAt     *time.Time `json:"last_content_at,omitempty"`
	CalculatedAt      time.Time  `json:"calculated_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Status classifica o risco de churn baseado no score.
func (h *TenantHealthScore) Status() string {
	switch {
	case h.Score >= 70:
		return "healthy"
	case h.Score >= 40:
		return "at_risk"
	default:
		return "churning"
	}
}

// HealthScoreResponse é a resposta pública do endpoint GET /health-scores/:tenant_id.
type HealthScoreResponse struct {
	*TenantHealthScore
	Status string `json:"status"`
}
