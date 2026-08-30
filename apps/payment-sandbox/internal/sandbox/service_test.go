package sandbox

import (
	"testing"

	"payment-sandbox/internal/application"
)

func TestFinalizeProcessingPaymentIntent(t *testing.T) {
	tests := []struct {
		name          string
		captureMethod string
		wantStatus    PaymentIntentStatus
	}{
		{name: "manual capture", captureMethod: "manual", wantStatus: PaymentIntentRequiresCapture},
		{name: "automatic capture", captureMethod: "automatic", wantStatus: PaymentIntentSucceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService()

			created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: tt.captureMethod}, "create-1", application.FingerprintString("create-1"))
			if err != nil {
				t.Fatalf("create failed: %v", err)
			}

			confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", application.FingerprintString("confirm-1|processing"))
			if err != nil {
				t.Fatalf("confirm failed: %v", err)
			}
			if confirmed.PaymentIntent.Status != PaymentIntentProcessing {
				t.Fatalf("expected processing, got %s", confirmed.PaymentIntent.Status)
			}
			if confirmed.Charge != nil {
				t.Fatal("expected no charge before finalization")
			}

			finalized, err := svc.FinalizeProcessingPaymentIntent(created.ID)
			if err != nil {
				t.Fatalf("finalize failed: %v", err)
			}
			if finalized.Status != tt.wantStatus {
				t.Fatalf("expected %s after finalization, got %s", tt.wantStatus, finalized.Status)
			}
		})
	}
}

func TestPaymentLifecycle(t *testing.T) {
	svc := NewService()

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", application.FingerprintString("create-1"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_visa"}, "", "confirm-1", application.FingerprintString("confirm-1|confirm"))
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if confirmed.PaymentIntent.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
	}
	if confirmed.Charge == nil {
		t.Fatal("expected charge after confirm")
	}

	captured, err := svc.CapturePaymentIntent(created.ID, CapturePaymentIntentRequest{})
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if captured.PaymentIntent.Status != PaymentIntentSucceeded {
		t.Fatalf("expected succeeded, got %s", captured.PaymentIntent.Status)
	}

	refunded, err := svc.CreateRefund(RefundRequest{ChargeID: captured.Charge.ID}, "refund-1", application.FingerprintString("refund-1|refund"))
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if refunded.Refund.Amount != 100 {
		t.Fatalf("expected full refund, got %d", refunded.Refund.Amount)
	}
	if refunded.Charge.Status != ChargeRefunded {
		t.Fatalf("expected refunded charge, got %s", refunded.Charge.Status)
	}
}

func TestScenarioResolution(t *testing.T) {
	engine := NewScenarioEngine()

	tests := []struct {
		name           string
		headerScenario string
		token          string
		wantScenario   ScenarioName
		wantError      bool
	}{
		{name: "header priority", headerScenario: "declined_insufficient_funds", token: "pm_card_visa", wantScenario: ScenarioDeclinedInsufficientFunds},
		{name: "token fallback", headerScenario: "", token: "pm_card_insufficient_funds", wantScenario: ScenarioDeclinedInsufficientFunds},
		{name: "unknown header", headerScenario: "unknown_scenario", token: "pm_card_visa", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := engine.Resolve(tt.headerScenario, tt.token)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected invalid scenario error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			if scenario != tt.wantScenario {
				t.Fatalf("expected %s, got %s", tt.wantScenario, scenario)
			}
		})
	}
}
