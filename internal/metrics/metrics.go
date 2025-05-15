package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AppRegistry is a custom registry for application-specific metrics.
	AppRegistry = prometheus.NewRegistry()

	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestSizeBytes  *prometheus.SummaryVec
	httpResponseSizeBytes *prometheus.SummaryVec

	registerMetricsOnce sync.Once
)

// MustRegisterMetrics ensures that metrics are registered only once with the AppRegistry.
func MustRegisterMetrics() {
	registerMetricsOnce.Do(func() {
		reg := promauto.With(AppRegistry) // Use promauto with our custom registry

		httpRequestsTotal = reg.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "app",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status_code"},
		)

		httpRequestDuration = reg.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "app",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latencies in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		httpRequestSizeBytes = reg.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: "app",
				Name:      "http_request_size_bytes",
				Help:      "HTTP request sizes in bytes.",
			},
			[]string{"method", "path"},
		)

		httpResponseSizeBytes = reg.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: "app",
				Name:      "http_response_size_bytes",
				Help:      "HTTP response sizes in bytes.",
			},
			[]string{"method", "path"},
		)

		// Register standard Go collectors with the custom registry as well
		AppRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		AppRegistry.MustRegister(collectors.NewGoCollector())
	})
}

// Middleware returns an Echo middleware handler for recording HTTP metrics.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	MustRegisterMetrics() // Ensure metrics are registered before middleware is used
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()
		path := c.Path() // Get the matched path template

		// If c.Path() is empty (e.g. 404), use the request URI as a fallback.
		// Be cautious with high cardinality on request URIs if you have many unique ones.
		if path == "" {
			path = req.URL.Path
		}

		err := next(c) // Execute the actual handler

		status := strconv.Itoa(c.Response().Status)
		elapsed := time.Since(start).Seconds()
		requestSize := float64(req.ContentLength)
		responseSize := float64(c.Response().Size)

		httpRequestsTotal.WithLabelValues(req.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(req.Method, path).Observe(elapsed)

		if requestSize > 0 {
			httpRequestSizeBytes.WithLabelValues(req.Method, path).Observe(requestSize)
		}
		if responseSize > 0 {
			httpResponseSizeBytes.WithLabelValues(req.Method, path).Observe(responseSize)
		}

		return err
	}
}
