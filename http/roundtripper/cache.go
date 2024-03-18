package roundtripper

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/clambin/go-common/cache"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httputil"
	"regexp"
	"time"
)

// WithCache creates a RoundTripper that caches the HTTP responses. table determines the caching behaviour per target path.
// If table is empty, all responses are cached for defaultExpiry amount of time.
// If table is not empty, only requests that match an entry in the table are cached, for the amount of time specified in the table's entry.
//
// Expired entries are periodically removed from the cache as per CleanupInterval. If CleanupInterval is zero,
// expired entries will never be removed.
func WithCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) Option {
	return WithInstrumentedCache(table, defaultExpiry, cleanupInterval, nil)
}

// WithInstrumentedCache causes RoundTripper to cache the HTTP responses, as WithCache.  Additionally, it measures how often
// the cache was consulted and hit as Prometheus metrics.
//
// namespace and subsystem are prepended to the metric names, e.g. api_errors_total will be called foo_bar_api_errors_total
// if namespace and subsystem are set to foo and bar respectively. Application will be set in the metrics' application label.
//
// If namespace, subsystem and application are blank, the call is equivalent to calling WithCache, i.e. no metrics are created.
func WithInstrumentedCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration, metrics CacheMetrics) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		c := cachingRoundTripper{
			next:    next,
			cache:   newResponseCache(table, defaultExpiry, cleanupInterval),
			metrics: metrics,
		}
		return &c
	}
}

var _ http.RoundTripper = &cachingRoundTripper{}

type cachingRoundTripper struct {
	next    http.RoundTripper
	cache   *responseCache
	metrics CacheMetrics
}

func (c *cachingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cacheKey, response, found, err := c.cache.get(request)
	if c.metrics != nil {
		c.metrics.Measure(request, found)
	}
	if found || err != nil {
		return response, err
	}

	if response, err = c.next.RoundTrip(request); err == nil {
		err = c.cache.put(cacheKey, request, response)
	}

	return response, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type CacheMetrics interface {
	Measure(r *http.Request, found bool)
	prometheus.Collector
}

var _ CacheMetrics = &defaultCacheMetrics{}

type defaultCacheMetrics struct {
	cache *prometheus.CounterVec // measures number of times the cache has been consulted
	hits  *prometheus.CounterVec // measures the number of times the cache was used
}

func NewCacheMetrics(namespace, subsystem string) CacheMetrics {
	return &defaultCacheMetrics{
		cache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_cache_total"),
			Help: "Number of times the cache was consulted",
		}, []string{"path", "method"}),
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_cache_hit_total"),
			Help: "Number of times the cache was used",
		}, []string{"path", "method"}),
	}
}

var _ prometheus.Collector = &defaultCacheMetrics{}

func (m *defaultCacheMetrics) Measure(r *http.Request, hit bool) {
	m.cache.WithLabelValues(r.URL.Path, r.Method).Inc()
	if hit {
		m.hits.WithLabelValues(r.URL.Path, r.Method).Inc()
	}
}

func (m *defaultCacheMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.cache.Describe(ch)
	m.hits.Describe(ch)
}

func (m *defaultCacheMetrics) Collect(ch chan<- prometheus.Metric) {
	m.cache.Collect(ch)
	m.hits.Collect(ch)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// CacheTable holds the endpoints that should be cached. If table is empty, all responses will be cached.
type CacheTable []*CacheTableEntry

// DefaultCacheTable is a CacheTable that caches all requests.
var DefaultCacheTable CacheTable

func (c CacheTable) shouldCache(r *http.Request) (match bool, expiry time.Duration) {
	if len(c) == 0 {
		return true, 0
	}

	for _, entry := range c {
		if match, expiry = entry.shouldCache(r); match {
			return
		}
	}
	return
}

func (c CacheTable) mustCompile() {
	for _, entry := range c {
		entry.mustCompile()
	}
}

// CacheTableEntry contains a single endpoint that should be cached. If the Path is a regular expression, IsRegExp must be true.
type CacheTableEntry struct {
	// Path is the URL Path for requests whose responses should be cached.
	// Can be a literal path, or a regular expression. In the latter case, set IsRegExp to true
	Path string
	// Methods is the list of HTTP Methods for which requests the response should be cached.
	// If empty, requests for any method will be cached.
	Methods []string
	// IsRegExp indicates if the Path is a regular expression.
	// CacheTableEntry will panic if Path does not contain a valid regular expression.
	IsRegExp bool
	// Expiry indicates how long a response should be cached.
	Expiry         time.Duration
	compiledRegExp *regexp.Regexp
}

// var CacheEverything []CacheTableEntry

func (entry *CacheTableEntry) shouldCache(r *http.Request) (match bool, expiry time.Duration) {
	match = entry.matchesPath(r)
	if !match {
		return
	}
	match = entry.matchesMethod(r)
	return match, entry.Expiry
}

func (entry *CacheTableEntry) matchesPath(r *http.Request) bool {
	path := r.URL.Path
	if entry.IsRegExp {
		return entry.compiledRegExp.MatchString(path)
	}
	return entry.Path == path
}

func (entry *CacheTableEntry) matchesMethod(r *http.Request) bool {
	if len(entry.Methods) == 0 {
		return true
	}
	for _, method := range entry.Methods {
		if method == r.Method {
			return true
		}
	}
	return false
}

func (entry *CacheTableEntry) mustCompile() {
	if !entry.IsRegExp {
		return
	}
	var err error
	if entry.compiledRegExp, err = regexp.Compile(entry.Path); err != nil {
		panic(fmt.Errorf("cacheTable: invalid regexp '%s': %w", entry.Path, err))
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type responseCache struct {
	table CacheTable
	cache *cache.Cache[string, []byte]
}

func newResponseCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) *responseCache {
	c := &responseCache{
		table: table,
		cache: cache.New[string, []byte](defaultExpiry, cleanupInterval),
	}
	c.table.mustCompile()
	return c
}

func (c *responseCache) get(req *http.Request) (string, *http.Response, bool, error) {
	key := getCacheKey(req)
	body, found := c.cache.Get(key)
	if !found {
		return key, nil, false, nil
	}

	r := bufio.NewReader(bytes.NewReader(body))
	resp, err := http.ReadResponse(r, req)
	return key, resp, found, err
}

func (c *responseCache) put(key string, req *http.Request, resp *http.Response) error {
	shouldCache, expiry := c.shouldCache(req)
	if !shouldCache {
		return nil
	}

	buf, err := httputil.DumpResponse(resp, true)
	if err == nil {
		c.cache.AddWithExpiry(key, buf, expiry)
	}
	return err
}

func (c *responseCache) shouldCache(r *http.Request) (bool, time.Duration) {
	shouldCache, expiry := c.table.shouldCache(r)
	if shouldCache && expiry == 0 {
		expiry = c.cache.GetDefaultExpiration()
	}
	return shouldCache, expiry
}

// TODO: allow this to be overridden

func getCacheKey(r *http.Request) string {
	return r.Method + " | " + r.URL.Path
}
