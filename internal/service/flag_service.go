package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	flagCacheTTL = 5 * time.Minute
	flagCacheKey = "pulse:flags:map"
)

type FlagService struct {
	repo  *repository.FlagRepository
	redis *redis.Client
}

func NewFlagService(repo *repository.FlagRepository, redis *redis.Client) *FlagService {
	return &FlagService{repo: repo, redis: redis}
}

// List returns all feature flags.
func (s *FlagService) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	flags, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: List: %w", err)
	}
	return flags, nil
}

// GetByKey returns a single flag by key.
func (s *FlagService) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	flag, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("service: GetByKey: %w", err)
	}
	return flag, nil
}

// Create validates uniqueness and creates a new feature flag.
func (s *FlagService) Create(ctx context.Context, req domain.CreateFlagRequest) (*domain.FeatureFlag, error) {
	existing, err := s.repo.GetByKey(ctx, req.Key)
	if err != nil {
		return nil, fmt.Errorf("service: Create check: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("flag with key '%s' already exists", req.Key)
	}

	flag := &domain.FeatureFlag{
		Key:         req.Key,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	if err := s.repo.Create(ctx, flag); err != nil {
		return nil, fmt.Errorf("service: Create: %w", err)
	}

	s.invalidateCache(ctx)
	return flag, nil
}

// Update applies partial updates to an existing flag by key.
func (s *FlagService) Update(ctx context.Context, key string, req domain.UpdateFlagRequest) (*domain.FeatureFlag, error) {
	updated, err := s.repo.Update(ctx, key, req)
	if err != nil {
		return nil, fmt.Errorf("service: Update: %w", err)
	}
	if updated == nil {
		return nil, nil
	}

	s.invalidateCache(ctx)
	return updated, nil
}

// Delete removes a feature flag by key.
func (s *FlagService) Delete(ctx context.Context, key string) error {
	if err := s.repo.Delete(ctx, key); err != nil {
		return fmt.Errorf("service: Delete: %w", err)
	}
	s.invalidateCache(ctx)
	return nil
}

// GetFlagsMap returns a map of key:enabled for all flags.
// Results are cached in Redis (DB 6) with TTL 5min.
func (s *FlagService) GetFlagsMap(ctx context.Context) (domain.FlagsMapResponse, error) {
	// Try cache first
	cached, err := s.redis.Get(ctx, flagCacheKey).Bytes()
	if err == nil {
		var resp domain.FlagsMapResponse
		if jsonErr := json.Unmarshal(cached, &resp); jsonErr == nil {
			return resp, nil
		}
	}

	// Cache miss — load from DB
	flags, err := s.repo.GetAll(ctx)
	if err != nil {
		return domain.FlagsMapResponse{}, fmt.Errorf("service: GetFlagsMap: %w", err)
	}

	result := make(map[string]bool, len(flags))
	for _, f := range flags {
		result[f.Key] = f.Enabled
	}

	resp := domain.FlagsMapResponse{Flags: result}

	// Cache the result
	if data, jsonErr := json.Marshal(resp); jsonErr == nil {
		if setErr := s.redis.Set(ctx, flagCacheKey, data, flagCacheTTL).Err(); setErr != nil {
			log.Printf("warning: failed to cache flags map: %v", setErr)
		}
	}

	return resp, nil
}

// invalidateCache deletes the flags map cache key.
func (s *FlagService) invalidateCache(ctx context.Context) {
	if err := s.redis.Del(ctx, flagCacheKey).Err(); err != nil {
		log.Printf("warning: failed to invalidate flags cache: %v", err)
	}
}
