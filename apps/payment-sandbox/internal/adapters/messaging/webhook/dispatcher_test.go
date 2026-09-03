package webhook

import (
	"context"
	"testing"
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

type transportFunc func(ctx context.Context, request Request) error

func (f transportFunc) Send(ctx context.Context, request Request) error { return f(ctx, request) }
