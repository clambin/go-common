package httpclient

import (
	"bufio"
	"bytes"
	"github.com/clambin/go-common/cache"
	"net/http"
	"net/http/httputil"
	"time"
)

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

/*
var readerPool = sync.Pool{
	NewRoundTripper: func() any { return bufio.NewReader(nil) },
}
*/

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

func getCacheKey(r *http.Request) string {
	return r.Method + " | " + r.URL.Path
}
