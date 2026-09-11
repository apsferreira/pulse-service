package pulse

import (
	"context"
	"encoding/json"
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
