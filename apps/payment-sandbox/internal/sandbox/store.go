package sandbox

import (
	"fmt"
	"sync"
)

type Store interface {
	WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error)
	NextID(prefix string) string
	NextReference(prefix string) string

	SavePaymentIntent(intent PaymentIntent) PaymentIntent
	GetPaymentIntent(id string) (PaymentIntent, error)
	SavePaymentAttempt(attempt PaymentAttempt) PaymentAttempt
	GetPaymentAttempt(id string) (PaymentAttempt, error)
	SaveCharge(charge Charge) Charge
	GetCharge(id string) (Charge, error)
	SaveRefund(refund Refund) Refund
	GetRefund(id string) (Refund, error)
}

type MemoryStore struct {
	mu          sync.Mutex
	seq         int64
	intents     map[string]PaymentIntent
	attempts    map[string]PaymentAttempt
	charges     map[string]Charge
	refunds     map[string]Refund
	idempotency map[string]idempotencyRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		intents:     make(map[string]PaymentIntent),
		attempts:    make(map[string]PaymentAttempt),
		charges:     make(map[string]Charge),
		refunds:     make(map[string]Refund),
		idempotency: make(map[string]idempotencyRecord),
	}
}

func (s *MemoryStore) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
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

func (s *MemoryStore) NextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}

func (s *MemoryStore) NextReference(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}

func (s *MemoryStore) SavePaymentIntent(intent PaymentIntent) PaymentIntent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intents[intent.ID] = intent
	return intent
}

func (s *MemoryStore) GetPaymentIntent(id string) (PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.intents[id]
	if !ok {
		return PaymentIntent{}, newError(404, "payment_intent_not_found", "payment intent not found")
	}
	return intent, nil
}

func (s *MemoryStore) SavePaymentAttempt(attempt PaymentAttempt) PaymentAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[attempt.ID] = attempt
	return attempt
}

func (s *MemoryStore) GetPaymentAttempt(id string) (PaymentAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt, ok := s.attempts[id]
	if !ok {
		return PaymentAttempt{}, newError(404, "payment_attempt_not_found", "payment attempt not found")
	}
	return attempt, nil
}

func (s *MemoryStore) SaveCharge(charge Charge) Charge {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.charges[charge.ID] = charge
	return charge
}

func (s *MemoryStore) GetCharge(id string) (Charge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	charge, ok := s.charges[id]
	if !ok {
		return Charge{}, newError(404, "charge_not_found", "charge not found")
	}
	return charge, nil
}

func (s *MemoryStore) SaveRefund(refund Refund) Refund {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refunds[refund.ID] = refund
	return refund
}

func (s *MemoryStore) GetRefund(id string) (Refund, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	refund, ok := s.refunds[id]
	if !ok {
		return Refund{}, newError(404, "refund_not_found", "refund not found")
	}
	return refund, nil
}
