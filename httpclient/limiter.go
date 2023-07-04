package httpclient

import (
	"fmt"
	"github.com/clambin/go-common/httpclient/internal"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"sync"
)

var _ http.RoundTripper = &limiter{}
var _ prometheus.Collector = &limiter{}

type limiter struct {
	next        http.RoundTripper
	metrics     *limiterMetrics
	parallel    *internal.Semaphore
	lock        sync.RWMutex
	inFlight    int
	maxInFlight int
}

// WithLimiter creates a RoundTripper that limits the number concurrent http requests to maxParallel.
func WithLimiter(maxParallel int64) Option {
	return WithInstrumentedLimiter(maxParallel, "", "", "")
}

// WithInstrumentedLimiter creates a RoundTripper that limits concurrent http requests, like WithLimiter. Additionally,
// it measures the current and maximum outstanding requests as Prometheus metrics.
//
// namespace and subsystem are prepended to the metric names, e.g. api_max_inflight will be called foo_bar_api_max_inflight
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
//
// If namespace, subsystem and application are blank, the call is equivalent to calling WithLimiter, i.e. no metrics are created.
func WithInstrumentedLimiter(maxParallel int64, namespace, subsystem, application string) Option {
	var m *limiterMetrics
	if namespace != "" || subsystem != "" || application != "" {
		m = newLimiterMetrics(namespace, subsystem, application)
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return &limiter{
			next:     next,
			metrics:  m,
			parallel: internal.NewSema(int(maxParallel)),
		}
	}
}

func (l *limiter) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := l.parallel.Acquire(request.Context()); err != nil {
		return nil, fmt.Errorf("acquire semaphore: %w", err)
	}
	defer l.parallel.Release()

	l.inc()
	defer l.dec()

	return l.next.RoundTrip(request)
}

func (l *limiter) inc() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.inFlight++
	if l.inFlight > l.maxInFlight {
		l.maxInFlight = l.inFlight
	}
	if l.metrics != nil {
		l.metrics.inFlightMetric.Set(float64(l.inFlight))
		l.metrics.maxInFlightMetric.Set(float64(l.maxInFlight))
	}
}

func (l *limiter) dec() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.inFlight--
	if l.metrics != nil {
		l.metrics.inFlightMetric.Set(float64(l.inFlight))
		l.metrics.maxInFlightMetric.Set(float64(l.maxInFlight))
	}
}

func (l *limiter) Describe(descs chan<- *prometheus.Desc) {
	if l.metrics != nil {
		l.metrics.Describe(descs)
	}
}

func (l *limiter) Collect(metrics chan<- prometheus.Metric) {
	if l.metrics != nil {
		l.metrics.Collect(metrics)
	}
}

var _ prometheus.Collector = &limiterMetrics{}

type limiterMetrics struct {
	inFlightMetric    prometheus.Gauge
	maxInFlightMetric prometheus.Gauge
}

func newLimiterMetrics(namespace, subsystem, application string) *limiterMetrics {
	return &limiterMetrics{
		inFlightMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "api_inflight",
			Help:        "Number of requests in flight",
			ConstLabels: map[string]string{"application": application},
		}),
		maxInFlightMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "api_max_inflight",
			Help:        "Maximum number of requests in flight",
			ConstLabels: map[string]string{"application": application},
		}),
	}
}

func (m *limiterMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.inFlightMetric.Describe(ch)
	m.maxInFlightMetric.Describe(ch)
}

func (m *limiterMetrics) Collect(ch chan<- prometheus.Metric) {
	m.inFlightMetric.Collect(ch)
	m.maxInFlightMetric.Collect(ch)
}
