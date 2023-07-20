package middleware

import (
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

// Logger logs incoming HTTP requests.  If requestWriter is nil, the logger defaults to DefaultRequestLogger.
func Logger(requestWriter RequestLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if requestWriter == nil {
			requestWriter = defaultRequestLogger{}
		}
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w}

			next.ServeHTTP(lrw, r)

			requestWriter.Log(r, lrw.statusCode, time.Since(start))
		}

		return http.HandlerFunc(fn)
	}
}

// A RequestLogger takes an incoming request, the resulting HTTP status code and the latency and logs it to a logger.
type RequestLogger interface {
	Log(r *http.Request, code int, latency time.Duration)
}

// DefaultRequestLogger logs the incoming request, the resulting HTTP status code and the latency to slog at Info log level.
var DefaultRequestLogger RequestLogger = defaultRequestLogger{}

type defaultRequestLogger struct{}

func (d defaultRequestLogger) Log(r *http.Request, statusCode int, latency time.Duration) {
	slog.Info("request",
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.Int("code", statusCode),
		slog.Duration("latency", latency),
	)
}
