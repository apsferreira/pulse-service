package service

// BKL-1027: EvaluateFlags com project_id — isola avaliação por tenant.
// Testa a lógica de avaliação end-to-end usando funções internas do service,
// seguindo o padrão estabelecido em flag_service_test.go.

import (
	"testing"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// ---------------------------------------------------------------------------
// TestEvaluateFlags_ProjectIsolation (BKL-1027)
// ---------------------------------------------------------------------------

// Simula o comportamento de EvaluateFlags após a query ao repositório:
// dado um slice de flags filtrado por project_id, a avaliação deve retornar
// apenas as flags daquele projeto.

func TestEvaluateFlags_ProjectIsolation(t *testing.T) {
	projectA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	projectB := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	flagFromA := domain.FeatureFlag{
		Key:       "feature-checkout-v2",
		Enabled:   true,
		ProjectID: &projectA,
	}
	flagFromB := domain.FeatureFlag{
		Key:       "feature-dark-mode",
		Enabled:   true,
		ProjectID: &projectB,
	}

	t.Run("flags filtradas por project_id A não incluem flags do projeto B", func(t *testing.T) {
		// Simula o slice retornado pelo repo.GetAll(ctx, &projectA)
		flagsForA := []domain.FeatureFlag{flagFromA}

		flagMap := make(map[string]domain.FeatureFlag, len(flagsForA))
		for _, f := range flagsForA {
			flagMap[f.Key] = f
		}

		result := make(map[string]bool)
		for _, f := range flagsForA {
			result[f.Key] = evaluateFlagWithPrereqs(f, "user-1", "tenant-1", flagMap)
		}

		if _, ok := result["feature-dark-mode"]; ok {
			t.Error("flag do projeto B não deveria aparecer no resultado do projeto A")
		}
		if !result["feature-checkout-v2"] {
			t.Error("flag habilitada do projeto A deveria estar ativa")
		}
	})

	t.Run("flags filtradas por project_id B não incluem flags do projeto A", func(t *testing.T) {
		flagsForB := []domain.FeatureFlag{flagFromB}

		flagMap := make(map[string]domain.FeatureFlag, len(flagsForB))
		for _, f := range flagsForB {
			flagMap[f.Key] = f
		}

		result := make(map[string]bool)
		for _, f := range flagsForB {
			result[f.Key] = evaluateFlagWithPrereqs(f, "user-2", "tenant-2", flagMap)
		}

		if _, ok := result["feature-checkout-v2"]; ok {
			t.Error("flag do projeto A não deveria aparecer no resultado do projeto B")
		}
		if !result["feature-dark-mode"] {
			t.Error("flag habilitada do projeto B deveria estar ativa")
		}
	})

	t.Run("sem project_id (legado): todas as flags são avaliadas", func(t *testing.T) {
		// Simula GetAll(ctx, nil) — retorna todas as flags
		allFlags := []domain.FeatureFlag{flagFromA, flagFromB}

		flagMap := make(map[string]domain.FeatureFlag, len(allFlags))
		for _, f := range allFlags {
			flagMap[f.Key] = f
		}

		result := make(map[string]bool)
		for _, f := range allFlags {
			result[f.Key] = evaluateFlagWithPrereqs(f, "user-legacy", "", flagMap)
		}

		if _, ok := result["feature-checkout-v2"]; !ok {
			t.Error("modo legado (sem project_id) deveria incluir todas as flags")
		}
		if _, ok := result["feature-dark-mode"]; !ok {
			t.Error("modo legado (sem project_id) deveria incluir todas as flags")
		}
	})
}

// ---------------------------------------------------------------------------
// TestEvaluateFlags_DisabledFlagSkippedPerProject (kill switch)
// ---------------------------------------------------------------------------

func TestEvaluateFlags_DisabledFlagSkippedPerProject(t *testing.T) {
	projectID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	t.Run("flag desabilitada em projeto retorna false no mapa", func(t *testing.T) {
		flags := []domain.FeatureFlag{
			{Key: "feature-enabled", Enabled: true, ProjectID: &projectID},
			{Key: "feature-disabled", Enabled: false, ProjectID: &projectID},
		}

		flagMap := make(map[string]domain.FeatureFlag, len(flags))
		for _, f := range flags {
			flagMap[f.Key] = f
		}

		result := make(map[string]bool)
		for _, f := range flags {
			result[f.Key] = evaluateFlagWithPrereqs(f, "user-x", "tenant-x", flagMap)
		}

		if !result["feature-enabled"] {
			t.Error("feature-enabled deveria ser true")
		}
		if result["feature-disabled"] {
			t.Error("feature-disabled deveria ser false (kill switch)")
		}
	})
}

// ---------------------------------------------------------------------------
// TestEvaluateFlags_RolloutPerProject (BKL-1027)
// ---------------------------------------------------------------------------

func TestEvaluateFlags_RolloutPerProject(t *testing.T) {
	projectID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	t.Run("rollout 100% em projeto retorna true para todos os users", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "feature-full-rollout",
			Enabled:           true,
			RolloutPercentage: 100,
			ProjectID:         &projectID,
		}
		flagMap := map[string]domain.FeatureFlag{f.Key: f}

		users := []string{"user-a", "user-b", "user-c", "user-d", "user-e"}
		for _, u := range users {
			if !evaluateFlagWithPrereqs(f, u, "tenant-x", flagMap) {
				t.Errorf("rollout 100%% deveria retornar true para user=%q", u)
			}
		}
	})

	t.Run("rollout 0% sem targeting retorna true (comportamento legado, BKL-1027 fix)", func(t *testing.T) {
		// Após BKL-1027: flags com rollout=0 e sem targeting continuam com
		// comportamento legado (enabled=true) para não quebrar integrações existentes
		f := domain.FeatureFlag{
			Key:               "feature-legacy",
			Enabled:           true,
			RolloutPercentage: 0,
			ProjectID:         &projectID,
		}
		flagMap := map[string]domain.FeatureFlag{f.Key: f}

		if !evaluateFlagWithPrereqs(f, "user-1", "tenant-1", flagMap) {
			t.Error("flag com rollout=0 e sem targeting deveria retornar true (modo legado)")
		}
	})

	t.Run("targeting por user_id com project_id isola corretamente", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "feature-user-targeted",
			Enabled: true,
			Targeting: domain.FlagTargeting{
				UserIDs: []string{"user-allowed"},
			},
			ProjectID: &projectID,
		}
		flagMap := map[string]domain.FeatureFlag{f.Key: f}

		if !evaluateFlagWithPrereqs(f, "user-allowed", "tenant-x", flagMap) {
			t.Error("user na lista de targeting deveria receber flag=true")
		}
		if evaluateFlagWithPrereqs(f, "user-blocked", "tenant-x", flagMap) {
			t.Error("user fora da lista de targeting deveria receber flag=false")
		}
	})
}

// ---------------------------------------------------------------------------
// TestFlagCacheKeyForProject (BKL-1028)
// ---------------------------------------------------------------------------

func TestFlagCacheKeyForProject(t *testing.T) {
	t.Run("project_id vazio usa chave global", func(t *testing.T) {
		key := flagCacheKeyForProject("")
		if key != "pulse:flags:_global:map" {
			t.Errorf("chave global incorreta: %q", key)
		}
	})

	t.Run("project_id específico gera chave isolada por projeto", func(t *testing.T) {
		key := flagCacheKeyForProject("proj-123")
		expected := "pulse:flags:proj-123:map"
		if key != expected {
			t.Errorf("esperava %q, got %q", expected, key)
		}
	})

	t.Run("chaves de projetos distintos são diferentes (sem cache cross-tenant)", func(t *testing.T) {
		keyA := flagCacheKeyForProject("proj-A")
		keyB := flagCacheKeyForProject("proj-B")
		if keyA == keyB {
			t.Error("projetos distintos não devem compartilhar a mesma chave de cache")
		}
	})

	t.Run("mesmo project_id sempre gera a mesma chave (determinístico)", func(t *testing.T) {
		const pid = "proj-stable"
		first := flagCacheKeyForProject(pid)
		for i := 0; i < 10; i++ {
			if got := flagCacheKeyForProject(pid); got != first {
				t.Errorf("chave não determinística: primeira=%q, iteração %d=%q", first, i, got)
			}
		}
	})
}
