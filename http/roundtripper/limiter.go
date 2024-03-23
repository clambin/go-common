package roundtripper

import (
	"fmt"
	"github.com/clambin/go-common/http/roundtripper/internal/sema"
	"net/http"
)

var _ http.RoundTripper = &limiter{}

type limiter struct {
	next     http.RoundTripper
	parallel *sema.Semaphore
}

// WithLimiter creates a RoundTripper that limits the number concurrent http requests to maxParallel.
func WithLimiter(maxParallel int64) Option {
	return func(next http.RoundTripper) http.RoundTripper {
		return &limiter{
			next:     next,
			parallel: sema.NewSema(int(maxParallel)),
		}
	}
}

func (l *limiter) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := l.parallel.Acquire(request.Context()); err != nil {
		return nil, fmt.Errorf("acquire semaphore: %w", err)
	}
	defer l.parallel.Release()
	return l.next.RoundTrip(request)
}
