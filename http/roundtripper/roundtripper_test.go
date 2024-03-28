package roundtripper_test

import (
	"bytes"
	"errors"
	"github.com/clambin/go-common/http/roundtripper"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRoundTripper(t *testing.T) {
	s := server{}

	c := http.Client{Transport: roundtripper.New(roundtripper.WithRoundTripper(&s))}
	if _, err := c.Get("/"); err != nil {
		t.Fatal(err)
	}
	if got := int(s.called.Load()); got != 1 {
		t.Errorf("s.called.Load() = %d, want 1", got)
	}
}

func TestRoundTripperFunc(t *testing.T) {
	f := func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent}, nil
	}
	tp := roundtripper.New(roundtripper.WithRoundTripper(roundtripper.RoundTripperFunc(f)))
	c := http.Client{Transport: tp}

	resp, err := c.Get("/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

type server struct {
	delay       time.Duration
	fail        bool
	called      atomic.Int32
	inFlight    atomic.Int32
	maxInFlight atomic.Int32
}

func (s *server) RoundTrip(_ *http.Request) (*http.Response, error) {
	s.inc()
	defer s.dec()

	time.Sleep(s.delay)

	if s.fail {
		return nil, errors.New("failed")
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`hello`)),
	}, nil
}

func (s *server) inc() {
	s.called.Add(1)
	s.inFlight.Add(1)
	s.maxInFlight.Store(max(s.inFlight.Load(), s.maxInFlight.Load()))
}

func (s *server) dec() {
	s.inFlight.Add(-1)
}
