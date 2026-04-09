package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- mock de EventServicer ---

type mockEventService struct {
	ingestEvent  *domain.Event
	ingestErr    error
	trackEvent   *domain.Event
	trackErr     error
	summary      domain.EventsSummaryResponse
	summaryErr   error
	identifyCount int
	identifyErr  error
	countResp    domain.EventCountResponse
	countErr     error
	listResp     domain.EventListResponse
	listErr      error
}

func (m *mockEventService) Ingest(ctx context.Context, req domain.IngestEventRequest, ip, ua string) (*domain.Event, error) {
	return m.ingestEvent, m.ingestErr
}

func (m *mockEventService) Track(ctx context.Context, req domain.TrackEventRequest, ip, ua string) (*domain.Event, error) {
	return m.trackEvent, m.trackErr
}

func (m *mockEventService) GetSummary(ctx context.Context) (domain.EventsSummaryResponse, error) {
	return m.summary, m.summaryErr
}

func (m *mockEventService) Identify(ctx context.Context, req domain.IdentifyRequest) (int, error) {
	return m.identifyCount, m.identifyErr
}

func (m *mockEventService) Count(ctx context.Context, eventName string, startDate, endDate *time.Time) (domain.EventCountResponse, error) {
	return m.countResp, m.countErr
}

func (m *mockEventService) List(ctx context.Context, eventName, userID string, limit, offset int) (domain.EventListResponse, error) {
	return m.listResp, m.listErr
}

// newTestEventApp monta um Fiber app com as rotas do EventHandler usando mock de serviço.
func newTestEventApp(svc *mockEventService, token string) *fiber.App {
	h := NewEventHandler(svc, token)
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(domain.ErrorResponse{Error: err.Error()})
		},
	})
	h.RegisterRoutes(app)
	return app
}

// --- POST /api/v1/events/track ---

func TestEventHandler_TrackEvent_Sucesso_Retorna201(t *testing.T) {
	uid := "user-abc"
	svc := &mockEventService{
		trackEvent: &domain.Event{
			Product:   "nitro",
			EventType: "page_view",
			UserID:    &uid,
		},
	}
	app := newTestEventApp(svc, "tok")

	payload := domain.TrackEventRequest{
		EventName: "page_view",
		Product:   "nitro",
		UserID:    &uid,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/track", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("esperado 201, recebido %d", resp.StatusCode)
	}
}

func TestEventHandler_TrackEvent_SemEventName_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "tok")

	// event_name vazio
	payload := domain.TrackEventRequest{Product: "nitro"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/track", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("event_name ausente deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_TrackEvent_BodyVazio_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "tok")

	req := httptest.NewRequest("POST", "/api/v1/events/track", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	// event_name vazio → 400
	if resp.StatusCode != 400 {
		t.Errorf("body vazio deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- POST /api/v1/events (IngestEvent) ---

func TestEventHandler_IngestEvent_Sucesso_Retorna201(t *testing.T) {
	svc := &mockEventService{
		ingestEvent: &domain.Event{Product: "libri", EventType: "book_opened"},
	}
	app := newTestEventApp(svc, "tok")

	payload := domain.IngestEventRequest{
		Product:   "libri",
		EventType: "book_opened",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("esperado 201, recebido %d", resp.StatusCode)
	}
}

// --- GET /api/v1/events/summary (requer token) ---

func TestEventHandler_GetSummary_SemToken_Retorna401(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/summary", nil)
	// Sem X-Service-Token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("sem token deveria retornar 401, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_GetSummary_ComToken_Retorna200(t *testing.T) {
	svc := &mockEventService{
		summary: domain.EventsSummaryResponse{
			Items:        []domain.EventSummaryItem{{Product: "nitro", EventType: "login", Count: 10}},
			TotalLast24h: 10,
		},
	}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/summary", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

// --- POST /api/v1/events/identify ---

func TestEventHandler_IdentifyUser_Sucesso(t *testing.T) {
	svc := &mockEventService{identifyCount: 3}
	app := newTestEventApp(svc, "tok")

	payload := domain.IdentifyRequest{
		AnonymousID: "anon-xyz",
		UserID:      "user-123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/identify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestEventHandler_IdentifyUser_CamposFaltando_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "tok")

	// anonymous_id ausente
	payload := domain.IdentifyRequest{UserID: "user-123"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/identify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("campos faltando deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- GET /api/v1/events/count (requer token) ---

func TestEventHandler_CountEvents_DataInvalida_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/count?start_date=data-errada", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("data inválida deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_CountEvents_Sucesso(t *testing.T) {
	svc := &mockEventService{
		countResp: domain.EventCountResponse{
			Items: []domain.EventCountItem{{EventName: "login", Count: 5}},
			Total: 5,
		},
	}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/count", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}
