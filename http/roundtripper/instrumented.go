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

// WithInstrumentedRoundTripper creates an [http.RoundTripper] that records requests metrics to the provided [RoundTripMetrics].
// The caller must register the RoundTripMetrics with a Prometheus registry.
func WithInstrumentedRoundTripper(m RoundTripMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &instrumentedRoundTripper{
			next:    next,
			metrics: m,
		}
	}
}

// RoundTrip implements the [http.RoundTripper] interface.
func (i *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := i.next.RoundTrip(req)
	i.metrics.Measure(req, resp, err, time.Since(start))
	return resp, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RoundTripMetrics measure metrics for each request processed by an instrumented roundtripper, created by WithInstrumentedRoundTripper.
//
// To create a custom RoundTripMetrics, implement the Measure method, which measures & records the necessary Prometheus metrics,
// and implements the [prometheus.Collector] interface. See [DefaultRoundTripMetrics] for an example.
type RoundTripMetrics interface {
	Measure(req *http.Request, resp *http.Response, err error, duration time.Duration)
	prometheus.Collector
}

var _ RoundTripMetrics = &DefaultRoundTripMetrics{}

// DefaultRoundTripMetrics measure the request's total count and duration (by method, path and status code)
type DefaultRoundTripMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.SummaryVec
}

// NewDefaultRoundTripMetrics returns a new DefaultRoundTripMetrics. The caller must register the returned metrics with a Prometheus registry.
//
// namespace and subsystem are prepended to the metric name.
// application is added to the metric as a label "application". If application is empty, the label is not added.
func NewDefaultRoundTripMetrics(namespace, subsystem, application string) *DefaultRoundTripMetrics {
	var constLabels map[string]string
	if application != "" {
		constLabels = map[string]string{"application": application}
	}

	return &DefaultRoundTripMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_requests_total",
			Help:        "total number of http requests",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "http_request_duration_seconds",
			Help:        "http request duration in seconds",
			ConstLabels: constLabels,
		},
			[]string{"method", "path", "code"},
		),
	}
}

// Measure measures the total number of requests and the duration of the current request.
func (d *DefaultRoundTripMetrics) Measure(req *http.Request, resp *http.Response, err error, duration time.Duration) {
	var code string
	if err == nil {
		code = strconv.Itoa(resp.StatusCode)
	}
	d.requests.WithLabelValues(req.Method, req.URL.Path, code).Add(1)
	d.duration.WithLabelValues(req.Method, req.URL.Path, code).Observe(duration.Seconds())
}

// Describe implements the [prometheus.Collector] interface.
func (d *DefaultRoundTripMetrics) Describe(ch chan<- *prometheus.Desc) {
	d.requests.Describe(ch)
	d.duration.Describe(ch)
}

// Collect implements the [prometheus.Collector] interface.
func (d *DefaultRoundTripMetrics) Collect(ch chan<- prometheus.Metric) {
	d.requests.Collect(ch)
	d.duration.Collect(ch)
}
