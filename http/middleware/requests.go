package middleware

import (
	"github.com/clambin/go-common/http/metrics"
	"net/http"
	"time"
)

var _ http.Handler = requestMetricsHandler{}

type requestMetricsHandler struct {
	next    http.Handler
	metrics metrics.RequestMetrics
}

func WithRequestMetrics(m metrics.RequestMetrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requestMetricsHandler{next: next, metrics: m}
	}
}

func (s requestMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	start := time.Now()
	s.next.ServeHTTP(lrw, r)
	s.metrics.Measure(r, lrw.statusCode, time.Since(start))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

type loggingResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

// Write implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ http.Handler = requestMetricsHandler{}

type inflightMetricsHandler struct {
	next    http.Handler
	metrics *metrics.InflightMetrics
}

func WithInflightMetrics(m *metrics.InflightMetrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return inflightMetricsHandler{next: next, metrics: m}
	}
}

func (h inflightMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.metrics.Inc()
	defer h.metrics.Dec()
	h.next.ServeHTTP(w, r)
}
