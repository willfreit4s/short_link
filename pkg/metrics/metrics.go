package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{
			"method",
			"route",
			"status",
		},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{
			"method",
			"route",
		},
	)

	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsInFlight,
	)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		HTTPRequestsInFlight.Inc()

		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()

		if route == "" {
			route = "unknown"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(
			method,
			route,
			status,
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			method,
			route,
		).Observe(duration)

		HTTPRequestsInFlight.Dec()
	}
}
