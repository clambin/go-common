package middleware_test

import (
	"bytes"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	out := bytes.NewBufferString("")
	opt := slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		// Remove time from the output for predictable test output.
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewTextHandler(out, &opt))
	slog.SetDefault(l)

	testCases := []struct {
		name   string
		logger middleware.RequestLogger
		want   string
	}{
		{
			name:   "default",
			logger: middleware.DefaultRequestLogger{},
			want:   "level=INFO msg=request path=/ method=GET code=200 latency=",
		},
		{
			name: "custom",
			logger: middleware.RequestLoggerFunc(func(r *http.Request, code int, latency time.Duration) {
				l.Debug("request",
					slog.String("client", r.RemoteAddr),
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.Int("code", code),
					slog.Duration("latency", latency),
				)
			}),
			want: `level=DEBUG msg=request client=127.0.0.1:5000 path=/ method=GET code=200 latency=`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()

			r := http.NewServeMux()
			r.Handle("/", middleware.Logger(tt.logger)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				_, _ = writer.Write([]byte("hello"))
			})))

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:5000"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, out.String(), tt.want)

		})
	}
}

func TestDefaultLogger(t *testing.T) {
	out := bytes.NewBufferString("")
	opt := slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		// Remove time from the output for predictable test output.
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewTextHandler(out, &opt))
	slog.SetDefault(l)

	r := http.NewServeMux()
	r.Handle("/", middleware.Logger(middleware.DefaultRequestLogger{})(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello"))
	})))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, out.String(), `level=INFO msg=request path=/ method=GET code=200 latency=`)
}

func BenchmarkLogger(b *testing.B) {
	out := bytes.NewBufferString("")
	opt := slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		// Remove time from the output for predictable test output.
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}
	l := slog.New(slog.NewTextHandler(out, &opt))
	slog.SetDefault(l)

	r := http.NewServeMux()
	r.Handle("/", middleware.Logger(middleware.DefaultRequestLogger{})(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello"))
	})))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fail()
		}
	}
}
