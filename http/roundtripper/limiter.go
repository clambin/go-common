package roundtripper

import (
	"fmt"
	"github.com/clambin/go-common/http/roundtripper/internal/sema"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"sync/atomic"
)

var _ http.RoundTripper = &limiter{}

type limiter struct {
	next     http.RoundTripper
	metrics  *LimiterMetrics
	parallel *sema.Semaphore
}

// WithLimiter creates a RoundTripper that limits the number concurrent http requests to maxParallel.
func WithLimiter(maxParallel int64) Option {
	return WithInstrumentedLimiter(maxParallel, nil)
}

// WithInstrumentedLimiter creates a RoundTripper that limits concurrent http requests, like WithLimiter. Additionally,
// it measures the current and maximum outstanding requests as Prometheus metrics.
//
// namespace and subsystem are prepended to the metric names, e.g. api_max_inflight will be called foo_bar_api_max_inflight
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
//
// If namespace, subsystem and application are blank, the call is equivalent to calling WithLimiter, i.e. no metrics are created.
func WithInstrumentedLimiter(maxParallel int64, metrics *LimiterMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &limiter{
			next:     next,
			metrics:  metrics,
			parallel: sema.NewSema(int(maxParallel)),
		}
	}
}

func (l *limiter) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := l.parallel.Acquire(request.Context()); err != nil {
		return nil, fmt.Errorf("acquire semaphore: %w", err)
	}
	defer l.parallel.Release()

	if l.metrics != nil {
		l.metrics.inc()
		defer l.metrics.dec()
	}

	return l.next.RoundTrip(request)
}

var _ prometheus.Collector = &LimiterMetrics{}

type LimiterMetrics struct {
	inFlightMetric    prometheus.Gauge
	inFlightMaxMetric prometheus.Gauge
	inFlight          atomic.Int32
	maxInFlight       atomic.Int32
}

func NewLimiterMetrics(namespace, subsystem string) *LimiterMetrics {
	return &LimiterMetrics{
		inFlightMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "api_inflight_current",
			Help:      "current in flight requests",
		}),
		inFlightMaxMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "api_inflight_max",
			Help:      "maximum in flight requests",
		}),
	}
}

func (m *LimiterMetrics) inc() {
	m.inFlight.Add(1)
	if current := m.inFlight.Load(); current > m.maxInFlight.Load() {
		m.maxInFlight.Store(current)
	}
	m.observe()
}

func (m *LimiterMetrics) dec() {
	m.inFlight.Add(-1)
	m.observe()
}

func (m *LimiterMetrics) observe() {
	m.inFlightMetric.Set(float64(m.inFlight.Load()))
	m.inFlightMaxMetric.Set(float64(m.maxInFlight.Load()))
}

func (m *LimiterMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.inFlightMetric.Describe(ch)
	m.inFlightMaxMetric.Describe(ch)

}

func (m *LimiterMetrics) Collect(ch chan<- prometheus.Metric) {
	m.inFlightMetric.Collect(ch)
	m.inFlightMaxMetric.Collect(ch)
}
