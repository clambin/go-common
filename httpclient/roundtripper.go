package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

// RoundTripper implements http.RoundTripper. It implements a net/http-compatible transport layer so it can be used
// by http.Client's Transport.
type RoundTripper struct {
	http.RoundTripper
	collectors []prometheus.Collector

	http.HandlerFunc
}

// NewRoundTripper returns a RoundTripper that implements the behaviour as specified by the different Option parameters.
// NewRoundTripper will construct a cascading roundTripper in the order of the provided options. E.g.
//
//	r := NewRoundTripper(WithMetrics(...), WithCache(...))
//
// returns a roundTripper that measures the client metrics, and then attempts to get the response from cache.
// Metrics are therefor captured for cached and non-cached responses. On the other hand:
//
//	r := NewRoundTripper(WithCache(...), WithMetrics(...))
//
// first attempts to get the response from cache, and failing that, performs real call, recording its http
// metrics.  Metrics will therefore only be captured for real http calls.
//
// NewRoundTripper makes no attempt to sanitize the order of the provided options. E.g. WithRoundTripper does not
// cascade to a next roundTripper; it directly performs the http call using the provided transport.
// Therefore, if we create the following RoundTripper:
//
//	r := NewRoundTripper(WithRoundTripper(...), WithMetrics(...))
//
// the WithMetrics option will not be used at all.
//
// If no options are provided, or the final option isn't WithRoundTripper, the http call is done using
// http.DefaultTransport.
//
// deprecated: use github.com/clambin/go-common/http/roundtripper instead
func NewRoundTripper(options ...Option) *RoundTripper {
	r := http.DefaultTransport
	var c []prometheus.Collector

	for i := len(options) - 1; i >= 0; i-- {
		r = options[i](r)
		if m, ok := r.(prometheus.Collector); ok {
			c = append(c, m)
		}
	}
	return &RoundTripper{RoundTripper: r, collectors: c}
}

// Describe implements the prometheus.Collector interface. A RoundTripper can therefore be directly registered and collected
// from a Prometheus registry.
func (r RoundTripper) Describe(descs chan<- *prometheus.Desc) {
	for _, c := range r.collectors {
		c.Describe(descs)
	}
}

// Collect implements the prometheus.Collector interface. A RoundTripper can therefore be directly registered and collected
// from a Prometheus registry.
func (r RoundTripper) Collect(metrics chan<- prometheus.Metric) {
	for _, c := range r.collectors {
		c.Collect(metrics)
	}
}

// Option is a function that can be passed to NewRoundTripper to specify the behaviour of the RoundTripper.
// See WithMetrics, WithCache, etc. for examples.
type Option func(current http.RoundTripper) http.RoundTripper

// The RoundTripperFunc type is an adapter to allow the use of ordinary functions as HTTP roundTrippers.
// If f is a function with the appropriate signature, RoundTripperFunc(f) is a roundTripper that calls f.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f(req)
func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// WithRoundTripper specifies the http.RoundTripper to make the final http client call. If no WithRoundTripper is
// provided, RoundTripper defaults to http.DefaultTransport.  Providing a nil roundTripper causes RoundTripper to panic.
func WithRoundTripper(roundTripper http.RoundTripper) Option {
	return func(_ http.RoundTripper) http.RoundTripper {
		return roundTripper
	}
}
