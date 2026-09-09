package operations

import (
	"context"
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

const localSeedTime = "2025-01-01T00:00:00Z"

type SeedRunner struct {
	uow ports.UnitOfWork
}

func NewSeedRunner(uow ports.UnitOfWork) *SeedRunner { return &SeedRunner{uow: uow} }

func (r *SeedRunner) Simulate(context.Context, string) error { return ErrNotImplemented }
func (r *SeedRunner) Replay(context.Context, string) error   { return ErrNotImplemented }
func (r *SeedRunner) Burst(context.Context) error            { return ErrNotImplemented }

func (r *SeedRunner) Seed(ctx context.Context) error {
	if r == nil || r.uow == nil {
		return ErrNotImplemented
	}
	now, err := time.Parse(time.RFC3339, localSeedTime)
	if err != nil {
		return err
	}
	return r.uow.Do(func(tx ports.Transaction) error {
		for _, fixture := range localFixtures(now) {
			if err := ctx.Err(); err != nil {
				return err
			}
			tx.SavePaymentIntent(fixture.intent)
			tx.SavePaymentAttempt(fixture.attempt)
			if fixture.charge != nil {
				tx.SaveCharge(*fixture.charge)
			}
			if fixture.refund != nil {
				tx.SaveRefund(*fixture.refund)
			}
			for _, entry := range fixture.ledger {
				tx.AppendLedgerEntry(entry)
			}
		}
		return nil
	})
}

type seedFixture struct {
	intent  domain.PaymentIntent
	attempt domain.PaymentAttempt
	charge  *domain.Charge
	refund  *domain.Refund
	ledger  []domain.LedgerEntry
}

func localFixtures(now time.Time) []seedFixture {
	amount := domain.Amount(1000)
	currency := domain.Currency("usd")
	merchant := "merchant_seed"
	fixtures := []seedFixture{
		{intent: intent("pi_seed_approved", "pa_seed_approved", amount, currency, merchant, domain.PaymentIntentSucceeded, domain.ScenarioApprovedImmediate, "ch_seed_approved", now), attempt: attempt("pa_seed_approved", "pi_seed_approved", "pm_card_visa", domain.PaymentAttemptAuthorized, now), charge: charge("ch_seed_approved", "pi_seed_approved", "pa_seed_approved", amount, domain.ChargeCaptured, now)},
		{intent: intent("pi_seed_declined", "pa_seed_declined", amount, currency, merchant, domain.PaymentIntentFailed, domain.ScenarioDeclinedInsufficientFunds, "", now), attempt: attemptWithDecline("pa_seed_declined", "pi_seed_declined", "pm_card_insufficient_funds", now)},
		{intent: intent("pi_seed_requires_action", "pa_seed_requires_action", amount, currency, merchant, domain.PaymentIntentRequiresAction, domain.ScenarioRequiresAction3DS, "", now), attempt: attempt("pa_seed_requires_action", "pi_seed_requires_action", "pm_card_authentication_required", domain.PaymentAttemptRequiresAction, now)},
		{intent: intent("pi_seed_processing", "pa_seed_processing", amount, currency, merchant, domain.PaymentIntentProcessing, domain.ScenarioProcessingThenSucceeded, "", now), attempt: attempt("pa_seed_processing", "pi_seed_processing", "pm_card_processing", domain.PaymentAttemptSubmitted, now)},
	}
	fixtures[0].refund = &domain.Refund{ID: "re_seed_approved", ChargeID: "ch_seed_approved", PaymentIntentID: "pi_seed_approved", Amount: domain.Amount(250), Status: domain.RefundSucceeded, CreatedAt: now, UpdatedAt: now}
	fixtures[0].charge.RefundedAmount = domain.Amount(250)
	fixtures[0].charge.Status = domain.ChargePartiallyRefunded
	fixtures[0].ledger = []domain.LedgerEntry{{ID: "le_seed_approved", EventName: "payment_succeeded", EntityType: "payment_intent", EntityID: "pi_seed_approved", PaymentIntentID: "pi_seed_approved", ChargeID: "ch_seed_approved", MerchantID: merchant, Currency: currency, BalanceBucket: "available", BalanceDelta: int64(amount), Amount: amount, CreatedAt: now}}
	return fixtures
}

func intent(id, attemptID string, amount domain.Amount, currency domain.Currency, merchant string, status domain.PaymentIntentStatus, scenario domain.ScenarioName, chargeID string, now time.Time) domain.PaymentIntent {
	return domain.PaymentIntent{ID: id, MerchantID: merchant, Amount: amount, Currency: currency, CaptureMethod: "automatic", Status: status, Scenario: string(scenario), LatestAttemptID: attemptID, ChargeID: chargeID, CreatedAt: now, UpdatedAt: now}
}

func attempt(id, intentID, token string, status domain.PaymentAttemptStatus, now time.Time) domain.PaymentAttempt {
	return domain.PaymentAttempt{ID: id, PaymentIntentID: intentID, PaymentMethodToken: token, Status: status, ProcessorReference: "proc_" + id[3:], RequestedAt: now, RespondedAt: now}
}

func attemptWithDecline(id, intentID, token string, now time.Time) domain.PaymentAttempt {
	a := attempt(id, intentID, token, domain.PaymentAttemptDeclined, now)
	a.DeclineCode = "insufficient_funds"
	return a
}

func charge(id, intentID, attemptID string, amount domain.Amount, status domain.ChargeStatus, now time.Time) *domain.Charge {
	return &domain.Charge{ID: id, PaymentIntentID: intentID, PaymentAttemptID: attemptID, Amount: amount, CapturedAmount: amount, Status: status, CreatedAt: now, UpdatedAt: now}
}
