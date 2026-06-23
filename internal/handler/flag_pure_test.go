package handler

// Testes de funções puras do flag_handler.go
// Cobre: resolveVariant, buildPreviewURL, CreateFlag body inválido

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- buildPreviewURL ---

func TestBuildPreviewURL_Variante(t *testing.T) {
	url := buildPreviewURL("minha-flag", "challenger", true)
	expected := "?pulse_override[minha-flag]=challenger"
	if url != expected {
		t.Errorf("esperado %q, got %q", expected, url)
	}
}

func TestBuildPreviewURL_Enabled(t *testing.T) {
	url := buildPreviewURL("minha-flag", "", true)
	expected := "?pulse_override[minha-flag]=on"
	if url != expected {
		t.Errorf("esperado %q, got %q", expected, url)
	}
}

func TestBuildPreviewURL_Disabled(t *testing.T) {
	url := buildPreviewURL("minha-flag", "", false)
	expected := "?pulse_override[minha-flag]=off"
	if url != expected {
		t.Errorf("esperado %q, got %q", expected, url)
	}
}

// --- resolveVariant ---

func TestResolveVariant_SemVariantes_RetornaVazio(t *testing.T) {
	flag := &domain.FeatureFlag{Key: "flag-sem-variantes", Variants: nil}
	got := resolveVariant(flag, "user-123")
	if got != "" {
		t.Errorf("sem variantes deveria retornar vazio, got %q", got)
	}
}

func TestResolveVariant_SemUserID_RetornaVazio(t *testing.T) {
	flag := &domain.FeatureFlag{
		Key: "flag-com-variantes",
		Variants: []domain.FlagVariant{
			{Key: "control", Weight: 50},
			{Key: "challenger", Weight: 50},
		},
	}
	got := resolveVariant(flag, "") // userID vazio
	if got != "" {
		t.Errorf("sem userID deveria retornar vazio, got %q", got)
	}
}

func TestResolveVariant_ComVariantesEUserID_RetornaVariante(t *testing.T) {
	flag := &domain.FeatureFlag{
		Key: "exp-flag",
		Variants: []domain.FlagVariant{
			{Key: "control", Weight: 50},
			{Key: "challenger", Weight: 50},
		},
	}
	got := resolveVariant(flag, "user-abc")
	if got != "control" && got != "challenger" {
		t.Errorf("deveria retornar 'control' ou 'challenger', got %q", got)
	}
}

// --- CreateFlag — BodyParser falho ---

func TestFlagHandler_CreateFlag_BodyInvalido_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader([]byte(`{invalid json`)))
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

// --- ListFlags — lista vazia retorna array vazio, não null ---

func TestFlagHandler_ListFlags_ListaVazia_RetornaArrayVazio(t *testing.T) {
	svc := &mockFlagService{listFlags: nil}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags", nil)
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
	// Lista nil deve ser normalizada para array vazio (não null no JSON)
	if body == nil {
		t.Error("lista nil deveria ser normalizada para array vazio")
	}
	if len(body) != 0 {
		t.Errorf("esperado 0 flags, recebido %d", len(body))
	}
}
