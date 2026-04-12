package service

// BKL-1001: Bayesian inference para experimentos A/B.
// Calcula distribuição Beta a posteriori para cada variante e
// P(B > A) via Monte Carlo simulation (10k samples).
// Zero dependências externas — usa apenas math/rand da stdlib.

import (
	"math"
	"math/rand"
)

// BayesianVariant representa uma variante com sua distribuição Beta a posteriori.
type BayesianVariant struct {
	Alpha float64 `json:"alpha"` // alpha = conversions + 1
	Beta  float64 `json:"beta"`  // beta  = non_conversions + 1
	Mean  float64 `json:"mean"`  // média da distribuição Beta = alpha / (alpha + beta)
}

// BayesianResult é o resultado da análise Bayesiana de um experimento A/B.
type BayesianResult struct {
	VariantA      BayesianVariant `json:"variant_a"`
	VariantB      BayesianVariant `json:"variant_b"`
	ProbBBeatsA   float64         `json:"prob_b_beats_a"`    // P(B > A) — 0.0 a 1.0
	Recommendation string         `json:"recommendation"`    // "deploy_b" | "keep_a" | "collect_more_data"
	SampleSize    int             `json:"sample_size"`
}

// BayesianInput é a entrada para análise Bayesiana.
type BayesianInput struct {
	AConversions    int `json:"a_conversions"`
	ATotal          int `json:"a_total"`
	BConversions    int `json:"b_conversions"`
	BTotal          int `json:"b_total"`
}

const monteCarloSamples = 10_000

// AnalyzeBayesian realiza análise Bayesiana com Beta distribution a posteriori.
// Usa prior não-informativo Beta(1,1) = uniforme.
// BKL-1001: P(B > A) calculado via Monte Carlo com 10k samples.
func AnalyzeBayesian(input BayesianInput) BayesianResult {
	// Alpha = conversions + 1 (prior Beta(1,1) é incorporado somando 1)
	aAlpha := float64(input.AConversions) + 1.0
	aBeta := float64(input.ATotal-input.AConversions) + 1.0
	bAlpha := float64(input.BConversions) + 1.0
	bBeta := float64(input.BTotal-input.BConversions) + 1.0

	varA := BayesianVariant{
		Alpha: aAlpha,
		Beta:  aBeta,
		Mean:  round4(aAlpha / (aAlpha + aBeta)),
	}
	varB := BayesianVariant{
		Alpha: bAlpha,
		Beta:  bBeta,
		Mean:  round4(bAlpha / (bAlpha + bBeta)),
	}

	// Monte Carlo: amostra 10k pares de Beta(a,b) e conta quantos B > A
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // seed fixo para reproducibilidade
	winsB := 0
	for i := 0; i < monteCarloSamples; i++ {
		sampleA := sampleBeta(rng, aAlpha, aBeta)
		sampleB := sampleBeta(rng, bAlpha, bBeta)
		if sampleB > sampleA {
			winsB++
		}
	}

	probBBeatsA := round4(float64(winsB) / float64(monteCarloSamples))

	recommendation := "collect_more_data"
	totalSamples := input.ATotal + input.BTotal
	if totalSamples >= 100 {
		switch {
		case probBBeatsA >= 0.95:
			recommendation = "deploy_b"
		case probBBeatsA <= 0.05:
			recommendation = "keep_a"
		}
	}

	return BayesianResult{
		VariantA:       varA,
		VariantB:       varB,
		ProbBBeatsA:    probBBeatsA,
		Recommendation: recommendation,
		SampleSize:     totalSamples,
	}
}

// sampleBeta gera uma amostra da distribuição Beta(alpha, beta) usando o método
// de Johnk (soma de gammas): Beta(a,b) = Gamma(a) / (Gamma(a) + Gamma(b)).
// Para alpha=beta=1 (uniforme), retorna rand.Float64() diretamente.
func sampleBeta(rng *rand.Rand, alpha, beta float64) float64 {
	if alpha == 1.0 && beta == 1.0 {
		return rng.Float64()
	}
	x := sampleGamma(rng, alpha)
	y := sampleGamma(rng, beta)
	if x+y == 0 {
		return 0.5
	}
	return x / (x + y)
}

// sampleGamma gera uma amostra de Gamma(shape, scale=1) usando o algoritmo
// de Marsaglia-Tsang (2000) para shape >= 1, e a transformação shape < 1.
func sampleGamma(rng *rand.Rand, shape float64) float64 {
	if shape < 1.0 {
		// Transformação: Gamma(k) = Gamma(k+1) * U^(1/k)
		u := rng.Float64()
		return sampleGamma(rng, shape+1.0) * math.Pow(u, 1.0/shape)
	}

	// Marsaglia-Tsang para shape >= 1
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)
	for {
		var x, v float64
		for {
			x = rng.NormFloat64()
			v = 1.0 + c*x
			if v > 0 {
				break
			}
		}
		v = v * v * v
		u := rng.Float64()
		if u < 1.0-0.0331*(x*x)*(x*x) {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
			return d * v
		}
	}
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
