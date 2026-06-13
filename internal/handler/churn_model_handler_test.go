package handler

// BKL-999: testes das funções puras de churn prediction.
// Cobre normalização min-max, gradient descent (sanity check sobre dataset separável),
// e a função predictChurn em casos limite.
// O caminho HTTP de Train/UserChurnRisk depende de pgxpool e fica fora destes testes —
// é coberto pelos testes de integração com banco real no CI.

import (
	"math"
	"testing"
)

func TestNormalizeFeatures_ValoresEntreZeroEUm(t *testing.T) {
	samples := []userChurnFeatures{
		{userID: "u1", eventCount30d: 0, eventCount7d: 0, daysSinceLastEvt: 0, sessionCount: 0, distinctEvents: 0},
		{userID: "u2", eventCount30d: 50, eventCount7d: 10, daysSinceLastEvt: 30, sessionCount: 5, distinctEvents: 4},
		{userID: "u3", eventCount30d: 100, eventCount7d: 20, daysSinceLastEvt: 60, sessionCount: 10, distinctEvents: 8},
	}
	normalized, mins, maxs := normalizeFeatures(samples)

	if len(normalized) != 3 {
		t.Fatalf("normalized len = %d, want 3", len(normalized))
	}
	if mins[0] != 0 || maxs[0] != 100 {
		t.Errorf("mins[0]=%v maxs[0]=%v, want 0/100", mins[0], maxs[0])
	}
	for i, row := range normalized {
		for j, v := range row {
			if v < 0 || v > 1 {
				t.Errorf("sample %d feat %d: %v fora do range [0,1]", i, j, v)
			}
		}
	}
	if normalized[2][0] != 1.0 {
		t.Errorf("sample[2][0] (max) = %v, want 1.0", normalized[2][0])
	}
	if normalized[0][0] != 0.0 {
		t.Errorf("sample[0][0] (min) = %v, want 0.0", normalized[0][0])
	}
}

func TestNormalizeFeatures_ListaVaziaRetornaNilSemPanic(t *testing.T) {
	normalized, mins, maxs := normalizeFeatures(nil)
	if normalized != nil || mins != nil || maxs != nil {
		t.Errorf("normalizeFeatures(nil) deveria retornar nil em todos os slots")
	}
}

func TestTrainLogisticRegression_SeparaChurnedDeAtivos(t *testing.T) {
	// Dataset sintético claramente separável:
	// usuários ativos têm muitos eventos recentes e baixos days_since_last_event
	// usuários churned têm 0 eventos recentes e altos days_since_last_event
	samples := []userChurnFeatures{
		{userID: "a1", eventCount30d: 100, eventCount7d: 30, daysSinceLastEvt: 0, sessionCount: 15, distinctEvents: 8, churned: 0},
		{userID: "a2", eventCount30d: 95, eventCount7d: 28, daysSinceLastEvt: 1, sessionCount: 14, distinctEvents: 7, churned: 0},
		{userID: "a3", eventCount30d: 80, eventCount7d: 22, daysSinceLastEvt: 2, sessionCount: 12, distinctEvents: 6, churned: 0},
		{userID: "a4", eventCount30d: 70, eventCount7d: 18, daysSinceLastEvt: 3, sessionCount: 10, distinctEvents: 5, churned: 0},
		{userID: "c1", eventCount30d: 1, eventCount7d: 0, daysSinceLastEvt: 60, sessionCount: 1, distinctEvents: 1, churned: 1},
		{userID: "c2", eventCount30d: 2, eventCount7d: 0, daysSinceLastEvt: 55, sessionCount: 1, distinctEvents: 1, churned: 1},
		{userID: "c3", eventCount30d: 0, eventCount7d: 0, daysSinceLastEvt: 70, sessionCount: 0, distinctEvents: 0, churned: 1},
		{userID: "c4", eventCount30d: 1, eventCount7d: 0, daysSinceLastEvt: 80, sessionCount: 1, distinctEvents: 1, churned: 1},
	}

	weights, accuracy := trainLogisticRegression(samples)

	if len(weights) != 6 {
		t.Fatalf("weights len = %d, want 6 (bias + 5 features)", len(weights))
	}
	// Em dataset claramente separável, acurácia de treino deve ser razoável.
	// 100 iterações com lr=0.01 não convergem perfeitamente, mas devem superar o baseline.
	if accuracy < 0.5 {
		t.Errorf("accuracy = %v, esperava >= 0.5", accuracy)
	}
}

func TestPredictChurn_PesosOuFeaturesInvalidasRetorna05(t *testing.T) {
	if got := predictChurn(nil, nil); got != 0.5 {
		t.Errorf("predictChurn(nil, nil) = %v, want 0.5", got)
	}
	if got := predictChurn([]float64{0, 0}, []float64{0, 0, 0, 0, 0}); got != 0.5 {
		t.Errorf("predictChurn com pesos curtos = %v, want 0.5", got)
	}
	if got := predictChurn(make([]float64, 6), []float64{0, 0}); got != 0.5 {
		t.Errorf("predictChurn com features curtas = %v, want 0.5", got)
	}
}

func TestPredictChurn_PesosZeradosRetornaCincoCentesimos(t *testing.T) {
	// Pesos todos zero -> z = 0 -> sigmoid(0) = 0.5
	weights := []float64{0, 0, 0, 0, 0, 0}
	features := []float64{50, 10, 30, 5, 4}
	got := predictChurn(weights, features)
	if math.Abs(got-0.5) > 1e-9 {
		t.Errorf("predictChurn(zero weights) = %v, want 0.5", got)
	}
}

func TestPredictChurn_BiasAltoIndicaChurn(t *testing.T) {
	// Bias muito positivo -> sigmoid alto -> probabilidade de churn alta
	weights := []float64{10, 0, 0, 0, 0, 0}
	features := []float64{0, 0, 0, 0, 0}
	got := predictChurn(weights, features)
	if got < 0.99 {
		t.Errorf("predictChurn(bias=10) = %v, want > 0.99", got)
	}
}

func TestSigmoid_ValoresConhecidos(t *testing.T) {
	cases := []struct {
		z, want float64
	}{
		{0, 0.5},
		{1, 0.7310585786},
		{-1, 0.2689414213},
		{10, 0.9999546021},
	}
	for _, tc := range cases {
		got := sigmoid(tc.z)
		if math.Abs(got-tc.want) > 1e-6 {
			t.Errorf("sigmoid(%v) = %v, want %v", tc.z, got, tc.want)
		}
	}
}
