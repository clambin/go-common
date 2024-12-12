package pubsub_test

import (
	"github.com/clambin/go-common/pubsub"
	"testing"
)

func TestPublisher(t *testing.T) {
	var p pubsub.Publisher[int]

	ch := p.Subscribe()
	defer p.Unsubscribe(ch)

	const count = 10000
	go func(n int) {
		for i := range n {
			p.Publish(i)
		}
	}(count)

	for i := range count {
		if val := <-ch; i != val {
			t.Errorf("got %d, want %d", val, i)
		}
	}
}
