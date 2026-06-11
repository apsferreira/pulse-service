package service

// BKL-1082: Email sequence de onboarding automático para novos tenants do Pulse.
// Sequência via Resend: dia 0 (welcome), dia 1 (SDK reminder), dia 7 (milestone),
// dia 30 (health check), dia 60 (win-back).
// Compliance LGPD: verifica unsubscribes antes de enviar.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const resendAPIURL = "https://api.resend.com/emails"

// EmailSequenceService processa e envia emails de onboarding.
type EmailSequenceService struct {
	db         *pgxpool.Pool
	resendKey  string
	fromEmail  string
	httpClient *http.Client
}

func NewEmailSequenceService(db *pgxpool.Pool) *EmailSequenceService {
	return &EmailSequenceService{
		db:        db,
		resendKey: os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("EMAIL_FROM"), // ex: noreply@institutoitinerante.com.br
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessDue verifica todos os projetos e envia emails cujo trigger_day foi atingido.
// Deve ser chamado 1x/dia via cron ou job.
func (s *EmailSequenceService) ProcessDue(ctx context.Context) error {
	if s.resendKey == "" {
		log.Println("[email_sequence] RESEND_API_KEY não configurado — skip")
		return nil
	}

	// Busca templates ativos
	templates, err := s.listActiveTemplates(ctx)
	if err != nil {
		return fmt.Errorf("email_sequence: listTemplates: %w", err)
	}

	// Busca projetos elegíveis (ativos, com email de contato)
	projects, err := s.listEligibleProjects(ctx)
	if err != nil {
		return fmt.Errorf("email_sequence: listProjects: %w", err)
	}

	sent := 0
	for _, proj := range projects {
		for _, tmpl := range templates {
			daysSinceCreation := int(time.Since(proj.CreatedAt).Hours() / 24)
			if daysSinceCreation < tmpl.TriggerDay {
				continue // ainda não chegou o dia
			}

			// Verifica se já foi enviado
			alreadySent, err := s.wasAlreadySent(ctx, proj.ID, tmpl.ID)
			if err != nil || alreadySent {
				continue
			}

			// Verifica unsubscribe
			unsubscribed, err := s.isUnsubscribed(ctx, proj.ContactEmail)
			if err != nil || unsubscribed {
				s.markSkipped(ctx, proj.ID, tmpl.ID, proj.ContactEmail)
				continue
			}

			// Monta body com variáveis
			body := s.renderTemplate(tmpl.BodyMD, map[string]string{
				"name":         proj.Name,
				"project_name": proj.Name,
				"api_key":      proj.APIKey,
			})

			if err := s.sendEmail(proj.ContactEmail, tmpl.Subject, body); err != nil {
				s.markFailed(ctx, proj.ID, tmpl.ID, proj.ContactEmail, err.Error())
				log.Printf("[email_sequence] falha ao enviar %s para project_id=%s: %v", tmpl.Name, proj.ID, err)
				continue
			}

			s.markSent(ctx, proj.ID, tmpl.ID, proj.ContactEmail)
			sent++
		}
	}

	log.Printf("[email_sequence] processamento concluído — %d emails enviados", sent)
	return nil
}

// renderTemplate substitui variáveis {{key}} no template.
func (s *EmailSequenceService) renderTemplate(tmpl string, vars map[string]string) string {
	for k, v := range vars {
		tmpl = strings.ReplaceAll(tmpl, "{{"+k+"}}", v)
	}
	return tmpl
}

// sendEmail envia email via Resend API.
func (s *EmailSequenceService) sendEmail(to, subject, markdownBody string) error {
	payload := map[string]interface{}{
		"from":    s.fromEmail,
		"to":      []string{to},
		"subject": subject,
		"text":    markdownBody, // Resend suporta text + html — usando text simples
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.resendKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}

// --- helpers de banco ---

type projectRow struct {
	ID           string
	Name         string
	APIKey       string
	ContactEmail string
	Plan         string
	CreatedAt    time.Time
}

type templateRow struct {
	ID         string
	Name       string
	TriggerDay int
	Subject    string
	BodyMD     string
}

func (s *EmailSequenceService) listActiveTemplates(ctx context.Context) ([]templateRow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, trigger_day, subject, body_md
		FROM email_sequence_templates
		WHERE is_active = true
		ORDER BY trigger_day ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []templateRow
	for rows.Next() {
		var t templateRow
		if err := rows.Scan(&t.ID, &t.Name, &t.TriggerDay, &t.Subject, &t.BodyMD); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *EmailSequenceService) listEligibleProjects(ctx context.Context) ([]projectRow, error) {
	// Busca projetos com email de contato via JOIN com a tabela de projetos
	// Por ora usa o api_key como fallback e owner_email se existir
	rows, err := s.db.Query(ctx, `
		SELECT id, name, api_key, COALESCE(owner_email, ''), COALESCE(plan, 'free'), created_at
		FROM projects
		WHERE is_active = true
		  AND COALESCE(owner_email, '') != ''
	`)
	if err != nil {
		// Se owner_email não existir na tabela ainda, retorna vazio sem erro
		log.Printf("[email_sequence] listEligibleProjects: %v (pode ser colunas ausentes)", err)
		return nil, nil
	}
	defer rows.Close()

	var out []projectRow
	for rows.Next() {
		var p projectRow
		if err := rows.Scan(&p.ID, &p.Name, &p.APIKey, &p.ContactEmail, &p.Plan, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *EmailSequenceService) wasAlreadySent(ctx context.Context, projectID, templateID string) (bool, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM email_sequence_events
		WHERE project_id = $1 AND template_id = $2 AND status = 'sent'
	`, projectID, templateID).Scan(&count)
	return count > 0, err
}

func (s *EmailSequenceService) isUnsubscribed(ctx context.Context, email string) (bool, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM email_unsubscribes WHERE email = $1
	`, email).Scan(&count)
	return count > 0, err
}

func (s *EmailSequenceService) markSent(ctx context.Context, projectID, templateID, email string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO email_sequence_events (project_id, template_id, email, status)
		VALUES ($1, $2, $3, 'sent')
		ON CONFLICT (project_id, template_id) DO UPDATE SET status = 'sent', sent_at = NOW()
	`, projectID, templateID, email)
}

func (s *EmailSequenceService) markFailed(ctx context.Context, projectID, templateID, email, errMsg string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO email_sequence_events (project_id, template_id, email, status, error)
		VALUES ($1, $2, $3, 'failed', $4)
		ON CONFLICT (project_id, template_id) DO UPDATE SET status = 'failed', error = $4
	`, projectID, templateID, email, errMsg)
}

func (s *EmailSequenceService) markSkipped(ctx context.Context, projectID, templateID, email string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO email_sequence_events (project_id, template_id, email, status)
		VALUES ($1, $2, $3, 'skipped')
		ON CONFLICT (project_id, template_id) DO NOTHING
	`, projectID, templateID, email)
}

