package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

// PrometheusMetrics implements a net/http middleware that records HTTP request statistics as Prometheus metrics
//
// deprecated: moved to github.com/clambin/go-common/http/middleware
type PrometheusMetrics struct {
	requests *prometheus.CounterVec
	duration durationMetric
}

type durationMetric interface {
	WithLabelValues(values ...string) prometheus.Observer
	prometheus.Collector
}

var _ prometheus.Collector = &PrometheusMetrics{}

// NewPrometheusMetrics creates a new PrometheusMetrics middleware for the provided PrometheusMetricsOptions options.
//
// Note: NewPrometheusMetrics will create the required Prometheus metrics, but will not register these with
// any prometheus registry. This is left up to the caller:
//
//	m := NewPrometheusMetrics(options)
//	prometheus.MustRegister(m)
//
// deprecated: moved to github.com/clambin/go-common/http/middleware
func NewPrometheusMetrics(o PrometheusMetricsOptions) *PrometheusMetrics {
	return o.makeMetrics()
}

// Handle returns an http.Handler that instruments the call to the next http.Handler
func (m *PrometheusMetrics) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(lrw, r)

		path := r.URL.Path
		obs := m.duration.WithLabelValues(r.Method, path)

		m.requests.WithLabelValues(r.Method, path, strconv.Itoa(lrw.statusCode)).Inc()
		obs.Observe(time.Since(start).Seconds())
	})
}

// Describe implements the prometheus.Collector interface
func (m *PrometheusMetrics) Describe(descs chan<- *prometheus.Desc) {
	m.requests.Describe(descs)
	m.duration.Describe(descs)
}

// Collect implements the prometheus.Collector interface
func (m *PrometheusMetrics) Collect(c chan<- prometheus.Metric) {
	m.requests.Collect(c)
	m.duration.Collect(c)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

// PrometheusMetricsOptions configure the metrics to be captured by a PrometheusMetrics middleware.
type PrometheusMetricsOptions struct {
	// Namespace is used as namespace to create the metrics' fully qualified name
	Namespace string
	// Subsystem is used as namespace to create the metrics' fully qualified name
	Subsystem string
	// Application will be added as a 'handler' label to all metrics
	Application string
	// MetricsType specifies if the duration metrics should be recorded as a Summary or a Histogram
	MetricsType PrometheusMetricsDurationType
	// Buckets specifies the buckets to use for a Histogram duration metric.
	// If left blank, prometheus.DefBuckets will be used
	Buckets []float64
}

func (o PrometheusMetricsOptions) makeMetrics() *PrometheusMetrics {
	var d durationMetric
	if o.MetricsType == Summary {
		d = o.makeSummary()
	} else {
		d = o.makeHistogram()
	}
	return &PrometheusMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_total"),
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": o.Application},
		}, []string{"method", "path", "code"}),
		duration: d,
	}
}

func (o PrometheusMetricsOptions) makeSummary() *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_duration_seconds"),
		Help:        "Request duration in seconds",
		ConstLabels: prometheus.Labels{"handler": o.Application},
	}, []string{"method", "path"})
}

func (o PrometheusMetricsOptions) makeHistogram() *prometheus.HistogramVec {
	buckets := o.Buckets
	if len(o.Buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_duration_seconds"),
		Help:        "Request duration in seconds",
		ConstLabels: prometheus.Labels{"handler": o.Application},
		Buckets:     buckets,
	}, []string{"method", "path"})
}

// PrometheusMetricsDurationType specifies the type of metrics to record for request duration. Use Summary if you are only interested in the average latency.
// Use Histogram if you want to use a histogram to measure a service level indicator (eg  latency of 95% of all requests).
type PrometheusMetricsDurationType int

const (
	// Summary measures the average duration.
	Summary PrometheusMetricsDurationType = iota
	// Histogram measures the latency in buckets and can be used to calculate a service level indicator. PrometheusMetricsOptions.Buckets
	// specify the buckets to be used. If none are provided, prometheus.DefBuckets will be used.
	Histogram
)

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
