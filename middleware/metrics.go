package middleware

import (
	"time"

	"github.com/JedizLaPulga/kese"
	"github.com/JedizLaPulga/kese/context"
	"github.com/JedizLaPulga/kese/metrics"
)

// MetricsConfig holds configuration for metrics middleware.
type MetricsConfig struct {
	// Metrics is the metrics collector
	Metrics *metrics.Metrics

	// SkipFunc allows skipping metrics collection for certain requests
	SkipFunc func(*context.Context) bool
}

// DefaultMetricsConfig returns default metrics configuration.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Metrics:  metrics.New(),
		SkipFunc: nil,
	}
}

// Metrics returns a middleware that collects request metrics.
//
// Example:
//
//	app.Use(middleware.Metrics())
//
//	// Expose metrics endpoint
//	app.GET("/metrics", func(c *context.Context) error {
//	    metrics.Handler().ServeHTTP(c.Writer, c.Request)
//	    c.SetWritten()
//	    return nil
//	})
func Metrics() kese.MiddlewareFunc {
	return MetricsWithConfig(DefaultMetricsConfig())
}

// MetricsWithConfig returns metrics middleware with custom configuration.
func MetricsWithConfig(config MetricsConfig) kese.MiddlewareFunc {
	return func(next kese.HandlerFunc) kese.HandlerFunc {
		return func(c *context.Context) error {
			// Skip if configured
			if config.SkipFunc != nil && config.SkipFunc(c) {
				return next(c)
			}

			// Track active requests
			config.Metrics.IncrementActive()
			defer config.Metrics.DecrementActive()

			// Record start time
			start := time.Now()

			// Call next handler
			err := next(c)

			// Record metrics
			duration := time.Since(start)
			statusCode := c.StatusCode()
			if statusCode == 0 {
				statusCode = 200
			}

			config.Metrics.RecordRequest(c.Method(), c.Path(), duration, statusCode)

			return err
		}
	}
}
