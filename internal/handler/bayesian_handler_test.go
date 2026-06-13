package handler

// BKL-1001: testes do endpoint de análise Bayesiana.
// Cobre validação de query params e o caminho feliz com recomendação determinística
// (seed fixo em service.AnalyzeBayesian garante reproducibilidade).

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/service"
)

func newBayesianApp() *fiber.App {
	app := fiber.New()
	NewBayesianHandler().RegisterRoutes(app)
	return app
}

func TestBayesianHandler_Analyze_Sucesso(t *testing.T) {
	app := newBayesianApp()

	req := httptest.NewRequest("GET",
		"/api/v1/experiments/exp-1/bayesian?a_conversions=50&a_total=1000&b_conversions=80&b_total=1000", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var got service.BayesianResult
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.VariantA.Alpha != 51 || got.VariantA.Beta != 951 {
		t.Errorf("variant A posterior: got alpha=%v beta=%v", got.VariantA.Alpha, got.VariantA.Beta)
	}
	if got.VariantB.Alpha != 81 || got.VariantB.Beta != 921 {
		t.Errorf("variant B posterior: got alpha=%v beta=%v", got.VariantB.Alpha, got.VariantB.Beta)
	}
	if got.ProbBBeatsA < 0.9 {
		t.Errorf("P(B>A) = %v, esperava >= 0.9 (B claramente melhor)", got.ProbBBeatsA)
	}
	if got.Recommendation != "deploy_b" {
		t.Errorf("Recommendation = %q, esperava deploy_b", got.Recommendation)
	}
	if got.SampleSize != 2000 {
		t.Errorf("SampleSize = %d, esperava 2000", got.SampleSize)
	}
}

func TestBayesianHandler_Analyze_CollectMoreData_QuandoAmostraPequena(t *testing.T) {
	app := newBayesianApp()

	req := httptest.NewRequest("GET",
		"/api/v1/experiments/exp-1/bayesian?a_conversions=2&a_total=10&b_conversions=5&b_total=10", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var got service.BayesianResult
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if got.Recommendation != "collect_more_data" {
		t.Errorf("Recommendation = %q, esperava collect_more_data (amostra < 100)", got.Recommendation)
	}
}

func TestBayesianHandler_Analyze_ValidacaoParametros(t *testing.T) {
	cases := []struct {
		name string
		qs   string
		want int
	}{
		{"sem a_conversions", "?a_total=10&b_conversions=5&b_total=10", 400},
		{"a_total zero", "?a_conversions=1&a_total=0&b_conversions=5&b_total=10", 400},
		{"b_total negativo", "?a_conversions=1&a_total=10&b_conversions=5&b_total=-1", 400},
		{"conversions > total", "?a_conversions=20&a_total=10&b_conversions=5&b_total=10", 400},
		{"a_conversions não-inteiro", "?a_conversions=abc&a_total=10&b_conversions=5&b_total=10", 400},
	}
	app := newBayesianApp()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/experiments/exp-1/bayesian"+tc.qs, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
		})
	}
}
