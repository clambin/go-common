package roundtripper_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/clambin/go-common/http/pkg/testutils"
	"github.com/clambin/go-common/http/roundtripper"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
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
	for i := 0; i < 100; i++ {
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

func BenchmarkWithLimiter(b *testing.B) {
	var body bytes.Buffer
	for i := 0; i < 10000; i++ {
		body.WriteString("hello\n")
	}
	rt := roundtripper.WithLimiter(100)(roundtripper.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(&body)}, nil

	}))

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
	var wg sync.WaitGroup

	b.ResetTimer()
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			_, _ = rt.RoundTrip(req)
		}()
	}
	wg.Wait()
}
