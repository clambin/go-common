package taskmanager

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestManager_Run(t *testing.T) {
	w1, w2 := &waiter{}, &waiter{}
	m := New(w1)
	assert.NoError(t, m.Add(w2))

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() {
		ch <- m.Run(ctx)
	}()

	require.Eventually(t, func() bool { return w1.getRunCounter() > 0 }, time.Second, time.Millisecond)
	require.Eventually(t, func() bool { return w2.getRunCounter() > 0 }, time.Second, time.Millisecond)

	assert.Error(t, m.Add(&waiter{}))

	cancel()
	assert.NoError(t, <-ch)
	assert.False(t, w1.running)
	assert.Equal(t, 1, w1.runCounter)
	assert.False(t, w2.running)
	assert.Equal(t, 1, w2.runCounter)
}

func TestManager_Run_Failing(t *testing.T) {
	m := New(&waiter{}, TaskFunc(func(_ context.Context) error {
		return errTaskFailed
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := m.Run(ctx)
	assert.ErrorIs(t, err, errTaskFailed)
}

func TestManager_Run_Twice(t *testing.T) {
	w1, w2 := &waiter{}, &waiter{}
	m := New(w1, w2)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() {
		ch <- m.Run(ctx)
	}()

	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		ch <- m.Run(ctx2)
	}()

	require.Eventually(t, func() bool { return w1.getRunCounter() > 0 }, time.Second, time.Millisecond)
	require.Eventually(t, func() bool { return w2.getRunCounter() > 0 }, time.Second, time.Millisecond)

	cancel2()
	cancel()
	//err := errors.Join(<-ch, <-ch)
	//require.ErrorIs(t, err, ErrAlreadyRunning)
	require.True(t, errors.Is(<-ch, ErrAlreadyRunning) || errors.Is(<-ch, ErrAlreadyRunning))
}

func TestManager_Run_Timeout(t *testing.T) {
	w1, w2 := &waiter{}, &waiter{}
	m := New(w1, w2)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := m.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, w1.runCounter)
	assert.Equal(t, 1, w2.runCounter)
}

type waiter struct {
	runCounter int
	running    bool
	lock       sync.Mutex
}

func (w *waiter) Run(ctx context.Context) error {
	w.lock.Lock()
	w.running = true
	w.runCounter++
	w.lock.Unlock()
	<-ctx.Done()
	w.lock.Lock()
	w.running = false
	w.lock.Unlock()
	return nil
}

func (w *waiter) getRunCounter() int {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.runCounter
}

var errTaskFailed = errors.New("task failed")
