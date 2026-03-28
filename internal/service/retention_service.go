package service

import (
	"context"
	"fmt"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type RetentionService struct {
	repo *repository.RetentionRepository
}

func NewRetentionService(repo *repository.RetentionRepository) *RetentionService {
	return &RetentionService{repo: repo}
}

var validProducts = map[string]bool{
	"libri": true,
	"nitro": true,
	"brio":  true,
}

func validateProduct(product string) error {
	if product != "" && !validProducts[product] {
		return fmt.Errorf("invalid product: must be one of libri, nitro, brio")
	}
	return nil
}

// GetRetention returns DAU/WAU/MAU per date for a product.
func (s *RetentionService) GetRetention(ctx context.Context, product string, startDate, endDate time.Time) (domain.RetentionResponse, error) {
	if err := validateProduct(product); err != nil {
		return domain.RetentionResponse{}, err
	}

	if endDate.Before(startDate) {
		return domain.RetentionResponse{}, fmt.Errorf("end_date must be after start_date")
	}

	// Cap range to 90 days
	maxRange := startDate.AddDate(0, 0, 90)
	if endDate.After(maxRange) {
		endDate = maxRange
	}

	items, err := s.repo.GetRetention(ctx, product, startDate, endDate)
	if err != nil {
		return domain.RetentionResponse{}, fmt.Errorf("retention_service: GetRetention: %w", err)
	}
	if items == nil {
		items = []domain.ActiveUsersItem{}
	}

	return domain.RetentionResponse{
		Product: product,
		Items:   items,
	}, nil
}

// GetCohorts returns cohort retention analysis.
func (s *RetentionService) GetCohorts(ctx context.Context, product, granularity string) (domain.CohortsResponse, error) {
	if err := validateProduct(product); err != nil {
		return domain.CohortsResponse{}, err
	}

	if granularity != "week" && granularity != "month" {
		granularity = "week"
	}

	cohorts, err := s.repo.GetCohorts(ctx, product, granularity)
	if err != nil {
		return domain.CohortsResponse{}, fmt.Errorf("retention_service: GetCohorts: %w", err)
	}
	if cohorts == nil {
		cohorts = []domain.CohortItem{}
	}

	return domain.CohortsResponse{
		Product:     product,
		Granularity: granularity,
		Cohorts:     cohorts,
	}, nil
}

// GetActiveUsers returns DAU/WAU/MAU breakdowns.
func (s *RetentionService) GetActiveUsers(ctx context.Context, product string, startDate, endDate time.Time) (domain.ActiveUsersResponse, error) {
	if err := validateProduct(product); err != nil {
		return domain.ActiveUsersResponse{}, err
	}

	if endDate.Before(startDate) {
		return domain.ActiveUsersResponse{}, fmt.Errorf("end_date must be after start_date")
	}

	// Cap range to 90 days
	maxRange := startDate.AddDate(0, 0, 90)
	if endDate.After(maxRange) {
		endDate = maxRange
	}

	dau, wau, mau, err := s.repo.GetActiveUsers(ctx, product, startDate, endDate)
	if err != nil {
		return domain.ActiveUsersResponse{}, fmt.Errorf("retention_service: GetActiveUsers: %w", err)
	}
	if dau == nil {
		dau = []domain.ActiveUsersCountItem{}
	}
	if wau == nil {
		wau = []domain.ActiveUsersCountItem{}
	}
	if mau == nil {
		mau = []domain.ActiveUsersCountItem{}
	}

	return domain.ActiveUsersResponse{
		Product: product,
		DAU:     dau,
		WAU:     wau,
		MAU:     mau,
	}, nil
}
