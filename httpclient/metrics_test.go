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
	cfg := &Metrics{}

	// makeLatencyTimer returns nil if no latency metric is set
	assert.Nil(t, cfg.makeLatencyTimer())

	r := prometheus.NewRegistry()
	cfg = NewMetrics("foo", "")
	r.MustRegister(cfg)

	// collect metrics
	timer := cfg.makeLatencyTimer("foo", "/bar", http.MethodGet)
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
	cfg := &Metrics{}

	// reportErrors doesn't crash when no errors metric is set
	cfg.reportErrors(nil)

	r := prometheus.NewRegistry()
	cfg = NewMetrics("bar", "")
	r.MustRegister(cfg)

	// collect metrics
	cfg.reportErrors(nil, "foo", "/bar", http.MethodGet)

	// do a measurement
	count := getErrorMetrics(t, r, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 0}, count)

	// record an error
	cfg.reportErrors(errors.New("some error"), "foo", "/bar", http.MethodGet)

	// counter should now be 1
	count = getErrorMetrics(t, r, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 1}, count)
}

func TestClientMetrics_Nil(t *testing.T) {
	cfg := Metrics{}

	timer := cfg.makeLatencyTimer("snafu")
	assert.Nil(t, timer)
	cfg.reportErrors(nil, "foo")
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
				counters[*metric.Label[1].Value] = *metric.Counter.Value
			}
		}
	}
	return counters
}
