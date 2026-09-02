package payments

import (
	"strings"

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

			scenarioName, err := s.scenarioEngine.Resolve(scenarioHeader, req.PaymentMethodToken)
			if err != nil {
				return nil, err
			}
			outcome, err := s.scenarioEngine.Outcome(scenarioName)
			if err != nil {
				return nil, err
			}

			now := s.clock.Now()
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
		_ = tx.Publish(domain.PaymentIntentFinalizedEvent{PaymentIntent: result.PaymentIntent})
		intentResult = result.PaymentIntent
		return nil
	})
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	return intentResult, nil
}

func (s *PaymentService) CapturePaymentIntent(intentID string, req domain.CapturePaymentIntentRequest) (domain.CapturePaymentIntentResponse, error) {
	var response domain.CapturePaymentIntentResponse
		err := s.uow.Do(func(tx ports.Transaction) error {
			intent, err := tx.GetPaymentIntent(intentID)
			if err != nil {
				return err
			}
			if err := intent.CanCapture(); err != nil {
				return err
			}
			charge, err := tx.GetCharge(intent.ChargeID)
			if err != nil {
				return err
			}

		now := s.clock.Now()
		result, err := intent.Capture(domain.CapturePaymentIntentCommand{Charge: &charge, Amount: domain.Amount(req.Amount), Now: now})
		if err != nil {
			return err
		}
		tx.SaveCharge(result.Charge)
		tx.SavePaymentIntent(result.PaymentIntent)
		_ = tx.Publish(domain.PaymentIntentCapturedEvent{PaymentIntent: result.PaymentIntent, Charge: result.Charge})
		response = domain.CapturePaymentIntentResponse{PaymentIntent: result.PaymentIntent, Charge: result.Charge}
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
