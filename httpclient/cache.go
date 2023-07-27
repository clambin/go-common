package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// WithCache creates a RoundTripper that caches the HTTP responses. table determines the caching behaviour per target path.
// If table is empty, all responses are cached for defaultExpiry amount of time.
// If table is not empty, only requests that match an entry in the table are cached, for the amount of time specified in the table's entry.
//
// Expired entries are periodically removed from the cache as per CleanupInterval. If CleanupInterval is zero,
// expired entries will never be removed.
func WithCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) Option {
	return WithInstrumentedCache(table, defaultExpiry, cleanupInterval, "", "", "")
}

// WithInstrumentedCache causes RoundTripper to cache the HTTP responses, as WithCache.  Additionally, it measures how often
// the cache was consulted and hit as Prometheus metrics.
//
// namespace and subsystem are prepended to the metric names, e.g. api_errors_total will be called foo_bar_api_errors_total
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
//
// If namespace, subsystem and application are blank, the call is equivalent to calling WithCache, i.e. no metrics are created.
func WithInstrumentedCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration, namespace, subsystem, application string) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		var metrics *cacheMetrics
		if namespace != "" || subsystem != "" || application != "" {
			metrics = newCacheMetrics(namespace, subsystem, application)
		}
		c := cachingRoundTripper{
			next:    next,
			cache:   newResponseCache(table, defaultExpiry, cleanupInterval),
			metrics: metrics,
		}
		return &c
	}
}

var _ http.RoundTripper = &cachingRoundTripper{}
var _ prometheus.Collector = &cachingRoundTripper{}

type cachingRoundTripper struct {
	next    http.RoundTripper
	cache   *responseCache
	metrics *cacheMetrics
}

func (c *cachingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cacheKey, response, found, err := c.cache.get(request)
	c.metrics.reportCache(found, request.URL.Path, request.Method)
	if found || err != nil {
		return response, err
	}

	if response, err = c.next.RoundTrip(request); err == nil {
		err = c.cache.put(cacheKey, request, response)
	}

	return response, err
}

func (c *cachingRoundTripper) Describe(descs chan<- *prometheus.Desc) {
	if c.metrics != nil {
		c.metrics.Describe(descs)
	}
}

func (c *cachingRoundTripper) Collect(metrics chan<- prometheus.Metric) {
	if c.metrics != nil {
		c.metrics.Collect(metrics)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type cacheMetrics struct {
	cache *prometheus.CounterVec // measures number of times the cache has been consulted
	hits  *prometheus.CounterVec // measures the number of times the cache was used
}

func newCacheMetrics(namespace, subsystem, application string) *cacheMetrics {
	return &cacheMetrics{
		cache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_cache_total"),
			Help:        "Number of times the cache was consulted",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "api_cache_hit_total"),
			Help:        "Number of times the cache was used",
			ConstLabels: map[string]string{"application": application},
		}, []string{"path", "method"}),
	}
}

var _ prometheus.Collector = &cacheMetrics{}

// Describe implements the prometheus.Collector interface
//
// cacheMetrics does not need to be registered and collected separately: it will be collected as part of the http.RoundTripper
// created by NewRoundTripper.
func (cm *cacheMetrics) Describe(ch chan<- *prometheus.Desc) {
	cm.cache.Describe(ch)
	cm.hits.Describe(ch)
}

// Collect implements the prometheus.Collector interface so clients can register roundTripperMetrics as a whole
//
// cacheMetrics does not need to be registered and collected separately: it will be collected as part of the http.RoundTripper
// created by NewRoundTripper.
func (cm *cacheMetrics) Collect(ch chan<- prometheus.Metric) {
	cm.cache.Collect(ch)
	cm.hits.Collect(ch)
}

func (cm *cacheMetrics) reportCache(hit bool, labelValues ...string) {
	if cm == nil {
		return
	}
	cm.cache.WithLabelValues(labelValues...).Inc()
	if hit {
		cm.hits.WithLabelValues(labelValues...).Inc()
	}
}
