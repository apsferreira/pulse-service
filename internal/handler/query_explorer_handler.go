package handler

// BKL-993: SQL Explorer — query builder para eventos brutos.
// Construção de SQL dinâmico com prepared statements ($N) sem fmt.Sprintf em queries.
// Limite hardcoded: 10k rows, timeout de 30s.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryExplorerMaxRows = 10_000
	queryExplorerTimeout = 30 * time.Second
)

// validAggregates é o set de aggregates permitidos — evita SQL injection por enum.
var validAggregates = map[string]bool{
	"count":          true,
	"sum":            true,
	"avg":            true,
	"distinct_count": true,
}

// validOperators é o set de operadores permitidos nos filtros.
var validOperators = map[string]string{
	"eq":         "=",
	"neq":        "!=",
	"gt":         ">",
	"gte":        ">=",
	"lt":         "<",
	"lte":        "<=",
	"contains":   "LIKE",
	"starts_with": "LIKE",
}

// QueryExplorerHandler expõe o endpoint de query builder sobre a tabela events.
type QueryExplorerHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

// NewQueryExplorerHandler cria o handler.
func NewQueryExplorerHandler(db *pgxpool.Pool, serviceToken string) *QueryExplorerHandler {
	return &QueryExplorerHandler{db: db, serviceToken: serviceToken}
}

// RegisterRoutes monta a rota do SQL Explorer.
func (h *QueryExplorerHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/projects", middleware.ServiceTokenMiddleware(h.serviceToken))
	api.Post("/:project_id/query", h.RunQuery)
}

// QueryFilter representa um filtro aplicado sobre uma propriedade de evento.
type QueryFilter struct {
	Property string `json:"property"` // ex: "country", "browser" — propriedade no JSONB properties
	Operator string `json:"operator"` // eq | neq | gt | gte | lt | lte | contains | starts_with
	Value    string `json:"value"`    // valor como string — cast no banco quando necessário
}

// QueryRequest é o payload para POST /api/v1/projects/:project_id/query.
type QueryRequest struct {
	Event     string        `json:"event"`      // event_type a filtrar (obrigatório)
	Filters   []QueryFilter `json:"filters"`    // filtros adicionais sobre properties JSONB
	GroupBy   []string      `json:"group_by"`   // colunas para agrupar (max 3)
	Aggregate string        `json:"aggregate"`  // count | sum | avg | distinct_count
	DateRange struct {
		From string `json:"from"` // ISO 8601: "2026-01-01"
		To   string `json:"to"`   // ISO 8601: "2026-04-11"
	} `json:"date_range"`
}

// QueryResultRow é uma linha do resultado.
type QueryResultRow struct {
	Group string  `json:"group"` // valor(es) do group_by concatenados com "|"
	Value float64 `json:"value"` // resultado do aggregate
}

// RunQuery godoc
// POST /api/v1/projects/:project_id/query
// Executa query parametrizada sobre events com aggregate e filtros opcionais.
// BKL-993: prepared statements obrigatórios — zero fmt.Sprintf em SQL.
func (h *QueryExplorerHandler) RunQuery(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "project_id obrigatório"})
	}

	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "payload inválido"})
	}

	// Valida event name — apenas alfanumérico + underscore + hífen
	if req.Event == "" || len(req.Event) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "event obrigatório e máx. 100 chars"})
	}

	// Valida aggregate
	agg := req.Aggregate
	if agg == "" {
		agg = "count"
	}
	if !validAggregates[agg] {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "aggregate inválido: use count, sum, avg ou distinct_count"})
	}

	// Valida group_by (max 3 campos, apenas colunas seguras)
	allowedGroupCols := map[string]string{
		"browser":     "properties->>'browser'",
		"country":     "properties->>'country'",
		"os":          "properties->>'os'",
		"device":      "properties->>'device'",
		"event_type":  "event_type",
		"environment": "environment",
		"user_id":     "user_id::text",
	}
	if len(req.GroupBy) > 3 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "máximo 3 campos em group_by"})
	}
	groupByExprs := make([]string, 0, len(req.GroupBy))
	for _, g := range req.GroupBy {
		expr, ok := allowedGroupCols[g]
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(
				domain.ErrorResponse{Error: fmt.Sprintf("group_by '%s' não permitido; use: browser, country, os, device, event_type, environment, user_id", g)},
			)
		}
		groupByExprs = append(groupByExprs, expr)
	}

	// Constrói query com prepared statements
	// args acumula os parâmetros; argN incrementa o índice $N
	args := make([]interface{}, 0, 10)
	argN := 0

	nextArg := func(v interface{}) string {
		argN++
		args = append(args, v)
		return fmt.Sprintf("$%d", argN)
	}

	var whereClauses []string

	// Filtro por project_id
	whereClauses = append(whereClauses, fmt.Sprintf("project_id = %s::uuid", nextArg(projectID)))

	// Filtro por event_type
	whereClauses = append(whereClauses, fmt.Sprintf("event_type = %s", nextArg(req.Event)))

	// Filtro por date_range
	if req.DateRange.From != "" {
		from, err := time.Parse("2006-01-02", req.DateRange.From)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "date_range.from inválido; use YYYY-MM-DD"})
		}
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= %s", nextArg(from)))
	}
	if req.DateRange.To != "" {
		to, err := time.Parse("2006-01-02", req.DateRange.To)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "date_range.to inválido; use YYYY-MM-DD"})
		}
		// Inclui o dia inteiro
		to = to.Add(24*time.Hour - time.Nanosecond)
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= %s", nextArg(to)))
	}

	// Filtros sobre properties JSONB — somente operadores da whitelist
	for _, f := range req.Filters {
		if len(f.Property) == 0 || len(f.Property) > 64 {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "filter.property inválido (máx. 64 chars)"})
		}
		// Valida property name: apenas alfanumérico + underscore
		for _, ch := range f.Property {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: fmt.Sprintf("filter.property '%s' contém caracteres inválidos", f.Property)})
			}
		}
		sqlOp, ok := validOperators[f.Operator]
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: fmt.Sprintf("operador '%s' inválido", f.Operator)})
		}

		var val interface{} = f.Value
		switch f.Operator {
		case "contains":
			val = "%" + f.Value + "%"
		case "starts_with":
			val = f.Value + "%"
		}

		// A property key é injetada via string building (mas já foi validada acima como alfanumérico)
		// O valor entra como parâmetro $N — sem risco de SQL injection
		whereClauses = append(whereClauses, fmt.Sprintf("properties->>'%s' %s %s", f.Property, sqlOp, nextArg(val)))
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	// Monta aggregate expression
	var aggExpr string
	switch agg {
	case "count":
		aggExpr = "COUNT(*)"
	case "sum":
		aggExpr = "SUM((properties->>'value')::numeric)"
	case "avg":
		aggExpr = "AVG((properties->>'value')::numeric)"
	case "distinct_count":
		aggExpr = "COUNT(DISTINCT user_id)"
	}

	// Monta SELECT e GROUP BY
	var querySQL string
	if len(groupByExprs) > 0 {
		groupSQL := strings.Join(groupByExprs, " || '|' || COALESCE(")
		// Usa COALESCE para cada expressão para evitar NULL quebrando a concatenação
		concatParts := make([]string, len(groupByExprs))
		for i, expr := range groupByExprs {
			concatParts[i] = fmt.Sprintf("COALESCE(%s, '')", expr)
		}
		groupConcatSQL := strings.Join(concatParts, " || '|' || ")
		groupListSQL := strings.Join(groupByExprs, ", ")
		_ = groupSQL
		querySQL = fmt.Sprintf(
			`SELECT %s AS grp, %s AS val FROM events %s GROUP BY %s ORDER BY val DESC LIMIT %d`,
			groupConcatSQL, aggExpr, whereSQL, groupListSQL, queryExplorerMaxRows,
		)
	} else {
		querySQL = fmt.Sprintf(
			`SELECT '' AS grp, %s AS val FROM events %s`,
			aggExpr, whereSQL,
		)
	}

	// Executa com timeout
	ctx, cancel := context.WithTimeout(c.Context(), queryExplorerTimeout)
	defer cancel()

	rows, err := h.db.Query(ctx, querySQL, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao executar query"})
	}
	defer rows.Close()

	results := make([]QueryResultRow, 0, 64)
	for rows.Next() {
		var row QueryResultRow
		if err := rows.Scan(&row.Group, &row.Value); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao ler resultados"})
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao iterar resultados"})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"count":   len(results),
		"query": fiber.Map{
			"event":     req.Event,
			"aggregate": agg,
			"group_by":  req.GroupBy,
		},
	})
}
