package httpclient_test

import (
	"context"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRoundTripper(t *testing.T) {
	r := httpclient.NewRoundTripper()
	c := &http.Client{Transport: r}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	}))
	defer s.Close()

	req, _ := http.NewRequest(http.MethodGet, s.URL+"/", nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Hello", string(body))
	_ = resp.Body.Close()
}

func TestRoundTripper_WithCache(t *testing.T) {
	r := httpclient.NewRoundTripper(httpclient.WithCache(httpclient.DefaultCacheTable, time.Minute, 5*time.Minute))
	c := &http.Client{Transport: r}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	}))

	var date string
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, s.URL+"/foo", nil)
		resp, err := c.Do(req)
		require.NoError(t, err, i)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		if i == 0 {
			date = resp.Header["Date"][0]
		} else {
			assert.Equal(t, date, resp.Header["Date"][0])
		}

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "Hello", string(body))
		_ = resp.Body.Close()
	}

	s.Close()
	req, _ := http.NewRequest(http.MethodGet, s.URL+"/bar", nil)
	_, err := c.Do(req)
	assert.Error(t, err)
}

func TestRoundTripper_WithCache_Stress(t *testing.T) {
	var bigResponse string
	for i := 0; i < 10000; i++ {
		bigResponse += "A"
	}

	r := httpclient.NewRoundTripper(httpclient.WithCache(httpclient.DefaultCacheTable, time.Minute, 5*time.Minute))
	c := &http.Client{Transport: r}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(bigResponse))
	}))

	const maxParallel = 100
	const Iterations = 10000

	p := semaphore.NewWeighted(maxParallel)
	ctx := context.Background()
	for i := 0; i < Iterations; i++ {
		require.NoError(t, p.Acquire(ctx, 1))
		go func() {
			defer p.Release(1)
			req, _ := http.NewRequest(http.MethodGet, s.URL+"/foo", nil)
			resp, err := c.Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, bigResponse, string(body))

			_ = resp.Body.Close()
		}()
	}
	require.NoError(t, p.Acquire(ctx, maxParallel))
}

func TestRoundTripper_Collect(t *testing.T) {
	r := httpclient.NewRoundTripper(
		httpclient.WithMetrics("", "", "foo"),
		httpclient.WithRoundTripper(&stubbedRoundTripper{}),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(r)

	c := &http.Client{Transport: r}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "http://localhost:8080/invalid", nil)
	resp, err = c.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	metrics, err := registry.Gather()
	require.NoError(t, err)
	for _, metric := range metrics {
		switch metric.GetName() {
		case "api_errors_total":
			assert.Len(t, metric.GetMetric(), 2)
		case "api_latency":
			assert.Len(t, metric.GetMetric(), 2)
		case "api_cache_total":
			assert.Len(t, metric.GetMetric(), 2)
		case "api_cache_hit_total":
			assert.Len(t, metric.GetMetric(), 1)
		default:
			t.Log(metric.GetName())
		}
	}
}

func BenchmarkRoundTripper_RoundTrip(b *testing.B) {
	r := httpclient.NewRoundTripper(
		httpclient.WithCache(httpclient.DefaultCacheTable, time.Minute, 0),
		httpclient.WithMetrics("", "", "foo"),
		httpclient.WithRoundTripper(&stubbedRoundTripper{}),
	)

	c := &http.Client{Transport: r}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)

	for i := 0; i < b.N; i++ {
		resp, err := c.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatal(resp.Status)
		}
	}
}

type stubbedRoundTripper struct{}

func (r *stubbedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	statusCode := http.StatusNotFound
	if req.URL.Path == "/" {
		statusCode = http.StatusOK
	}
	return &http.Response{StatusCode: statusCode}, nil
}
