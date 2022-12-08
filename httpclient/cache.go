package httpclient

import (
	"bufio"
	"bytes"
	"github.com/clambin/go-common/cache"
	"net/http"
	"net/http/httputil"
	"time"
)

// Cache will cache calls based in the provided CacheTable
type Cache struct {
	Table CacheTable
	cache cache.Cacher[string, []byte]
}

func NewCache(table CacheTable, defaultExpiry, cleanupInterval time.Duration) *Cache {
	c := Cache{
		Table: table,
		cache: cache.New[string, []byte](defaultExpiry, cleanupInterval),
	}
	c.Table.compile()
	return &c
}

func (c *Cache) Get(req *http.Request) (string, *http.Response, bool, error) {
	key := getCacheKey(req)
	body, found := c.cache.Get(key)

	var resp *http.Response
	var err error
	if found {
		resp, err = http.ReadResponse(bufio.NewReader(bytes.NewBuffer(body)), req)
		return key, resp, found, err
	}
	return key, resp, found, err
}

func (c *Cache) Put(key string, req *http.Request, resp *http.Response) error {
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

func (c *Cache) shouldCache(r *http.Request) (cache bool, expiry time.Duration) {
	cache, expiry = c.Table.shouldCache(r)
	if cache && expiry == 0 {
		expiry = c.cache.GetDefaultExpiration()
	}
	return
}

func getCacheKey(r *http.Request) string {
	return r.URL.Path
}
