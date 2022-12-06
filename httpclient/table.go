package httpclient

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// CacheTable holds the Endpoints that should be cached. If Table is empty, all responses will be cached.
type CacheTable struct {
	Table    []CacheTableEntry
	compiled bool
	lock     sync.Mutex
}

func (c *CacheTable) shouldCache(r *http.Request) (match bool, expiry time.Duration) {
	if len(c.Table) == 0 {
		return true, 0
	}

	c.compileIfNeeded()

	for _, entry := range c.Table {
		if match, expiry = entry.shouldCache(r); match {
			return
		}
	}
	return
}

func (c *CacheTable) compileIfNeeded() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.compiled {
		return
	}

	for index := range c.Table {
		if c.Table[index].IsRegExp {
			var err error
			c.Table[index].compiledRegExp, err = regexp.Compile(c.Table[index].Endpoint)
			if err != nil {
				panic(fmt.Errorf("cacheTable: invalid regexp '%s': %w", c.Table[index].Endpoint, err))
			}
		}
	}
	c.compiled = true
}

// CacheTableEntry contains a single endpoint that should be cached. If the Endpoint is a regular expression, IsRegExp must be set.
// CacheTable will then compile it when needed. CacheTable will panic if the regular expression is invalid.
type CacheTableEntry struct {
	// Endpoint is the URL Path for requests whose responses should be cached.
	// Can be a literal path, or a regular expression. In the latter case,
	// set IsRegExp to true
	Endpoint string
	// Methods is the list of HTTP Methods for which requests the response should be cached.
	// If empty, requests for any method will be cached.
	Methods []string
	// IsRegExp indicated the Endpoint is a regular expression.
	// Note: CacheTableEntry will panic if Endpoint does not contain a valid regular expression.
	IsRegExp bool
	// Expiry indicates how long a response should be cached.
	Expiry         time.Duration
	compiledRegExp *regexp.Regexp
}

// var CacheEverything []CacheTableEntry

func (entry CacheTableEntry) shouldCache(r *http.Request) (match bool, expiry time.Duration) {
	match = entry.matchesEndpoint(r)
	if !match {
		return
	}
	match = entry.matchesMethods(r)
	return match, entry.Expiry
}

func (entry CacheTableEntry) matchesEndpoint(r *http.Request) bool {
	endpoint := r.URL.Path
	if entry.IsRegExp {
		return entry.compiledRegExp.MatchString(endpoint)
	}
	return entry.Endpoint == endpoint
}

func (entry CacheTableEntry) matchesMethods(r *http.Request) bool {
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
