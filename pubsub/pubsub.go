package pubsub

import "sync"

type Publisher[T any] struct {
	clients map[chan T]struct{}
	lock    sync.RWMutex
}

func (p *Publisher[T]) Subscribe() chan T {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.clients == nil {
		p.clients = make(map[chan T]struct{})
	}
	ch := make(chan T)
	p.clients[ch] = struct{}{}
	return ch
}

func (p *Publisher[T]) Unsubscribe(ch chan T) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.clients, ch)
}

func (p *Publisher[T]) Publish(data T) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for ch := range p.clients {
		ch <- data
	}
}
