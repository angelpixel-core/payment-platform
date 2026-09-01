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
func (f *fakePersistenceMetricsRecorder) RecordUnitOfWork(_ context.Context, backend, outcome string, duration time.Duration) {
	f.calls = append(f.calls, persistenceMetricCall{backend: backend, resource: "uow", operation: "do", outcome: outcome, duration: duration})
}
func (f *fakePersistenceMetricsRecorder) RecordOutboxOperation(context.Context, string, string, string, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordOutboxPending(context.Context, string, int64) {}

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
		{backend: "memory", resource: "uow", operation: "do", outcome: "success"},
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

func TestMemoryUnitOfWorkRecordsMetrics(t *testing.T) {
	recorder := &fakePersistenceMetricsRecorder{}
	store := NewStore(recorder)
	uow := NewUnitOfWork(store, noopPublisher{})

	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		return nil
	}); err != nil {
		t.Fatalf("uow success failed: %v", err)
	}
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_2"})
		return domain.NewError(500, "boom", "boom")
	}); err == nil {
		t.Fatal("expected rollback error")
	}

failing := NewUnitOfWork(store, failingPublisher{})
	if err := failing.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_3"})
		tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: domain.PaymentIntent{ID: "pi_3"}})
		return nil
	}); err == nil {
		t.Fatal("expected commit error")
	}

	uowCalls := filterPersistenceCalls(recorder.calls, "uow")
	wantOutcomes := []string{"success", "rollback", "commit_error"}
	if len(uowCalls) != len(wantOutcomes) {
		t.Fatalf("expected %d uow calls, got %d: %#v", len(wantOutcomes), len(uowCalls), uowCalls)
	}
	for i, want := range wantOutcomes {
		got := uowCalls[i]
		if got.backend != "memory" || got.resource != "uow" || got.operation != "do" || got.outcome != want {
			t.Fatalf("call %d: expected outcome %s, got %#v", i, want, got)
		}
		if got.duration <= 0 {
			t.Fatalf("call %d: expected positive duration, got %s", i, got.duration)
		}
	}
}

func filterPersistenceCalls(calls []persistenceMetricCall, resource string) []persistenceMetricCall {
	filtered := make([]persistenceMetricCall, 0, len(calls))
	for _, call := range calls {
		if call.resource == resource {
			filtered = append(filtered, call)
		}
	}
	return filtered
}

type noopPublisher struct{}

func (noopPublisher) Publish(domain.Event) error           { return nil }
func (noopPublisher) Subscribe(string, ports.EventHandler) {}
