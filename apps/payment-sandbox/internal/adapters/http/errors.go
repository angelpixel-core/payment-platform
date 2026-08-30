package httpadapter

import (
	"errors"
	"net/http"

	"payment-sandbox/internal/domain"
)

func writeError(w http.ResponseWriter, err error) {
	var se *domain.Error
	if errors.As(err, &se) {
		writeJSON(w, se.StatusCode, map[string]any{"error": se})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": domain.NewError(http.StatusInternalServerError, "internal_error", err.Error())})
}
