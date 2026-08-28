package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu          sync.Mutex
	now         func() time.Time
	seq         int64
	intents     map[string]*PaymentIntent
	attempts    map[string]*PaymentAttempt
	charges     map[string]*Charge
	refunds     map[string]*Refund
	idempotency map[string]idempotencyRecord
}

type idempotencyRecord struct {
	fingerprint string
	value       any
}

func NewService() *Service {
	return &Service{
		now:         time.Now,
		intents:     make(map[string]*PaymentIntent),
		attempts:    make(map[string]*PaymentAttempt),
		charges:     make(map[string]*Charge),
		refunds:     make(map[string]*Refund),
		idempotency: make(map[string]idempotencyRecord),
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
		intent := &PaymentIntent{
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
		s.mu.Lock()
		s.intents[intent.ID] = intent
		s.mu.Unlock()
		return *clonePaymentIntent(intent), nil
	})
	if err != nil {
		return PaymentIntent{}, err
	}
	return result.(PaymentIntent), nil
}

func (s *Service) ConfirmPaymentIntent(intentID string, req ConfirmPaymentIntentRequest, idempotencyKey, fingerprint string) (ConfirmPaymentIntentResponse, error) {
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

		now := s.now()
		attempt := &PaymentAttempt{
			ID:                 s.nextID("pa"),
			PaymentIntentID:    intent.ID,
			PaymentMethodToken: req.PaymentMethodToken,
			Status:             PaymentAttemptAuthorized,
			ProcessorReference: s.nextReference("pr"),
			RequestedAt:        now,
			RespondedAt:        now,
		}
		s.mu.Lock()
		s.attempts[attempt.ID] = attempt
		s.mu.Unlock()

		charge := &Charge{
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

		intent.LatestAttemptID = attempt.ID
		intent.ChargeID = charge.ID
		intent.UpdatedAt = now
		charge.UpdatedAt = now
		s.mu.Lock()
		s.charges[charge.ID] = charge
		s.mu.Unlock()
		return ConfirmPaymentIntentResponse{
			PaymentIntent:  *clonePaymentIntent(intent),
			PaymentAttempt: *clonePaymentAttempt(attempt),
			Charge:         cloneCharge(charge),
		}, nil
	})
	if err != nil {
		return ConfirmPaymentIntentResponse{}, err
	}
	return result.(ConfirmPaymentIntentResponse), nil
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

	return CapturePaymentIntentResponse{
		PaymentIntent: *clonePaymentIntent(intent),
		Charge:        *cloneCharge(charge),
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
		refund := &Refund{
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
		s.mu.Lock()
		s.refunds[refund.ID] = refund
		s.mu.Unlock()
		return RefundResponse{
			Refund: *cloneRefund(refund),
			Charge: *cloneCharge(charge),
		}, nil
	})
	if err != nil {
		return RefundResponse{}, err
	}
	return result.(RefundResponse), nil
}

func (s *Service) withIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	s.mu.Lock()

	if existing, ok := s.idempotency[key]; ok {
		s.mu.Unlock()
		if existing.fingerprint != fingerprint {
			return nil, newError(409, "idempotency_conflict", "idempotency key was already used with a different request")
		}
		return existing.value, nil
	}
	s.mu.Unlock()

	value, err := fn()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idempotency[key] = idempotencyRecord{fingerprint: fingerprint, value: value}
	return value, nil
}

func (s *Service) intent(id string) (*PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.intents[id]
	if !ok {
		return nil, newError(404, "payment_intent_not_found", "payment intent not found")
	}
	return intent, nil
}

func (s *Service) charge(id string) (*Charge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	charge, ok := s.charges[id]
	if !ok {
		return nil, newError(404, "charge_not_found", "charge not found")
	}
	return charge, nil
}

func (s *Service) nextID(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}

func (s *Service) nextReference(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
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
