package ports

import (
	"time"

	"payment-sandbox/internal/domain"
)

type Clock interface {
	Now() time.Time
}

type Store interface {
	WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error)
	NextID(prefix string) string
	NextReference(prefix string) string

	SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent
	GetPaymentIntent(id string) (domain.PaymentIntent, error)
	SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt
	GetPaymentAttempt(id string) (domain.PaymentAttempt, error)
	SaveCharge(charge domain.Charge) domain.Charge
	GetCharge(id string) (domain.Charge, error)
	SaveRefund(refund domain.Refund) domain.Refund
	GetRefund(id string) (domain.Refund, error)
}
