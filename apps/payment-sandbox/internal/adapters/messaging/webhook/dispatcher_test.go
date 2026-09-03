package webhook

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestNewDeliveryClonesPayloadAndDispatchReusesIt(t *testing.T) {
	original := []byte(`{"event_type":"payment.succeeded"}`)
	delivery := NewDelivery("payment.succeeded", "evt_1", "del_1", "https://example.test/webhooks", 1, original)
	original[0] = 'x'

	if string(delivery.Payload) != `{"event_type":"payment.succeeded"}` {
		t.Fatalf("expected delivery payload to be immutable copy, got %q", string(delivery.Payload))
	}

	var got Request
	transport := transportFunc(func(ctx context.Context, request Request) error {
		got = request
		return nil
	})

	dispatcher := NewDispatcher(transport)
	if err := dispatcher.Dispatch(context.Background(), delivery); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if got.Endpoint != delivery.Endpoint {
		t.Fatalf("expected endpoint %q, got %q", delivery.Endpoint, got.Endpoint)
	}
	if string(got.Body) != string(delivery.Payload) {
		t.Fatalf("expected body %q, got %q", string(delivery.Payload), string(got.Body))
	}
}

func TestDispatcherRetriesUntilSuccess(t *testing.T) {
	var attempts int
	var slept []time.Duration
	transport := transportFunc(func(ctx context.Context, request Request) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	dispatcher := NewDispatcher(transport)
	dispatcher.sleep = func(d time.Duration) { slept = append(slept, d) }

	delivery := NewDelivery("payment.failed", "evt_2", "del_2", "https://example.test/webhooks", 1, []byte(`{"event_type":"payment.failed"}`))
	if err := dispatcher.Dispatch(context.Background(), delivery); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(slept) != 2 {
		t.Fatalf("expected 2 sleep intervals, got %d", len(slept))
	}
	if slept[0] != 10*time.Millisecond || slept[1] != 20*time.Millisecond {
		t.Fatalf("expected linear backoff [10ms 20ms], got %#v", slept)
	}
	history := dispatcher.AttemptHistory()
	if len(history) != 3 {
		t.Fatalf("expected 3 attempt records, got %d", len(history))
	}
	if history[0] != (AttemptRecord{Attempt: 1, Outcome: "failure"}) || history[1] != (AttemptRecord{Attempt: 2, Outcome: "failure"}) || history[2] != (AttemptRecord{Attempt: 3, Outcome: "success"}) {
		t.Fatalf("unexpected attempt history: %#v", history)
	}
	if dispatcher.FinalState() != "delivered" {
		t.Fatalf("expected final state delivered, got %q", dispatcher.FinalState())
	}
	trace := dispatcher.Trace(delivery)
	if trace.DeliveryID != delivery.DeliveryID || trace.EventID != delivery.EventID || trace.EventType != delivery.EventType {
		t.Fatalf("unexpected trace metadata: %#v", trace)
	}
	if trace.Endpoint != delivery.Endpoint {
		t.Fatalf("expected trace endpoint %q, got %q", delivery.Endpoint, trace.Endpoint)
	}
	if trace.FinalState != "delivered" {
		t.Fatalf("expected trace final state delivered, got %q", trace.FinalState)
	}
	if len(trace.Attempts) != 3 {
		t.Fatalf("expected trace attempts length 3, got %d", len(trace.Attempts))
	}
}

func TestDispatcherKeepsPayloadBytesStableAcrossRetries(t *testing.T) {
	var attempts int
	var bodies [][]byte
	transport := transportFunc(func(ctx context.Context, request Request) error {
		attempts++
		bodies = append(bodies, append([]byte(nil), request.Body...))
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	dispatcher := NewDispatcher(transport)
	dispatcher.sleep = func(time.Duration) {}

	payload := []byte(`{"event_type":"payment.succeeded","data":{"charge_id":"ch_1"}}`)
	delivery := NewDelivery("payment.succeeded", "evt_immutable", "del_immutable", "https://example.test/webhooks", 1, payload)
	if err := dispatcher.Dispatch(context.Background(), delivery); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(bodies) != 3 {
		t.Fatalf("expected 3 captured bodies, got %d", len(bodies))
	}
	for i, body := range bodies {
		if !reflect.DeepEqual(body, delivery.Payload) {
			t.Fatalf("attempt %d body changed: got %q want %q", i+1, string(body), string(delivery.Payload))
		}
	}
	if !reflect.DeepEqual(payload, []byte(`{"event_type":"payment.succeeded","data":{"charge_id":"ch_1"}}`)) {
		t.Fatalf("expected original payload literal to remain unchanged")
	}
}

func TestDispatcherKeepsDeliveryIDStableAcrossRetries(t *testing.T) {
	var attempts int
	transport := transportFunc(func(ctx context.Context, request Request) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	dispatcher := NewDispatcher(transport)
	dispatcher.sleep = func(time.Duration) {}

	delivery := NewDelivery("payment.failed", "evt_delivery_id", "del_delivery_id", "https://example.test/webhooks", 1, []byte(`{"event_type":"payment.failed"}`))
	if err := dispatcher.Dispatch(context.Background(), delivery); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	trace := dispatcher.Trace(delivery)
	if trace.DeliveryID != delivery.DeliveryID {
		t.Fatalf("expected trace delivery id %q, got %q", delivery.DeliveryID, trace.DeliveryID)
	}
	if trace.DeliveryID != "del_delivery_id" {
		t.Fatalf("expected stable delivery id del_delivery_id, got %q", trace.DeliveryID)
	}
	if len(trace.Attempts) != 3 {
		t.Fatalf("expected 3 trace attempts, got %d", len(trace.Attempts))
	}
}

func TestDispatcherReturnsFinalErrorAfterRetries(t *testing.T) {
	var attempts int
	transport := transportFunc(func(ctx context.Context, request Request) error {
		attempts++
		return errors.New("permanent failure")
	})

	dispatcher := NewDispatcher(transport)
	dispatcher.maxAttempts = 2
	dispatcher.sleep = func(time.Duration) {}

	delivery := NewDelivery("payment.processing", "evt_3", "del_3", "https://example.test/webhooks", 1, []byte(`{"event_type":"payment.processing"}`))
	err := dispatcher.Dispatch(context.Background(), delivery)
	if err == nil {
		t.Fatalf("expected final error")
	}
	var dispatchErr DispatchError
	if !errors.As(err, &dispatchErr) {
		t.Fatalf("expected DispatchError, got %T: %v", err, err)
	}
	if dispatchErr.DeliveryID != delivery.DeliveryID {
		t.Fatalf("expected delivery id %q, got %q", delivery.DeliveryID, dispatchErr.DeliveryID)
	}
	if dispatchErr.Attempts != 2 {
		t.Fatalf("expected 2 attempts in error, got %d", dispatchErr.Attempts)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	history := dispatcher.AttemptHistory()
	if len(history) != 2 {
		t.Fatalf("expected 2 attempt records, got %d", len(history))
	}
	if history[0] != (AttemptRecord{Attempt: 1, Outcome: "failure"}) || history[1] != (AttemptRecord{Attempt: 2, Outcome: "failure"}) {
		t.Fatalf("unexpected attempt history: %#v", history)
	}
	if dispatcher.FinalState() != "failed" {
		t.Fatalf("expected final state failed, got %q", dispatcher.FinalState())
	}
	trace := dispatcher.Trace(delivery)
	if trace.FinalState != "failed" {
		t.Fatalf("expected trace final state failed, got %q", trace.FinalState)
	}
	if len(trace.Attempts) != 2 {
		t.Fatalf("expected trace attempts length 2, got %d", len(trace.Attempts))
	}
}

type transportFunc func(ctx context.Context, request Request) error

func (f transportFunc) Send(ctx context.Context, request Request) error { return f(ctx, request) }
