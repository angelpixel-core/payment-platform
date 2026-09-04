package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"payment-sandbox/internal/adapters/observability/metrics"
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
	metrics     metrics.MetricsRecorder
}

func NewStore(recorder metrics.MetricsRecorder) *MemoryStore {
	return &MemoryStore{
		intents:     make(map[string]domain.PaymentIntent),
		attempts:    make(map[string]domain.PaymentAttempt),
		charges:     make(map[string]domain.Charge),
		refunds:     make(map[string]domain.Refund),
		idempotency: make(map[string]idempotencyRecord),
		metrics:     recorder,
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
	start := time.Now()
	defer s.recordPersistence("payment_intent", "save", nil, start)
	return s.savePaymentIntent(intent)
}

func (s *MemoryStore) savePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.intents[intent.ID] = intent
	return intent
}

func (s *MemoryStore) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	start := time.Now()
	intent, err := s.getPaymentIntent(id)
	s.recordPersistence("payment_intent", "get", err, start)
	return intent, err
}

func (s *MemoryStore) ListPaymentIntents() []domain.PaymentIntent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.PaymentIntent, 0, len(s.intents))
	for _, intent := range s.intents {
		out = append(out, intent)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) getPaymentIntent(id string) (domain.PaymentIntent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.intents[id]
	if !ok {
		return domain.PaymentIntent{}, domain.NewError(404, "payment_intent_not_found", "payment intent not found")
	}
	return intent, nil
}

func (s *MemoryStore) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	start := time.Now()
	defer s.recordPersistence("payment_attempt", "save", nil, start)
	return s.savePaymentAttempt(attempt)
}

func (s *MemoryStore) savePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[attempt.ID] = attempt
	return attempt
}

func (s *MemoryStore) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	start := time.Now()
	attempt, err := s.getPaymentAttempt(id)
	s.recordPersistence("payment_attempt", "get", err, start)
	return attempt, err
}

func (s *MemoryStore) ListPaymentAttempts() []domain.PaymentAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.PaymentAttempt, 0, len(s.attempts))
	for _, attempt := range s.attempts {
		out = append(out, attempt)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RequestedAt.Equal(out[j].RequestedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].RequestedAt.Before(out[j].RequestedAt)
	})
	return out
}

func (s *MemoryStore) getPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt, ok := s.attempts[id]
	if !ok {
		return domain.PaymentAttempt{}, domain.NewError(404, "payment_attempt_not_found", "payment attempt not found")
	}
	return attempt, nil
}

func (s *MemoryStore) SaveCharge(charge domain.Charge) domain.Charge {
	start := time.Now()
	defer s.recordPersistence("charge", "save", nil, start)
	return s.saveCharge(charge)
}

func (s *MemoryStore) saveCharge(charge domain.Charge) domain.Charge {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.charges[charge.ID] = charge
	return charge
}

func (s *MemoryStore) GetCharge(id string) (domain.Charge, error) {
	start := time.Now()
	charge, err := s.getCharge(id)
	s.recordPersistence("charge", "get", err, start)
	return charge, err
}

func (s *MemoryStore) ListCharges() []domain.Charge {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Charge, 0, len(s.charges))
	for _, charge := range s.charges {
		out = append(out, charge)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) getCharge(id string) (domain.Charge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	charge, ok := s.charges[id]
	if !ok {
		return domain.Charge{}, domain.NewError(404, "charge_not_found", "charge not found")
	}
	return charge, nil
}

func (s *MemoryStore) SaveRefund(refund domain.Refund) domain.Refund {
	start := time.Now()
	defer s.recordPersistence("refund", "save", nil, start)
	return s.saveRefund(refund)
}

func (s *MemoryStore) saveRefund(refund domain.Refund) domain.Refund {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refunds[refund.ID] = refund
	return refund
}

func (s *MemoryStore) GetRefund(id string) (domain.Refund, error) {
	start := time.Now()
	refund, err := s.getRefund(id)
	s.recordPersistence("refund", "get", err, start)
	return refund, err
}

func (s *MemoryStore) ListRefunds() []domain.Refund {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Refund, 0, len(s.refunds))
	for _, refund := range s.refunds {
		out = append(out, refund)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) getRefund(id string) (domain.Refund, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	refund, ok := s.refunds[id]
	if !ok {
		return domain.Refund{}, domain.NewError(404, "refund_not_found", "refund not found")
	}
	return refund, nil
}

func (s *MemoryStore) recordPersistence(resource, operation string, err error, start time.Time) {
	if s == nil || s.metrics == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	s.metrics.RecordPersistenceOperation(
		context.Background(),
		"memory",
		resource,
		operation,
		outcome,
		time.Since(start),
	)
}
