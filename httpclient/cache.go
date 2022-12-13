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
	cache cache.Cacher[string, []byte]
}

func newCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) *responseCache {
	c := responseCache{
		table: table,
		cache: cache.New[string, []byte](defaultExpiry, cleanupInterval),
	}
	c.table.compile()
	return &c
}

func (c *responseCache) get(req *http.Request) (string, *http.Response, bool, error) {
	key := getCacheKey(req)
	body, found := c.cache.Get(key)
	if !found {
		return key, nil, false, nil
	}

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(body)), req)
	if err == nil {
		resp.Request = req
	}
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

func (c *responseCache) shouldCache(r *http.Request) (cache bool, expiry time.Duration) {
	cache, expiry = c.table.shouldCache(r)
	if cache && expiry == 0 {
		expiry = c.cache.GetDefaultExpiration()
	}
	return
}

func getCacheKey(r *http.Request) string {
	return r.URL.Path
}
