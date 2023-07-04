package internal_test

import (
	"context"
	"github.com/clambin/go-common/httpclient/internal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	s := internal.NewSema(3)

	ctx := context.Background()
	assert.NoError(t, s.Acquire(ctx))
	assert.NoError(t, s.Acquire(ctx))
	assert.NoError(t, s.Acquire(ctx))

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	err := s.Acquire(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	cancel()

	s.Release()
	s.Release()
	s.Release()

	assert.Panics(t, func() { s.Release() })
}
