package outbox

import (
	"testing"

	"payment-sandbox/internal/adapters/eventing/inprocess"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type testEvent struct{ name string }

func (e testEvent) EventName() string { return e.name }

func TestPublisherQueuesAndDispatches(t *testing.T) {
	downstream := inprocess.NewPublisher()
	called := 0
	downstream.Subscribe("test.event", func(event domain.Event) error {
		called++
		return nil
	})

	publisher := NewPublisher(downstream)
	if err := publisher.Publish(testEvent{name: "test.event"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected downstream handler once, got %d", called)
	}
	records := publisher.Snapshot()
	if len(records) != 1 || records[0].State != RecordDispatched {
		t.Fatalf("expected dispatched record, got %#v", records)
	}
}

func TestPublisherKeepsPendingOnDownstreamError(t *testing.T) {
	publisher := NewPublisher(&failingPublisher{})
	if err := publisher.Publish(testEvent{name: "test.event"}); err == nil {
		t.Fatal("expected publish error")
	}
	records := publisher.Snapshot()
	if len(records) != 1 || records[0].State != RecordPending {
		t.Fatalf("expected pending record, got %#v", records)
	}
}

type failingPublisher struct{}

func (failingPublisher) Publish(domain.Event) error           { return domain.NewError(500, "boom", "boom") }
func (failingPublisher) Subscribe(string, ports.EventHandler) {}
