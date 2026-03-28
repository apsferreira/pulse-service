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
	EventName  string                 `json:"event_name"`
	UserID     *string                `json:"user_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
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
