package metrics

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type recordedCustomMetric struct {
	name  string
	value float64
}

type fakeCustomMetricSink struct {
	calls []recordedCustomMetric
}

func (f *fakeCustomMetricSink) RecordCustomMetric(name string, value float64) {
	f.calls = append(f.calls, recordedCustomMetric{name: name, value: value})
}

func TestRecorderRecordsHTTPAndPaymentFlowMetrics(t *testing.T) {
	sink := &fakeCustomMetricSink{}
	recorder, err := NewRecorder(sink)
	if err != nil {
		t.Fatalf("new recorder failed: %v", err)
	}

	recorder.RecordHTTPRequest(context.Background(), http.MethodGet, "/v1/payments", 201, 15*time.Millisecond)
	recorder.RecordPaymentFlow(context.Background(), "Refund/Create", "Error", 23*time.Millisecond)
	recorder.RecordPaymentCommand(context.Background(), "Refund/Create", "Error", 31*time.Millisecond)

	want := []recordedCustomMetric{
		{name: "HTTP/Requests", value: 1},
		{name: "HTTP/RequestDurationMs", value: 15},
		{name: "Payments/refund_create/Count", value: 1},
		{name: "Payments/refund_create/DurationMs", value: 23},
		{name: "Payments/refund_create/Errors", value: 1},
		{name: "Commands/refund_create/Count", value: 1},
		{name: "Commands/refund_create/DurationMs", value: 31},
		{name: "Commands/refund_create/Errors", value: 1},
	}

	if len(sink.calls) != len(want) {
		t.Fatalf("expected %d custom metric calls, got %d: %#v", len(want), len(sink.calls), sink.calls)
	}
	for i := range want {
		if sink.calls[i] != want[i] {
			t.Fatalf("call %d: expected %#v, got %#v", i, want[i], sink.calls[i])
		}
	}
}

func TestRecorderWorksWithoutCustomMetricSink(t *testing.T) {
	recorder, err := NewRecorder(nil)
	if err != nil {
		t.Fatalf("new recorder failed: %v", err)
	}

	recorder.RecordHTTPRequest(context.Background(), http.MethodGet, "/health", 200, 10*time.Millisecond)
	recorder.RecordPaymentFlow(context.Background(), "payment_intent.create", "success", 12*time.Millisecond)
	recorder.RecordPaymentCommand(context.Background(), "payment_intent.create", "success", 14*time.Millisecond)
}
