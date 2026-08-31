package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObservabilityMiddlewareSetsRequestID(t *testing.T) {
	h := Observability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got == "" {
			t.Fatal("expected request id in context")
		}
		w.WriteHeader(http.StatusCreated)
	}), slog.New(slog.NewTextHandler(io.Discard, nil)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Fatal("expected request id header")
	}
}

func TestObservabilityMiddlewareRecovers(t *testing.T) {
	h := Observability(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), slog.New(slog.NewTextHandler(io.Discard, nil)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
