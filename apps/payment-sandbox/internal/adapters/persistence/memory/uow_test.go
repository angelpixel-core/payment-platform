package memory

import (
	"testing"

	"payment-sandbox/internal/adapters/messaging/inprocess"
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

func TestUnitOfWorkAtomicityOnError(t *testing.T) {
	store := NewStore()
	publisher := inprocess.NewPublisher()
	called := 0
	publisher.Subscribe("payment_intent.created", func(event domain.Event) error {
		called++
		return nil
	})

	uow := NewUnitOfWork(store, publisher)
	err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.SavePaymentAttempt(domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1"})
		return domain.NewError(500, "boom", "boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if called != 0 {
		t.Fatalf("expected no published events, got %d", called)
	}
	if _, err := store.GetPaymentIntent("pi_1"); err == nil {
		t.Fatal("expected no committed payment intent")
	}
	if _, err := store.GetPaymentAttempt("pa_1"); err == nil {
		t.Fatal("expected no committed payment attempt")
	}
}

func TestUnitOfWorkRollbackOnPublishFailure(t *testing.T) {
	store := NewStore()
	uow := NewUnitOfWork(store, failingPublisher{})

	err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.SaveCharge(domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1"})
		tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: domain.PaymentIntent{ID: "pi_1"}})
		return nil
	})
	if err == nil {
		t.Fatal("expected publish failure")
	}
	if _, err := store.GetPaymentIntent("pi_1"); err == nil {
		t.Fatal("expected rollback of payment intent")
	}
	if _, err := store.GetCharge("ch_1"); err == nil {
		t.Fatal("expected rollback of charge")
	}
}

type failingPublisher struct{}

func (failingPublisher) Publish(domain.Event) error           { return domain.NewError(500, "boom", "boom") }
func (failingPublisher) Subscribe(string, ports.EventHandler) {}
