package sandbox

import "payment-sandbox/internal/application"

type Service struct {
	app *application.PaymentService
}

func NewService() *Service {
	return &Service{app: application.NewPaymentService(NewMemoryStore(), NewScenarioEngine())}
}

func (s *Service) CreatePaymentIntent(req CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (PaymentIntent, error) {
	return s.app.CreatePaymentIntent(req, idempotencyKey, fingerprint)
}

func (s *Service) ConfirmPaymentIntent(intentID string, req ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (ConfirmPaymentIntentResponse, error) {
	return s.app.ConfirmPaymentIntent(intentID, req, scenarioHeader, idempotencyKey, fingerprint)
}

func (s *Service) FinalizeProcessingPaymentIntent(intentID string) (PaymentIntent, error) {
	return s.app.FinalizeProcessingPaymentIntent(intentID)
}

func (s *Service) CapturePaymentIntent(intentID string, req CapturePaymentIntentRequest) (CapturePaymentIntentResponse, error) {
	return s.app.CapturePaymentIntent(intentID, req)
}

func (s *Service) CreateRefund(req RefundRequest, idempotencyKey, fingerprint string) (RefundResponse, error) {
	return s.app.CreateRefund(req, idempotencyKey, fingerprint)
}
