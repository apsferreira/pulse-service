package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- mock de FlagServicer ---

type mockFlagService struct {
	flagsMap    domain.FlagsMapResponse
	flagsMapErr error
	evalResp    domain.FlagsMapResponse
	evalErr     error
	listFlags   []domain.FeatureFlag
	listErr     error
	getFlag     *domain.FeatureFlag
	getErr      error
	createFlag  *domain.FeatureFlag
	createErr   error
	updateFlag  *domain.FeatureFlag
	updateErr   error
	deleteErr   error
	variants    domain.VariantResponse
	variantsErr error
}

func (m *mockFlagService) GetFlagsMap(ctx context.Context) (domain.FlagsMapResponse, error) {
	return m.flagsMap, m.flagsMapErr
}

func (m *mockFlagService) EvaluateFlags(ctx context.Context, userID, tenantID string) (domain.FlagsMapResponse, error) {
	return m.evalResp, m.evalErr
}

func (m *mockFlagService) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	return m.listFlags, m.listErr
}

func (m *mockFlagService) ListByProject(ctx context.Context, projectID string) ([]domain.FeatureFlag, error) {
	return m.listFlags, m.listErr
}

func (m *mockFlagService) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	return m.getFlag, m.getErr
}

func (m *mockFlagService) Create(ctx context.Context, req domain.CreateFlagRequest) (*domain.FeatureFlag, error) {
	return m.createFlag, m.createErr
}

func (m *mockFlagService) Update(ctx context.Context, key string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	return m.updateFlag, m.updateErr
}

func (m *mockFlagService) Delete(ctx context.Context, key string) error {
	return m.deleteErr
}

func (m *mockFlagService) GetVariants(ctx context.Context, userID, tenantID string) (domain.VariantResponse, error) {
	return m.variants, m.variantsErr
}

// newTestFlagApp monta um Fiber app com as rotas do FlagHandler usando mock de serviço.
func newTestFlagApp(svc *mockFlagService, token string) *fiber.App {
	h := NewFlagHandler(svc, token)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// --- GetFlagsMap (público, sem auth) ---

func TestFlagHandler_GetFlagsMap_Publico(t *testing.T) {
	svc := &mockFlagService{
		flagsMap: domain.FlagsMapResponse{
			Flags: map[string]bool{"feature-a": true, "feature-b": false},
		},
	}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}

	var body domain.FlagsMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if body.Flags["feature-a"] != true {
		t.Error("feature-a deveria ser true")
	}
	if body.Flags["feature-b"] != false {
		t.Error("feature-b deveria ser false")
	}
}

func TestFlagHandler_GetFlagsMap_SemToken_Retorna200(t *testing.T) {
	// Endpoint público — sem token deve retornar 200
	svc := &mockFlagService{flagsMap: domain.FlagsMapResponse{Flags: map[string]bool{}}}
	app := newTestFlagApp(svc, "token-secreto")

	req := httptest.NewRequest("GET", "/flags", nil)
	// Intencionalmente sem X-Service-Token
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("endpoint público não deve exigir token, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_GetFlagsMap_ComContexto_UsaEvaluate(t *testing.T) {
	// Quando user_id é fornecido, deve usar EvaluateFlags
	svc := &mockFlagService{
		evalResp: domain.FlagsMapResponse{Flags: map[string]bool{"feat": true}},
	}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/flags?user_id=user-123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

// --- CreateFlag (requer X-Service-Token) ---

func TestFlagHandler_CreateFlag_Sucesso(t *testing.T) {
	created := &domain.FeatureFlag{Key: "nova-flag", Enabled: true}
	svc := &mockFlagService{createFlag: created}
	app := newTestFlagApp(svc, "svc-tok")

	payload := domain.CreateFlagRequest{Key: "nova-flag", Enabled: true}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("esperado 201, recebido %d", resp.StatusCode)
	}
}

func TestFlagHandler_CreateFlag_SemToken_Retorna401(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "svc-tok")

	payload := domain.CreateFlagRequest{Key: "nova-flag", Enabled: true}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Sem X-Service-Token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("sem token deveria retornar 401, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_CreateFlag_KeyInvalida_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "svc-tok")

	// Key com espaço — padrão inválido
	payload := domain.CreateFlagRequest{Key: "flag invalida!", Enabled: true}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("key inválida deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_CreateFlag_KeyMuitoCurta_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "svc-tok")

	payload := domain.CreateFlagRequest{Key: "a", Enabled: true} // < 2 chars
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/flags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("key de 1 char deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- GetFlagByKey ---

func TestFlagHandler_GetFlagByKey_Encontrado(t *testing.T) {
	flag := &domain.FeatureFlag{Key: "minha-flag", Enabled: true}
	svc := &mockFlagService{getFlag: flag}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags/minha-flag", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestFlagHandler_GetFlagByKey_NaoEncontrado_Retorna404(t *testing.T) {
	svc := &mockFlagService{getFlag: nil} // repo retorna nil
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("GET", "/api/v1/flags/nao-existe", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("esperado 404, recebido %d", resp.StatusCode)
	}
}

// --- EvaluateBatch ---

func TestFlagHandler_EvaluateBatch_SemKeys_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	payload := map[string]any{"keys": []string{}}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/flags/evaluate-batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("keys vazio deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_EvaluateBatch_MaisDe50Keys_Retorna400(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	keys := make([]string, 51)
	for i := range keys {
		keys[i] = "flag-key"
	}
	payload := map[string]any{"keys": keys}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/flags/evaluate-batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("mais de 50 keys deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestFlagHandler_EvaluateBatch_Sucesso(t *testing.T) {
	svc := &mockFlagService{
		evalResp: domain.FlagsMapResponse{Flags: map[string]bool{"flag-a": true, "flag-b": false}},
	}
	app := newTestFlagApp(svc, "tok")

	payload := map[string]any{
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
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}

	var body2 domain.FlagsMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&body2); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if body2.Flags["flag-a"] != true {
		t.Error("flag-a deveria ser true")
	}
}

// --- ListFlags ---

func TestFlagHandler_ListFlags_Retorna200(t *testing.T) {
	svc := &mockFlagService{
		listFlags: []domain.FeatureFlag{
			{Key: "f1", Enabled: true},
			{Key: "f2", Enabled: false},
		},
	}
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
}

// --- DeleteFlag ---

func TestFlagHandler_DeleteFlag_Sucesso_Retorna204(t *testing.T) {
	svc := &mockFlagService{}
	app := newTestFlagApp(svc, "tok")

	req := httptest.NewRequest("DELETE", "/api/v1/flags/minha-flag", nil)
	req.Header.Set("X-Service-Token", "tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("esperado 204, recebido %d", resp.StatusCode)
	}
}
