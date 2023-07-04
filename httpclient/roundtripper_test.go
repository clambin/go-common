package httpclient_test

import (
	"bytes"
	"errors"
	"github.com/clambin/go-common/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestWithRoundTripper(t *testing.T) {
	s := stubbedServer{}

	c := http.Client{Transport: httpclient.NewRoundTripper(httpclient.WithRoundTripper(&s))}
	_, err := c.Get("/")
	assert.NoError(t, err)
	assert.Equal(t, 1, s.called)

	c = http.Client{Transport: httpclient.NewRoundTripper(httpclient.WithRoundTripper(nil))}
	assert.Panics(t, func() {
		_, _ = c.Get("/")
	})
}

func TestRoundTripperFunc(t *testing.T) {
	f := func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent}, nil
	}
	tp := httpclient.NewRoundTripper(httpclient.WithRoundTripper(httpclient.RoundTripperFunc(f)))
	c := http.Client{Transport: tp}

	resp, err := c.Get("/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestCustom_RoundTripper(t *testing.T) {
	f := func(req *http.Request) (*http.Response, error) {
		statusCode := http.StatusNoContent
		if h := req.Header.Get("X-Chain"); h != "foo" {
			statusCode = http.StatusBadRequest
		}
		return &http.Response{StatusCode: statusCode}, nil
	}
	r := httpclient.NewRoundTripper(
		WithHeader(),
		httpclient.WithRoundTripper(httpclient.RoundTripperFunc(f)),
	)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := r.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func WithHeader() httpclient.Option {
	return func(current http.RoundTripper) http.RoundTripper {
		return &Header{next: current}
	}
}

type Header struct{ next http.RoundTripper }

func (f *Header) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("X-Chain", "foo")
	return f.next.RoundTrip(r)
}

var _ http.RoundTripper = &stubbedServer{}

type stubbedServer struct {
	delay       time.Duration
	fail        bool
	called      int
	inFlight    int
	maxInFlight int
	lock        sync.Mutex
}

func (s *stubbedServer) RoundTrip(_ *http.Request) (*http.Response, error) {
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

func (s *stubbedServer) inc() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.called++
	s.inFlight++
	if s.inFlight > s.maxInFlight {
		s.maxInFlight = s.inFlight
	}
}

func (s *stubbedServer) dec() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.inFlight--
}
