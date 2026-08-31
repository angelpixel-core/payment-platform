package memory

import (
	"testing"

	"payment-sandbox/internal/adapters/eventing/inprocess"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

func TestUnitOfWorkDo(t *testing.T) {
	store := NewStore()
	publisher := inprocess.NewPublisher()
	called := 0
	publisher.Subscribe("payment_intent.created", func(event domain.Event) error {
		called++
		return nil
	})

	uow := NewUnitOfWork(store, publisher)
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		return tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: domain.PaymentIntent{ID: "pi_1"}})
	}); err != nil {
		t.Fatalf("do failed: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected downstream publish once, got %d", called)
	}
	got, err := store.GetPaymentIntent("pi_1")
	if err != nil {
		t.Fatalf("get payment intent failed: %v", err)
	}
	if got.ID != "pi_1" {
		t.Fatalf("expected stored payment intent")
	}
}
