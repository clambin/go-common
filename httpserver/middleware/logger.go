package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	return middleware.RequestLogger(logger{logger: l})
}

type logger struct {
	logger *slog.Logger
}

func (l logger) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &LogEntry{logger: l.logger, request: r}
}

var _ middleware.LogEntry = &LogEntry{}

type LogEntry struct {
	logger  *slog.Logger
	request *http.Request
}

func (l LogEntry) Write(status, _ int, _ http.Header, elapsed time.Duration, _ interface{}) {
	l.logger.Info("request processed",
		slog.Group("request",
			slog.String("from", l.request.RemoteAddr),
			slog.String("path", l.request.URL.Path),
			slog.String("method", l.request.Method),
			slog.Int("status", status),
			//slog.Int("responseSize", bytes),
			slog.Float64("elapsed", float64(elapsed.Microseconds())/1000),
		),
	)
}

func (l LogEntry) Panic(v interface{}, _ []byte) {
	middleware.PrintPrettyStack(v)
}
