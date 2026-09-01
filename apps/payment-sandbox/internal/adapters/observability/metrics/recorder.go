package metrics

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type customMetricRecorder interface {
	RecordCustomMetric(name string, value float64)
}

type MetricsRecorder interface {
	RecordHTTPRequest(ctx context.Context, method, route string, status int, duration time.Duration)
	RecordPaymentFlow(ctx context.Context, flow, outcome string, duration time.Duration)
	RecordPaymentCommand(ctx context.Context, command, outcome string, duration time.Duration)
	RecordPersistenceOperation(ctx context.Context, backend, resource, operation, outcome string, duration time.Duration)
	RecordUnitOfWork(ctx context.Context, backend, outcome string, duration time.Duration)
	RecordOutboxOperation(ctx context.Context, backend, operation, outcome string, duration time.Duration)
	RecordOutboxPending(ctx context.Context, backend string, pending int64)
}

type Recorder struct {
	customMetrics    customMetricRecorder
	httpRequests     metric.Int64Counter
	httpDurations    metric.Int64Histogram
	paymentFlows     metric.Int64Counter
	paymentDurations metric.Int64Histogram
	paymentErrors    metric.Int64Counter
	paymentCommands  metric.Int64Counter
	commandDurations metric.Int64Histogram
	commandErrors    metric.Int64Counter
	persistenceOps   metric.Int64Counter
	persistenceDur   metric.Int64Histogram
	persistenceErr   metric.Int64Counter
	uowOps           metric.Int64Counter
	uowDur           metric.Int64Histogram
	uowErr           metric.Int64Counter
	outboxOps        metric.Int64Counter
	outboxDur        metric.Int64Histogram
	outboxErr        metric.Int64Counter
	outboxPending    metric.Int64UpDownCounter
	outboxPendingVal map[string]int64
	outboxMu         sync.Mutex
}

func NewRecorder(customMetrics customMetricRecorder) (*Recorder, error) {
	meter := otel.Meter("payment-sandbox/metrics")

	httpRequests, err := meter.Int64Counter(
		"payment_sandbox_http_requests_total",
		metric.WithDescription("Total HTTP requests received"),
	)
	if err != nil {
		return nil, err
	}
	httpDurations, err := meter.Int64Histogram(
		"payment_sandbox_http_request_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("HTTP request duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	paymentFlows, err := meter.Int64Counter(
		"payment_sandbox_payment_flows_total",
		metric.WithDescription("Total payment flow executions"),
	)
	if err != nil {
		return nil, err
	}
	paymentDurations, err := meter.Int64Histogram(
		"payment_sandbox_payment_flow_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Payment flow duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	paymentErrors, err := meter.Int64Counter(
		"payment_sandbox_payment_flow_errors_total",
		metric.WithDescription("Total failed payment flow executions"),
	)
	if err != nil {
		return nil, err
	}
	paymentCommands, err := meter.Int64Counter(
		"payment_sandbox_payment_commands_total",
		metric.WithDescription("Total payment command executions"),
	)
	if err != nil {
		return nil, err
	}
	commandDurations, err := meter.Int64Histogram(
		"payment_sandbox_payment_command_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Payment command duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	commandErrors, err := meter.Int64Counter(
		"payment_sandbox_payment_command_errors_total",
		metric.WithDescription("Total failed payment command executions"),
	)
	if err != nil {
		return nil, err
	}
	persistenceOps, err := meter.Int64Counter(
		"payment_sandbox_persistence_operations_total",
		metric.WithDescription("Total persistence operations"),
	)
	if err != nil {
		return nil, err
	}
	persistenceDur, err := meter.Int64Histogram(
		"payment_sandbox_persistence_operation_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Persistence operation duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	persistenceErr, err := meter.Int64Counter(
		"payment_sandbox_persistence_operation_errors_total",
		metric.WithDescription("Total failed persistence operations"),
	)
	if err != nil {
		return nil, err
	}
	uowOps, err := meter.Int64Counter(
		"payment_sandbox_unit_of_work_total",
		metric.WithDescription("Total unit of work executions"),
	)
	if err != nil {
		return nil, err
	}
	uowDur, err := meter.Int64Histogram(
		"payment_sandbox_unit_of_work_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Unit of work duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	uowErr, err := meter.Int64Counter(
		"payment_sandbox_unit_of_work_errors_total",
		metric.WithDescription("Total failed unit of work executions"),
	)
	if err != nil {
		return nil, err
	}
	outboxOps, err := meter.Int64Counter(
		"payment_sandbox_outbox_operations_total",
		metric.WithDescription("Total outbox operations"),
	)
	if err != nil {
		return nil, err
	}
	outboxDur, err := meter.Int64Histogram(
		"payment_sandbox_outbox_operation_duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Outbox operation duration in milliseconds"),
	)
	if err != nil {
		return nil, err
	}
	outboxErr, err := meter.Int64Counter(
		"payment_sandbox_outbox_operation_errors_total",
		metric.WithDescription("Total failed outbox operations"),
	)
	if err != nil {
		return nil, err
	}
	outboxPending, err := meter.Int64UpDownCounter(
		"payment_sandbox_outbox_pending_events",
		metric.WithDescription("Current pending outbox events"),
	)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		customMetrics:    customMetrics,
		httpRequests:     httpRequests,
		httpDurations:    httpDurations,
		paymentFlows:     paymentFlows,
		paymentDurations: paymentDurations,
		paymentErrors:    paymentErrors,
		paymentCommands:  paymentCommands,
		commandDurations: commandDurations,
		commandErrors:    commandErrors,
		persistenceOps:   persistenceOps,
		persistenceDur:   persistenceDur,
		persistenceErr:   persistenceErr,
		uowOps:           uowOps,
		uowDur:           uowDur,
		uowErr:           uowErr,
		outboxOps:        outboxOps,
		outboxDur:        outboxDur,
		outboxErr:        outboxErr,
		outboxPending:    outboxPending,
		outboxPendingVal: make(map[string]int64),
	}, nil
}

func (r *Recorder) RecordHTTPRequest(ctx context.Context, method, route string, status int, duration time.Duration) {
	if r == nil {
		return
	}
	r.httpRequests.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", route),
			attribute.String("http.status_class", statusClass(status)),
		),
	)
	r.httpDurations.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", status),
		),
	)
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("HTTP/Requests", 1)
		r.customMetrics.RecordCustomMetric("HTTP/RequestDurationMs", float64(duration.Milliseconds()))
	}
}

func (r *Recorder) RecordPaymentFlow(ctx context.Context, flow, outcome string, duration time.Duration) {
	if r == nil {
		return
	}
	flow = sanitizeMetricPart(flow)
	outcome = sanitizeMetricPart(outcome)
	r.paymentFlows.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("flow.name", flow),
			attribute.String("flow.outcome", outcome),
		),
	)
	r.paymentDurations.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("flow.name", flow),
			attribute.String("flow.outcome", outcome),
		),
	)
	if strings.EqualFold(outcome, "error") {
		r.paymentErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("flow.name", flow)))
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("Payments/"+flow+"/Count", 1)
		r.customMetrics.RecordCustomMetric("Payments/"+flow+"/DurationMs", float64(duration.Milliseconds()))
		if strings.EqualFold(outcome, "error") {
			r.customMetrics.RecordCustomMetric("Payments/"+flow+"/Errors", 1)
		}
	}
}

func (r *Recorder) RecordPaymentCommand(ctx context.Context, command, outcome string, duration time.Duration) {
	if r == nil {
		return
	}
	command = sanitizeMetricPart(command)
	outcome = sanitizeMetricPart(outcome)
	r.paymentCommands.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("command.name", command),
			attribute.String("command.outcome", outcome),
		),
	)
	r.commandDurations.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("command.name", command),
			attribute.String("command.outcome", outcome),
		),
	)
	if strings.EqualFold(outcome, "error") {
		r.commandErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("command.name", command)))
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("Commands/"+command+"/Count", 1)
		r.customMetrics.RecordCustomMetric("Commands/"+command+"/DurationMs", float64(duration.Milliseconds()))
		if strings.EqualFold(outcome, "error") {
			r.customMetrics.RecordCustomMetric("Commands/"+command+"/Errors", 1)
		}
	}
}

func (r *Recorder) RecordPersistenceOperation(ctx context.Context, backend, resource, operation, outcome string, duration time.Duration) {
	if r == nil {
		return
	}
	backend = sanitizeMetricPart(backend)
	resource = sanitizeMetricPart(resource)
	operation = sanitizeMetricPart(operation)
	outcome = sanitizeMetricPart(outcome)
	r.persistenceOps.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("persistence.backend", backend),
			attribute.String("persistence.resource", resource),
			attribute.String("persistence.operation", operation),
			attribute.String("persistence.outcome", outcome),
		),
	)
	r.persistenceDur.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("persistence.backend", backend),
			attribute.String("persistence.resource", resource),
			attribute.String("persistence.operation", operation),
			attribute.String("persistence.outcome", outcome),
		),
	)
	if strings.EqualFold(outcome, "error") {
		r.persistenceErr.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("persistence.backend", backend),
				attribute.String("persistence.resource", resource),
				attribute.String("persistence.operation", operation),
			),
		)
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("Persistence/"+backend+"/"+resource+"/"+operation+"/Count", 1)
		r.customMetrics.RecordCustomMetric("Persistence/"+backend+"/"+resource+"/"+operation+"/DurationMs", float64(duration.Milliseconds()))
		if strings.EqualFold(outcome, "error") {
			r.customMetrics.RecordCustomMetric("Persistence/"+backend+"/"+resource+"/"+operation+"/Errors", 1)
		}
	}
}

func (r *Recorder) RecordUnitOfWork(ctx context.Context, backend, outcome string, duration time.Duration) {
	if r == nil {
		return
	}
	backend = sanitizeMetricPart(backend)
	outcome = sanitizeMetricPart(outcome)
	r.uowOps.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("uow.backend", backend),
			attribute.String("uow.outcome", outcome),
		),
	)
	r.uowDur.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("uow.backend", backend),
			attribute.String("uow.outcome", outcome),
		),
	)
	if strings.EqualFold(outcome, "rollback") || strings.EqualFold(outcome, "commit_error") {
		r.uowErr.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("uow.backend", backend),
				attribute.String("uow.outcome", outcome),
			),
		)
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("UnitOfWork/"+backend+"/Count", 1)
		r.customMetrics.RecordCustomMetric("UnitOfWork/"+backend+"/DurationMs", float64(duration.Milliseconds()))
		if strings.EqualFold(outcome, "rollback") || strings.EqualFold(outcome, "commit_error") {
			r.customMetrics.RecordCustomMetric("UnitOfWork/"+backend+"/Errors", 1)
		}
	}
}

func (r *Recorder) RecordOutboxOperation(ctx context.Context, backend, operation, outcome string, duration time.Duration) {
	if r == nil {
		return
	}
	backend = sanitizeMetricPart(backend)
	operation = sanitizeMetricPart(operation)
	outcome = sanitizeMetricPart(outcome)
	r.outboxOps.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("outbox.backend", backend),
			attribute.String("outbox.operation", operation),
			attribute.String("outbox.outcome", outcome),
		),
	)
	r.outboxDur.Record(ctx, duration.Milliseconds(),
		metric.WithAttributes(
			attribute.String("outbox.backend", backend),
			attribute.String("outbox.operation", operation),
			attribute.String("outbox.outcome", outcome),
		),
	)
	if strings.EqualFold(outcome, "failure") {
		r.outboxErr.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("outbox.backend", backend),
				attribute.String("outbox.operation", operation),
			),
		)
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("Outbox/"+backend+"/"+operation+"/Count", 1)
		r.customMetrics.RecordCustomMetric("Outbox/"+backend+"/"+operation+"/DurationMs", float64(duration.Milliseconds()))
		if strings.EqualFold(outcome, "failure") {
			r.customMetrics.RecordCustomMetric("Outbox/"+backend+"/"+operation+"/Errors", 1)
		}
	}
}

func (r *Recorder) RecordOutboxPending(ctx context.Context, backend string, pending int64) {
	if r == nil {
		return
	}
	backend = sanitizeMetricPart(backend)
	r.outboxMu.Lock()
	prev := r.outboxPendingVal[backend]
	r.outboxPendingVal[backend] = pending
	r.outboxMu.Unlock()
	delta := pending - prev
	if delta != 0 {
		r.outboxPending.Add(ctx, delta, metric.WithAttributes(attribute.String("outbox.backend", backend)))
	}
	if r.customMetrics != nil {
		r.customMetrics.RecordCustomMetric("Outbox/"+backend+"/Pending", float64(pending))
	}
}

func statusClass(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

func sanitizeMetricPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, ".", "_")
	return value
}
