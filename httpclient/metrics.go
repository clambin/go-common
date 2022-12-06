package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics contains Prometheus metrics to capture during API calls. Each metric is expected to have two labels:
// the first will contain the application issuing the request. The second will contain the endpoint (i.e. Path) of the request.
type Metrics struct {
	latency *prometheus.SummaryVec // measures latency of an API call
	errors  *prometheus.CounterVec // measures any errors returned by an API call
}

// NewMetrics creates a standard set of Prometheus metrics to capture during API calls.
func NewMetrics(namespace, subsystem string) *Metrics {
	return &Metrics{
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_latency"),
			Help: "latency of Reporter API calls",
		}, []string{"application", "endpoint", "method"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_errors_total"),
			Help: "Number of failed Reporter API calls",
		}, []string{"application", "endpoint", "method"}),
	}
}

var _ prometheus.Collector = &Metrics{}

// Describe implements the prometheus.Collector interface so clients can register Metrics as a whole
func (pm *Metrics) Describe(ch chan<- *prometheus.Desc) {
	pm.latency.Describe(ch)
	pm.errors.Describe(ch)
}

// Collect implements the prometheus.Collector interface so clients can register Metrics as a whole
func (pm *Metrics) Collect(ch chan<- prometheus.Metric) {
	pm.latency.Collect(ch)
	pm.errors.Collect(ch)
}

func (pm *Metrics) reportErrors(err error, labelValues ...string) {
	if pm == nil || pm.errors == nil {
		return
	}

	var value float64
	if err != nil {
		value = 1.0
	}
	pm.errors.WithLabelValues(labelValues...).Add(value)
}

func (pm *Metrics) makeLatencyTimer(labelValues ...string) (timer *prometheus.Timer) {
	if pm != nil && pm.latency != nil {
		timer = prometheus.NewTimer(pm.latency.WithLabelValues(labelValues...))
	}
	return
}
