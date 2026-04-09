package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
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

// List returns all feature flags (sem filtro de projeto).
func (s *FlagService) List(ctx context.Context) ([]domain.FeatureFlag, error) {
	flags, err := s.repo.GetAll(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("service: List: %w", err)
	}
	return flags, nil
}

// ListByProject retorna flags filtradas por project_id. BKL-323.
func (s *FlagService) ListByProject(ctx context.Context, projectID string) ([]domain.FeatureFlag, error) {
	flags, err := s.repo.GetAll(ctx, &projectID)
	if err != nil {
		return nil, fmt.Errorf("service: ListByProject: %w", err)
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
		Key:               req.Key,
		Description:       req.Description,
		Enabled:           req.Enabled,
		RolloutPercentage: req.RolloutPercentage,
		Targeting:         req.Targeting,
		Variants:          req.Variants, // BKL-094
		ProjectID:         req.ProjectID, // BKL-323
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
// Quando sem contexto (user_id/tenant_id vazio), retorna o campo enabled diretamente.
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
	flags, err := s.repo.GetAll(ctx, nil)
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

// EvaluateFlags avalia flags com contexto de user_id e tenant_id.
// Lógica de avaliação (em ordem de prioridade):
// 1. Se flag.Enabled = false → false (kill switch)
// 2. Se targeting tem user_ids e user_id está na lista → true
// 3. Se targeting tem tenant_ids e tenant_id está na lista → true
// 4. Se rollout_percentage > 0 → hash(user_id + flag_key) % 100 < rollout_percentage
// 5. Se rollout_percentage = 0 e targeting vazio → enabled (comportamento legado)
func (s *FlagService) EvaluateFlags(ctx context.Context, userID, tenantID string) (domain.FlagsMapResponse, error) {
	flags, err := s.repo.GetAll(ctx, nil)
	if err != nil {
		return domain.FlagsMapResponse{}, fmt.Errorf("service: EvaluateFlags: %w", err)
	}

	result := make(map[string]bool, len(flags))
	for _, f := range flags {
		result[f.Key] = evaluateFlag(f, userID, tenantID)
	}

	return domain.FlagsMapResponse{Flags: result}, nil
}

// evaluateFlag determina se uma flag está ativa para um dado contexto.
func evaluateFlag(f domain.FeatureFlag, userID, tenantID string) bool {
	// Kill switch: se desabilitada, sempre false
	if !f.Enabled {
		return false
	}

	hasTargeting := len(f.Targeting.UserIDs) > 0 || len(f.Targeting.TenantIDs) > 0
	hasRollout := f.RolloutPercentage > 0

	// Sem rollout e sem targeting → comportamento legado (enabled = true)
	if !hasTargeting && !hasRollout {
		return true
	}

	// Targeting por user_id
	if userID != "" && len(f.Targeting.UserIDs) > 0 {
		for _, id := range f.Targeting.UserIDs {
			if id == userID {
				return true
			}
		}
	}

	// Targeting por tenant_id
	if tenantID != "" && len(f.Targeting.TenantIDs) > 0 {
		for _, id := range f.Targeting.TenantIDs {
			if id == tenantID {
				return true
			}
		}
	}

	// Rollout percentual hash-based (determinístico por user)
	if hasRollout && userID != "" {
		bucket := hashBucket(userID, f.Key)
		return bucket < f.RolloutPercentage
	}

	// Se tem targeting mas o contexto não bateu → false
	if hasTargeting {
		return false
	}

	return f.Enabled
}

// hashBucket retorna um número 0-99 deterministico para um dado user+flag.
// Mesmo user+flag sempre retorna o mesmo bucket → experiência consistente.
func hashBucket(userID, flagKey string) int {
	h := sha256.Sum256([]byte(userID + ":" + flagKey))
	num := binary.BigEndian.Uint32(h[:4])
	return int(num % 100)
}

// GetVariants retorna a variante A/B selecionada para cada flag com variantes configuradas. BKL-094.
// A seleção é determinística: mesmo userID+flagKey sempre retorna a mesma variante.
// Flags sem variantes ou desabilitadas não aparecem na resposta.
func (s *FlagService) GetVariants(ctx context.Context, userID, tenantID string) (domain.VariantResponse, error) {
	flags, err := s.repo.GetAll(ctx, nil)
	if err != nil {
		return domain.VariantResponse{}, fmt.Errorf("service: GetVariants: %w", err)
	}

	variants := make(map[string]string)
	for _, f := range flags {
		if !f.Enabled || len(f.Variants) == 0 {
			continue
		}
		// Só inclui se o usuário está no rollout (ou sem rollout configurado)
		if !evaluateFlag(f, userID, tenantID) {
			continue
		}
		variant := selectVariant(f.Variants, userID, f.Key)
		if variant != "" {
			variants[f.Key] = variant
		}
	}

	return domain.VariantResponse{Variants: variants}, nil
}

// selectVariant escolhe deterministicamente uma variante com base em pesos cumulativos.
// Usa hashBucket para consistência: mesmo userID+flagKey → mesma variante.
func selectVariant(variants []domain.FlagVariant, userID, flagKey string) string {
	if len(variants) == 0 {
		return ""
	}

	// Calcula total de pesos
	total := 0
	for _, v := range variants {
		total += v.Weight
	}
	if total == 0 {
		return variants[0].Key
	}

	// Bucket 0-99 → mapeia para peso cumulativo
	bucket := hashBucket(userID, flagKey+":variant")
	// Normaliza para escala 0-total
	scaled := (bucket * total) / 100

	cumulative := 0
	for _, v := range variants {
		cumulative += v.Weight
		if scaled < cumulative {
			return v.Key
		}
	}
	return variants[len(variants)-1].Key
}

// invalidateCache deletes the flags map cache key.
func (s *FlagService) invalidateCache(ctx context.Context) {
	if err := s.redis.Del(ctx, flagCacheKey).Err(); err != nil {
		log.Printf("warning: failed to invalidate flags cache: %v", err)
	}
}
