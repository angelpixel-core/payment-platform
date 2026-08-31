package payments

import (
	"testing"

	"payment-sandbox/internal/adapters/persistence/memory"
	"payment-sandbox/internal/domain"
)

func TestPaymentQueryService(t *testing.T) {
	store := memory.NewStore()
	query := NewPaymentQueryService(store)

	intent := domain.PaymentIntent{ID: "pi_1", Amount: 100, Currency: "USD", Status: domain.PaymentIntentSucceeded}
	attempt := domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1", Status: domain.PaymentAttemptAuthorized}
	charge := domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1", Amount: 100, Status: domain.ChargeCaptured}
	refund := domain.Refund{ID: "re_1", ChargeID: "ch_1", Amount: 100, Status: domain.RefundSucceeded}

	store.SavePaymentIntent(intent)
	store.SavePaymentAttempt(attempt)
	store.SaveCharge(charge)
	store.SaveRefund(refund)

	gotIntent, err := query.GetPaymentIntent("pi_1")
	if err != nil {
		t.Fatalf("get intent failed: %v", err)
	}
	if gotIntent.ID != intent.ID || gotIntent.Amount != 100 || gotIntent.Currency != "USD" {
		t.Fatalf("unexpected intent: %+v", gotIntent)
	}

	gotAttempt, err := query.GetPaymentAttempt("pa_1")
	if err != nil {
		t.Fatalf("get attempt failed: %v", err)
	}
	if gotAttempt.ID != attempt.ID || gotAttempt.PaymentIntentID != attempt.PaymentIntentID {
		t.Fatalf("unexpected attempt: %+v", gotAttempt)
	}

	gotCharge, err := query.GetCharge("ch_1")
	if err != nil {
		t.Fatalf("get charge failed: %v", err)
	}
	if gotCharge.ID != charge.ID || gotCharge.Status != string(charge.Status) {
		t.Fatalf("unexpected charge: %+v", gotCharge)
	}

	gotRefund, err := query.GetRefund("re_1")
	if err != nil {
		t.Fatalf("get refund failed: %v", err)
	}
	if gotRefund.ID != refund.ID || gotRefund.Status != string(refund.Status) {
		t.Fatalf("unexpected refund: %+v", gotRefund)
	}
}
