package roundtripper

import (
	"net/http"
)

// New returns a RoundTripper that implements the behaviour as specified by the different Option parameters.
// New will construct a cascading roundTripper in the order of the provided options. E.g.
//
//	r := New(WithMetrics(...), WithCache(...))
//
// returns a roundTripper that measures the client metrics, and then attempts to get the response from cache.
// Metrics are therefor captured for cached and non-cached responses. On the other hand:
//
//	r := New(WithCache(...), WithMetrics(...))
//
// first attempts to get the response from cache, and failing that, performs real call, recording its http
// metrics.  Metrics will therefore only be captured for real http calls.
//
// New makes no attempt to sanitize the order of the provided options. E.g. WithRoundTripper does not
// cascade to a next roundTripper; it directly performs the http call using the provided transport.
// Therefore, if we create the following RoundTripper:
//
//	r := New(WithRoundTripper(...), WithMetrics(...))
//
// the WithMetrics option will not be used at all.
//
// If no options are provided, or the final option isn't WithRoundTripper, the http call is done using
// http.DefaultTransport.
func New(options ...Option) http.RoundTripper {
	r := http.DefaultTransport
	for i := len(options) - 1; i >= 0; i-- {
		r = options[i](r)
	}
	return r
}

// Option is a function that can be passed to New to specify the behaviour of the RoundTripper.
// See WithMetrics, WithCache, etc. for examples.
type Option func(current http.RoundTripper) http.RoundTripper

// WithRoundTripper specifies the http.RoundTripper to make the final http client call. If no WithRoundTripper is
// provided, RoundTripper defaults to http.DefaultTransport.  Providing a nil roundTripper causes RoundTripper to panic.
func WithRoundTripper(roundTripper http.RoundTripper) Option {
	return func(_ http.RoundTripper) http.RoundTripper {
		return roundTripper
	}
}

// The RoundTripperFunc type is an adapter to allow the use of ordinary functions as HTTP roundTrippers.
// If f is a function with the appropriate signature, RoundTripperFunc(f) is a roundTripper that calls f.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f(req)
func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
