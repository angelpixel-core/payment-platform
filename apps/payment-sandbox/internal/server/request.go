package server

import (
	"net/http"
	"strings"
)

func requestIdempotencyKey(r *http.Request, fallback string) string {
	if key := strings.TrimSpace(r.Header.Get("Idempotency-Key")); key != "" {
		return key
	}
	return strings.TrimSpace(fallback)
}
