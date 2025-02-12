package middleware_test

import (
	"bytes"
	"github.com/clambin/go-common/httputils/middleware"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  slog.Level
		logger middleware.RequestLogFormatter
		want   string
	}{
		{
			name:   "default",
			logger: middleware.DefaultRequestLogFormatter,
			want:   `level=INFO msg="http request" path=/ method=GET code=200 latency=`,
		},
		{
			name:   "none",
			level:  slog.LevelDebug,
			logger: middleware.DefaultRequestLogFormatter,
		},
		{
			name: "custom",
			logger: middleware.RequestLogFormatterFunc(func(r *http.Request, code int, latency time.Duration) []slog.Attr {
				return []slog.Attr{
					slog.String("client", r.RemoteAddr),
					slog.String("path", r.URL.Path), slog.String("method", r.Method),
					slog.Int("code", code), slog.Duration("latency", latency),
				}
			}),
			want: `level=INFO msg="http request" client=127.0.0.1:5000 path=/ method=GET code=200 latency=`,
		},
	}

	var out bytes.Buffer
	l := slog.New(slog.NewTextHandler(&out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()

			h := middleware.RequestLogger(l, tt.level, tt.logger)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("hello"))
			}))

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:5000"
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
			}
			if got := strings.Contains(out.String(), tt.want); got != true {
				t.Errorf("got response %s, want %s", out.String(), tt.want)
			}
		})
	}
}
