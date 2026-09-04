package ports

import "payment-sandbox/internal/domain"

type Store interface {
	WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error)
	NextID(prefix string) string
	NextReference(prefix string) string

	SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent
	GetPaymentIntent(id string) (domain.PaymentIntent, error)
	ListPaymentIntents() []domain.PaymentIntent
	SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt
	GetPaymentAttempt(id string) (domain.PaymentAttempt, error)
	ListPaymentAttempts() []domain.PaymentAttempt
	SaveCharge(charge domain.Charge) domain.Charge
	GetCharge(id string) (domain.Charge, error)
	ListCharges() []domain.Charge
	SaveRefund(refund domain.Refund) domain.Refund
	GetRefund(id string) (domain.Refund, error)
	ListRefunds() []domain.Refund
}
