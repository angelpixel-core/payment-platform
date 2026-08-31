package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeMetricsRecorder struct{ called int }

func (f *fakeMetricsRecorder) RecordHTTPRequest(context.Context, string, string, int, time.Duration) {
	f.called++
}

func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	recorder := &fakeMetricsRecorder{}
	h := Metrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), recorder)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.called != 1 {
		t.Fatalf("expected metrics recorder to be called once, got %d", recorder.called)
	}
}
