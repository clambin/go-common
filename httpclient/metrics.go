package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
)

// roundTripperMetrics contains Prometheus metrics to capture during API calls. Each metric is expected to have two labels:
// the first will contain the application issuing the request. The second will contain the Path of the request.
type roundTripperMetrics struct {
	latency *prometheus.SummaryVec // measures latency of an API call
	errors  *prometheus.CounterVec // measures any errors returned by an API call
	cache   *prometheus.CounterVec // measures number of times the cache has been consulted
	hits    *prometheus.CounterVec // measures the number of times the cache was used
}

func newMetrics(namespace, subsystem, application string) *roundTripperMetrics {
	return &roundTripperMetrics{
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
		cache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_cache_total"),
			Help:        "Number of times the cache was consulted",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_cache_hit_total"),
			Help:        "Number of times the cache was used",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
	}
}

var _ prometheus.Collector = &roundTripperMetrics{}

// Describe implements the prometheus.Collector interface so clients can register roundTripperMetrics as a whole
func (pm *roundTripperMetrics) Describe(ch chan<- *prometheus.Desc) {
	pm.latency.Describe(ch)
	pm.errors.Describe(ch)
	pm.cache.Describe(ch)
	pm.hits.Describe(ch)
}

// Collect implements the prometheus.Collector interface so clients can register roundTripperMetrics as a whole
func (pm *roundTripperMetrics) Collect(ch chan<- prometheus.Metric) {
	pm.latency.Collect(ch)
	pm.errors.Collect(ch)
	pm.cache.Collect(ch)
	pm.hits.Collect(ch)
}

func (pm *roundTripperMetrics) reportErrors(err error, labelValues ...string) {
	if pm == nil {
		return
	}

	var value float64
	if err != nil {
		value = 1.0
	}
	pm.errors.WithLabelValues(labelValues...).Add(value)
}

func (pm *roundTripperMetrics) makeLatencyTimer(labelValues ...string) (timer *prometheus.Timer) {
	if pm != nil {
		timer = prometheus.NewTimer(pm.latency.WithLabelValues(labelValues...))
	}
	return
}

func (pm *roundTripperMetrics) reportCache(hit bool, labelValues ...string) {
	if pm == nil {
		return
	}
	pm.cache.WithLabelValues(labelValues...).Inc()
	if hit {
		pm.hits.WithLabelValues(labelValues...).Inc()
	}
}
