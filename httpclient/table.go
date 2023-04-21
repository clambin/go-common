package httpclient

import (
	"fmt"
	"net/http"
	"regexp"
	"time"
)

// CacheTable holds the endpoints that should be cached. If table is empty, all responses will be cached.
type CacheTable []*CacheTableEntry

// DefaultCacheTable is a CacheTable that caches all requests.
var DefaultCacheTable []*CacheTableEntry

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
