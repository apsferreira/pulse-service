// BKL-1084: ChurnTrendService detecta tenants com health score < 30 por 7 dias consecutivos.
// Dispara alerta de email para ADMIN_EMAIL via Resend API.
// Chamado 1x/dia, após o cálculo de health scores (BKL-801).
package service

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

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	churnTrendThreshold     = 30  // score abaixo deste valor
	churnTrendDays          = 7   // dias consecutivos abaixo do threshold
	churnTrendCheckInterval = 24 * time.Hour
)

// ChurnTrendScheduler detecta e alerta churn trends (7d consecutivos score < 30).
type ChurnTrendScheduler struct {
	db       *pgxpool.Pool
	client   *http.Client
	interval time.Duration
	stop     chan struct{}
}

// NewChurnTrendScheduler cria o scheduler com pool de DB.
func NewChurnTrendScheduler(db *pgxpool.Pool) *ChurnTrendScheduler {
	return &ChurnTrendScheduler{
		db:       db,
		client:   &http.Client{Timeout: 10 * time.Second},
		interval: churnTrendCheckInterval,
		stop:     make(chan struct{}),
	}
}

// Start inicia o loop em background.
func (s *ChurnTrendScheduler) Start(ctx context.Context) {
	go func() {
		// Primeira execução imediata
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

// Stop encerra o scheduler.
func (s *ChurnTrendScheduler) Stop() {
	close(s.stop)
}

// tick executa a verificação diária:
// 1. Persiste o score atual no histórico
// 2. Detecta tenants com 7 dias consecutivos abaixo de 30
// 3. Envia alerta se necessário
func (s *ChurnTrendScheduler) tick(ctx context.Context) {
	// Passo 1: registrar score atual no histórico
	if err := s.recordTodayScores(ctx); err != nil {
		log.Printf("[churn-trend] erro ao registrar histórico: %v", err)
	}

	// Passo 2: buscar tenants em tendência de churn
	tenants, err := s.findTrendingChurners(ctx)
	if err != nil {
		log.Printf("[churn-trend] erro ao buscar tenants em trend: %v", err)
		return
	}
	if len(tenants) == 0 {
		log.Printf("[churn-trend] nenhum tenant com %d dias consecutivos score < %d", churnTrendDays, churnTrendThreshold)
		return
	}

	// Passo 3: enviar alerta
	if err := s.sendAlert(ctx, tenants); err != nil {
		log.Printf("[churn-trend] erro ao enviar alerta: %v", err)
		return
	}
	log.Printf("[churn-trend] alerta enviado para %d tenants em risco crítico de churn", len(tenants))
}

// TrendingChurner representa um tenant em risco crítico.
type TrendingChurner struct {
	TenantID       string
	AvgScore7d     float64
	MinScore7d     int
	ConsecutiveDays int
}

// recordTodayScores insere os scores do dia no histórico.
func (s *ChurnTrendScheduler) recordTodayScores(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenant_health_score_history (tenant_id, score, recorded_at)
		SELECT tenant_id, score, CURRENT_DATE
		FROM tenant_health_scores
		ON CONFLICT (tenant_id, recorded_at) DO UPDATE SET score = EXCLUDED.score
	`)
	return err
}

// findTrendingChurners retorna tenants com score < 30 por 7+ dias consecutivos.
func (s *ChurnTrendScheduler) findTrendingChurners(ctx context.Context) ([]TrendingChurner, error) {
	rows, err := s.db.Query(ctx, `
		WITH recent AS (
		  SELECT
		    tenant_id,
		    score,
		    recorded_at,
		    ROW_NUMBER() OVER (PARTITION BY tenant_id ORDER BY recorded_at DESC) AS rn
		  FROM tenant_health_score_history
		  WHERE recorded_at >= CURRENT_DATE - INTERVAL '7 days'
		    AND score < $1
		),
		consecutive AS (
		  SELECT
		    tenant_id,
		    COUNT(*) AS days_below,
		    AVG(score) AS avg_score,
		    MIN(score) AS min_score
		  FROM recent
		  WHERE rn <= $2
		  GROUP BY tenant_id
		  HAVING COUNT(*) >= $2
		)
		SELECT tenant_id, avg_score, min_score, days_below
		FROM consecutive
		ORDER BY avg_score ASC
	`, churnTrendThreshold, churnTrendDays)
	if err != nil {
		return nil, fmt.Errorf("findTrendingChurners: %w", err)
	}
	defer rows.Close()

	var results []TrendingChurner
	for rows.Next() {
		var t TrendingChurner
		if err := rows.Scan(&t.TenantID, &t.AvgScore7d, &t.MinScore7d, &t.ConsecutiveDays); err != nil {
			continue
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// sendAlert envia email para ADMIN_EMAIL via Resend.
func (s *ChurnTrendScheduler) sendAlert(ctx context.Context, tenants []TrendingChurner) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	if apiKey == "" || adminEmail == "" {
		log.Printf("[churn-trend] RESEND_API_KEY ou ADMIN_EMAIL não configurados — %d tenants em risco ignorados", len(tenants))
		return nil
	}
	if fromEmail == "" {
		fromEmail = "noreply@institutoitinerante.com.br"
	}

	subject := fmt.Sprintf("[Pulse] 🚨 Churn Crítico — %d tenant(s) com score < %d por %d+ dias", len(tenants), churnTrendThreshold, churnTrendDays)
	body := buildTrendAlertEmail(tenants)

	payload, _ := json.Marshal(map[string]any{
		"from":    fromEmail,
		"to":      []string{adminEmail},
		"subject": subject,
		"html":    body,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend error %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

func buildTrendAlertEmail(tenants []TrendingChurner) string {
	var sb strings.Builder
	sb.WriteString(`<div style="font-family:sans-serif;max-width:600px;margin:auto;color:#0D1B26">`)
	sb.WriteString(`<h2 style="color:#DC2626">🚨 Alerta Crítico de Churn</h2>`)
	sb.WriteString(fmt.Sprintf(`<p>Os seguintes tenants estão com health score abaixo de <strong>%d</strong> por <strong>%d+ dias consecutivos</strong>:</p>`, churnTrendThreshold, churnTrendDays))

	sb.WriteString(`<table style="width:100%;border-collapse:collapse;margin-top:16px">`)
	sb.WriteString(`<thead><tr style="background:#F5F8FA">`)
	sb.WriteString(`<th style="padding:8px;text-align:left;border:1px solid #E2E8F0">Tenant</th>`)
	sb.WriteString(`<th style="padding:8px;text-align:center;border:1px solid #E2E8F0">Score Médio 7d</th>`)
	sb.WriteString(`<th style="padding:8px;text-align:center;border:1px solid #E2E8F0">Score Mínimo</th>`)
	sb.WriteString(`<th style="padding:8px;text-align:center;border:1px solid #E2E8F0">Dias Abaixo</th>`)
	sb.WriteString(`</tr></thead><tbody>`)

	for _, t := range tenants {
		sb.WriteString(`<tr>`)
		sb.WriteString(fmt.Sprintf(`<td style="padding:8px;border:1px solid #E2E8F0"><code>%s</code></td>`, t.TenantID))
		sb.WriteString(fmt.Sprintf(`<td style="padding:8px;text-align:center;border:1px solid #E2E8F0;color:#DC2626;font-weight:bold">%.0f</td>`, t.AvgScore7d))
		sb.WriteString(fmt.Sprintf(`<td style="padding:8px;text-align:center;border:1px solid #E2E8F0">%d</td>`, t.MinScore7d))
		sb.WriteString(fmt.Sprintf(`<td style="padding:8px;text-align:center;border:1px solid #E2E8F0">%d</td>`, t.ConsecutiveDays))
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</tbody></table>`)
	sb.WriteString(fmt.Sprintf(`<p style="margin-top:24px;color:#6B7280;font-size:12px">Enviado automaticamente pelo Pulse — %s</p>`, time.Now().Format("02/01/2006 15:04")))
	sb.WriteString(`</div>`)
	return sb.String()
}
