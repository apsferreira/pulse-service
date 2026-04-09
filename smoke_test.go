//go:build smoke

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// smokeBase retorna a URL base do servidor real a testar.
// Lê PULSE_BASE_URL ou usa o padrão local.
func smokeBase() string {
	if v := os.Getenv("PULSE_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func smokeGet(t *testing.T, path string, expectedStatus int, headers map[string]string) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", smokeBase()+path, nil)
	if err != nil {
		t.Fatalf("smoke: erro ao criar request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("smoke: erro ao executar request para %s: %v", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != expectedStatus {
		t.Errorf("smoke: GET %s esperado %d, recebido %d — body: %s",
			path, expectedStatus, resp.StatusCode, string(body))
	}
}

// TestSmoke_HealthCheck verifica que o serviço está respondendo.
func TestSmoke_HealthCheck(t *testing.T) {
	smokeGet(t, "/health", http.StatusOK, nil)
}

// TestSmoke_FlagsPublico verifica o endpoint público de flags.
func TestSmoke_FlagsPublico(t *testing.T) {
	smokeGet(t, "/flags", http.StatusOK, nil)
}

// TestSmoke_FlagsAdmin_SemToken verifica que a rota admin rejeita sem token.
func TestSmoke_FlagsAdmin_SemToken(t *testing.T) {
	smokeGet(t, "/api/v1/flags", http.StatusUnauthorized, nil)
}

// TestSmoke_CostsOverview_SemToken verifica autenticação no endpoint de custos.
func TestSmoke_CostsOverview_SemToken(t *testing.T) {
	smokeGet(t, "/api/v1/costs/overview", http.StatusUnauthorized, nil)
}

// TestSmoke_EventsSummary_SemToken verifica autenticação no endpoint de eventos.
func TestSmoke_EventsSummary_SemToken(t *testing.T) {
	smokeGet(t, "/api/v1/events/summary", http.StatusUnauthorized, nil)
}

// TestSmoke_FlagsAdmin_ComToken verifica que token válido permite acesso.
func TestSmoke_FlagsAdmin_ComToken(t *testing.T) {
	token := os.Getenv("PULSE_SERVICE_TOKEN")
	if token == "" {
		t.Skip("PULSE_SERVICE_TOKEN não configurado — pulando teste autenticado")
	}
	smokeGet(t, "/api/v1/flags", http.StatusOK, map[string]string{
		"X-Service-Token": token,
	})
}

func init() {
	fmt.Printf("Smoke tests executando contra: %s\n", smokeBase())
}
