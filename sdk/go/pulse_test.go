package pulse

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestTrack_SendsProductAndUserID trava o contrato real com o pulse-service:
// antes de 2026-09-11, Track() não enviava product nem user_id — o que
// levou o pulse-service a normalizar todo evento para product="default"
// (ver EventService.Track no pulse-service) e foi a causa raiz de 13
// serviços do ecossistema terem escrito cliente HTTP próprio em vez de usar
// este SDK.
func TestTrack_SendsProductAndUserID(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events/track" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		if got := r.Header.Get("X-Service-Token"); got != "test-token" {
			t.Errorf("X-Service-Token: got %q, want %q", got, "test-token")
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	c.Track(context.Background(), "checkout.started", "meu-produto", "user-123", Props{"amount": 100})

	select {
	case body := <-received:
		if body["event_name"] != "checkout.started" {
			t.Errorf("event_name: got %v", body["event_name"])
		}
		if body["product"] != "meu-produto" {
			t.Errorf("product: got %v — sem esse campo o pulse-service normaliza pra \"default\"", body["product"])
		}
		if body["user_id"] != "user-123" {
			t.Errorf("user_id: got %v", body["user_id"])
		}
		props, ok := body["properties"].(map[string]interface{})
		if !ok || props["amount"] != float64(100) {
			t.Errorf("properties: got %v", body["properties"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout esperando a requisição fire-and-forget")
	}
}

// TestTrack_EmptyUserIDOmiteNaoQuebra garante que um evento anônimo (sem
// usuário autenticado) ainda envia product corretamente.
func TestTrack_EmptyUserIDOmiteNaoQuebra(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	c.Track(context.Background(), "page.viewed", "meu-produto", "", Props{"path": "/home"})

	select {
	case body := <-received:
		if body["product"] != "meu-produto" {
			t.Errorf("product: got %v", body["product"])
		}
		if body["user_id"] != "" {
			t.Errorf("user_id deveria ser string vazia, got %v", body["user_id"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

// TestTrackWithProject_EnviaProjectID trava o contrato novo (ADR-052, SocialMake
// Opção B): sem project_id, o evento não pode ser consultado depois via
// QueryEvents (filtra por project_id::uuid).
func TestTrackWithProject_EnviaProjectID(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	c.TrackWithProject(context.Background(), "order.paid", "storemake", "user-1", "11111111-1111-1111-1111-111111111111", Props{"attribution_ref": "sm_abc123"})

	select {
	case body := <-received:
		if body["project_id"] != "11111111-1111-1111-1111-111111111111" {
			t.Errorf("project_id: got %v", body["project_id"])
		}
		props, ok := body["properties"].(map[string]interface{})
		if !ok || props["attribution_ref"] != "sm_abc123" {
			t.Errorf("properties.attribution_ref: got %v", body["properties"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

// TestTrackWithProject_SemProjectIDNaoEnviaCampo preserva retrocompatibilidade:
// chamador que não tem project_id ainda não quebra (campo simplesmente omitido).
func TestTrackWithProject_SemProjectIDNaoEnviaCampo(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	c.TrackWithProject(context.Background(), "page.viewed", "storemake", "user-1", "", Props{"path": "/home"})

	select {
	case body := <-received:
		if _, has := body["project_id"]; has {
			t.Errorf("project_id não deveria estar presente quando vazio: %v", body["project_id"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

// TestQueryEvents_MontaFiltroCorreto trava o contrato de QueryEvents com
// query_explorer_handler.go: event, filters (property/operator/value) e
// aggregate chegam no payload exatamente como o handler espera.
func TestQueryEvents_MontaFiltroCorreto(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["event"] != "order.paid" {
			t.Errorf("event: got %v", body["event"])
		}
		filters, _ := body["filters"].([]interface{})
		if len(filters) != 1 {
			t.Fatalf("filters: esperava 1, got %d", len(filters))
		}
		f := filters[0].(map[string]interface{})
		if f["property"] != "attribution_ref" || f["operator"] != "eq" || f["value"] != "sm_abc123" {
			t.Errorf("filter incorreto: %v", f)
		}
		if body["aggregate"] != "count" {
			t.Errorf("aggregate: got %v", body["aggregate"])
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{{"group": "", "value": 3}},
			"count":   1,
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	resp, err := c.QueryEvents(context.Background(), QueryRequest{
		ProjectID: "11111111-1111-1111-1111-111111111111",
		Event:     "order.paid",
		Filters:   []QueryFilter{{Property: "attribution_ref", Operator: "eq", Value: "sm_abc123"}},
		Aggregate: "count",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if gotPath != "/api/v1/projects/11111111-1111-1111-1111-111111111111/query" {
		t.Errorf("path inesperado: %s", gotPath)
	}
	if len(resp.Results) != 1 || resp.Results[0].Value != 3 {
		t.Errorf("resultado inesperado: %+v", resp.Results)
	}
}

// TestQueryEvents_SemProjectID_RetornaErro — ProjectID é obrigatório (vai no
// path da URL, sem ele não há endpoint válido a chamar).
func TestQueryEvents_SemProjectID_RetornaErro(t *testing.T) {
	c := New("http://example.invalid", "test-token")
	_, err := c.QueryEvents(context.Background(), QueryRequest{Event: "order.paid"})
	if err == nil {
		t.Fatal("esperava erro sem ProjectID")
	}
}

// TestQueryEvents_ServidorIndisponivel_RetornaErrQueryUnavailable — STD-017:
// chamador trata erro de indisponibilidade via errors.Is, nunca por string.
func TestQueryEvents_ServidorIndisponivel_RetornaErrQueryUnavailable(t *testing.T) {
	c := New("http://127.0.0.1:1", "test-token")
	_, err := c.QueryEvents(context.Background(), QueryRequest{ProjectID: "x", Event: "order.paid"})
	if !errors.Is(err, ErrQueryUnavailable) {
		t.Errorf("esperava errors.Is(err, ErrQueryUnavailable), got %v", err)
	}
}

// TestQueryEvents_StatusNaoOK_RetornaErrQueryUnavailable
func TestQueryEvents_StatusNaoOK_RetornaErrQueryUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	_, err := c.QueryEvents(context.Background(), QueryRequest{ProjectID: "x", Event: "order.paid"})
	if !errors.Is(err, ErrQueryUnavailable) {
		t.Errorf("esperava errors.Is(err, ErrQueryUnavailable), got %v", err)
	}
}

// TestIdentify_SendsProduct trava o contrato de Identify() com
// IdentifyRequest.Product no pulse-service.
func TestIdentify_SendsProduct(t *testing.T) {
	received := make(chan map[string]string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events/identify" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	c.Identify("meu-produto", "anon-1", "user-123")

	select {
	case body := <-received:
		if body["product"] != "meu-produto" {
			t.Errorf("product: got %v", body["product"])
		}
		if body["anonymous_id"] != "anon-1" || body["user_id"] != "user-123" {
			t.Errorf("ids incorretos: %v", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

// TestTrack_NaoBloqueiaQuandoServerIndisponivel confirma o comportamento
// fire-and-forget: Track() retorna imediatamente mesmo se o pulse-service
// estiver fora do ar.
func TestTrack_NaoBloqueiaQuandoServerIndisponivel(t *testing.T) {
	c := New("http://127.0.0.1:1", "test-token") // porta inexistente, conexão recusada
	done := make(chan struct{})
	go func() {
		c.Track(context.Background(), "evt", "produto", "user", nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Track() bloqueou o caller")
	}
}
