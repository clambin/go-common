package internal

import (
	"context"
)

type Semaphore struct {
	ch chan struct{}
}

func NewSema(max int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, max)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	if len(s.ch) < 1 {
		panic("releasing non-acquired semaphore")
	}
	<-s.ch
}
