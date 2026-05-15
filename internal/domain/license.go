package domain

import "time"

// LicenseKey representa uma licença de uso do Pulse self-hosted.
type LicenseKey struct {
	ID                 string     `json:"id"`
	Key                string     `json:"key"`
	Plan               string     `json:"plan"`
	CustomerEmail      string     `json:"customer_email"`
	CustomerName       string     `json:"customer_name,omitempty"`
	OrderID            string     `json:"order_id,omitempty"`
	ValidFrom          time.Time  `json:"valid_from"`
	ValidUntil         time.Time  `json:"valid_until"`
	ActivatedAt        *time.Time `json:"activated_at,omitempty"`
	ActivationHostname string     `json:"activation_hostname,omitempty"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// LicenseValidateResponse é a resposta do endpoint POST /api/v1/license/validate.
type LicenseValidateResponse struct {
	Valid      bool      `json:"valid"`
	Plan       string    `json:"plan"`
	ValidUntil time.Time `json:"valid_until"`
	DaysLeft   int       `json:"days_left"`
	Message    string    `json:"message,omitempty"`
}
