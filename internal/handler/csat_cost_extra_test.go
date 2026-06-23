package handler

// Testes adicionais para cobrir caminhos de erro do CSATHandler e CostHandler.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// === CSATHandler caminhos de erro adicionais ===

func TestCSATHandler_CreateSurvey_QuestionVazia_Retorna400(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	// question está vazia
	payload := domain.CreateSurveyRequest{Product: "nitro", Question: ""}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/csat/surveys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("question vazia deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_ListSurveys_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCSATRepo{surveysErr: errors.New("db error")}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_GetNPS_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCSATRepo{npsErr: errors.New("db error")}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys/"+uuid.New().String()+"/nps", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_ListResponses_IDInvalido_Retorna400(t *testing.T) {
	repo := &mockCSATRepo{}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys/nao-uuid/responses", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("ID inválido deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_ListResponses_Sucesso(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		responses: []domain.CSATResponse{
			{ID: uuid.New(), Score: 8, SurveyID: surveyID},
		},
	}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/csat/surveys/%s/responses?days=30", surveyID), nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if count, ok := result["count"].(float64); !ok || int(count) != 1 {
		t.Errorf("esperado count=1, got %v", result["count"])
	}
}

func TestCSATHandler_UpdateSurvey_SurveyNaoEncontrada_Retorna404(t *testing.T) {
	repo := &mockCSATRepo{surveyErr: errors.New("not found")}
	app := newTestCSATApp(repo, "tok")

	payload := domain.CreateSurveyRequest{Question: "Q"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/api/v1/csat/surveys/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("survey não encontrada deveria retornar 404, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_SubmitResponse_ErroRepo_Retorna500(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{
			ID:     surveyID,
			Scale:  10,
			Active: true,
		},
		submitErr: errors.New("db error"),
	}
	app := newTestCSATApp(repo, "tok")

	payload := domain.SubmitResponseRequest{Score: 5}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", fmt.Sprintf("/public/csat/%s/respond", surveyID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_ListResponsesByRespondent_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCSATRepo{responsesErr: errors.New("db error")}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/responses?respondent=user-123", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// === CostHandler caminhos de erro adicionais ===

func TestCostHandler_RecordCost_BodyInvalido_Retorna400(t *testing.T) {
	repo := &mockCostRepo{}
	app := newTestCostApp(repo, "tok")

	req := httptest.NewRequest("POST", "/api/v1/costs", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("body inválido deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestCostHandler_GetOverview_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCostRepo{overviewErr: errors.New("db error")}
	app := newTestCostApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/overview", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCostHandler_ListRecords_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCostRepo{recordsErr: errors.New("db error")}
	app := newTestCostApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/records", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCostHandler_GetTotal_ErroRepo_Retorna500(t *testing.T) {
	repo := &mockCostRepo{totalErr: errors.New("db error")}
	app := newTestCostApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/total", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de repo deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCostHandler_GetTotal_TotalBRLCalculado(t *testing.T) {
	repo := &mockCostRepo{totalUSD: 10.0}
	app := newTestCostApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/total", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	totalBRL, ok := body["total_brl"].(float64)
	if !ok {
		t.Fatal("total_brl deve ser float64")
	}
	// total_brl = total_usd * USDToBRL (5.0)
	expectedBRL := 10.0 * domain.USDToBRL
	if totalBRL != expectedBRL {
		t.Errorf("total_brl esperado %.2f, recebido %.2f", expectedBRL, totalBRL)
	}
}
