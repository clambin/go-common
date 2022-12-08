/*
Package httpclient provides a standard way of writing API clients.
TODO!
It's meant to be a drop-in replacement for HTTPClient's Transport .
Currently, it supports generating Prometheus metrics when performing API calls, and caching API responses.

InstrumentedClient generates Prometheus metrics when performing API calls. Currently, it records request latency and errors.

Cache caches responses to HTTP requests, based on the provided CacheTableEntry slice. If the slice is empty, all responses will be cached.

Note: NewCacher will create a Caller that also generates Prometheus metrics by chaining the request to an InstrumentedClient.
To avoid this, create a Cache object directly:

	c := &httpclient.Cache{
		Caller: &httpclient.BaseClient{},
		Table: httpclient.CacheTable{Table: cacheEntries},
		Cache: cache.New[string, []byte](cacheExpiry, cacheCleanup),
*/
package httpclient
