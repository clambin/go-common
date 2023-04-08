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
func WithMetrics(metrics middleware.PrometheusMetricsOptions) Option {
	return func(s *Server) {
		s.instrumentedHandler = middleware.NewPrometheusMetrics(metrics)
	}
}
