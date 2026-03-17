package service

import (
	"context"
	"fmt"

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
