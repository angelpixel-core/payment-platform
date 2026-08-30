package application

import (
	"crypto/sha256"
	"encoding/hex"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/modules/payments"
	"payment-sandbox/internal/modules/refunds"
	"payment-sandbox/internal/ports"
)

type PaymentService struct {
	payments *payments.PaymentService
	refunds   *refunds.Service
}

func NewPaymentService(store ports.Store, scenarioEngine ports.ScenarioResolver, events ports.EventPublisher) *PaymentService {
	clock := systemClock{}
	return &PaymentService{
		payments: payments.NewService(store, clock, scenarioEngine, events),
		refunds:   refunds.NewService(store, clock, events),
	}
}

func (s *PaymentService) CreatePaymentIntent(req domain.CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (domain.PaymentIntent, error) {
	return s.payments.CreatePaymentIntent(req, idempotencyKey, fingerprint)
}

func (s *PaymentService) ConfirmPaymentIntent(intentID string, req domain.ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (domain.ConfirmPaymentIntentResponse, error) {
	return s.payments.ConfirmPaymentIntent(intentID, req, scenarioHeader, idempotencyKey, fingerprint)
}

func (s *PaymentService) FinalizeProcessingPaymentIntent(intentID string) (domain.PaymentIntent, error) {
	return s.payments.FinalizeProcessingPaymentIntent(intentID)
}

func (s *PaymentService) CapturePaymentIntent(intentID string, req domain.CapturePaymentIntentRequest) (domain.CapturePaymentIntentResponse, error) {
	return s.payments.CapturePaymentIntent(intentID, req)
}

func (s *PaymentService) CreateRefund(req domain.RefundRequest, idempotencyKey, fingerprint string) (domain.RefundResponse, error) {
	return s.refunds.CreateRefund(req, idempotencyKey, fingerprint)
}

func Fingerprint(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func FingerprintString(value string) string { return Fingerprint([]byte(value)) }
