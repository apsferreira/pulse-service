package repository

// BKL-1080: Repositório para event_schemas — schema registry de eventos.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventSchemaRepository struct {
	db *pgxpool.Pool
}

func NewEventSchemaRepository(db *pgxpool.Pool) *EventSchemaRepository {
	return &EventSchemaRepository{db: db}
}

// GetByEventName retorna o schema de um event_type (nil se não existir).
func (r *EventSchemaRepository) GetByEventName(ctx context.Context, projectID, eventName string) (*domain.EventSchema, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, project_id, event_name, COALESCE(description,''), strict, created_at, updated_at
		 FROM event_schemas
		 WHERE project_id = $1::uuid AND event_name = $2`,
		projectID, eventName,
	)
	var s domain.EventSchema
	if err := row.Scan(&s.ID, &s.ProjectID, &s.EventName, &s.Description, &s.Strict, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("event_schema_repository: GetByEventName: %w", err)
	}

	// Carrega propriedades
	props, err := r.listProperties(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.Properties = props
	return &s, nil
}

// List retorna todos os schemas de um projeto.
func (r *EventSchemaRepository) List(ctx context.Context, projectID string) ([]domain.EventSchema, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, event_name, COALESCE(description,''), strict, created_at, updated_at
		 FROM event_schemas
		 WHERE project_id = $1::uuid
		 ORDER BY event_name ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("event_schema_repository: List: %w", err)
	}
	defer rows.Close()

	var schemas []domain.EventSchema
	for rows.Next() {
		var s domain.EventSchema
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.EventName, &s.Description, &s.Strict, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("event_schema_repository: List scan: %w", err)
		}
		props, err := r.listProperties(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.Properties = props
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

// Create cria um novo schema com suas propriedades.
func (r *EventSchemaRepository) Create(ctx context.Context, projectID string, req domain.CreateEventSchemaRequest) (*domain.EventSchema, error) {
	s := domain.EventSchema{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		EventName:   req.EventName,
		Description: req.Description,
		Strict:      req.Strict,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("event_schema_repository: Create begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO event_schemas (id, project_id, event_name, description, strict, created_at, updated_at)
		 VALUES ($1, $2::uuid, $3, $4, $5, $6, $7)
		 ON CONFLICT (project_id, event_name) DO UPDATE
		 SET description = EXCLUDED.description, strict = EXCLUDED.strict, updated_at = EXCLUDED.updated_at`,
		s.ID, s.ProjectID, s.EventName, s.Description, s.Strict, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("event_schema_repository: Create insert schema: %w", err)
	}

	// Remove propriedades antigas e recria (upsert simples)
	if _, err := tx.Exec(ctx, `DELETE FROM event_schema_properties WHERE schema_id = $1::uuid`, s.ID); err != nil {
		return nil, fmt.Errorf("event_schema_repository: Create delete props: %w", err)
	}

	for _, p := range req.Properties {
		propID := uuid.New().String()
		if _, err := tx.Exec(ctx,
			`INSERT INTO event_schema_properties (id, schema_id, name, type, required, description)
			 VALUES ($1, $2::uuid, $3, $4, $5, $6)`,
			propID, s.ID, p.Name, string(p.Type), p.Required, p.Description,
		); err != nil {
			return nil, fmt.Errorf("event_schema_repository: Create insert prop: %w", err)
		}
		s.Properties = append(s.Properties, domain.EventSchemaProperty{
			ID: propID, SchemaID: s.ID, Name: p.Name, Type: p.Type,
			Required: p.Required, Description: p.Description,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("event_schema_repository: Create commit: %w", err)
	}
	return &s, nil
}

// Delete remove um schema e suas propriedades.
func (r *EventSchemaRepository) Delete(ctx context.Context, projectID, eventName string) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM event_schemas WHERE project_id = $1::uuid AND event_name = $2`,
		projectID, eventName,
	)
	if err != nil {
		return fmt.Errorf("event_schema_repository: Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("event_schema_repository: Delete: schema não encontrado")
	}
	return nil
}

func (r *EventSchemaRepository) listProperties(ctx context.Context, schemaID string) ([]domain.EventSchemaProperty, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, schema_id, name, type, required, COALESCE(description,'')
		 FROM event_schema_properties
		 WHERE schema_id = $1::uuid
		 ORDER BY name ASC`,
		schemaID,
	)
	if err != nil {
		return nil, fmt.Errorf("event_schema_repository: listProperties: %w", err)
	}
	defer rows.Close()

	var props []domain.EventSchemaProperty
	for rows.Next() {
		var p domain.EventSchemaProperty
		if err := rows.Scan(&p.ID, &p.SchemaID, &p.Name, &p.Type, &p.Required, &p.Description); err != nil {
			return nil, fmt.Errorf("event_schema_repository: listProperties scan: %w", err)
		}
		props = append(props, p)
	}
	return props, rows.Err()
}
