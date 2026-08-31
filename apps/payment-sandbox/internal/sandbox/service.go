package sandbox

import (
	"payment-sandbox/internal/adapters/eventing/inprocess"
	"payment-sandbox/internal/application"
)

type Service struct {
	commands *application.PaymentService
	queries  *application.PaymentQueryService
}

func NewService() *Service {
	store := NewMemoryStore()
	return &Service{commands: application.NewPaymentService(store, NewScenarioEngine(), inprocess.NewPublisher()), queries: application.NewPaymentQueryService(store)}
}

func (s *Service) CreatePaymentIntent(req CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (PaymentIntent, error) {
	return s.commands.CreatePaymentIntent(req, idempotencyKey, fingerprint)
}

func (s *Service) ConfirmPaymentIntent(intentID string, req ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (ConfirmPaymentIntentResponse, error) {
	return s.commands.ConfirmPaymentIntent(intentID, req, scenarioHeader, idempotencyKey, fingerprint)
}

func (s *Service) FinalizeProcessingPaymentIntent(intentID string) (PaymentIntent, error) {
	return s.commands.FinalizeProcessingPaymentIntent(intentID)
}

func (s *Service) CapturePaymentIntent(intentID string, req CapturePaymentIntentRequest) (CapturePaymentIntentResponse, error) {
	return s.commands.CapturePaymentIntent(intentID, req)
}

func (s *Service) CreateRefund(req RefundRequest, idempotencyKey, fingerprint string) (RefundResponse, error) {
	return s.commands.CreateRefund(req, idempotencyKey, fingerprint)
}

func (s *Service) GetPaymentIntent(id string) (PaymentIntent, error) {
	return s.queries.GetPaymentIntent(id)
}
func (s *Service) GetPaymentAttempt(id string) (PaymentAttempt, error) {
	return s.queries.GetPaymentAttempt(id)
}
func (s *Service) GetCharge(id string) (Charge, error) { return s.queries.GetCharge(id) }
func (s *Service) GetRefund(id string) (Refund, error) { return s.queries.GetRefund(id) }
