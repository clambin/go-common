// Package prometheus provides a Task that runs an HTTP server for the Prometheus default registry.
package prometheus

import (
	"github.com/clambin/go-common/taskmanager/httpserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// New creates a new httpserver.HTTPServer for the Prometheus default registry.
func New(options ...Option) *httpserver.HTTPServer {
	cfg := config{
		address: ":9090",
		path:    "/metrics",
		handler: promhttp.Handler(),
	}
	for _, option := range options {
		option(&cfg)
	}

	r := http.NewServeMux()
	r.Handle(cfg.path, cfg.handler)
	return httpserver.New(cfg.address, r)
}

// Option configures the HTTPServer
type Option func(cfg *config)

type config struct {
	address string
	path    string
	handler http.Handler
}

// WithAddr overrides the default listener address, i.e. ":9090"
func WithAddr(address string) Option {
	return func(cfg *config) {
		cfg.address = address
	}
}

// WithPath overrides the default path, i.e. "/metrics"
func WithPath(path string) Option {
	return func(cfg *config) {
		cfg.path = path
	}
}
