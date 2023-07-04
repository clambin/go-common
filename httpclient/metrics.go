package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

// WithMetrics creates a RoundTripper that measures latency and errors of outbound requests.
//
// namespace and subsystem are prepended to the metric names, e.g. api_latenct will be called foo_bar_api_latency
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
func WithMetrics(namespace, subsystem, application string) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &instrumentedRoundTripper{
			metrics: newMetrics(namespace, subsystem, application),
			next:    next,
		}
	}
}

var _ http.RoundTripper = &instrumentedRoundTripper{}
var _ prometheus.Collector = &instrumentedRoundTripper{}

type instrumentedRoundTripper struct {
	metrics roundTripperMetrics
	next    http.RoundTripper
}

func (m *instrumentedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	path := request.URL.Path
	timer := m.metrics.makeLatencyTimer(path, request.Method)

	response, err := m.next.RoundTrip(request)

	if timer != nil {
		timer.ObserveDuration()
	}
	m.metrics.reportErrors(err, path, request.Method)
	return response, err
}

func (m *instrumentedRoundTripper) Describe(descs chan<- *prometheus.Desc) {
	m.metrics.latency.Describe(descs)
	m.metrics.errors.Describe(descs)
}

func (m *instrumentedRoundTripper) Collect(metrics chan<- prometheus.Metric) {
	m.metrics.latency.Collect(metrics)
	m.metrics.errors.Collect(metrics)
}

// roundTripperMetrics contains Prometheus metrics to capture during API calls. Each metric is expected to have two labels:
// the first will contain the application issuing the request. The second will contain the Path of the request.
type roundTripperMetrics struct {
	latency *prometheus.SummaryVec // measures latency of an API call
	errors  *prometheus.CounterVec // measures any errors returned by an API call
}

func newMetrics(namespace, subsystem, application string) roundTripperMetrics {
	return roundTripperMetrics{
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_latency"),
			Help:        "latency of Reporter API calls",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_errors_total"),
			Help:        "Number of failed Reporter API calls",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
	}
}

func (pm *roundTripperMetrics) makeLatencyTimer(labelValues ...string) (timer *prometheus.Timer) {
	return prometheus.NewTimer(pm.latency.WithLabelValues(labelValues...))
}

func (pm *roundTripperMetrics) reportErrors(err error, labelValues ...string) {
	var value float64
	if err != nil {
		value = 1.0
	}
	pm.errors.WithLabelValues(labelValues...).Add(value)
}
