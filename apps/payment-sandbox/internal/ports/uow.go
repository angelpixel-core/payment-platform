package ports

import "payment-sandbox/internal/domain"

type Transaction interface {
	Store
	Publish(event domain.Event) error
}

type UnitOfWork interface {
	Do(fn func(tx Transaction) error) error
}
