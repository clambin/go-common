package httpclient_test

import (
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	r := httpclient.NewRoundTripper(
		httpclient.WithCache{
			Table:           httpclient.CacheTable{},
			DefaultExpiry:   time.Minute,
			CleanupInterval: 5 * time.Minute,
		})
	c := &http.Client{Transport: r}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	}))

	var date string
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, s.URL+"/foo", nil)
		resp, err := c.Do(req)
		require.NoError(t, err)
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

func TestRoundTripper_Collect(t *testing.T) {
	r := httpclient.NewRoundTripper(
		httpclient.WithCache{},
		httpclient.WithRoundTripperMetrics{Application: "foo"},
		httpclient.WithRoundTripper{RoundTripper: &stubbedRoundTripper{}},
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(r)

	c := &http.Client{Transport: r}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "this is not the server you're looking for", http.StatusNotFound)
	}))
	defer s.Close()

	req, _ := http.NewRequest(http.MethodGet, s.URL+"/", nil)
	resp, err := c.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodGet, s.URL+"/invalid", nil)
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

type stubbedRoundTripper struct{}

func (r *stubbedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	statusCode := http.StatusNotFound
	if req.URL.Path == "/" {
		statusCode = http.StatusOK
	}
	return &http.Response{StatusCode: statusCode}, nil
}
