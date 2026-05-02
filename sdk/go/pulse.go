// Package pulse provides a Go SDK for the Pulse feature flag service.
//
// Usage:
//
//	client := pulse.New("https://pulse-api.institutoitinerante.com.br", "service-token")
//	enabled := client.IsEnabled(ctx, "minha-flag", true)
//	variant := client.GetVariant(ctx, "pricing-test", "control", pulse.WithUser("user-123"))
//	client.Track("page.viewed", pulse.Props{"page": "/home"})
package pulse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,100}$`)

// Client é o cliente SDK do Pulse. Seguro para uso concorrente.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	mu           sync.RWMutex
	flagCache    map[string]bool
	variantCache map[string]string
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// Props é o tipo das propriedades de um evento de tracking.
type Props map[string]interface{}

// Option configura a avaliação de flags (user/tenant context).
type Option func(*evalConfig)

type evalConfig struct {
	userID   string
	tenantID string
}

// WithUser define o user ID para avaliação de rollout e A/B.
func WithUser(userID string) Option {
	return func(c *evalConfig) { c.userID = userID }
}

// WithTenant define o tenant ID para avaliação de targeting.
func WithTenant(tenantID string) Option {
	return func(c *evalConfig) { c.tenantID = tenantID }
}

// ClientOption configura o Client na criação.
type ClientOption func(*Client)

// WithHTTPClient substitui o http.Client padrão.
func WithHTTPClient(c *http.Client) ClientOption {
	return func(client *Client) { client.httpClient = c }
}

// WithTimeout define o timeout das requisições HTTP.
func WithTimeout(d time.Duration) ClientOption {
	return func(client *Client) { client.httpClient.Timeout = d }
}

// WithCacheTTL ativa cache local de todas as flags com o TTL informado.
// Recomendado para hot paths — evita uma chamada HTTP por flag check.
func WithCacheTTL(d time.Duration) ClientOption {
	return func(client *Client) { client.cacheTTL = d }
}

// New cria um Client com timeout de 3s e sem cache.
// baseURL: URL base da API (ex: "https://pulse-api.institutoitinerante.com.br").
// token: service token da conta do serviço registrado em service_clients.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		token:        token,
		httpClient:   &http.Client{Timeout: 3 * time.Second},
		flagCache:    make(map[string]bool),
		variantCache: make(map[string]string),
		cacheTTL:     0,
	}
}

// NewWithOptions cria um Client com opções adicionais.
func NewWithOptions(baseURL, token string, opts ...ClientOption) *Client {
	c := New(baseURL, token)
	for _, o := range opts {
		o(c)
	}
	return c
}

// IsEnabled retorna se a flag está habilitada.
// Fail-safe: retorna defaultVal em caso de erro de rede, timeout ou flag inexistente.
func (c *Client) IsEnabled(ctx context.Context, flagKey string, defaultVal bool, opts ...Option) bool {
	if !validKeyPattern.MatchString(flagKey) {
		return defaultVal
	}
	cfg := applyOpts(opts)

	if c.cacheTTL > 0 {
		if v, ok := c.fromFlagCache(flagKey); ok {
			return v
		}
		// Cache miss: recarrega todas as flags de uma vez
		flags, variants, err := c.fetchAllFlags(ctx, cfg.userID, cfg.tenantID)
		if err != nil {
			return defaultVal
		}
		c.setAllCache(flags, variants)
		return flags[flagKey]
	}

	// Sem cache: avalia flag individual via endpoint admin (requer token)
	enabled, err := c.evaluateSingle(ctx, flagKey)
	if err != nil {
		return defaultVal
	}
	return enabled
}

// GetVariant retorna a variante A/B para uma flag.
// Retorna defaultValue se a flag estiver desabilitada, não existir ou Pulse estiver indisponível.
func (c *Client) GetVariant(ctx context.Context, flagKey, defaultValue string, opts ...Option) string {
	if !validKeyPattern.MatchString(flagKey) {
		return defaultValue
	}
	cfg := applyOpts(opts)

	if c.cacheTTL > 0 {
		if v, ok := c.fromVariantCache(flagKey); ok {
			return v
		}
	}

	variants, err := c.fetchVariants(ctx, cfg.userID, cfg.tenantID)
	if err != nil {
		return defaultValue
	}
	if c.cacheTTL > 0 {
		c.mu.Lock()
		c.variantCache = variants
		c.mu.Unlock()
	}
	if v, ok := variants[flagKey]; ok {
		return v
	}
	return defaultValue
}

// Track envia um evento de tracking para o Pulse (fire-and-forget).
// Erros são silenciados — não bloqueia o caller.
func (c *Client) Track(eventName string, props Props) {
	go func() {
		payload := map[string]interface{}{
			"event_name": eventName,
			"properties": props,
		}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/api/v1/events/track", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if c.token != "" {
			req.Header.Set("X-Service-Token", c.token)
		}
		resp, err := c.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()
}

// Identify associa um ID anônimo a um user ID real.
func (c *Client) Identify(anonymousID, userID string) {
	go func() {
		payload := map[string]string{"anonymous_id": anonymousID, "user_id": userID}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/api/v1/events/identify", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()
}

// evaluateSingle avalia uma flag individual via GET /api/v1/flags/:key/evaluate.
func (c *Client) evaluateSingle(ctx context.Context, flagKey string) (bool, error) {
	endpoint := fmt.Sprintf("%s/api/v1/flags/%s/evaluate", c.baseURL, url.PathEscape(flagKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Service-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("pulse: status %d", resp.StatusCode)
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, err
	}
	return body.Enabled, nil
}

// fetchAllFlags busca todas as flags via GET /flags (com contexto opcional de user/tenant).
func (c *Client) fetchAllFlags(ctx context.Context, userID, tenantID string) (map[string]bool, map[string]string, error) {
	u := c.baseURL + "/flags"
	params := url.Values{}
	if userID != "" {
		params.Set("user_id", userID)
	}
	if tenantID != "" {
		params.Set("tenant_id", tenantID)
	}
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-Service-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("pulse: status %d", resp.StatusCode)
	}

	var result struct {
		Flags    map[string]bool   `json:"flags"`
		Variants map[string]string `json:"variants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, err
	}
	if result.Flags == nil {
		result.Flags = make(map[string]bool)
	}
	if result.Variants == nil {
		result.Variants = make(map[string]string)
	}
	return result.Flags, result.Variants, nil
}

// fetchVariants busca variantes via GET /flags/variants.
func (c *Client) fetchVariants(ctx context.Context, userID, tenantID string) (map[string]string, error) {
	u := c.baseURL + "/flags/variants"
	params := url.Values{}
	if userID != "" {
		params.Set("user_id", userID)
	}
	if tenantID != "" {
		params.Set("tenant_id", tenantID)
	}
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return make(map[string]string), nil
	}
	req.Header.Set("X-Service-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return make(map[string]string), nil // fail-open
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return make(map[string]string), nil
	}

	var result struct {
		Variants map[string]string `json:"variants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return make(map[string]string), nil
	}
	if result.Variants == nil {
		result.Variants = make(map[string]string)
	}
	return result.Variants, nil
}

func (c *Client) fromFlagCache(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cacheTTL > 0 && time.Since(c.cacheTime) > c.cacheTTL {
		return false, false
	}
	v, ok := c.flagCache[key]
	return v, ok
}

func (c *Client) fromVariantCache(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cacheTTL > 0 && time.Since(c.cacheTime) > c.cacheTTL {
		return "", false
	}
	v, ok := c.variantCache[key]
	return v, ok
}

func (c *Client) setAllCache(flags map[string]bool, variants map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flagCache = flags
	c.variantCache = variants
	c.cacheTime = time.Now()
}

func applyOpts(opts []Option) *evalConfig {
	cfg := &evalConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}
