package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"payment-sandbox/internal/sandbox"
)

type Server struct {
	svc *sandbox.Service
	mux *http.ServeMux
}

func New(svc *sandbox.Service) http.Handler {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /v1/payment_intents", s.handleCreatePaymentIntent)
	s.mux.HandleFunc("POST /v1/payment_intents/{id}/confirm", s.handleConfirmPaymentIntent)
	s.mux.HandleFunc("POST /v1/payment_intents/{id}/capture", s.handleCapturePaymentIntent)
	s.mux.HandleFunc("POST /v1/refunds", s.handleCreateRefund)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	var req sandbox.CreatePaymentIntentRequest
	payload, err := readJSON(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}
	idem := requestIdempotencyKey(r, req.IdempotencyKey)
	result, err := s.svc.CreatePaymentIntent(req, idem, sandbox.Fingerprint(payload))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sandbox.CreatePaymentIntentResponse{PaymentIntent: result})
}

func (s *Server) handleConfirmPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var req sandbox.ConfirmPaymentIntentRequest
	payload, err := readJSON(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}
	intentID := r.PathValue("id")
	idem := requestIdempotencyKey(r, req.IdempotencyKey)
	result, err := s.svc.ConfirmPaymentIntent(intentID, req, idem, sandbox.Fingerprint(payload))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCapturePaymentIntent(w http.ResponseWriter, r *http.Request) {
	var req sandbox.CapturePaymentIntentRequest
	if err := readJSONBody(r, &req); err != nil {
		writeError(w, err)
		return
	}
	intentID := r.PathValue("id")
	result, err := s.svc.CapturePaymentIntent(intentID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateRefund(w http.ResponseWriter, r *http.Request) {
	var req sandbox.RefundRequest
	payload, err := readJSON(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}
	idem := requestIdempotencyKey(r, req.IdempotencyKey)
	result, err := s.svc.CreateRefund(req, idem, sandbox.Fingerprint(payload))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func requestIdempotencyKey(r *http.Request, fallback string) string {
	if key := strings.TrimSpace(r.Header.Get("Idempotency-Key")); key != "" {
		return key
	}
	return strings.TrimSpace(fallback)
}

func readJSON(r *http.Request, dst any) ([]byte, error) {
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

func readJSONBody(r *http.Request, dst any) error {
	_, err := readJSON(r, dst)
	return err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	var se *sandbox.Error
	if errors.As(err, &se) {
		writeJSON(w, se.StatusCode, map[string]any{"error": se})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]string{"code": "internal_error", "message": err.Error()}})
}
