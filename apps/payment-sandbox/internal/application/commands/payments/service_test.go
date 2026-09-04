package payments

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

type fakeScenarioResolver struct{}

func (fakeScenarioResolver) Resolve(headerScenario, paymentMethodToken string) (domain.ScenarioName, error) {
	switch {
	case headerScenario != "":
		return domain.NormalizeScenarioName(headerScenario), nil
	case paymentMethodToken == "pm_card_processing":
		return domain.ScenarioProcessingThenSucceeded, nil
	default:
		return domain.ScenarioApprovedImmediate, nil
	}
}

func (fakeScenarioResolver) Outcome(name domain.ScenarioName) (domain.ScenarioOutcome, error) {
	switch domain.NormalizeScenarioName(string(name)) {
	case domain.ScenarioApprovedImmediate:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioApprovedImmediate, IntentStatus: domain.PaymentIntentRequiresCapture, AttemptStatus: domain.PaymentAttemptAuthorized, ChargeStatus: domain.ChargeAuthorized, CreatesCharge: true}, nil
	case domain.ScenarioDeclinedInsufficientFunds:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioDeclinedInsufficientFunds, IntentStatus: domain.PaymentIntentFailed, AttemptStatus: domain.PaymentAttemptDeclined, DeclineCode: "insufficient_funds"}, nil
	case domain.ScenarioRequiresAction3DS:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioRequiresAction3DS, IntentStatus: domain.PaymentIntentRequiresAction, AttemptStatus: domain.PaymentAttemptRequiresAction}, nil
	case domain.ScenarioProcessingThenSucceeded:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioProcessingThenSucceeded, IntentStatus: domain.PaymentIntentProcessing, AttemptStatus: domain.PaymentAttemptSubmitted, FinalizesLater: true}, nil
	default:
		return domain.ScenarioOutcome{}, domain.NewError(422, "invalid_scenario", "unknown scenario")
	}
}

type eventSpy struct {
	events []string
}

func (s *eventSpy) Publish(event domain.Event) error {
	s.events = append(s.events, event.EventName())
	return nil
}

func (s *eventSpy) Subscribe(string, ports.EventHandler) {}

func TestPaymentServicePublishesEvents(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, svc *PaymentService)
		wantEvents []string
	}{
		{
			name: "create emits created",
			setup: func(t *testing.T, svc *PaymentService) {
				if _, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", fingerprintString("create-1")); err != nil {
					t.Fatalf("create failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created"},
		},
		{
			name: "confirm emits confirmed",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", fingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				if _, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{}, "", "confirm-1", fingerprintString("confirm-1")); err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed"},
		},
		{
			name: "finalize emits finalized",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", fingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				confirmed, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", fingerprintString("confirm-1|processing"))
				if err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
				if confirmed.PaymentIntent.Status != domain.PaymentIntentProcessing {
					t.Fatalf("expected processing, got %s", confirmed.PaymentIntent.Status)
				}
				if _, err := svc.FinalizeProcessingPaymentIntent(created.ID); err != nil {
					t.Fatalf("finalize failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed", "payment_intent.finalized"},
		},
		{
			name: "capture emits captured",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", fingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				confirmed, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{}, "", "confirm-1", fingerprintString("confirm-1"))
				if err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
				if confirmed.PaymentIntent.Status != domain.PaymentIntentRequiresCapture {
					t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
				}
				if _, err := svc.CapturePaymentIntent(created.ID, domain.CapturePaymentIntentRequest{IdempotencyKey: "capture-1"}, fingerprintString("capture-1|amount=100")); err != nil {
					t.Fatalf("capture failed: %v", err)
				}
				if _, err := svc.CapturePaymentIntent(created.ID, domain.CapturePaymentIntentRequest{IdempotencyKey: "capture-1"}, fingerprintString("capture-1|amount=100")); err != nil {
					t.Fatalf("capture retry failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed", "payment_intent.captured"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &eventSpy{}
			uow := memory.NewUnitOfWork(memory.NewStore(nil), spy)
			svc := NewService(uow, staticClock{}, fakeScenarioResolver{})
			tt.setup(t, svc)
			if len(spy.events) != len(tt.wantEvents) {
				t.Fatalf("expected %d events, got %d: %v", len(tt.wantEvents), len(spy.events), spy.events)
			}
			for i, want := range tt.wantEvents {
				if spy.events[i] != want {
					t.Fatalf("event %d: expected %s, got %s", i, want, spy.events[i])
				}
			}
		})
	}
}

type staticClock struct{}

func (staticClock) Now() time.Time { return time.Unix(0, 0).UTC() }
