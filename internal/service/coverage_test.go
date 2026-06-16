// Package service — testes de cobertura para funções puras (Issue #52).
package service

import (
	"math"
	"strings"
	"testing"
)

// ─── calcStats ───────────────────────────────────────────────────────────────

func TestCalcStats_Vazio(t *testing.T) {
	mean, stddev := calcStats(nil)
	if mean != 0 || stddev != 0 {
		t.Errorf("esperado (0,0), obtido (%.2f,%.2f)", mean, stddev)
	}
}

func TestCalcStats_UnicoValor(t *testing.T) {
	mean, stddev := calcStats([]float64{10.0})
	if math.Abs(mean-10.0) > 0.001 {
		t.Errorf("mean esperado 10.0, obtido %.4f", mean)
	}
	if math.Abs(stddev) > 0.001 {
		t.Errorf("stddev esperado 0.0, obtido %.4f", stddev)
	}
}

func TestCalcStats_ValoresIguais(t *testing.T) {
	mean, stddev := calcStats([]float64{5, 5, 5, 5})
	if math.Abs(mean-5.0) > 0.001 || math.Abs(stddev) > 0.001 {
		t.Errorf("esperado mean=5 stddev=0, obtido mean=%.2f stddev=%.2f", mean, stddev)
	}
}

func TestCalcStats_Dispersao(t *testing.T) {
	// [2,4,4,4,5,5,7,9] → mean=5, stddev=2
	mean, stddev := calcStats([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if math.Abs(mean-5.0) > 0.001 {
		t.Errorf("mean esperado 5.0, obtido %.4f", mean)
	}
	if math.Abs(stddev-2.0) > 0.001 {
		t.Errorf("stddev esperado 2.0, obtido %.4f", stddev)
	}
}

func TestCalcStats_NegativosMisturados(t *testing.T) {
	_, stddev := calcStats([]float64{-3, -1, 1, 3})
	if stddev <= 0 {
		t.Error("stddev deve ser positivo para valores não-uniformes")
	}
}

// ─── parseWindowDays ─────────────────────────────────────────────────────────

func TestParseWindowDays_Vazio(t *testing.T) {
	days, err := parseWindowDays("")
	if err != nil {
		t.Fatalf("não esperava erro, obteve: %v", err)
	}
	if days != 7 {
		t.Errorf("esperado 7, obtido %d", days)
	}
}

func TestParseWindowDays_Dias(t *testing.T) {
	casos := []struct {
		input string
		want  int
	}{
		{"7d", 7},
		{"14d", 14},
		{"28d", 28},
		{"1d", 1},
	}
	for _, c := range casos {
		t.Run(c.input, func(t *testing.T) {
			days, err := parseWindowDays(c.input)
			if err != nil {
				t.Fatalf("input=%q: não esperava erro, obteve: %v", c.input, err)
			}
			if days != c.want {
				t.Errorf("input=%q: esperado %d, obtido %d", c.input, c.want, days)
			}
		})
	}
}

func TestParseWindowDays_Semanas(t *testing.T) {
	days, err := parseWindowDays("4w")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if days != 28 {
		t.Errorf("4w deve ser 28 dias, obtido %d", days)
	}
}

func TestParseWindowDays_Invalidos(t *testing.T) {
	invalidos := []string{"x", "0d", "-1d", "7x", "d"}
	for _, inv := range invalidos {
		t.Run(inv, func(t *testing.T) {
			_, err := parseWindowDays(inv)
			if err == nil {
				t.Errorf("input=%q deveria retornar erro", inv)
			}
		})
	}
}

// ─── buildDefaultExplanation ─────────────────────────────────────────────────

func TestBuildDefaultExplanation_Spike(t *testing.T) {
	r := &AnomalyResult{
		Metric:         "eventos",
		CurrentValue:   200,
		ExpectedValue:  100,
		DeviationSigma: 3.5,
		Direction:      AnomalyDirectionSpike,
	}
	msg := buildDefaultExplanation(r)
	if !strings.Contains(msg, "Aumento") {
		t.Errorf("esperava 'Aumento' em: %q", msg)
	}
	if !strings.Contains(msg, "eventos") {
		t.Errorf("esperava 'eventos' em: %q", msg)
	}
}

func TestBuildDefaultExplanation_Drop(t *testing.T) {
	r := &AnomalyResult{
		Metric:         "logins",
		CurrentValue:   20,
		ExpectedValue:  100,
		DeviationSigma: 2.1,
		Direction:      AnomalyDirectionDrop,
	}
	msg := buildDefaultExplanation(r)
	if !strings.Contains(msg, "Queda") {
		t.Errorf("esperava 'Queda' em: %q", msg)
	}
}

func TestBuildDefaultExplanation_ExpectedZero(t *testing.T) {
	r := &AnomalyResult{
		Metric:        "posts",
		CurrentValue:  50,
		ExpectedValue: 0,
		Direction:     AnomalyDirectionSpike,
	}
	msg := buildDefaultExplanation(r)
	if msg == "" {
		t.Error("mensagem não deve ser vazia")
	}
}

// ─── scoreFromCount ──────────────────────────────────────────────────────────

func TestScoreFromCount(t *testing.T) {
	thresholds := map[int]int{1: 10, 4: 20, 10: 30}

	casos := []struct {
		count int
		want  int
	}{
		{0, 0},
		{-5, 0},
		{1, 10},
		{3, 10},
		{4, 20},
		{9, 20},
		{10, 30},
		{100, 30},
	}
	for _, c := range casos {
		t.Run("", func(t *testing.T) {
			got := scoreFromCount(c.count, 30, thresholds)
			if got != c.want {
				t.Errorf("count=%d: esperado %d, obtido %d", c.count, c.want, got)
			}
		})
	}
}

func TestScoreFromCount_LimiteMax(t *testing.T) {
	thresholds := map[int]int{1: 50}
	got := scoreFromCount(10, 30, thresholds)
	if got != 30 {
		t.Errorf("score não deve exceder max=30, obtido %d", got)
	}
}

// ─── capitalizeFirst ─────────────────────────────────────────────────────────

func TestCapitalizeFirst(t *testing.T) {
	casos := []struct{ input, want string }{
		{"", ""},
		{"anomalia", "Anomalia"},
		// a função subtrai 32 do rune — comportamento definido apenas para ASCII minúsculo
		{"a", "A"},
		{"spike", "Spike"},
	}
	for _, c := range casos {
		got := capitalizeFirst(c.input)
		if got != c.want {
			t.Errorf("input=%q: esperado %q, obtido %q", c.input, c.want, got)
		}
	}
}

// ─── AnomalyDirection constants ──────────────────────────────────────────────

func TestAnomalyDirectionValores(t *testing.T) {
	if AnomalyDirectionSpike != "spike" {
		t.Errorf("esperado 'spike', obtido %q", AnomalyDirectionSpike)
	}
	if AnomalyDirectionDrop != "drop" {
		t.Errorf("esperado 'drop', obtido %q", AnomalyDirectionDrop)
	}
	if AnomalyDirectionNone != "none" {
		t.Errorf("esperado 'none', obtido %q", AnomalyDirectionNone)
	}
}
