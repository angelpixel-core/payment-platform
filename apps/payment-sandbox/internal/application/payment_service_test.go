package application

import (
	"testing"

	"payment-sandbox/internal/adapters/memory"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

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
				if _, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", FingerprintString("create-1")); err != nil {
					t.Fatalf("create failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created"},
		},
		{
			name: "confirm emits confirmed",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", FingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				if _, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{}, "", "confirm-1", FingerprintString("confirm-1")); err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed"},
		},
		{
			name: "finalize emits finalized",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", FingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				confirmed, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", FingerprintString("confirm-1|processing"))
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
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", FingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				confirmed, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{}, "", "confirm-1", FingerprintString("confirm-1"))
				if err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
				if confirmed.PaymentIntent.Status != domain.PaymentIntentRequiresCapture {
					t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
				}
				if _, err := svc.CapturePaymentIntent(created.ID, domain.CapturePaymentIntentRequest{}); err != nil {
					t.Fatalf("capture failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed", "payment_intent.captured"},
		},
		{
			name: "refund emits refund_created",
			setup: func(t *testing.T, svc *PaymentService) {
				created, err := svc.CreatePaymentIntent(domain.CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", FingerprintString("create-1"))
				if err != nil {
					t.Fatalf("create failed: %v", err)
				}
				confirmed, err := svc.ConfirmPaymentIntent(created.ID, domain.ConfirmPaymentIntentRequest{}, "", "confirm-1", FingerprintString("confirm-1"))
				if err != nil {
					t.Fatalf("confirm failed: %v", err)
				}
				captured, err := svc.CapturePaymentIntent(created.ID, domain.CapturePaymentIntentRequest{})
				if err != nil {
					t.Fatalf("capture failed: %v", err)
				}
				if captured.PaymentIntent.Status != domain.PaymentIntentSucceeded || confirmed.Charge == nil {
					t.Fatalf("expected successful capture flow")
				}
				if _, err := svc.CreateRefund(domain.RefundRequest{ChargeID: captured.Charge.ID}, "refund-1", FingerprintString("refund-1")); err != nil {
					t.Fatalf("refund failed: %v", err)
				}
			},
			wantEvents: []string{"payment_intent.created", "payment_intent.confirmed", "payment_intent.captured", "refund.created"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &eventSpy{}
			uow := memory.NewUnitOfWork(memory.NewStore(), spy)
			svc := NewPaymentService(uow, fakeScenarioResolver{})
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
