package sandbox

import "testing"

func TestFinalizeProcessingPaymentIntent(t *testing.T) {
	svc := NewService()

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-1", FingerprintString("create-1"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", FingerprintString("confirm-1|processing"))
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
	if finalized.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture after finalization, got %s", finalized.Status)
	}
}

func TestScenarioResolution(t *testing.T) {
	config := DefaultScenarioConfig()

	if scenario, err := config.Resolve("declined_insufficient_funds", "pm_card_visa"); err != nil {
		t.Fatalf("resolve failed: %v", err)
	} else if scenario != ScenarioDeclinedInsufficientFunds {
		t.Fatalf("expected header priority, got %s", scenario)
	}

	if scenario, err := config.Resolve("", "pm_card_insufficient_funds"); err != nil {
		t.Fatalf("resolve failed: %v", err)
	} else if scenario != ScenarioDeclinedInsufficientFunds {
		t.Fatalf("expected token fallback, got %s", scenario)
	}

	if _, err := config.Resolve("unknown_scenario", "pm_card_visa"); err == nil {
		t.Fatal("expected invalid scenario error")
	}
}
