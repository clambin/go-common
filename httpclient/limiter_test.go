package httpclient_test

import (
	"bytes"
	"context"
	"github.com/clambin/go-common/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestLimiter_RoundTrip(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()

	const maxParallel = 10
	s := stubbedServer{delay: 10 * time.Millisecond}
	r := httpclient.NewRoundTripper(
		httpclient.WithInstrumentedLimiter(maxParallel, "foo", "bar", "snafu"),
		httpclient.WithRoundTripper(&s),
	)
	c := http.Client{Transport: r}

	reg.MustRegister(r)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Get("/")
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, maxParallel, s.maxInFlight)

	assert.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP foo_bar_api_max_inflight Maximum number of requests in flight
# TYPE foo_bar_api_max_inflight gauge
foo_bar_api_max_inflight{application="snafu"} 10
`), "foo_bar_max_inflight"))
}

func TestLimiter_RoundTrip_Exceeded(t *testing.T) {
	s := stubbedServer{delay: time.Second}
	r := httpclient.NewRoundTripper(
		httpclient.WithLimiter(1),
		httpclient.WithRoundTripper(&s),
	)
	c := http.Client{Transport: r}

	go func() {
		_, _ = c.Get("/")
	}()

	// wait for the first request to reach the server
	assert.Eventually(t, func() bool {
		s.lock.Lock()
		defer s.lock.Unlock()

		return s.called > 0
	}, time.Second, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	_, err := c.Do(req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
