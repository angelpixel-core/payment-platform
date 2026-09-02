package sandbox

import (
	"testing"

	"payment-sandbox/internal/adapters/persistence/memory"
)

func TestMemoryStoreRoundTrip(t *testing.T) {
	store := memory.NewStore(nil)

	intent := PaymentIntent{ID: "pi_1", Amount: 100}
	store.SavePaymentIntent(intent)

	gotIntent, err := store.GetPaymentIntent("pi_1")
	if err != nil {
		t.Fatalf("get intent failed: %v", err)
	}
	if gotIntent.ID != intent.ID || gotIntent.Amount != intent.Amount {
		t.Fatalf("unexpected intent: %+v", gotIntent)
	}

	value, err := store.WithIdempotency("key", "fingerprint", func() (any, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("idempotency failed: %v", err)
	}
	if value.(string) != "ok" {
		t.Fatalf("unexpected idempotency value: %v", value)
	}

	again, err := store.WithIdempotency("key", "fingerprint", func() (any, error) {
		return "wrong", nil
	})
	if err != nil {
		t.Fatalf("idempotency replay failed: %v", err)
	}
	if again.(string) != "ok" {
		t.Fatalf("expected cached value, got %v", again)
	}
}
