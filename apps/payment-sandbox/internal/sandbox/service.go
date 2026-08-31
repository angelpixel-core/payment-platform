package sandbox

import (
	"payment-sandbox/internal/adapters/eventing/inprocess"
	"payment-sandbox/internal/adapters/eventing/outbox"
	"payment-sandbox/internal/adapters/memory"
	"payment-sandbox/internal/application"
	appEvents "payment-sandbox/internal/application/events"
	"payment-sandbox/internal/application/queries"
)

type Service struct {
	commands *application.PaymentService
	queries  *queries.PaymentQueryService
	recorder *appEvents.Recorder
}

func NewService() *Service {
	store := NewMemoryStore()
	dispatcher := inprocess.NewPublisher()
	recorder := appEvents.NewRecorder()
	appEvents.RegisterInternalHandlers(dispatcher, recorder)
	publisher := outbox.NewPublisher(dispatcher)
	uow := memory.NewUnitOfWork(store, publisher)
	return &Service{commands: application.NewPaymentService(uow, NewScenarioEngine()), queries: queries.NewPaymentQueryService(store), recorder: recorder}
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

func (s *Service) GetPaymentIntent(id string) (PaymentIntentView, error) {
	return s.queries.GetPaymentIntent(id)
}
func (s *Service) GetPaymentAttempt(id string) (PaymentAttemptView, error) {
	return s.queries.GetPaymentAttempt(id)
}
func (s *Service) GetCharge(id string) (ChargeView, error) { return s.queries.GetCharge(id) }
func (s *Service) GetRefund(id string) (RefundView, error) { return s.queries.GetRefund(id) }

func (s *Service) EventRecorder() *appEvents.Recorder { return s.recorder }
