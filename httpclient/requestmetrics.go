package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// WithMetrics creates a RoundTripper that measures outbound HTTP requests and exposes them to Prometheus.
//
// namespace and subsystem are prepended to the metric names, e.g. api_latency will be called foo_bar_api_latency
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
func WithMetrics(namespace, subsystem, application string) Option {
	return WithCustomMetrics(defaultRequestMeasurer{
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_latency"),
			Help:        "latency of HTTP calls",
			ConstLabels: map[string]string{"application": application},
		}, []string{"method", "path"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_errors_total"),
			Help:        "Number of failed HTTP calls",
			ConstLabels: map[string]string{"application": application},
		}, []string{"method", "path"}),
	})
}

// WithCustomMetrics creates a RoundTripper that measures outbound HTTP requests and exposes them to Prometheus.
// The provided RequestMeasurer determines which characteristics to measure and how to present them to Prometheus.
// NewDefaultRequestMeasurer measures the request's latency and error count.
func WithCustomMetrics(o RequestMeasurer) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return measuringRoundTripper{next: next, RequestMeasurer: o}
	}
}

// A RequestMeasurer measures performance characteristics of an outbound HTTP request.
type RequestMeasurer interface {
	MeasureRequest(*http.Request, *http.Response, error, time.Duration)
	prometheus.Collector
}

var _ http.RoundTripper = &measuringRoundTripper{}
var _ prometheus.Collector = &measuringRoundTripper{}

type measuringRoundTripper struct {
	next http.RoundTripper
	RequestMeasurer
}

func (r measuringRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := r.next.RoundTrip(req)
	r.MeasureRequest(req, resp, err, time.Since(start))
	return resp, err
}

var _ RequestMeasurer = defaultRequestMeasurer{}

type defaultRequestMeasurer struct {
	latency *prometheus.SummaryVec // measures latency of an API call
	errors  *prometheus.CounterVec // measures any errors returned by an API call
}

// MeasureRequest measures the HTTP request's latency and error count.
func (d defaultRequestMeasurer) MeasureRequest(req *http.Request, _ *http.Response, err error, duration time.Duration) {
	d.latency.WithLabelValues(req.Method, req.URL.Path).Observe(duration.Seconds())
	var val float64
	if err != nil {
		val = 1
	}
	d.errors.WithLabelValues(req.Method, req.URL.Path).Add(val)
}

// Describe implements the prometheus.Collector interface.
//
// A RequestMeasurer does not need to be registered and collected separately: it will be collected as part of the http.RoundTripper
// created by NewRoundTripper.
func (d defaultRequestMeasurer) Describe(descs chan<- *prometheus.Desc) {
	d.latency.Describe(descs)
	d.errors.Describe(descs)
}

// Collect implements the prometheus.Collector interface.
//
// A RequestMeasurer does not need to be registered and collected separately: it will be collected as part of the http.RoundTripper
// created by NewRoundTripper.
func (d defaultRequestMeasurer) Collect(metrics chan<- prometheus.Metric) {
	d.latency.Collect(metrics)
	d.errors.Collect(metrics)
}
