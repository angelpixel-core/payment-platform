package observability

import (
	"reflect"
	"testing"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
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

func TestRegisterInternalHandlersSubscribesExpectedEvents(t *testing.T) {
	pub := &capturePublisher{}
	recorder := NewRecorder()

	RegisterInternalHandlers(pub, recorder)

	want := []string{
		"payment_intent.created",
		"payment_intent.confirmed",
		"payment_intent.finalized",
		"payment_intent.captured",
		"refund.created",
	}
	if !reflect.DeepEqual(pub.eventNames, want) {
		t.Fatalf("unexpected subscriptions: got %#v want %#v", pub.eventNames, want)
	}
	for _, handler := range pub.handlers {
		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
	}
}

type capturePublisher struct {
	eventNames []string
	handlers   []ports.EventHandler
}

func (p *capturePublisher) Publish(domain.Event) error { return nil }

func (p *capturePublisher) Subscribe(eventName string, handler ports.EventHandler) {
	p.eventNames = append(p.eventNames, eventName)
	p.handlers = append(p.handlers, handler)
}
