package payments

import (
	"testing"
	"time"

	"payment-sandbox/internal/adapters/persistence/memory"
	"payment-sandbox/internal/domain"
)

func TestPaymentQueryService(t *testing.T) {
	store := memory.NewStore(nil)
	query := NewPaymentQueryService(store)

	intent := domain.PaymentIntent{ID: "pi_1", Amount: 100, Currency: "USD", CaptureMethod: "manual", Status: domain.PaymentIntentSucceeded, ChargeID: "ch_1", LatestAttemptID: "pa_1"}
	attempt := domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1", Status: domain.PaymentAttemptAuthorized}
	capturedAt := time.Now().UTC().Truncate(time.Second)
	charge := domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1", Amount: 100, CapturedAmount: 100, CapturedAt: &capturedAt, Status: domain.ChargeCaptured}
	refund := domain.Refund{ID: "re_1", ChargeID: "ch_1", PaymentIntentID: "pi_1", Amount: 100, Status: domain.RefundSucceeded}

	store.SavePaymentIntent(intent)
	store.SavePaymentAttempt(attempt)
	store.SaveCharge(charge)
	store.SaveRefund(refund)
	charge.RefundedAmount = 100
	charge.Status = domain.ChargeRefunded
	store.SaveCharge(charge)

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

	lifecycle, err := query.GetPaymentLifecycle("pi_1")
	if err != nil {
		t.Fatalf("get lifecycle failed: %v", err)
	}
	if lifecycle.Status != "refunded" {
		t.Fatalf("expected refunded lifecycle, got %s", lifecycle.Status)
	}
	if lifecycle.RefundableAmount != 0 || lifecycle.IsRefundable {
		t.Fatalf("expected non-refundable lifecycle, got %+v", lifecycle)
	}
	if lifecycle.LatestAttempt == nil || lifecycle.Charge == nil {
		t.Fatalf("expected nested attempt and charge in lifecycle: %+v", lifecycle)
	}

	report, err := query.GetTransactionReport()
	if err != nil {
		t.Fatalf("get report failed: %v", err)
	}
	if report.Count != 1 || len(report.Transactions) != 1 {
		t.Fatalf("expected one report line, got %+v", report)
	}
	if report.BalanceProjection.Count != 3 || len(report.BalanceProjection.Balances) != 3 {
		t.Fatalf("expected three balance projection lines, got %+v", report.BalanceProjection)
	}
	if report.SettlementProjection.Count != 1 || len(report.SettlementProjection.Batches) != 1 {
		t.Fatalf("expected one settlement batch, got %+v", report.SettlementProjection)
	}
	settlement := report.SettlementProjection.Batches[0]
	if settlement.MerchantID != intent.MerchantID || settlement.Currency != intent.Currency.String() {
		t.Fatalf("unexpected settlement scope: %+v", settlement)
	}
	if settlement.GrossAmount != 100 || settlement.RefundedAmount != 100 || settlement.NetAmount != 0 || settlement.ChargeCount != 1 || settlement.RefundCount != 1 {
		t.Fatalf("unexpected settlement batch: %+v", settlement)
	}
	if len(settlement.ChargeIDs) != 1 || settlement.ChargeIDs[0] != charge.ID {
		t.Fatalf("unexpected settlement charge ids: %+v", settlement)
	}
	line := report.Transactions[0]
	if line.PaymentIntent.ID != intent.ID {
		t.Fatalf("unexpected report intent: %+v", line.PaymentIntent)
	}
	if line.LatestAttempt == nil || line.Charge == nil || len(line.Refunds) != 1 {
		t.Fatalf("expected report to include attempt, charge, and refund: %+v", line)
	}
}

func TestTransactionReportBalanceProjection(t *testing.T) {
	store := memory.NewStore(nil)
	query := NewPaymentQueryService(store)

	store.SavePaymentIntent(domain.PaymentIntent{
		ID:            "pi_reserved",
		MerchantID:    "merchant_1",
		Amount:        100,
		Currency:      "usd",
		CaptureMethod: "manual",
		Status:        domain.PaymentIntentRequiresCapture,
	})
	store.SavePaymentIntent(domain.PaymentIntent{
		ID:            "pi_available",
		MerchantID:    "merchant_1",
		Amount:        200,
		Currency:      "usd",
		CaptureMethod: "automatic",
		Status:        domain.PaymentIntentSucceeded,
	})
	store.SaveCharge(domain.Charge{
		ID:              "ch_available",
		PaymentIntentID: "pi_available",
		Amount:          200,
		CapturedAmount:  200,
		CapturedAt:      func() *time.Time { v := time.Now().UTC().Truncate(time.Second); return &v }(),
		Status:          domain.ChargeCaptured,
	})
	store.SavePaymentIntent(domain.PaymentIntent{
		ID:            "pi_liquidable",
		MerchantID:    "merchant_1",
		Amount:        300,
		Currency:      "usd",
		CaptureMethod: "manual",
		Status:        domain.PaymentIntentSucceeded,
		ChargeID:      "ch_liquidable",
	})
	store.SaveCharge(domain.Charge{
		ID:              "ch_liquidable",
		PaymentIntentID: "pi_liquidable",
		Amount:          300,
		CapturedAmount:  300,
		CapturedAt:      func() *time.Time { v := time.Now().UTC().Truncate(time.Second); return &v }(),
		Status:          domain.ChargeCaptured,
	})

	report, err := query.GetTransactionReport()
	if err != nil {
		t.Fatalf("get report failed: %v", err)
	}
	if report.BalanceProjection.Count != 3 {
		t.Fatalf("expected three balance lines, got %+v", report.BalanceProjection)
	}
	if report.SettlementProjection.Count != 1 || len(report.SettlementProjection.Batches) != 1 {
		t.Fatalf("expected one settlement batch, got %+v", report.SettlementProjection)
	}

	balances := make(map[BalanceAccountType]int64)
	for _, line := range report.BalanceProjection.Balances {
		if line.MerchantID != "merchant_1" || line.Currency != "usd" {
			t.Fatalf("unexpected balance line scope: %+v", line)
		}
		balances[line.AccountType] = line.Amount
	}
	if balances[BalanceAccountReserved] != 100 {
		t.Fatalf("expected reserved 100, got %+v", balances)
	}
	if balances[BalanceAccountAvailable] != 200 {
		t.Fatalf("expected available 200, got %+v", balances)
	}
	if balances[BalanceAccountLiquidable] != 300 {
		t.Fatalf("expected liquidable 300, got %+v", balances)
	}
	batch := report.SettlementProjection.Batches[0]
	if batch.GrossAmount != 300 || batch.NetAmount != 300 || batch.ChargeCount != 1 {
		t.Fatalf("unexpected settlement batch: %+v", batch)
	}
}
