package events

import (
	"testing"

	"payment-sandbox/internal/adapters/eventing/inprocess"
)

type testEvent struct{ name string }

func (e testEvent) EventName() string { return e.name }

func TestRecorderAndRegistration(t *testing.T) {
	publisher := inprocess.NewPublisher()
	recorder := NewRecorder()
	RegisterInternalHandlers(publisher, recorder)

	if err := publisher.Publish(testEvent{name: "payment_intent.created"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if err := publisher.Publish(testEvent{name: "payment_intent.created"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if err := publisher.Publish(testEvent{name: "refund.created"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	names, counts := recorder.Snapshot()
	if len(names) != 3 {
		t.Fatalf("expected 3 events, got %d", len(names))
	}
	if counts["payment_intent.created"] != 2 || counts["refund.created"] != 1 {
		t.Fatalf("unexpected counts: %#v", counts)
	}
}

func TestRegisterInternalHandlersNoop(t *testing.T) {
	RegisterInternalHandlers(nil, nil)
}
