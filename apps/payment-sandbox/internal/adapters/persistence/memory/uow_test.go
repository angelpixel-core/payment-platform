package memory

import (
	"testing"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

func TestUnitOfWorkDo(t *testing.T) {
	store := NewStore(nil)
	publisher := inprocess.NewPublisher()
	called := 0
	publisher.Subscribe("payment_intent.created", func(event domain.Event) error {
		called++
		return nil
	})

	uow := NewUnitOfWork(store, publisher)
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.AppendLedgerEntry(domain.LedgerEntry{ID: "le_1", EventName: "payment_intent.created", EntityType: "payment_intent", EntityID: "pi_1", PaymentIntentID: "pi_1"})
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
	entries := store.ListLedgerEntries()
	if len(entries) != 1 || entries[0].EventName != "payment_intent.created" {
		t.Fatalf("expected one ledger entry, got %#v", entries)
	}
}

func TestUnitOfWorkAtomicityOnError(t *testing.T) {
	store := NewStore(nil)
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
		tx.AppendLedgerEntry(domain.LedgerEntry{ID: "le_1", EventName: "payment_intent.created", EntityType: "payment_intent", EntityID: "pi_1", PaymentIntentID: "pi_1"})
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
	if entries := store.ListLedgerEntries(); len(entries) != 0 {
		t.Fatalf("expected no committed ledger entries, got %#v", entries)
	}
}

func TestUnitOfWorkRollbackOnPublishFailure(t *testing.T) {
	store := NewStore(nil)
	uow := NewUnitOfWork(store, failingPublisher{})

	err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.SaveCharge(domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1"})
		tx.AppendLedgerEntry(domain.LedgerEntry{ID: "le_1", EventName: "payment_intent.created", EntityType: "payment_intent", EntityID: "pi_1", PaymentIntentID: "pi_1"})
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
	if entries := store.ListLedgerEntries(); len(entries) != 0 {
		t.Fatalf("expected rollback of ledger entries, got %#v", entries)
	}
}

type failingPublisher struct{}

func (failingPublisher) Publish(domain.Event) error           { return domain.NewError(500, "boom", "boom") }
func (failingPublisher) Subscribe(string, ports.EventHandler) {}
