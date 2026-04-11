package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserIdentity representa o mapeamento entre canonical_user_id (auth-service) e
// a identidade do usuário em um produto específico. BKL-763.
type UserIdentity struct {
	ID              int       `json:"id"`
	CanonicalUserID uuid.UUID `json:"canonical_user_id"`
	Product         string    `json:"product"`
	ProductUserID   string    `json:"product_user_id"`
	Email           string    `json:"email,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserIdentityUpsertRequest é o payload para POST /api/v1/user-identities.
type UserIdentityUpsertRequest struct {
	CanonicalUserID uuid.UUID `json:"canonical_user_id"`
	Product         string    `json:"product"`
	ProductUserID   string    `json:"product_user_id"`
	Email           string    `json:"email,omitempty"`
}

// UserIdentityListResponse agrupa identidades de um canonical_user_id.
type UserIdentityListResponse struct {
	CanonicalUserID uuid.UUID      `json:"canonical_user_id"`
	Identities      []UserIdentity `json:"identities"`
	Total           int            `json:"total"`
}
