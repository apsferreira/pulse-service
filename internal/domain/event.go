package domain

import "time"

type Event struct {
	ID         string                 `json:"id" db:"id"`
	Product    string                 `json:"product" db:"product"`
	EventType  string                 `json:"event_type" db:"event_type"`
	UserID     *string                `json:"user_id,omitempty" db:"user_id"`
	SessionID  *string                `json:"session_id,omitempty" db:"session_id"`
	Properties map[string]interface{} `json:"properties" db:"properties"`
	IPAddress  *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

type IngestEventRequest struct {
	Product    string                 `json:"product" validate:"required"`
	EventType  string                 `json:"event_type" validate:"required"`
	UserID     *string                `json:"user_id,omitempty"`
	SessionID  *string                `json:"session_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type EventSummaryItem struct {
	Product   string `json:"product"`
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

type EventsSummaryResponse struct {
	Items        []EventSummaryItem `json:"items"`
	TotalLast24h int                `json:"total_last_24h"`
}

// TrackEventRequest is the payload for POST /api/v1/events/track.
type TrackEventRequest struct {
	EventName   string                 `json:"event_name"`
	Product     string                 `json:"product,omitempty"`
	UserID      *string                `json:"user_id,omitempty"`
	AnonymousID *string                `json:"anonymous_id,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// IdentifyRequest vincula um anonymous_id a um user_id real.
// Usado apos login/signup para unificar a jornada do usuario.
type IdentifyRequest struct {
	AnonymousID string `json:"anonymous_id"`
	UserID      string `json:"user_id"`
	Product     string `json:"product,omitempty"`
}

// EventCountItem represents a count grouped by event_name.
type EventCountItem struct {
	EventName string `json:"event_name"`
	Count     int    `json:"count"`
}

// EventCountResponse is the response for GET /api/v1/events/count.
type EventCountResponse struct {
	Items []EventCountItem `json:"items"`
	Total int              `json:"total"`
}

// EventListResponse is the response for GET /api/v1/events.
type EventListResponse struct {
	Items  []Event `json:"items"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// ── Retention metrics ────────────────────────────────────────────

// ActiveUsersItem represents DAU/WAU/MAU for a specific date.
type ActiveUsersItem struct {
	Date string `json:"date"`
	DAU  int    `json:"dau"`
	WAU  int    `json:"wau"`
	MAU  int    `json:"mau"`
}

// RetentionResponse is the response for GET /api/v1/metrics/retention.
type RetentionResponse struct {
	Product string            `json:"product"`
	Items   []ActiveUsersItem `json:"items"`
}

// CohortBucket represents retention for a cohort in a given period offset.
type CohortBucket struct {
	Period        int     `json:"period"`
	ActiveUsers   int     `json:"active_users"`
	RetentionPct  float64 `json:"retention_pct"`
}

// CohortItem represents a signup cohort and its retention over time.
type CohortItem struct {
	CohortLabel string         `json:"cohort_label"`
	CohortSize  int            `json:"cohort_size"`
	Buckets     []CohortBucket `json:"buckets"`
}

// CohortsResponse is the response for GET /api/v1/metrics/cohorts.
type CohortsResponse struct {
	Product    string       `json:"product"`
	Granularity string     `json:"granularity"`
	Cohorts    []CohortItem `json:"cohorts"`
}

// ActiveUsersCountItem represents a single period count.
type ActiveUsersCountItem struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

// ActiveUsersResponse is the response for GET /api/v1/metrics/active-users.
type ActiveUsersResponse struct {
	Product string                 `json:"product"`
	DAU     []ActiveUsersCountItem `json:"dau"`
	WAU     []ActiveUsersCountItem `json:"wau"`
	MAU     []ActiveUsersCountItem `json:"mau"`
}
