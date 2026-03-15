package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

const flagCacheTTL = 60 * time.Second

type FlagService struct {
	repo  *repository.FlagRepository
	redis *redis.Client
}

func NewFlagService(repo *repository.FlagRepository, redis *redis.Client) *FlagService {
	return &FlagService{repo: repo, redis: redis}
}

// List returns all feature flags, optionally filtered by product.
func (s *FlagService) List(ctx context.Context, product *string) ([]domain.FeatureFlag, error) {
	flags, err := s.repo.GetAll(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("service: List: %w", err)
	}
	return flags, nil
}

// Create validates uniqueness and creates a new feature flag.
func (s *FlagService) Create(ctx context.Context, req domain.CreateFlagRequest) (*domain.FeatureFlag, error) {
	product := ""
	if req.Product != nil {
		product = *req.Product
	}

	existing, err := s.repo.GetByName(ctx, req.Name, product)
	if err != nil {
		return nil, fmt.Errorf("service: Create check: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("flag '%s' already exists for product '%s'", req.Name, product)
	}

	rolloutPct := 100
	if req.RolloutPct != nil {
		rolloutPct = *req.RolloutPct
	}

	flag := &domain.FeatureFlag{
		Name:        req.Name,
		Description: req.Description,
		IsEnabled:   req.IsEnabled,
		Product:     req.Product,
		RolloutPct:  rolloutPct,
	}

	if err := s.repo.Create(ctx, flag); err != nil {
		return nil, fmt.Errorf("service: Create: %w", err)
	}

	// Invalidate cache for the product
	s.InvalidateCache(ctx, product)

	return flag, nil
}

// Update applies partial updates to an existing flag.
func (s *FlagService) Update(ctx context.Context, id string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("service: Update: %w", err)
	}
	if updated == nil {
		return nil, nil
	}

	// Invalidate cache for all products since we don't know the product without extra query
	s.InvalidateCache(ctx, "")

	return updated, nil
}

// Delete removes a feature flag by ID.
func (s *FlagService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service: Delete: %w", err)
	}
	s.InvalidateCache(ctx, "")
	return nil
}

// EvaluateFlags resolves the effective flag values for a given user/tenant/product combination.
// Resolution order: user override > tenant override > rollout % > default enabled value.
// Results are cached in Redis for 60 seconds.
func (s *FlagService) EvaluateFlags(ctx context.Context, product, userID, tenantID string) (domain.FlagsResponse, error) {
	cacheKey := fmt.Sprintf("pulse:flags:%s:%s", product, userID)

	// Try cache first
	cached, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var resp domain.FlagsResponse
		if jsonErr := json.Unmarshal(cached, &resp); jsonErr == nil {
			return resp, nil
		}
	}

	// Load flags for this product (including global flags)
	var productPtr *string
	if product != "" {
		productPtr = &product
	}
	flags, err := s.repo.GetAll(ctx, productPtr)
	if err != nil {
		return domain.FlagsResponse{}, fmt.Errorf("service: EvaluateFlags load flags: %w", err)
	}

	// Load overrides for this user/tenant
	overrides, err := s.repo.GetOverridesForUser(ctx, userID, tenantID)
	if err != nil {
		return domain.FlagsResponse{}, fmt.Errorf("service: EvaluateFlags load overrides: %w", err)
	}

	// Build override index keyed by flagID (user override takes priority over tenant)
	overrideByFlagID := make(map[string]bool)
	// Process tenant overrides first, then user overrides (user wins)
	for _, o := range overrides {
		if o.TenantID != nil && tenantID != "" && *o.TenantID == tenantID {
			overrideByFlagID[o.FlagID] = o.IsEnabled
		}
	}
	for _, o := range overrides {
		if o.UserID != nil && userID != "" && *o.UserID == userID {
			overrideByFlagID[o.FlagID] = o.IsEnabled
		}
	}

	result := make(map[string]bool, len(flags))
	for _, f := range flags {
		if enabled, hasOverride := overrideByFlagID[f.ID]; hasOverride {
			result[f.Name] = enabled
			continue
		}

		if !f.IsEnabled {
			result[f.Name] = false
			continue
		}

		// Apply rollout percentage using consistent hash of userID
		if f.RolloutPct < 100 && userID != "" {
			hash := crc32.ChecksumIEEE([]byte(f.Name + ":" + userID))
			bucket := int(hash % 100)
			result[f.Name] = bucket < f.RolloutPct
		} else {
			result[f.Name] = f.IsEnabled
		}
	}

	resp := domain.FlagsResponse{Flags: result}

	// Cache the result
	if data, err := json.Marshal(resp); err == nil {
		s.redis.Set(ctx, cacheKey, data, flagCacheTTL)
	}

	return resp, nil
}

// InvalidateCache deletes Redis keys matching pulse:flags:{product}:*.
// If product is empty, invalidates all flag cache keys.
func (s *FlagService) InvalidateCache(ctx context.Context, product string) {
	pattern := "pulse:flags:*"
	if product != "" {
		pattern = fmt.Sprintf("pulse:flags:%s:*", product)
	}

	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redis.Del(ctx, iter.Val())
	}
}
