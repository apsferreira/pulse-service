package handler

// Testes de funções puras e helpers do handler package.
// Cobre: isPaidPlan, isUnlimitedPlan, upgradeGate, WithXxx builders, broadcastEvent.

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- isPaidPlan ---

func TestIsPaidPlan(t *testing.T) {
	tests := []struct {
		plan string
		want bool
	}{
		{"free", false},
		{"starter", true},
		{"growth", true},
		{"pro", true},
		{"enterprise", true},
		{"internal", true},
		{"", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		got := isPaidPlan(tt.plan)
		if got != tt.want {
			t.Errorf("isPaidPlan(%q) = %v, want %v", tt.plan, got, tt.want)
		}
	}
}

// --- isUnlimitedPlan ---

func TestIsUnlimitedPlan(t *testing.T) {
	tests := []struct {
		plan string
		want bool
	}{
		{"free", false},
		{"starter", false},
		{"growth", false},
		{"pro", true},
		{"enterprise", true},
		{"internal", true},
		{"", false},
	}
	for _, tt := range tests {
		got := isUnlimitedPlan(tt.plan)
		if got != tt.want {
			t.Errorf("isUnlimitedPlan(%q) = %v, want %v", tt.plan, got, tt.want)
		}
	}
}

// --- upgradeGate ---

func TestUpgradeGate_RetornaHTTP402(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upgradeGate(c, "free", "starter")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 402 {
		t.Errorf("upgradeGate deveria retornar 402, recebeu %d", resp.StatusCode)
	}
}

func TestUpgradeGate_CodeUpgradeRequired(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upgradeGate(c, "free", "starter")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}

	var body UpgradeGateResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if body.Code != "upgrade_required" {
		t.Errorf("code esperado 'upgrade_required', got %q", body.Code)
	}
	if body.Plan != "free" {
		t.Errorf("plan esperado 'free', got %q", body.Plan)
	}
	if body.UpgradeURL == "" {
		t.Error("upgrade_url não deve ser vazia")
	}
}

func TestUpgradeGate_PlanoGrowth_URLCorreta(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upgradeGate(c, "starter", "growth")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	var body UpgradeGateResponse
	json.NewDecoder(resp.Body).Decode(&body) //nolint:errcheck
	if body.UpgradeURL == "" {
		t.Error("upgrade_url para growth não deve ser vazia")
	}
}

// --- EventHandler builders (With* chain) ---

func TestEventHandler_WithBillingURL_RetornaInstancia(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	h2 := h.WithBillingURL("https://example.com/billing")
	if h2 == nil {
		t.Error("WithBillingURL deveria retornar o handler")
	}
	if h2 != h {
		t.Error("WithBillingURL deveria retornar o mesmo ponteiro")
	}
}

func TestEventHandler_WithRedis_RetornaInstancia(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	h2 := h.WithRedis(nil)
	if h2 == nil {
		t.Error("WithRedis(nil) não deve retornar nil")
	}
}

func TestEventHandler_WithIdentityRepo_NilSafe(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	h2 := h.WithIdentityRepo(nil)
	if h2 == nil {
		t.Error("WithIdentityRepo(nil) não deve retornar nil")
	}
}

func TestEventHandler_WithHub_NilSafe(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	h2 := h.WithHub(nil)
	if h2 == nil {
		t.Error("WithHub(nil) não deve retornar nil")
	}
}

// --- FlagHandler builders ---

func TestFlagHandler_WithSSEBroker_NilSafe(t *testing.T) {
	svc := &mockFlagService{}
	h := NewFlagHandler(svc, "tok")

	h2 := h.WithSSEBroker(nil)
	if h2 == nil {
		t.Error("WithSSEBroker(nil) não deve retornar nil")
	}
}

func TestFlagHandler_WithWebhookHandler_NilSafe(t *testing.T) {
	svc := &mockFlagService{}
	h := NewFlagHandler(svc, "tok")

	h2 := h.WithWebhookHandler(nil)
	if h2 == nil {
		t.Error("WithWebhookHandler(nil) não deve retornar nil")
	}
}

func TestFlagHandler_WithFlagRepo_NilSafe(t *testing.T) {
	svc := &mockFlagService{}
	h := NewFlagHandler(svc, "tok")

	h2 := h.WithFlagRepo(nil)
	if h2 == nil {
		t.Error("WithFlagRepo(nil) não deve retornar nil")
	}
}

func TestFlagHandler_WithProjectRepo_NilSafe(t *testing.T) {
	svc := &mockFlagService{}
	h := NewFlagHandler(svc, "tok")

	h2 := h.WithProjectRepo(nil)
	if h2 == nil {
		t.Error("WithProjectRepo(nil) não deve retornar nil")
	}
}

// --- broadcastEvent com ProjectID válido e hub nil ---

func TestBroadcastEvent_ProjectIDValido_HubNil_NaoPanica(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	projectID := "proj-uuid-123"
	event := &domain.Event{
		Product:   "nitro",
		EventType: "click",
		ProjectID: &projectID,
	}

	// Hub é nil — broadcastEvent deve retornar sem panic
	h.broadcastEvent(event)
}

func TestBroadcastEvent_ProjectIDVazio_NaoPanica(t *testing.T) {
	svc := &mockEventService{}
	h := NewEventHandler(svc, "tok")

	emptyID := ""
	event := &domain.Event{
		Product:   "nitro",
		EventType: "click",
		ProjectID: &emptyID,
	}

	// ProjectID vazio → não deve broadcast
	h.broadcastEvent(event)
}
