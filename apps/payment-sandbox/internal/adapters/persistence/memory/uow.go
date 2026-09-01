package memory

import (
	"sync"
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type UnitOfWork struct {
	store     *MemoryStore
	publisher ports.EventPublisher
}

type transaction struct {
	store     *MemoryStore
	publisher ports.EventPublisher
	intents   map[string]domain.PaymentIntent
	attempts  map[string]domain.PaymentAttempt
	charges   map[string]domain.Charge
	refunds   map[string]domain.Refund
	events    []domain.Event
	idem      map[string]idempotencyRecord
	mu        sync.Mutex
}

func NewUnitOfWork(store *MemoryStore, publisher ports.EventPublisher) *UnitOfWork {
	return &UnitOfWork{store: store, publisher: publisher}
}

var _ ports.UnitOfWork = (*UnitOfWork)(nil)
var _ ports.Transaction = (*transaction)(nil)

func (u *UnitOfWork) Do(fn func(tx ports.Transaction) error) error {
	tx := &transaction{store: u.store, publisher: u.publisher}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.commit()
}

func (tx *transaction) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	tx.store.mu.Lock()
	if existing, ok := tx.store.idempotency[key]; ok {
		tx.store.mu.Unlock()
		if existing.fingerprint != fingerprint {
			return nil, domain.NewError(409, "idempotency_conflict", "idempotency key was already used with a different request")
		}
		return existing.value, nil
	}
	tx.store.mu.Unlock()

	value, err := fn()
	if err != nil {
		return nil, err
	}
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.idem == nil {
		tx.idem = make(map[string]idempotencyRecord)
	}
	tx.idem[key] = idempotencyRecord{fingerprint: fingerprint, value: value}
	return value, nil
}

func (tx *transaction) NextID(prefix string) string        { return tx.store.NextID(prefix) }
func (tx *transaction) NextReference(prefix string) string { return tx.store.NextReference(prefix) }

func (tx *transaction) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	start := time.Now()
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.intents == nil {
		tx.intents = make(map[string]domain.PaymentIntent)
	}
	tx.intents[intent.ID] = intent
	tx.store.recordPersistence("payment_intent", "save", nil, start)
	return intent
}

func (tx *transaction) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	start := time.Now()
	tx.mu.Lock()
	if intent, ok := tx.intents[id]; ok {
		tx.mu.Unlock()
		tx.store.recordPersistence("payment_intent", "get", nil, start)
		return intent, nil
	}
	tx.mu.Unlock()
	intent, err := tx.store.getPaymentIntent(id)
	tx.store.recordPersistence("payment_intent", "get", err, start)
	return intent, err
}

func (tx *transaction) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	start := time.Now()
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.attempts == nil {
		tx.attempts = make(map[string]domain.PaymentAttempt)
	}
	tx.attempts[attempt.ID] = attempt
	tx.store.recordPersistence("payment_attempt", "save", nil, start)
	return attempt
}

func (tx *transaction) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	start := time.Now()
	tx.mu.Lock()
	if attempt, ok := tx.attempts[id]; ok {
		tx.mu.Unlock()
		tx.store.recordPersistence("payment_attempt", "get", nil, start)
		return attempt, nil
	}
	tx.mu.Unlock()
	attempt, err := tx.store.getPaymentAttempt(id)
	tx.store.recordPersistence("payment_attempt", "get", err, start)
	return attempt, err
}

func (tx *transaction) SaveCharge(charge domain.Charge) domain.Charge {
	start := time.Now()
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.charges == nil {
		tx.charges = make(map[string]domain.Charge)
	}
	tx.charges[charge.ID] = charge
	tx.store.recordPersistence("charge", "save", nil, start)
	return charge
}

func (tx *transaction) GetCharge(id string) (domain.Charge, error) {
	start := time.Now()
	tx.mu.Lock()
	if charge, ok := tx.charges[id]; ok {
		tx.mu.Unlock()
		tx.store.recordPersistence("charge", "get", nil, start)
		return charge, nil
	}
	tx.mu.Unlock()
	charge, err := tx.store.getCharge(id)
	tx.store.recordPersistence("charge", "get", err, start)
	return charge, err
}

func (tx *transaction) SaveRefund(refund domain.Refund) domain.Refund {
	start := time.Now()
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.refunds == nil {
		tx.refunds = make(map[string]domain.Refund)
	}
	tx.refunds[refund.ID] = refund
	tx.store.recordPersistence("refund", "save", nil, start)
	return refund
}

func (tx *transaction) GetRefund(id string) (domain.Refund, error) {
	start := time.Now()
	tx.mu.Lock()
	if refund, ok := tx.refunds[id]; ok {
		tx.mu.Unlock()
		tx.store.recordPersistence("refund", "get", nil, start)
		return refund, nil
	}
	tx.mu.Unlock()
	refund, err := tx.store.getRefund(id)
	tx.store.recordPersistence("refund", "get", err, start)
	return refund, err
}

func (tx *transaction) Publish(event domain.Event) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.events = append(tx.events, event)
	return nil
}

func (tx *transaction) commit() error {
	tx.store.mu.Lock()
	backup := snapshotStore(tx.store)
	applyTransactionState(tx.store, tx)
	tx.store.mu.Unlock()

	for _, event := range tx.events {
		if tx.publisher != nil {
			if err := tx.publisher.Publish(event); err != nil {
				tx.store.mu.Lock()
				restoreStore(tx.store, backup)
				tx.store.mu.Unlock()
				return err
			}
		}
	}
	return nil
}

type storeSnapshot struct {
	seq         int64
	intents     map[string]domain.PaymentIntent
	attempts    map[string]domain.PaymentAttempt
	charges     map[string]domain.Charge
	refunds     map[string]domain.Refund
	idempotency map[string]idempotencyRecord
}

func snapshotStore(store *MemoryStore) storeSnapshot {
	return storeSnapshot{
		seq:         store.seq,
		intents:     cloneIntents(store.intents),
		attempts:    cloneAttempts(store.attempts),
		charges:     cloneCharges(store.charges),
		refunds:     cloneRefunds(store.refunds),
		idempotency: cloneIdempotency(store.idempotency),
	}
}

func applyTransactionState(store *MemoryStore, tx *transaction) {
	if tx.intents != nil {
		if store.intents == nil {
			store.intents = make(map[string]domain.PaymentIntent)
		}
		for k, v := range tx.intents {
			store.intents[k] = v
		}
	}
	if tx.attempts != nil {
		if store.attempts == nil {
			store.attempts = make(map[string]domain.PaymentAttempt)
		}
		for k, v := range tx.attempts {
			store.attempts[k] = v
		}
	}
	if tx.charges != nil {
		if store.charges == nil {
			store.charges = make(map[string]domain.Charge)
		}
		for k, v := range tx.charges {
			store.charges[k] = v
		}
	}
	if tx.refunds != nil {
		if store.refunds == nil {
			store.refunds = make(map[string]domain.Refund)
		}
		for k, v := range tx.refunds {
			store.refunds[k] = v
		}
	}
	if tx.idem != nil {
		if store.idempotency == nil {
			store.idempotency = make(map[string]idempotencyRecord)
		}
		for k, v := range tx.idem {
			store.idempotency[k] = v
		}
	}
}

func restoreStore(store *MemoryStore, snapshot storeSnapshot) {
	store.seq = snapshot.seq
	store.intents = snapshot.intents
	store.attempts = snapshot.attempts
	store.charges = snapshot.charges
	store.refunds = snapshot.refunds
	store.idempotency = snapshot.idempotency
}

func cloneIntents(src map[string]domain.PaymentIntent) map[string]domain.PaymentIntent {
	if src == nil {
		return nil
	}
	dst := make(map[string]domain.PaymentIntent, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneAttempts(src map[string]domain.PaymentAttempt) map[string]domain.PaymentAttempt {
	if src == nil {
		return nil
	}
	dst := make(map[string]domain.PaymentAttempt, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneCharges(src map[string]domain.Charge) map[string]domain.Charge {
	if src == nil {
		return nil
	}
	dst := make(map[string]domain.Charge, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneRefunds(src map[string]domain.Refund) map[string]domain.Refund {
	if src == nil {
		return nil
	}
	dst := make(map[string]domain.Refund, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneIdempotency(src map[string]idempotencyRecord) map[string]idempotencyRecord {
	if src == nil {
		return nil
	}
	dst := make(map[string]idempotencyRecord, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
