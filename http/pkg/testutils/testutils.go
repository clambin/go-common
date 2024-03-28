// Package testutils contains candidates to be moved to go-common/testutils
package testutils

import (
	"context"
	"time"
)

func Panics(f func()) bool {
	var didPanic bool
	func(f func()) {
		defer func() {
			if err := recover(); err != nil {
				didPanic = true
			}
		}()
		f()
	}(f)
	return didPanic
}

func Eventually(f func() bool, timeout time.Duration, interval time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ch := make(chan bool)
	go func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				ch <- false
				return
			case <-ticker.C:
				if ok := f(); ok {
					ch <- true
					return
				}
			}
		}
	}(ctx)

	return <-ch
}
