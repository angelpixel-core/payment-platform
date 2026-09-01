package metrics

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type customMetricRecorder interface {
	RecordCustomMetric(name string, value float64)
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
