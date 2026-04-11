package handler

// BKL-1055: Testes de isolamento multi-tenant para o Pulse.
// Valida que projeto A não vê dados de projeto B — flags, avaliação, variantes.
// Gate de CI: falha se qualquer vazamento de dados entre tenants for detectado.
//
// Estes testes usam mocks do serviço para ser independentes de DB/Redis.
// Para testes de integração com DB real, ver /tests/integration/ (futuro BKL-1105).

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// mockIsolationFlagService simula um serviço de flags com dados de dois projetos distintos.
type mockIsolationFlagService struct {
	projectAFlags domain.FlagsMapResponse
	projectBFlags domain.FlagsMapResponse
}

// GetFlagsMap retorna flags do projeto correto — nunca mistura entre projetos.
func (m *mockIsolationFlagService) GetFlagsMap(_ context.Context, projectID string) (domain.FlagsMapResponse, error) {
	switch projectID {
	case "project-a":
		return m.projectAFlags, nil
	case "project-b":
		return m.projectBFlags, nil
	default:
		return domain.FlagsMapResponse{Flags: map[string]bool{}}, nil
	}
}

func (m *mockIsolationFlagService) EvaluateFlags(_ context.Context, _, _, projectID string) (domain.FlagsMapResponse, error) {
	return m.GetFlagsMap(context.Background(), projectID)
}

func (m *mockIsolationFlagService) GetVariants(_ context.Context, _, _, projectID string) (domain.VariantResponse, error) {
	if projectID == "project-a" {
		return domain.VariantResponse{Variants: map[string]string{"exp-flag": "variant-a"}}, nil
	}
	return domain.VariantResponse{Variants: map[string]string{"exp-flag": "variant-b"}}, nil
}

func (m *mockIsolationFlagService) List(_ context.Context) ([]domain.FeatureFlag, error) {
	return nil, nil
}

func (m *mockIsolationFlagService) ListByProject(_ context.Context, projectID string) ([]domain.FeatureFlag, error) {
	switch projectID {
	case "project-a":
		return []domain.FeatureFlag{{Key: "flag-alpha", Enabled: true, ProjectID: "project-a"}}, nil
	case "project-b":
		return []domain.FeatureFlag{{Key: "flag-beta", Enabled: true, ProjectID: "project-b"}}, nil
	}
	return nil, nil
}

func (m *mockIsolationFlagService) GetByKey(_ context.Context, _ string) (*domain.FeatureFlag, error) {
	return nil, nil
}

func (m *mockIsolationFlagService) Create(_ context.Context, _ domain.CreateFlagRequest) (*domain.FeatureFlag, error) {
	return nil, nil
}

func (m *mockIsolationFlagService) Update(_ context.Context, _ string, _ domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	return nil, nil
}

func (m *mockIsolationFlagService) Delete(_ context.Context, _ string) error {
	return nil
}

// newIsolationTestApp monta o app de testes com o mock de isolamento.
func newIsolationTestApp(svc *mockIsolationFlagService) *fiber.App {
	h := NewFlagHandler(svc, "test-token")
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// TestIsolation_GetFlagsMap_ProjetoA_NaoVeProjetoB garante que GET /flags?project_id=project-a
// não retorna flags do projeto-b.
func TestIsolation_GetFlagsMap_ProjetoA_NaoVeProjetoB(t *testing.T) {
	svc := &mockIsolationFlagService{
		projectAFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"flag-alpha": true,
			"flag-common": true,
		}},
		projectBFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"flag-beta":   true,
			"flag-common": false, // mesma chave, valor diferente
		}},
	}
	app := newIsolationTestApp(svc)

	// Projeto A
	reqA := httptest.NewRequest("GET", "/flags?project_id=project-a", nil)
	respA, err := app.Test(reqA)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if respA.StatusCode != 200 {
		t.Fatalf("esperado 200 para projeto-a, recebido %d", respA.StatusCode)
	}

	var bodyA domain.FlagsMapResponse
	if err := json.NewDecoder(respA.Body).Decode(&bodyA); err != nil {
		t.Fatalf("decode erro: %v", err)
	}

	// flag-beta pertence ao projeto B — não deve aparecer para projeto A
	if _, found := bodyA.Flags["flag-beta"]; found {
		t.Error("ISOLAMENTO VIOLADO: projeto-a recebeu flag-beta que pertence ao projeto-b")
	}

	// flag-alpha deve estar presente e ativa
	if !bodyA.Flags["flag-alpha"] {
		t.Error("flag-alpha deveria estar ativa para projeto-a")
	}

	// flag-common deve ter o valor do projeto A (true), não do projeto B (false)
	if !bodyA.Flags["flag-common"] {
		t.Error("ISOLAMENTO VIOLADO: flag-common do projeto-a deveria ser true, recebeu false (valor do projeto-b)")
	}
}

// TestIsolation_GetFlagsMap_ProjetoB_NaoVeProjetoA garante a direção inversa.
func TestIsolation_GetFlagsMap_ProjetoB_NaoVeProjetoA(t *testing.T) {
	svc := &mockIsolationFlagService{
		projectAFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"flag-alpha": true,
		}},
		projectBFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"flag-beta": true,
		}},
	}
	app := newIsolationTestApp(svc)

	reqB := httptest.NewRequest("GET", "/flags?project_id=project-b", nil)
	respB, err := app.Test(reqB)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}

	var bodyB domain.FlagsMapResponse
	if err := json.NewDecoder(respB.Body).Decode(&bodyB); err != nil {
		t.Fatalf("decode erro: %v", err)
	}

	if _, found := bodyB.Flags["flag-alpha"]; found {
		t.Error("ISOLAMENTO VIOLADO: projeto-b recebeu flag-alpha que pertence ao projeto-a")
	}

	if !bodyB.Flags["flag-beta"] {
		t.Error("flag-beta deveria estar ativa para projeto-b")
	}
}

// TestIsolation_EvaluateFlags_ContextoUsuario_IsoladoPorProjeto valida que EvaluateFlags
// com user_id também respeita o project_id.
func TestIsolation_EvaluateFlags_ContextoUsuario_IsoladoPorProjeto(t *testing.T) {
	svc := &mockIsolationFlagService{
		projectAFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"premium-feature": true,
		}},
		projectBFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"premium-feature": false,
		}},
	}
	app := newIsolationTestApp(svc)

	// Usuário do projeto A avalia "premium-feature" → deve receber true
	reqA := httptest.NewRequest("GET", "/flags?user_id=user-123&project_id=project-a", nil)
	respA, _ := app.Test(reqA)

	var bodyA domain.FlagsMapResponse
	json.NewDecoder(respA.Body).Decode(&bodyA) //nolint:errcheck

	// Mesmo user com projeto B → deve receber false (isolamento por projeto, não por usuário)
	reqB := httptest.NewRequest("GET", "/flags?user_id=user-123&project_id=project-b", nil)
	respB, _ := app.Test(reqB)

	var bodyB domain.FlagsMapResponse
	json.NewDecoder(respB.Body).Decode(&bodyB) //nolint:errcheck

	if bodyA.Flags["premium-feature"] == bodyB.Flags["premium-feature"] {
		t.Logf("projectA.premium-feature=%v, projectB.premium-feature=%v", bodyA.Flags["premium-feature"], bodyB.Flags["premium-feature"])
		t.Error("ISOLAMENTO VIOLADO: mesmo usuário em projetos diferentes deveria ver flags diferentes")
	}
}

// TestIsolation_EvaluateBatch_NaoVazaDadosEntreProjects valida o endpoint batch.
func TestIsolation_EvaluateBatch_NaoVazaDadosEntreProjects(t *testing.T) {
	svc := &mockIsolationFlagService{
		projectAFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"feat-x": true,
			"feat-y": false,
		}},
		projectBFlags: domain.FlagsMapResponse{Flags: map[string]bool{
			"feat-x": false, // mesmo nome, valor oposto
			"feat-z": true,
		}},
	}
	app := newIsolationTestApp(svc)

	// Batch para projeto A
	payloadA := map[string]any{
		"keys":       []string{"feat-x", "feat-y"},
		"user_id":    "user-abc",
		"project_id": "project-a",
	}
	bodyA, _ := json.Marshal(payloadA)
	reqA := httptest.NewRequest("POST", "/flags/evaluate-batch", bytes.NewReader(bodyA))
	reqA.Header.Set("Content-Type", "application/json")
	respA, _ := app.Test(reqA)

	var resultA domain.FlagsMapResponse
	json.NewDecoder(respA.Body).Decode(&resultA) //nolint:errcheck

	// Batch para projeto B (mesmo user, mesma flag)
	payloadB := map[string]any{
		"keys":       []string{"feat-x"},
		"user_id":    "user-abc",
		"project_id": "project-b",
	}
	bodyB, _ := json.Marshal(payloadB)
	reqB := httptest.NewRequest("POST", "/flags/evaluate-batch", bytes.NewReader(bodyB))
	reqB.Header.Set("Content-Type", "application/json")
	respB, _ := app.Test(reqB)

	var resultB domain.FlagsMapResponse
	json.NewDecoder(respB.Body).Decode(&resultB) //nolint:errcheck

	// feat-x deve ser true para projeto-a e false para projeto-b
	if resultA.Flags["feat-x"] == resultB.Flags["feat-x"] {
		t.Errorf("ISOLAMENTO VIOLADO: feat-x = %v para projeto-a e %v para projeto-b (deveriam ser diferentes)",
			resultA.Flags["feat-x"], resultB.Flags["feat-x"])
	}

	// feat-z pertence ao projeto B — não deve aparecer na resposta de projeto A
	if _, found := resultA.Flags["feat-z"]; found {
		t.Error("ISOLAMENTO VIOLADO: feat-z do projeto-b apareceu na resposta de projeto-a")
	}
}

// TestIsolation_GetVariants_IsoladoPorProjeto valida isolamento de variantes A/B.
func TestIsolation_GetVariants_IsoladoPorProjeto(t *testing.T) {
	svc := &mockIsolationFlagService{
		projectAFlags: domain.FlagsMapResponse{Flags: map[string]bool{}},
		projectBFlags: domain.FlagsMapResponse{Flags: map[string]bool{}},
	}
	app := newIsolationTestApp(svc)

	reqA := httptest.NewRequest("GET", "/flags/variants?user_id=user-1&project_id=project-a", nil)
	respA, _ := app.Test(reqA)

	var varA domain.VariantResponse
	json.NewDecoder(respA.Body).Decode(&varA) //nolint:errcheck

	reqB := httptest.NewRequest("GET", "/flags/variants?user_id=user-1&project_id=project-b", nil)
	respB, _ := app.Test(reqB)

	var varB domain.VariantResponse
	json.NewDecoder(respB.Body).Decode(&varB) //nolint:errcheck

	// O mesmo usuário em projetos diferentes deve receber variantes diferentes
	if varA.Variants["exp-flag"] == varB.Variants["exp-flag"] {
		t.Errorf("ISOLAMENTO VIOLADO: variante exp-flag = '%s' para ambos os projetos (esperado variant-a vs variant-b)",
			varA.Variants["exp-flag"])
	}
}
