package httpserver

import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHTTPServer_Run(t *testing.T) {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	s := New(":8080", r)
	m := taskmanager.New(s)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() {
		ch <- m.Run(ctx)
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		return err == nil && string(body) == "hello"
	}, time.Second, time.Millisecond)

	cancel()
	assert.NoError(t, <-ch)
}

func TestHTTPServer_Run_Fail(t *testing.T) {
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	s := New("not-a-valid-addr", r)
	m := taskmanager.New(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan error)
	go func() {
		ch <- m.Run(ctx)
	}()
	assert.Error(t, <-ch)
}
