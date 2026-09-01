package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMetricsMiddleware(b *testing.B) {
	recorder := &fakeMetricsRecorder{}
	h := Metrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), recorder)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Pattern = "GET /health"
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			b.Fatal("expected request id in context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-Request-Id", "bench-request-id")
		req, _ = withRequestID(req)
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkObservabilityMiddleware(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Observability(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Pattern = "GET /health"
		req.Header.Set("X-Request-Id", "bench-request-id")
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkRecoveryPath(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Observability(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		req.Pattern = "GET /panic"
		req.Header.Set("X-Request-Id", "bench-request-id")
		h.ServeHTTP(rec, req)
	}
}
