package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CostRepository struct {
	db *pgxpool.Pool
}

func NewCostRepository(db *pgxpool.Pool) *CostRepository {
	return &CostRepository{db: db}
}

// Record persiste um novo registro de custo de API.
func (r *CostRepository) Record(ctx context.Context, cost *APICost) error {
	cost.ID = uuid.New()
	cost.CreatedAt = time.Now().UTC()

	meta, err := json.Marshal(cost.Metadata)
	if err != nil {
		return fmt.Errorf("cost_repository: Record marshal metadata: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO api_costs
		 (id, product, operation, provider, model, tokens_in, tokens_out, images, cost_usd, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		cost.ID, cost.Product, cost.Operation, cost.Provider, cost.Model,
		cost.TokensIn, cost.TokensOut, cost.Images, cost.CostUSD,
		meta, cost.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("cost_repository: Record: %w", err)
	}
	return nil
}

// GetOverview retorna o total e o breakdown por produto/operacao para o periodo.
// Se product for vazio, agrega todos os produtos.
func (r *CostRepository) GetOverview(ctx context.Context, product string, from, to time.Time) (*domain.CostOverview, error) {
	var (
		rows interface{ Close() }
		err  error
	)

	type breakdownRow struct {
		product   string
		operation string
		count     int
		totalUSD  float64
	}

	var breakdownArgs []interface{}
	var breakdownQuery string

	if product != "" {
		breakdownQuery = `
			SELECT product, operation, COUNT(*) AS cnt, COALESCE(SUM(cost_usd), 0) AS total_usd
			FROM api_costs
			WHERE product = $1 AND created_at >= $2 AND created_at <= $3
			GROUP BY product, operation
			ORDER BY total_usd DESC`
		breakdownArgs = []interface{}{product, from, to}
	} else {
		breakdownQuery = `
			SELECT product, operation, COUNT(*) AS cnt, COALESCE(SUM(cost_usd), 0) AS total_usd
			FROM api_costs
			WHERE created_at >= $1 AND created_at <= $2
			GROUP BY product, operation
			ORDER BY total_usd DESC`
		breakdownArgs = []interface{}{from, to}
	}

	pgRows, err := r.db.Query(ctx, breakdownQuery, breakdownArgs...)
	if err != nil {
		return nil, fmt.Errorf("cost_repository: GetOverview query: %w", err)
	}
	defer pgRows.Close()
	rows = pgRows

	var breakdown []domain.CostSummary
	var totalUSD float64

	for pgRows.Next() {
		var br breakdownRow
		if err := pgRows.Scan(&br.product, &br.operation, &br.count, &br.totalUSD); err != nil {
			return nil, fmt.Errorf("cost_repository: GetOverview scan: %w", err)
		}
		totalUSD += br.totalUSD
		breakdown = append(breakdown, domain.CostSummary{
			Product:   br.product,
			Operation: br.operation,
			Count:     br.count,
			TotalUSD:  br.totalUSD,
			TotalBRL:  br.totalUSD * domain.USDToBRL,
		})
	}
	if err := pgRows.Err(); err != nil {
		return nil, fmt.Errorf("cost_repository: GetOverview rows: %w", err)
	}

	_ = rows // satisfaz linter — rows.Close() ja foi deferido via pgRows

	if breakdown == nil {
		breakdown = []domain.CostSummary{}
	}

	return &domain.CostOverview{
		From:      from,
		To:        to,
		TotalUSD:  totalUSD,
		TotalBRL:  totalUSD * domain.USDToBRL,
		Breakdown: breakdown,
	}, nil
}

// ListByProduct retorna registros individuais filtrados por produto e periodo.
func (r *CostRepository) ListByProduct(ctx context.Context, product string, from, to time.Time, limit int) ([]domain.APICost, error) {
	var (
		query string
		args  []interface{}
	)

	if product != "" {
		query = `
			SELECT id, product, operation, provider, COALESCE(model, ''), tokens_in, tokens_out, images, cost_usd, COALESCE(metadata, '{}'), created_at
			FROM api_costs
			WHERE product = $1 AND created_at >= $2 AND created_at <= $3
			ORDER BY created_at DESC
			LIMIT $4`
		args = []interface{}{product, from, to, limit}
	} else {
		query = `
			SELECT id, product, operation, provider, COALESCE(model, ''), tokens_in, tokens_out, images, cost_usd, COALESCE(metadata, '{}'), created_at
			FROM api_costs
			WHERE created_at >= $1 AND created_at <= $2
			ORDER BY created_at DESC
			LIMIT $3`
		args = []interface{}{from, to, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cost_repository: ListByProduct query: %w", err)
	}
	defer rows.Close()

	var costs []domain.APICost
	for rows.Next() {
		var c domain.APICost
		var metaRaw []byte
		if err := rows.Scan(
			&c.ID, &c.Product, &c.Operation, &c.Provider, &c.Model,
			&c.TokensIn, &c.TokensOut, &c.Images, &c.CostUSD,
			&metaRaw, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("cost_repository: ListByProduct scan: %w", err)
		}
		c.CostBRL = c.CostUSD * domain.USDToBRL
		if metaRaw != nil {
			if err := json.Unmarshal(metaRaw, &c.Metadata); err != nil {
				c.Metadata = make(map[string]string)
			}
		}
		costs = append(costs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cost_repository: ListByProduct rows: %w", err)
	}
	if costs == nil {
		costs = []domain.APICost{}
	}
	return costs, nil
}

// TotalByPeriod retorna o custo total em USD para o periodo inteiro (todos os produtos).
func (r *CostRepository) TotalByPeriod(ctx context.Context, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(cost_usd), 0) FROM api_costs WHERE created_at >= $1 AND created_at <= $2`,
		from, to,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("cost_repository: TotalByPeriod: %w", err)
	}
	return total, nil
}

// APICost e um alias local para evitar repetir domain.APICost nos campos privados do Record.
// Aqui usamos domain.APICost diretamente; este type alias resolve a referencia sem conflito.
type APICost = domain.APICost
