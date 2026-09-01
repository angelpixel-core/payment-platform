package sandbox

import (
	"context"
	"database/sql"
	"time"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/adapters/messaging/outbox"
	"payment-sandbox/internal/adapters/observability/metrics"
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
	metrics  metrics.MetricsRecorder
}

func NewService() *Service {
	return NewServiceWithMetrics(nil)
}

func NewServiceWithMetrics(metricsRecorder metrics.MetricsRecorder) *Service {
	store := NewMemoryStore(metricsRecorder)
	dispatcher := inprocess.NewPublisher()
	eventRecorder := appobs.NewRecorder()
	appobs.RegisterInternalHandlers(dispatcher, eventRecorder)
	publisher := outbox.NewPublisher(dispatcher, metricsRecorder)
	uow := memory.NewUnitOfWork(store, publisher)
	clock := clockadapter.NewClock()
	return &Service{commands: commandpayments.NewService(uow, clock, NewScenarioEngine()), refunds: commandrefunds.NewService(uow, clock), queries: payments.NewPaymentQueryService(store), recorder: eventRecorder, metrics: metricsRecorder}
}

func NewPostgresService(db *sql.DB) *Service {
	return NewPostgresServiceWithMetrics(db, nil)
}

func NewPostgresServiceWithMetrics(db *sql.DB, metricsRecorder metrics.MetricsRecorder) *Service {
	store := postgres.NewStore(db, metricsRecorder)
	dispatcher := inprocess.NewPublisher()
	eventRecorder := appobs.NewRecorder()
	appobs.RegisterInternalHandlers(dispatcher, eventRecorder)
	uow := postgres.NewUnitOfWork(db, dispatcher, metricsRecorder)
	clock := clockadapter.NewClock()
	return &Service{commands: commandpayments.NewService(uow, clock, NewScenarioEngine()), refunds: commandrefunds.NewService(uow, clock), queries: payments.NewPaymentQueryService(store), recorder: eventRecorder, metrics: metricsRecorder}
}

func (s *Service) CreatePaymentIntent(req CreatePaymentIntentRequest, idempotencyKey, fingerprint string) (PaymentIntent, error) {
	start := time.Now()
	var err error
	defer s.recordCommand("payment_intent.create", start, &err)
	defer s.recordFlow("payment_intent.create", start, &err)
	result, err := s.commands.CreatePaymentIntent(req, idempotencyKey, fingerprint)
	return result, err
}

func (s *Service) ConfirmPaymentIntent(intentID string, req ConfirmPaymentIntentRequest, scenarioHeader, idempotencyKey, fingerprint string) (ConfirmPaymentIntentResponse, error) {
	start := time.Now()
	var err error
	defer s.recordCommand("payment_intent.confirm", start, &err)
	defer s.recordFlow("payment_intent.confirm", start, &err)
	result, err := s.commands.ConfirmPaymentIntent(intentID, req, scenarioHeader, idempotencyKey, fingerprint)
	return result, err
}

func (s *Service) FinalizeProcessingPaymentIntent(intentID string) (PaymentIntent, error) {
	start := time.Now()
	var err error
	defer s.recordCommand("payment_intent.finalize_processing", start, &err)
	defer s.recordFlow("payment_intent.finalize_processing", start, &err)
	result, err := s.commands.FinalizeProcessingPaymentIntent(intentID)
	return result, err
}

func (s *Service) CapturePaymentIntent(intentID string, req CapturePaymentIntentRequest) (CapturePaymentIntentResponse, error) {
	start := time.Now()
	var err error
	defer s.recordCommand("payment_intent.capture", start, &err)
	defer s.recordFlow("payment_intent.capture", start, &err)
	result, err := s.commands.CapturePaymentIntent(intentID, req)
	return result, err
}

func (s *Service) CreateRefund(req RefundRequest, idempotencyKey, fingerprint string) (RefundResponse, error) {
	start := time.Now()
	var err error
	defer s.recordCommand("refund.create", start, &err)
	defer s.recordFlow("refund.create", start, &err)
	result, err := s.refunds.CreateRefund(req, idempotencyKey, fingerprint)
	return result, err
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

func (s *Service) recordFlow(flow string, start time.Time, err *error) {
	if s.metrics == nil {
		return
	}
	outcome := "success"
	if err != nil && *err != nil {
		outcome = "error"
	}
	s.metrics.RecordPaymentFlow(context.Background(), flow, outcome, time.Since(start))
}

func (s *Service) recordCommand(command string, start time.Time, err *error) {
	if s.metrics == nil {
		return
	}
	outcome := "success"
	if err != nil && *err != nil {
		outcome = "error"
	}
	s.metrics.RecordPaymentCommand(context.Background(), command, outcome, time.Since(start))
}
