package memory

import (
	"fmt"
	"sync"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type idempotencyRecord struct {
	fingerprint string
	value       any
}

type MemoryStore struct {
	mu          sync.Mutex
	seq         int64
	intents     map[string]domain.PaymentIntent
	attempts    map[string]domain.PaymentAttempt
	charges     map[string]domain.Charge
	refunds     map[string]domain.Refund
	idempotency map[string]idempotencyRecord
}

func NewStore() *MemoryStore {
	return &MemoryStore{
		intents:     make(map[string]domain.PaymentIntent),
		attempts:    make(map[string]domain.PaymentAttempt),
		charges:     make(map[string]domain.Charge),
		refunds:     make(map[string]domain.Refund),
		idempotency: make(map[string]idempotencyRecord),
	}
}

var _ ports.Store = (*MemoryStore)(nil)

func (s *MemoryStore) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	s.mu.Lock()

	if existing, ok := s.idempotency[key]; ok {
		s.mu.Unlock()
		if existing.fingerprint != fingerprint {
			return nil, domain.NewError(409, "idempotency_conflict", "idempotency key was already used with a different request")
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

func (s *MemoryStore) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intents[intent.ID] = intent
	return intent
}

func (s *MemoryStore) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.intents[id]
	if !ok {
		return domain.PaymentIntent{}, domain.NewError(404, "payment_intent_not_found", "payment intent not found")
	}
	return intent, nil
}

func (s *MemoryStore) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[attempt.ID] = attempt
	return attempt
}

func (s *MemoryStore) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt, ok := s.attempts[id]
	if !ok {
		return domain.PaymentAttempt{}, domain.NewError(404, "payment_attempt_not_found", "payment attempt not found")
	}
	return attempt, nil
}

func (s *MemoryStore) SaveCharge(charge domain.Charge) domain.Charge {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.charges[charge.ID] = charge
	return charge
}

func (s *MemoryStore) GetCharge(id string) (domain.Charge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	charge, ok := s.charges[id]
	if !ok {
		return domain.Charge{}, domain.NewError(404, "charge_not_found", "charge not found")
	}
	return charge, nil
}

func (s *MemoryStore) SaveRefund(refund domain.Refund) domain.Refund {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refunds[refund.ID] = refund
	return refund
}

func (s *MemoryStore) GetRefund(id string) (domain.Refund, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	refund, ok := s.refunds[id]
	if !ok {
		return domain.Refund{}, domain.NewError(404, "refund_not_found", "refund not found")
	}
	return refund, nil
}
