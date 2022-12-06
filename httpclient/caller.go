package httpclient

import (
	"net/http"
)

// Caller interface of a generic API caller
//
//go:generate mockery --name Caller
type Caller interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// BaseClient performs the actual HTTP request
type BaseClient struct {
	HTTPClient *http.Client
}

var _ Caller = &BaseClient{}

// Do performs the actual HTTP request
func (b *BaseClient) Do(req *http.Request) (resp *http.Response, err error) {
	httpClient := b.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return httpClient.Do(req)
}
