package service

// BKL-814: ChurnAlertScheduler — alerta diário de churn risk para admin IIT.
// Busca tenants com health score < 40 e envia email resumo via Resend API.
// Requer env vars: RESEND_API_KEY, RESEND_FROM, ADMIN_EMAIL.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

const churnRiskThreshold = 40

// ChurnAlertScheduler verifica tenants em risco e notifica o admin IIT.
type ChurnAlertScheduler struct {
	repo     *repository.HealthScoreRepository
	interval time.Duration
	stop     chan struct{}
}

// NewChurnAlertScheduler cria o scheduler com intervalo recomendado de 24h.
func NewChurnAlertScheduler(repo *repository.HealthScoreRepository, interval time.Duration) *ChurnAlertScheduler {
	return &ChurnAlertScheduler{
		repo:     repo,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start inicia o loop em background.
func (s *ChurnAlertScheduler) Start(ctx context.Context) {
	go func() {
		s.tick(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.tick(ctx)
			case <-s.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop para o scheduler.
func (s *ChurnAlertScheduler) Stop() {
	close(s.stop)
}

func (s *ChurnAlertScheduler) tick(ctx context.Context) {
	atRisk, err := s.repo.ListAtRisk(ctx, churnRiskThreshold)
	if err != nil {
		log.Printf("[churn-alert] erro ao buscar tenants em risco: %v", err)
		return
	}
	if len(atRisk) == 0 {
		log.Printf("[churn-alert] nenhum tenant com score < %d", churnRiskThreshold)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Printf("[churn-alert] ADMIN_EMAIL não configurado — ignorando %d tenants em risco", len(atRisk))
		return
	}

	if err := sendChurnAlertEmail(adminEmail, atRisk); err != nil {
		log.Printf("[churn-alert] erro ao enviar email: %v", err)
		return
	}

	log.Printf("[churn-alert] alerta enviado com %d tenants em risco", len(atRisk))
}

func sendChurnAlertEmail(to string, tenants []*domain.TenantHealthScore) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("RESEND_FROM")
	if apiKey == "" || from == "" {
		return fmt.Errorf("RESEND_API_KEY ou RESEND_FROM não configurado")
	}

	subject := fmt.Sprintf("[Pulse] Alerta de Churn Risk — %d tenant(s) com score < %d", len(tenants), churnRiskThreshold)
	htmlBody := buildChurnAlertHTML(tenants)

	payload := map[string]interface{}{
		"from":    from,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("enviar email: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend retornou status %d", resp.StatusCode)
	}
	return nil
}

func buildChurnAlertHTML(tenants []*domain.TenantHealthScore) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"></head>
<body style="font-family:Arial,sans-serif;background:#f9fafb;padding:32px;">
<div style="max-width:700px;margin:0 auto;background:#fff;border-radius:12px;padding:32px;box-shadow:0 1px 3px rgba(0,0,0,.1);">
  <h1 style="color:#dc2626;font-size:20px;margin:0 0 8px;">⚠️ Alerta de Churn Risk — Pulse</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 24px;">`)
	sb.WriteString(fmt.Sprintf("%d tenant(s) com health score abaixo de %d. Requer atenção imediata.", len(tenants), churnRiskThreshold))
	sb.WriteString(`</p>
  <table width="100%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;font-size:13px;">
    <thead>
      <tr style="background:#f3f4f6;">
        <th style="text-align:left;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Tenant ID</th>
        <th style="text-align:center;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Score</th>
        <th style="text-align:center;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Status</th>
        <th style="text-align:left;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Último Login</th>
        <th style="text-align:left;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Último Conteúdo</th>
        <th style="text-align:center;padding:10px 12px;color:#374151;border-bottom:2px solid #e5e7eb;">Calculado em</th>
      </tr>
    </thead>
    <tbody>`)

	for i, hs := range tenants {
		bgColor := "#ffffff"
		if i%2 == 1 {
			bgColor = "#f9fafb"
		}
		statusColor := "#dc2626"
		statusLabel := "🔴 Churning"
		if hs.Score >= 40 {
			statusColor = "#d97706"
			statusLabel = "🟡 At Risk"
		}

		lastLogin := "—"
		if hs.LastLoginAt != nil {
			lastLogin = hs.LastLoginAt.Format("02/01 15:04")
		}
		lastContent := "—"
		if hs.LastContentAt != nil {
			lastContent = hs.LastContentAt.Format("02/01 15:04")
		}

		sb.WriteString(fmt.Sprintf(`
      <tr style="background:%s;">
        <td style="padding:9px 12px;color:#374151;border-bottom:1px solid #e5e7eb;font-family:monospace;font-size:11px;">%s</td>
        <td style="padding:9px 12px;text-align:center;border-bottom:1px solid #e5e7eb;"><strong style="color:%s;">%d</strong></td>
        <td style="padding:9px 12px;text-align:center;border-bottom:1px solid #e5e7eb;font-size:12px;">%s</td>
        <td style="padding:9px 12px;border-bottom:1px solid #e5e7eb;color:#6b7280;">%s</td>
        <td style="padding:9px 12px;border-bottom:1px solid #e5e7eb;color:#6b7280;">%s</td>
        <td style="padding:9px 12px;text-align:center;border-bottom:1px solid #e5e7eb;color:#6b7280;font-size:11px;">%s</td>
      </tr>`,
			bgColor,
			hs.TenantID[:8]+"...", // truncado por segurança
			statusColor,
			hs.Score,
			statusLabel,
			lastLogin,
			lastContent,
			hs.CalculatedAt.Format("02/01 15:04"),
		))
	}

	sb.WriteString(`
    </tbody>
  </table>
  <p style="color:#9ca3af;font-size:11px;margin:24px 0 0;text-align:center;">
    Pulse — IIT Analytics &bull; Alerta automático diário (BKL-814)
  </p>
</div>
</body>
</html>`)

	return sb.String()
}
