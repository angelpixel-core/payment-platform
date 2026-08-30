package application

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type ScenarioResolver interface {
	Resolve(headerScenario, paymentMethodToken string) (domain.ScenarioName, error)
	Outcome(name domain.ScenarioName) (domain.ScenarioOutcome, error)
}

type PaymentService struct {
	clock          ports.Clock
	scenarioEngine ScenarioResolver
	store          ports.Store
	events         ports.EventPublisher
}

type idempotencyRecord struct {
	fingerprint string
	value       any
}

func NewPaymentService(store ports.Store, scenarioEngine ScenarioResolver, events ports.EventPublisher) *PaymentService {
	return &PaymentService{
		clock:          systemClock{},
		scenarioEngine: scenarioEngine,
		store:          store,
		events:         events,
	}
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
		intent := domain.PaymentIntent{
			ID:             s.nextID("pi"),
			MerchantID:     req.MerchantID,
			CustomerID:     req.CustomerID,
			Amount:         req.Amount,
			Currency:       strings.ToUpper(req.Currency),
			CaptureMethod:  normalizeCaptureMethod(req.CaptureMethod),
			Status:         domain.PaymentIntentRequiresPaymentMethod,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		s.store.SavePaymentIntent(intent)
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
		if intent.Status != domain.PaymentIntentRequiresPaymentMethod && intent.Status != domain.PaymentIntentRequiresConfirmation {
			return nil, domain.NewError(409, "invalid_intent_state", "payment intent cannot be confirmed in its current state")
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
		attempt := domain.PaymentAttempt{
			ID:                 s.nextID("pa"),
			PaymentIntentID:    intent.ID,
			PaymentMethodToken: req.PaymentMethodToken,
			Status:             outcome.AttemptStatus,
			ProcessorReference: s.nextReference("pr"),
			RequestedAt:        now,
			RespondedAt:        now,
		}
		s.store.SavePaymentAttempt(attempt)

		intent.Scenario = string(outcome.Scenario)
		intent.LatestAttemptID = attempt.ID
		intent.UpdatedAt = now

		if outcome.FinalizesLater {
			intent.Status = outcome.IntentStatus
			s.store.SavePaymentIntent(*intent)
			return domain.ConfirmPaymentIntentResponse{
				PaymentIntent:  *intent,
				PaymentAttempt: attempt,
				Charge:         nil,
			}, nil
		}

		charge := domain.Charge{
			ID:               s.nextID("ch"),
			PaymentIntentID:  intent.ID,
			PaymentAttemptID: attempt.ID,
			Amount:           intent.Amount,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		switch outcome.Scenario {
		case domain.ScenarioApprovedImmediate:
			charge.Status = outcome.ChargeStatus
			if intent.CaptureMethod == "automatic" {
				charge.CapturedAmount = intent.Amount
				charge.Status = domain.ChargeCaptured
				intent.Status = domain.PaymentIntentSucceeded
			} else {
				intent.Status = domain.PaymentIntentRequiresCapture
			}
		case domain.ScenarioDeclinedInsufficientFunds:
			intent.Status = outcome.IntentStatus
			attempt.DeclineCode = outcome.DeclineCode
			charge = domain.Charge{}
		case domain.ScenarioRequiresAction3DS:
			intent.Status = outcome.IntentStatus
			charge = domain.Charge{}
		default:
			return nil, domain.NewError(422, "invalid_scenario", "scenario not supported")
		}

		var confirmedCharge *domain.Charge
		if charge.ID != "" {
			intent.ChargeID = charge.ID
			charge.UpdatedAt = now
			s.store.SaveCharge(charge)
			confirmedCharge = &charge
		}
		s.store.SavePaymentIntent(*intent)
		s.store.SavePaymentAttempt(attempt)
		return domain.ConfirmPaymentIntentResponse{
			PaymentIntent:  *intent,
			PaymentAttempt: attempt,
			Charge:         confirmedCharge,
		}, nil
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
	if intent.Status != domain.PaymentIntentProcessing {
		return domain.PaymentIntent{}, domain.NewError(409, "invalid_intent_state", "payment intent cannot be finalized in its current state")
	}
	if domain.NormalizeScenarioName(intent.Scenario) != domain.ScenarioProcessingThenSucceeded {
		return domain.PaymentIntent{}, domain.NewError(409, "invalid_scenario", "payment intent is not in the processing scenario")
	}

	attempt, err := s.paymentAttempt(intent.LatestAttemptID)
	if err != nil {
		return domain.PaymentIntent{}, err
	}
	if attempt.Status != domain.PaymentAttemptSubmitted {
		return domain.PaymentIntent{}, domain.NewError(409, "invalid_attempt_state", "processing payment intent cannot be finalized in its current attempt state")
	}

	now := s.clock.Now()
	charge := domain.Charge{
		ID:               s.nextID("ch"),
		PaymentIntentID:  intent.ID,
		PaymentAttemptID: attempt.ID,
		Amount:           intent.Amount,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if intent.CaptureMethod == "automatic" {
		charge.CapturedAmount = intent.Amount
		charge.Status = domain.ChargeCaptured
		intent.Status = domain.PaymentIntentSucceeded
	} else {
		charge.Status = domain.ChargeAuthorized
		intent.Status = domain.PaymentIntentRequiresCapture
	}
	intent.ChargeID = charge.ID
	intent.UpdatedAt = now
	attempt.Status = domain.PaymentAttemptAuthorized
	attempt.RespondedAt = now
	charge.UpdatedAt = now

	s.store.SaveCharge(charge)
	s.store.SavePaymentAttempt(*attempt)
	s.store.SavePaymentIntent(*intent)

	return *intent, nil
}

func (s *PaymentService) CapturePaymentIntent(intentID string, req domain.CapturePaymentIntentRequest) (domain.CapturePaymentIntentResponse, error) {
	intent, err := s.intent(intentID)
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	if intent.Status != domain.PaymentIntentRequiresCapture {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(409, "invalid_intent_state", "payment intent cannot be captured in its current state")
	}
	charge, err := s.charge(intent.ChargeID)
	if err != nil {
		return domain.CapturePaymentIntentResponse{}, err
	}
	if charge.Status != domain.ChargeAuthorized {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(409, "invalid_charge_state", "charge cannot be captured in its current state")
	}

	amount := req.Amount
	if amount == 0 {
		amount = charge.Amount
	}
	if amount <= 0 {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(400, "invalid_amount", "capture amount must be greater than zero")
	}
	if amount > charge.Amount {
		return domain.CapturePaymentIntentResponse{}, domain.NewError(400, "invalid_amount", "capture amount cannot exceed charge amount")
	}

	now := s.clock.Now()
	charge.CapturedAmount = amount
	charge.Status = domain.ChargeCaptured
	charge.UpdatedAt = now
	intent.Status = domain.PaymentIntentSucceeded
	intent.UpdatedAt = now
	s.store.SaveCharge(*charge)
	s.store.SavePaymentIntent(*intent)

	return domain.CapturePaymentIntentResponse{
		PaymentIntent: *intent,
		Charge:        *charge,
	}, nil
}

func (s *PaymentService) CreateRefund(req domain.RefundRequest, idempotencyKey, fingerprint string) (domain.RefundResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if strings.TrimSpace(req.ChargeID) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "invalid_charge_id", "charge_id is required")
	}

	key := "create_refund:" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		charge, err := s.charge(req.ChargeID)
		if err != nil {
			return nil, err
		}
		if charge.CapturedAmount == 0 {
			return nil, domain.NewError(409, "invalid_charge_state", "charge must be captured before refunding")
		}

		remaining := charge.CapturedAmount - charge.RefundedAmount
		amount := req.Amount
		if amount == 0 {
			amount = remaining
		}
		if amount <= 0 {
			return nil, domain.NewError(400, "invalid_amount", "refund amount must be greater than zero")
		}
		if amount > remaining {
			return nil, domain.NewError(400, "invalid_amount", "refund amount cannot exceed remaining captured amount")
		}

		now := s.clock.Now()
		refund := domain.Refund{
			ID:              s.nextID("re"),
			ChargeID:        charge.ID,
			PaymentIntentID: charge.PaymentIntentID,
			Amount:          amount,
			Status:          domain.RefundSucceeded,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		charge.RefundedAmount += amount
		if charge.RefundedAmount == charge.CapturedAmount {
			charge.Status = domain.ChargeRefunded
		} else {
			charge.Status = domain.ChargePartiallyRefunded
		}
		charge.UpdatedAt = now
		s.store.SaveRefund(refund)
		s.store.SaveCharge(*charge)
		return domain.RefundResponse{
			Refund: refund,
			Charge: *charge,
		}, nil
	})
	if err != nil {
		return domain.RefundResponse{}, err
	}
	return result.(domain.RefundResponse), nil
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

func (s *PaymentService) nextID(prefix string) string {
	return s.store.NextID(prefix)
}

func (s *PaymentService) nextReference(prefix string) string {
	return s.store.NextReference(prefix)
}

func normalizeCaptureMethod(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "manual"
	}
	if value != "manual" && value != "automatic" {
		return "manual"
	}
	return value
}

func Fingerprint(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func FingerprintString(value string) string {
	return Fingerprint([]byte(value))
}
