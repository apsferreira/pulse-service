// Package modelrevalidation implementa o job semanal de revalidação dos modelos ML.
// fix #66: detecta drift do churn model (regressão logística por projeto) e recalcula
// pesos dos experimentos MAB ativos para evitar priors desatualizados.
//
// Para cada projeto com churn model treinado:
//   1. Recalcula acurácia do modelo persistido sobre uma amostra fresca dos últimos 90 dias
//   2. Persiste last_validated_at + last_validation_accuracy
//   3. Insere evento 'model.performance' na tabela events (project_id, model_version, accuracy)
//   4. Se accuracy < accuracyAlertThreshold, dispara notificação via notification-service
//
// Para cada experimento MAB ativo: chama RecalculateWeights (mesma lógica do endpoint manual).
package modelrevalidation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AccuracyAlertThreshold — accuracy abaixo deste valor dispara notificação.
const AccuracyAlertThreshold = 0.70

// ChurnRevalidator é a interface do ChurnModelHandler usada pelo job.
type ChurnRevalidator interface {
	RevalidateAll(ctx context.Context) ([]ChurnRevalidationResult, error)
}

// ChurnRevalidationResult espelha handler.RevalidationResult sem import cíclico.
type ChurnRevalidationResult struct {
	ProjectID    string
	ModelVersion int
	Accuracy     float64
	SampleSize   int
}

// MABRevalidator é a interface mínima do MABRepository usada pelo job.
type MABRevalidator interface {
	ListActiveExperimentIDs(ctx context.Context) ([]uuid.UUID, error)
	RecalculateWeights(ctx context.Context, experimentID uuid.UUID) error
}

// Scheduler orquestra a revalidação semanal de churn model + MAB.
type Scheduler struct {
	db                 *pgxpool.Pool
	churn              ChurnRevalidator
	mab                MABRevalidator
	notificationURL    string
	notificationToken  string
	interval           time.Duration
}

// New cria o scheduler. interval default = 7 dias.
func New(db *pgxpool.Pool, churn ChurnRevalidator, mab MABRevalidator, notificationURL, notificationToken string, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 7 * 24 * time.Hour
	}
	return &Scheduler{
		db:                db,
		churn:             churn,
		mab:               mab,
		notificationURL:   notificationURL,
		notificationToken: notificationToken,
		interval:          interval,
	}
}

// Start inicia o loop em background; respeita ctx.Done para parar.
func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("[model-revalidation] iniciado — intervalo: %s", s.interval)
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		s.run(ctx)
		for {
			select {
			case <-ctx.Done():
				log.Println("[model-revalidation] parado")
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()
}

func (s *Scheduler) run(ctx context.Context) {
	s.revalidateChurnModels(ctx)
	s.recalculateMABWeights(ctx)
}

func (s *Scheduler) revalidateChurnModels(ctx context.Context) {
	if s.churn == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	results, err := s.churn.RevalidateAll(cctx)
	if err != nil {
		log.Printf("[model-revalidation] churn: %v", err)
		return
	}
	log.Printf("[model-revalidation] churn: %d projeto(s) revalidado(s)", len(results))

	for _, r := range results {
		s.logPerformanceEvent(ctx, r)
		if r.Accuracy < AccuracyAlertThreshold {
			s.sendDriftAlert(ctx, r)
		}
	}
}

func (s *Scheduler) recalculateMABWeights(ctx context.Context) {
	if s.mab == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ids, err := s.mab.ListActiveExperimentIDs(cctx)
	if err != nil {
		log.Printf("[model-revalidation] mab list: %v", err)
		return
	}
	updated := 0
	for _, id := range ids {
		if err := s.mab.RecalculateWeights(cctx, id); err != nil {
			log.Printf("[model-revalidation] mab recalc %s: %v", id, err)
			continue
		}
		updated++
	}
	log.Printf("[model-revalidation] mab: %d experimento(s) com pesos recalculados", updated)
}

// logPerformanceEvent grava um evento 'model.performance' na tabela events para auditoria.
// Não falha o job se a inserção der erro — apenas loga.
func (s *Scheduler) logPerformanceEvent(ctx context.Context, r ChurnRevalidationResult) {
	props := map[string]interface{}{
		"model_version": r.ModelVersion,
		"accuracy":      r.Accuracy,
		"sample_size":   r.SampleSize,
	}
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err = s.db.Exec(cctx,
		`INSERT INTO events (id, product, event_type, properties, environment, project_id, created_at)
		 VALUES (gen_random_uuid()::text, 'pulse', 'model.performance', $1::jsonb, 'production', $2::uuid, NOW())`,
		string(propsJSON), r.ProjectID,
	)
	if err != nil {
		log.Printf("[model-revalidation] log event %s: %v", r.ProjectID, err)
	}
}

// sendDriftAlert chama POST {notificationURL}/api/v1/notifications com payload de drift.
// Silencioso se notificationURL não estiver configurado (self-hosted sem notification-service).
func (s *Scheduler) sendDriftAlert(ctx context.Context, r ChurnRevalidationResult) {
	if s.notificationURL == "" {
		log.Printf("[model-revalidation] DRIFT project=%s accuracy=%.4f (notification-service não configurado)", r.ProjectID, r.Accuracy)
		return
	}
	payload := map[string]interface{}{
		"channel":  "email",
		"template": "model_drift",
		"subject":  fmt.Sprintf("[Pulse] Churn model drift — accuracy %.2f%% < %.0f%%", r.Accuracy*100, AccuracyAlertThreshold*100),
		"body": fmt.Sprintf(
			"O modelo de churn do projeto %s caiu para accuracy=%.4f (model_version=%d, sample_size=%d). Recomendado retreinar via POST /churn-model/train.",
			r.ProjectID, r.Accuracy, r.ModelVersion, r.SampleSize,
		),
		"metadata": map[string]interface{}{
			"project_id":    r.ProjectID,
			"model_version": r.ModelVersion,
			"accuracy":      r.Accuracy,
			"threshold":     AccuracyAlertThreshold,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[model-revalidation] alert marshal: %v", err)
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.notificationURL+"/api/v1/notifications", bytes.NewReader(body))
	if err != nil {
		log.Printf("[model-revalidation] alert req: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.notificationToken != "" {
		req.Header.Set("X-Service-Token", s.notificationToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[model-revalidation] alert send: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("[model-revalidation] alert status=%d project=%s", resp.StatusCode, r.ProjectID)
	}
}
