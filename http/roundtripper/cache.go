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

// CacheOptions contains configuration options for a caching roundTripper.
type CacheOptions struct {
	// CacheTable determines which requests are cached and for how long. If empty, all requests are cached.
	CacheTable
	// DefaultExpiration defines how long requests are cached when CacheTable is empty.
	DefaultExpiration time.Duration
	// CleanupInterval defines how often the cache is scrubbed (i.e. when stale requests are removed).
	// If zero, stale requests are never removed.
	CleanupInterval time.Duration
	// GetKey returns the cache key for a request. By default, the request's Path & Method are used.
	GetKey func(*http.Request) string
	// Metrics contains the Prometheus metrics for the cache. The caller must register the requests with a Prometheus registry.
	// If nil, no metrics are collected.
	CacheMetrics
}

// WithCache creates a RoundTripper that caches the HTTP responses. See CacheOptions for the available options.
func WithCache(options CacheOptions) Option {
	if options.GetKey == nil {
		options.GetKey = func(r *http.Request) string {
			return r.Method + "|" + r.URL.Path
		}
	}
	c := responseCache{
		table: options.CacheTable,
		cache: cache.New[string, []byte](options.DefaultExpiration, options.CleanupInterval),
	}
	c.table.mustCompile()

	return func(next http.RoundTripper) http.RoundTripper {
		c := cachingRoundTripper{
			next:    next,
			cache:   c,
			getKey:  options.GetKey,
			metrics: options.CacheMetrics,
		}
		return &c
	}
}

var _ http.RoundTripper = &cachingRoundTripper{}

type cachingRoundTripper struct {
	next    http.RoundTripper
	cache   responseCache
	getKey  func(*http.Request) string
	metrics CacheMetrics
}

func (c *cachingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cacheKey := c.getKey(request)
	response, found, err := c.cache.get(cacheKey, request)
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

// CacheTable defines which responses are cached. If table is empty, all responses will be cached.
type CacheTable []*CacheTableEntry

// DefaultCacheTable caches all requests.
var DefaultCacheTable CacheTable

func (c CacheTable) shouldCache(r *http.Request) (bool, time.Duration) {
	if len(c) == 0 {
		return true, 0
	}

	for _, entry := range c {
		if match, expiry := entry.shouldCache(r); match {
			return match, expiry
		}
	}
	return false, 0
}

func (c CacheTable) mustCompile() {
	for _, entry := range c {
		entry.mustCompile()
	}
}

// CacheTableEntry contains a single endpoint that should be cached. If the Path is a regular expression, IsRegExp must be true.
type CacheTableEntry struct {
	// Path is the URL Path for requests whose responses should be cached. If blank, all paths will be cached.
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

func (entry *CacheTableEntry) shouldCache(r *http.Request) (bool, time.Duration) {
	if entry.matchesPath(r) && entry.matchesMethod(r) {
		return true, entry.Expiry
	}
	return false, 0
}

func (entry *CacheTableEntry) matchesPath(r *http.Request) bool {
	if entry.Path == "" {
		return true
	}
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

func (c *responseCache) get(key string, req *http.Request) (resp *http.Response, ok bool, err error) {
	var body []byte
	if body, ok = c.cache.Get(key); ok {
		r := bufio.NewReader(bytes.NewReader(body))
		resp, err = http.ReadResponse(r, req)
	}
	return resp, ok, err
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

func NewCacheMetrics(namespace, subsystem, application string) CacheMetrics {
	var constLabels map[string]string
	if application != "" {
		constLabels = map[string]string{"application": application}
	}
	return &defaultCacheMetrics{
		cache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "http_cache_total"),
			Help:        "Number of times the cache was consulted",
			ConstLabels: constLabels,
		}, []string{"path", "method"}),
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(namespace, subsystem, "http_cache_hit_total"),
			Help:        "Number of times the cache was used",
			ConstLabels: constLabels,
		}, []string{"path", "method"}),
	}
}

var _ prometheus.Collector = &defaultCacheMetrics{}

func (m *defaultCacheMetrics) Measure(r *http.Request, hit bool) {
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	m.cache.WithLabelValues(path, r.Method).Inc()
	if hit {
		m.hits.WithLabelValues(path, r.Method).Inc()
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
