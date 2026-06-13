package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- mock de domain.CSATRepository ---

type mockCSATRepo struct {
	survey        *domain.CSATSurvey
	surveyErr     error
	surveys       []domain.CSATSurvey
	surveysErr    error
	createErr     error
	updateErr     error
	submitErr     error
	npsResult     *domain.NPSResult
	npsErr        error
	responses     []domain.CSATResponse
	responsesErr  error
}

func (m *mockCSATRepo) CreateSurvey(ctx context.Context, s *domain.CSATSurvey) error {
	if m.createErr != nil {
		return m.createErr
	}
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

func (m *mockCSATRepo) GetSurvey(ctx context.Context, id uuid.UUID) (*domain.CSATSurvey, error) {
	return m.survey, m.surveyErr
}

func (m *mockCSATRepo) ListSurveys(ctx context.Context, product string) ([]domain.CSATSurvey, error) {
	return m.surveys, m.surveysErr
}

func (m *mockCSATRepo) UpdateSurvey(ctx context.Context, s *domain.CSATSurvey) error {
	return m.updateErr
}

func (m *mockCSATRepo) SubmitResponse(ctx context.Context, r *domain.CSATResponse) error {
	if m.submitErr != nil {
		return m.submitErr
	}
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	return nil
}

func (m *mockCSATRepo) GetNPS(ctx context.Context, surveyID uuid.UUID, from, to time.Time) (*domain.NPSResult, error) {
	return m.npsResult, m.npsErr
}

func (m *mockCSATRepo) ListResponses(ctx context.Context, surveyID uuid.UUID, from, to time.Time, limit int) ([]domain.CSATResponse, error) {
	return m.responses, m.responsesErr
}

func (m *mockCSATRepo) ListResponsesByRespondent(ctx context.Context, respondent string, limit int) ([]domain.CSATResponse, error) {
	return m.responses, m.responsesErr
}

// newTestCSATApp monta um Fiber app com as rotas do CSATHandler usando mock.
func newTestCSATApp(repo domain.CSATRepository, token string) *fiber.App {
	h := NewCSATHandler(repo, token)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// --- POST /public/csat/:survey_id/respond (público, sem auth) ---

func TestCSATHandler_SubmitResponse_SurveyAtiva_Retorna201(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{
			ID:      surveyID,
			Product: "nitro",
			Scale:   10,
			Active:  true,
		},
	}
	app := newTestCSATApp(repo, "tok")

	payload := domain.SubmitResponseRequest{
		Score:      9,
		Comment:    "Ótimo produto",
		Respondent: "user@example.com",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/public/csat/"+surveyID.String()+"/respond", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Sem token — endpoint público

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("survey ativa deveria retornar 201, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_SubmitResponse_SurveyInativa_Retorna422(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{
			ID:     surveyID,
			Scale:  10,
			Active: false, // inativa
		},
	}
	app := newTestCSATApp(repo, "tok")

	payload := domain.SubmitResponseRequest{Score: 8}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/public/csat/"+surveyID.String()+"/respond", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 422 {
		t.Errorf("survey inativa deveria retornar 422, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_SubmitResponse_ScoreFaltandoSurveyID_Retorna400(t *testing.T) {
	app := newTestCSATApp(&mockCSATRepo{}, "tok")

	req := httptest.NewRequest("POST", "/public/csat/id-invalido/respond", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("survey_id inválido deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_SubmitResponse_ScoreForaDoIntervalo_Retorna400(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{
			ID:     surveyID,
			Scale:  10,
			Active: true,
		},
	}
	app := newTestCSATApp(repo, "tok")

	// Score 11 > Scale 10 → inválido
	payload := domain.SubmitResponseRequest{Score: 11}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/public/csat/"+surveyID.String()+"/respond", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("score fora do intervalo deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_SubmitResponse_SemAuth_EhPublico(t *testing.T) {
	// Garante que o endpoint NÃO exige token — qualquer cliente pode responder
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{
			ID:     surveyID,
			Scale:  10,
			Active: true,
		},
	}
	app := newTestCSATApp(repo, "token-admin-qualquer")

	payload := domain.SubmitResponseRequest{Score: 7}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/public/csat/"+surveyID.String()+"/respond", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Propositalmente sem X-Service-Token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode == 401 {
		t.Error("endpoint público NÃO deve exigir autenticação — retornou 401")
	}
}

// --- POST /api/v1/csat/surveys (requer token) ---

func TestCSATHandler_CreateSurvey_Sucesso(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	payload := domain.CreateSurveyRequest{
		Product:  "nitro",
		Question: "Como você avalia nosso produto?",
		Scale:    10,
		Type:     "nps",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/csat/surveys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("esperado 201, recebido %d", resp.StatusCode)
	}
}

func TestCSATHandler_CreateSurvey_SemToken_Retorna401(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	payload := domain.CreateSurveyRequest{Product: "nitro", Question: "?"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/csat/surveys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Sem token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("sem token deveria retornar 401, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_CreateSurvey_CamposObrigatoriosFaltando_Retorna400(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	// product vazio
	payload := domain.CreateSurveyRequest{Question: "Pergunta sem produto"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/csat/surveys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("campos obrigatórios faltando deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- GET /api/v1/csat/surveys/:id/nps ---

func TestCSATHandler_GetNPS_IDInvalido_Retorna400(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys/id-invalido/nps", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("ID inválido deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_GetNPS_Sucesso(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		npsResult: &domain.NPSResult{
			SurveyID:   surveyID,
			Total:      100,
			Promoters:  70,
			Detractors: 10,
			NPS:        60.0,
		},
	}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys/"+surveyID.String()+"/nps", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}
