package postgres

import (
	"context"
	"database/sql"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type UnitOfWork struct {
	db        *sql.DB
	publisher ports.EventPublisher
}

func NewUnitOfWork(db *sql.DB, publisher ports.EventPublisher) *UnitOfWork {
	return &UnitOfWork{db: db, publisher: publisher}
}

var _ ports.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Do(fn func(tx ports.Transaction) error) error {
	tx, err := u.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	inner := &transaction{store: &Store{db: u.db}, tx: tx, publisher: u.publisher}
	if err := fn(inner); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := inner.commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (u *UnitOfWork) Publish(event domain.Event) error {
	return u.publisher.Publish(event)
}
