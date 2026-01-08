package health

import (
	"net/http"
	"sync"
)

// CheckFunc is a function that performs a health check.
// Return nil if healthy, error otherwise.
type CheckFunc func() error

// Status represents the health status.
type Status string

const (
	// StatusHealthy indicates all checks passed
	StatusHealthy Status = "healthy"
	// StatusUnhealthy indicates at least one check failed
	StatusUnhealthy Status = "unhealthy"
)

// HealthChecker manages health checks.
type HealthChecker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// New creates a new health checker.
func New() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
	}
}

// AddCheck adds a named health check.
//
// Example:
//
//	health.AddCheck("database", func() error {
//	    return db.Ping()
//	})
func (h *HealthChecker) AddCheck(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// RemoveCheck removes a health check.
func (h *HealthChecker) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// Check runs all health checks and returns the status.
func (h *HealthChecker) Check() (Status, map[string]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]string)
	allHealthy := true

	for name, check := range h.checks {
		if err := check(); err != nil {
			results[name] = err.Error()
			allHealthy = false
		} else {
			results[name] = "ok"
		}
	}

	if allHealthy {
		return StatusHealthy, results
	}
	return StatusUnhealthy, results
}

// ServeHTTP implements http.Handler for the health checker.
func (h *HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, checks := h.Check()

	statusCode := http.StatusOK
	if status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Simple JSON response without importing encoding/json
	w.Write([]byte(`{"status":"`))
	w.Write([]byte(status))
	w.Write([]byte(`","checks":{`))

	first := true
	for name, result := range checks {
		if !first {
			w.Write([]byte(`,`))
		}
		w.Write([]byte(`"`))
		w.Write([]byte(name))
		w.Write([]byte(`":"`))
		w.Write([]byte(result))
		w.Write([]byte(`"`))
		first = false
	}

	w.Write([]byte(`}}`))
}

// Default global health checker
var defaultChecker = New()

// AddCheck adds a check to the default health checker.
func AddCheck(name string, check CheckFunc) {
	defaultChecker.AddCheck(name, check)
}
