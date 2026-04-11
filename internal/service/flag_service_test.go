package service

import (
	"testing"

	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// ---------------------------------------------------------------------------
// TestHashBucket
// ---------------------------------------------------------------------------

func TestHashBucket(t *testing.T) {
	t.Run("resultado sempre entre 0 e 99", func(t *testing.T) {
		inputs := []struct{ user, flag string }{
			{"user-abc", "flag-x"},
			{"user-xyz", "flag-y"},
			{"", "flag-empty-user"},
			{"user-1", ""},
			{"", ""},
		}
		for _, tc := range inputs {
			got := hashBucket(tc.user, tc.flag)
			if got < 0 || got > 99 {
				t.Errorf("hashBucket(%q, %q) = %d; fora do intervalo [0,99]", tc.user, tc.flag, got)
			}
		}
	})

	t.Run("mesmo input retorna mesmo output (determinístico)", func(t *testing.T) {
		const userID = "user-deterministic"
		const flagKey = "flag-deterministic"
		first := hashBucket(userID, flagKey)
		for i := 0; i < 10; i++ {
			if got := hashBucket(userID, flagKey); got != first {
				t.Errorf("hashBucket não determinístico: primeira=%d, iteração %d=%d", first, i, got)
			}
		}
	})

	t.Run("inputs diferentes produzem resultados potencialmente diferentes", func(t *testing.T) {
		// Testa um conjunto grande de pares para verificar que não há colisão total
		seen := make(map[int]int)
		for i := 0; i < 50; i++ {
			userID := "user-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
			bucket := hashBucket(userID, "flag-variation-test")
			seen[bucket]++
		}
		// Com 50 inputs e 100 buckets, espera-se pelo menos 2 buckets distintos
		if len(seen) < 2 {
			t.Errorf("hashBucket produziu apenas %d valor(es) distinto(s) para 50 inputs diferentes", len(seen))
		}
	})
}

// ---------------------------------------------------------------------------
// TestSelectVariant
// ---------------------------------------------------------------------------

func TestSelectVariant(t *testing.T) {
	t.Run("lista vazia retorna string vazia", func(t *testing.T) {
		got := selectVariant(nil, "user-1", "flag-a")
		if got != "" {
			t.Errorf("esperava string vazia, got %q", got)
		}
		got = selectVariant([]domain.FlagVariant{}, "user-1", "flag-a")
		if got != "" {
			t.Errorf("esperava string vazia para slice vazio, got %q", got)
		}
	})

	t.Run("pesos totais zero retorna primeira variante", func(t *testing.T) {
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 0},
			{Key: "challenger", Weight: 0},
		}
		got := selectVariant(variants, "user-1", "flag-b")
		if got != "control" {
			t.Errorf("esperava 'control' quando pesos=0, got %q", got)
		}
	})

	t.Run("2 variantes 50/50 sempre retorna uma das duas", func(t *testing.T) {
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 50},
			{Key: "challenger", Weight: 50},
		}
		valid := map[string]bool{"control": true, "challenger": true}
		for _, userID := range []string{"user-a", "user-b", "user-c", "user-d", "user-e"} {
			got := selectVariant(variants, userID, "flag-ab")
			if !valid[got] {
				t.Errorf("selectVariant retornou %q, esperava 'control' ou 'challenger'", got)
			}
		}
	})

	t.Run("mesmo user+flag sempre retorna mesma variante (determinístico)", func(t *testing.T) {
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 50},
			{Key: "challenger", Weight: 50},
		}
		const userID = "user-stable"
		const flagKey = "flag-stable"
		first := selectVariant(variants, userID, flagKey)
		for i := 0; i < 10; i++ {
			if got := selectVariant(variants, userID, flagKey); got != first {
				t.Errorf("selectVariant não determinístico: primeira=%q, iteração %d=%q", first, i, got)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestEvaluateFlag
// ---------------------------------------------------------------------------

func TestEvaluateFlag(t *testing.T) {
	t.Run("flag disabled retorna false (kill switch)", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "flag-disabled",
			Enabled: false,
		}
		if got := evaluateFlag(f, "user-1", "tenant-1"); got {
			t.Error("esperava false para flag desabilitada")
		}
	})

	t.Run("flag enabled sem targeting sem rollout retorna true (comportamento legado)", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "flag-legacy",
			Enabled:           true,
			RolloutPercentage: 0,
		}
		if got := evaluateFlag(f, "user-1", "tenant-1"); !got {
			t.Error("esperava true para flag enabled sem targeting e sem rollout")
		}
	})

	t.Run("targeting por user_id: usuário na lista retorna true", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "flag-user-targeting",
			Enabled: true,
			Targeting: domain.FlagTargeting{
				UserIDs: []string{"user-allowed", "user-other"},
			},
		}
		if got := evaluateFlag(f, "user-allowed", "tenant-x"); !got {
			t.Error("esperava true para usuário na lista de targeting")
		}
	})

	t.Run("targeting por user_id: usuário fora da lista retorna false", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "flag-user-targeting",
			Enabled: true,
			Targeting: domain.FlagTargeting{
				UserIDs: []string{"user-allowed"},
			},
		}
		if got := evaluateFlag(f, "user-blocked", "tenant-x"); got {
			t.Error("esperava false para usuário fora da lista de targeting")
		}
	})

	t.Run("targeting por tenant_id: tenant na lista retorna true", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "flag-tenant-targeting",
			Enabled: true,
			Targeting: domain.FlagTargeting{
				TenantIDs: []string{"tenant-gold", "tenant-silver"},
			},
		}
		if got := evaluateFlag(f, "user-1", "tenant-gold"); !got {
			t.Error("esperava true para tenant na lista de targeting")
		}
	})

	t.Run("targeting por tenant_id: tenant fora da lista retorna false", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:     "flag-tenant-targeting",
			Enabled: true,
			Targeting: domain.FlagTargeting{
				TenantIDs: []string{"tenant-gold"},
			},
		}
		if got := evaluateFlag(f, "user-1", "tenant-bronze"); got {
			t.Error("esperava false para tenant fora da lista de targeting")
		}
	})

	t.Run("rollout 100% retorna true para qualquer userID", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "flag-full-rollout",
			Enabled:           true,
			RolloutPercentage: 100,
		}
		for _, userID := range []string{"user-a", "user-b", "user-c", "user-z"} {
			if got := evaluateFlag(f, userID, ""); !got {
				t.Errorf("esperava true para rollout 100%%, userID=%q", userID)
			}
		}
	})

	t.Run("rollout 0% sem targeting retorna true (comportamento legado)", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "flag-zero-rollout-no-targeting",
			Enabled:           true,
			RolloutPercentage: 0,
		}
		if got := evaluateFlag(f, "user-1", "tenant-1"); !got {
			t.Error("esperava true para rollout=0 e sem targeting (modo legado)")
		}
	})

	t.Run("rollout 50%: mesmo userID sempre retorna mesmo resultado (determinístico)", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "flag-half-rollout",
			Enabled:           true,
			RolloutPercentage: 50,
		}
		const userID = "user-consistent-rollout"
		first := evaluateFlag(f, userID, "")
		for i := 0; i < 10; i++ {
			if got := evaluateFlag(f, userID, ""); got != first {
				t.Errorf("evaluateFlag não determinístico no rollout 50%%: primeira=%v, iteração %d=%v", first, i, got)
			}
		}
	})

	t.Run("rollout 50%: distribui entre users (não todos true ou todos false)", func(t *testing.T) {
		f := domain.FeatureFlag{
			Key:               "flag-distribution-rollout",
			Enabled:           true,
			RolloutPercentage: 50,
		}
		trueCount, falseCount := 0, 0
		users := []string{
			"user-01", "user-02", "user-03", "user-04", "user-05",
			"user-06", "user-07", "user-08", "user-09", "user-10",
			"user-11", "user-12", "user-13", "user-14", "user-15",
			"user-16", "user-17", "user-18", "user-19", "user-20",
		}
		for _, u := range users {
			if evaluateFlag(f, u, "") {
				trueCount++
			} else {
				falseCount++
			}
		}
		if trueCount == 0 || falseCount == 0 {
			t.Errorf("rollout 50%% deveria distribuir entre true/false; true=%d, false=%d", trueCount, falseCount)
		}
	})
}

// ---------------------------------------------------------------------------
// TestEvaluateFlagWithPrereqs
// ---------------------------------------------------------------------------

func TestEvaluateFlagWithPrereqs(t *testing.T) {
	// Flag principal habilitada, sem targeting/rollout
	mainFlag := domain.FeatureFlag{
		Key:     "flag-main",
		Enabled: true,
	}

	t.Run("sem prerequisites passa para evaluateFlag normalmente", func(t *testing.T) {
		f := mainFlag
		f.Prerequisites = nil
		allFlags := map[string]domain.FeatureFlag{"flag-main": f}

		got := evaluateFlagWithPrereqs(f, "user-1", "tenant-1", allFlags)
		if !got {
			t.Error("esperava true para flag habilitada sem prerequisites")
		}
	})

	t.Run("prerequisito inexistente retorna false", func(t *testing.T) {
		f := mainFlag
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-does-not-exist"},
		}
		allFlags := map[string]domain.FeatureFlag{"flag-main": f}

		got := evaluateFlagWithPrereqs(f, "user-1", "tenant-1", allFlags)
		if got {
			t.Error("esperava false quando prerequisito não existe no mapa")
		}
	})

	t.Run("prerequisito disabled retorna false", func(t *testing.T) {
		f := mainFlag
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-prereq"},
		}
		prereq := domain.FeatureFlag{
			Key:     "flag-prereq",
			Enabled: false,
		}
		allFlags := map[string]domain.FeatureFlag{
			"flag-main":  f,
			"flag-prereq": prereq,
		}

		got := evaluateFlagWithPrereqs(f, "user-1", "tenant-1", allFlags)
		if got {
			t.Error("esperava false quando prerequisito está desabilitado")
		}
	})

	t.Run("prerequisito enabled sem variante obrigatória avalia flag principal", func(t *testing.T) {
		f := mainFlag
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-prereq", RequiredVariant: ""},
		}
		prereq := domain.FeatureFlag{
			Key:     "flag-prereq",
			Enabled: true,
		}
		allFlags := map[string]domain.FeatureFlag{
			"flag-main":  f,
			"flag-prereq": prereq,
		}

		// flag-main está enabled sem targeting → deve retornar true
		got := evaluateFlagWithPrereqs(f, "user-1", "tenant-1", allFlags)
		if !got {
			t.Error("esperava true quando prerequisito está habilitado e sem variante obrigatória")
		}
	})

	t.Run("prerequisito com variante obrigatória não satisfeita retorna false", func(t *testing.T) {
		// Usa variantes com pesos que forçam um resultado previsível pelo hash.
		// Vamos encontrar um userID cujo bucket de variante seleciona "control",
		// então declaramos RequiredVariant = "challenger" — deve retornar false.
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 100},
			{Key: "challenger", Weight: 0},
		}
		prereq := domain.FeatureFlag{
			Key:      "flag-prereq-variant",
			Enabled:  true,
			Variants: variants,
		}
		f := mainFlag
		f.Key = "flag-needs-variant"
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-prereq-variant", RequiredVariant: "challenger"},
		}
		allFlags := map[string]domain.FeatureFlag{
			"flag-needs-variant":   f,
			"flag-prereq-variant":  prereq,
		}

		// Com weight 100/0, selectVariant sempre retorna "control".
		// RequiredVariant = "challenger" → deve bloquear.
		got := evaluateFlagWithPrereqs(f, "user-any", "tenant-any", allFlags)
		if got {
			t.Error("esperava false quando variante do prerequisito não satisfaz RequiredVariant")
		}
	})

	t.Run("prerequisito com variante obrigatória satisfeita avalia flag principal", func(t *testing.T) {
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 0},
			{Key: "challenger", Weight: 100},
		}
		prereq := domain.FeatureFlag{
			Key:      "flag-prereq-variant",
			Enabled:  true,
			Variants: variants,
		}
		f := mainFlag
		f.Key = "flag-needs-variant"
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-prereq-variant", RequiredVariant: "challenger"},
		}
		allFlags := map[string]domain.FeatureFlag{
			"flag-needs-variant":  f,
			"flag-prereq-variant": prereq,
		}

		// Com weight 0/100, selectVariant sempre retorna "challenger".
		// RequiredVariant = "challenger" → prerequisito satisfeito → avalia flag-main (true).
		got := evaluateFlagWithPrereqs(f, "user-any", "tenant-any", allFlags)
		if !got {
			t.Error("esperava true quando variante do prerequisito satisfaz RequiredVariant e flag principal está habilitada")
		}
	})

	t.Run("múltiplos prerequisites: um falho bloqueia avaliação", func(t *testing.T) {
		prereqOK := domain.FeatureFlag{Key: "flag-prereq-ok", Enabled: true}
		prereqKO := domain.FeatureFlag{Key: "flag-prereq-ko", Enabled: false}
		f := mainFlag
		f.Prerequisites = []domain.FlagPrerequisite{
			{FlagKey: "flag-prereq-ok"},
			{FlagKey: "flag-prereq-ko"},
		}
		allFlags := map[string]domain.FeatureFlag{
			"flag-main":      f,
			"flag-prereq-ok": prereqOK,
			"flag-prereq-ko": prereqKO,
		}

		got := evaluateFlagWithPrereqs(f, "user-1", "tenant-1", allFlags)
		if got {
			t.Error("esperava false quando ao menos um prerequisito está desabilitado")
		}
	})
}
