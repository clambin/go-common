package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

type RequestMetrics interface {
	Measure(req *http.Request, statusCode int, duration time.Duration)
	prometheus.Collector
}

var _ RequestMetrics = RequestSummaryMetrics{}

type RequestSummaryMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.SummaryVec
}

const (
	RequestTotal     = "http_requests_total"
	RequestsDuration = "http_request_duration_seconds"
)

func NewRequestSummaryMetrics(namespace, subsystem string, constLabels map[string]string) *RequestSummaryMetrics {
	return &RequestSummaryMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        RequestTotal,
			Help:        "total number of http requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        RequestsDuration,
			Help:        "duration of http requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
	}
}

func (m RequestSummaryMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	code := strconv.Itoa(statusCode)
	path := req.URL.Path
	if path == "" {
		path = "/"
	}
	m.requests.WithLabelValues(req.Method, path, code).Inc()
	m.duration.WithLabelValues(req.Method, path, code).Observe(duration.Seconds())
}

func (m RequestSummaryMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.duration.Describe(ch)
}

func (m RequestSummaryMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.duration.Collect(ch)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ RequestMetrics = RequestHistogramMetrics{}

type RequestHistogramMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewRequestHistogramMetrics(namespace, subsystem string, constLabels map[string]string, buckets ...float64) *RequestHistogramMetrics {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return &RequestHistogramMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        RequestTotal,
			Help:        "total number of http requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        RequestsDuration,
			Help:        "duration of http requests",
			ConstLabels: constLabels,
			Buckets:     buckets,
		},
			[]string{"method", "path", "code"},
		),
	}
}

func (m RequestHistogramMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	code := strconv.Itoa(statusCode)
	path := req.URL.Path
	if path == "" {
		path = "/"
	}
	m.requests.WithLabelValues(req.Method, path, code).Inc()
	m.duration.WithLabelValues(req.Method, path, code).Observe(duration.Seconds())

}

func (m RequestHistogramMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.duration.Describe(ch)
}

func (m RequestHistogramMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.duration.Collect(ch)
}
