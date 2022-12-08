package httpclient

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	pcg "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestClientMetrics_MakeLatencyTimer(t *testing.T) {
	r := prometheus.NewRegistry()
	cfg := newMetrics("foo", "", "test")
	r.MustRegister(cfg)

	// collect metrics
	timer := cfg.makeLatencyTimer("/bar", http.MethodGet)
	require.NotNil(t, timer)
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	// one measurement should be collected
	m, err := r.Gather()
	require.NoError(t, err)
	var found bool
	for _, entry := range m {
		if *entry.Name == "foo_api_latency" {
			require.Equal(t, pcg.MetricType_SUMMARY, *entry.Type)
			require.NotZero(t, entry.Metric)
			assert.NotZero(t, entry.Metric[0].Summary.GetSampleCount())
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestClientMetrics_ReportErrors(t *testing.T) {
	r := prometheus.NewRegistry()
	cfg := newMetrics("bar", "", "test")
	r.MustRegister(cfg)

	// collect metrics
	cfg.reportErrors(nil, "/bar", http.MethodGet)

	// do a measurement
	count := getErrorMetrics(t, r, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 0}, count)

	// record an error
	cfg.reportErrors(errors.New("some error"), "/bar", http.MethodGet)

	// counter should now be 1
	count = getErrorMetrics(t, r, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 1}, count)
}

func getErrorMetrics(t *testing.T, g prometheus.Gatherer, prefix string) map[string]float64 {
	t.Helper()

	counters := make(map[string]float64)
	m, err := g.Gather()
	require.NoError(t, err)
	for _, entry := range m {
		if *entry.Name == prefix+"api_errors_total" {
			require.Equal(t, pcg.MetricType_COUNTER, *entry.Type)
			for _, metric := range entry.Metric {
				for _, label := range metric.GetLabel() {
					if *label.Name == "path" {
						counters[*label.Value] = *metric.Counter.Value
						break
					}
				}
			}
		}
	}
	return counters
}
