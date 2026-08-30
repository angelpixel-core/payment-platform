package domain

import (
	"testing"
	"time"
)

func TestPaymentIntentConfirm(t *testing.T) {
	now := time.Unix(100, 0)

	t.Run("approved immediate automatic capture", func(t *testing.T) {
		intent := PaymentIntent{ID: "pi_1", Amount: 100, CaptureMethod: "automatic", Status: PaymentIntentRequiresPaymentMethod}
		attempt := &PaymentAttempt{ID: "pa_1"}
		charge := &Charge{ID: "ch_1"}

		outcome := ScenarioOutcome{Scenario: ScenarioApprovedImmediate, IntentStatus: PaymentIntentRequiresCapture, AttemptStatus: PaymentAttemptAuthorized, ChargeStatus: ChargeAuthorized, CreatesCharge: true}
		result, err := intent.Confirm(ConfirmPaymentIntentCommand{Outcome: outcome, Attempt: attempt, Charge: charge, Now: now})
		if err != nil {
			t.Fatalf("confirm failed: %v", err)
		}

		if result.PaymentIntent.Status != PaymentIntentSucceeded {
			t.Fatalf("expected succeeded, got %s", result.PaymentIntent.Status)
		}
		if result.PaymentIntent.ChargeID != "ch_1" {
			t.Fatalf("expected charge id to be linked")
		}
		if result.Charge == nil {
			t.Fatalf("expected charge result")
		}
		if result.Charge.Status != ChargeCaptured || result.Charge.CapturedAmount != 100 {
			t.Fatalf("expected captured charge, got %+v", result.Charge)
		}
	})

	t.Run("declined sets failure and decline code", func(t *testing.T) {
		intent := PaymentIntent{ID: "pi_1", Status: PaymentIntentRequiresPaymentMethod}
		attempt := &PaymentAttempt{ID: "pa_1"}

		outcome := ScenarioOutcome{Scenario: ScenarioDeclinedInsufficientFunds, IntentStatus: PaymentIntentFailed, AttemptStatus: PaymentAttemptDeclined, DeclineCode: "insufficient_funds"}
		result, err := intent.Confirm(ConfirmPaymentIntentCommand{Outcome: outcome, Attempt: attempt, Now: now})
		if err != nil {
			t.Fatalf("confirm failed: %v", err)
		}

		if result.PaymentIntent.Status != PaymentIntentFailed {
			t.Fatalf("expected failed, got %s", result.PaymentIntent.Status)
		}
		if result.PaymentAttempt.DeclineCode != "insufficient_funds" {
			t.Fatalf("expected decline code to be set")
		}
	})
}

func TestPaymentIntentFinalizeProcessing(t *testing.T) {
	now := time.Unix(200, 0)

	intent := PaymentIntent{ID: "pi_1", Amount: 100, CaptureMethod: "manual", Status: PaymentIntentProcessing, Scenario: string(ScenarioProcessingThenSucceeded)}
	attempt := &PaymentAttempt{ID: "pa_1", Status: PaymentAttemptSubmitted}
	charge := &Charge{ID: "ch_1"}

	result, err := intent.FinalizeProcessing(FinalizeProcessingCommand{Attempt: attempt, Charge: charge, Now: now})
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}

	if result.PaymentIntent.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture, got %s", result.PaymentIntent.Status)
	}
	if result.PaymentAttempt.Status != PaymentAttemptAuthorized {
		t.Fatalf("expected attempt authorized, got %s", result.PaymentAttempt.Status)
	}
	if result.Charge.Status != ChargeAuthorized {
		t.Fatalf("expected authorized charge, got %s", result.Charge.Status)
	}
}

func TestPaymentIntentCapture(t *testing.T) {
	now := time.Unix(300, 0)

	intent := PaymentIntent{ID: "pi_1", Status: PaymentIntentRequiresCapture, ChargeID: "ch_1"}
	charge := &Charge{ID: "ch_1", Amount: 100, Status: ChargeAuthorized}

	result, err := intent.Capture(CapturePaymentIntentCommand{Charge: charge, Now: now})
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	if result.PaymentIntent.Status != PaymentIntentSucceeded {
		t.Fatalf("expected succeeded, got %s", result.PaymentIntent.Status)
	}
	if result.Charge.CapturedAmount != 100 || result.Charge.Status != ChargeCaptured {
		t.Fatalf("expected captured charge, got %+v", result.Charge)
	}
}

func TestPaymentIntentAggregateRules(t *testing.T) {
	t.Run("rejects invalid confirm state", func(t *testing.T) {
		intent := PaymentIntent{Status: PaymentIntentSucceeded}
		if err := intent.CanConfirm(); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("rejects invalid capture state", func(t *testing.T) {
		intent := PaymentIntent{Status: PaymentIntentProcessing}
		if err := intent.CanCapture(); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("rejects capture without charge", func(t *testing.T) {
		intent := PaymentIntent{Status: PaymentIntentRequiresCapture}
		if err := intent.CanCapture(); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("rejects invalid finalize state", func(t *testing.T) {
		intent := PaymentIntent{Status: PaymentIntentProcessing, Scenario: string(ScenarioApprovedImmediate)}
		if err := intent.CanFinalizeProcessing(); err == nil {
			t.Fatalf("expected error")
		}
	})
}
