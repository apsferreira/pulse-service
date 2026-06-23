package handler

// Testes adicionais para cobrir caminhos de erro e edge cases do EventHandler.
// Complementa event_handler_test.go aumentando cobertura dos métodos parcialmente cobertos.

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// --- IngestEvent error path ---

func TestEventHandler_IngestEvent_ServicoRetornaErro_Retorna400(t *testing.T) {
	svc := &mockEventService{ingestErr: errors.New("validação falhou")}
	app := newTestEventApp(svc, "tok")

	payload := domain.IngestEventRequest{Product: "nitro", EventType: "click"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("erro de serviço deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_IngestEvent_BodyInvalido_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "tok")

	req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader([]byte(`{body malformado`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("body inválido deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- TrackEvent error path ---

func TestEventHandler_TrackEvent_ServicoRetornaErro_Retorna400(t *testing.T) {
	svc := &mockEventService{trackErr: errors.New("produto inválido")}
	app := newTestEventApp(svc, "tok")

	uid := "user-abc"
	payload := domain.TrackEventRequest{EventName: "click", Product: "nitro", UserID: &uid}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/track", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("erro de serviço deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- IdentifyUser — mais caminhos ---

func TestEventHandler_IdentifyUser_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockEventService{identifyErr: errors.New("db error")}
	app := newTestEventApp(svc, "tok")

	payload := domain.IdentifyRequest{AnonymousID: "anon-1", UserID: "user-1"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/identify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_IdentifyUser_SemUserID_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "tok")

	payload := domain.IdentifyRequest{AnonymousID: "anon-1"} // user_id ausente
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/identify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("user_id ausente deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

// --- CountEvents — mais caminhos ---

func TestEventHandler_CountEvents_DataFimInvalida_Retorna400(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/count?end_date=data-invalida", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("end_date inválida deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_CountEvents_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockEventService{countErr: errors.New("db error")}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/count", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_CountEvents_DataInicioFormatoYMD_Retorna200(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "svc-tok")

	// Formato YYYY-MM-DD deve ser aceito
	req := httptest.NewRequest("GET", "/api/v1/events/count?start_date=2024-01-01&end_date=2024-12-31", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("data no formato YYYY-MM-DD deveria ser aceita, recebeu %d", resp.StatusCode)
	}
}

// --- ListEvents ---

func TestEventHandler_ListEvents_SemToken_Retorna401(t *testing.T) {
	svc := &mockEventService{}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	// sem token

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("sem token deveria retornar 401, recebeu %d", resp.StatusCode)
	}
}

func TestEventHandler_ListEvents_ComToken_Retorna200(t *testing.T) {
	svc := &mockEventService{
		listResp: domain.EventListResponse{
			Items:  []domain.Event{{Product: "nitro", EventType: "click"}},
			Total:  1,
			Limit:  50,
			Offset: 0,
		},
	}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events?limit=10&offset=0", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("esperado 200, recebido %d", resp.StatusCode)
	}
}

func TestEventHandler_ListEvents_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockEventService{listErr: errors.New("db error")}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- GetSummary error path ---

func TestEventHandler_GetSummary_ServicoRetornaErro_Retorna500(t *testing.T) {
	svc := &mockEventService{summaryErr: errors.New("db error")}
	app := newTestEventApp(svc, "svc-tok")

	req := httptest.NewRequest("GET", "/api/v1/events/summary", nil)
	req.Header.Set("X-Service-Token", "svc-tok")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("erro de serviço deveria retornar 500, recebeu %d", resp.StatusCode)
	}
}

// --- BatchIngest error path ---

func TestEventHandler_BatchIngest_ServicoRetornaErro_Retorna400(t *testing.T) {
	svc := &mockEventService{batchErr: errors.New("batch error")}
	app := newTestEventApp(svc, "tok")

	payload := domain.BatchIngestRequest{
		Events: []domain.IngestEventRequest{
			{Product: "nitro", EventType: "click"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/events/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("erro de serviço deveria retornar 400, recebeu %d", resp.StatusCode)
	}
}
