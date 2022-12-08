package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

type RoundTripper struct {
	roundTripper http.RoundTripper
	metrics      *RoundTripperMetrics
	cache        *Cache
}

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

func NewRoundTripper(options ...RoundTripperOption) *RoundTripper {
	r := &RoundTripper{}
	for _, option := range options {
		option.apply(r)
	}
	if r.roundTripper == nil {
		r.roundTripper = http.DefaultTransport
	}
	return r
}

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

	response, err = http.DefaultTransport.RoundTrip(request)

	if timer != nil {
		timer.ObserveDuration()
	}
	r.metrics.reportErrors(err, path, request.Method)

	if r.cache != nil {
		err = r.cache.Put(cacheKey, request, response)
	}

	return response, err
}

func (r *RoundTripper) Describe(descs chan<- *prometheus.Desc) {
	if r.metrics != nil {
		r.metrics.Describe(descs)
	}
}

func (r *RoundTripper) Collect(metrics chan<- prometheus.Metric) {
	if r.metrics != nil {
		r.metrics.Collect(metrics)
	}
}

// RoundTripperOption specified configuration options for Client
type RoundTripperOption interface {
	apply(r *RoundTripper)
}

type WithRoundTripperMetrics struct {
	Namespace   string
	Subsystem   string
	Application string
}

func (o WithRoundTripperMetrics) apply(r *RoundTripper) {
	r.metrics = newMetrics(o.Namespace, o.Subsystem, o.Application)
}

type WithCache struct {
	Table           CacheTable
	DefaultExpiry   time.Duration
	CleanupInterval time.Duration
}

func (o WithCache) apply(r *RoundTripper) {
	r.cache = NewCache(o.Table, o.DefaultExpiry, o.CleanupInterval)
}
