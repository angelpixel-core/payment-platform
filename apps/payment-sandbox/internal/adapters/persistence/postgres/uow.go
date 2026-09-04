package postgres

import (
	"context"
	"database/sql"
	"time"

	"payment-sandbox/internal/adapters/observability/metrics"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type UnitOfWork struct {
	store     *Store
	publisher ports.EventPublisher
}

func NewUnitOfWork(db *sql.DB, publisher ports.EventPublisher, recorder metrics.MetricsRecorder) *UnitOfWork {
	return &UnitOfWork{store: &Store{db: db, metrics: recorder}, publisher: publisher}
}

var _ ports.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Do(fn func(tx ports.Transaction) error) (err error) {
	start := time.Now()
	outcome := "success"
	defer func() {
		if u.store != nil {
			u.store.recordUnitOfWork(outcome, time.Since(start))
		}
	}()
	tx, err := u.store.db.BeginTx(context.Background(), nil)
	if err != nil {
		outcome = "commit_error"
		return err
	}
	inner := &transaction{store: u.store, tx: tx, publisher: u.publisher}
	if err = fn(inner); err != nil {
		outcome = "rollback"
		_ = tx.Rollback()
		return err
	}
	if err = inner.commit(); err != nil {
		outcome = "commit_error"
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (u *UnitOfWork) Publish(event domain.Event) error {
	return u.publisher.Publish(event)
}

func (tx *transaction) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	start := time.Now()
	_, _ = upsertPaymentIntent(context.Background(), tx.tx, intent)
	tx.store.recordPersistence("payment_intent", "save", nil, start)
	return intent
}

func (tx *transaction) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	start := time.Now()
	intent, err := getPaymentIntent(context.Background(), tx.tx, id)
	tx.store.recordPersistence("payment_intent", "get", err, start)
	return intent, err
}

func (tx *transaction) ListPaymentIntents() []domain.PaymentIntent {
	items, err := listPaymentIntents(context.Background(), tx.tx)
	if err != nil {
		return nil
	}
	return items
}

func (tx *transaction) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	start := time.Now()
	_, _ = upsertPaymentAttempt(context.Background(), tx.tx, attempt)
	tx.store.recordPersistence("payment_attempt", "save", nil, start)
	return attempt
}

func (tx *transaction) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	start := time.Now()
	attempt, err := getPaymentAttempt(context.Background(), tx.tx, id)
	tx.store.recordPersistence("payment_attempt", "get", err, start)
	return attempt, err
}

func (tx *transaction) ListPaymentAttempts() []domain.PaymentAttempt {
	items, err := listPaymentAttempts(context.Background(), tx.tx)
	if err != nil {
		return nil
	}
	return items
}

func (tx *transaction) SaveCharge(charge domain.Charge) domain.Charge {
	start := time.Now()
	_, _ = upsertCharge(context.Background(), tx.tx, charge)
	tx.store.recordPersistence("charge", "save", nil, start)
	return charge
}

func (tx *transaction) GetCharge(id string) (domain.Charge, error) {
	start := time.Now()
	charge, err := getCharge(context.Background(), tx.tx, id)
	tx.store.recordPersistence("charge", "get", err, start)
	return charge, err
}

func (tx *transaction) ListCharges() []domain.Charge {
	items, err := listCharges(context.Background(), tx.tx)
	if err != nil {
		return nil
	}
	return items
}

func (tx *transaction) SaveRefund(refund domain.Refund) domain.Refund {
	start := time.Now()
	_, _ = upsertRefund(context.Background(), tx.tx, refund)
	tx.store.recordPersistence("refund", "save", nil, start)
	return refund
}

func (tx *transaction) GetRefund(id string) (domain.Refund, error) {
	start := time.Now()
	refund, err := getRefund(context.Background(), tx.tx, id)
	tx.store.recordPersistence("refund", "get", err, start)
	return refund, err
}

func (tx *transaction) ListRefunds() []domain.Refund {
	items, err := listRefunds(context.Background(), tx.tx)
	if err != nil {
		return nil
	}
	return items
}
