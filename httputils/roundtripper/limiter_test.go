package roundtripper_test

import (
	"context"
	"errors"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/go-common/testutils"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLimiter_RoundTrip(t *testing.T) {
	s := server{delay: 10 * time.Millisecond}

	const maxParallel = 10
	c := http.Client{
		Transport: roundtripper.WithLimiter(maxParallel)(&s),
	}

	var eg errgroup.Group
	for range 100 {
		eg.Go(func() error {
			_, err := c.Get("/")
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	if got := int(s.maxInFlight.Load()); got != maxParallel {
		t.Errorf("got %d, want %d", got, maxParallel)
	}
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
	if ok := testutils.Eventually(func() bool { return s.called.Load() > 0 }, time.Second, 10*time.Millisecond); !ok {
		t.Error("condition never satisfied")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded error, got %v", err)
	}
}

// Current:
// BenchmarkWithLimiter-16    	  191450	     12661 ns/op	   44680 B/op	       4 allocs/op
func BenchmarkWithLimiter(b *testing.B) {
	rt := roundtripper.WithLimiter(100)(roundtripper.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(strings.Repeat("hello\n", 10_000)))}, nil
	}))

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	var wg sync.WaitGroup

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rt.RoundTrip(req)
		}()
	}
	wg.Wait()
}
