package httpclient_test

import (
	"bytes"
	"fmt"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

func ExampleWithMetrics() {
	transport := httpclient.NewRoundTripper(
		httpclient.WithMetrics("", "", "example"),
	)
	client := &http.Client{
		Transport: transport,
	}

	// The RoundTripper needs to be registered with a Prometheus registry so Prometheus will collect it.
	prometheus.DefaultRegisterer.MustRegister(transport)

	if resp, err := client.Get("https://example.com"); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func ExampleWithCache() {
	cacheTable := []*httpclient.CacheTableEntry{
		{
			Path:     "/foo/.+",
			IsRegExp: true,
			Expiry:   5 * time.Second,
		},
	}
	c := &http.Client{
		Transport: httpclient.NewRoundTripper(httpclient.WithCache(cacheTable, time.Second, time.Minute)),
	}

	if resp, err := c.Get("https://example.com"); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func ExampleRoundTripperFunc() {
	stub := func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`hello`)),
		}, nil
	}
	tp := httpclient.NewRoundTripper(httpclient.WithRoundTripper(httpclient.RoundTripperFunc(stub)))

	c := http.Client{Transport: tp}

	if resp, err := c.Get("/"); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func Example_chained() {
	tp := httpclient.NewRoundTripper(
		httpclient.WithCache(httpclient.DefaultCacheTable, time.Second, time.Minute),
		httpclient.WithMetrics("foo", "bar", "example"),
	)

	// The RoundTripper needs to be registered with a Prometheus registry so Prometheus will collect it.
	prometheus.MustRegister(tp)

	c := http.Client{Transport: tp}

	if resp, err := c.Get("https://example.com"); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}
