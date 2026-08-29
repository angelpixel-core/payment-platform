package server

import (
	"errors"
	"net/http"

	"payment-sandbox/internal/sandbox"
)

func writeError(w http.ResponseWriter, err error) {
	var se *sandbox.Error
	if errors.As(err, &se) {
		writeJSON(w, se.StatusCode, map[string]any{"error": se})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]string{"code": "internal_error", "message": err.Error()}})
}
