package service

// BKL-1043: Chi-squared test para A/B testing no Pulse.
// Implementação zero-dependência em Go puro.
//
// Fórmula chi-squared:
//   χ² = Σ (observed - expected)² / expected
// onde expected = (row_total * col_total) / grand_total
//
// p-value calculado via approximação CDF da distribuição qui-quadrado com 1 grau de liberdade
// usando o complemento da função erro: p = 1 - erf(√(χ²/2)) / √π
//
// Threshold: α = 0.05 (95% de confiança — padrão da indústria de A/B testing)

import (
	"math"
)

// ExperimentResult contém os resultados do teste estatístico de um experimento A/B.
type ExperimentResult struct {
	// Variante de controle
	ControlConversions int `json:"control_conversions"`
	ControlTotal       int `json:"control_total"`
	ControlRate        float64 `json:"control_rate"` // taxa de conversão 0-1

	// Variante de teste (challenger)
	TreatmentConversions int `json:"treatment_conversions"`
	TreatmentTotal       int `json:"treatment_total"`
	TreatmentRate        float64 `json:"treatment_rate"`

	// Estatística
	ChiSquared               float64 `json:"chi_squared"`
	PValue                   float64 `json:"p_value"`
	StatisticallySignificant bool    `json:"statistically_significant"` // p < 0.05
	ConfidenceLevel          float64 `json:"confidence_level"`          // 1 - p_value (ex: 0.95)
	// BKL-419: campo legível de significância estatística
	Confidence string `json:"confidence"` // "90%", "95%", "99%" ou "not_significant"

	// Lift relativo entre tratamento e controle
	RelativeLift    float64 `json:"relative_lift"` // ex: +12.5 (%)
	AbsoluteLift    float64 `json:"absolute_lift"` // ex: +0.03 (3pp)

	// Interpretação
	Recommendation string `json:"recommendation"` // "keep", "reject", "collect_more_data"
	SampleSize     int    `json:"sample_size"` // total de observações
}

// ExperimentInput é o payload para análise estatística de experimento A/B.
type ExperimentInput struct {
	// Controle
	ControlConversions int `json:"control_conversions"`
	ControlTotal       int `json:"control_total"`

	// Tratamento
	TreatmentConversions int `json:"treatment_conversions"`
	TreatmentTotal       int `json:"treatment_total"`
}

// AnalyzeExperiment realiza um chi-squared test 2x2 e retorna resultado estatístico.
// BKL-1043: zero dependências externas — usa apenas math.
func AnalyzeExperiment(input ExperimentInput) ExperimentResult {
	cConv := input.ControlConversions
	cTotal := input.ControlTotal
	tConv := input.TreatmentConversions
	tTotal := input.TreatmentTotal

	if cTotal <= 0 || tTotal <= 0 {
		return ExperimentResult{Recommendation: "collect_more_data"}
	}

	cRate := float64(cConv) / float64(cTotal)
	tRate := float64(tConv) / float64(tTotal)

	// Tabela 2x2:
	// [cConv,    cTotal-cConv]   (controle: converteu, não converteu)
	// [tConv,    tTotal-tConv]   (tratamento: converteu, não converteu)

	a := float64(cConv)
	b := float64(cTotal - cConv)
	c := float64(tConv)
	d := float64(tTotal - tConv)

	n := a + b + c + d

	// Expected values
	eA := (a + c) * (a + b) / n
	eB := (b + d) * (a + b) / n
	eC := (a + c) * (c + d) / n
	eD := (b + d) * (c + d) / n

	chiSq := 0.0
	for _, pair := range [][2]float64{{a, eA}, {b, eB}, {c, eC}, {d, eD}} {
		obs, exp := pair[0], pair[1]
		if exp > 0 {
			chiSq += (obs-exp) * (obs-exp) / exp
		}
	}

	// p-value para chi-squared com 1 grau de liberdade
	pValue := chiSquaredPValue(chiSq, 1)

	significant := pValue < 0.05
	confidence := 1.0 - pValue

	// BKL-419: nível de confiança legível via lookup de chi-squared (df=1)
	// Thresholds: p=0.001 → χ²≥10.83 (99.9%), p=0.01 → χ²≥6.63 (99%), p=0.05 → χ²≥3.84 (95%), p=0.10 → χ²≥2.71 (90%)
	confidenceLabel := "not_significant"
	switch {
	case chiSq >= 10.83:
		confidenceLabel = "99.9%"
	case chiSq >= 6.63:
		confidenceLabel = "99%"
	case chiSq >= 3.84:
		confidenceLabel = "95%"
	case chiSq >= 2.71:
		confidenceLabel = "90%"
	}

	// Lift
	absoluteLift := tRate - cRate
	relativeLift := 0.0
	if cRate > 0 {
		relativeLift = absoluteLift / cRate * 100
	}

	// Recomendação
	recommendation := "collect_more_data"
	if significant {
		if tRate > cRate {
			recommendation = "keep" // tratamento ganhou
		} else {
			recommendation = "reject" // controle ganhou, reverter tratamento
		}
	}

	return ExperimentResult{
		ControlConversions:       cConv,
		ControlTotal:             cTotal,
		ControlRate:              cRate,
		TreatmentConversions:     tConv,
		TreatmentTotal:           tTotal,
		TreatmentRate:            tRate,
		ChiSquared:               math.Round(chiSq*10000) / 10000,
		PValue:                   math.Round(pValue*10000) / 10000,
		StatisticallySignificant: significant,
		ConfidenceLevel:          math.Round(confidence*10000) / 10000,
		Confidence:               confidenceLabel,
		AbsoluteLift:             math.Round(absoluteLift*10000) / 10000,
		RelativeLift:             math.Round(relativeLift*100) / 100,
		Recommendation:           recommendation,
		SampleSize:               cTotal + tTotal,
	}
}

// chiSquaredPValue calcula o p-value para um valor chi-squared com df graus de liberdade.
// Para df=1: usa aproximação via função erro complementar (erfc).
// Para df>1: aproximação de Wilson-Hilferty.
func chiSquaredPValue(chiSq float64, df int) float64 {
	if chiSq <= 0 {
		return 1.0
	}

	if df == 1 {
		// Para 1 grau de liberdade: p = 1 - erf(√(χ²/2)) = erfc(√(χ²/2))
		return math.Erfc(math.Sqrt(chiSq/2))
	}

	// Aproximação de Wilson-Hilferty para df > 1
	// χ² com df ~ Normal(df, 2df) transformado
	k := float64(df)
	z := math.Pow(chiSq/k, 1.0/3.0) - (1.0 - 2.0/(9.0*k))
	sigma := math.Sqrt(2.0 / (9.0 * k))
	if sigma <= 0 {
		return 1.0
	}
	zScore := z / sigma
	// p-value = P(Z > zScore) para distribuição normal padrão
	return 0.5 * math.Erfc(zScore/math.Sqrt2)
}
