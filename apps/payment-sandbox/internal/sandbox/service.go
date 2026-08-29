package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type Service struct {
	now            func() time.Time
	scenarioEngine *ScenarioEngine
	store          Store
}

type idempotencyRecord struct {
	fingerprint string
	value       any
}

func NewService() *Service {
	return &Service{
		now:            time.Now,
		scenarioEngine: NewScenarioEngine(),
		store:          NewMemoryStore(),
	}
}

func (s *Service) CreatePaymentIntent(req CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (PaymentIntent, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return PaymentIntent{}, newError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if req.Amount <= 0 {
		return PaymentIntent{}, newError(400, "invalid_amount", "amount must be greater than zero")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return PaymentIntent{}, newError(400, "invalid_currency", "currency is required")
	}

	key := "create_payment_intent:" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		now := s.now()
		intent := PaymentIntent{
			ID:             s.nextID("pi"),
			MerchantID:     req.MerchantID,
			CustomerID:     req.CustomerID,
			Amount:         req.Amount,
			Currency:       strings.ToUpper(req.Currency),
			CaptureMethod:  normalizeCaptureMethod(req.CaptureMethod),
			Status:         PaymentIntentRequiresPaymentMethod,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		s.store.SavePaymentIntent(intent)
		return intent, nil
	})
	if err != nil {
		return PaymentIntent{}, err
	}
	return result.(PaymentIntent), nil
}

func (s *Service) ConfirmPaymentIntent(intentID string, req ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (ConfirmPaymentIntentResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return ConfirmPaymentIntentResponse{}, newError(400, "missing_idempotency_key", "idempotency key is required")
	}
	key := "confirm_payment_intent:" + intentID + ":" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		intent, err := s.intent(intentID)
		if err != nil {
			return nil, err
		}
		if intent.Status != PaymentIntentRequiresPaymentMethod && intent.Status != PaymentIntentRequiresConfirmation {
			return nil, newError(409, "invalid_intent_state", "payment intent cannot be confirmed in its current state")
		}

		scenarioName, err := s.scenarioEngine.Resolve(scenarioHeader, req.PaymentMethodToken)
		if err != nil {
			return nil, err
		}
		outcome, err := s.scenarioEngine.Outcome(scenarioName)
		if err != nil {
			return nil, err
		}

		now := s.now()
		attempt := PaymentAttempt{
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
			return ConfirmPaymentIntentResponse{
				PaymentIntent:  *intent,
				PaymentAttempt: attempt,
				Charge:         nil,
			}, nil
		}

		charge := Charge{
			ID:               s.nextID("ch"),
			PaymentIntentID:  intent.ID,
			PaymentAttemptID: attempt.ID,
			Amount:           intent.Amount,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		switch outcome.Scenario {
		case ScenarioApprovedImmediate:
			charge.Status = outcome.ChargeStatus
			if intent.CaptureMethod == "automatic" {
				charge.CapturedAmount = intent.Amount
				charge.Status = ChargeCaptured
				intent.Status = PaymentIntentSucceeded
			} else {
				intent.Status = PaymentIntentRequiresCapture
			}
		case ScenarioDeclinedInsufficientFunds:
			intent.Status = outcome.IntentStatus
			attempt.DeclineCode = outcome.DeclineCode
			charge = Charge{}
		case ScenarioRequiresAction3DS:
			intent.Status = outcome.IntentStatus
			charge = Charge{}
		default:
			return nil, newError(422, "invalid_scenario", "scenario not supported")
		}

		var confirmedCharge *Charge
		if charge.ID != "" {
			intent.ChargeID = charge.ID
			charge.UpdatedAt = now
			s.store.SaveCharge(charge)
			confirmedCharge = &charge
		}
		s.store.SavePaymentIntent(*intent)
		s.store.SavePaymentAttempt(attempt)
		return ConfirmPaymentIntentResponse{
			PaymentIntent:  *intent,
			PaymentAttempt: attempt,
			Charge:         confirmedCharge,
		}, nil
	})
	if err != nil {
		return ConfirmPaymentIntentResponse{}, err
	}
	return result.(ConfirmPaymentIntentResponse), nil
}

func (s *Service) FinalizeProcessingPaymentIntent(intentID string) (PaymentIntent, error) {
	intent, err := s.intent(intentID)
	if err != nil {
		return PaymentIntent{}, err
	}
	if intent.Status != PaymentIntentProcessing {
		return PaymentIntent{}, newError(409, "invalid_intent_state", "payment intent cannot be finalized in its current state")
	}
	if normalizeScenarioName(intent.Scenario) != ScenarioProcessingThenSucceeded {
		return PaymentIntent{}, newError(409, "invalid_scenario", "payment intent is not in the processing scenario")
	}

	attempt, err := s.paymentAttempt(intent.LatestAttemptID)
	if err != nil {
		return PaymentIntent{}, err
	}
	if attempt.Status != PaymentAttemptSubmitted {
		return PaymentIntent{}, newError(409, "invalid_attempt_state", "processing payment intent cannot be finalized in its current attempt state")
	}

	now := s.now()
	charge := Charge{
		ID:               s.nextID("ch"),
		PaymentIntentID:  intent.ID,
		PaymentAttemptID: attempt.ID,
		Amount:           intent.Amount,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if intent.CaptureMethod == "automatic" {
		charge.CapturedAmount = intent.Amount
		charge.Status = ChargeCaptured
		intent.Status = PaymentIntentSucceeded
	} else {
		charge.Status = ChargeAuthorized
		intent.Status = PaymentIntentRequiresCapture
	}
	intent.ChargeID = charge.ID
	intent.UpdatedAt = now
	attempt.Status = PaymentAttemptAuthorized
	attempt.RespondedAt = now
	charge.UpdatedAt = now

	s.store.SaveCharge(charge)
	s.store.SavePaymentAttempt(*attempt)
	s.store.SavePaymentIntent(*intent)

	return *intent, nil
}

func (s *Service) CapturePaymentIntent(intentID string, req CapturePaymentIntentRequest) (CapturePaymentIntentResponse, error) {
	intent, err := s.intent(intentID)
	if err != nil {
		return CapturePaymentIntentResponse{}, err
	}
	if intent.Status != PaymentIntentRequiresCapture {
		return CapturePaymentIntentResponse{}, newError(409, "invalid_intent_state", "payment intent cannot be captured in its current state")
	}
	charge, err := s.charge(intent.ChargeID)
	if err != nil {
		return CapturePaymentIntentResponse{}, err
	}
	if charge.Status != ChargeAuthorized {
		return CapturePaymentIntentResponse{}, newError(409, "invalid_charge_state", "charge cannot be captured in its current state")
	}

	amount := req.Amount
	if amount == 0 {
		amount = charge.Amount
	}
	if amount <= 0 {
		return CapturePaymentIntentResponse{}, newError(400, "invalid_amount", "capture amount must be greater than zero")
	}
	if amount > charge.Amount {
		return CapturePaymentIntentResponse{}, newError(400, "invalid_amount", "capture amount cannot exceed charge amount")
	}

	now := s.now()
	charge.CapturedAmount = amount
	charge.Status = ChargeCaptured
	charge.UpdatedAt = now
	intent.Status = PaymentIntentSucceeded
	intent.UpdatedAt = now
	s.store.SaveCharge(*charge)
	s.store.SavePaymentIntent(*intent)

	return CapturePaymentIntentResponse{
		PaymentIntent: *intent,
		Charge:        *charge,
	}, nil
}

func (s *Service) CreateRefund(req RefundRequest, idempotencyKey, fingerprint string) (RefundResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return RefundResponse{}, newError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if strings.TrimSpace(req.ChargeID) == "" {
		return RefundResponse{}, newError(400, "invalid_charge_id", "charge_id is required")
	}

	key := "create_refund:" + idempotencyKey
	result, err := s.withIdempotency(key, fingerprint, func() (any, error) {
		charge, err := s.charge(req.ChargeID)
		if err != nil {
			return nil, err
		}
		if charge.CapturedAmount == 0 {
			return nil, newError(409, "invalid_charge_state", "charge must be captured before refunding")
		}

		remaining := charge.CapturedAmount - charge.RefundedAmount
		amount := req.Amount
		if amount == 0 {
			amount = remaining
		}
		if amount <= 0 {
			return nil, newError(400, "invalid_amount", "refund amount must be greater than zero")
		}
		if amount > remaining {
			return nil, newError(400, "invalid_amount", "refund amount cannot exceed remaining captured amount")
		}

		now := s.now()
		refund := Refund{
			ID:              s.nextID("re"),
			ChargeID:        charge.ID,
			PaymentIntentID: charge.PaymentIntentID,
			Amount:          amount,
			Status:          RefundSucceeded,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		charge.RefundedAmount += amount
		if charge.RefundedAmount == charge.CapturedAmount {
			charge.Status = ChargeRefunded
		} else {
			charge.Status = ChargePartiallyRefunded
		}
		charge.UpdatedAt = now
		s.store.SaveRefund(refund)
		s.store.SaveCharge(*charge)
		return RefundResponse{
			Refund: refund,
			Charge: *charge,
		}, nil
	})
	if err != nil {
		return RefundResponse{}, err
	}
	return result.(RefundResponse), nil
}

func (s *Service) withIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	return s.store.WithIdempotency(key, fingerprint, fn)
}

func (s *Service) intent(id string) (*PaymentIntent, error) {
	intent, err := s.store.GetPaymentIntent(id)
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

func (s *Service) charge(id string) (*Charge, error) {
	charge, err := s.store.GetCharge(id)
	if err != nil {
		return nil, err
	}
	return &charge, nil
}

func (s *Service) paymentAttempt(id string) (*PaymentAttempt, error) {
	attempt, err := s.store.GetPaymentAttempt(id)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (s *Service) nextID(prefix string) string {
	return s.store.NextID(prefix)
}

func (s *Service) nextReference(prefix string) string {
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

func clonePaymentIntent(src *PaymentIntent) *PaymentIntent {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func clonePaymentAttempt(src *PaymentAttempt) *PaymentAttempt {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func cloneCharge(src *Charge) *Charge {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func cloneRefund(src *Refund) *Refund {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func Fingerprint(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func FingerprintString(value string) string {
	return Fingerprint([]byte(value))
}
