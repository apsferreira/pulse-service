package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/pkg/privacy"
	"github.com/institutoitinerante/pulse-service/internal/pkg/products"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// Ingest validates and persists a new event.
func (s *EventService) Ingest(ctx context.Context, req domain.IngestEventRequest, ipAddress, userAgent string) (*domain.Event, error) {
	// fix(#64): normaliza product (lowercase + trim) e alerta se fora da allowlist
	req.Product = products.NormalizeAndCheck(req.Product)
	if req.Product == "" {
		return nil, fmt.Errorf("product is required")
	}
	if req.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}

	env := req.Environment
	if env == "" {
		env = "production"
	}

	event := &domain.Event{
		Product:     req.Product,
		EventType:   req.EventType,
		UserID:      req.UserID,
		SessionID:   req.SessionID,
		Properties:  props,
		Environment: env,           // BKL-750
		ProjectID:   req.ProjectID, // BKL-1038
	}
	if anon := privacy.AnonymizeIP(ipAddress); anon != "" {
		event.IPAddress = &anon
	}
	if userAgent != "" {
		event.UserAgent = &userAgent
	}

	if err := s.repo.Insert(ctx, event); err != nil {
		return nil, fmt.Errorf("event_service: Ingest: %w", err)
	}
	return event, nil
}

// GetSummary returns event counts grouped by product/event_type for last 24h.
// projectID is optional — empty string returns all projects.
func (s *EventService) GetSummary(ctx context.Context, projectID string) (domain.EventsSummaryResponse, error) {
	items, err := s.repo.GetSummaryLast24h(ctx, projectID)
	if err != nil {
		return domain.EventsSummaryResponse{}, fmt.Errorf("event_service: GetSummary: %w", err)
	}
	if items == nil {
		items = []domain.EventSummaryItem{}
	}

	total := 0
	for _, item := range items {
		total += item.Count
	}

	return domain.EventsSummaryResponse{
		Items:        items,
		TotalLast24h: total,
	}, nil
}

// Track validates and persists a new event via the simplified track endpoint.
func (s *EventService) Track(ctx context.Context, req domain.TrackEventRequest, ipAddress, userAgent string) (*domain.Event, error) {
	if req.EventName == "" {
		return nil, fmt.Errorf("event_name is required")
	}

	// ProjectID (BKL-1038) é opcional, mas quando enviado precisa ser um UUID
	// válido — é usado tanto para isolar o canal de broadcast quanto como
	// project_id::uuid em POST /api/v1/projects/:id/query. Um valor malformado
	// aceito aqui só quebraria silenciosamente na hora da consulta.
	if req.ProjectID != nil && *req.ProjectID != "" {
		if _, err := uuid.Parse(*req.ProjectID); err != nil {
			return nil, fmt.Errorf("project_id inválido: deve ser um UUID")
		}
	}

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}

	// fix(#64): normaliza product (lowercase + trim) e alerta se fora da allowlist
	product := products.NormalizeAndCheck(req.Product)
	if product == "" {
		product = "default"
	}

	// Usa user_id se disponivel, senao anonymous_id como user_id temporario
	userID := req.UserID
	if (userID == nil || *userID == "") && req.AnonymousID != nil && *req.AnonymousID != "" {
		anonPrefixed := "anon:" + *req.AnonymousID
		userID = &anonPrefixed
	}

	event := &domain.Event{
		Product:    product,
		EventType:  req.EventName,
		UserID:     userID,
		Properties: props,
		ProjectID:  req.ProjectID,
	}
	if anon := privacy.AnonymizeIP(ipAddress); anon != "" {
		event.IPAddress = &anon
	}
	if userAgent != "" {
		event.UserAgent = &userAgent
	}

	if err := s.repo.Insert(ctx, event); err != nil {
		return nil, fmt.Errorf("event_service: Track: %w", err)
	}
	return event, nil
}

// BatchIngest valida e persiste até 500 eventos em lote.
// BKL-1041: usa InsertBatch (pgx CopyFrom) para alta performance.
func (s *EventService) BatchIngest(ctx context.Context, req domain.BatchIngestRequest, ipAddress, userAgent string) (domain.BatchIngestResponse, error) {
	if len(req.Events) == 0 {
		return domain.BatchIngestResponse{}, fmt.Errorf("events array cannot be empty")
	}
	if len(req.Events) > 500 {
		return domain.BatchIngestResponse{}, fmt.Errorf("batch too large: max 500 events per request, got %d", len(req.Events))
	}

	events := make([]*domain.Event, 0, len(req.Events))
	rejected := 0

	for _, r := range req.Events {
		// fix(#64): normaliza product (lowercase + trim) e alerta se fora da allowlist
		r.Product = products.NormalizeAndCheck(r.Product)
		if r.Product == "" || r.EventType == "" {
			rejected++
			continue
		}

		props := r.Properties
		if props == nil {
			props = make(map[string]interface{})
		}
		env := r.Environment
		if env == "" {
			env = "production"
		}

		e := &domain.Event{
			Product:     r.Product,
			EventType:   r.EventType,
			UserID:      r.UserID,
			SessionID:   r.SessionID,
			Properties:  props,
			Environment: env,
			ProjectID:   r.ProjectID,
		}
		if anon := privacy.AnonymizeIP(ipAddress); anon != "" {
			e.IPAddress = &anon
		}
		if userAgent != "" {
			e.UserAgent = &userAgent
		}
		events = append(events, e)
	}

	n, err := s.repo.InsertBatch(ctx, events)
	if err != nil {
		return domain.BatchIngestResponse{}, fmt.Errorf("event_service: BatchIngest: %w", err)
	}

	return domain.BatchIngestResponse{
		Accepted: n,
		Rejected: rejected,
	}, nil
}

// Identify vincula um anonymous_id a um user_id real.
// Atualiza todos os eventos do anonymous_id para o user_id.
func (s *EventService) Identify(ctx context.Context, req domain.IdentifyRequest) (int, error) {
	if req.AnonymousID == "" || req.UserID == "" {
		return 0, fmt.Errorf("anonymous_id e user_id sao obrigatorios")
	}
	anonKey := "anon:" + req.AnonymousID
	count, err := s.repo.MergeAnonymousUser(ctx, anonKey, req.UserID)
	if err != nil {
		return 0, fmt.Errorf("event_service: Identify: %w", err)
	}
	return count, nil
}

// Count returns event counts grouped by event_name for a given period.
func (s *EventService) Count(ctx context.Context, eventName string, startDate, endDate *time.Time) (domain.EventCountResponse, error) {
	items, err := s.repo.CountByEventName(ctx, eventName, startDate, endDate)
	if err != nil {
		return domain.EventCountResponse{}, fmt.Errorf("event_service: Count: %w", err)
	}
	if items == nil {
		items = []domain.EventCountItem{}
	}

	total := 0
	for _, item := range items {
		total += item.Count
	}

	return domain.EventCountResponse{
		Items: items,
		Total: total,
	}, nil
}

// List returns events with optional filters and pagination.
func (s *EventService) List(ctx context.Context, eventName, userID string, limit, offset int) (domain.EventListResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	events, total, err := s.repo.ListEvents(ctx, eventName, userID, limit, offset)
	if err != nil {
		return domain.EventListResponse{}, fmt.Errorf("event_service: List: %w", err)
	}
	if events == nil {
		events = []domain.Event{}
	}

	return domain.EventListResponse{
		Items:  events,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
