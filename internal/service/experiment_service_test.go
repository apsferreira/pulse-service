package service

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// TestAnalyzeExperiment
// ---------------------------------------------------------------------------

func TestAnalyzeExperiment(t *testing.T) {
	t.Run("dados insuficientes retorna collect_more_data", func(t *testing.T) {
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   0,
			ControlTotal:         0,
			TreatmentConversions: 0,
			TreatmentTotal:       0,
		})
		if result.Recommendation != "collect_more_data" {
			t.Errorf("esperava 'collect_more_data' para dados zerados, got %q", result.Recommendation)
		}
	})

	t.Run("controle com total zero retorna collect_more_data", func(t *testing.T) {
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   0,
			ControlTotal:         0,
			TreatmentConversions: 50,
			TreatmentTotal:       100,
		})
		if result.Recommendation != "collect_more_data" {
			t.Errorf("esperava 'collect_more_data' quando ControlTotal=0, got %q", result.Recommendation)
		}
	})

	t.Run("tratamento claramente melhor → recommendation keep", func(t *testing.T) {
		// 200 conversões em 1000 (20%) vs 300 em 1000 (30%) — diferença grande, p < 0.05
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   200,
			ControlTotal:         1000,
			TreatmentConversions: 300,
			TreatmentTotal:       1000,
		})
		if !result.StatisticallySignificant {
			t.Error("esperava resultado estatisticamente significativo")
		}
		if result.Recommendation != "keep" {
			t.Errorf("esperava 'keep' (tratamento venceu), got %q", result.Recommendation)
		}
		if result.TreatmentRate <= result.ControlRate {
			t.Errorf("TreatmentRate (%.3f) deveria ser maior que ControlRate (%.3f)", result.TreatmentRate, result.ControlRate)
		}
	})

	t.Run("controle claramente melhor → recommendation reject", func(t *testing.T) {
		// controle 30% vs tratamento 10% — tratamento pior
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   300,
			ControlTotal:         1000,
			TreatmentConversions: 100,
			TreatmentTotal:       1000,
		})
		if !result.StatisticallySignificant {
			t.Error("esperava resultado estatisticamente significativo")
		}
		if result.Recommendation != "reject" {
			t.Errorf("esperava 'reject' (controle venceu), got %q", result.Recommendation)
		}
	})

	t.Run("taxas idênticas → sem significância", func(t *testing.T) {
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   100,
			ControlTotal:         500,
			TreatmentConversions: 100,
			TreatmentTotal:       500,
		})
		if result.StatisticallySignificant {
			t.Error("taxas idênticas não deveriam ser estatisticamente significativas")
		}
		if result.Recommendation != "collect_more_data" {
			t.Errorf("esperava 'collect_more_data' para taxas iguais, got %q", result.Recommendation)
		}
	})

	t.Run("sample_size é soma de controle e tratamento", func(t *testing.T) {
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   50,
			ControlTotal:         200,
			TreatmentConversions: 60,
			TreatmentTotal:       300,
		})
		if result.SampleSize != 500 {
			t.Errorf("esperava SampleSize=500, got %d", result.SampleSize)
		}
	})

	t.Run("lift relativo calculado corretamente", func(t *testing.T) {
		// controle 10%, tratamento 20% → lift absoluto = +10pp, relativo = +100%
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   100,
			ControlTotal:         1000,
			TreatmentConversions: 200,
			TreatmentTotal:       1000,
		})
		if math.Abs(result.ControlRate-0.10) > 0.001 {
			t.Errorf("ControlRate esperado 0.10, got %.4f", result.ControlRate)
		}
		if math.Abs(result.TreatmentRate-0.20) > 0.001 {
			t.Errorf("TreatmentRate esperado 0.20, got %.4f", result.TreatmentRate)
		}
		if math.Abs(result.RelativeLift-100.0) > 1.0 {
			t.Errorf("RelativeLift esperado ~100%%, got %.2f", result.RelativeLift)
		}
	})

	t.Run("PValue e ChiSquared são arredondados a 4 casas decimais", func(t *testing.T) {
		result := AnalyzeExperiment(ExperimentInput{
			ControlConversions:   200,
			ControlTotal:         1000,
			TreatmentConversions: 300,
			TreatmentTotal:       1000,
		})
		// Verifica que ChiSquared tem no máximo 4 casas decimais
		rounded := math.Round(result.ChiSquared*10000) / 10000
		if rounded != result.ChiSquared {
			t.Errorf("ChiSquared não foi arredondado a 4 casas: %v", result.ChiSquared)
		}
	})
}

// ---------------------------------------------------------------------------
// TestChiSquaredPValue
// ---------------------------------------------------------------------------

func TestChiSquaredPValue(t *testing.T) {
	t.Run("chi-squared zero retorna p=1 (sem evidência)", func(t *testing.T) {
		p := chiSquaredPValue(0, 1)
		if p != 1.0 {
			t.Errorf("esperava p=1.0 para chi²=0, got %.4f", p)
		}
	})

	t.Run("chi-squared muito alto retorna p próximo de zero (alta significância)", func(t *testing.T) {
		// chi²=100 com df=1 → p virtualmente zero
		p := chiSquaredPValue(100, 1)
		if p >= 0.05 {
			t.Errorf("esperava p<0.05 para chi²=100, got %.6f", p)
		}
	})

	t.Run("chi-squared 3.84 (threshold α=0.05 com df=1) retorna p≈0.05", func(t *testing.T) {
		// Valor crítico padrão para α=0.05 com 1 grau de liberdade
		p := chiSquaredPValue(3.841, 1)
		if p >= 0.06 || p <= 0.04 {
			t.Errorf("esperava p≈0.05 para chi²=3.841, got %.4f", p)
		}
	})

	t.Run("df>1 usa aproximação de Wilson-Hilferty (resultado plausível)", func(t *testing.T) {
		// chi²=5.99 com df=2 → threshold α=0.05
		p := chiSquaredPValue(5.99, 2)
		if p <= 0 || p > 1 {
			t.Errorf("p-value inválido para df=2: %.6f", p)
		}
		if p >= 0.10 {
			t.Errorf("esperava p<0.10 para chi²=5.99 df=2, got %.4f", p)
		}
	})
}
