package httpclient

import (
	"bytes"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
	cfg.reportErrors(errors.New("some error"), "/bar", http.MethodGet)

	err := testutil.GatherAndCompare(r, bytes.NewBufferString(`# HELP bar_api_errors_total Number of failed Reporter API calls
# TYPE bar_api_errors_total counter
bar_api_errors_total{application="test",method="GET",path="/bar"} 1
`))
	assert.NoError(t, err)

}

func TestClientMetrics_Cache(t *testing.T) {
	r := prometheus.NewRegistry()
	cfg := newMetrics("foo", "", "test")
	r.MustRegister(cfg)

	cfg.reportCache(false, "/bar", http.MethodGet)

	err := testutil.GatherAndCompare(r, bytes.NewBufferString(`# HELP foo_api_cache_total Number of times the cache was consulted
# TYPE foo_api_cache_total counter
foo_api_cache_total{application="test",method="GET",path="/bar"} 1
`))
	assert.NoError(t, err)

	cfg.reportCache(true, "/bar", http.MethodGet)

	err = testutil.GatherAndCompare(r, bytes.NewBufferString(`# HELP foo_api_cache_hit_total Number of times the cache was used
# TYPE foo_api_cache_hit_total counter
foo_api_cache_hit_total{application="test",method="GET",path="/bar"} 1
# HELP foo_api_cache_total Number of times the cache was consulted
# TYPE foo_api_cache_total counter
foo_api_cache_total{application="test",method="GET",path="/bar"} 2
`))
	assert.NoError(t, err)

}
