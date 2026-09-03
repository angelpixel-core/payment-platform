package webhook

import (
	"context"
	"errors"
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
}

type transportFunc func(ctx context.Context, request Request) error

func (f transportFunc) Send(ctx context.Context, request Request) error { return f(ctx, request) }
