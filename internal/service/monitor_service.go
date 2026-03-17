package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type MonitorService struct {
	repo *repository.MonitorRepository
}

func NewMonitorService(repo *repository.MonitorRepository) *MonitorService {
	return &MonitorService{repo: repo}
}

// List returns all monitors with their last check and 24h stats.
func (s *MonitorService) List(ctx context.Context) ([]domain.MonitorStatus, error) {
	monitors, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("monitor_service: List: %w", err)
	}

	statuses := make([]domain.MonitorStatus, 0, len(monitors))
	for _, m := range monitors {
		ms, err := s.buildStatus(ctx, m)
		if err != nil {
			log.Printf("monitor_service: List build status for %s: %v", m.ID, err)
			// include with zeroed stats rather than failing the whole list
			statuses = append(statuses, domain.MonitorStatus{Monitor: m})
			continue
		}
		statuses = append(statuses, ms)
	}
	return statuses, nil
}

// GetByID returns a single monitor with stats.
func (s *MonitorService) GetByID(ctx context.Context, id string) (*domain.MonitorStatus, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("monitor_service: GetByID: %w", err)
	}
	if m == nil {
		return nil, nil
	}
	ms, err := s.buildStatus(ctx, *m)
	if err != nil {
		return nil, err
	}
	return &ms, nil
}

// Create validates and inserts a new monitor.
func (s *MonitorService) Create(ctx context.Context, req domain.CreateMonitorRequest) (*domain.Monitor, error) {
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return nil, fmt.Errorf("name must be between 2 and 100 characters")
	}
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	method := "GET"
	if req.Method != "" {
		method = req.Method
	}
	intervalSec := 60
	if req.IntervalSeconds != nil {
		intervalSec = *req.IntervalSeconds
	}
	timeoutSec := 10
	if req.TimeoutSeconds != nil {
		timeoutSec = *req.TimeoutSeconds
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	m := &domain.Monitor{
		Name:            req.Name,
		URL:             req.URL,
		Method:          method,
		IntervalSeconds: intervalSec,
		TimeoutSeconds:  timeoutSec,
		IsActive:        isActive,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("monitor_service: Create: %w", err)
	}
	return m, nil
}

// Update applies partial updates.
func (s *MonitorService) Update(ctx context.Context, id string, req domain.UpdateMonitorRequest) (*domain.Monitor, error) {
	m, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("monitor_service: Update: %w", err)
	}
	return m, nil
}

// Delete removes a monitor.
func (s *MonitorService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("monitor_service: Delete: %w", err)
	}
	return nil
}

// GetHistory returns the last 24h of checks for a given monitor.
func (s *MonitorService) GetHistory(ctx context.Context, id string) ([]domain.MonitorCheck, error) {
	checks, err := s.repo.GetHistoryLast24h(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("monitor_service: GetHistory: %w", err)
	}
	if checks == nil {
		checks = []domain.MonitorCheck{}
	}
	return checks, nil
}

// CheckMonitor performs an HTTP request against the monitor URL and persists the result.
func (s *MonitorService) CheckMonitor(ctx context.Context, m domain.Monitor) domain.MonitorCheck {
	check := domain.MonitorCheck{
		MonitorID: m.ID,
	}

	timeout := time.Duration(m.TimeoutSeconds) * time.Second
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequestWithContext(ctx, m.Method, m.URL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		check.Status = "down"
		check.ErrorMessage = &errMsg
		_ = s.repo.InsertCheck(ctx, &check)
		return check
	}
	req.Header.Set("User-Agent", "pulse-monitor/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := int(time.Since(start).Milliseconds())
	check.ResponseTimeMs = &elapsed

	if err != nil {
		errMsg := err.Error()
		// Distinguish timeout from generic down
		if ctx.Err() != nil || elapsed >= m.TimeoutSeconds*1000 {
			check.Status = "timeout"
		} else {
			check.Status = "down"
		}
		check.ErrorMessage = &errMsg
		_ = s.repo.InsertCheck(ctx, &check)
		return check
	}
	defer func() { _ = resp.Body.Close() }()
	// Drain body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	code := resp.StatusCode
	check.StatusCode = &code
	if code >= 200 && code < 400 {
		check.Status = "up"
	} else {
		check.Status = "down"
		errMsg := fmt.Sprintf("unexpected status code: %d", code)
		check.ErrorMessage = &errMsg
	}

	_ = s.repo.InsertCheck(ctx, &check)
	return check
}

// RunChecks fetches all active monitors and checks each one.
func (s *MonitorService) RunChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	monitors, err := s.repo.GetAllActive(ctx)
	if err != nil {
		log.Printf("monitor_service: RunChecks fetch monitors: %v", err)
		return
	}

	for _, m := range monitors {
		go func(mon domain.Monitor) {
			checkCtx, checkCancel := context.WithTimeout(context.Background(), time.Duration(mon.TimeoutSeconds+2)*time.Second)
			defer checkCancel()
			result := s.CheckMonitor(checkCtx, mon)
			log.Printf("monitor_service: checked %s (%s) -> %s", mon.Name, mon.URL, result.Status)
		}(m)
	}
}

// StartBackgroundChecker launches a goroutine that runs checks every 60 seconds.
func (s *MonitorService) StartBackgroundChecker() {
	go func() {
		log.Println("monitor_service: background checker started")
		// Run an initial check immediately at startup
		s.RunChecks()

		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			s.RunChecks()
		}
	}()
}

// buildStatus assembles a MonitorStatus by fetching last check and 24h stats.
func (s *MonitorService) buildStatus(ctx context.Context, m domain.Monitor) (domain.MonitorStatus, error) {
	lastCheck, err := s.repo.GetLastCheck(ctx, m.ID)
	if err != nil {
		return domain.MonitorStatus{}, fmt.Errorf("monitor_service: buildStatus last check: %w", err)
	}

	uptime, avgResp, err := s.repo.GetUptimeStats(ctx, m.ID)
	if err != nil {
		return domain.MonitorStatus{}, fmt.Errorf("monitor_service: buildStatus uptime stats: %w", err)
	}

	return domain.MonitorStatus{
		Monitor:       m,
		LastCheck:     lastCheck,
		Uptime24h:     uptime,
		AvgResponseMs: avgResp,
	}, nil
}
