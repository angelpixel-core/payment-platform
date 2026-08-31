package sandbox

import (
	"database/sql"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/adapters/messaging/outbox"
	"payment-sandbox/internal/adapters/persistence/memory"
	"payment-sandbox/internal/adapters/persistence/postgres"
	clockadapter "payment-sandbox/internal/adapters/time/system"
	commandpayments "payment-sandbox/internal/application/commands/payments"
	commandrefunds "payment-sandbox/internal/application/commands/refunds"
	"payment-sandbox/internal/application/queries/payments"
	appobs "payment-sandbox/internal/application/support/observability"
)

type Service struct {
	commands *commandpayments.PaymentService
	refunds  *commandrefunds.Service
	queries  *payments.PaymentQueryService
	recorder *appobs.Recorder
}

func NewService() *Service {
	store := NewMemoryStore()
	dispatcher := inprocess.NewPublisher()
	recorder := appobs.NewRecorder()
	appobs.RegisterInternalHandlers(dispatcher, recorder)
	publisher := outbox.NewPublisher(dispatcher)
	uow := memory.NewUnitOfWork(store, publisher)
	clock := clockadapter.NewClock()
	return &Service{commands: commandpayments.NewService(uow, clock, NewScenarioEngine()), refunds: commandrefunds.NewService(uow, clock), queries: payments.NewPaymentQueryService(store), recorder: recorder}
}

func NewPostgresService(db *sql.DB) *Service {
	store := postgres.NewStore(db)
	dispatcher := inprocess.NewPublisher()
	recorder := appobs.NewRecorder()
	appobs.RegisterInternalHandlers(dispatcher, recorder)
	uow := postgres.NewUnitOfWork(db, dispatcher)
	clock := clockadapter.NewClock()
	return &Service{commands: commandpayments.NewService(uow, clock, NewScenarioEngine()), refunds: commandrefunds.NewService(uow, clock), queries: payments.NewPaymentQueryService(store), recorder: recorder}
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
	return s.refunds.CreateRefund(req, idempotencyKey, fingerprint)
}

func (s *Service) GetPaymentIntent(id string) (PaymentIntentView, error) {
	return s.queries.GetPaymentIntent(id)
}
func (s *Service) GetPaymentAttempt(id string) (PaymentAttemptView, error) {
	return s.queries.GetPaymentAttempt(id)
}
func (s *Service) GetCharge(id string) (ChargeView, error) { return s.queries.GetCharge(id) }
func (s *Service) GetRefund(id string) (RefundView, error) { return s.queries.GetRefund(id) }

func (s *Service) EventRecorder() *appobs.Recorder { return s.recorder }
