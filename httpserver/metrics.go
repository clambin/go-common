package httpserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type metrics struct {
	requests *prometheus.CounterVec
	duration durationMetric
}

var _ prometheus.Collector = &metrics{}

func newMetrics(o WithMetrics) *metrics {
	var d durationMetric
	if o.MetricsType == Summary {
		d = newSummary(o)
	} else {
		d = newHistogramMetrics(o)
	}
	return &metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_total"),
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": o.Application},
		}, []string{"method", "path", "code"}),
		duration: d,
	}
}

func (m metrics) Describe(descs chan<- *prometheus.Desc) {
	m.requests.Describe(descs)
	m.duration.Describe(descs)
}

func (m metrics) Collect(c chan<- prometheus.Metric) {
	m.requests.Collect(c)
	m.duration.Collect(c)
}

/////////////////////////

type durationMetric interface {
	With(method, path string) prometheus.Observer
	prometheus.Collector
}

type histogramMetrics struct {
	duration *prometheus.HistogramVec
}

var _ durationMetric = &histogramMetrics{}

func newHistogramMetrics(o WithMetrics) durationMetric {
	if len(o.Buckets) == 0 {
		o.Buckets = prometheus.DefBuckets
	}
	return &histogramMetrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_duration_seconds"),
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": o.Application},
			Buckets:     o.Buckets,
		}, []string{"method", "path"}),
	}
}

// With returns the Observer to record request duration
func (m *histogramMetrics) With(method, path string) prometheus.Observer {
	return m.duration.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// Describe implements the prometheus Collector interface
func (m *histogramMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.duration.Describe(ch)
}

// Collect implements the prometheus Collector interface
func (m *histogramMetrics) Collect(ch chan<- prometheus.Metric) {
	m.duration.Collect(ch)
}

type summary struct {
	duration *prometheus.SummaryVec
}

var _ durationMetric = &summary{}

func newSummary(o WithMetrics) *summary {
	return &summary{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:        prometheus.BuildFQName(o.Namespace, o.Subsystem, "http_requests_duration_seconds"),
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": o.Application},
		}, []string{"method", "path"}),
	}
}

// With returns the Observer to record request duration
func (m *summary) With(method, path string) prometheus.Observer {
	return m.duration.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// Describe implements the prometheus Collector interface
func (m *summary) Describe(ch chan<- *prometheus.Desc) {
	m.duration.Describe(ch)
}

// Collect implements the prometheus Collector interface
func (m *summary) Collect(ch chan<- prometheus.Metric) {
	m.duration.Collect(ch)
}
