package handler

// BKL-1036: LGPD Art. 18 §5° — portabilidade de dados por user_id.
// GET /api/v1/users/:user_id/export?project_id=
// Retorna todos os eventos do usuário em JSON. Para volumes grandes, resposta
// é gerada inline (sem async job — volumes típicos são manejáveis).

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserExportHandler struct {
	db           *pgxpool.Pool
	serviceToken string
}

func NewUserExportHandler(db *pgxpool.Pool, serviceToken string) *UserExportHandler {
	return &UserExportHandler{db: db, serviceToken: serviceToken}
}

func (h *UserExportHandler) RegisterRoutes(app *fiber.App) {
	auth := middleware.ServiceTokenMiddleware(h.serviceToken)
	app.Get("/api/v1/users/:user_id/export", auth, h.Export)
	app.Delete("/api/v1/users/:user_id/data", auth, h.DeleteUserData)
}

// Export godoc
// GET /api/v1/users/:user_id/export?project_id=
// Retorna todos os eventos do usuário em JSON (portabilidade LGPD Art. 18 §5°).
func (h *UserExportHandler) Export(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "user_id obrigatório"})
	}
	projectID := c.Query("project_id")

	// Busca eventos do usuário
	events, err := h.fetchUserEvents(c.UserContext(), userID, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	// Busca avaliações de flags do usuário
	flagEvals, err := h.fetchUserFlagEvals(c.UserContext(), userID, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	export := fiber.Map{
		"user_id":         userID,
		"exported_at":     time.Now().UTC().Format(time.RFC3339),
		"legal_basis":     "LGPD Art. 18 §5° — portabilidade de dados",
		"events":          events,
		"flag_evaluations": flagEvals,
		"total_events":    len(events),
		"total_flag_evals": len(flagEvals),
	}

	// Retorna como download JSON
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao serializar dados"})
	}

	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="pulse-export-%s.json"`, userID))
	return c.Send(data)
}

// DeleteUserData godoc
// DELETE /api/v1/users/:user_id/data?project_id=
// Remove todos os dados pessoais do usuário (LGPD Art. 18 VI — exclusão).
func (h *UserExportHandler) DeleteUserData(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "user_id obrigatório"})
	}
	projectID := c.Query("project_id")

	deletedEvents, err := h.deleteUserEvents(c.UserContext(), userID, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	deletedFlagEvals, err := h.deleteUserFlagEvals(c.UserContext(), userID, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{
		"user_id":               userID,
		"deleted_events":        deletedEvents,
		"deleted_flag_evals":    deletedFlagEvals,
		"deleted_at":            time.Now().UTC().Format(time.RFC3339),
		"legal_basis":           "LGPD Art. 18 VI — exclusão de dados",
	})
}

// --- helpers de banco ---

func (h *UserExportHandler) fetchUserEvents(ctx context.Context, userID, projectID string) ([]map[string]interface{}, error) {
	query := `SELECT id, product, event_type, session_id, properties, environment, project_id, created_at
		FROM events WHERE user_id = $1`
	args := []interface{}{userID}
	if projectID != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
	}
	query += " ORDER BY created_at DESC LIMIT 10000"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("export: fetchEvents: %w", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var (
			id, product, eventType string
			sessionID, projID      *string
			propsRaw               []byte
			environment            string
			createdAt              time.Time
		)
		if err := rows.Scan(&id, &product, &eventType, &sessionID, &propsRaw, &environment, &projID, &createdAt); err != nil {
			return nil, err
		}
		var props map[string]interface{}
		_ = json.Unmarshal(propsRaw, &props)

		events = append(events, map[string]interface{}{
			"id":          id,
			"product":     product,
			"event_type":  eventType,
			"session_id":  sessionID,
			"properties":  props,
			"environment": environment,
			"project_id":  projID,
			"created_at":  createdAt.Format(time.RFC3339),
		})
	}
	return events, rows.Err()
}

func (h *UserExportHandler) fetchUserFlagEvals(ctx context.Context, userID, projectID string) ([]map[string]interface{}, error) {
	query := `SELECT id, flag_key, project_id, result, variant, reason, evaluated_at
		FROM flag_evaluations WHERE user_id = $1`
	args := []interface{}{userID}
	if projectID != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
	}
	query += " ORDER BY evaluated_at DESC LIMIT 5000"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("export: fetchFlagEvals: %w", err)
	}
	defer rows.Close()

	var evals []map[string]interface{}
	for rows.Next() {
		var (
			id, flagKey string
			projID      *string
			result      bool
			variant     *string
			reason      string
			evaluatedAt time.Time
		)
		if err := rows.Scan(&id, &flagKey, &projID, &result, &variant, &reason, &evaluatedAt); err != nil {
			return nil, err
		}
		evals = append(evals, map[string]interface{}{
			"id":           id,
			"flag_key":     flagKey,
			"project_id":   projID,
			"result":       result,
			"variant":      variant,
			"reason":       reason,
			"evaluated_at": evaluatedAt.Format(time.RFC3339),
		})
	}
	return evals, rows.Err()
}

func (h *UserExportHandler) deleteUserEvents(ctx context.Context, userID, projectID string) (int64, error) {
	query := "DELETE FROM events WHERE user_id = $1"
	args := []interface{}{userID}
	if projectID != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
	}
	tag, err := h.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("deleteUserEvents: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (h *UserExportHandler) deleteUserFlagEvals(ctx context.Context, userID, projectID string) (int64, error) {
	query := "DELETE FROM flag_evaluations WHERE user_id = $1"
	args := []interface{}{userID}
	if projectID != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
	}
	tag, err := h.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("deleteUserFlagEvals: %w", err)
	}
	return tag.RowsAffected(), nil
}
