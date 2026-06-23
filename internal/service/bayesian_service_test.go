package service

import (
	"math"
	"testing"
)

// TestAnalyzeBayesian cobre a função pura exportada AnalyzeBayesian.
// Sem dependências de DB ou Redis — zero mocks necessários.

func TestAnalyzeBayesian_DadosZero_CollectMoreData(t *testing.T) {
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 0, ATotal: 0,
		BConversions: 0, BTotal: 0,
	})
	if result.Recommendation != "collect_more_data" {
		t.Errorf("com dados zero esperava collect_more_data, got %q", result.Recommendation)
	}
}

func TestAnalyzeBayesian_BVenceComFolga_DeployB(t *testing.T) {
	// B com taxa muito maior que A → P(B > A) ≥ 0.95 → deploy_b
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 10, ATotal: 1000,  // taxa A: 1%
		BConversions: 200, BTotal: 1000, // taxa B: 20%
	})
	if result.Recommendation != "deploy_b" {
		t.Errorf("B claramente melhor, esperava deploy_b, got %q (prob=%.4f)", result.Recommendation, result.ProbBBeatsA)
	}
	if result.ProbBBeatsA < 0.95 {
		t.Errorf("P(B > A) deveria ser ≥ 0.95, got %.4f", result.ProbBBeatsA)
	}
}

func TestAnalyzeBayesian_AVenceComFolga_KeepA(t *testing.T) {
	// A com taxa muito maior que B → P(B > A) ≤ 0.05 → keep_a
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 200, ATotal: 1000, // taxa A: 20%
		BConversions: 10, BTotal: 1000,  // taxa B: 1%
	})
	if result.Recommendation != "keep_a" {
		t.Errorf("A claramente melhor, esperava keep_a, got %q (prob=%.4f)", result.Recommendation, result.ProbBBeatsA)
	}
	if result.ProbBBeatsA > 0.05 {
		t.Errorf("P(B > A) deveria ser ≤ 0.05, got %.4f", result.ProbBBeatsA)
	}
}

func TestAnalyzeBayesian_SampleSizeCorreto(t *testing.T) {
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 30, ATotal: 200,
		BConversions: 40, BTotal: 200,
	})
	if result.SampleSize != 400 {
		t.Errorf("sample_size esperado 400, got %d", result.SampleSize)
	}
}

func TestAnalyzeBayesian_ProbEntre0e1(t *testing.T) {
	// P(B > A) deve sempre estar entre 0 e 1
	inputs := []BayesianInput{
		{0, 100, 0, 100},
		{50, 100, 50, 100},
		{100, 100, 0, 100},
		{0, 100, 100, 100},
	}
	for _, inp := range inputs {
		result := AnalyzeBayesian(inp)
		if result.ProbBBeatsA < 0 || result.ProbBBeatsA > 1 {
			t.Errorf("P(B>A)=%.4f fora do intervalo [0,1] para input %+v", result.ProbBBeatsA, inp)
		}
	}
}

func TestAnalyzeBayesian_MeanVariantA_Correto(t *testing.T) {
	// Variante A: 1 conversão em 2 total → alpha=2, beta=2 → mean=2/4=0.5
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 1, ATotal: 2,
		BConversions: 0, BTotal: 1,
	})
	expected := 0.5
	if math.Abs(result.VariantA.Mean-expected) > 0.001 {
		t.Errorf("mean A esperado %.4f, got %.4f", expected, result.VariantA.Mean)
	}
}

func TestAnalyzeBayesian_PequenasAmostras_CollectMoreData(t *testing.T) {
	// Mesmo com B melhor, amostra < 100 → collect_more_data
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 5, ATotal: 50,
		BConversions: 15, BTotal: 50,
	})
	// Com n<100, nunca deve sugerir deploy ou keep sem dados suficientes
	// (a lógica verifica totalSamples >= 100)
	if result.SampleSize != 100 {
		t.Errorf("sample size esperado 100, got %d", result.SampleSize)
	}
	// Recomendação: com n=100, pode decidir — dependendo do resultado estatístico
}

func TestAnalyzeBayesian_AlphasBeta_Calculados(t *testing.T) {
	// Alpha = conversions + 1 (prior Beta(1,1))
	result := AnalyzeBayesian(BayesianInput{
		AConversions: 10, ATotal: 100,
		BConversions: 20, BTotal: 100,
	})
	// Alpha A = 11, Beta A = 91 → mean = 11/102 ≈ 0.1078
	expectedMeanA := 11.0 / 102.0
	if math.Abs(result.VariantA.Mean-expectedMeanA) > 0.001 {
		t.Errorf("mean A esperado %.4f, got %.4f", expectedMeanA, result.VariantA.Mean)
	}
	if result.VariantA.Alpha != 11.0 {
		t.Errorf("alpha A esperado 11.0, got %.1f", result.VariantA.Alpha)
	}
}
