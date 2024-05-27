package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

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
	LabelValues LabelValuesFunc
}

type LabelValuesFunc func(*http.Request, int) (method string, path string, code string)

type RequestMetrics interface {
	Measure(*http.Request, int, time.Duration)
	prometheus.Collector
}

func NewRequestMetrics(o Options) RequestMetrics {
	if len(o.Buckets) == 0 {
		o.Buckets = prometheus.DefBuckets
	}
	if o.LabelValues == nil {
		o.LabelValues = func(req *http.Request, statusCode int) (string, string, string) {
			code := strconv.Itoa(statusCode)
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			return req.Method, path, code
		}
	}

	r := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   o.Namespace,
		Subsystem:   o.Subsystem,
		Name:        RequestTotal,
		Help:        "total number of http requests",
		ConstLabels: o.ConstLabels,
	},
		[]string{"method", "path", "code"},
	)

	var d DurationMetric
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

	return requestMetrics{
		CounterVec:     r,
		DurationMetric: d,
		labelValues:    o.LabelValues,
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type requestMetrics struct {
	*prometheus.CounterVec
	DurationMetric
	labelValues func(*http.Request, int) (string, string, string)
}

type CounterMetric interface {
	WithLabelValues(...string) prometheus.Counter
	prometheus.Collector
}

type DurationMetric interface {
	WithLabelValues(...string) prometheus.Observer
	prometheus.Collector
}

func (m requestMetrics) Labels(req *http.Request, statusCode int) (string, string, string) {
	return m.labelValues(req, statusCode)
}

func (m requestMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	l1, l2, l3 := m.Labels(req, statusCode)
	m.CounterVec.WithLabelValues(l1, l2, l3).Inc()
	m.DurationMetric.WithLabelValues(l1, l2, l3).Observe(duration.Seconds())
}

func (m requestMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.CounterVec.Describe(ch)
	m.DurationMetric.Describe(ch)
}

func (m requestMetrics) Collect(ch chan<- prometheus.Metric) {
	m.CounterVec.Collect(ch)
	m.DurationMetric.Collect(ch)
}
