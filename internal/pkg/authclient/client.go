package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// UserProfile é o subconjunto de campos que o Pulse precisa do auth-service.
type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Client faz lookups de perfil no auth-service via X-Service-Token.
// Resultados são cacheados em memória por cacheTTL para evitar burst de chamadas.
type Client struct {
	baseURL      string
	serviceToken string
	httpClient   *http.Client

	mu       sync.Mutex
	cache    map[string]*cacheEntry
	cacheTTL time.Duration
}

type cacheEntry struct {
	profile   *UserProfile
	fetchedAt time.Time
}

func New(baseURL, serviceToken string) *Client {
	return &Client{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient:   &http.Client{Timeout: 3 * time.Second},
		cache:        make(map[string]*cacheEntry),
		cacheTTL:     5 * time.Minute,
	}
}

// GetProfile busca nome+email de um user UUID no auth-service.
// Retorna nil sem erro quando o user não for encontrado (404) ou quando
// baseURL está vazia (enriquecimento desabilitado).
func (c *Client) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if c.baseURL == "" {
		return nil, nil
	}

	c.mu.Lock()
	if e, ok := c.cache[userID]; ok && time.Since(e.fetchedAt) < c.cacheTTL {
		c.mu.Unlock()
		return e.profile, nil
	}
	c.mu.Unlock()

	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Token", c.serviceToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[authclient] lookup %s: %v", userID, err)
		return nil, nil // fail-open: não bloqueia a lista
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.store(userID, nil)
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[authclient] lookup %s: status %d", userID, resp.StatusCode)
		return nil, nil
	}

	var p UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	c.store(userID, &p)
	return &p, nil
}

// GetProfilesBatch busca vários perfis em paralelo (máx 10 goroutines concorrentes).
// Retorna mapa user_id → *UserProfile; ausências ficam como nil.
func (c *Client) GetProfilesBatch(ctx context.Context, userIDs []string) map[string]*UserProfile {
	result := make(map[string]*UserProfile, len(userIDs))
	if c.baseURL == "" {
		return result
	}

	sem := make(chan struct{}, 10)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range userIDs {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			p, _ := c.GetProfile(ctx, id)
			mu.Lock()
			result[id] = p
			mu.Unlock()
		}()
	}
	wg.Wait()
	return result
}

func (c *Client) store(userID string, p *UserProfile) {
	c.mu.Lock()
	c.cache[userID] = &cacheEntry{profile: p, fetchedAt: time.Now()}
	c.mu.Unlock()
}
