package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
)

type InflightMetrics struct {
	current     atomic.Int32
	maxSeen     atomic.Int32
	inflight    prometheus.Gauge
	maxInflight prometheus.Gauge
}

func NewInflightMetrics(namespace, subsystem string, labels map[string]string) *InflightMetrics {
	return &InflightMetrics{
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

func (m *InflightMetrics) Inc() {
	val := m.current.Add(1)
	m.inflight.Set(float64(val))

	if val > m.maxSeen.Load() {
		m.maxSeen.Store(val)
		m.maxInflight.Set(float64(val))
	}
}

func (m *InflightMetrics) Dec() {
	m.current.Add(-1)
	m.inflight.Set(float64(m.current.Load()))
}

func (m *InflightMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.inflight.Describe(ch)
	m.maxInflight.Describe(ch)
}

func (m *InflightMetrics) Collect(ch chan<- prometheus.Metric) {
	m.inflight.Collect(ch)
	m.maxInflight.Collect(ch)
}
