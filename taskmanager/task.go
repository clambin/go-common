package taskmanager

import (
	"context"
)

// Task is the interface that any taskmanager task must adhere to. It consists of a single Run method. The task should
// run until the provided context is marked as Done, or a fatal error occurs.
type Task interface {
	Run(ctx context.Context) error
}

// The TaskFunc type is an adapter to allow the use of ordinary functions as a Task. If f is a function with the appropriate signature,
// TaskFunc(f) is a Task that calls f.
type TaskFunc func(ctx context.Context) error

// Run calls f(ctx).
func (f TaskFunc) Run(ctx context.Context) error {
	return f(ctx)
}
