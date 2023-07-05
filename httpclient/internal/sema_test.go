package internal_test

import (
	"context"
	"github.com/clambin/go-common/httpclient/internal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	const maxCount = 1000
	s := internal.NewSema(maxCount)

	ctx := context.Background()
	for i := 0; i < maxCount; i++ {
		assert.NoError(t, s.Acquire(ctx))
	}

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	err := s.Acquire(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	cancel()

	for i := 0; i < maxCount; i++ {
		s.Release()
	}
	assert.Panics(t, func() { s.Release() })

	ctx = context.Background()
	assert.NoError(t, s.Acquire(ctx))
	s.Release()
}
