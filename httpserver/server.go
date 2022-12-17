package httpserver

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/set"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"net/http"
	"time"
)

// Server implements a configurable HTTP Server. See the different WithXXX structs for available options.
type Server struct {
	port                int
	handlers            []Handler
	instrumentedHandler *InstrumentedHandler
	listener            net.Listener
	server              *http.Server
}

var _ prometheus.Collector = &Server{}

// Handler contains a path to be registered in the Server's HTTP server
type Handler struct {
	// Path of the endpoint (e.g. "/health"). Can be any path that's valid for gorilla/mux router's Path() method.
	Path string
	// Methods that the handler should support. If empty, defaults to http.MethodGet.
	Methods []string
	// Handler that implements the endpoint.
	Handler http.Handler
}

// New returns a Server with the specified options
func New(options ...Option) (s *Server, err error) {
	s = new(Server)
	for _, o := range options {
		o.apply(s)
	}

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	r := http.NewServeMux()
	for _, h := range s.handlers {
		r.Handle(h.Path, s.makeHandler(h))
	}
	s.server = &http.Server{Handler: r}
	return
}

// Serve starts the HTTP server. When the server is shut down, it returns http.ErrServerClosed.
func (s *Server) Serve() error {
	return s.server.Serve(s.listener)
}

// Shutdown performs a graceful shutdown of the HTTP server
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetPort returns the HTTP Server's listening port
func (s *Server) GetPort() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// ServeHTTP calls the server's handler. Mainly intended to be used in unit tests without starting the underlying HTTP server.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.server.Handler.ServeHTTP(w, req)
}

// Describe implements the prometheus.Collector interface
func (s *Server) Describe(descs chan<- *prometheus.Desc) {
	if s.instrumentedHandler != nil {
		s.instrumentedHandler.metrics.Describe(descs)
	}
}

// Collect implements the prometheus.Collector interface
func (s *Server) Collect(c chan<- prometheus.Metric) {
	if s.instrumentedHandler != nil {
		s.instrumentedHandler.metrics.Collect(c)
	}
}

func (s *Server) makeHandler(h Handler) http.Handler {
	handler := h.Handler
	if s.instrumentedHandler != nil {
		handler = s.instrumentedHandler.handle(handler)
	}
	if len(h.Methods) == 0 {
		h.Methods = []string{http.MethodGet}
	}
	m := &MethodFilteredHandler{methods: set.Create(h.Methods)}
	return m.handle(handler)
}
