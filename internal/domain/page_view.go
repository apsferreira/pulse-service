package domain

import "time"

// PageView representa uma visualização de página rastreada pelo SDK.
type PageView struct {
	ID           string    `json:"id"`
	SessionID    *string   `json:"session_id,omitempty"`
	UserID       *string   `json:"user_id,omitempty"`
	Product      *string   `json:"product,omitempty"`
	PageTitle    *string   `json:"page_title,omitempty"`
	PagePath     *string   `json:"page_path,omitempty"`
	TimeOnPageMs *int      `json:"time_on_page_ms,omitempty"`
	Referrer     *string   `json:"referrer,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// TrackPageViewRequest é o payload de POST /api/v1/events/page_view.
type TrackPageViewRequest struct {
	SessionID    string  `json:"session_id"    validate:"omitempty,max=100"`
	UserID       *string `json:"user_id"       validate:"omitempty,uuid"`
	Product      string  `json:"product"       validate:"omitempty,max=50"`
	PageTitle    string  `json:"page_title"    validate:"omitempty,max=255"`
	PagePath     string  `json:"page_path"     validate:"required,max=500"`
	TimeOnPageMs *int    `json:"time_on_page_ms" validate:"omitempty,min=0"`
	Referrer     string  `json:"referrer"      validate:"omitempty,max=500"`
}

// TopPageItem representa uma entrada no ranking de páginas mais acessadas.
type TopPageItem struct {
	PagePath     string  `json:"page_path"`
	TotalViews   int     `json:"total_views"`
	AvgTimeOnPageMs float64 `json:"avg_time_on_page_ms"`
}

// TopPagesResponse é a resposta de GET /api/v1/analytics/top-pages.
type TopPagesResponse struct {
	Product string        `json:"product"`
	From    string        `json:"from"`
	To      string        `json:"to"`
	Items   []TopPageItem `json:"items"`
}

// UserJourneyItem representa uma página na jornada de um usuário.
type UserJourneyItem struct {
	ID           string    `json:"id"`
	SessionID    *string   `json:"session_id,omitempty"`
	Product      *string   `json:"product,omitempty"`
	PageTitle    *string   `json:"page_title,omitempty"`
	PagePath     *string   `json:"page_path,omitempty"`
	TimeOnPageMs *int      `json:"time_on_page_ms,omitempty"`
	Referrer     *string   `json:"referrer,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserJourneyResponse é a resposta de GET /api/v1/analytics/user-journey.
type UserJourneyResponse struct {
	UserID  string            `json:"user_id"`
	Product string            `json:"product"`
	Items   []UserJourneyItem `json:"items"`
}
