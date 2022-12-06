package httpclient_test

import (
	"fmt"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

func ExampleInstrumentedClient() {
	metrics := httpclient.NewMetrics("foo", "bar")
	prometheus.DefaultRegisterer.MustRegister(metrics)

	c := httpclient.InstrumentedClient{
		Options:     httpclient.Options{PrometheusMetrics: metrics},
		Application: "test",
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if resp, err := c.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func ExampleCacher() {
	metrics := httpclient.NewMetrics("foo", "bar")
	prometheus.DefaultRegisterer.MustRegister(metrics)

	table := []httpclient.CacheTableEntry{
		{
			Endpoint: "/foo/.+",
			IsRegExp: true,
			Expiry:   5 * time.Second,
		},
	}

	c := httpclient.NewCacher(nil, "test", httpclient.Options{PrometheusMetrics: metrics}, table, time.Minute, time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if resp, err := c.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}
