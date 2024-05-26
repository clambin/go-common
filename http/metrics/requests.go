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

const (
	RequestTotal     = "http_requests_total"
	RequestsDuration = "http_request_duration_seconds"
)

type DurationType int

const (
	SummaryDuration DurationType = iota
	HistogramDuration
)

type Options struct {
	Namespace   string
	Subsystem   string
	ConstLabels prometheus.Labels
	DurationType
	Buckets     []float64
	MakeMetrics func() (CounterMetric, DurationMetric)
	Labels      func(req *http.Request, statusCode int, _ time.Duration) []string
}

func NewRequestMetrics(o Options) RequestMetrics {
	var r CounterMetric
	var d DurationMetric

	if len(o.Buckets) == 0 {
		o.Buckets = prometheus.DefBuckets
	}

	if o.MakeMetrics != nil {
		r, d = o.MakeMetrics()
		return requestMetrics{
			CounterMetric:  r,
			DurationMetric: d,
		}
	}

	r = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        RequestTotal,
		Help:        "total number of http requests",
		ConstLabels: o.ConstLabels,
	},
		[]string{"method", "path", "code"},
	)

	switch o.DurationType {
	case SummaryDuration:
		d = prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:   o.Namespace,
			Subsystem:   o.Subsystem,
			Name:        RequestsDuration,
			Help:        "duration of http requests",
			ConstLabels: o.ConstLabels,
		},
			[]string{"method", "path", "code"},
		)
	case HistogramDuration:
		d = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   o.Namespace,
			Subsystem:   o.Subsystem,
			Name:        RequestsDuration,
			Help:        "duration of http requests",
			ConstLabels: o.ConstLabels,
			Buckets:     o.Buckets,
		},
			[]string{"method", "path", "code"},
		)
	}

	return requestMetrics{CounterMetric: r, DurationMetric: d}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type requestMetrics struct {
	CounterMetric
	DurationMetric
}

type CounterMetric interface {
	WithLabelValues(...string) prometheus.Counter
	prometheus.Collector
}

type DurationMetric interface {
	WithLabelValues(...string) prometheus.Observer
	prometheus.Collector
}

func (m requestMetrics) Labels(req *http.Request, statusCode int, _ time.Duration) []string {
	code := strconv.Itoa(statusCode)
	path := req.URL.Path
	if path == "" {
		path = "/"
	}
	return []string{req.Method, path, code}
}

func (m requestMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	labels := m.Labels(req, statusCode, duration)
	m.CounterMetric.WithLabelValues(labels...).Inc()
	m.DurationMetric.WithLabelValues(labels...).Observe(duration.Seconds())
}

func (m requestMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.CounterMetric.Describe(ch)
	m.DurationMetric.Describe(ch)
}

func (m requestMetrics) Collect(ch chan<- prometheus.Metric) {
	m.CounterMetric.Collect(ch)
	m.DurationMetric.Collect(ch)
}
