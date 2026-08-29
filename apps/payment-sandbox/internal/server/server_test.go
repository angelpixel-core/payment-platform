package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-sandbox/internal/sandbox"
)

type createEnvelope struct {
	PaymentIntent sandbox.PaymentIntent `json:"payment_intent"`
}

type confirmEnvelope struct {
	PaymentIntent  sandbox.PaymentIntent  `json:"payment_intent"`
	PaymentAttempt sandbox.PaymentAttempt `json:"payment_attempt"`
	Charge         *sandbox.Charge        `json:"charge"`
}

type captureEnvelope struct {
	PaymentIntent sandbox.PaymentIntent `json:"payment_intent"`
	Charge        sandbox.Charge        `json:"charge"`
}

type refundEnvelope struct {
	Refund sandbox.Refund `json:"refund"`
	Charge sandbox.Charge `json:"charge"`
}

func TestPaymentLifecycle(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":         100,
		"currency":       "usd",
		"capture_method": "manual",
	}, "create-1", nil)
	intentID := created.PaymentIntent.ID

	_, confirmed := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+intentID+"/confirm", map[string]any{}, "confirm-1", nil)
	if confirmed.PaymentIntent.Status != "requires_capture" {
		t.Fatalf("expected requires_capture, got %s", confirmed.PaymentIntent.Status)
	}
	if confirmed.Charge == nil {
		t.Fatal("expected charge in confirm response")
	}
	chargeID := confirmed.Charge.ID

	_, captured := doPost[captureEnvelope](t, client.URL+"/v1/payment_intents/"+intentID+"/capture", map[string]any{}, "", nil)
	if captured.PaymentIntent.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %s", captured.PaymentIntent.Status)
	}

	_, refunded := doPost[refundEnvelope](t, client.URL+"/v1/refunds", map[string]any{"charge_id": chargeID}, "refund-1", nil)
	if refunded.Refund.Amount != 100 {
		t.Fatalf("expected full refund, got %d", refunded.Refund.Amount)
	}
	if refunded.Charge.Status != "refunded" {
		t.Fatalf("expected refunded charge status, got %s", refunded.Charge.Status)
	}
}

func TestIdempotency(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, first := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "idem-create", nil)
	_, second := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "idem-create", nil)
	if first.PaymentIntent.ID != second.PaymentIntent.ID {
		t.Fatalf("expected same id on idempotent create")
	}

	intentID := first.PaymentIntent.ID
	_, confirm1 := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+intentID+"/confirm", map[string]any{}, "idem-confirm", nil)
	_, confirm2 := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+intentID+"/confirm", map[string]any{}, "idem-confirm", nil)
	if confirm1.PaymentAttempt.ID != confirm2.PaymentAttempt.ID {
		t.Fatalf("expected same attempt on idempotent confirm")
	}
}

func TestInvalidTransition(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "create-invalid", nil)

	resp, body := doRawPost(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/capture", map[string]any{})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if !bytes.Contains(body, []byte("invalid_intent_state")) {
		t.Fatalf("expected invalid_intent_state error, got %s", string(body))
	}
}

func TestScenarioHeaderPriority(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "scenario-priority-create", nil)

	_, confirmed := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{
		"payment_method_token": "pm_card_visa",
	}, "scenario-priority-confirm", map[string]string{"X-Sandbox-Scenario": "declined_insufficient_funds"})

	if confirmed.PaymentIntent.Status != "failed" {
		t.Fatalf("expected failed from header scenario, got %s", confirmed.PaymentIntent.Status)
	}
	if confirmed.Charge != nil {
		t.Fatalf("expected no charge for declined scenario")
	}
	if confirmed.PaymentAttempt.Status != "declined" {
		t.Fatalf("expected declined attempt, got %s", confirmed.PaymentAttempt.Status)
	}
}

func TestScenarioTokenFallback(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "scenario-fallback-create", nil)

	_, confirmed := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{
		"payment_method_token": "pm_card_insufficient_funds",
	}, "scenario-fallback-confirm", nil)

	if confirmed.PaymentIntent.Status != "failed" {
		t.Fatalf("expected failed from token fallback, got %s", confirmed.PaymentIntent.Status)
	}
	if confirmed.Charge != nil {
		t.Fatalf("expected no charge for declined token scenario")
	}
}

func TestScenarioInvalid(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "scenario-invalid-create", nil)

	resp, body := doPostRawWithHeaders(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{
		"payment_method_token": "pm_card_visa",
	}, "scenario-invalid-confirm", map[string]string{"X-Sandbox-Scenario": "unknown_scenario"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	if !bytes.Contains(body, []byte("invalid_scenario")) {
		t.Fatalf("expected invalid_scenario error, got %s", string(body))
	}
}

func TestHealth(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	resp, err := http.Get(client.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMalformedJSONReturnsInternalError(t *testing.T) {
	client := httptest.NewServer(New(sandbox.NewService()))
	defer client.Close()

	req, err := http.NewRequest(http.MethodPost, client.URL+"/v1/payment_intents", bytes.NewBufferString("{"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func doPost[T any](t *testing.T, url string, payload any, idem string, headers map[string]string) (int, T) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if idem != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	return resp.StatusCode, out
}

func doRawPost(t *testing.T, url string, payload any) (*http.Response, []byte) {
	return doPostRawWithHeaders(t, url, payload, "", nil)
}

func doPostRawWithHeaders(t *testing.T, url string, payload any, idem string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if idem != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	return resp, buf.Bytes()
}
