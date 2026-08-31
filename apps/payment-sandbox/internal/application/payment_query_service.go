package application

import (
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type PaymentQueryService struct {
	store ports.Store
}

func NewPaymentQueryService(store ports.Store) *PaymentQueryService {
	return &PaymentQueryService{store: store}
}

func (s *PaymentQueryService) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	return s.store.GetPaymentIntent(id)
}

func (s *PaymentQueryService) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	return s.store.GetPaymentAttempt(id)
}

func (s *PaymentQueryService) GetCharge(id string) (domain.Charge, error) {
	return s.store.GetCharge(id)
}

func (s *PaymentQueryService) GetRefund(id string) (domain.Refund, error) {
	return s.store.GetRefund(id)
}
