package outbox

import (
	"context"
	"sync"
	"time"

	"payment-sandbox/internal/adapters/observability/metrics"
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
	recorder   metrics.MetricsRecorder
}

func NewPublisher(downstream ports.EventPublisher, recorder metrics.MetricsRecorder) *Publisher {
	return &Publisher{downstream: downstream, recorder: recorder}
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
	start := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = append(p.queue, event)
	p.history = append(p.history, Record{Event: event, State: RecordPending})
	p.recordOutbox("enqueue", "success", time.Since(start))
	p.recordPendingLocked()
}

func (p *Publisher) Flush() error {
	for {
		start := time.Now()
		p.mu.Lock()
		if len(p.queue) == 0 {
			p.recordPendingLocked()
			p.mu.Unlock()
			return nil
		}
		event := p.queue[0]
		p.queue = p.queue[1:]
		idx := p.nextPendingIndexLocked()
		p.mu.Unlock()

		if p.downstream != nil {
			if err := p.downstream.Publish(event); err != nil {
				p.recordOutbox("publish", "failure", time.Since(start))
				p.mu.Lock()
				p.recordPendingLocked()
				p.mu.Unlock()
				return err
			}
		}

		p.mu.Lock()
		if idx >= 0 && idx < len(p.history) {
			p.history[idx].State = RecordDispatched
		}
		p.recordOutbox("publish", "success", time.Since(start))
		p.recordPendingLocked()
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

func (p *Publisher) recordOutbox(operation, outcome string, duration time.Duration) {
	if p == nil || p.recorder == nil {
		return
	}
	p.recorder.RecordOutboxOperation(context.Background(), "memory", operation, outcome, duration)
}

func (p *Publisher) recordPendingLocked() {
	if p == nil || p.recorder == nil {
		return
	}
	var pending int64
	for _, record := range p.history {
		if record.State == RecordPending {
			pending++
		}
	}
	p.recorder.RecordOutboxPending(context.Background(), "memory", pending)
}
