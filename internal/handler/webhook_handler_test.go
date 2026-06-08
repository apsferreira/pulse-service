package handler

// Testes de validação de URL de webhook contra SSRF.
// Garante que URLs internas (K3s, localhost, RFC1918) são rejeitadas
// e que apenas HTTPS externo é aceito.

import (
	"net/http"
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
