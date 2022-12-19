package httpserver

import (
	"github.com/clambin/go-common/set"
	"net/http"
	"strconv"
	"time"
)

type InstrumentedHandler struct {
	metrics *metrics
}

func (h *InstrumentedHandler) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		path := req.URL.Path

		obs := h.metrics.duration.With(req.Method, path)

		start := time.Now()
		next.ServeHTTP(lrw, req)

		h.metrics.requests.WithLabelValues(req.Method, path, strconv.Itoa(lrw.statusCode)).Inc()
		obs.Observe(time.Since(start).Seconds())
	})
}

type MethodFilteredHandler struct {
	methods set.Set[string]
}

func (h *MethodFilteredHandler) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !h.methods.Has(req.Method) {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, req)
	})
}

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
