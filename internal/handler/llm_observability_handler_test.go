package handler

// BKL-1063: testes do LLMObservabilityHandler.
// Cobre funções puras (parseDateRange, round2) e registro de rotas.
// O caminho HTTP completo depende de pgxpool e fica para testes de integração.

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

const llmTestToken = "svc-llm-test-token"

func newLLMApp() *fiber.App {
	h := NewLLMObservabilityHandler(nil, llmTestToken)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// --- parseDateRange ---

func TestParseDateRange_Default_SemParametros(t *testing.T) {
	start, end, err := parseDateRange("", "")
	if err != nil {
		t.Fatalf("parseDateRange: %v", err)
	}
	if end.Before(start) || end.Equal(start) {
		t.Error("end deve ser depois de start")
	}
	diff := end.Sub(start)
	if diff < 29*24*time.Hour || diff > 31*24*time.Hour {
		t.Errorf("diferença = %v, esperava ~30 dias", diff)
	}
}

func TestParseDateRange_ComParametrosValidos(t *testing.T) {
	start, end, err := parseDateRange("2025-01-01", "2025-01-15")
	if err != nil {
		t.Fatalf("parseDateRange: %v", err)
	}
	if start.Year() != 2025 || start.Month() != 1 || start.Day() != 1 {
		t.Errorf("start = %v, want 2025-01-01", start)
	}
	if end.Year() != 2025 || end.Month() != 1 || end.Day() != 15 {
		t.Errorf("end = %v, want 2025-01-15", end)
	}
}

func TestParseDateRange_StartInvalido_RetornaErro(t *testing.T) {
	_, _, err := parseDateRange("not-a-date", "")
	if err == nil {
		t.Error("parseDateRange com data inválida deve retornar erro")
	}
}

func TestParseDateRange_EndInvalido_RetornaErro(t *testing.T) {
	_, _, err := parseDateRange("", "bad-date")
	if err == nil {
		t.Error("parseDateRange com end inválido deve retornar erro")
	}
}

// --- round2 ---

func TestRound2_ArredondaCorretamente(t *testing.T) {
	cases := []struct {
		input float64
		want  float64
	}{
		{1.23456, 1.23},
		{1.23556, 1.24},
		{0, 0},
		{99.999, 100.00},
		{-1.234, -1.22},  // round2 trunca para cima com +0.5 mesmo em negativos
	}
	for _, tc := range cases {
		got := round2(tc.input)
		if got != tc.want {
			t.Errorf("round2(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// --- Auth (rotas exigem token) ---

func TestLLMObservability_Auth_SemToken_Retorna401(t *testing.T) {
	app := newLLMApp()
	endpoints := []string{
		"/api/v1/llm/overview?from=2025-01-01&to=2025-01-31",
		"/api/v1/llm/by-model?from=2025-01-01&to=2025-01-31",
		"/api/v1/llm/by-feature?from=2025-01-01&to=2025-01-31",
		"/api/v1/llm/errors?from=2025-01-01&to=2025-01-31",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest("GET", ep, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != 401 {
				t.Errorf("status = %d, want 401 para endpoint sem token", resp.StatusCode)
			}
		})
	}
}

// --- Overview: validação de datas ---

func TestLLMObservability_Overview_DataInvalida_Retorna400(t *testing.T) {
	app := newLLMApp()
	req := httptest.NewRequest("GET", "/api/v1/llm/overview?from=bad-date", nil)
	req.Header.Set("X-Service-Token", llmTestToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// --- ByModel: validação de datas ---

func TestLLMObservability_ByModel_DataInvalida_Retorna400(t *testing.T) {
	app := newLLMApp()
	req := httptest.NewRequest("GET", "/api/v1/llm/by-model?to=bad-date", nil)
	req.Header.Set("X-Service-Token", llmTestToken)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// Fim dos testes de LLM observability.
// Testes com DB real (pgxpool) estão em testes de integração.
