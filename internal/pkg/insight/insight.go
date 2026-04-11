// Package insight implementa geração de narrativa semanal de métricas via Claude Haiku.
// BKL-864: converte dados de retenção/CSAT/custo em texto acionável em PT-BR.
// Usa chamada HTTP direta à API Anthropic — sem SDK extra no go.mod.
// Custo estimado: ~$0.08/mês com 4 chamadas/mês.
// Fallback: retorna string vazia se ANTHROPIC_API_KEY não configurada ou erro de rede.
package insight

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
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	insightModel    = "claude-haiku-4-5-20251001"
	maxTokens       = 300
	requestTimeout  = 20 * time.Second
)

// MetricsSnapshot contém os dados de métricas para geração da narrativa.
type MetricsSnapshot struct {
	// Retenção
	ChurnRate     float64 // % churn semana atual
	PrevChurnRate float64 // % churn semana anterior
	DAU           int     // usuários ativos diários (média semanal)
	PrevDAU       int     // DAU semana anterior
	// CSAT
	CSATScore     float64 // média CSAT 0-10
	PrevCSATScore float64 // CSAT semana anterior
	NPSScore      int     // NPS atual (-100 a 100)
	// Custos
	APICostBRL   float64 // custo API IA na semana (R$)
	PrevAPICostBRL float64
	// Breakdown por produto
	ProductChurn map[string]float64 // produto → churn%
	// Contexto temporal
	WeekStart string // "2026-04-07"
	WeekEnd   string // "2026-04-13"
}

// Generator gera insights narrativos usando Claude Haiku.
// Nil-safe: se ANTHROPIC_API_KEY não configurada, Generate retorna texto vazio.
type Generator struct {
	apiKey string
	client *http.Client
}

// New cria um Generator. Retorna nil se ANTHROPIC_API_KEY não estiver configurada.
func New() *Generator {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		log.Println("[insight] ANTHROPIC_API_KEY não configurada — narrativa semanal desabilitada")
		return nil
	}
	return &Generator{
		apiKey: key,
		client: &http.Client{Timeout: requestTimeout},
	}
}

// Generate produz um insight narrativo em PT-BR baseado nas métricas fornecidas.
// Em caso de erro ou Generator nil, retorna ("", err).
func (g *Generator) Generate(ctx context.Context, m MetricsSnapshot) (string, error) {
	if g == nil {
		return "", nil
	}

	prompt := buildPrompt(m)

	body, err := json.Marshal(map[string]any{
		"model":      insightModel,
		"max_tokens": maxTokens,
		"system":     systemPrompt,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("insight: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("insight: new request: %w", err)
	}
	req.Header.Set("x-api-key", g.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("insight: HTTP call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("insight: API error %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("insight: decode response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("insight: empty response from Haiku")
	}

	return strings.TrimSpace(result.Content[0].Text), nil
}

func buildPrompt(m MetricsSnapshot) string {
	// Produto com maior churn
	worstProduct := ""
	worstChurn := 0.0
	for prod, churn := range m.ProductChurn {
		if churn > worstChurn {
			worstChurn = churn
			worstProduct = prod
		}
	}

	churnDelta := m.ChurnRate - m.PrevChurnRate
	dauDelta := m.DAU - m.PrevDAU
	csatDelta := m.CSATScore - m.PrevCSATScore
	costDelta := m.APICostBRL - m.PrevAPICostBRL

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Semana: %s a %s\n\n", m.WeekStart, m.WeekEnd))
	sb.WriteString(fmt.Sprintf("Churn semanal: %.1f%% (anterior: %.1f%%, delta: %+.1f%%)\n", m.ChurnRate, m.PrevChurnRate, churnDelta))
	sb.WriteString(fmt.Sprintf("DAU médio: %d usuários (anterior: %d, delta: %+d)\n", m.DAU, m.PrevDAU, dauDelta))
	sb.WriteString(fmt.Sprintf("CSAT: %.1f/10 (anterior: %.1f/10, delta: %+.1f)\n", m.CSATScore, m.PrevCSATScore, csatDelta))
	sb.WriteString(fmt.Sprintf("NPS: %d\n", m.NPSScore))
	sb.WriteString(fmt.Sprintf("Custo IA: R$ %.2f (anterior: R$ %.2f, delta: %+.2f)\n", m.APICostBRL, m.PrevAPICostBRL, costDelta))
	if worstProduct != "" {
		sb.WriteString(fmt.Sprintf("Produto com maior churn: %s (%.1f%%)\n", worstProduct, worstChurn))
	}
	sb.WriteString("\nGere o insight narrativo acionável.")

	return sb.String()
}

const systemPrompt = `Você é um analista de produto sênior do Instituto Itinerante de Tecnologia (IIT).
Analise as métricas da semana e gere um insight narrativo conciso em português do Brasil.

Formato OBRIGATÓRIO (máximo 4 frases, sem bullet points, sem markdown):
1. Frase sobre o indicador mais crítico (churn, DAU ou CSAT)
2. Frase sobre possível causa ou produto afetado
3. Frase com ação recomendada para a próxima semana
4. Frase opcional sobre custo de IA se relevante (>20% variação)

Exemplo de saída: "Churn subiu 8% esta semana. Produto mais afetado: Nitro — possível causa: bug no release de 07/abr. Recomendar rollback imediato e comunicação proativa aos usuários afetados. Custo de IA estável em R$ 12,50."

Seja direto, acionável e específico. Não repita os números brutos — interprete-os.`

// WeeklyInsightReport é o resultado do insight semanal persistido.
type WeeklyInsightReport struct {
	ID        int       `json:"id"`
	WeekStart time.Time `json:"week_start"`
	WeekEnd   time.Time `json:"week_end"`
	Narrative string    `json:"narrative"`
	Metrics   MetricsSnapshot `json:"metrics"`
	CreatedAt time.Time `json:"created_at"`
}
