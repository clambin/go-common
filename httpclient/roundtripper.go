package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// RoundTripper performs an HTTP request based on the specified list of RoundTripperOption items.
// It implements the http.RoundTripper interface.
type RoundTripper struct {
	*roundTripperMetrics
	roundTripper http.RoundTripper
	cache        *responseCache
}

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

// NewRoundTripper creates a new RoundTripper
func NewRoundTripper(options ...RoundTripperOption) *RoundTripper {
	r := RoundTripper{roundTripper: http.DefaultTransport}
	for _, option := range options {
		option(&r)
	}
	return &r
}

// RoundTrip performs the HTTP request
func (r *RoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	var cacheKey string
	if r.cache != nil {
		var found bool
		cacheKey, response, found, err = r.cache.get(request)
		r.reportCache(found, request.URL.Path, request.Method)
		if found || err != nil {
			return response, err
		}
	}

	path := request.URL.Path
	timer := r.makeLatencyTimer(path, request.Method)

	response, err = r.roundTripper.RoundTrip(request)

	if timer != nil {
		timer.ObserveDuration()
	}
	r.reportErrors(err, path, request.Method)

	if err == nil && r.cache != nil {
		err = r.cache.put(cacheKey, request, response)
	}

	return response, err
}

// RoundTripperOption specified configuration options for Client
type RoundTripperOption func(*RoundTripper)

// WithMetrics causes RoundTripper to collect Prometheus metrics for each call made. RoundTripper implements
// the prometheus.Collector interface, so the roundtripper can be registered with a prometheus Registry to collect current metric.
//
// Namespace and Subsystem will be prepended to the metric names, e.g. api_errors_total will be called foo_bar_api_errors_total
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metric's application label.
func WithMetrics(namespace, subsystem, application string) func(*RoundTripper) {
	return func(r *RoundTripper) {
		r.roundTripperMetrics = newMetrics(namespace, subsystem, application)
	}
}

// WithCache causes RoundTripper to cache the HTTP responses. Table dictates the caching behaviour per target path.
// If Table is empty, all responses will be cached for DefaultExpiry amount of time. Expired entries will periodically
// be removed from the cache as per CleanupInterval. If CleanupInterval is zero, expired entries will never be removed.
func WithCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) func(tripper *RoundTripper) {
	return func(r *RoundTripper) {
		r.cache = newCache(table, defaultExpiry, cleanupInterval)
	}
}

// WithRoundTripper assigns a final RoundTripper of the chain. Use this to control the final HTTP exchange's behaviour
// (e.g. using a proxy to make the HTTP call).
func WithRoundTripper(roundTripper http.RoundTripper) func(*RoundTripper) {
	return func(r *RoundTripper) {
		r.roundTripper = roundTripper
	}
}
