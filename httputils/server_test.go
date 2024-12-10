package http

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRunServer(t *testing.T) {
	s := http.Server{Addr: ":8888", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- RunServer(ctx, &s)
	}()

	for {
		_, err := http.Get("http://localhost:8888/")
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	cancel()

	for {
		_, err := http.Get("http://localhost:8888/")
		if err != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("server failed to start/stop: %s", err.Error())
	}
}
