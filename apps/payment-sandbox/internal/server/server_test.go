package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-sandbox/internal/bootstrap"
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

type intentViewEnvelope struct {
	PaymentIntent sandbox.PaymentIntentView `json:"payment_intent"`
}

type attemptViewEnvelope struct {
	PaymentAttempt sandbox.PaymentAttemptView `json:"payment_attempt"`
}

type chargeViewEnvelope struct {
	Charge sandbox.ChargeView `json:"charge"`
}

type refundViewEnvelope struct {
	Refund sandbox.RefundView `json:"refund"`
}

type errorEnvelope struct {
	Error sandbox.Error `json:"error"`
}

func TestPaymentLifecycle(t *testing.T) {
	client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
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

func TestQueryEndpoints(t *testing.T) {
	client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
	defer client.Close()

	_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "query-create", nil)

	resp, body := doGet(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var gotIntent intentViewEnvelope
	if err := json.Unmarshal(body, &gotIntent); err != nil {
		t.Fatalf("decode intent failed: %v", err)
	}
	if gotIntent.PaymentIntent.ID != created.PaymentIntent.ID {
		t.Fatalf("expected intent %s, got %s", created.PaymentIntent.ID, gotIntent.PaymentIntent.ID)
	}

	_, confirmed := doPost[confirmEnvelope](t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{}, "query-confirm", nil)
	resp, body = doGet(t, client.URL+"/v1/payment_attempts/"+confirmed.PaymentAttempt.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var gotAttempt attemptViewEnvelope
	if err := json.Unmarshal(body, &gotAttempt); err != nil {
		t.Fatalf("decode attempt failed: %v", err)
	}
	if gotAttempt.PaymentAttempt.ID != confirmed.PaymentAttempt.ID {
		t.Fatalf("expected attempt %s, got %s", confirmed.PaymentAttempt.ID, gotAttempt.PaymentAttempt.ID)
	}

	resp, body = doGet(t, client.URL+"/v1/charges/"+confirmed.Charge.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var gotCharge chargeViewEnvelope
	if err := json.Unmarshal(body, &gotCharge); err != nil {
		t.Fatalf("decode charge failed: %v", err)
	}
	if gotCharge.Charge.ID != confirmed.Charge.ID {
		t.Fatalf("expected charge %s, got %s", confirmed.Charge.ID, gotCharge.Charge.ID)
	}

	_, captured := doPost[captureEnvelope](t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/capture", map[string]any{}, "query-capture", nil)
	_, refunded := doPost[refundEnvelope](t, client.URL+"/v1/refunds", map[string]any{"charge_id": captured.Charge.ID}, "query-refund", nil)
	resp, body = doGet(t, client.URL+"/v1/refunds/"+refunded.Refund.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var gotRefund refundViewEnvelope
	if err := json.Unmarshal(body, &gotRefund); err != nil {
		t.Fatalf("decode refund failed: %v", err)
	}
	if gotRefund.Refund.ID != refunded.Refund.ID {
		t.Fatalf("expected refund %s, got %s", refunded.Refund.ID, gotRefund.Refund.ID)
	}
}

func TestIdempotency(t *testing.T) {
	client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
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

func TestScenarioResponses(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		payload           map[string]any
		wantStatus        int
		wantIntentStatus  string
		wantAttemptStatus string
		wantCharge        bool
		wantErrorCode     string
		wantErrorMessage  string
	}{
		{
			name:              "header priority",
			headers:           map[string]string{"X-Sandbox-Scenario": "declined_insufficient_funds"},
			payload:           map[string]any{"payment_method_token": "pm_card_visa"},
			wantStatus:        http.StatusOK,
			wantIntentStatus:  "failed",
			wantAttemptStatus: "declined",
			wantCharge:        false,
		},
		{
			name:              "token fallback",
			payload:           map[string]any{"payment_method_token": "pm_card_insufficient_funds"},
			wantStatus:        http.StatusOK,
			wantIntentStatus:  "failed",
			wantAttemptStatus: "declined",
			wantCharge:        false,
		},
		{
			name:             "unknown header",
			headers:          map[string]string{"X-Sandbox-Scenario": "unknown_scenario"},
			payload:          map[string]any{"payment_method_token": "pm_card_visa"},
			wantStatus:       http.StatusUnprocessableEntity,
			wantErrorCode:    "invalid_scenario",
			wantErrorMessage: "unknown sandbox scenario \"unknown_scenario\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
			defer client.Close()

			_, created := doPost[createEnvelope](t, client.URL+"/v1/payment_intents", map[string]any{
				"amount":   100,
				"currency": "usd",
			}, "scenario-create-"+tt.name, nil)

			resp, body := doPostRawWithHeaders(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", tt.payload, "scenario-confirm-"+tt.name, tt.headers)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			if tt.wantStatus != http.StatusOK {
				got := decodeError(t, body)
				if got.Error.Code != tt.wantErrorCode || got.Error.Message != tt.wantErrorMessage {
					t.Fatalf("unexpected error: %+v", got)
				}
				return
			}

			var confirmed confirmEnvelope
			if err := json.Unmarshal(body, &confirmed); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if string(confirmed.PaymentIntent.Status) != tt.wantIntentStatus {
				t.Fatalf("expected intent status %s, got %s", tt.wantIntentStatus, confirmed.PaymentIntent.Status)
			}
			if string(confirmed.PaymentAttempt.Status) != tt.wantAttemptStatus {
				t.Fatalf("expected attempt status %s, got %s", tt.wantAttemptStatus, confirmed.PaymentAttempt.Status)
			}
			if (confirmed.Charge != nil) != tt.wantCharge {
				t.Fatalf("expected charge presence %t, got %t", tt.wantCharge, confirmed.Charge != nil)
			}
		})
	}
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, clientURL string) (*http.Response, []byte)
		wantStatus       int
		wantErrorCode    string
		wantErrorMessage string
	}{
		{
			name: "invalid transition",
			setup: func(t *testing.T, clientURL string) (*http.Response, []byte) {
				_, created := doPost[createEnvelope](t, clientURL+"/v1/payment_intents", map[string]any{"amount": 100, "currency": "usd"}, "create-invalid", nil)
				return doRawPost(t, clientURL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/capture", map[string]any{})
			},
			wantStatus:       http.StatusConflict,
			wantErrorCode:    "invalid_intent_state",
			wantErrorMessage: "payment intent cannot be captured in its current state",
		},
		{
			name: "invalid amount",
			setup: func(t *testing.T, clientURL string) (*http.Response, []byte) {
				return doPostRawWithHeaders(t, clientURL+"/v1/payment_intents", map[string]any{"amount": 0, "currency": "usd"}, "invalid-amount", nil)
			},
			wantStatus:       http.StatusBadRequest,
			wantErrorCode:    "invalid_amount",
			wantErrorMessage: "amount must be greater than zero",
		},
		{
			name: "malformed json",
			setup: func(t *testing.T, clientURL string) (*http.Response, []byte) {
				req, err := http.NewRequest(http.MethodPost, clientURL+"/v1/payment_intents", bytes.NewBufferString("{"))
				if err != nil {
					t.Fatalf("request creation failed: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("request failed: %v", err)
				}
				defer resp.Body.Close()
				return resp, mustReadAll(t, resp)
			},
			wantStatus:       http.StatusInternalServerError,
			wantErrorCode:    "internal_error",
			wantErrorMessage: "unexpected end of JSON input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
			defer client.Close()

			resp, body := tt.setup(t, client.URL)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			got := decodeError(t, body)
			if got.Error.Code != tt.wantErrorCode || got.Error.Message != tt.wantErrorMessage {
				t.Fatalf("unexpected error: %+v", got)
			}
		})
	}
}

func TestHTTPContractShapes(t *testing.T) {
	client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
	defer client.Close()

	createResp, createBody := doPostRawWithHeaders(t, client.URL+"/v1/payment_intents", map[string]any{
		"amount":         100,
		"currency":       "usd",
		"capture_method": "manual",
	}, "shape-create", nil)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.StatusCode)
	}
	assertJSONHasKeys(t, createBody, "payment_intent")
	assertNestedJSONHasKeys(t, createBody, "payment_intent", "id", "amount", "currency", "capture_method", "status")

	var created createEnvelope
	if err := json.Unmarshal(createBody, &created); err != nil {
		t.Fatalf("decode create failed: %v", err)
	}

	confirmResp, confirmBody := doPostRawWithHeaders(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{}, "shape-confirm", nil)
	if confirmResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", confirmResp.StatusCode)
	}
	assertJSONHasKeys(t, confirmBody, "payment_intent", "payment_attempt", "charge")
	assertNestedJSONHasKeys(t, confirmBody, "payment_intent", "id", "status", "latest_attempt_id")
	assertNestedJSONHasKeys(t, confirmBody, "payment_attempt", "id", "payment_intent_id", "status")
	assertNestedJSONHasKeys(t, confirmBody, "charge", "id", "payment_intent_id", "status")

	var confirmed confirmEnvelope
	if err := json.Unmarshal(confirmBody, &confirmed); err != nil {
		t.Fatalf("decode confirm failed: %v", err)
	}

	captureResp, captureBody := doPostRawWithHeaders(t, client.URL+"/v1/payment_intents/"+created.PaymentIntent.ID+"/capture", map[string]any{}, "shape-capture", nil)
	if captureResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", captureResp.StatusCode)
	}
	assertJSONHasKeys(t, captureBody, "payment_intent", "charge")
	assertNestedJSONHasKeys(t, captureBody, "payment_intent", "id", "status", "charge_id")
	assertNestedJSONHasKeys(t, captureBody, "charge", "id", "status", "captured_amount")

	var captured captureEnvelope
	if err := json.Unmarshal(captureBody, &captured); err != nil {
		t.Fatalf("decode capture failed: %v", err)
	}

	refundResp, refundBody := doPostRawWithHeaders(t, client.URL+"/v1/refunds", map[string]any{"charge_id": captured.Charge.ID}, "shape-refund", nil)
	if refundResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", refundResp.StatusCode)
	}
	assertJSONHasKeys(t, refundBody, "refund", "charge")
	assertNestedJSONHasKeys(t, refundBody, "refund", "id", "charge_id", "status")
	assertNestedJSONHasKeys(t, refundBody, "charge", "id", "status", "refunded_amount")
}

func TestHealth(t *testing.T) {
	client := httptest.NewServer(bootstrap.New(sandbox.NewService()))
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

func doGet(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	return resp, buf.Bytes()
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

func decodeError(t *testing.T, body []byte) errorEnvelope {
	t.Helper()
	var out errorEnvelope
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode error body failed: %v", err)
	}
	return out
}

func mustReadAll(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	return buf.Bytes()
}

func assertJSONHasKeys(t *testing.T, body []byte, keys ...string) {
	t.Helper()
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode json failed: %v", err)
	}
	for _, key := range keys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload %s", key, string(body))
		}
	}
}

func assertNestedJSONHasKeys(t *testing.T, body []byte, key string, nestedKeys ...string) {
	t.Helper()
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode json failed: %v", err)
	}
	raw, ok := payload[key]
	if !ok {
		t.Fatalf("expected key %q in payload %s", key, string(body))
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(raw, &nested); err != nil {
		t.Fatalf("decode nested json for %q failed: %v", key, err)
	}
	for _, nestedKey := range nestedKeys {
		if _, ok := nested[nestedKey]; !ok {
			t.Fatalf("expected nested key %q in %q payload %s", nestedKey, key, string(body))
		}
	}
}
