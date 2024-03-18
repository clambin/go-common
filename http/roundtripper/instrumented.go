package roundtripper

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

type instrumentedRoundTripper struct {
	next    http.RoundTripper
	metrics RoundTripMetrics
}

func WithInstrumentedRoundTripper(m RoundTripMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &instrumentedRoundTripper{
			next:    next,
			metrics: m,
		}
	}
}

func (i *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := i.next.RoundTrip(req)
	i.metrics.Measure(req, resp, err, time.Since(start))
	return resp, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type RoundTripMetrics interface {
	Measure(req *http.Request, resp *http.Response, err error, latency time.Duration)
	prometheus.Collector
}

var _ RoundTripMetrics = &DefaultRoundTripMetrics{}

type DefaultRoundTripMetrics struct {
	latency  *prometheus.SummaryVec
	requests *prometheus.CounterVec
}

func NewDefaultRoundTripMetrics(namespace, subsystem string) *DefaultRoundTripMetrics {
	return &DefaultRoundTripMetrics{
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "latency",
			Help:      "request latency",
		},
			[]string{"method", "path", "code"},
		),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "total number of requests",
		},
			[]string{"method", "path", "code"},
		),
	}
}

func (d *DefaultRoundTripMetrics) Measure(req *http.Request, resp *http.Response, err error, latency time.Duration) {
	var code string
	if err == nil {
		code = strconv.Itoa(resp.StatusCode)
	}
	d.latency.WithLabelValues(req.Method, req.URL.Path, code).Observe(latency.Seconds())
	d.requests.WithLabelValues(req.Method, req.URL.Path, code).Add(1)
}

func (d *DefaultRoundTripMetrics) Describe(ch chan<- *prometheus.Desc) {
	d.latency.Describe(ch)
	d.requests.Describe(ch)
}

func (d *DefaultRoundTripMetrics) Collect(ch chan<- prometheus.Metric) {
	d.latency.Collect(ch)
	d.requests.Collect(ch)
}
