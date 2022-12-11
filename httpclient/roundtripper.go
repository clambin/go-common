package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// RoundTripper performs an HTTP request based on the specified list of RoundTripperOption items.
// It implements the http.RoundTripper interface.
type RoundTripper struct {
	roundTripper http.RoundTripper
	metrics      *roundTripperMetrics
	cache        *Cache
}

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

// NewRoundTripper creates a new RoundTripper
func NewRoundTripper(options ...RoundTripperOption) *RoundTripper {
	r := &RoundTripper{roundTripper: http.DefaultTransport}
	for _, option := range options {
		option.apply(r)
	}
	return r
}

// RoundTrip performs the HTTP request
func (r *RoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	var cacheKey string
	if r.cache != nil {
		var found bool
		cacheKey, response, found, err = r.cache.Get(request)
		r.metrics.reportCache(found, request.URL.Path, request.Method)
		if found || err != nil {
			return response, err
		}
	}

	path := request.URL.Path
	timer := r.metrics.makeLatencyTimer(path, request.Method)

	response, err = r.roundTripper.RoundTrip(request)

	if timer != nil {
		timer.ObserveDuration()
	}
	r.metrics.reportErrors(err, path, request.Method)

	if err == nil && r.cache != nil {
		err = r.cache.Put(cacheKey, request, response)
	}

	return response, err
}

// Describe implements the prometheus.Collector interface
func (r *RoundTripper) Describe(descs chan<- *prometheus.Desc) {
	if r.metrics != nil {
		r.metrics.Describe(descs)
	}
}

// Collect implements the prometheus.Collector interface
func (r *RoundTripper) Collect(metrics chan<- prometheus.Metric) {
	if r.metrics != nil {
		r.metrics.Collect(metrics)
	}
}

// RoundTripperOption specified configuration options for Client
type RoundTripperOption interface {
	apply(r *RoundTripper)
}

// WithRoundTripperMetrics causes RoundTripper to collect Prometheus metrics for each call made. RoundTripper implements
// the prometheus.Collector interface, so the roundtripper can be registered with a prometheus Registry to collect current metric.
//
// Namespace and Subsystem will be prepended to the metric names, e.g. api_errors_total will be called foo_bar_api_errors_total
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metric's application label.
type WithRoundTripperMetrics struct {
	Namespace   string
	Subsystem   string
	Application string
}

func (o WithRoundTripperMetrics) apply(r *RoundTripper) {
	r.metrics = newMetrics(o.Namespace, o.Subsystem, o.Application)
}

// WithCache causes RoundTripper to cache the HTTP responses. Table dictates the caching behaviour per target path.
// If Table is empty, all responses will be cached for DefaultExpiry amount of time. Expired entries will periodically
// be removed from the cache as per CleanupInterval. If CleanupInterval is zero, expired entries will never be removed.
type WithCache struct {
	Table           CacheTable
	DefaultExpiry   time.Duration
	CleanupInterval time.Duration
}

func (o WithCache) apply(r *RoundTripper) {
	r.cache = NewCache(o.Table, o.DefaultExpiry, o.CleanupInterval)
}

// WithRoundTripper assigns a final RoundTripper of the chain. Use this to control the final HTTP exchange's behaviour
// (e.g. using a proxy to make the HTTP call).
type WithRoundTripper struct {
	RoundTripper http.RoundTripper
}

func (o WithRoundTripper) apply(r *RoundTripper) {
	r.roundTripper = o.RoundTripper
}
