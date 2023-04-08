package httpserver

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Option specified configuration options for Server
type Option func(*Server)

// WithAddr specifies the Server's listening address.  If the port is zero, Server will listen on a random port.
// Use GetPort() to determine the actual listening port
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithPrometheus adds a Prometheus metrics endpoint to the server at the specified Path. Default path is "/metrics"
func WithPrometheus(path string) Option {
	return func(s *Server) {
		if path == "" {
			path = "/metrics"
		}
		s.handlers = append(s.handlers, Handler{
			Path:    path,
			Handler: promhttp.Handler(),
			Methods: []string{http.MethodGet},
		})
	}
}

// WithHandlers adds the specified handlers to the server
func WithHandlers(handlers []Handler) Option {
	return func(s *Server) {
		s.handlers = append(s.handlers, handlers...)
	}
}

// WithMetrics will collect the specified metrics to instrument the Server's Handlers.
func WithMetrics(namespace, subsystem, application string, metricsType MetricsType, buckets []float64) Option {
	return func(s *Server) {
		// TODO: this is ugly is f*ck
		mwOptions := middleware.PrometheusMetricsOptions{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Application: application,
			Buckets:     buckets,
		}
		switch metricsType {
		case Summary:
			mwOptions.MetricsType = middleware.Summary
		case Histogram:
			mwOptions.MetricsType = middleware.Histogram
		}
		s.instrumentedHandler = middleware.NewPrometheusMetrics(mwOptions)
	}
}

// MetricsType specifies the type of metrics to record for request duration. Use Summary if you are only interested in the average latency.
// Use Histogram if you want to use a histogram to measure a service level indicator (eg  latency of 95% of all requests).
type MetricsType int

const (
	// Summary measures the average duration.
	Summary MetricsType = iota
	// Histogram measures the latency in buckets and can be used to calculate a service level indicator. WithMetrics.Buckets
	// specify the buckets to be used. If none are provided, prometheus.DefBuckets will be used.
	Histogram
)
