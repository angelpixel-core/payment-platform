package payments

import (
	"strings"
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type PaymentService struct {
	uow            ports.UnitOfWork
	clock          ports.Clock
	scenarioEngine ports.ScenarioResolver
}

func NewService(uow ports.UnitOfWork, clock ports.Clock, scenarioEngine ports.ScenarioResolver) *PaymentService {
	return &PaymentService{uow: uow, clock: clock, scenarioEngine: scenarioEngine}
}

func (s *PaymentService) CreatePaymentIntent(req domain.CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (domain.PaymentIntent, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.PaymentIntent{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if req.Amount <= 0 {
		return domain.PaymentIntent{}, domain.NewError(400, "invalid_amount", "amount must be greater than zero")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return domain.PaymentIntent{}, domain.NewError(400, "invalid_currency", "currency is required")
	}
	amount, err := domain.NewAmount(req.Amount)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	currency, err := domain.NewCurrency(req.Currency)
	if err != nil {
		return domain.PaymentIntent{}, err
	}

	var intent domain.PaymentIntent
	err = s.uow.Do(func(tx ports.Transaction) error {
		key := "create_payment_intent:" + idempotencyKey
		result, err := tx.WithIdempotency(key, fingerprint, func() (any, error) {
			now := s.clock.Now()
			intent := domain.PaymentIntent{ID: tx.NextID("pi"), MerchantID: req.MerchantID, CustomerID: req.CustomerID, Amount: amount, Currency: currency, CaptureMethod: normalizeCaptureMethod(req.CaptureMethod), Status: domain.PaymentIntentRequiresPaymentMethod, IdempotencyKey: idempotencyKey, CreatedAt: now, UpdatedAt: now}
			tx.SavePaymentIntent(intent)
			tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.created", "payment_intent", intent.ID, intent.ID, "", "", "", intent.MerchantID, intent.Currency, "", 0, intent.Amount, now))
			_ = tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: intent})
			return intent, nil
		})
		if err != nil {
			return err
		}
		intent = result.(domain.PaymentIntent)
		return nil
	})
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	return intent, nil
}

func (s *PaymentService) ConfirmPaymentIntent(intentID string, req domain.ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (domain.ConfirmPaymentIntentResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.ConfirmPaymentIntentResponse{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	var response domain.ConfirmPaymentIntentResponse
	err := s.uow.Do(func(tx ports.Transaction) error {
		key := "confirm_payment_intent:" + intentID + ":" + idempotencyKey
		result, err := tx.WithIdempotency(key, fingerprint, func() (any, error) {
			intent, err := tx.GetPaymentIntent(intentID)
			if err != nil {
				return nil, err
			}

			now := s.clock.Now()
			scenarioName, err := s.scenarioEngine.Resolve(scenarioHeader, req.PaymentMethodToken)
			if err != nil {
				return nil, err
			}
			outcome, err := s.scenarioEngine.Outcome(scenarioName)
			if err != nil {
				return nil, err
			}

			attempt := domain.PaymentAttempt{ID: tx.NextID("pa"), PaymentIntentID: intent.ID, PaymentMethodToken: req.PaymentMethodToken, Status: outcome.AttemptStatus, ProcessorReference: tx.NextReference("pr"), RequestedAt: now, RespondedAt: now}
			var charge *domain.Charge
			if outcome.CreatesCharge {
				charge = &domain.Charge{ID: tx.NextID("ch")}
			}
			result, err := intent.Confirm(domain.ConfirmPaymentIntentCommand{Outcome: outcome, Attempt: &attempt, Charge: charge, Now: now})
			if err != nil {
				return nil, err
			}

			tx.SavePaymentIntent(result.PaymentIntent)
			tx.SavePaymentAttempt(result.PaymentAttempt)
			if result.Charge != nil {
				tx.SaveCharge(*result.Charge)
			}
			appendConfirmLedgerEntries(tx, result.PaymentIntent, result.PaymentAttempt, result.Charge, now)
			_ = tx.Publish(domain.PaymentIntentConfirmedEvent{PaymentIntent: result.PaymentIntent, PaymentAttempt: result.PaymentAttempt, Charge: result.Charge})
			return domain.ConfirmPaymentIntentResponse{PaymentIntent: result.PaymentIntent, PaymentAttempt: result.PaymentAttempt, Charge: result.Charge}, nil
		})
		if err != nil {
			return err
		}
		response = result.(domain.ConfirmPaymentIntentResponse)
		return nil
	})
	if err != nil {
		return domain.ConfirmPaymentIntentResponse{}, err
	}
	return response, nil
}

func (s *PaymentService) FinalizeProcessingPaymentIntent(intentID string) (domain.PaymentIntent, error) {
	var intentResult domain.PaymentIntent
	err := s.uow.Do(func(tx ports.Transaction) error {
		intent, err := tx.GetPaymentIntent(intentID)
		if err != nil {
			return err
		}

		attempt, err := tx.GetPaymentAttempt(intent.LatestAttemptID)
		if err != nil {
			return err
		}

		now := s.clock.Now()
		charge := domain.Charge{ID: tx.NextID("ch")}
		result, err := intent.FinalizeProcessing(domain.FinalizeProcessingCommand{Attempt: &attempt, Charge: &charge, Now: now})
		if err != nil {
			return err
		}

		tx.SaveCharge(result.Charge)
		tx.SavePaymentAttempt(result.PaymentAttempt)
		tx.SavePaymentIntent(result.PaymentIntent)
		appendFinalizeLedgerEntries(tx, result.PaymentIntent, result.PaymentAttempt, result.Charge, now)
		_ = tx.Publish(domain.PaymentIntentFinalizedEvent{PaymentIntent: result.PaymentIntent})
		intentResult = result.PaymentIntent
		return nil
	})
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	return intentResult, nil
}

func (s *PaymentService) CapturePaymentIntent(intentID string, req domain.CapturePaymentIntentRequest, fingerprint string) (domain.CapturePaymentIntentResponse, error) {
	var response domain.CapturePaymentIntentResponse
	err := s.uow.Do(func(tx ports.Transaction) error {
		key := "capture_payment_intent:" + intentID + ":" + strings.TrimSpace(req.IdempotencyKey)
		result, err := tx.WithIdempotency(key, fingerprint, func() (any, error) {
			intent, err := tx.GetPaymentIntent(intentID)
			if err != nil {
				return nil, err
			}
			if err := intent.CanCapture(); err != nil {
				return nil, err
			}
			charge, err := tx.GetCharge(intent.ChargeID)
			if err != nil {
				return nil, err
			}
			intentForCharge, err := tx.GetPaymentIntent(charge.PaymentIntentID)
			if err != nil {
				return nil, err
			}

			now := s.clock.Now()
			result, err := intent.Capture(domain.CapturePaymentIntentCommand{Charge: &charge, Amount: domain.Amount(req.Amount), Now: now})
			if err != nil {
				return nil, err
			}
			tx.SaveCharge(result.Charge)
			tx.SavePaymentIntent(result.PaymentIntent)
			appendCaptureLedgerEntries(tx, result.PaymentIntent, result.Charge, intentForCharge.MerchantID, intentForCharge.Currency, now)
			_ = tx.Publish(domain.PaymentIntentCapturedEvent{PaymentIntent: result.PaymentIntent, Charge: result.Charge})
			return domain.CapturePaymentIntentResponse{PaymentIntent: result.PaymentIntent, Charge: result.Charge}, nil
		})
		if err != nil {
			return err
		}
		response = result.(domain.CapturePaymentIntentResponse)
		return nil
	})
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	return response, nil
}

func normalizeCaptureMethod(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || (value != "manual" && value != "automatic") {
		return "manual"
	}
	return value
}

func chargeID(charge *domain.Charge) string {
	if charge == nil {
		return ""
	}
	return charge.ID
}

func appendConfirmLedgerEntries(tx ports.Transaction, intent domain.PaymentIntent, attempt domain.PaymentAttempt, charge *domain.Charge, now time.Time) {
	if intent.Status == domain.PaymentIntentProcessing {
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.confirmed", "payment_intent", intent.ID, intent.ID, attempt.ID, chargeID(charge), "", intent.MerchantID, intent.Currency, "reserved", int64(intent.Amount), intent.Amount, now))
		return
	}
	if intent.Status == domain.PaymentIntentRequiresCapture {
		if charge == nil {
			return
		}
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.confirmed", "payment_intent", intent.ID, intent.ID, attempt.ID, charge.ID, "", intent.MerchantID, intent.Currency, "reserved", int64(intent.Amount), intent.Amount, now))
		return
	}
	if intent.Status == domain.PaymentIntentSucceeded && charge != nil {
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.confirmed", "payment_intent", intent.ID, intent.ID, attempt.ID, charge.ID, "", intent.MerchantID, intent.Currency, "available", int64(charge.CapturedAmount), charge.CapturedAmount, now))
	}
}

func appendFinalizeLedgerEntries(tx ports.Transaction, intent domain.PaymentIntent, attempt domain.PaymentAttempt, charge domain.Charge, now time.Time) {
	if intent.CaptureMethod == "automatic" {
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.finalized", "payment_intent", intent.ID, intent.ID, attempt.ID, charge.ID, "", intent.MerchantID, intent.Currency, "reserved", -int64(intent.Amount), intent.Amount, now))
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.finalized", "payment_intent", intent.ID, intent.ID, attempt.ID, charge.ID, "", intent.MerchantID, intent.Currency, "available", int64(charge.CapturedAmount), charge.CapturedAmount, now))
		return
	}
	// Manual processing finalization keeps the same reserved bucket until capture.
}

func appendCaptureLedgerEntries(tx ports.Transaction, intent domain.PaymentIntent, charge domain.Charge, merchantID string, currency domain.Currency, now time.Time) {
	if strings.EqualFold(intent.CaptureMethod, "automatic") {
		tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.captured", "charge", charge.ID, intent.ID, "", charge.ID, "", merchantID, currency, "available", int64(charge.CapturedAmount), charge.CapturedAmount, now))
		return
	}
	tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.captured", "charge", charge.ID, intent.ID, "", charge.ID, "", merchantID, currency, "reserved", -int64(charge.CapturedAmount), charge.CapturedAmount, now))
	tx.AppendLedgerEntry(ledgerEntry(tx.NextID("le"), "payment_intent.captured", "charge", charge.ID, intent.ID, "", charge.ID, "", merchantID, currency, "liquidable", int64(charge.CapturedAmount), charge.CapturedAmount, now))
}

func ledgerEntry(id, eventName, entityType, entityID, paymentIntentID, paymentAttemptID, chargeID, refundID, merchantID string, currency domain.Currency, bucket string, delta int64, amount domain.Amount, now time.Time) domain.LedgerEntry {
	return domain.LedgerEntry{ID: id, EventName: eventName, EntityType: entityType, EntityID: entityID, PaymentIntentID: paymentIntentID, PaymentAttemptID: paymentAttemptID, ChargeID: chargeID, RefundID: refundID, MerchantID: merchantID, Currency: currency, BalanceBucket: bucket, BalanceDelta: delta, Amount: amount, CreatedAt: now}
}
