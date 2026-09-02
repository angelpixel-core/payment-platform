package httpadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

func RequestIdempotencyKey(r *http.Request, fallback string) string {
	if key := strings.TrimSpace(r.Header.Get("Idempotency-Key")); key != "" {
		return key
	}
	return strings.TrimSpace(fallback)
}

func fingerprint(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func fingerprintString(value string) string {
	return fingerprint([]byte(value))
}
