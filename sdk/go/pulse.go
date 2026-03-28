// Package pulse provides a Go SDK for the pulse-service feature flags and event tracking API.
package pulse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// validKeyPattern restricts flag keys to safe characters only.
var validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,100}$`)

// Client is the pulse-service SDK client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		client.httpClient = c
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(client *Client) {
		client.httpClient.Timeout = d
	}
}

// NewClient creates a new pulse-service SDK client.
// baseURL should be the root URL of the pulse-service (e.g. "https://pulse.institutoitinerante.com.br").
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("pulse: baseURL is required")
	}
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("pulse: invalid baseURL: %w", err)
	}

	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// --- Feature Flags ---

// FlagsMapResponse matches the GET /flags JSON response.
type FlagsMapResponse struct {
	Flags map[string]bool `json:"flags"`
}

// GetFlags returns all feature flags as a map of key -> enabled.
func (c *Client) GetFlags(ctx context.Context) (map[string]bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/flags", nil)
	if err != nil {
		return nil, fmt.Errorf("pulse: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pulse: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pulse: unexpected status %d", resp.StatusCode)
	}

	var result FlagsMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("pulse: failed to decode response: %w", err)
	}
	return result.Flags, nil
}

// FeatureFlag represents a single feature flag with full details.
type FeatureFlag struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetFlag returns a single feature flag by key (requires service token auth on the server).
// This calls the admin endpoint GET /api/v1/flags/:key.
func (c *Client) GetFlag(ctx context.Context, key string) (*FeatureFlag, error) {
	if !validKeyPattern.MatchString(key) {
		return nil, fmt.Errorf("pulse: invalid flag key %q", key)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/flags/"+url.PathEscape(key), nil)
	if err != nil {
		return nil, fmt.Errorf("pulse: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pulse: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pulse: unexpected status %d", resp.StatusCode)
	}

	var flag FeatureFlag
	if err := json.NewDecoder(resp.Body).Decode(&flag); err != nil {
		return nil, fmt.Errorf("pulse: failed to decode response: %w", err)
	}
	return &flag, nil
}

// IsEnabled checks if a feature flag is enabled.
// Uses the public GET /flags endpoint and checks the key in the returned map.
// Returns false if the flag does not exist or on error.
func (c *Client) IsEnabled(ctx context.Context, key string) (bool, error) {
	if !validKeyPattern.MatchString(key) {
		return false, fmt.Errorf("pulse: invalid flag key %q", key)
	}

	flags, err := c.GetFlags(ctx)
	if err != nil {
		return false, err
	}
	return flags[key], nil
}

// --- Event Tracking ---

// TrackEventRequest is the payload for POST /api/v1/events/track.
type TrackEventRequest struct {
	EventName  string                 `json:"event_name"`
	UserID     *string                `json:"user_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// TrackEventResponse is the response from POST /api/v1/events/track.
type TrackEventResponse struct {
	ID        string                 `json:"id"`
	Product   string                 `json:"product"`
	EventType string                 `json:"event_type"`
	UserID    *string                `json:"user_id,omitempty"`
	SessionID *string                `json:"session_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// TrackEvent sends a tracking event to the pulse-service.
func (c *Client) TrackEvent(ctx context.Context, event TrackEventRequest) (*TrackEventResponse, error) {
	if event.EventName == "" {
		return nil, fmt.Errorf("pulse: event_name is required")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("pulse: failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/events/track", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("pulse: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pulse: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("pulse: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result TrackEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("pulse: failed to decode response: %w", err)
	}
	return &result, nil
}
