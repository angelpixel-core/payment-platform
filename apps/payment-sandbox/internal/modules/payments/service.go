package payments

import (
	"strings"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type PaymentService struct {
	clock          ports.Clock
	scenarioEngine ports.ScenarioResolver
	store          ports.Store
	events         ports.EventPublisher
}

func NewService(store ports.Store, clock ports.Clock, scenarioEngine ports.ScenarioResolver, events ports.EventPublisher) *PaymentService {
	return &PaymentService{clock: clock, scenarioEngine: scenarioEngine, store: store, events: events}
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

	key := "create_payment_intent:" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		now := s.clock.Now()
		intent := domain.PaymentIntent{ID: s.nextID("pi"), MerchantID: req.MerchantID, CustomerID: req.CustomerID, Amount: req.Amount, Currency: strings.ToUpper(req.Currency), CaptureMethod: normalizeCaptureMethod(req.CaptureMethod), Status: domain.PaymentIntentRequiresPaymentMethod, IdempotencyKey: idempotencyKey, CreatedAt: now, UpdatedAt: now}
		s.store.SavePaymentIntent(intent)
		s.publish(domain.PaymentIntentCreatedEvent{PaymentIntent: intent})
		return intent, nil
	})
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	return result.(domain.PaymentIntent), nil
}

func (s *PaymentService) ConfirmPaymentIntent(intentID string, req domain.ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (domain.ConfirmPaymentIntentResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.ConfirmPaymentIntentResponse{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	key := "confirm_payment_intent:" + intentID + ":" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		intent, err := s.intent(intentID)
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
		attempt := domain.PaymentAttempt{ID: s.nextID("pa"), PaymentIntentID: intent.ID, PaymentMethodToken: req.PaymentMethodToken, Status: outcome.AttemptStatus, ProcessorReference: s.nextReference("pr"), RequestedAt: now, RespondedAt: now}
		var charge *domain.Charge
		if outcome.CreatesCharge {
			charge = &domain.Charge{ID: s.nextID("ch")}
		}
		result, err := intent.Confirm(outcome, &attempt, charge, now)
		if err != nil {
			return nil, err
		}

		s.store.SavePaymentIntent(result.PaymentIntent)
		s.store.SavePaymentAttempt(result.PaymentAttempt)
		if result.Charge != nil {
			s.store.SaveCharge(*result.Charge)
		}
		s.publish(domain.PaymentIntentConfirmedEvent{PaymentIntent: result.PaymentIntent, PaymentAttempt: result.PaymentAttempt, Charge: result.Charge})
		return domain.ConfirmPaymentIntentResponse{PaymentIntent: result.PaymentIntent, PaymentAttempt: result.PaymentAttempt, Charge: result.Charge}, nil
	})
	if err != nil {
		return domain.ConfirmPaymentIntentResponse{}, err
	}
	return result.(domain.ConfirmPaymentIntentResponse), nil
}

func (s *PaymentService) FinalizeProcessingPaymentIntent(intentID string) (domain.PaymentIntent, error) {
	intent, err := s.intent(intentID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}

	attempt, err := s.paymentAttempt(intent.LatestAttemptID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}

	now := s.clock.Now()
	charge := domain.Charge{ID: s.nextID("ch")}
	result, err := intent.FinalizeProcessing(attempt, &charge, now)
	if err != nil {
		return domain.PaymentIntent{}, err
	}

	s.store.SaveCharge(result.Charge)
	s.store.SavePaymentAttempt(result.PaymentAttempt)
	s.store.SavePaymentIntent(result.PaymentIntent)
	s.publish(domain.PaymentIntentFinalizedEvent{PaymentIntent: result.PaymentIntent})

	return result.PaymentIntent, nil
}

func (s *PaymentService) CapturePaymentIntent(intentID string, req domain.CapturePaymentIntentRequest) (domain.CapturePaymentIntentResponse, error) {
	intent, err := s.intent(intentID)
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	if intent.ChargeID == "" {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(409, "invalid_intent_state", "payment intent cannot be captured in its current state")
	}
	charge, err := s.charge(intent.ChargeID)
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	if charge.Status != domain.ChargeAuthorized {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(409, "invalid_charge_state", "charge cannot be captured in its current state")
	}

	now := s.clock.Now()
	result, err := intent.Capture(charge, req.Amount, now)
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	s.store.SaveCharge(result.Charge)
	s.store.SavePaymentIntent(result.PaymentIntent)
	s.publish(domain.PaymentIntentCapturedEvent{PaymentIntent: result.PaymentIntent, Charge: result.Charge})

	return domain.CapturePaymentIntentResponse{PaymentIntent: result.PaymentIntent, Charge: result.Charge}, nil
}

func (s *PaymentService) withIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	return s.store.WithIdempotency(key, fingerprint, fn)
}
func (s *PaymentService) intent(id string) (*domain.PaymentIntent, error) {
	intent, err := s.store.GetPaymentIntent(id)
	if err != nil {
		return nil, err
	}
	return &intent, nil
}
func (s *PaymentService) charge(id string) (*domain.Charge, error) {
	charge, err := s.store.GetCharge(id)
	if err != nil {
		return nil, err
	}
	return &charge, nil
}
func (s *PaymentService) paymentAttempt(id string) (*domain.PaymentAttempt, error) {
	attempt, err := s.store.GetPaymentAttempt(id)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}
func (s *PaymentService) nextID(prefix string) string        { return s.store.NextID(prefix) }
func (s *PaymentService) nextReference(prefix string) string { return s.store.NextReference(prefix) }
func (s *PaymentService) publish(event domain.Event) {
	if s.events != nil && event != nil {
		_ = s.events.Publish(event)
	}
}
func normalizeCaptureMethod(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || (value != "manual" && value != "automatic") {
		return "manual"
	}
	return value
}
