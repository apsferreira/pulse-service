package domain

import "time"

// InsightType identifica a categoria de um insight gerado pelo agente autônomo.
type InsightType string

const (
	InsightTypeDormantFlag   InsightType = "dormant_flag"   // flag ativa sem avaliações há 30+ dias
	InsightTypeFullRollout   InsightType = "full_rollout"   // flag em 100% há 14+ dias
	InsightTypeSpike         InsightType = "spike"          // volume de eventos > 3x a média diária
	InsightTypeDrop          InsightType = "drop"           // volume de eventos < 20% da média diária
	InsightTypeNegativeRate  InsightType = "negative_rate"  // >90% das avaliações retornam false
	InsightTypeABSuggestion  InsightType = "ab_suggestion"  // rollout parcial sem experimento A/B
)

// InsightSeverity classifica a urgência do insight.
type InsightSeverity string

const (
	InsightSeverityInfo     InsightSeverity = "info"
	InsightSeverityWarning  InsightSeverity = "warning"
	InsightSeverityCritical InsightSeverity = "critical"
)

// ProjectInsight representa um insight acionável gerado pelo agente de análise.
type ProjectInsight struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	InsightType InsightType     `json:"insight_type"`
	FlagKey     *string         `json:"flag_key,omitempty"`
	Message     string          `json:"message"`
	Severity    InsightSeverity `json:"severity"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	IsDismissed bool            `json:"is_dismissed"`
	DismissedAt *time.Time      `json:"dismissed_at,omitempty"`
	DismissedBy *string         `json:"dismissed_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
}

// DismissInsightRequest é o payload para POST /api/v1/insights/:id/dismiss.
type DismissInsightRequest struct {
	DismissedBy string `json:"dismissed_by" validate:"required,max=255"`
}
