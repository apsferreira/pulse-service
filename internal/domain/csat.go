package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CSATSurvey representa uma pesquisa de satisfacao (NPS ou CSAT).
type CSATSurvey struct {
	ID        uuid.UUID `json:"id"`
	Product   string    `json:"product"`
	Question  string    `json:"question"`
	Scale     int       `json:"scale"`
	Type      string    `json:"type"` // "nps" ou "csat"
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateSurveyRequest e o payload para criacao de uma nova survey.
type CreateSurveyRequest struct {
	Product  string `json:"product"  validate:"required,max=100"`
	Question string `json:"question" validate:"required"`
	Scale    int    `json:"scale"`
	Type     string `json:"type"` // default "nps"
}

// CSATResponse representa uma resposta enviada por um respondente.
type CSATResponse struct {
	ID         uuid.UUID         `json:"id"`
	SurveyID   uuid.UUID         `json:"survey_id"`
	Score      int               `json:"score"`
	Comment    string            `json:"comment,omitempty"`
	Respondent string            `json:"respondent,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// SubmitResponseRequest e o payload enviado pelo widget embed publico.
type SubmitResponseRequest struct {
	Score      int               `json:"score"      validate:"required"`
	Comment    string            `json:"comment"`
	Respondent string            `json:"respondent"`
	Metadata   map[string]string `json:"metadata"`
}

// NPSPeriod representa o intervalo de tempo de um calculo NPS.
type NPSPeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// NPSResult agrega metricas NPS para um survey em um periodo.
type NPSResult struct {
	SurveyID   uuid.UUID `json:"survey_id"`
	Product    string    `json:"product"`
	Total      int       `json:"total"`
	Promoters  int       `json:"promoters"`   // score 9-10
	Neutrals   int       `json:"neutrals"`    // score 7-8 (alias: passives)
	Detractors int       `json:"detractors"`  // score 0-6
	NPS        float64   `json:"nps_score"`   // (promoters - detractors) / total * 100
	AvgScore   float64   `json:"average_score"`
	Period     NPSPeriod `json:"period"`
}

// CSATRepository define as operacoes de persistencia do modulo CSAT/NPS.
type CSATRepository interface {
	CreateSurvey(ctx context.Context, s *CSATSurvey) error
	GetSurvey(ctx context.Context, id uuid.UUID) (*CSATSurvey, error)
	ListSurveys(ctx context.Context, product string) ([]CSATSurvey, error)
	UpdateSurvey(ctx context.Context, s *CSATSurvey) error
	SubmitResponse(ctx context.Context, r *CSATResponse) error
	GetNPS(ctx context.Context, surveyID uuid.UUID, from, to time.Time) (*NPSResult, error)
	ListResponses(ctx context.Context, surveyID uuid.UUID, from, to time.Time, limit int) ([]CSATResponse, error)
	ListResponsesByRespondent(ctx context.Context, respondent string, limit int) ([]CSATResponse, error)
}
