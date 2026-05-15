// Package agent implementa o agente autônomo de análise de feature flags e eventos.
// BKL-insights: roda periodicamente (default 6h), analisa dados por projeto ativo
// e gera insights acionáveis na tabela project_insights.
package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// InsightsAgent analisa feature flags e eventos e persiste insights acionáveis.
type InsightsAgent struct {
	repo     *repository.ProjectInsightRepository
	interval time.Duration
}

// NewInsightsAgent cria o agente. intervalHours ≤ 0 usa o default de 6h.
func NewInsightsAgent(repo *repository.ProjectInsightRepository, intervalHours int) *InsightsAgent {
	if intervalHours <= 0 {
		intervalHours = 6
	}
	return &InsightsAgent{
		repo:     repo,
		interval: time.Duration(intervalHours) * time.Hour,
	}
}

// Start inicia o loop do agente. Executa imediatamente ao startup e depois a cada interval.
// Bloqueante — chamar em goroutine separada.
func (a *InsightsAgent) Start(ctx context.Context) {
	log.Printf("[insights-agent] iniciado — intervalo: %s", a.interval)

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	// Primeira execução ao startup
	a.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[insights-agent] parado")
			return
		case <-ticker.C:
			a.runOnce(ctx)
		}
	}
}

// RunNow executa uma análise completa imediatamente (usado pelo endpoint de trigger manual).
func (a *InsightsAgent) RunNow(ctx context.Context) error {
	return a.runOnce(ctx)
}

// runOnce executa a análise para todos os projetos ativos.
func (a *InsightsAgent) runOnce(ctx context.Context) error {
	// Limpa insights expirados antes de gerar novos
	purged, err := a.repo.PurgeExpired(ctx)
	if err != nil {
		log.Printf("[insights-agent] erro ao purgar expirados: %v", err)
	} else if purged > 0 {
		log.Printf("[insights-agent] insights expirados removidos: %d", purged)
	}

	projects, err := a.repo.ActiveProjects(ctx)
	if err != nil {
		log.Printf("[insights-agent] erro ao buscar projetos: %v", err)
		return err
	}

	total := 0
	for _, proj := range projects {
		n, err := a.analyzeProject(ctx, proj.ID, proj.Slug)
		if err != nil {
			log.Printf("[insights-agent] erro projeto=%s: %v", proj.Slug, err)
			continue
		}
		total += n
	}

	if total > 0 {
		log.Printf("[insights-agent] ciclo concluído — %d insights gerados em %d projetos",
			total, len(projects))
	}
	return nil
}

// analyzeProject roda todas as análises para um projeto e retorna a contagem de novos insights.
func (a *InsightsAgent) analyzeProject(ctx context.Context, projectID, slug string) (int, error) {
	n := 0

	analyses := []func(context.Context, string, string) (int, error){
		a.analyzeDormantFlags,
		a.analyzeFullRollout,
		a.analyzeEventVolume,
		a.analyzeNegativeRate,
		a.analyzePartialRollout,
	}

	for _, fn := range analyses {
		count, err := fn(ctx, projectID, slug)
		if err != nil {
			log.Printf("[insights-agent] projeto=%s análise com erro: %v", slug, err)
			continue
		}
		n += count
	}

	return n, nil
}

// analyzeDormantFlags — análise A: flags ativas sem avaliações nos últimos 30 dias.
func (a *InsightsAgent) analyzeDormantFlags(ctx context.Context, projectID, slug string) (int, error) {
	flags, err := a.repo.DormantFlags(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("dormant_flags: %w", err)
	}

	n := 0
	for _, f := range flags {
		expires := expiresAt7Days()
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeDormantFlag,
			FlagKey:     &f.Key,
			Message:     fmt.Sprintf("A flag '%s' está ativa há 30+ dias sem avaliações. Considere desativá-la.", f.Key),
			Severity:    domain.InsightSeverityWarning,
			Metadata:    map[string]any{"flag_key": f.Key},
			ExpiresAt:   &expires,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s dormant_flag insert erro: %v", slug, err)
			continue
		}
		n++
	}
	return n, nil
}

// analyzeFullRollout — análise B: flags em 100% de rollout por mais de 14 dias.
func (a *InsightsAgent) analyzeFullRollout(ctx context.Context, projectID, slug string) (int, error) {
	flags, err := a.repo.FullRolloutFlags(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("full_rollout: %w", err)
	}

	n := 0
	for _, f := range flags {
		expires := expiresAt7Days()
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeFullRollout,
			FlagKey:     &f.Key,
			Message:     fmt.Sprintf("A flag '%s' está em 100%% de rollout há %d dias. Considere removê-la do código e deletar a flag.", f.Key, f.Days),
			Severity:    domain.InsightSeverityWarning,
			Metadata:    map[string]any{"flag_key": f.Key, "days_active": f.Days},
			ExpiresAt:   &expires,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s full_rollout insert erro: %v", slug, err)
			continue
		}
		n++
	}
	return n, nil
}

// analyzeEventVolume — análises C e D: spike (>3x média) e drop (<20% da média).
func (a *InsightsAgent) analyzeEventVolume(ctx context.Context, projectID, slug string) (int, error) {
	last24h, avg7d, err := a.repo.EventVolumeStats(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("event_volume: %w", err)
	}

	n := 0
	expires := expiresAt7Days()

	// Spike: volume > 3x a média diária
	if avg7d > 0 && float64(last24h) > 3*avg7d {
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeSpike,
			Message: fmt.Sprintf(
				"Spike detectado: %d eventos nas últimas 24h (média: %.0f/dia). Verifique se é uso legítimo ou erro no SDK.",
				last24h, avg7d,
			),
			Severity: domain.InsightSeverityCritical,
			Metadata: map[string]any{
				"last_24h": last24h,
				"avg_7d":   avg7d,
			},
			ExpiresAt: &expires,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s spike insert erro: %v", slug, err)
		} else {
			n++
		}
	}

	// Drop: volume < 20% da média (somente se a média for relevante — >100/dia)
	if avg7d > 100 && float64(last24h) < 0.2*avg7d {
		expires2 := expiresAt7Days()
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeDrop,
			Message: fmt.Sprintf(
				"Drop detectado: apenas %d eventos nas últimas 24h (média: %.0f/dia). O SDK pode estar com problema.",
				last24h, avg7d,
			),
			Severity: domain.InsightSeverityCritical,
			Metadata: map[string]any{
				"last_24h": last24h,
				"avg_7d":   avg7d,
			},
			ExpiresAt: &expires2,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s drop insert erro: %v", slug, err)
		} else {
			n++
		}
	}

	return n, nil
}

// analyzeNegativeRate — análise E: flags avaliadas como false em >90% dos casos.
func (a *InsightsAgent) analyzeNegativeRate(ctx context.Context, projectID, slug string) (int, error) {
	flags, err := a.repo.NegativeRateFlags(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("negative_rate: %w", err)
	}

	n := 0
	for _, f := range flags {
		expires := expiresAt7Days()
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeNegativeRate,
			FlagKey:     &f.Key,
			Message: fmt.Sprintf(
				"A flag '%s' é avaliada como false em %.1f%% dos casos. Se for intencional, considere simplificar o targeting.",
				f.Key, f.Pct,
			),
			Severity: domain.InsightSeverityInfo,
			Metadata: map[string]any{"flag_key": f.Key, "negative_pct": f.Pct},
			ExpiresAt: &expires,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s negative_rate insert erro: %v", slug, err)
			continue
		}
		n++
	}
	return n, nil
}

// analyzePartialRollout — análise F: flags com rollout 10-90% sem experimento A/B.
func (a *InsightsAgent) analyzePartialRollout(ctx context.Context, projectID, slug string) (int, error) {
	flags, err := a.repo.PartialRolloutFlags(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("partial_rollout: %w", err)
	}

	n := 0
	for _, f := range flags {
		expires := expiresAt7Days()
		ins := &domain.ProjectInsight{
			ProjectID:   projectID,
			InsightType: domain.InsightTypeABSuggestion,
			FlagKey:     &f.Key,
			Message: fmt.Sprintf(
				"A flag '%s' está em rollout parcial (%.0f%%). Configure um A/B test para medir o impacto com dados estatísticos.",
				f.Key, f.Pct,
			),
			Severity: domain.InsightSeverityInfo,
			Metadata: map[string]any{"flag_key": f.Key, "rollout_pct": f.Pct},
			ExpiresAt: &expires,
		}
		if err := a.repo.Insert(ctx, ins); err != nil {
			log.Printf("[insights-agent] projeto=%s ab_suggestion insert erro: %v", slug, err)
			continue
		}
		n++
	}
	return n, nil
}

// expiresAt7Days retorna NOW() + 7 dias em UTC.
func expiresAt7Days() time.Time {
	return time.Now().UTC().Add(7 * 24 * time.Hour)
}
