package payments

import (
	"sort"
	"strings"
	"time"

	"payment-sandbox/internal/domain"
)

type BalanceAccountType string

const (
	BalanceAccountAvailable  BalanceAccountType = "available"
	BalanceAccountReserved   BalanceAccountType = "reserved"
	BalanceAccountLiquidable BalanceAccountType = "liquidable"
)

type TransactionReportLine struct {
	PaymentIntent PaymentIntentView   `json:"payment_intent"`
	LatestAttempt *PaymentAttemptView `json:"latest_attempt,omitempty"`
	Charge        *ChargeView         `json:"charge,omitempty"`
	Refunds       []RefundView        `json:"refunds,omitempty"`
}

type SettlementBatchStatus string

const (
	SettlementBatchPending SettlementBatchStatus = "pending"
	SettlementBatchReady   SettlementBatchStatus = "ready"
)

type SettlementChargeLine struct {
	ChargeID        string    `json:"charge_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	CapturedAt      time.Time `json:"captured_at"`
	Amount          int64     `json:"amount"`
	RefundedAmount  int64     `json:"refunded_amount"`
}

type SettlementBatchLine struct {
	MerchantID     string                 `json:"merchant_id"`
	Currency       string                 `json:"currency"`
	SettlementDate string                 `json:"settlement_date"`
	Status         SettlementBatchStatus  `json:"status"`
	GrossAmount    int64                  `json:"gross_amount"`
	RefundedAmount int64                  `json:"refunded_amount"`
	NetAmount      int64                  `json:"net_amount"`
	ChargeCount    int                    `json:"charge_count"`
	RefundCount    int                    `json:"refund_count"`
	ChargeIDs      []string               `json:"charge_ids"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Charges        []SettlementChargeLine `json:"charges"`
}

type SettlementProjectionView struct {
	GeneratedAt time.Time             `json:"generated_at"`
	Batches     []SettlementBatchLine `json:"batches"`
	Count       int                   `json:"count"`
}

type BalanceProjectionLine struct {
	MerchantID  string             `json:"merchant_id"`
	Currency    string             `json:"currency"`
	AccountType BalanceAccountType `json:"account_type"`
	Amount      int64              `json:"amount"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type BalanceProjectionView struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Balances    []BalanceProjectionLine `json:"balances"`
	Count       int                     `json:"count"`
}

type TransactionReportView struct {
	GeneratedAt          time.Time                `json:"generated_at"`
	Transactions         []TransactionReportLine  `json:"transactions"`
	BalanceProjection    BalanceProjectionView    `json:"balance_projection"`
	SettlementProjection SettlementProjectionView `json:"settlement_projection"`
	Count                int                      `json:"count"`
}

func BuildTransactionReport(intents []domain.PaymentIntent, attempts []domain.PaymentAttempt, charges []domain.Charge, refunds []domain.Refund) TransactionReportView {
	attemptViews := make(map[string]PaymentAttemptView, len(attempts))
	for _, attempt := range attempts {
		attemptViews[attempt.ID] = PaymentAttemptView{
			ID:                 attempt.ID,
			PaymentIntentID:    attempt.PaymentIntentID,
			PaymentMethodToken: attempt.PaymentMethodToken,
			Status:             string(attempt.Status),
			DeclineCode:        attempt.DeclineCode,
			ProcessorReference: attempt.ProcessorReference,
			RequestedAt:        attempt.RequestedAt,
			RespondedAt:        attempt.RespondedAt,
		}
	}

	chargeByIntentID := make(map[string]ChargeView, len(charges))
	for _, charge := range charges {
		chargeByIntentID[charge.PaymentIntentID] = ChargeView{
			ID:               charge.ID,
			PaymentIntentID:  charge.PaymentIntentID,
			PaymentAttemptID: charge.PaymentAttemptID,
			Amount:           charge.Amount.Int64(),
			CapturedAmount:   charge.CapturedAmount.Int64(),
			RefundedAmount:   charge.RefundedAmount.Int64(),
			Status:           string(charge.Status),
			CreatedAt:        charge.CreatedAt,
			CapturedAt:       charge.CapturedAt,
			UpdatedAt:        charge.UpdatedAt,
		}
	}

	refundsByIntentID := make(map[string][]RefundView)
	for _, refund := range refunds {
		refundsByIntentID[refund.PaymentIntentID] = append(refundsByIntentID[refund.PaymentIntentID], RefundView{
			ID:              refund.ID,
			ChargeID:        refund.ChargeID,
			PaymentIntentID: refund.PaymentIntentID,
			Amount:          refund.Amount.Int64(),
			Status:          string(refund.Status),
			CreatedAt:       refund.CreatedAt,
			UpdatedAt:       refund.UpdatedAt,
		})
	}

	lines := make([]TransactionReportLine, 0, len(intents))
	for _, intent := range intents {
		line := TransactionReportLine{
			PaymentIntent: paymentIntentViewFromDomain(intent),
		}
		if intent.LatestAttemptID != "" {
			if attempt, ok := attemptViews[intent.LatestAttemptID]; ok {
				copyAttempt := attempt
				line.LatestAttempt = &copyAttempt
			}
		}
		if charge, ok := chargeByIntentID[intent.ID]; ok {
			copyCharge := charge
			line.Charge = &copyCharge
		}
		if refundLines, ok := refundsByIntentID[intent.ID]; ok {
			line.Refunds = append([]RefundView(nil), refundLines...)
		}
		lines = append(lines, line)
	}

	sort.Slice(lines, func(i, j int) bool {
		if lines[i].PaymentIntent.CreatedAt.Equal(lines[j].PaymentIntent.CreatedAt) {
			return lines[i].PaymentIntent.ID < lines[j].PaymentIntent.ID
		}
		return lines[i].PaymentIntent.CreatedAt.Before(lines[j].PaymentIntent.CreatedAt)
	})

	balance := BuildBalanceProjection(intents, charges)
	settlement := BuildSettlementProjection(intents, charges, refunds)
	return TransactionReportView{GeneratedAt: time.Now().UTC(), Transactions: lines, BalanceProjection: balance, SettlementProjection: settlement, Count: len(lines)}
}

func BuildBalanceProjection(intents []domain.PaymentIntent, charges []domain.Charge) BalanceProjectionView {
	now := time.Now().UTC()
	type bucket struct {
		line BalanceProjectionLine
	}

	chargeByIntentID := make(map[string]domain.Charge, len(charges))
	for _, charge := range charges {
		chargeByIntentID[charge.PaymentIntentID] = charge
	}

	balances := make(map[string]*bucket)
	ensure := func(merchantID, currency string, accountType BalanceAccountType) *bucket {
		key := balanceKey(merchantID, currency, accountType)
		if existing, ok := balances[key]; ok {
			return existing
		}
		b := &bucket{line: BalanceProjectionLine{MerchantID: merchantID, Currency: currency, AccountType: accountType, UpdatedAt: now}}
		balances[key] = b
		return b
	}

	for _, intent := range intents {
		merchantID := intent.MerchantID
		currency := intent.Currency.String()
		ensure(merchantID, currency, BalanceAccountAvailable)
		ensure(merchantID, currency, BalanceAccountReserved)
		ensure(merchantID, currency, BalanceAccountLiquidable)

		if intent.Status == domain.PaymentIntentRequiresCapture || intent.Status == domain.PaymentIntentProcessing {
			b := ensure(merchantID, currency, BalanceAccountReserved)
			b.line.Amount += intent.Amount.Int64()
			b.line.UpdatedAt = maxTime(b.line.UpdatedAt, intent.UpdatedAt)
		}

		charge, ok := chargeByIntentID[intent.ID]
		if !ok {
			continue
		}
		net := charge.CapturedAmount.Int64() - charge.RefundedAmount.Int64()
		if net <= 0 {
			continue
		}
		accountType := BalanceAccountLiquidable
		if strings.EqualFold(intent.CaptureMethod, "automatic") {
			accountType = BalanceAccountAvailable
		}
		b := ensure(merchantID, currency, accountType)
		b.line.Amount += net
		b.line.UpdatedAt = maxTime(b.line.UpdatedAt, charge.UpdatedAt)
	}

	out := make([]BalanceProjectionLine, 0, len(balances))
	for _, bucket := range balances {
		out = append(out, bucket.line)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].MerchantID == out[j].MerchantID {
			if out[i].Currency == out[j].Currency {
				return balanceAccountOrder(out[i].AccountType) < balanceAccountOrder(out[j].AccountType)
			}
			return out[i].Currency < out[j].Currency
		}
		return out[i].MerchantID < out[j].MerchantID
	})

	return BalanceProjectionView{GeneratedAt: now, Balances: out, Count: len(out)}
}

func BuildSettlementProjection(intents []domain.PaymentIntent, charges []domain.Charge, refunds []domain.Refund) SettlementProjectionView {
	now := time.Now().UTC()
	day := func(t time.Time) string {
		return t.UTC().Format("2006-01-02")
	}

	type batch struct {
		line SettlementBatchLine
	}

	intentByID := make(map[string]domain.PaymentIntent, len(intents))
	for _, intent := range intents {
		intentByID[intent.ID] = intent
	}

	refundsByChargeID := make(map[string][]domain.Refund, len(refunds))
	for _, refund := range refunds {
		refundsByChargeID[refund.ChargeID] = append(refundsByChargeID[refund.ChargeID], refund)
	}

	batches := make(map[string]*batch)
	ensure := func(merchantID, currency, settlementDate string) *batch {
		key := merchantID + "|" + currency + "|" + settlementDate
		if existing, ok := batches[key]; ok {
			return existing
		}
		b := &batch{line: SettlementBatchLine{MerchantID: merchantID, Currency: currency, SettlementDate: settlementDate, Status: settlementBatchStatusForDate(settlementDate, now), UpdatedAt: now}}
		batches[key] = b
		return b
	}

	for _, charge := range charges {
		if charge.CapturedAt == nil {
			continue
		}
		intent, ok := intentByID[charge.PaymentIntentID]
		if !ok || !strings.EqualFold(intent.CaptureMethod, "manual") {
			continue
		}
		settlementDate := day(*charge.CapturedAt)
		merchantID := intent.MerchantID
		currency := intent.Currency.String()
		b := ensure(merchantID, currency, settlementDate)
		gross := charge.CapturedAmount.Int64()
		refunded := charge.RefundedAmount.Int64()
		net := gross - refunded
		b.line.GrossAmount += gross
		b.line.RefundedAmount += refunded
		b.line.NetAmount += net
		b.line.ChargeCount++
		b.line.RefundCount += len(refundsByChargeID[charge.ID])
		b.line.ChargeIDs = append(b.line.ChargeIDs, charge.ID)
		b.line.UpdatedAt = maxTime(b.line.UpdatedAt, maxTime(*charge.CapturedAt, charge.UpdatedAt))
		b.line.Charges = append(b.line.Charges, SettlementChargeLine{ChargeID: charge.ID, PaymentIntentID: charge.PaymentIntentID, CapturedAt: *charge.CapturedAt, Amount: gross, RefundedAmount: refunded})
	}

	out := make([]SettlementBatchLine, 0, len(batches))
	for _, batch := range batches {
		sort.Strings(batch.line.ChargeIDs)
		sort.Slice(batch.line.Charges, func(i, j int) bool { return batch.line.Charges[i].ChargeID < batch.line.Charges[j].ChargeID })
		out = append(out, batch.line)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].MerchantID == out[j].MerchantID {
			if out[i].Currency == out[j].Currency {
				return out[i].SettlementDate < out[j].SettlementDate
			}
			return out[i].Currency < out[j].Currency
		}
		return out[i].MerchantID < out[j].MerchantID
	})

	return SettlementProjectionView{GeneratedAt: now, Batches: out, Count: len(out)}
}

func paymentIntentViewFromDomain(intent domain.PaymentIntent) PaymentIntentView {
	return PaymentIntentView{
		ID:              intent.ID,
		Status:          string(intent.Status),
		Amount:          intent.Amount.Int64(),
		Currency:        intent.Currency.String(),
		ChargeID:        intent.ChargeID,
		LatestAttemptID: intent.LatestAttemptID,
		CreatedAt:       intent.CreatedAt,
		UpdatedAt:       intent.UpdatedAt,
	}
}

func balanceKey(merchantID, currency string, accountType BalanceAccountType) string {
	return merchantID + "|" + currency + "|" + string(accountType)
}

func balanceAccountOrder(accountType BalanceAccountType) int {
	switch accountType {
	case BalanceAccountAvailable:
		return 0
	case BalanceAccountReserved:
		return 1
	case BalanceAccountLiquidable:
		return 2
	default:
		return 99
	}
}

func maxTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if b.After(a) {
		return b
	}
	return a
}

func settlementBatchStatusForDate(settlementDate string, now time.Time) SettlementBatchStatus {
	if settlementDate == now.UTC().Format("2006-01-02") {
		return SettlementBatchPending
	}
	if settlementDate < now.UTC().Format("2006-01-02") {
		return SettlementBatchReady
	}
	return SettlementBatchPending
}
