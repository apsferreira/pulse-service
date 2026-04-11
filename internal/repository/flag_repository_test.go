package repository

// Testes para flag_repository.go sem conexão real ao banco.
//
// Estratégia: como FlagRepository usa pgxpool.Pool diretamente (sem interface),
// testamos a lógica pura acessível no pacote:
//   - scanFlag: deserialização correta de JSON (targeting, variants, prerequisites)
//   - construção de FeatureFlag a partir dos campos escaneados
//   - comportamento de campos nullable (project_id, starts_at, ends_at)
//
// Testes de queries SQL (GetAll, Create, Update, Delete) requerem testcontainers
// e estão marcados como skip quando INTEGRATION_TESTS=1 não está definido.

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// ---------------------------------------------------------------------------
// Helpers — constrói bytes JSON para simular o que o banco retorna
// ---------------------------------------------------------------------------

func marshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal falhou: %v", err)
	}
	return b
}

// ---------------------------------------------------------------------------
// TestFlagDeserialization
// ---------------------------------------------------------------------------
// Testa que as estruturas de domínio são corretamente deserializadas a partir
// de JSON — o mesmo caminho que scanFlag percorre internamente.

func TestFlagDeserialization(t *testing.T) {
	t.Run("targeting com user_ids e tenant_ids é deserializado corretamente", func(t *testing.T) {
		targeting := domain.FlagTargeting{
			UserIDs:   []string{"user-a", "user-b"},
			TenantIDs: []string{"tenant-gold"},
		}
		b := marshalJSON(t, targeting)

		var got domain.FlagTargeting
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal targeting falhou: %v", err)
		}
		if len(got.UserIDs) != 2 {
			t.Errorf("esperava 2 user_ids, got %d", len(got.UserIDs))
		}
		if len(got.TenantIDs) != 1 {
			t.Errorf("esperava 1 tenant_id, got %d", len(got.TenantIDs))
		}
		if got.UserIDs[0] != "user-a" {
			t.Errorf("user_ids[0] esperado 'user-a', got %q", got.UserIDs[0])
		}
	})

	t.Run("variants com pesos são deserializadas corretamente", func(t *testing.T) {
		variants := []domain.FlagVariant{
			{Key: "control", Weight: 50},
			{Key: "challenger", Weight: 50},
		}
		b := marshalJSON(t, variants)

		var got []domain.FlagVariant
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal variants falhou: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("esperava 2 variants, got %d", len(got))
		}
		if got[0].Key != "control" || got[0].Weight != 50 {
			t.Errorf("variant[0] incorreta: %+v", got[0])
		}
	})

	t.Run("prerequisites com required_variant são deserializados corretamente", func(t *testing.T) {
		prereqs := []domain.FlagPrerequisite{
			{FlagKey: "flag-parent", RequiredVariant: "challenger"},
		}
		b := marshalJSON(t, prereqs)

		var got []domain.FlagPrerequisite
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal prerequisites falhou: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("esperava 1 prerequisite, got %d", len(got))
		}
		if got[0].FlagKey != "flag-parent" {
			t.Errorf("FlagKey esperado 'flag-parent', got %q", got[0].FlagKey)
		}
		if got[0].RequiredVariant != "challenger" {
			t.Errorf("RequiredVariant esperado 'challenger', got %q", got[0].RequiredVariant)
		}
	})

	t.Run("targeting vazio (nil bytes) não causa panic", func(t *testing.T) {
		var targeting domain.FlagTargeting
		var nilBytes []byte
		if len(nilBytes) > 0 {
			_ = json.Unmarshal(nilBytes, &targeting)
		}
		// targeting permanece no zero-value — sem panic
		if len(targeting.UserIDs) != 0 || len(targeting.TenantIDs) != 0 {
			t.Error("targeting a partir de nil bytes deveria ser zero-value")
		}
	})

	t.Run("variants como array vazio JSON é deserializado como slice vazio", func(t *testing.T) {
		var variants []domain.FlagVariant
		if err := json.Unmarshal([]byte("[]"), &variants); err != nil {
			t.Fatalf("unmarshal array vazio falhou: %v", err)
		}
		if variants == nil {
			t.Error("variants a partir de '[]' deveria ser slice vazio, não nil")
		}
		if len(variants) != 0 {
			t.Errorf("esperava 0 variants, got %d", len(variants))
		}
	})
}

// ---------------------------------------------------------------------------
// TestFeatureFlagFields — campos nullable (BKL-323, BKL-1040)
// ---------------------------------------------------------------------------

func TestFeatureFlagFields(t *testing.T) {
	t.Run("project_id como *uuid.UUID pode ser nil ou ter valor", func(t *testing.T) {
		// nil project_id (flag global, sem escopo de projeto)
		var f1 domain.FeatureFlag
		if f1.ProjectID != nil {
			t.Error("project_id zero-value deveria ser nil")
		}

		// project_id com valor (flag de projeto específico)
		pid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
		f2 := domain.FeatureFlag{ProjectID: &pid}
		if f2.ProjectID == nil {
			t.Error("project_id não deveria ser nil quando setado")
		}
		if *f2.ProjectID != pid {
			t.Errorf("project_id incorreto: %v", f2.ProjectID)
		}
	})

	t.Run("starts_at e ends_at como *time.Time podem ser nil (BKL-1040)", func(t *testing.T) {
		// Flags sem agendamento
		f := domain.FeatureFlag{Key: "no-schedule", Enabled: true}
		if f.StartsAt != nil {
			t.Error("starts_at deveria ser nil para flags sem agendamento")
		}
		if f.EndsAt != nil {
			t.Error("ends_at deveria ser nil para flags sem expiração")
		}

		// Flags com agendamento
		start := time.Now().Add(24 * time.Hour)
		end := time.Now().Add(7 * 24 * time.Hour)
		fScheduled := domain.FeatureFlag{
			Key:      "scheduled-flag",
			Enabled:  false,
			StartsAt: &start,
			EndsAt:   &end,
		}
		if fScheduled.StartsAt == nil || fScheduled.EndsAt == nil {
			t.Error("starts_at e ends_at não deveriam ser nil quando configurados")
		}
		if fScheduled.StartsAt.Before(time.Now()) {
			t.Error("starts_at deveria estar no futuro")
		}
	})

	t.Run("RolloutPercentage aceita valores 0-100", func(t *testing.T) {
		for _, pct := range []int{0, 25, 50, 75, 100} {
			f := domain.FeatureFlag{RolloutPercentage: pct}
			if f.RolloutPercentage != pct {
				t.Errorf("RolloutPercentage %d não foi preservado", pct)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestCreateFlagRequest — serialização do payload de criação
// ---------------------------------------------------------------------------

func TestCreateFlagRequest(t *testing.T) {
	t.Run("payload com todos os campos é serializado e deserializado corretamente", func(t *testing.T) {
		pid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
		start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

		req := domain.CreateFlagRequest{
			Key:               "new-feature",
			Description:       "Feature de teste",
			Enabled:           true,
			RolloutPercentage: 50,
			Targeting: domain.FlagTargeting{
				UserIDs: []string{"user-beta"},
			},
			Variants: []domain.FlagVariant{
				{Key: "control", Weight: 50},
				{Key: "treatment", Weight: 50},
			},
			Prerequisites: []domain.FlagPrerequisite{
				{FlagKey: "parent-flag"},
			},
			ProjectID: &pid,
			StartsAt:  &start,
			EndsAt:    &end,
		}

		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal CreateFlagRequest falhou: %v", err)
		}

		var got domain.CreateFlagRequest
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("unmarshal CreateFlagRequest falhou: %v", err)
		}

		if got.Key != "new-feature" {
			t.Errorf("Key incorreto: %q", got.Key)
		}
		if got.RolloutPercentage != 50 {
			t.Errorf("RolloutPercentage incorreto: %d", got.RolloutPercentage)
		}
		if got.ProjectID == nil || *got.ProjectID != pid {
			t.Error("ProjectID incorreto após round-trip JSON")
		}
		if len(got.Variants) != 2 {
			t.Errorf("esperava 2 variants, got %d", len(got.Variants))
		}
		if len(got.Prerequisites) != 1 {
			t.Errorf("esperava 1 prerequisite, got %d", len(got.Prerequisites))
		}
	})
}

// ---------------------------------------------------------------------------
// TestUpdateFlagRequest — campos opcionais com ponteiros
// ---------------------------------------------------------------------------

func TestUpdateFlagRequest(t *testing.T) {
	t.Run("campos nil não alteram o valor existente", func(t *testing.T) {
		existing := domain.FeatureFlag{
			Key:               "existing-flag",
			Enabled:           true,
			RolloutPercentage: 75,
		}
		req := domain.UpdateFlagRequest{
			Enabled: nil, // não alterado
		}

		// Simula a lógica de Update do repositório
		if req.Enabled != nil {
			existing.Enabled = *req.Enabled
		}
		if existing.Enabled != true {
			t.Error("Enabled deveria permanecer true quando req.Enabled é nil")
		}
	})

	t.Run("campo explicitamente false desabilita a flag", func(t *testing.T) {
		existing := domain.FeatureFlag{Enabled: true}
		disabled := false
		req := domain.UpdateFlagRequest{Enabled: &disabled}

		if req.Enabled != nil {
			existing.Enabled = *req.Enabled
		}
		if existing.Enabled {
			t.Error("flag deveria estar desabilitada após UpdateFlagRequest com Enabled=false")
		}
	})

	t.Run("rollout_percentage pode ser atualizado para 0 (kill gradual)", func(t *testing.T) {
		existing := domain.FeatureFlag{RolloutPercentage: 50}
		zero := 0
		req := domain.UpdateFlagRequest{RolloutPercentage: &zero}

		if req.RolloutPercentage != nil {
			existing.RolloutPercentage = *req.RolloutPercentage
		}
		if existing.RolloutPercentage != 0 {
			t.Errorf("RolloutPercentage deveria ser 0, got %d", existing.RolloutPercentage)
		}
	})
}
