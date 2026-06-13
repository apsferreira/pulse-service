package handler

// BKL-1154: testes do MABHandler com mock do repositório.
// Cobre validação de input, autenticação, caminhos felizes e erros de repositório.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

const mabTestToken = "svc-mab-test-token"

// --- mock do MABStore ---

type mockMABStore struct {
	createExperimentFn func(ctx context.Context, projectID uuid.UUID, req *domain.CreateMABExperimentRequest) (*domain.MABExperiment, error)
	listExperimentsFn  func(ctx context.Context, projectID uuid.UUID) ([]domain.MABExperiment, error)
	getExperimentFn    func(ctx context.Context, projectID uuid.UUID, flagKey string) (*domain.MABExperiment, error)
	selectVariantFn    func(ctx context.Context, experimentID uuid.UUID, epsilon float64) (*domain.SelectVariantResponse, error)
	recordEventFn      func(ctx context.Context, experimentID uuid.UUID, variantKey string, isConversion bool) error
	recalculateWtsFn   func(ctx context.Context, experimentID uuid.UUID) error
	getExperimentByID  func(ctx context.Context, id uuid.UUID) (*domain.MABExperiment, error)
	setActiveFn        func(ctx context.Context, experimentID uuid.UUID, active bool) error
	countExperimentsFn func(ctx context.Context, projectID uuid.UUID) (int, error)
}

func (m *mockMABStore) CreateExperiment(ctx context.Context, projectID uuid.UUID, req *domain.CreateMABExperimentRequest) (*domain.MABExperiment, error) {
	if m.createExperimentFn != nil {
		return m.createExperimentFn(ctx, projectID, req)
	}
	id := uuid.New()
	now := time.Now()
	return &domain.MABExperiment{
		ID:        id,
		ProjectID: projectID,
		FlagKey:   req.FlagKey,
		Epsilon:   req.Epsilon,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Variants: []domain.MABVariantStats{
			{ID: uuid.New(), ExperimentID: id, VariantKey: req.Variants[0].VariantKey, Weight: 0.5},
			{ID: uuid.New(), ExperimentID: id, VariantKey: req.Variants[1].VariantKey, Weight: 0.5},
		},
	}, nil
}

func (m *mockMABStore) ListExperiments(ctx context.Context, projectID uuid.UUID) ([]domain.MABExperiment, error) {
	if m.listExperimentsFn != nil {
		return m.listExperimentsFn(ctx, projectID)
	}
	return []domain.MABExperiment{}, nil
}

func (m *mockMABStore) GetExperiment(ctx context.Context, projectID uuid.UUID, flagKey string) (*domain.MABExperiment, error) {
	if m.getExperimentFn != nil {
		return m.getExperimentFn(ctx, projectID, flagKey)
	}
	return nil, nil
}

func (m *mockMABStore) SelectVariant(ctx context.Context, experimentID uuid.UUID, epsilon float64) (*domain.SelectVariantResponse, error) {
	if m.selectVariantFn != nil {
		return m.selectVariantFn(ctx, experimentID, epsilon)
	}
	return &domain.SelectVariantResponse{VariantKey: "control", IsExplore: false, Weight: 0.5}, nil
}

func (m *mockMABStore) RecordEvent(ctx context.Context, experimentID uuid.UUID, variantKey string, isConversion bool) error {
	if m.recordEventFn != nil {
		return m.recordEventFn(ctx, experimentID, variantKey, isConversion)
	}
	return nil
}

func (m *mockMABStore) RecalculateWeights(ctx context.Context, experimentID uuid.UUID) error {
	if m.recalculateWtsFn != nil {
		return m.recalculateWtsFn(ctx, experimentID)
	}
	return nil
}

func (m *mockMABStore) GetExperimentByID(ctx context.Context, id uuid.UUID) (*domain.MABExperiment, error) {
	if m.getExperimentByID != nil {
		return m.getExperimentByID(ctx, id)
	}
	return &domain.MABExperiment{ID: id, FlagKey: "my-flag"}, nil
}

func (m *mockMABStore) SetActive(ctx context.Context, experimentID uuid.UUID, active bool) error {
	if m.setActiveFn != nil {
		return m.setActiveFn(ctx, experimentID, active)
	}
	return nil
}

func (m *mockMABStore) CountExperiments(ctx context.Context, projectID uuid.UUID) (int, error) {
	if m.countExperimentsFn != nil {
		return m.countExperimentsFn(ctx, projectID)
	}
	return 0, nil
}

// --- helpers de teste ---

func newMABApp(repo MABStore) *fiber.App {
	h := NewMABHandler(repo, mabTestToken)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

var validUUID = "00000000-0000-0000-0000-000000000000"

func mabRequest(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", mabTestToken)
	return req
}

// --- Auth ---

func TestMABHandler_Auth_SemServiceTokenRetorna401(t *testing.T) {
	app := newMABApp(&mockMABStore{})
	req := httptest.NewRequest("GET", "/api/v1/projects/"+validUUID+"/mab/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

// --- Validação de input ---

func TestMABHandler_CreateExperiment_ValidacaoDeInput(t *testing.T) {
	cases := []struct {
		name       string
		projectID  string
		body       any
		bodyRaw    []byte
		wantStatus int
	}{
		{
			name:       "projectId não-UUID",
			projectID:  "not-a-uuid",
			body:       domain.CreateMABExperimentRequest{FlagKey: "k", Variants: []domain.MABVariantRequest{{VariantKey: "a"}, {VariantKey: "b"}}, Epsilon: 0.1},
			wantStatus: 400,
		},
		{
			name:       "body inválido (json corrompido)",
			projectID:  validUUID,
			bodyRaw:    []byte("{not-json"),
			wantStatus: 400,
		},
		{
			name:       "flag_key ausente",
			projectID:  validUUID,
			body:       domain.CreateMABExperimentRequest{Variants: []domain.MABVariantRequest{{VariantKey: "a"}, {VariantKey: "b"}}, Epsilon: 0.1},
			wantStatus: 400,
		},
		{
			name:       "menos de 2 variantes",
			projectID:  validUUID,
			body:       domain.CreateMABExperimentRequest{FlagKey: "k", Variants: []domain.MABVariantRequest{{VariantKey: "a"}}, Epsilon: 0.1},
			wantStatus: 400,
		},
		{
			name:       "epsilon negativo",
			projectID:  validUUID,
			body:       domain.CreateMABExperimentRequest{FlagKey: "k", Variants: []domain.MABVariantRequest{{VariantKey: "a"}, {VariantKey: "b"}}, Epsilon: -0.1},
			wantStatus: 400,
		},
		{
			name:       "epsilon > 1",
			projectID:  validUUID,
			body:       domain.CreateMABExperimentRequest{FlagKey: "k", Variants: []domain.MABVariantRequest{{VariantKey: "a"}, {VariantKey: "b"}}, Epsilon: 1.5},
			wantStatus: 400,
		},
	}

	app := newMABApp(&mockMABStore{})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			if tc.bodyRaw != nil {
				body = tc.bodyRaw
			} else {
				body, _ = json.Marshal(tc.body)
			}
			req := httptest.NewRequest("POST", "/api/v1/projects/"+tc.projectID+"/mab/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Service-Token", mabTestToken)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

func TestMABHandler_CreateExperiment_Sucesso_Retorna201(t *testing.T) {
	mock := &mockMABStore{}
	app := newMABApp(mock)

	payload := domain.CreateMABExperimentRequest{
		FlagKey: "test-flag",
		Epsilon: 0.1,
		Variants: []domain.MABVariantRequest{
			{VariantKey: "control"},
			{VariantKey: "treatment"},
		},
	}
	body, _ := json.Marshal(payload)

	req := mabRequest("POST", "/api/v1/projects/"+validUUID+"/mab/", body)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}

	var exp domain.MABExperiment
	if err := json.NewDecoder(resp.Body).Decode(&exp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if exp.FlagKey != "test-flag" {
		t.Errorf("FlagKey = %q, want test-flag", exp.FlagKey)
	}
	if len(exp.Variants) != 2 {
		t.Errorf("variants = %d, want 2", len(exp.Variants))
	}
}

func TestMABHandler_CreateExperiment_RepoError_Retorna500(t *testing.T) {
	mock := &mockMABStore{
		createExperimentFn: func(_ context.Context, _ uuid.UUID, _ *domain.CreateMABExperimentRequest) (*domain.MABExperiment, error) {
			return nil, assertAnError{}
		},
	}
	app := newMABApp(mock)

	payload := domain.CreateMABExperimentRequest{
		FlagKey: "err-flag",
		Epsilon: 0.1,
		Variants: []domain.MABVariantRequest{
			{VariantKey: "a"}, {VariantKey: "b"},
		},
	}
	body, _ := json.Marshal(payload)
	req := mabRequest("POST", "/api/v1/projects/"+validUUID+"/mab/", body)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- List ---

func TestMABHandler_ListExperiments_Sucesso_Retorna200(t *testing.T) {
	mock := &mockMABStore{
		listExperimentsFn: func(_ context.Context, _ uuid.UUID) ([]domain.MABExperiment, error) {
			return []domain.MABExperiment{
				{ID: uuid.New(), FlagKey: "exp-a"},
				{ID: uuid.New(), FlagKey: "exp-b"},
			}, nil
		},
	}
	app := newMABApp(mock)
	req := mabRequest("GET", "/api/v1/projects/"+validUUID+"/mab/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	total, _ := body["total"].(float64)
	if int(total) != 2 {
		t.Errorf("total = %v, want 2", total)
	}
}

func TestMABHandler_ListExperiments_ProjectIdInvalido_Retorna400(t *testing.T) {
	app := newMABApp(&mockMABStore{})
	req := mabRequest("GET", "/api/v1/projects/bad/mab/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// --- Get ---

func TestMABHandler_GetExperiment_NaoEncontrado_Retorna404(t *testing.T) {
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return nil, nil
		},
	}
	app := newMABApp(mock)
	req := mabRequest("GET", "/api/v1/projects/"+validUUID+"/mab/unknown-flag", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestMABHandler_GetExperiment_RepoError_Retorna500(t *testing.T) {
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return nil, assertAnError{}
		},
	}
	app := newMABApp(mock)
	req := mabRequest("GET", "/api/v1/projects/"+validUUID+"/mab/some-flag", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- SelectVariant ---

func TestMABHandler_SelectVariant_Sucesso_Retorna200(t *testing.T) {
	expID := uuid.New()
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{ID: expID, FlagKey: "my-flag", Epsilon: 0.1, IsActive: true}, nil
		},
	}
	app := newMABApp(mock)
	req := mabRequest("GET", "/api/v1/projects/"+validUUID+"/mab/my-flag/select", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var sel domain.SelectVariantResponse
	if err := json.NewDecoder(resp.Body).Decode(&sel); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if sel.VariantKey == "" {
		t.Error("VariantKey não deve ser vazio")
	}
}

func TestMABHandler_SelectVariant_ExperimentoInativo_Retorna409(t *testing.T) {
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{ID: uuid.New(), FlagKey: "my-flag", IsActive: false}, nil
		},
	}
	app := newMABApp(mock)
	req := mabRequest("GET", "/api/v1/projects/"+validUUID+"/mab/my-flag/select", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 409 {
		t.Errorf("experimento inativo = %d, want 409", resp.StatusCode)
	}
}

// --- RecordEvent ---

func TestMABHandler_RecordEvent_Sucesso_Retorna204(t *testing.T) {
	expID := uuid.New()
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{ID: expID, FlagKey: "my-flag", IsActive: true}, nil
		},
	}
	app := newMABApp(mock)

	payload := domain.RecordMABEventRequest{VariantKey: "control", IsConversion: true}
	body, _ := json.Marshal(payload)
	req := mabRequest("POST", "/api/v1/projects/"+validUUID+"/mab/my-flag/events", body)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 204 {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestMABHandler_RecordEvent_VariantKeyVazio_Retorna400(t *testing.T) {
	app := newMABApp(&mockMABStore{})
	body, _ := json.Marshal(domain.RecordMABEventRequest{VariantKey: ""})
	req := mabRequest("POST", "/api/v1/projects/"+validUUID+"/mab/my-flag/events", body)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// --- RecalculateWeights ---

func TestMABHandler_RecalculateWeights_Sucesso_Retorna200(t *testing.T) {
	expID := uuid.New()
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{ID: expID, FlagKey: "my-flag", IsActive: true}, nil
		},
		getExperimentByID: func(_ context.Context, id uuid.UUID) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{
				ID: id, FlagKey: "my-flag",
				Variants: []domain.MABVariantStats{
					{VariantKey: "control", Weight: 0.6},
					{VariantKey: "treatment", Weight: 0.4},
				},
			}, nil
		},
	}
	app := newMABApp(mock)
	req := mabRequest("POST", "/api/v1/projects/"+validUUID+"/mab/my-flag/recalculate", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var exp domain.MABExperiment
	if err := json.NewDecoder(resp.Body).Decode(&exp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(exp.Variants) != 2 {
		t.Errorf("variants = %d, want 2", len(exp.Variants))
	}
}

// --- SetActive ---

func TestMABHandler_SetActive_Sucesso_Retorna200(t *testing.T) {
	expID := uuid.New()
	mock := &mockMABStore{
		getExperimentFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.MABExperiment, error) {
			return &domain.MABExperiment{ID: expID, FlagKey: "my-flag", IsActive: false}, nil
		},
	}
	app := newMABApp(mock)

	payload := domain.SetMABActiveRequest{IsActive: true}
	body, _ := json.Marshal(payload)
	req := mabRequest("PATCH", "/api/v1/projects/"+validUUID+"/mab/my-flag/active", body)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["is_active"] != true {
		t.Error("is_active deve ser true")
	}
}

// --- Validação de autenticação em todos os endpoints ---

type assertAnError struct{}

func (assertAnError) Error() string { return "assertAnError" }
