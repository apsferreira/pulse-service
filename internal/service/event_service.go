package service

import (
	"context"
	"fmt"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
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

	event := &domain.Event{
		Product:    req.Product,
		EventType:  req.EventType,
		UserID:     req.UserID,
		SessionID:  req.SessionID,
		Properties: props,
	}
	if ipAddress != "" {
		event.IPAddress = &ipAddress
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
func (s *EventService) GetSummary(ctx context.Context) (domain.EventsSummaryResponse, error) {
	items, err := s.repo.GetSummaryLast24h(ctx)
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

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}

	product := req.Product
	if product == "" {
		product = "default"
	}

	event := &domain.Event{
		Product:    product,
		EventType:  req.EventName,
		UserID:     req.UserID,
		Properties: props,
	}
	if ipAddress != "" {
		event.IPAddress = &ipAddress
	}
	if userAgent != "" {
		event.UserAgent = &userAgent
	}

	if err := s.repo.Insert(ctx, event); err != nil {
		return nil, fmt.Errorf("event_service: Track: %w", err)
	}
	return event, nil
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
