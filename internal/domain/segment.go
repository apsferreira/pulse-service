package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RuleOperator define os operadores de comparação suportados nas regras de targeting.
type RuleOperator string

const (
	OpEquals      RuleOperator = "equals"
	OpNotEquals   RuleOperator = "not_equals"
	OpContains    RuleOperator = "contains"
	OpNotContains RuleOperator = "not_contains"
	OpStartsWith  RuleOperator = "starts_with"
	OpEndsWith    RuleOperator = "ends_with"
	OpIn          RuleOperator = "in"
	OpNotIn       RuleOperator = "not_in"
	OpGt          RuleOperator = "gt"
	OpGte         RuleOperator = "gte"
	OpLt          RuleOperator = "lt"
	OpLte         RuleOperator = "lte"
)

// Rule representa uma condição atômica de targeting.
// Attribute é o caminho do atributo no contexto do usuário (ex: "user.plan", "user.country").
// Value pode ser string, float64 ou []any (para in/not_in).
type Rule struct {
	Attribute string       `json:"attribute"`
	Operator  RuleOperator `json:"operator"`
	Value     any          `json:"value"`
}

// RuleGroup é um conjunto de regras avaliadas com AND.
// O segmento avalia grupos com OR: basta um grupo satisfeito para o segmento corresponder.
type RuleGroup struct {
	Rules []Rule `json:"rules"`
}

// Segment representa um segmento de usuários com regras compostas de targeting.
type Segment struct {
	ID          uuid.UUID   `json:"id"`
	ProjectID   uuid.UUID   `json:"project_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rules       []RuleGroup `json:"rules"` // OR entre grupos
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CreateSegmentRequest é o payload para criação de um segmento.
type CreateSegmentRequest struct {
	Name        string      `json:"name"        validate:"required,max=100"`
	Description string      `json:"description" validate:"max=500"`
	Rules       []RuleGroup `json:"rules"       validate:"required,min=1"`
}

// UpdateSegmentRequest é o payload para atualização parcial de um segmento.
type UpdateSegmentRequest struct {
	Name        *string     `json:"name"        validate:"omitempty,max=100"`
	Description *string     `json:"description" validate:"omitempty,max=500"`
	Rules       []RuleGroup `json:"rules"`
}

// SegmentAssignment representa a associação entre uma flag e um segmento.
type SegmentAssignment struct {
	FlagID            uuid.UUID `json:"flag_id"`
	SegmentID         uuid.UUID `json:"segment_id"`
	RolloutPercentage int       `json:"rollout_percentage"`
}

// AssignSegmentRequest é o payload para associar um segmento a uma flag.
type AssignSegmentRequest struct {
	SegmentID         uuid.UUID `json:"segment_id"         validate:"required"`
	RolloutPercentage int       `json:"rollout_percentage" validate:"min=0,max=100"`
}

// EvaluateSegmentRequest é o payload do endpoint de avaliação pontual de segmento.
type EvaluateSegmentRequest struct {
	Context map[string]any `json:"context" validate:"required"`
}

// EvaluateSegment verifica se um contexto de atributos satisfaz as regras do segmento.
// ctx deve ter chaves no formato "user.plan", "user.country", "account.age", etc.
// Retorna true se ao menos um RuleGroup for completamente satisfeito (lógica OR/AND).
// Segmento sem grupos retorna false (fail-safe).
func EvaluateSegment(segment *Segment, ctx map[string]any) bool {
	if len(segment.Rules) == 0 {
		return false
	}
	for _, group := range segment.Rules {
		if evaluateGroup(group, ctx) {
			return true
		}
	}
	return false
}

// evaluateGroup avalia um grupo de regras com AND: todas precisam ser satisfeitas.
func evaluateGroup(group RuleGroup, ctx map[string]any) bool {
	for _, rule := range group.Rules {
		if !evaluateRule(rule, ctx) {
			return false
		}
	}
	return true
}

// evaluateRule avalia uma regra atômica contra o contexto.
// Suporta operadores de string, numéricos e de coleção (in/not_in).
func evaluateRule(rule Rule, ctx map[string]any) bool {
	val, ok := ctx[rule.Attribute]
	if !ok {
		return false
	}

	switch rule.Operator {
	case OpEquals:
		return compareEquality(val, rule.Value)

	case OpNotEquals:
		return !compareEquality(val, rule.Value)

	case OpContains:
		s, sv := toStrings(val, rule.Value)
		return strings.Contains(s, sv)

	case OpNotContains:
		s, sv := toStrings(val, rule.Value)
		return !strings.Contains(s, sv)

	case OpStartsWith:
		s, sv := toStrings(val, rule.Value)
		return strings.HasPrefix(s, sv)

	case OpEndsWith:
		s, sv := toStrings(val, rule.Value)
		return strings.HasSuffix(s, sv)

	case OpIn:
		return containsInList(val, rule.Value)

	case OpNotIn:
		return !containsInList(val, rule.Value)

	case OpGt:
		a, b, ok := toFloats(val, rule.Value)
		return ok && a > b

	case OpGte:
		a, b, ok := toFloats(val, rule.Value)
		return ok && a >= b

	case OpLt:
		a, b, ok := toFloats(val, rule.Value)
		return ok && a < b

	case OpLte:
		a, b, ok := toFloats(val, rule.Value)
		return ok && a <= b
	}

	return false
}

// compareEquality compara dois valores por igualdade, suportando string e número.
func compareEquality(a, b any) bool {
	// Comparação direta (mesmo tipo)
	if fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) {
		return true
	}
	// Comparação numérica (float64 — JSON unmarshal converte números para float64)
	af, bOk := toFloat(a)
	bf, bfOk := toFloat(b)
	if bOk && bfOk {
		return af == bf
	}
	return false
}

// toStrings converte dois valores quaisquer para string para comparações de substring.
func toStrings(a, b any) (string, string) {
	return fmt.Sprintf("%v", a), fmt.Sprintf("%v", b)
}

// toFloat tenta converter um valor para float64.
func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case int32:
		return float64(t), true
	}
	return 0, false
}

// toFloats converte dois valores para float64; retorna ok=false se qualquer um falhar.
func toFloats(a, b any) (float64, float64, bool) {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	return af, bf, aok && bok
}

// containsInList verifica se val está presente na lista rule.Value.
// rule.Value deve ser []any ou []string (formato padrão de JSON unmarshal).
func containsInList(val, list any) bool {
	valStr := fmt.Sprintf("%v", val)
	switch l := list.(type) {
	case []any:
		for _, item := range l {
			if fmt.Sprintf("%v", item) == valStr {
				return true
			}
		}
	case []string:
		for _, item := range l {
			if item == valStr {
				return true
			}
		}
	}
	return false
}
