package ports

import "payment-sandbox/internal/domain"

type Transaction interface {
	Store
	Publish(event domain.Event) error
}

// INFO: UnitOfWork coordinates atomic multi-repository writes.
// Use it for business operations that must commit together, such as
// persisting a payment and its ledger counterparty in the same transaction.
type UnitOfWork interface {
	Do(fn func(tx Transaction) error) error
}
