package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CSATRepository implementa domain.CSATRepository com pgxpool.
type CSATRepository struct {
	db *pgxpool.Pool
}

// NewCSATRepository cria um novo CSATRepository.
func NewCSATRepository(db *pgxpool.Pool) *CSATRepository {
	return &CSATRepository{db: db}
}

// CreateSurvey insere uma nova survey no banco.
func (r *CSATRepository) CreateSurvey(ctx context.Context, s *domain.CSATSurvey) error {
	s.ID = uuid.New()
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO csat_surveys (id, product, question, scale, type, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.Product, s.Question, s.Scale, s.Type, s.Active, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("csat_repository: CreateSurvey: %w", err)
	}
	return nil
}

// GetSurvey retorna uma survey pelo ID.
func (r *CSATRepository) GetSurvey(ctx context.Context, id uuid.UUID) (*domain.CSATSurvey, error) {
	var s domain.CSATSurvey
	err := r.db.QueryRow(ctx,
		`SELECT id, product, question, scale, type, active, created_at, updated_at
		 FROM csat_surveys
		 WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.Product, &s.Question, &s.Scale, &s.Type, &s.Active, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("csat_repository: GetSurvey: %w", err)
	}
	return &s, nil
}

// ListSurveys retorna surveys, filtrando por produto se informado.
func (r *CSATRepository) ListSurveys(ctx context.Context, product string) ([]domain.CSATSurvey, error) {
	var (
		query string
		args  []interface{}
	)

	if product != "" {
		query = `SELECT id, product, question, scale, type, active, created_at, updated_at
		          FROM csat_surveys
		          WHERE product = $1
		          ORDER BY created_at DESC`
		args = []interface{}{product}
	} else {
		query = `SELECT id, product, question, scale, type, active, created_at, updated_at
		          FROM csat_surveys
		          ORDER BY created_at DESC`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("csat_repository: ListSurveys query: %w", err)
	}
	defer rows.Close()

	var surveys []domain.CSATSurvey
	for rows.Next() {
		var s domain.CSATSurvey
		if err := rows.Scan(&s.ID, &s.Product, &s.Question, &s.Scale, &s.Type, &s.Active, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("csat_repository: ListSurveys scan: %w", err)
		}
		surveys = append(surveys, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("csat_repository: ListSurveys rows: %w", err)
	}
	if surveys == nil {
		surveys = []domain.CSATSurvey{}
	}
	return surveys, nil
}

// UpdateSurvey atualiza os campos editaveis de uma survey.
func (r *CSATRepository) UpdateSurvey(ctx context.Context, s *domain.CSATSurvey) error {
	s.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE csat_surveys
		 SET product = $1, question = $2, scale = $3, type = $4, active = $5, updated_at = $6
		 WHERE id = $7`,
		s.Product, s.Question, s.Scale, s.Type, s.Active, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("csat_repository: UpdateSurvey: %w", err)
	}
	return nil
}

// SubmitResponse persiste uma resposta de respondente.
func (r *CSATRepository) SubmitResponse(ctx context.Context, resp *domain.CSATResponse) error {
	resp.ID = uuid.New()
	resp.CreatedAt = time.Now().UTC()

	meta, err := json.Marshal(resp.Metadata)
	if err != nil {
		return fmt.Errorf("csat_repository: SubmitResponse marshal metadata: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO csat_responses (id, survey_id, score, comment, respondent, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		resp.ID, resp.SurveyID, resp.Score, resp.Comment, resp.Respondent, meta, resp.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("csat_repository: SubmitResponse: %w", err)
	}
	return nil
}

// GetNPS calcula promotores, neutros, detratores e NPS score para o periodo.
func (r *CSATRepository) GetNPS(ctx context.Context, surveyID uuid.UUID, from, to time.Time) (*domain.NPSResult, error) {
	// Busca dados da survey para popular Product no resultado
	survey, err := r.GetSurvey(ctx, surveyID)
	if err != nil {
		return nil, fmt.Errorf("csat_repository: GetNPS get survey: %w", err)
	}

	var total, promoters, passives, detractors int
	var avgScore float64

	err = r.db.QueryRow(ctx,
		`SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN score >= 9 THEN 1 ELSE 0 END), 0) AS promoters,
			COALESCE(SUM(CASE WHEN score BETWEEN 7 AND 8 THEN 1 ELSE 0 END), 0) AS passives,
			COALESCE(SUM(CASE WHEN score <= 6 THEN 1 ELSE 0 END), 0) AS detractors,
			COALESCE(AVG(score::float), 0) AS avg_score
		 FROM csat_responses
		 WHERE survey_id = $1 AND created_at BETWEEN $2 AND $3`,
		surveyID, from, to,
	).Scan(&total, &promoters, &passives, &detractors, &avgScore)
	if err != nil {
		return nil, fmt.Errorf("csat_repository: GetNPS query: %w", err)
	}

	var nps float64
	if total > 0 {
		nps = float64(promoters-detractors) / float64(total) * 100
	}

	return &domain.NPSResult{
		SurveyID:   surveyID,
		Product:    survey.Product,
		Total:      total,
		Promoters:  promoters,
		Neutrals:   passives,
		Detractors: detractors,
		NPS:        nps,
		AvgScore:   avgScore,
		Period: domain.NPSPeriod{
			From: from,
			To:   to,
		},
	}, nil
}

// ListResponses retorna respostas individuais de uma survey no periodo.
func (r *CSATRepository) ListResponses(ctx context.Context, surveyID uuid.UUID, from, to time.Time, limit int) ([]domain.CSATResponse, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, survey_id, score, COALESCE(comment, ''), COALESCE(respondent, ''), COALESCE(metadata, '{}'), created_at
		 FROM csat_responses
		 WHERE survey_id = $1 AND created_at BETWEEN $2 AND $3
		 ORDER BY created_at DESC
		 LIMIT $4`,
		surveyID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("csat_repository: ListResponses query: %w", err)
	}
	defer rows.Close()

	var responses []domain.CSATResponse
	for rows.Next() {
		var resp domain.CSATResponse
		var metaRaw []byte
		if err := rows.Scan(
			&resp.ID, &resp.SurveyID, &resp.Score, &resp.Comment, &resp.Respondent,
			&metaRaw, &resp.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("csat_repository: ListResponses scan: %w", err)
		}
		if metaRaw != nil {
			if err := json.Unmarshal(metaRaw, &resp.Metadata); err != nil {
				resp.Metadata = make(map[string]string)
			}
		}
		responses = append(responses, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("csat_repository: ListResponses rows: %w", err)
	}
	if responses == nil {
		responses = []domain.CSATResponse{}
	}
	return responses, nil
}
