package outbox

import (
	"sync"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type RecordState string

const (
	RecordPending    RecordState = "pending"
	RecordDispatched RecordState = "dispatched"
)

type Record struct {
	Event domain.Event
	State RecordState
}

type Publisher struct {
	mu         sync.Mutex
	queue      []domain.Event
	history    []Record
	downstream ports.EventPublisher
}

func NewPublisher(downstream ports.EventPublisher) *Publisher {
	return &Publisher{downstream: downstream}
}

var _ ports.EventPublisher = (*Publisher)(nil)

func (p *Publisher) Publish(event domain.Event) error {
	p.Enqueue(event)
	return p.Flush()
}

func (p *Publisher) Subscribe(eventName string, handler ports.EventHandler) {
	if p.downstream != nil {
		p.downstream.Subscribe(eventName, handler)
	}
}

func (p *Publisher) Enqueue(event domain.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = append(p.queue, event)
	p.history = append(p.history, Record{Event: event, State: RecordPending})
}

func (p *Publisher) Flush() error {
	for {
		p.mu.Lock()
		if len(p.queue) == 0 {
			p.mu.Unlock()
			return nil
		}
		event := p.queue[0]
		p.queue = p.queue[1:]
		idx := p.nextPendingIndexLocked()
		p.mu.Unlock()

		if p.downstream != nil {
			if err := p.downstream.Publish(event); err != nil {
				return err
			}
		}

		p.mu.Lock()
		if idx >= 0 && idx < len(p.history) {
			p.history[idx].State = RecordDispatched
		}
		p.mu.Unlock()
	}
}

func (p *Publisher) Snapshot() []Record {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]Record, len(p.history))
	copy(result, p.history)
	return result
}

func (p *Publisher) nextPendingIndexLocked() int {
	for i := len(p.history) - 1; i >= 0; i-- {
		if p.history[i].State == RecordPending {
			return i
		}
	}
	return -1
}
