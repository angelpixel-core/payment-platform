package inprocess

import (
	"sync"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type Publisher struct {
	mu       sync.RWMutex
	handlers map[string][]ports.EventHandler
}

func NewPublisher() *Publisher {
	return &Publisher{handlers: make(map[string][]ports.EventHandler)}
}

var _ ports.EventPublisher = (*Publisher)(nil)

func (p *Publisher) Subscribe(eventName string, handler ports.EventHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[eventName] = append(p.handlers[eventName], handler)
}

func (p *Publisher) Publish(event domain.Event) error {
	p.mu.RLock()
	handlers := append([]ports.EventHandler(nil), p.handlers[event.EventName()]...)
	p.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return err
		}
	}
	return nil
}
