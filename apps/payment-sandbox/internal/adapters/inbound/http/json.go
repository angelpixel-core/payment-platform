package httpadapter

import (
	"encoding/json"
	"io"
	"net/http"
)

func ReadJSON(r *http.Request, dst any) ([]byte, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		body = []byte("{}")
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return nil, err
	}
	return body, nil
}

func ReadJSONBody(r *http.Request, dst any) error {
	_, err := ReadJSON(r, dst)
	return err
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
