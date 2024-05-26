package roundtripper

import (
	"github.com/clambin/go-common/http/metrics"
	"net/http"
	"time"
)

var _ http.RoundTripper = &roundTripMeasurer{}

type roundTripMeasurer struct {
	next    http.RoundTripper
	metrics metrics.RequestMetrics
}

// WithRequestMetrics creates a [http.RoundTripper] that measures request count and duration.
// The caller must register the metrics with a Prometheus registry.
func WithRequestMetrics(m metrics.RequestMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &roundTripMeasurer{
			next:    next,
			metrics: m,
		}
	}
}

func (i *roundTripMeasurer) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := i.next.RoundTrip(req)
	var statusCode int
	if err == nil {
		statusCode = resp.StatusCode
	}
	i.metrics.Measure(req, statusCode, time.Since(start))
	return resp, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ http.RoundTripper = &inflightMeasurer{}

type inflightMeasurer struct {
	metrics metrics.InFlightMetrics
	next    http.RoundTripper
}

// WithInflightMetrics creates a [http.RoundTripper] that measures outstanding requests.
// The caller must register the metrics with a Prometheus registry.
func WithInflightMetrics(m metrics.InFlightMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &inflightMeasurer{
			next:    next,
			metrics: m,
		}
	}
}

func (i inflightMeasurer) RoundTrip(request *http.Request) (*http.Response, error) {
	i.metrics.Inc()
	defer i.metrics.Dec()
	return i.next.RoundTrip(request)
}
