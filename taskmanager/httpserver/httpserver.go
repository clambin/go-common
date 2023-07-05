// Package httpserver implements a Task that runs an HTTP server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// HTTPServer runs a Task that starts an HTTP server.
type HTTPServer struct {
	ShutdownTimeout time.Duration
	server          *http.Server
}

// New returns an HTTPServer for the provided listener address and router.
func New(addr string, router http.Handler) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		ShutdownTimeout: 5 * time.Second,
	}
}

// Run starts the HTTP server.
//
// When the context is marked as Done, it will attempt to shut down the HTTP server gracefully,
// waiting ShutdownTimeout duration for it to shut down. The default for ShutdownTimeout is 5 seconds.
func (s *HTTPServer) Run(ctx context.Context) error {
	ch := make(chan error)
	go func() {
		ch <- s.server.ListenAndServe()
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
	}

	ctx2, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()
	if err := s.server.Shutdown(ctx2); err != nil {
		<-ch
		return fmt.Errorf("http shutdown: %w", err)
	}

	err := <-ch
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return err
}
