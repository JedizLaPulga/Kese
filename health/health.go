package health

import (
	"net/http"
	"sync"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
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

// Handler returns an HTTP handler for health checks.
//
// Example:
//
//	app.GET("/health", healthChecker.Handler())
func (h *HealthChecker) Handler() kese.HandlerFunc {
	return func(c *context.Context) error {
		status, checks := h.Check()

		response := map[string]interface{}{
			"status": status,
			"checks": checks,
		}

		statusCode := http.StatusOK
		if status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, response)
	}
}

// LivenessHandler returns a simple liveness check (always returns 200).
// Useful for Kubernetes liveness probes.
func (h *HealthChecker) LivenessHandler() kese.HandlerFunc {
	return func(c *context.Context) error {
		return c.JSON(200, map[string]string{"status": "alive"})
	}
}

// ReadinessHandler returns a readiness check (checks all health checks).
// Useful for Kubernetes readiness probes.
func (h *HealthChecker) ReadinessHandler() kese.HandlerFunc {
	return h.Handler()
}

// Default global health checker
var defaultChecker = New()

// AddCheck adds a check to the default health checker.
func AddCheck(name string, check CheckFunc) {
	defaultChecker.AddCheck(name, check)
}

// Handler returns the default health check handler.
func Handler() kese.HandlerFunc {
	return defaultChecker.Handler()
}
