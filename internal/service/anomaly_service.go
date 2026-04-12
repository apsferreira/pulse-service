// Package service — AnomalyService: detecção de anomalias em métricas de eventos (BKL-865).
// Algoritmo: média móvel + desvio padrão das últimas N semanas.
// Detecta valores >2σ acima ou abaixo da média histórica.
// Gemini 2.5 Flash gera explicação human-readable via HTTP direto (sem SDK extra).
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AnomalyDirection indica se a anomalia é um pico ou uma queda.
type AnomalyDirection string

const (
	AnomalyDirectionSpike AnomalyDirection = "spike"
	AnomalyDirectionDrop  AnomalyDirection = "drop"
	AnomalyDirectionNone  AnomalyDirection = "none"

	anomalyGeminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-04-17:generateContent"
	anomalyHTTPTimeout    = 20 * time.Second
	sigmaThreshold        = 2.0
)

// AnomalyResult é a resposta do endpoint de anomalias.
type AnomalyResult struct {
	Metric          string           `json:"metric"`
	CurrentValue    float64          `json:"current_value"`
	ExpectedValue   float64          `json:"expected_value"`
	DeviationSigma  float64          `json:"deviation_sigma"`
	IsAnomaly       bool             `json:"is_anomaly"`
	Direction       AnomalyDirection `json:"direction"`
	Explanation     string           `json:"explanation"`
	WindowWeeks     int              `json:"window_weeks"`
	ProjectID       string           `json:"project_id"`
}

// AnomalyService detecta anomalias em séries temporais de eventos.
type AnomalyService struct {
	db         *pgxpool.Pool
	httpClient *http.Client
	geminiKey  string
}

// NewAnomalyService cria o serviço de anomalias.
func NewAnomalyService(db *pgxpool.Pool) *AnomalyService {
	return &AnomalyService{
		db:         db,
		httpClient: &http.Client{Timeout: anomalyHTTPTimeout},
		geminiKey:  os.Getenv("GEMINI_API_KEY"),
	}
}

// Detect analisa uma métrica e retorna o resultado de anomalia.
// window é "7d", "14d", "28d" etc. projectID filtra por projeto específico.
func (s *AnomalyService) Detect(ctx context.Context, projectID, metric, window string) (*AnomalyResult, error) {
	windowDays, err := parseWindowDays(window)
	if err != nil {
		return nil, fmt.Errorf("anomaly: %w", err)
	}
	windowWeeks := windowDays / 7
	if windowWeeks < 1 {
		windowWeeks = 1
	}

	// Coleta série histórica de contagens semanais para a métrica
	weeklyValues, err := s.getWeeklyCounts(ctx, projectID, metric, windowWeeks+4) // +4 semanas para contexto
	if err != nil {
		return nil, fmt.Errorf("anomaly: coletar série: %w", err)
	}
	if len(weeklyValues) < 2 {
		return &AnomalyResult{
			Metric:         metric,
			ProjectID:      projectID,
			IsAnomaly:      false,
			Direction:      AnomalyDirectionNone,
			WindowWeeks:    windowWeeks,
			Explanation:    "Dados insuficientes para análise de anomalia (mínimo 2 semanas)",
		}, nil
	}

	// Valor atual = semana mais recente
	current := weeklyValues[len(weeklyValues)-1]
	// Histórico = todas as semanas menos a atual
	historical := weeklyValues[:len(weeklyValues)-1]

	mean, stddev := calcStats(historical)
	var deviationSigma float64
	if stddev > 0 {
		deviationSigma = (current - mean) / stddev
	}

	isAnomaly := math.Abs(deviationSigma) >= sigmaThreshold
	direction := AnomalyDirectionNone
	if isAnomaly {
		if current > mean {
			direction = AnomalyDirectionSpike
		} else {
			direction = AnomalyDirectionDrop
		}
	}

	result := &AnomalyResult{
		Metric:         metric,
		CurrentValue:   current,
		ExpectedValue:  math.Round(mean*100) / 100,
		DeviationSigma: math.Round(deviationSigma*100) / 100,
		IsAnomaly:      isAnomaly,
		Direction:      direction,
		WindowWeeks:    windowWeeks,
		ProjectID:      projectID,
	}

	// Gerar explicação via Gemini (apenas se anomalia detectada; fallback se API indisponível)
	if isAnomaly {
		explanation, err := s.generateExplanation(ctx, result, historical)
		if err != nil {
			log.Printf("[anomaly] Gemini indisponível, usando explicação padrão: %v", err)
			result.Explanation = buildDefaultExplanation(result)
		} else {
			result.Explanation = explanation
		}
	} else {
		result.Explanation = fmt.Sprintf("Métrica %s está dentro do padrão normal (%.1fσ do esperado).", metric, deviationSigma)
	}

	return result, nil
}

// getWeeklyCounts retorna contagens semanais ordenadas do mais antigo para o mais recente.
func (s *AnomalyService) getWeeklyCounts(ctx context.Context, projectID, metric string, numWeeks int) ([]float64, error) {
	since := time.Now().UTC().AddDate(0, 0, -numWeeks*7)

	query := `
		SELECT
			DATE_TRUNC('week', created_at) AS week,
			COUNT(*) AS cnt
		FROM events
		WHERE event_type = $1
		  AND created_at >= $2`
	args := []any{metric, since}

	if projectID != "" {
		query += ` AND project_id = $3`
		args = append(args, projectID)
	}
	query += `
		GROUP BY week
		ORDER BY week ASC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("getWeeklyCounts query: %w", err)
	}
	defer rows.Close()

	var values []float64
	for rows.Next() {
		var week time.Time
		var cnt int64
		if err := rows.Scan(&week, &cnt); err != nil {
			return nil, fmt.Errorf("getWeeklyCounts scan: %w", err)
		}
		values = append(values, float64(cnt))
	}
	return values, rows.Err()
}

// calcStats calcula média e desvio padrão de uma série.
func calcStats(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	stddev = math.Sqrt(variance)
	return
}

// generateExplanation chama Gemini 2.5 Flash para gerar explicação human-readable.
func (s *AnomalyService) generateExplanation(ctx context.Context, result *AnomalyResult, historical []float64) (string, error) {
	if s.geminiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	direction := "aumento"
	if result.Direction == AnomalyDirectionDrop {
		direction = "queda"
	}
	pctChange := 0.0
	if result.ExpectedValue > 0 {
		pctChange = ((result.CurrentValue - result.ExpectedValue) / result.ExpectedValue) * 100
	}

	prompt := fmt.Sprintf(`Você é um analista de métricas de produto.
Uma anomalia foi detectada na métrica "%s" do projeto "%s".

DADOS:
- Valor atual: %.0f eventos na semana
- Valor esperado (média histórica): %.0f eventos
- Variação: %+.1f%%
- Desvio: %.1fσ
- Direção: %s

Gere uma explicação concisa em português brasileiro, no formato:
"%s de X%% em <métrica> detectad<o/a>. Possíveis causas: [causa 1, causa 2]. Recomendação: [ação]."

Máximo 2 frases. Seja específico e acionável.`,
		result.Metric, result.ProjectID,
		result.CurrentValue, result.ExpectedValue,
		pctChange, result.DeviationSigma,
		direction,
		capitalizeFirst(direction),
	)

	reqBody := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 200,
			"temperature":     0.2,
		},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", anomalyGeminiEndpoint, s.geminiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// buildDefaultExplanation gera explicação sem LLM (fallback).
func buildDefaultExplanation(r *AnomalyResult) string {
	pctChange := 0.0
	if r.ExpectedValue > 0 {
		pctChange = ((r.CurrentValue - r.ExpectedValue) / r.ExpectedValue) * 100
	}
	direction := "Aumento"
	if r.Direction == AnomalyDirectionDrop {
		direction = "Queda"
	}
	return fmt.Sprintf(
		"%s de %.0f%% em %s detectado (%.1fσ). Verifique releases recentes, campanhas de marketing ou problemas técnicos.",
		direction, math.Abs(pctChange), r.Metric, r.DeviationSigma,
	)
}

// parseWindowDays converte string "7d", "14d", "28d" em dias inteiros.
func parseWindowDays(window string) (int, error) {
	if window == "" {
		return 7, nil
	}
	if len(window) < 2 {
		return 0, fmt.Errorf("window inválido: %q (use 7d, 14d, 28d)", window)
	}
	suffix := window[len(window)-1]
	numStr := window[:len(window)-1]

	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil || n <= 0 {
		return 0, fmt.Errorf("window inválido: %q", window)
	}
	switch suffix {
	case 'd':
		return n, nil
	case 'w':
		return n * 7, nil
	default:
		return 0, fmt.Errorf("window sufixo desconhecido: %q (use d ou w)", string(suffix))
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return string([]rune(s)[0]-32) + s[1:]
}
