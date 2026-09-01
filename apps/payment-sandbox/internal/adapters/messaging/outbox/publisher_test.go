package outbox

import (
	"context"
	"testing"
	"time"

	"payment-sandbox/internal/adapters/messaging/inprocess"
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

	recorder := &fakeOutboxMetricsRecorder{}
	publisher := NewPublisher(downstream, recorder)
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
	if len(recorder.ops) != 2 {
		t.Fatalf("expected 2 outbox ops, got %#v", recorder.ops)
	}
	if recorder.pendings[len(recorder.pendings)-1] != 0 {
		t.Fatalf("expected no pending records at end, got %#v", recorder.pendings)
	}
}

func TestPublisherKeepsPendingOnDownstreamError(t *testing.T) {
	recorder := &fakeOutboxMetricsRecorder{}
	publisher := NewPublisher(&failingPublisher{}, recorder)
	if err := publisher.Publish(testEvent{name: "test.event"}); err == nil {
		 t.Fatal("expected publish error")
	}
	records := publisher.Snapshot()
	if len(records) != 1 || records[0].State != RecordPending {
		t.Fatalf("expected pending record, got %#v", records)
	}
	if len(recorder.ops) != 2 {
		t.Fatalf("expected 2 outbox ops, got %#v", recorder.ops)
	}
	if recorder.pendings[len(recorder.pendings)-1] != 1 {
		t.Fatalf("expected pending record to remain, got %#v", recorder.pendings)
	}
}

type outboxOpCall struct {
	backend   string
	operation string
	outcome   string
	duration  time.Duration
}

type fakeOutboxMetricsRecorder struct {
	ops      []outboxOpCall
	pendings []int64
}

func (f *fakeOutboxMetricsRecorder) RecordHTTPRequest(context.Context, string, string, int, time.Duration) {}
func (f *fakeOutboxMetricsRecorder) RecordPaymentFlow(context.Context, string, string, time.Duration) {}
func (f *fakeOutboxMetricsRecorder) RecordPaymentCommand(context.Context, string, string, time.Duration) {}
func (f *fakeOutboxMetricsRecorder) RecordPersistenceOperation(context.Context, string, string, string, string, time.Duration) {}
func (f *fakeOutboxMetricsRecorder) RecordUnitOfWork(context.Context, string, string, time.Duration) {}
func (f *fakeOutboxMetricsRecorder) RecordOutboxOperation(_ context.Context, backend, operation, outcome string, duration time.Duration) {
	f.ops = append(f.ops, outboxOpCall{backend: backend, operation: operation, outcome: outcome, duration: duration})
}
func (f *fakeOutboxMetricsRecorder) RecordOutboxPending(_ context.Context, _ string, pending int64) {
	f.pendings = append(f.pendings, pending)
}

type failingPublisher struct{}

func (failingPublisher) Publish(domain.Event) error           { return domain.NewError(500, "boom", "boom") }
func (failingPublisher) Subscribe(string, ports.EventHandler) {}
