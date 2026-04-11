package domain

import "time"

// FlagEvaluationReason descreve por que a flag resultou num valor específico.
type FlagEvaluationReason string

const (
	ReasonKillSwitch      FlagEvaluationReason = "kill_switch"
	ReasonTargetingUser   FlagEvaluationReason = "targeting_user"
	ReasonTargetingTenant FlagEvaluationReason = "targeting_tenant"
	ReasonRollout         FlagEvaluationReason = "rollout"
	ReasonDefault         FlagEvaluationReason = "default"
	ReasonDisabled        FlagEvaluationReason = "disabled"
)

// FlagEvaluation é o registro de uma avaliação de feature flag. BKL-1074.
type FlagEvaluation struct {
	ID          string               `json:"id"`
	FlagKey     string               `json:"flag_key"`
	ProjectID   *string              `json:"project_id,omitempty"`
	UserID      *string              `json:"user_id,omitempty"`
	TenantID    *string              `json:"tenant_id,omitempty"`
	Result      bool                 `json:"result"`
	Variant     *string              `json:"variant,omitempty"`
	Reason      FlagEvaluationReason `json:"reason"`
	EvaluatedAt time.Time            `json:"evaluated_at"`
}

// FlagEvaluationListResponse é a resposta paginada do audit log.
type FlagEvaluationListResponse struct {
	Items  []FlagEvaluation `json:"items"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}
