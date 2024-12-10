package sema_test

import (
	"context"
	"errors"
	"github.com/clambin/go-common/httputils/roundtripper/internal/sema"
	"github.com/clambin/go-common/testutils"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	const maxCount = 1000
	s := sema.NewSema(maxCount)

	ctx := context.Background()
	for i := 0; i < maxCount; i++ {
		if err := s.Acquire(ctx); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	if err := s.Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got wrong error: %v, wanted: %v", err, context.DeadlineExceeded)
	}
	cancel()

	for i := 0; i < maxCount; i++ {
		s.Release()
	}
	if ok := testutils.Panics(s.Release); !ok {
		t.Error("expected panic")
	}
}
