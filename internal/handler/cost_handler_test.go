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

// --- mock de domain.CostRepository ---

type mockCostRepo struct {
	recordErr   error
	overview    *domain.CostOverview
	overviewErr error
	records     []domain.APICost
	recordsErr  error
	totalUSD    float64
	totalErr    error
}

func (m *mockCostRepo) Record(ctx context.Context, cost *domain.APICost) error {
	if m.recordErr != nil {
		return m.recordErr
	}
	cost.ID = uuid.New()
	cost.CreatedAt = time.Now()
	cost.CostBRL = cost.CostUSD * domain.USDToBRL
	return nil
}

func (m *mockCostRepo) GetOverview(ctx context.Context, product string, from, to time.Time) (*domain.CostOverview, error) {
	return m.overview, m.overviewErr
}

func (m *mockCostRepo) ListByProduct(ctx context.Context, product string, from, to time.Time, limit int) ([]domain.APICost, error) {
	return m.records, m.recordsErr
}

func (m *mockCostRepo) TotalByPeriod(ctx context.Context, from, to time.Time) (float64, error) {
	return m.totalUSD, m.totalErr
}

// newTestCostApp monta um Fiber app com as rotas do CostHandler usando mock.
func newTestCostApp(repo domain.CostRepository, token string) *fiber.App {
	h := NewCostHandler(repo, token)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// --- POST /api/v1/costs ---

func TestCostHandler_RecordCost_Sucesso_Retorna201(t *testing.T) {
	repo := &mockCostRepo{}
	app := newTestCostApp(repo, "svc-tok")

	payload := domain.RecordCostRequest{
		Product:   "socialmake",
		Operation: "generate_caption",
		Provider:  "google",
		Model:     "gemini-2.0-flash",
		TokensIn:  500,
		TokensOut: 200,
		CostUSD:   0.0015,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/costs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("esperado 201, recebido %d", resp.StatusCode)
	}

	var cost domain.APICost
	if err := json.NewDecoder(resp.Body).Decode(&cost); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}
	if cost.CostBRL == 0 {
		t.Error("cost_brl não deve ser zero — conversão USD→BRL deve ser aplicada")
	}
}

func TestCostHandler_RecordCost_SemToken_Retorna401(t *testing.T) {
	repo := &mockCostRepo{}
	app := newTestCostApp(repo, "svc-tok")

	payload := domain.RecordCostRequest{
		Product:   "socialmake",
		Operation: "generate_caption",
		Provider:  "google",
		CostUSD:   0.001,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/costs", bytes.NewReader(body))
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

func TestCostHandler_RecordCost_CamposObrigatoriosFaltando_Retorna400(t *testing.T) {
	repo := &mockCostRepo{}
	app := newTestCostApp(repo, "svc-tok")

	// product, operation e provider ausentes
	payload := domain.RecordCostRequest{CostUSD: 0.001}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/costs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("campos obrigatórios faltando deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- GET /api/v1/costs/overview ---

func TestCostHandler_GetOverview_Sucesso_Retorna200(t *testing.T) {
	now := time.Now()
	repo := &mockCostRepo{
		overview: &domain.CostOverview{
			From:     now.AddDate(0, 0, -30),
			To:       now,
			TotalUSD: 1.50,
			TotalBRL: 7.50,
			Breakdown: []domain.CostSummary{
				{Product: "socialmake", Operation: "generate_caption", Count: 100, TotalUSD: 1.50},
			},
		},
	}
	app := newTestCostApp(repo, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/overview?product=socialmake", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestCostHandler_GetOverview_SemToken_Retorna401(t *testing.T) {
	repo := &mockCostRepo{}
	app := newTestCostApp(repo, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/overview", nil)
	// Sem token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("sem token deveria retornar 401, recebeu %d", resp.StatusCode)
	}
}

// --- GET /api/v1/costs/records ---

func TestCostHandler_ListRecords_Sucesso(t *testing.T) {
	repo := &mockCostRepo{
		records: []domain.APICost{
			{Product: "socialmake", Operation: "generate_caption", CostUSD: 0.001},
		},
	}
	app := newTestCostApp(repo, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/records?product=socialmake&days=7", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

// --- GET /api/v1/costs/total ---

func TestCostHandler_GetTotal_Retorna200(t *testing.T) {
	repo := &mockCostRepo{totalUSD: 42.50}
	app := newTestCostApp(repo, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/costs/total?days=30", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	totalUSD, ok := body["total_usd"].(float64)
	if !ok {
		t.Fatal("total_usd deve ser float64")
	}
	if totalUSD != 42.50 {
		t.Errorf("total_usd esperado 42.50, recebido %v", totalUSD)
	}
}

// --- parsePeriod helper (unit test direto) ---

func TestParsePeriod_DefaultDays(t *testing.T) {
	app := fiber.New()
	var from, to time.Time

	app.Get("/test", func(c *fiber.Ctx) error {
		from, to = parsePeriod(c)
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("erro: %v", err)
	}

	// Sem query param "days" → default 30 dias
	diff := to.Sub(from)
	if diff < 29*24*time.Hour || diff > 31*24*time.Hour {
		t.Errorf("período padrão deveria ser ~30 dias, obtido %v", diff)
	}
}

func TestParsePeriod_DaysCustomizado(t *testing.T) {
	app := fiber.New()
	var from, to time.Time

	app.Get("/test", func(c *fiber.Ctx) error {
		from, to = parsePeriod(c)
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/test?days=7", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("erro: %v", err)
	}

	diff := to.Sub(from)
	if diff < 6*24*time.Hour || diff > 8*24*time.Hour {
		t.Errorf("período customizado deveria ser ~7 dias, obtido %v", diff)
	}
}
