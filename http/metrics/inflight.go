package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
)

type InFlightMetrics interface {
	Inc()
	Dec()
	prometheus.Collector
}

var _ InFlightMetrics = &inflightMetric{}

type inflightMetric struct {
	current     atomic.Int32
	maxSeen     atomic.Int32
	inflight    prometheus.Gauge
	maxInflight prometheus.Gauge
}

func NewInflightMetric(namespace, subsystem string, labels map[string]string) InFlightMetrics {
	return &inflightMetric{
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "inflight_requests",
			Help:        "number of requests currently in flight",
			ConstLabels: labels,
		}),
		maxInflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "inflight_requests_max",
			Help:        "highest number of in flight requests",
			ConstLabels: labels,
		}),
	}
}

func (i *inflightMetric) Inc() {
	val := i.current.Add(1)
	i.inflight.Set(float64(val))

	if val > i.maxSeen.Load() {
		i.maxSeen.Store(val)
		i.maxInflight.Set(float64(val))
	}
}

func (i *inflightMetric) Dec() {
	i.current.Add(-1)
	i.inflight.Set(float64(i.current.Load()))
}

func (i *inflightMetric) Describe(ch chan<- *prometheus.Desc) {
	i.inflight.Describe(ch)
	i.maxInflight.Describe(ch)
}

func (i *inflightMetric) Collect(ch chan<- prometheus.Metric) {
	i.inflight.Collect(ch)
	i.maxInflight.Collect(ch)
}
