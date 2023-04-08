package httpserver_test

import (
	"errors"
	"fmt"
	"github.com/clambin/go-common/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type endpoint struct {
	path   string
	method string
	result int
}

type testCase struct {
	name      string
	options   []httpserver.Option
	endpoints []endpoint
}

func TestServer_ServeHTTP(t *testing.T) {
	testCases := []testCase{
		{
			name: "prometheus only",
			options: []httpserver.Option{
				httpserver.WithPrometheus(""),
			},
			endpoints: []endpoint{
				{path: "/metrics", method: http.MethodGet, result: http.StatusOK},
				{path: "/foo", method: http.MethodGet, result: http.StatusNotFound},
			},
		},
		{
			name: "handlers only",
			options: []httpserver.Option{
				httpserver.WithHandlers([]httpserver.Handler{
					{
						Path: "/foo",
						Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							_, _ = w.Write([]byte("OK"))
						}),
						Methods: []string{http.MethodPost},
					},
				}),
			},
			endpoints: []endpoint{
				{path: "/foo", method: http.MethodPost, result: http.StatusOK},
				{path: "/foo", method: http.MethodGet, result: http.StatusMethodNotAllowed},
				{path: "/metrics", method: http.MethodGet, result: http.StatusNotFound},
			},
		},
		{
			name: "combined",
			options: []httpserver.Option{
				httpserver.WithPrometheus(""),
				httpserver.WithHandlers([]httpserver.Handler{
					{
						Path: "/foo",
						Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							_, _ = w.Write([]byte("OK"))
						}),
					},
				}),
			},
			endpoints: []endpoint{
				{path: "/foo", method: http.MethodGet, result: http.StatusOK},
				{path: "/foo", method: http.MethodPost, result: http.StatusMethodNotAllowed},
				{path: "/metrics", method: http.MethodGet, result: http.StatusOK},
				{path: "/metrics", method: http.MethodPost, result: http.StatusMethodNotAllowed},
			},
		},
		{
			name: "fixed port",
			options: []httpserver.Option{
				httpserver.WithAddr(":8080"),
				httpserver.WithPrometheus(""),
			},
			endpoints: []endpoint{
				{path: "/metrics", method: http.MethodGet, result: http.StatusOK},
				{path: "/metrics", method: http.MethodPost, result: http.StatusMethodNotAllowed},
				{path: "/foo", method: http.MethodGet, result: http.StatusNotFound},
			},
		},
	}

	var wg sync.WaitGroup
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := httpserver.New(tt.options...)
			require.NoError(t, err)

			for _, ep := range tt.endpoints {
				req, _ := http.NewRequest(ep.method, ep.path, nil)
				resp := httptest.NewRecorder()

				s.ServeHTTP(resp, req)

				assert.Equal(t, ep.result, resp.Code, ep.path)
			}
		})
	}
	wg.Wait()
}

func TestServer_ServerHTTP_WithMetrics(t *testing.T) {
	testCases := []struct {
		name         string
		metrics      httpserver.Option
		evalCount    func(t *testing.T, r prometheus.Gatherer)
		evalDuration func(t *testing.T, r prometheus.Gatherer)
	}{
		{
			name:         "histogram",
			metrics:      httpserver.WithMetrics("", "", "foobar", httpserver.Histogram, nil),
			evalCount:    evalRequestsCounter,
			evalDuration: evalDurationHistogram,
		},
		{
			name:         "summary",
			metrics:      httpserver.WithMetrics("", "", "foobar", httpserver.Summary, nil),
			evalCount:    evalRequestsCounter,
			evalDuration: evalDurationSummary,
		},
	}

	var wg sync.WaitGroup
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := httpserver.New(
				httpserver.WithHandlers([]httpserver.Handler{{
					Path: "/foo",
					Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte("OK"))
					}),
				}}),
				tt.metrics,
			)
			require.NoError(t, err)
			r := prometheus.NewRegistry()
			r.MustRegister(s)

			req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
			resp := httptest.NewRecorder()

			s.ServeHTTP(resp, req)
			assert.Equal(t, resp.Code, http.StatusOK)

			if tt.evalCount != nil {
				tt.evalCount(t, r)
			}

			if tt.evalDuration != nil {
				tt.evalDuration(t, r)
			}
		})
	}

	wg.Wait()

}

func TestServer_Serve(t *testing.T) {
	s, err := httpserver.New(httpserver.WithHandlers([]httpserver.Handler{{
		Path: "/",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("OK"))
		}),
	}}))
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err2 := s.Serve()
		require.True(t, errors.Is(err2, http.ErrServerClosed))
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", s.GetPort()))
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	_ = s.Shutdown(time.Minute)
	wg.Wait()
}

func TestServer_Run_BadPort(t *testing.T) {
	_, err := httpserver.New(httpserver.WithAddr(":-1"))
	assert.Error(t, err)
}

type metricInfo struct {
	metric *io_prometheus_client.Metric
	labels map[string]string
}

func getMetricInfo(t *testing.T, g prometheus.Gatherer, name string) (output []metricInfo) {
	t.Helper()

	metrics, err := g.Gather()
	require.NoError(t, err)

	for _, metric := range metrics {
		if metric.GetName() != name {
			continue
		}
		for _, m := range metric.GetMetric() {
			info := metricInfo{
				metric: m,
				labels: make(map[string]string),
			}
			for _, l := range m.GetLabel() {
				info.labels[l.GetName()] = l.GetValue()
			}
			output = append(output, info)
		}
	}
	return output
}

func evalRequestsCounter(t *testing.T, r prometheus.Gatherer) {
	t.Helper()
	metrics := getMetricInfo(t, r, "http_requests_total")
	require.Len(t, metrics, 1)
	assert.Equal(t, 1.0, metrics[0].metric.GetCounter().GetValue())

}

func evalDurationHistogram(t *testing.T, r prometheus.Gatherer) {
	t.Helper()
	metrics := getMetricInfo(t, r, "http_requests_duration_seconds")
	require.Len(t, metrics, 1)
	assert.Len(t, metrics[0].labels, 3)
	assert.Equal(t, "foobar", metrics[0].labels["handler"])

	assert.Equal(t, uint64(1), metrics[0].metric.GetHistogram().GetSampleCount())
	assert.Len(t, metrics[0].labels, 3)
	assert.Equal(t, "foobar", metrics[0].labels["handler"])
}

func evalDurationSummary(t *testing.T, r prometheus.Gatherer) {
	t.Helper()
	metrics := getMetricInfo(t, r, "http_requests_duration_seconds")
	require.Len(t, metrics, 1)
	assert.Len(t, metrics[0].labels, 3)
	assert.Equal(t, "foobar", metrics[0].labels["handler"])

	assert.Equal(t, uint64(1), metrics[0].metric.GetSummary().GetSampleCount())
	assert.Len(t, metrics[0].labels, 3)
	assert.Equal(t, "foobar", metrics[0].labels["handler"])
}
