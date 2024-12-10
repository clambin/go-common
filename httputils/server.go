package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// RunServer starts an HTTP Server and gracefully shuts it down when the context expires.
//
// Once the server has been shut down, it may not be reused.
func RunServer(ctx context.Context, s *http.Server) error {
	subCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error)
	go func() {
		<-subCtx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		err := s.Shutdown(shutdownCtx)
		if err != nil {
			err = fmt.Errorf("http server failed to stop: %w", err)
		}
		errCh <- err
	}()

	err := s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("http server failed to start: %w", err)
	}
	cancel()
	return errors.Join(err, <-errCh)
}
