package refunds

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"payment-sandbox/internal/adapters/persistence/memory"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

func fingerprintString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type staticClock struct{}

func (staticClock) Now() time.Time { return time.Unix(0, 0).UTC() }

type eventSpy struct{ events []string }

func (s *eventSpy) Publish(event domain.Event) error {
	s.events = append(s.events, event.EventName())
	return nil
}

func (s *eventSpy) Subscribe(string, ports.EventHandler) {}

func TestRefundServicePublishesEvents(t *testing.T) {
	store := memory.NewStore(nil)
	store.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1", Currency: "USD", Status: domain.PaymentIntentSucceeded})
	store.SaveCharge(domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1", Amount: 100, CapturedAmount: 100, Status: domain.ChargeCaptured})

	spy := &eventSpy{}
	uow := memory.NewUnitOfWork(store, spy)
	svc := NewService(uow, staticClock{})

	if _, err := svc.CreateRefund(domain.RefundRequest{ChargeID: "ch_1"}, "refund-1", fingerprintString("refund-1")); err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	if len(spy.events) != 1 || spy.events[0] != "refund.created" {
		t.Fatalf("unexpected events: %#v", spy.events)
	}
}
