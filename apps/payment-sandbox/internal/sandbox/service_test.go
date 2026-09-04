package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func fingerprintString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func TestDefaultSandboxUsesMemory(t *testing.T) {
	svc := NewService()
	if svc.EventRecorder() == nil {
		t.Fatal("expected event recorder on default sandbox service")
	}

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "memory-default-create", fingerprintString("memory-default-create"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := svc.GetPaymentIntent(created.ID)
	if err != nil {
		t.Fatalf("get payment intent failed: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected intent %s, got %s", created.ID, got.ID)
	}
}

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

			created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: tt.captureMethod}, "create-1", fingerprintString("create-1"))
			if err != nil {
				t.Fatalf("create failed: %v", err)
			}

			confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", fingerprintString("confirm-1|processing"))
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

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", fingerprintString("create-1"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_visa"}, "", "confirm-1", fingerprintString("confirm-1|confirm"))
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if confirmed.PaymentIntent.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
	}
	if confirmed.Charge == nil {
		t.Fatal("expected charge after confirm")
	}

	captured, err := svc.CapturePaymentIntent(created.ID, CapturePaymentIntentRequest{}, "")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if captured.PaymentIntent.Status != PaymentIntentSucceeded {
		t.Fatalf("expected succeeded, got %s", captured.PaymentIntent.Status)
	}

	refunded, err := svc.CreateRefund(RefundRequest{ChargeID: captured.Charge.ID}, "refund-1", fingerprintString("refund-1|refund"))
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

func TestInternalEventHandlers(t *testing.T) {
	svc := NewService()

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-events", fingerprintString("create-events"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{}, "", "confirm-events", fingerprintString("confirm-events"))
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	_, err = svc.CapturePaymentIntent(created.ID, CapturePaymentIntentRequest{}, "")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	_, err = svc.CreateRefund(RefundRequest{ChargeID: confirmed.Charge.ID}, "refund-events", fingerprintString("refund-events"))
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}

	_, counts := svc.EventRecorder().Snapshot()
	if counts["payment_intent.created"] != 1 || counts["payment_intent.confirmed"] != 1 || counts["payment_intent.captured"] != 1 || counts["refund.created"] != 1 {
		t.Fatalf("unexpected event counts: %#v", counts)
	}
}
