package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metrics holds application metrics.
type Metrics struct {
	mu              sync.RWMutex
	requestCount    map[string]int
	requestDuration map[string][]time.Duration
	activeRequests  int
	totalRequests   int
	totalErrors     int
}

// New creates a new metrics collector.
func New() *Metrics {
	return &Metrics{
		requestCount:    make(map[string]int),
		requestDuration: make(map[string][]time.Duration),
	}
}

// RecordRequest records a completed request.
func (m *Metrics) RecordRequest(method, path string, duration time.Duration, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := method + " " + path
	m.requestCount[key]++
	m.requestDuration[key] = append(m.requestDuration[key], duration)
	m.totalRequests++

	if statusCode >= 400 {
		m.totalErrors++
	}
}

// IncrementActive increments active request count.
func (m *Metrics) IncrementActive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeRequests++
}

// DecrementActive decrements active request count.
func (m *Metrics) DecrementActive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeRequests--
}

// ServeHTTP implements http.Handler for Prometheus-style metrics endpoint.
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Active requests
	fmt.Fprintf(w, "# HELP kese_active_requests Number of active requests\n")
	fmt.Fprintf(w, "# TYPE kese_active_requests gauge\n")
	fmt.Fprintf(w, "kese_active_requests %d\n\n", m.activeRequests)

	// Total requests
	fmt.Fprintf(w, "# HELP kese_requests_total Total number of requests\n")
	fmt.Fprintf(w, "# TYPE kese_requests_total counter\n")
	fmt.Fprintf(w, "kese_requests_total %d\n\n", m.totalRequests)

	// Total errors
	fmt.Fprintf(w, "# HELP kese_errors_total Total number of errors (4xx, 5xx)\n")
	fmt.Fprintf(w, "# TYPE kese_errors_total counter\n")
	fmt.Fprintf(w, "kese_errors_total %d\n\n", m.totalErrors)

	// Request count by route
	fmt.Fprintf(w, "# HELP kese_requests_by_route_total Requests by route\n")
	fmt.Fprintf(w, "# TYPE kese_requests_by_route_total counter\n")
	for route, count := range m.requestCount {
		fmt.Fprintf(w, "kese_requests_by_route_total{route=\"%s\"} %d\n", route, count)
	}
	fmt.Fprintln(w)

	// Average duration by route
	fmt.Fprintf(w, "# HELP kese_request_duration_seconds Average request duration\n")
	fmt.Fprintf(w, "# TYPE kese_request_duration_seconds summary\n")
	for route, durations := range m.requestDuration {
		if len(durations) > 0 {
			var total time.Duration
			for _, d := range durations {
				total += d
			}
			avg := total / time.Duration(len(durations))
			fmt.Fprintf(w, "kese_request_duration_seconds{route=\"%s\"} %.6f\n",
				route, avg.Seconds())
		}
	}
}

// Default global metrics
var defaultMetrics = New()

// RecordRequest records to the default metrics.
func RecordRequest(method, path string, duration time.Duration, statusCode int) {
	defaultMetrics.RecordRequest(method, path, duration, statusCode)
}

// Handler returns the default metrics HTTP handler.
func Handler() http.Handler {
	return defaultMetrics
}
