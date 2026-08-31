package middleware

import (
	"encoding/json"
	"net/http"

	"payment-sandbox/internal/domain"
)

func toError(v any) error {
	if err, ok := v.(error); ok {
		return err
	}
	return domain.NewError(http.StatusInternalServerError, "panic", "panic")
}

func writeErrorJSON(w http.ResponseWriter, status int, err *domain.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err})
}
