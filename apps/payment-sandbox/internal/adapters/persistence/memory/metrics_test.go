package memory

import (
	"context"
	"testing"
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type persistenceMetricCall struct {
	backend   string
	resource  string
	operation string
	outcome   string
	duration  time.Duration
}

type fakePersistenceMetricsRecorder struct {
	calls []persistenceMetricCall
}

func (f *fakePersistenceMetricsRecorder) RecordHTTPRequest(context.Context, string, string, int, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPaymentFlow(context.Context, string, string, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPaymentCommand(context.Context, string, string, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPersistenceOperation(_ context.Context, backend, resource, operation, outcome string, duration time.Duration) {
	f.calls = append(f.calls, persistenceMetricCall{backend: backend, resource: resource, operation: operation, outcome: outcome, duration: duration})
}

func TestMemoryStoreRecordsPersistenceMetrics(t *testing.T) {
	recorder := &fakePersistenceMetricsRecorder{}
	store := NewStore(recorder)

	store.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
	if _, err := store.GetPaymentIntent("pi_1"); err != nil {
		t.Fatalf("get payment intent failed: %v", err)
	}
	if _, err := store.GetCharge("missing"); err == nil {
		t.Fatal("expected missing charge error")
	}

	uow := NewUnitOfWork(store, noopPublisher{})
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentAttempt(domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1"})
		if _, err := tx.GetPaymentAttempt("pa_1"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("uow do failed: %v", err)
	}

	want := []persistenceMetricCall{
		{backend: "memory", resource: "payment_intent", operation: "save", outcome: "success"},
		{backend: "memory", resource: "payment_intent", operation: "get", outcome: "success"},
		{backend: "memory", resource: "charge", operation: "get", outcome: "error"},
		{backend: "memory", resource: "payment_attempt", operation: "save", outcome: "success"},
		{backend: "memory", resource: "payment_attempt", operation: "get", outcome: "success"},
	}

	if len(recorder.calls) != len(want) {
		t.Fatalf("expected %d calls, got %d: %#v", len(want), len(recorder.calls), recorder.calls)
	}
	for i := range want {
		got := recorder.calls[i]
		if got.backend != want[i].backend || got.resource != want[i].resource || got.operation != want[i].operation || got.outcome != want[i].outcome {
			t.Fatalf("call %d: expected %#v, got %#v", i, want[i], got)
		}
		if got.duration <= 0 {
			t.Fatalf("call %d: expected positive duration, got %s", i, got.duration)
		}
	}
}

type noopPublisher struct{}

func (noopPublisher) Publish(domain.Event) error           { return nil }
func (noopPublisher) Subscribe(string, ports.EventHandler) {}
