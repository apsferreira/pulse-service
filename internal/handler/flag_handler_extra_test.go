package handler

// Testes adicionais para cobrir caminhos de erro e edge cases do FlagHandler.
// Complementa flag_handler_test.go.

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- UpdateFlag ---

func TestFlagHandler_UpdateFlag_Sucesso_Retorna200(t *testing.T) {
	updated := &domain.FeatureFlag{Key: "minha-flag", Enabled: false}
	svc := &mockFlagService{updateFlag: updated}
	app := newTestFlagApp(svc, "tok")

	enabled := false
	payload := domain.UpdateFlagRequest{Enabled: &enabled}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/api/v1/flags/minha-flag", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestFlagHandler_UpdateFlag_KeyMuitoCurta_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	payload := domain.UpdateFlagRequest{}
	body, _ := json.Marshal(payload)

	// key de 1 char é inválida (< 2 chars)
	req := httptest.NewRequest("PUT", "/api/v1/flags/a", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("key de 1 char deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_UpdateFlag_NaoEncontrada_Retorna404(t *testing.T) {
	svc := &mockFlagService{updateFlag: nil} // serviço retorna nil sem erro
	app := newTestFlagApp(svc, "tok")

	payload := domain.UpdateFlagRequest{}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/api/v1/flags/nao-existe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("flag não encontrada deveria retornar 404, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_UpdateFlag_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{updateErr: errors.New("db error")}
	app := newTestFlagApp(svc, "tok")

	payload := domain.UpdateFlagRequest{}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/api/v1/flags/minha-flag", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- DeleteFlag error paths ---

func TestFlagHandler_DeleteFlag_KeyUmChar_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	// key de 1 char é inválida
	req := httptest.NewRequest("DELETE", "/api/v1/flags/x", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("key de 1 char deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_DeleteFlag_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{deleteErr: errors.New("db error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("DELETE", "/api/v1/flags/minha-flag", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- GetFlagByKey error paths ---

func TestFlagHandler_GetFlagByKey_KeyInvalida_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags/a", nil) // 1 char — inválido (< 2)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("key de 1 char deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_GetFlagByKey_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{getErr: errors.New("db error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags/minha-flag", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- ListFlags ---

func TestFlagHandler_ListFlags_FiltroPorProjectID(t *testing.T) {
	svc := &mockFlagService{
		listFlags: []domain.FeatureFlag{
			{Key: "flag-projeto", Enabled: true},
		},
	}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags?project_id=proj-abc", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}

	var body []domain.FeatureFlag
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if len(body) != 1 {
		t.Errorf("esperado 1 flag, recebido %d", len(body))
	}
}

func TestFlagHandler_ListFlags_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{listErr: errors.New("db error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- GetFlagsMap error paths ---

func TestFlagHandler_GetFlagsMap_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{flagsMapErr: errors.New("redis error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_GetFlagsMap_EvaluateRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{evalErr: errors.New("eval error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags?user_id=user-123", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de evaluate deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- GetVariants ---

func TestFlagHandler_GetVariants_Sucesso(t *testing.T) {
	svc := &mockFlagService{
		variants: domain.VariantResponse{
			Variants: map[string]string{"exp-flag": "challenger"},
		},
	}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags/variants?user_id=user-123", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestFlagHandler_GetVariants_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockFlagService{variantsErr: errors.New("error")}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags/variants?user_id=user-123", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- EvaluateBatch — fail-open ---

func TestFlagHandler_EvaluateBatch_ServicoRetornaErro_FailOpen(t *testing.T) {
	// Fail-open: erro no serviço retorna todas as flags como false
	svc := &mockFlagService{evalErr: errors.New("redis error")}
	app := newTestFlagApp(svc, "tok")

	payload := map[string]interface{}{
		"keys":    []string{"flag-a", "flag-b"},
		"user_id": "user-123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/flags/evaluate-batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	// Fail-open: retorna 200 com todas as flags false
	if resp.StatusCode != 200 {
		t.Errorf("fail-open deveria retornar 200, recebeu %d", resp.StatusCode)
	}

	var result domain.FlagsMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if result.Flags["flag-a"] != false || result.Flags["flag-b"] != false {
		t.Error("fail-open deveria retornar todas as flags como false")
	}
}

// --- CreateFlag error paths ---

func TestFlagHandler_CreateFlag_ServicoRetornaErro_Retorna409(t *testing.T) {
	svc := &mockFlagService{createErr: errors.New("chave duplicada")}
	app := newTestFlagApp(svc, "tok")

	payload := domain.CreateFlagRequest{Key: "nova-flag", Enabled: true}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 409 {
		t.Errorf("conflito de flag deveria retornar 409, recebeu %d", resp.StatusCode)
	}
}
