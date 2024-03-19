package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

var _ http.Handler = serverMetricsHandler{}

type serverMetricsHandler struct {
	next    http.Handler
	metrics ServerMetrics
}

func WithServerMetrics(m ServerMetrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return serverMetricsHandler{next: next, metrics: m}
	}
}

func (s serverMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	start := time.Now()
	s.next.ServeHTTP(lrw, r)
	s.metrics.Measure(r, lrw.statusCode, time.Since(start))
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type ServerMetrics interface {
	Measure(req *http.Request, statusCode int, duration time.Duration)
	prometheus.Collector
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ServerMetrics = &DefaultServerSummaryMetrics{}

type DefaultServerSummaryMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.SummaryVec
}

func NewDefaultServerSummaryMetrics(namespace, subsystem, application string) *DefaultServerSummaryMetrics {
	var constLabels map[string]string
	if application != "" {
		constLabels = map[string]string{"application": application}
	}
	return &DefaultServerSummaryMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_server_requests_total",
			Help:        "total number of http server requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_server_request_duration_seconds",
			Help:        "total number of http server requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
	}
}

func (d DefaultServerSummaryMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	code := strconv.Itoa(statusCode)
	d.requests.WithLabelValues(req.Method, req.URL.Path, code).Inc()
	d.duration.WithLabelValues(req.Method, req.URL.Path, code).Observe(duration.Seconds())
}

func (d DefaultServerSummaryMetrics) Describe(ch chan<- *prometheus.Desc) {
	d.requests.Describe(ch)
	d.duration.Describe(ch)
}

func (d DefaultServerSummaryMetrics) Collect(ch chan<- prometheus.Metric) {
	d.requests.Collect(ch)
	d.duration.Collect(ch)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ServerMetrics = &DefaultServerHistogramMetrics{}

type DefaultServerHistogramMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewDefaultServerHistogramMetrics(namespace, subsystem, application string, buckets ...float64) *DefaultServerHistogramMetrics {
	var constLabels map[string]string
	if application != "" {
		constLabels = map[string]string{"application": application}
	}
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return &DefaultServerHistogramMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_server_requests_total",
			Help:        "total number of http server requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_server_request_duration_seconds",
			Help:        "total number of http server requests",
			ConstLabels: constLabels,
			Buckets:     buckets,
		},
			[]string{"method", "path", "code"},
		),
	}
}

func (d DefaultServerHistogramMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	code := strconv.Itoa(statusCode)
	d.requests.WithLabelValues(req.Method, req.URL.Path, code).Inc()
	d.duration.WithLabelValues(req.Method, req.URL.Path, code).Observe(duration.Seconds())
}

func (d DefaultServerHistogramMetrics) Describe(ch chan<- *prometheus.Desc) {
	d.requests.Describe(ch)
	d.duration.Describe(ch)
}

func (d DefaultServerHistogramMetrics) Collect(ch chan<- prometheus.Metric) {
	d.requests.Collect(ch)
	d.duration.Collect(ch)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ServerMetrics = NoMetrics{}

type NoMetrics struct{}

func (n NoMetrics) Measure(_ *http.Request, _ int, _ time.Duration) {
}

func (n NoMetrics) Describe(_ chan<- *prometheus.Desc) {
}

func (n NoMetrics) Collect(_ chan<- prometheus.Metric) {
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
