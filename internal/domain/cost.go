package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// USDToBRL taxa de conversao fixa para exibicao.
const USDToBRL = 5.00

// APICost representa um registro de custo de chamada de API externa.
type APICost struct {
	ID        uuid.UUID         `json:"id"`
	Product   string            `json:"product"`
	Operation string            `json:"operation"`
	Provider  string            `json:"provider"`
	Model     string            `json:"model,omitempty"`
	TokensIn  int               `json:"tokens_in"`
	TokensOut int               `json:"tokens_out"`
	Images    int               `json:"images"`
	CostUSD   float64           `json:"cost_usd"`
	CostBRL   float64           `json:"cost_brl"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// RecordCostRequest e o payload de ingestao externa (sem id/created_at).
type RecordCostRequest struct {
	Product   string            `json:"product"   validate:"required,max=100"`
	Operation string            `json:"operation" validate:"required,max=100"`
	Provider  string            `json:"provider"  validate:"required,max=50"`
	Model     string            `json:"model"     validate:"max=100"`
	TokensIn  int               `json:"tokens_in"`
	TokensOut int               `json:"tokens_out"`
	Images    int               `json:"images"`
	CostUSD   float64           `json:"cost_usd"  validate:"required"`
	Metadata  map[string]string `json:"metadata"`
}

// CostSummary agrega custos por produto e operacao.
type CostSummary struct {
	Product   string  `json:"product"`
	Operation string  `json:"operation"`
	Count     int     `json:"count"`
	TotalUSD  float64 `json:"total_usd"`
	TotalBRL  float64 `json:"total_brl"`
}

// CostOverview e a visao consolidada para um periodo.
type CostOverview struct {
	From      time.Time     `json:"from"`
	To        time.Time     `json:"to"`
	TotalUSD  float64       `json:"total_usd"`
	TotalBRL  float64       `json:"total_brl"`
	Breakdown []CostSummary `json:"breakdown"`
}

// CostRepository define as operacoes de persistencia de custos.
type CostRepository interface {
	Record(ctx context.Context, cost *APICost) error
	GetOverview(ctx context.Context, product string, from, to time.Time) (*CostOverview, error)
	ListByProduct(ctx context.Context, product string, from, to time.Time, limit int) ([]APICost, error)
	TotalByPeriod(ctx context.Context, from, to time.Time) (float64, error)
}
