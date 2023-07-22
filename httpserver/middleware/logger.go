package middleware

import (
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

// RequestLogger logs incoming HTTP requests.
func RequestLogger(logger *slog.Logger, logLevel slog.Level, formatter RequestLogFormatter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w}

			next.ServeHTTP(lrw, r)

			logger.LogAttrs(r.Context(), logLevel, "http request", formatter.FormatRequest(r, lrw.statusCode, time.Since(start))...)
		}

		return http.HandlerFunc(fn)
	}
}

// A RequestLogFormatter takes the HTTP request, the resulting HTTP status code and latency and formats the log entry.
type RequestLogFormatter interface {
	FormatRequest(*http.Request, int, time.Duration) []slog.Attr
}

// DefaultRequestLogFormatter is the default RequestLogFormatter. It logs the request's HTTP method and the path.
var DefaultRequestLogFormatter RequestLogFormatter = &defaultRequestLogFormatter{}

type defaultRequestLogFormatter struct{}

func (d defaultRequestLogFormatter) FormatRequest(r *http.Request, statusCode int, latency time.Duration) []slog.Attr {
	return []slog.Attr{slog.String("path", r.URL.Path), slog.String("method", r.Method),
		slog.Int("code", statusCode), slog.Duration("latency", latency),
	}
}

// The RequestLogFormatterFunc type is an adapter that allows an ordinary function to be used as a RequestLogFormatter.
type RequestLogFormatterFunc func(r *http.Request, statusCode int, latency time.Duration) []slog.Attr

// FormatRequest calls f(r, statusCode, latency)
func (f RequestLogFormatterFunc) FormatRequest(r *http.Request, statusCode int, latency time.Duration) []slog.Attr {
	return f(r, statusCode, latency)
}
