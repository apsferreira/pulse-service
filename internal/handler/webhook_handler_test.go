package handler

// Testes de validação de URL de webhook contra SSRF.
// Garante que URLs internas (K3s, localhost, RFC1918) são rejeitadas
// e que apenas HTTPS externo é aceito.

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestIsAllowedWebhookURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		// URLs externas válidas (devem ser aceitas)
		{"URL HTTPS externa simples", "https://hooks.example.com/webhook", true},
		{"URL HTTPS com path", "https://hooks.example.com/hooks/abc123", true},
		{"URL HTTPS com query", "https://hooks.example.com/webhook?key=val", true},
		{"URL HTTPS Slack", "https://hooks.slack.com/services/T00/B00/xxxx", true},
		{"URL HTTPS Datadog", "https://api.datadoghq.com/api/v1/events", true},
		{"URL HTTPS n8n", "https://n8n.exemplo.com.br/webhook/pulse", true},
		{"URL HTTPS Make/Zapier", "https://hook.us.make.com/abc123def", true},

		// Protocolo não-HTTPS (devem ser rejeitados)
		{"HTTP simples", "http://example.com/webhook", false},
		{"HTTP localhost", "http://localhost:8080/webhook", false},
		{"URL sem protocolo", "example.com/webhook", false},
		{"URL vazia", "", false},
		{"esquema FTP", "ftp://example.com/webhook", false},

		// Localhost (devem ser rejeitados)
		{"localhost HTTPS", "https://localhost:8080/webhook", false},
		{"127.0.0.1", "https://127.0.0.1:8080/webhook", false},
		{"127.0.0.2", "https://127.0.0.2:8443/hook", false},
		{"127.1", "https://127.1:8080/webhook", false},

		// RFC1918 10.x.x.x (devem ser rejeitados)
		{"10.0.0.1", "https://10.0.0.1/webhook", false},
		{"10.96.0.1 (K8s DNS)", "https://10.96.0.1/api/v1/secrets", false},
		{"10.0.0.0/8", "https://10.10.10.10:443/hook", false},
		{"10.244.0.1 (pod IP)", "https://10.244.0.1:3000/hook", false},

		// RFC1918 192.168.x.x (devem ser rejeitados)
		{"192.168.0.1", "https://192.168.0.1/webhook", false},
		{"192.168.1.100", "https://192.168.1.100:8080/hook", false},
		{"192.168.50.1", "https://192.168.50.1/health", false},

		// RFC1918 172.16-31.x.x (devem ser rejeitados)
		{"172.16.0.1", "https://172.16.0.1/webhook", false},
		{"172.20.0.1 (Vault)", "https://172.20.0.1:8200/v1/secret", false},
		{"172.31.255.255", "https://172.31.255.255/hook", false},
		{"172.17.0.1 (Docker bridge)", "https://172.17.0.1:2375/hook", false},
		{"172.18.0.1", "https://172.18.0.1/hook", false},
		{"172.19.0.1", "https://172.19.0.1/hook", false},
		{"172.22.0.1", "https://172.22.0.1/hook", false},

		// 172.x fora do range RFC1918 (devem ser aceitos)
		{"172.32.0.1 (público)", "https://172.32.0.1/webhook", true},
		{"172.15.0.1 (público)", "https://172.15.0.1/webhook", true},
		{"172.33.1.1 (público)", "https://172.33.1.1/webhook", true},

		// Serviços internos K8s (devem ser rejeitados)
		{"argocd.svc.cluster.local", "https://argocd.argocd.svc.cluster.local/api", false},
		{"vault.vault.svc.cluster.local", "https://vault.vault.svc.cluster.local/v1/auth", false},
		{"postgresql.svc.cluster.local", "https://postgresql.svc.cluster.local:5432/health", false},
		{"serviço .svc apenas", "https://meu-servico.namespace.svc/hook", false},
		{"prometheus-operated.svc", "https://prometheus-operated.monitoring.svc:9090/api", false},
		{"subdominio longo .svc", "https://servico.namespace.svc.cluster.local/hook", false},
		{"k8s dashboard", "https://kubernetes-dashboard.kube-system.svc.cluster.local", false},

		// URLs malformadas (devem ser rejeitadas)
		{"string inválida pura", "não é uma url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedWebhookURL(tt.url)
			if got != tt.want {
				t.Errorf("isAllowedWebhookURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestSSHCheckRedirect(t *testing.T) {
	tests := []struct {
		name        string
		redirectURL string
		wantErr     bool
	}{
		{"redirect externo válido", "https://hooks.example.com/hook", false},
		{"redirect para localhost", "https://localhost:8080/hook", true},
		{"redirect para 10.x", "https://10.0.0.1:8200/hook", true},
		{"redirect para .svc", "https://vault.vault.svc.cluster.local/hook", true},
		{"redirect para 192.168.x", "https://192.168.1.1/hook", true},
		{"HTTP redirect", "http://example.com/hook", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.redirectURL, nil)
			if err != nil {
				t.Fatalf("erro ao criar request: %v", err)
			}
			err = sshCheckRedirect(req, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("sshCheckRedirect(%q) = %v, wantErr=%v", tt.redirectURL, err, tt.wantErr)
			}
		})
	}
}

func TestSSHCheckRedirectMaxRedirects(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/hook", nil)
	via := make([]*http.Request, 5)
	for i := range via {
		via[i], _ = http.NewRequest("GET", "https://example.com/redirect", nil)
	}
	err := sshCheckRedirect(req, via)
	if err == nil {
		t.Error("sshCheckRedirect should reject after 5 redirects")
	}
}

// --- formatSlack ---

func TestWebhookHandler_FormatSlack_GeraPayloadCorreto(t *testing.T) {
	h := NewWebhookHandler(nil, "tok")
	payload := map[string]interface{}{"event": "event.tracked", "count": 42}
	body, err := h.formatSlack("event.tracked", payload)
	if err != nil {
		t.Fatalf("formatSlack: %v", err)
	}

	var slack struct {
		Text   string `json:"text"`
		Blocks []any  `json:"blocks"`
	}
	if err := json.Unmarshal(body, &slack); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if slack.Text == "" {
		t.Error("Slack text não deve ser vazio")
	}
	if len(slack.Blocks) == 0 {
		t.Error("Slack blocks não deve ser vazio")
	}
}

// --- formatDatadog ---

func TestWebhookHandler_FormatDatadog_GeraPayloadCorreto(t *testing.T) {
	h := NewWebhookHandler(nil, "tok")
	payload := map[string]interface{}{"event": "flag.evaluated", "flag": "test"}
	body, err := h.formatDatadog("flag.evaluated", payload)
	if err != nil {
		t.Fatalf("formatDatadog: %v", err)
	}

	var dd struct {
		Title     string   `json:"title"`
		AlertType string   `json:"alert_type"`
		Tags      []string `json:"tags"`
	}
	if err := json.Unmarshal(body, &dd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dd.Title != "Pulse: flag.evaluated" {
		t.Errorf("title = %q, want Pulse: flag.evaluated", dd.Title)
	}
	if dd.AlertType != "info" {
		t.Errorf("alert_type = %q, want info", dd.AlertType)
	}
	hasSourceTag := false
	for _, tag := range dd.Tags {
		if tag == "source:pulse" {
			hasSourceTag = true
			break
		}
	}
	if !hasSourceTag {
		t.Error("Datadog tags devem incluir source:pulse")
	}
}

// --- generateSecret ---

func TestGenerateSecret_RetornaStringHex(t *testing.T) {
	secret, err := generateSecret()
	if err != nil {
		t.Fatalf("generateSecret: %v", err)
	}
	if len(secret) != 64 {
		t.Errorf("secret len = %d, want 64 (32 bytes em hex)", len(secret))
	}
	// Verifica se é hex válido
	for _, c := range secret {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("caractere não-hex: %c", c)
		}
	}
}

func TestGenerateSecret_UnicidadeEntreChamadas(t *testing.T) {
	s1, _ := generateSecret()
	s2, _ := generateSecret()
	if s1 == s2 {
		t.Error("dois secrets consecutivos não devem ser iguais")
	}
}

// --- dispatchWithType: valida assinatura HMAC ---

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestWebhookHandler_DispatchWithType_AssinaPayloadComHMAC(t *testing.T) {
	h := NewWebhookHandler(nil, "tok")
	payload := map[string]interface{}{"event": "webhook.test", "data": "hello"}

	captured := make(chan *http.Request, 1)
	h.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			select {
			case captured <- r:
			default:
			}
			return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
		}),
	}

	err := h.dispatchWithType(context.Background(), "https://example.com/hook", "test-secret", WebhookTypeCustom, "webhook.test", payload)
	if err != nil {
		t.Fatalf("dispatchWithType: %v", err)
	}

	req := <-captured
	sig := req.Header.Get("X-Pulse-Signature")
	if !strings.HasPrefix(sig, "sha256=") {
		t.Errorf("X-Pulse-Signature = %q, deve começar com sha256=", sig)
	}
	if req.Header.Get("X-Pulse-Event") != "webhook.test" {
		t.Errorf("X-Pulse-Event = %q, want webhook.test", req.Header.Get("X-Pulse-Event"))
	}
	ct := req.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
