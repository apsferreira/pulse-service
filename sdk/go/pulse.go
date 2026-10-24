// Package pulse provides a Go SDK for the Pulse feature flag service.
//
// Usage:
//
//	client := pulse.New("https://seu-pulse.exemplo.com", "service-token")
//	enabled := client.IsEnabled(ctx, "minha-flag", true)
//	variant := client.GetVariant(ctx, "pricing-test", "control", pulse.WithUser("user-123"))
//	client.Track(ctx, "page.viewed", "meu-produto", userID, pulse.Props{"page": "/home"})
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
// baseURL: URL base da API (ex: "https://seu-pulse.exemplo.com").
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
//
// product é obrigatório: identifica qual produto do ecossistema IIT originou
// o evento. O pulse-service usa esse campo para segmentar eventos por
// produto — se vier vazio, o EventService normaliza para "product=default",
// misturando a telemetria de origens diferentes. Antes de 2026-09-11 esta
// função não enviava product nem user_id de forma alguma (só event_name e
// properties), o que levou 13 serviços do ecossistema a escreverem cliente
// HTTP próprio em vez de importar este SDK — ver ADR correspondente.
//
// userID pode ser vazio quando o evento não está associado a um usuário
// autenticado (ex.: evento anônimo, ainda não identificado).
func (c *Client) Track(ctx context.Context, eventName, product, userID string, props Props) {
	go func() {
		payload := map[string]interface{}{
			"event_name": eventName,
			"product":    product,
			"user_id":    userID,
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

// TrackWithProject envia um evento de tracking associado a um projectID — o UUID
// estável do projeto/tenant dono do evento (BKL-1038), ex: o UUID da loja no
// StoreMake. Sem projectID, o evento não pode ser consultado depois via
// QueryEvents/POST /api/v1/projects/:id/query (filtra por project_id::uuid) nem
// isolado no canal de broadcast. Fire-and-forget, mesmo comportamento de Track.
//
// Adicionado para o funil de atribuição do SocialMake (ADR-052, Opção B): o
// endpoint simplificado /events/track nunca populava event.ProjectID antes desta
// versão — era aceito só por POST /api/v1/events (ingest completo).
func (c *Client) TrackWithProject(ctx context.Context, eventName, product, userID, projectID string, props Props) {
	go func() {
		payload := map[string]interface{}{
			"event_name": eventName,
			"product":    product,
			"user_id":    userID,
			"properties": props,
		}
		if projectID != "" {
			payload["project_id"] = projectID
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

// QueryFilter é um filtro de igualdade/comparação sobre uma propriedade do
// evento (properties JSONB). Property aceita qualquer nome alfanumérico — o
// endpoint de query não restringe por allowlist nos filtros (só no GroupBy).
type QueryFilter struct {
	Property string // ex: "attribution_ref" — chave dentro de properties
	Operator string // eq | neq | gt | gte | lt | lte | contains | starts_with
	Value    string
}

// QueryRequest é o payload de QueryEvents (POST /api/v1/projects/:id/query).
type QueryRequest struct {
	// ProjectID é obrigatório — o UUID do projeto/tenant dono dos eventos
	// (mesmo valor passado em TrackWithProject).
	ProjectID string
	// Event é o event_type a filtrar (obrigatório), ex: "order.paid".
	Event   string
	Filters []QueryFilter
	// Aggregate: count (default) | sum | avg | distinct_count.
	Aggregate string
	// DateFrom/DateTo no formato YYYY-MM-DD. Zero value = sem filtro de data.
	DateFrom string
	DateTo   string
}

// QueryResultRow é uma linha do resultado de QueryEvents.
type QueryResultRow struct {
	Group string
	Value float64
}

// QueryEventsResponse é a resposta de QueryEvents.
type QueryEventsResponse struct {
	Results []QueryResultRow
	Count   int
}

// ErrQueryUnavailable é o sentinel devolvido quando o pulse-service está
// indisponível ou retorna erro — convenção STD-017: o chamador trata via
// errors.Is, nunca inspecionando string de erro ou status code.
var ErrQueryUnavailable = fmt.Errorf("pulse: query indisponível")

// QueryEvents consulta eventos filtrados por projeto e propriedades via
// POST /api/v1/projects/:id/query (query_explorer_handler.go no pulse-service).
// Diferente de um funil (breakdown_by), este endpoint filtra por um valor EXATO
// de propriedade sem limite de cardinalidade — é o mecanismo certo para
// perguntas do tipo "quantos order.paid têm attribution_ref = X", uma por post,
// em vez de pedir ao Pulse para enumerar os N valores mais frequentes (que tem
// um LIMIT 10 embutido em funnels/analyze e descartaria posts fora do top 10).
//
// Síncrono (ao contrário de Track/Identify) porque o chamador precisa do
// resultado para responder a uma página de Analytics. Timeout do httpClient do
// Client (default 3s) se aplica — ajuste via WithTimeout se precisar de mais
// para agregações pesadas.
func (c *Client) QueryEvents(ctx context.Context, req QueryRequest) (*QueryEventsResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("pulse: QueryEvents requer ProjectID")
	}
	if req.Event == "" {
		return nil, fmt.Errorf("pulse: QueryEvents requer Event")
	}

	type filterDTO struct {
		Property string `json:"property"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}
	filters := make([]filterDTO, 0, len(req.Filters))
	for _, f := range req.Filters {
		filters = append(filters, filterDTO{Property: f.Property, Operator: f.Operator, Value: f.Value})
	}

	body, err := json.Marshal(map[string]interface{}{
		"event":     req.Event,
		"filters":   filters,
		"aggregate": req.Aggregate,
		"date_range": map[string]string{
			"from": req.DateFrom,
			"to":   req.DateTo,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("pulse: erro ao montar query: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/projects/%s/query", c.baseURL, url.PathEscape(req.ProjectID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryUnavailable, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("X-Service-Token", c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrQueryUnavailable, resp.StatusCode)
	}

	var raw struct {
		Results []struct {
			Group string  `json:"group"`
			Value float64 `json:"value"`
		} `json:"results"`
		Count int `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryUnavailable, err)
	}

	out := &QueryEventsResponse{Count: raw.Count, Results: make([]QueryResultRow, 0, len(raw.Results))}
	for _, r := range raw.Results {
		out.Results = append(out.Results, QueryResultRow{Group: r.Group, Value: r.Value})
	}
	return out, nil
}

// Identify associa um ID anônimo a um user ID real, marcado com o produto de
// origem (ver IdentifyRequest.Product no pulse-service).
func (c *Client) Identify(product, anonymousID, userID string) {
	go func() {
		payload := map[string]string{"product": product, "anonymous_id": anonymousID, "user_id": userID}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/api/v1/events/identify", bytes.NewReader(body))
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
