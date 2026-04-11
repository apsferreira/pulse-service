package service

import (
	"context"
	"log"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// HealthScoreService calcula e persiste health scores por tenant.
// BKL-801: score 0-100 com 4 dimensões.
type HealthScoreService struct {
	repo *repository.HealthScoreRepository
}

func NewHealthScoreService(repo *repository.HealthScoreRepository) *HealthScoreService {
	return &HealthScoreService{repo: repo}
}

// GetByTenantID retorna o health score atual de um tenant.
// Retorna score zero se ainda não calculado.
func (s *HealthScoreService) GetByTenantID(ctx context.Context, tenantID string) (*domain.TenantHealthScore, error) {
	hs, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if hs == nil {
		// Tenant sem score ainda — retorna zero score
		return &domain.TenantHealthScore{
			TenantID:     tenantID,
			Score:        0,
			CalculatedAt: time.Now().UTC(),
		}, nil
	}
	return hs, nil
}

// ListAll retorna todos os health scores (BKL-1083: visão admin interna).
func (s *HealthScoreService) ListAll(ctx context.Context) ([]*domain.TenantHealthScore, error) {
	return s.repo.ListAll(ctx)
}

// CalculateAll recalcula o health score de todos os tenants ativos.
// Deve ser chamado pelo cron diário.
func (s *HealthScoreService) CalculateAll(ctx context.Context) {
	tenants, err := s.repo.ListDistinctTenants(ctx)
	if err != nil {
		log.Printf("[health-score] falha ao listar tenants: %v", err)
		return
	}

	log.Printf("[health-score] calculando score para %d tenants", len(tenants))

	for _, tenantID := range tenants {
		if err := s.calculateOne(ctx, tenantID); err != nil {
			log.Printf("[health-score] erro ao calcular tenant=%s: %v", tenantID, err)
		}
	}
}

// calculateOne calcula e persiste o health score de um tenant.
func (s *HealthScoreService) calculateOne(ctx context.Context, tenantID string) error {
	// --- Sinal 1: Login frequency (30 pontos máximo) ---
	// Eventos: user.login, session.started
	loginCount, err := s.repo.CountEventsLast30Days(ctx, tenantID, "user.login", "session.started")
	if err != nil {
		return err
	}
	lastLoginAt, _ := s.repo.LastEventAt(ctx, tenantID, "user.login", "session.started")

	loginScore := scoreFromCount(loginCount, 30, map[int]int{
		1:  10, // pelo menos 1 login = 10pts
		4:  20, // 4+ logins = 20pts
		10: 30, // 10+ logins = 30pts (máximo)
	})

	// --- Sinal 2: Content creation (40 pontos máximo) ---
	// Eventos: post.published, post.created, post.scheduled
	contentCount, err := s.repo.CountEventsLast30Days(ctx, tenantID, "post.published", "post.created", "post.scheduled")
	if err != nil {
		return err
	}
	lastContentAt, _ := s.repo.LastEventAt(ctx, tenantID, "post.published", "post.created", "post.scheduled")

	contentScore := scoreFromCount(contentCount, 40, map[int]int{
		1:  15, // 1 post = 15pts
		4:  30, // 4+ posts = 30pts
		10: 40, // 10+ posts = 40pts (máximo)
	})

	// --- Sinal 3: Feature adoption (20 pontos máximo) ---
	// Eventos: connection.oauth_success, ai.suggestion_used, calendar.viewed, approval.requested
	adoptionCount, err := s.repo.CountEventsLast30Days(ctx, tenantID,
		"connection.oauth_success", "ai.suggestion_used", "calendar.viewed", "approval.requested",
		"analytics.viewed", "post.submitted_for_approval")
	if err != nil {
		return err
	}

	adoptionScore := scoreFromCount(adoptionCount, 20, map[int]int{
		1: 10, // qualquer feature usada = 10pts
		3: 15, // 3+ usos = 15pts
		6: 20, // 6+ usos = 20pts (máximo)
	})

	// --- Sinal 4: Subscription (10 pontos máximo) ---
	// Eventos: subscription_created, payment_confirmed
	subCount, err := s.repo.CountEventsLast30Days(ctx, tenantID, "subscription_created", "payment_confirmed")
	if err != nil {
		return err
	}

	subscriptionScore := 0
	if subCount > 0 {
		subscriptionScore = 10 // tem assinatura ativa = 10pts
	}

	totalScore := loginScore + contentScore + adoptionScore + subscriptionScore

	hs := &domain.TenantHealthScore{
		TenantID:          tenantID,
		Score:             totalScore,
		LoginScore:        loginScore,
		ContentScore:      contentScore,
		AdoptionScore:     adoptionScore,
		SubscriptionScore: subscriptionScore,
		LastLoginAt:       lastLoginAt,
		LastContentAt:     lastContentAt,
	}

	return s.repo.Upsert(ctx, hs)
}

// scoreFromCount mapeia uma contagem de eventos para uma pontuação.
// thresholds: mapa de contagem mínima → pontos atribuídos. Usa o maior threshold satisfeito.
func scoreFromCount(count, max int, thresholds map[int]int) int {
	if count <= 0 {
		return 0
	}
	score := 0
	for threshold, pts := range thresholds {
		if count >= threshold && pts > score {
			score = pts
		}
	}
	if score > max {
		return max
	}
	return score
}

// StartDailyJob inicia o job de cálculo diário de health scores.
func StartHealthScoreJob(ctx context.Context, svc *HealthScoreService) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		// Executa imediatamente na inicialização
		svc.CalculateAll(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.CalculateAll(ctx)
			}
		}
	}()
	log.Println("[health-score] job diário iniciado (intervalo: 24h)")
}
