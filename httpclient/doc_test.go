package httpclient_test

import (
	"fmt"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

func Example_withMetrics() {
	transport := httpclient.NewRoundTripper(httpclient.WithRoundTripperMetrics{Application: "example"})
	client := &http.Client{
		Transport: transport,
	}

	prometheus.DefaultRegisterer.MustRegister(transport)

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func Example_withCache() {
	c := &http.Client{
		Transport: httpclient.NewRoundTripper(httpclient.WithCache{
			Table: []*httpclient.CacheTableEntry{
				{
					Path:     "/foo/.+",
					IsRegExp: true,
					Expiry:   5 * time.Second,
				},
			},
		}),
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if resp, err := c.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}
