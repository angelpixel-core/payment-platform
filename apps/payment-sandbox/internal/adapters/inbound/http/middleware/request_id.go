package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func withRequestID(r *http.Request) (*http.Request, string) {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
	if requestID == "" {
		requestID = uuid.NewString()
	}
	ctx := context.WithValue(r.Context(), requestIDKey, requestID)
	return r.WithContext(ctx), requestID
}
