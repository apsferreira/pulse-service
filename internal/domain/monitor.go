package domain

import "time"

type Monitor struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	URL             string    `json:"url" db:"url"`
	Method          string    `json:"method" db:"method"`
	IntervalSeconds int       `json:"interval_seconds" db:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds" db:"timeout_seconds"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type MonitorCheck struct {
	ID             string    `json:"id" db:"id"`
	MonitorID      string    `json:"monitor_id" db:"monitor_id"`
	Status         string    `json:"status" db:"status"` // "up", "down", "timeout"
	StatusCode     *int      `json:"status_code,omitempty" db:"status_code"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty" db:"response_time_ms"`
	ErrorMessage   *string   `json:"error_message,omitempty" db:"error_message"`
	CheckedAt      time.Time `json:"checked_at" db:"checked_at"`
}

type MonitorStatus struct {
	Monitor       Monitor       `json:"monitor"`
	LastCheck     *MonitorCheck `json:"last_check,omitempty"`
	Uptime24h     float64       `json:"uptime_24h"`   // % uptime last 24h
	AvgResponseMs float64       `json:"avg_response_ms"` // avg response time last 24h
}

type CreateMonitorRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	URL             string `json:"url" validate:"required"`
	Method          string `json:"method"`
	IntervalSeconds *int   `json:"interval_seconds"`
	TimeoutSeconds  *int   `json:"timeout_seconds"`
	IsActive        *bool  `json:"is_active"`
}

type UpdateMonitorRequest struct {
	Name            *string `json:"name"`
	URL             *string `json:"url"`
	Method          *string `json:"method"`
	IntervalSeconds *int    `json:"interval_seconds"`
	TimeoutSeconds  *int    `json:"timeout_seconds"`
	IsActive        *bool   `json:"is_active"`
}
