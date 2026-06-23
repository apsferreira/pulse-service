package handler

// Testes adicionais do UpdateSurvey para aumentar cobertura.

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

func TestCSATHandler_UpdateSurvey_ErroRepo_Retorna500(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey:    &domain.CSATSurvey{ID: surveyID, Product: "nitro", Question: "Q", Scale: 10, Active: true},
		updateErr: errors.New("db error"),
	}
	app := newTestCSATApp(repo, "tok")

	payload := domain.CreateSurveyRequest{Question: "Nova pergunta"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/csat/surveys/%s", surveyID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de update deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_UpdateSurvey_AtualizaActive(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{ID: surveyID, Product: "nitro", Question: "Q?", Scale: 10, Active: true},
	}
	app := newTestCSATApp(repo, "tok")

	// Desativar survey: active=false
	active := false
	body, _ := json.Marshal(map[string]interface{}{"active": active})

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/csat/surveys/%s", surveyID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200 ao atualizar active, recebeu %d", resp.StatusCode)
	}
}

func TestCSATHandler_UpdateSurvey_AtualizaScale(t *testing.T) {
	surveyID := uuid.New()
	repo := &mockCSATRepo{
		survey: &domain.CSATSurvey{ID: surveyID, Product: "nitro", Question: "Q?", Scale: 5, Active: true},
	}
	app := newTestCSATApp(repo, "tok")

	payload := domain.CreateSurveyRequest{Scale: 10}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/csat/surveys/%s", surveyID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200 ao atualizar scale, recebeu %d", resp.StatusCode)
	}

	var result domain.CSATSurvey
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Scale != 10 {
		t.Errorf("scale esperado 10, got %d", result.Scale)
	}
}

func TestCSATHandler_GetNPS_NoRows_Retorna404(t *testing.T) {
	// erro contendo "no rows" deve retornar 404 em vez de 500
	repo := &mockCSATRepo{npsErr: fmt.Errorf("sql: no rows in result set")}
	app := newTestCSATApp(repo, "tok")

	req := httptest.NewRequest("GET", "/api/v1/csat/surveys/"+uuid.New().String()+"/nps", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("no rows deveria retornar 404, recebeu %d", resp.StatusCode)
	}
}
