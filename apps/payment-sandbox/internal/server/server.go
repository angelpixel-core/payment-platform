package server

import (
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
	scenarioHeader := strings.TrimSpace(r.Header.Get("X-Sandbox-Scenario"))
	idem := requestIdempotencyKey(r, req.IdempotencyKey)
	fingerprint := sandbox.FingerprintString(string(payload) + "|scenario=" + scenarioHeader + "|token=" + req.PaymentMethodToken)
	result, err := s.svc.ConfirmPaymentIntent(intentID, req, scenarioHeader, idem, fingerprint)
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
