package sandbox

import (
	"context"
	"testing"
	"time"
)

type recordedFlow struct {
	flow     string
	outcome  string
	duration time.Duration
}

type recordedCommand struct {
	command  string
	outcome  string
	duration time.Duration
}

type fakeFlowMetricsRecorder struct {
	flowCalls    []recordedFlow
	commandCalls []recordedCommand
}

func (f *fakeFlowMetricsRecorder) RecordPaymentFlow(_ context.Context, flow, outcome string, duration time.Duration) {
	f.flowCalls = append(f.flowCalls, recordedFlow{flow: flow, outcome: outcome, duration: duration})
}

func (f *fakeFlowMetricsRecorder) RecordHTTPRequest(context.Context, string, string, int, time.Duration) {
}

func (f *fakeFlowMetricsRecorder) RecordPaymentCommand(_ context.Context, command, outcome string, duration time.Duration) {
	f.commandCalls = append(f.commandCalls, recordedCommand{command: command, outcome: outcome, duration: duration})
}

func (f *fakeFlowMetricsRecorder) RecordPersistenceOperation(context.Context, string, string, string, string, time.Duration) {
}
func (f *fakeFlowMetricsRecorder) RecordUnitOfWork(context.Context, string, string, time.Duration) {}
func (f *fakeFlowMetricsRecorder) RecordOutboxOperation(context.Context, string, string, string, time.Duration) {
}
func (f *fakeFlowMetricsRecorder) RecordOutboxPending(context.Context, string, int64) {}

func TestServiceRecordsPaymentFlowMetrics(t *testing.T) {
	recorder := &fakeFlowMetricsRecorder{}
	svc := NewServiceWithMetrics(recorder)

	created, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd"}, "create-1", fingerprintString("create-1"))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	confirmed, err := svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{PaymentMethodToken: "pm_card_processing"}, "", "confirm-1", fingerprintString("confirm-1|processing"))
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if confirmed.PaymentIntent.Status != PaymentIntentProcessing {
		t.Fatalf("expected processing, got %s", confirmed.PaymentIntent.Status)
	}

	finalized, err := svc.FinalizeProcessingPaymentIntent(created.ID)
	if err != nil {
		t.Fatalf("finalize failed: %v", err)
	}
	if finalized.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture, got %s", finalized.Status)
	}

	created, err = svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, "create-2", fingerprintString("create-2"))
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}

	confirmed, err = svc.ConfirmPaymentIntent(created.ID, ConfirmPaymentIntentRequest{}, "", "confirm-2", fingerprintString("confirm-2"))
	if err != nil {
		t.Fatalf("second confirm failed: %v", err)
	}
	if confirmed.PaymentIntent.Status != PaymentIntentRequiresCapture {
		t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
	}

	captured, err := svc.CapturePaymentIntent(created.ID, CapturePaymentIntentRequest{}, "")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}

	refunded, err := svc.CreateRefund(RefundRequest{ChargeID: captured.Charge.ID}, "refund-1", fingerprintString("refund-1|refund"))
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if refunded.Refund.Amount != 100 {
		t.Fatalf("expected full refund, got %d", refunded.Refund.Amount)
	}

	if _, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 0, Currency: "usd"}, "create-error", fingerprintString("create-error")); err == nil {
		t.Fatal("expected create to fail")
	}

	gotFlow := recorder.flowCalls
	wantFlow := []recordedFlow{
		{flow: "payment_intent.create", outcome: "success"},
		{flow: "payment_intent.confirm", outcome: "success"},
		{flow: "payment_intent.finalize_processing", outcome: "success"},
		{flow: "payment_intent.create", outcome: "success"},
		{flow: "payment_intent.confirm", outcome: "success"},
		{flow: "payment_intent.capture", outcome: "success"},
		{flow: "refund.create", outcome: "success"},
		{flow: "payment_intent.create", outcome: "error"},
	}

	if len(gotFlow) != len(wantFlow) {
		t.Fatalf("expected %d flow metric calls, got %d: %#v", len(wantFlow), len(gotFlow), gotFlow)
	}
	for i := range wantFlow {
		if gotFlow[i].flow != wantFlow[i].flow || gotFlow[i].outcome != wantFlow[i].outcome {
			t.Fatalf("flow call %d: expected %#v, got %#v", i, wantFlow[i], gotFlow[i])
		}
		if gotFlow[i].duration <= 0 {
			t.Fatalf("flow call %d: expected positive duration, got %s", i, gotFlow[i].duration)
		}
	}

	gotCommand := recorder.commandCalls
	wantCommand := []recordedCommand{
		{command: "payment_intent.create", outcome: "success"},
		{command: "payment_intent.confirm", outcome: "success"},
		{command: "payment_intent.finalize_processing", outcome: "success"},
		{command: "payment_intent.create", outcome: "success"},
		{command: "payment_intent.confirm", outcome: "success"},
		{command: "payment_intent.capture", outcome: "success"},
		{command: "refund.create", outcome: "success"},
		{command: "payment_intent.create", outcome: "error"},
	}

	if len(gotCommand) != len(wantCommand) {
		t.Fatalf("expected %d command metric calls, got %d: %#v", len(wantCommand), len(gotCommand), gotCommand)
	}
	for i := range wantCommand {
		if gotCommand[i].command != wantCommand[i].command || gotCommand[i].outcome != wantCommand[i].outcome {
			t.Fatalf("command call %d: expected %#v, got %#v", i, wantCommand[i], gotCommand[i])
		}
		if gotCommand[i].duration <= 0 {
			t.Fatalf("command call %d: expected positive duration, got %s", i, gotCommand[i].duration)
		}
	}
}
