package domain

import (
	"time"

	"github.com/google/uuid"
)

type MABExperiment struct {
	ID        uuid.UUID        `json:"id"`
	ProjectID uuid.UUID        `json:"project_id"`
	FlagKey   string           `json:"flag_key"`
	Epsilon   float64          `json:"epsilon"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Variants  []MABVariantStats `json:"variants,omitempty"`
}

type MABVariantStats struct {
	ID           uuid.UUID `json:"id"`
	ExperimentID uuid.UUID `json:"experiment_id"`
	VariantKey   string    `json:"variant_key"`
	Impressions  int64     `json:"impressions"`
	Conversions  int64     `json:"conversions"`
	Weight       float64   `json:"weight"`
	ConvRate     float64   `json:"conversion_rate"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateMABExperimentRequest struct {
	FlagKey  string             `json:"flag_key" validate:"required"`
	Epsilon  float64            `json:"epsilon" validate:"gte=0,lte=1"`
	Variants []MABVariantRequest `json:"variants" validate:"required,min=2"`
}

type MABVariantRequest struct {
	VariantKey string `json:"variant_key" validate:"required"`
}

type RecordMABEventRequest struct {
	VariantKey   string `json:"variant_key" validate:"required"`
	IsConversion bool   `json:"is_conversion"`
}

type SetMABActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type SelectVariantResponse struct {
	VariantKey string  `json:"variant_key"`
	IsExplore  bool    `json:"is_explore"`
	Weight     float64 `json:"weight"`
}
