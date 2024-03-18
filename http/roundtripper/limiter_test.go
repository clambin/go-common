package roundtripper_test

import (
	"bytes"
	"context"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestLimiter_RoundTrip(t *testing.T) {
	s := server{delay: 10 * time.Millisecond}

	const maxParallel = 10
	metrics := roundtripper.NewLimiterMetrics("foo", "bar")
	c := http.Client{
		Transport: roundtripper.WithInstrumentedLimiter(maxParallel, metrics)(&s),
	}

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

	assert.NoError(t, testutil.CollectAndCompare(metrics, bytes.NewBufferString(`
# HELP foo_bar_api_inflight_current current in flight requests
# TYPE foo_bar_api_inflight_current gauge
foo_bar_api_inflight_current 0

# HELP foo_bar_api_inflight_max maximum in flight requests
# TYPE foo_bar_api_inflight_max gauge
foo_bar_api_inflight_max 10
`)))
}

func TestLimiter_RoundTrip_Exceeded(t *testing.T) {
	s := server{delay: time.Second}
	r := roundtripper.WithLimiter(1)(&s)
	c := http.Client{Transport: r}

	go func() {
		_, _ = c.Get("/")
	}()

	// wait for the first request to reach the server
	// subsequent requests will block on the Limiter's semaphore
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

func BenchmarkWithLimiter(b *testing.B) {
	var body bytes.Buffer
	for i := 0; i < 10000; i++ {
		body.WriteString("hello\n")
	}
	rt := roundtripper.WithLimiter(50)(roundtripper.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(&body)}, nil

	}))

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rt.RoundTrip(req)
		}()
	}
	wg.Wait()
}
