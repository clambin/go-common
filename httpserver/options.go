package httpserver

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Option specified configuration options for Server
type Option interface {
	apply(server *Server)
}

// WithPort specifies the Server's listening port. If no port is specified, Server will listen on a random port.
// Use GetPort() to determine the actual listening port
type WithPort struct {
	Port int
}

func (o WithPort) apply(s *Server) {
	s.port = o.Port
}

// WithPrometheus adds a Prometheus metrics endpoint to the server at the specified Path. Default path is "/metrics"
type WithPrometheus struct {
	Path string
}

func (o WithPrometheus) apply(s *Server) {
	if o.Path == "" {
		o.Path = "/metrics"
	}
	s.handlers = append(s.handlers, Handler{
		Path:    o.Path,
		Handler: promhttp.Handler(),
		Methods: []string{http.MethodGet},
	})
}

// WithHandlers adds the specified handlers to the server
type WithHandlers struct {
	Handlers []Handler
}

func (o WithHandlers) apply(s *Server) {
	s.handlers = append(s.handlers, o.Handlers...)
}

// WithMetrics will collect the specified metrics to instrument the Server's Handlers.
type WithMetrics struct {
	Namespace   string
	Subsystem   string
	Application string
	MetricsType MetricsType
	Buckets     []float64
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

func (o WithMetrics) apply(s *Server) {
	// TODO: this is ugly is f*ck
	mwOptions := middleware.PrometheusMetricsOptions{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Application: o.Application,
		Buckets:     o.Buckets,
	}
	switch o.MetricsType {
	case Summary:
		mwOptions.MetricsType = middleware.Summary
	case Histogram:
		mwOptions.MetricsType = middleware.Histogram
	}
	s.instrumentedHandler = middleware.NewPrometheusMetrics(mwOptions)
}
