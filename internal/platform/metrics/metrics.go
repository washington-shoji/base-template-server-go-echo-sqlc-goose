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
	AppRegistry = prometheus.NewRegistry()

	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestSizeBytes  *prometheus.SummaryVec
	httpResponseSizeBytes *prometheus.SummaryVec
	jobsProcessedTotal    *prometheus.CounterVec

	registerOnce sync.Once
)

func MustRegister() {
	registerOnce.Do(func() {
		reg := promauto.With(AppRegistry)
		httpRequestsTotal = reg.NewCounterVec(prometheus.CounterOpts{
			Namespace: "app", Name: "http_requests_total", Help: "Total HTTP requests.",
		}, []string{"method", "path", "status_code"})
		httpRequestDuration = reg.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "app", Name: "http_request_duration_seconds", Help: "HTTP latencies.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"})
		httpRequestSizeBytes = reg.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: "app", Name: "http_request_size_bytes", Help: "Request sizes.",
		}, []string{"method", "path"})
		httpResponseSizeBytes = reg.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: "app", Name: "http_response_size_bytes", Help: "Response sizes.",
		}, []string{"method", "path"})
		jobsProcessedTotal = reg.NewCounterVec(prometheus.CounterOpts{
			Namespace: "app", Name: "jobs_processed_total", Help: "Background jobs processed.",
		}, []string{"type", "status"})
		AppRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		AppRegistry.MustRegister(collectors.NewGoCollector())
	})
}

func JobProcessed(jobType, status string) {
	MustRegister()
	jobsProcessedTotal.WithLabelValues(jobType, status).Inc()
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	MustRegister()
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()
		path := c.Path()
		if path == "" {
			path = "unmatched"
		}

		err := next(c)

		status := c.Response().Status
		if status == 0 {
			if err != nil {
				status = 500
			} else {
				status = 200
			}
		}
		elapsed := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(req.Method, path, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(req.Method, path).Observe(elapsed)
		if req.ContentLength > 0 {
			httpRequestSizeBytes.WithLabelValues(req.Method, path).Observe(float64(req.ContentLength))
		}
		if c.Response().Size > 0 {
			httpResponseSizeBytes.WithLabelValues(req.Method, path).Observe(float64(c.Response().Size))
		}
		return err
	}
}
