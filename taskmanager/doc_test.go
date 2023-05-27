package taskmanager_test

import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/go-common/taskmanager/httpserver"
	"net/http"
	"os"
	"os/signal"
)

func Example() {
	m := taskmanager.New()

	// Add an HTTP Server.
	r := http.NewServeMux()
	r.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	_ = m.Add(httpserver.New(":8080", r))

	// Add a Goroutine. We use TaskFunc to convert a func to a Task
	// without having to declare a struct that adheres to the Task interface.
	_ = m.Add(taskmanager.TaskFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}))

	// Run until the program is interrupted.
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt)
	defer done()

	// Start the task manager. This will run until the context is marked as done.
	if err := m.Run(ctx); err != nil {
		panic(err)
	}
}
