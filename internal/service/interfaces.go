//go:generate mockgen -source=interfaces.go -destination=mocks/mock_interfaces.go -package=mocks

package service

import (
	"context"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// FlagServicer define as operações do serviço de feature flags.
// Permite substituição por mocks nos testes de handler.
type FlagServicer interface {
	List(ctx context.Context) ([]domain.FeatureFlag, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.FeatureFlag, error) // BKL-323
	GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error)
	Create(ctx context.Context, req domain.CreateFlagRequest) (*domain.FeatureFlag, error)
	Update(ctx context.Context, key string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error)
	Delete(ctx context.Context, key string) error
	GetFlagsMap(ctx context.Context, projectID string) (domain.FlagsMapResponse, error)          // BKL-1028
	EvaluateFlags(ctx context.Context, userID, tenantID, projectID string) (domain.FlagsMapResponse, error) // BKL-1027
	GetVariants(ctx context.Context, userID, tenantID, projectID string) (domain.VariantResponse, error)    // BKL-1027
}

// EventServicer define as operações do serviço de eventos.
type EventServicer interface {
	Ingest(ctx context.Context, req domain.IngestEventRequest, ipAddress, userAgent string) (*domain.Event, error)
	BatchIngest(ctx context.Context, req domain.BatchIngestRequest, ipAddress, userAgent string) (domain.BatchIngestResponse, error) // BKL-1041
	GetSummary(ctx context.Context, projectID string) (domain.EventsSummaryResponse, error)
	Track(ctx context.Context, req domain.TrackEventRequest, ipAddress, userAgent string) (*domain.Event, error)
	Identify(ctx context.Context, req domain.IdentifyRequest) (int, error)
	Count(ctx context.Context, eventName string, startDate, endDate *time.Time) (domain.EventCountResponse, error)
	List(ctx context.Context, eventName, userID string, limit, offset int) (domain.EventListResponse, error)
}

// Verificação em tempo de compilação: as structs concretas satisfazem as interfaces.
var _ FlagServicer = (*FlagService)(nil)
var _ EventServicer = (*EventService)(nil)
