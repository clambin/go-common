package middleware_test

import (
	"bytes"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	out := bytes.NewBufferString("")
	opt := slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		// Remove time from the output for predictable test output.
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewTextHandler(out, &opt))

	r := chi.NewRouter()
	r.Use(middleware.Logger(l))
	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, out.String(), `level=INFO msg="request processed" request.from="" request.path=/ping request.method=GET request.status=200 request.elapsed=`)
}
