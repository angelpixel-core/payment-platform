package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeMetricsRecorder struct {
	called int
	method string
	route  string
	status int
	duration time.Duration
}

func (f *fakeMetricsRecorder) RecordHTTPRequest(_ context.Context, method, route string, status int, duration time.Duration) {
	f.called++
	f.method = method
	f.route = route
	f.status = status
	f.duration = duration
}

func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	recorder := &fakeMetricsRecorder{}
	h := Metrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), recorder)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Pattern = "GET /health"
	h.ServeHTTP(rec, req)

	if recorder.called != 1 {
		t.Fatalf("expected metrics recorder to be called once, got %d", recorder.called)
	}
	if recorder.method != http.MethodGet {
		t.Fatalf("expected method %q, got %q", http.MethodGet, recorder.method)
	}
	if recorder.route != "GET /health" {
		t.Fatalf("expected route %q, got %q", "GET /health", recorder.route)
	}
	if recorder.status != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.status)
	}
	if recorder.duration <= 0 {
		t.Fatalf("expected positive duration, got %s", recorder.duration)
	}
}
