package handler

// GET /api/v1/users                        — lista usuários únicos com contagens
// GET /api/v1/users/:user_id/timeline      — flag evaluations + events misturados
// GET /api/v1/users/:user_id/identities    — cross-produto via user_identities

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserInsightsHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewUserInsightsHandler(db *pgxpool.Pool, serviceToken string) *UserInsightsHandler {
	return &UserInsightsHandler{db: db, serviceToken: serviceToken}
}

func (h *UserInsightsHandler) RegisterRoutes(app *fiber.App) {
	auth := middleware.ServiceTokenMiddleware(h.serviceToken)
	app.Get("/api/v1/users", auth, h.ListUsers)
	app.Get("/api/v1/users/:user_id/timeline", auth, h.GetTimeline)
	app.Get("/api/v1/users/:user_id/identities", auth, h.GetIdentities)
}

type UserSummary struct {
	UserID     string    `json:"user_id"`
	EventCount int       `json:"event_count"`
	FlagCount  int       `json:"flag_count"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Products   []string  `json:"products"`
}

type TimelineItem struct {
	Kind         string                 `json:"kind"` // "flag" | "event"
	Timestamp    time.Time              `json:"timestamp"`
	FlagKey      *string                `json:"flag_key,omitempty"`
	FlagResult   *bool                  `json:"flag_result,omitempty"`
	FlagReason   *string                `json:"flag_reason,omitempty"`
	EventType    *string                `json:"event_type,omitempty"`
	EventProduct *string                `json:"event_product,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// ListUsers GET /api/v1/users?project_id=&flag_key=&product=&since=24h|7d|30d&search=&limit=&offset=
func (h *UserInsightsHandler) ListUsers(c *fiber.Ctx) error {
	projectID := c.Query("project_id")
	flagKey := c.Query("flag_key")
	product := c.Query("product")
	search := c.Query("search")
	since := c.Query("since", "7d")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	sinceTime := userSinceParam(since)
	args := []interface{}{sinceTime}
	idx := 2

	flagCond := ""
	if flagKey != "" {
		flagCond = fmt.Sprintf(" AND fe.flag_key = $%d", idx)
		args = append(args, flagKey)
		idx++
	}
	flagProjCond := ""
	if projectID != "" {
		flagProjCond = fmt.Sprintf(" AND fe.project_id::text = $%d", idx)
		args = append(args, projectID)
		idx++
	}
	evProdCond := ""
	if product != "" {
		evProdCond = fmt.Sprintf(" AND ev.product = $%d", idx)
		args = append(args, product)
		idx++
	}
	evProjCond := ""
	if projectID != "" {
		evProjCond = fmt.Sprintf(" AND ev.project_id::text = $%d", idx)
		args = append(args, projectID)
		idx++
	}
	searchCond := ""
	if search != "" {
		searchCond = fmt.Sprintf(" AND u.user_id ILIKE $%d", idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	_ = idx

	dataQuery := fmt.Sprintf(`
		WITH flag_users AS (
			SELECT fe.user_id::text AS user_id, MAX(fe.evaluated_at) AS last_seen,
			       COUNT(*) AS flag_cnt, 0::bigint AS ev_cnt
			FROM flag_evaluations fe
			WHERE fe.user_id IS NOT NULL AND fe.evaluated_at >= $1%s%s
			GROUP BY fe.user_id
		),
		event_users AS (
			SELECT ev.user_id::text AS user_id, MAX(ev.created_at) AS last_seen,
			       0::bigint AS flag_cnt, COUNT(*) AS ev_cnt
			FROM events ev
			WHERE ev.user_id IS NOT NULL AND ev.created_at >= $1%s%s
			GROUP BY ev.user_id
		),
		combined AS (
			SELECT user_id, last_seen, flag_cnt, ev_cnt FROM flag_users
			UNION ALL
			SELECT user_id, last_seen, flag_cnt, ev_cnt FROM event_users
		),
		u AS (
			SELECT user_id,
				MAX(last_seen)  AS last_seen_at,
				SUM(flag_cnt)   AS flag_count,
				SUM(ev_cnt)     AS event_count
			FROM combined
			GROUP BY user_id
		)
		SELECT u.user_id, u.event_count::int, u.flag_count::int, u.last_seen_at,
			COALESCE(
				(SELECT ARRAY_AGG(DISTINCT ev2.product)
				 FROM events ev2
				 WHERE ev2.user_id::text = u.user_id AND ev2.created_at >= $1),
				ARRAY[]::text[]
			) AS products
		FROM u
		WHERE 1=1%s
		ORDER BY u.last_seen_at DESC
		LIMIT %d OFFSET %d
	`, flagCond, flagProjCond, evProdCond, evProjCond, searchCond, limit, offset)

	rows, err := h.db.Query(c.Context(), dataQuery, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to list users"})
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.UserID, &u.EventCount, &u.FlagCount, &u.LastSeenAt, &u.Products); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "scan error"})
		}
		if u.Products == nil {
			u.Products = []string{}
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "rows error"})
	}
	if users == nil {
		users = []UserSummary{}
	}

	countQuery := fmt.Sprintf(`
		WITH flag_users AS (
			SELECT fe.user_id::text AS user_id FROM flag_evaluations fe
			WHERE fe.user_id IS NOT NULL AND fe.evaluated_at >= $1%s%s
			GROUP BY fe.user_id
		),
		event_users AS (
			SELECT ev.user_id::text AS user_id FROM events ev
			WHERE ev.user_id IS NOT NULL AND ev.created_at >= $1%s%s
			GROUP BY ev.user_id
		),
		u AS (SELECT user_id FROM flag_users UNION SELECT user_id FROM event_users)
		SELECT COUNT(*) FROM u WHERE 1=1%s
	`, flagCond, flagProjCond, evProdCond, evProjCond, searchCond)

	var total int
	if err := h.db.QueryRow(c.Context(), countQuery, args...).Scan(&total); err != nil {
		total = len(users)
	}

	return c.JSON(fiber.Map{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetTimeline GET /api/v1/users/:user_id/timeline?since=7d&limit=100
func (h *UserInsightsHandler) GetTimeline(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	since := c.Query("since", "30d")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	sinceTime := userSinceParam(since)
	var items []TimelineItem

	// Flag evaluations
	flagRows, err := h.db.Query(c.Context(), `
		SELECT flag_key, result, reason, evaluated_at
		FROM flag_evaluations
		WHERE user_id::text = $1 AND evaluated_at >= $2
		ORDER BY evaluated_at DESC
		LIMIT $3
	`, userID, sinceTime, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to load flags"})
	}
	defer flagRows.Close()

	for flagRows.Next() {
		var fk, reason string
		var result bool
		var ts time.Time
		if err := flagRows.Scan(&fk, &result, &reason, &ts); err != nil {
			continue
		}
		fkCopy := fk
		reasonCopy := reason
		resultCopy := result
		items = append(items, TimelineItem{
			Kind:       "flag",
			Timestamp:  ts,
			FlagKey:    &fkCopy,
			FlagResult: &resultCopy,
			FlagReason: &reasonCopy,
		})
	}
	flagRows.Close()

	// Events
	evRows, err := h.db.Query(c.Context(), `
		SELECT event_type, product, properties, created_at
		FROM events
		WHERE user_id::text = $1 AND created_at >= $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, sinceTime, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to load events"})
	}
	defer evRows.Close()

	for evRows.Next() {
		var evType, prod string
		var propsRaw []byte
		var ts time.Time
		if err := evRows.Scan(&evType, &prod, &propsRaw, &ts); err != nil {
			continue
		}
		var props map[string]interface{}
		if propsRaw != nil {
			_ = json.Unmarshal(propsRaw, &props)
		}
		evTypeCopy := evType
		prodCopy := prod
		items = append(items, TimelineItem{
			Kind:         "event",
			Timestamp:    ts,
			EventType:    &evTypeCopy,
			EventProduct: &prodCopy,
			Properties:   props,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.After(items[j].Timestamp)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	if items == nil {
		items = []TimelineItem{}
	}

	return c.JSON(fiber.Map{"items": items, "user_id": userID})
}

// GetIdentities GET /api/v1/users/:user_id/identities
func (h *UserInsightsHandler) GetIdentities(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	rows, err := h.db.Query(c.Context(), `
		SELECT ui.canonical_user_id::text, ui.product, ui.product_user_id, COALESCE(ui.email, '')
		FROM user_identities ui
		WHERE ui.product_user_id = $1
		   OR ui.canonical_user_id = (
		        SELECT canonical_user_id FROM user_identities
		        WHERE product_user_id = $1 LIMIT 1
		      )
		ORDER BY ui.product ASC
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "failed to load identities"})
	}
	defer rows.Close()

	type Identity struct {
		CanonicalUserID string `json:"canonical_user_id"`
		Product         string `json:"product"`
		ProductUserID   string `json:"product_user_id"`
		Email           string `json:"email,omitempty"`
	}
	var identities []Identity
	for rows.Next() {
		var id Identity
		if err := rows.Scan(&id.CanonicalUserID, &id.Product, &id.ProductUserID, &id.Email); err != nil {
			continue
		}
		identities = append(identities, id)
	}
	if identities == nil {
		identities = []Identity{}
	}
	return c.JSON(fiber.Map{"identities": identities, "user_id": userID})
}

func userSinceParam(s string) time.Time {
	switch s {
	case "24h":
		return time.Now().UTC().Add(-24 * time.Hour)
	case "30d":
		return time.Now().UTC().AddDate(0, 0, -30)
	default: // "7d"
		return time.Now().UTC().AddDate(0, 0, -7)
	}
}
