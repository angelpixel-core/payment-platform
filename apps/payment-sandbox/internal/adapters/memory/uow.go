package memory

import (
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
}

func NewUnitOfWork(store *MemoryStore, publisher ports.EventPublisher) *UnitOfWork {
	return &UnitOfWork{store: store, publisher: publisher}
}

var _ ports.UnitOfWork = (*UnitOfWork)(nil)
var _ ports.Transaction = (*transaction)(nil)

func (u *UnitOfWork) Do(fn func(tx ports.Transaction) error) error {
	return fn(&transaction{store: u.store, publisher: u.publisher})
}

func (tx *transaction) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	return tx.store.WithIdempotency(key, fingerprint, fn)
}
func (tx *transaction) NextID(prefix string) string        { return tx.store.NextID(prefix) }
func (tx *transaction) NextReference(prefix string) string { return tx.store.NextReference(prefix) }
func (tx *transaction) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	return tx.store.SavePaymentIntent(intent)
}
func (tx *transaction) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	return tx.store.GetPaymentIntent(id)
}
func (tx *transaction) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	return tx.store.SavePaymentAttempt(attempt)
}
func (tx *transaction) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	return tx.store.GetPaymentAttempt(id)
}
func (tx *transaction) SaveCharge(charge domain.Charge) domain.Charge {
	return tx.store.SaveCharge(charge)
}
func (tx *transaction) GetCharge(id string) (domain.Charge, error) { return tx.store.GetCharge(id) }
func (tx *transaction) SaveRefund(refund domain.Refund) domain.Refund {
	return tx.store.SaveRefund(refund)
}
func (tx *transaction) GetRefund(id string) (domain.Refund, error) { return tx.store.GetRefund(id) }
func (tx *transaction) Publish(event domain.Event) error {
	if tx.publisher == nil || event == nil {
		return nil
	}
	return tx.publisher.Publish(event)
}
