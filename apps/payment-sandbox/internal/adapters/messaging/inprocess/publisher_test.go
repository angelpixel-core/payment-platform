package inprocess

import (
	"testing"

	"payment-sandbox/internal/domain"
)

type testEvent struct{}

func (testEvent) EventName() string { return "test.event" }

func TestPublisherDispatchesHandlers(t *testing.T) {
	p := NewPublisher()
	called := 0
	p.Subscribe("test.event", func(event domain.Event) error {
		called++
		return nil
	})

	if err := p.Publish(testEvent{}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected handler to be called once, got %d", called)
	}
}
