package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Manager groups a number of Task instances
type Manager struct {
	lock    sync.Mutex
	tasks   []Task
	running bool
}

// ErrAlreadyRunning indicates the Manager is already running and the requested function cannot be performed.
var ErrAlreadyRunning = errors.New("task manager already running")

// New returns a Manager for the specified Task objects.
func New(tasks ...Task) *Manager {
	return &Manager{tasks: tasks}
}

// Add adds a Task to the Manager.  This can only be done when the Manager is not running.
// Otherwise, it returns ErrAlreadyRunning.
func (m *Manager) Add(task ...Task) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.running {
		return ErrAlreadyRunning
	}

	m.tasks = append(m.tasks, task...)
	return nil
}

// Run starts the different Task objects as separate Goroutines and waits for the context to be marked as Done.
// It then stops all Tasks and returns any returned errors.
//
// If any Task returns an error before the context is marked as Done, Run will stop all Goroutines and return
// the error of the failing Task.
//
// Only one instance of a Manager can run at a single time. If Run is called while the Manager is already running,
// Run returns ErrAlreadyRunning.
func (m *Manager) Run(ctx context.Context) error {
	if err := m.setRunning(true); err != nil {
		return ErrAlreadyRunning
	}

	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan error)
	m.startTasks(subCtx, ch)

	select {
	case <-ctx.Done():
		_ = m.setRunning(false)
		return m.shutdown(ch)
	case err := <-ch:
		return err
	}
}

func (m *Manager) startTasks(ctx context.Context, ch chan error) {
	for _, task := range m.tasks {
		go func(t Task) {
			ch <- t.Run(ctx)
		}(task)
	}
}

func (m *Manager) shutdown(ch chan error) error {
	var errs []error
	for range m.tasks {
		errs = append(errs, <-ch)
	}
	// while we're supporting pre-1.20
	//return errors.Join(errs...)
	return joinErrors(errs...)
}

func (m *Manager) setRunning(running bool) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.running == running {
		return fmt.Errorf("manager running state already %v", running)
	}
	m.running = running
	return nil
}
