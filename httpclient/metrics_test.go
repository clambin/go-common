package httpclient_test

import (
	"bytes"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestWithMetrics(t *testing.T) {
	s := stubbedServer{}
	c := httpclient.NewRoundTripper(
		httpclient.WithMetrics("foo", "bar", "test"),
		httpclient.WithRoundTripper(&s),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := c.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NoError(t, testutil.CollectAndCompare(c, bytes.NewBufferString(`
# HELP foo_bar_api_errors_total Number of failed Reporter API calls
# TYPE foo_bar_api_errors_total counter
foo_bar_api_errors_total{application="test",method="GET",path="/"} 0
`), "foo_bar_api_errors_total"))

	s.fail = true
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	resp, err = c.RoundTrip(req)
	require.Error(t, err)

	// TODO: check latency summary?
	assert.NoError(t, testutil.CollectAndCompare(c, bytes.NewBufferString(`
# HELP foo_bar_api_errors_total Number of failed Reporter API calls
# TYPE foo_bar_api_errors_total counter
foo_bar_api_errors_total{application="test",method="GET",path="/"} 1
`), "foo_bar_api_errors_total"))
}
