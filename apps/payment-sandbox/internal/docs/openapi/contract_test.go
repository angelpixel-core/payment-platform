package openapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"

	"payment-sandbox/internal/bootstrap"
	"payment-sandbox/internal/sandbox"
)

func TestOpenAPIContract(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	server := httptest.NewServer(bootstrap.New(sandbox.NewService()))
	t.Cleanup(server.Close)

	var created createPaymentIntentEnvelope
	callContract(t, h, server, http.MethodPost, "/v1/payment_intents", map[string]any{
		"amount":   100,
		"currency": "usd",
	}, "create-contract", nil, &created)

	var confirmed confirmPaymentIntentEnvelope
	callContract(t, h, server, http.MethodPost, "/v1/payment_intents/"+created.PaymentIntent.ID+"/confirm", map[string]any{
		"payment_method_token": "pm_card_visa",
	}, "confirm-contract", map[string]string{"X-Sandbox-Scenario": "approved_immediate"}, &confirmed)

	var captured capturePaymentIntentEnvelope
	callContract(t, h, server, http.MethodPost, "/v1/payment_intents/"+created.PaymentIntent.ID+"/capture", map[string]any{}, "capture-contract", nil, &captured)

	var refunded refundEnvelope
	callContract(t, h, server, http.MethodPost, "/v1/refunds", map[string]any{
		"charge_id": captured.Charge.ID,
	}, "refund-contract", nil, &refunded)

	var gotIntent paymentIntentEnvelope
	callContract(t, h, server, http.MethodGet, "/v1/payment_intents/"+created.PaymentIntent.ID, nil, "", nil, &gotIntent)
}

type contractHarness struct {
	doc *openapi3.T
}

type createPaymentIntentEnvelope struct {
	PaymentIntent paymentIntentView `json:"payment_intent"`
}

type confirmPaymentIntentEnvelope struct {
	PaymentIntent  paymentIntentView  `json:"payment_intent"`
	PaymentAttempt paymentAttemptView `json:"payment_attempt"`
	Charge         *chargeView        `json:"charge"`
}

type capturePaymentIntentEnvelope struct {
	PaymentIntent paymentIntentView `json:"payment_intent"`
	Charge        chargeView        `json:"charge"`
}

type refundEnvelope struct {
	Refund refundView `json:"refund"`
	Charge chargeView `json:"charge"`
}

type paymentIntentEnvelope struct {
	PaymentIntent paymentIntentView `json:"payment_intent"`
}

type paymentIntentView struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	ChargeID        string `json:"charge_id"`
	LatestAttemptID string `json:"latest_attempt_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type paymentAttemptView struct {
	ID                 string `json:"id"`
	PaymentIntentID    string `json:"payment_intent_id"`
	PaymentMethodToken string `json:"payment_method_token"`
	Status             string `json:"status"`
	DeclineCode        string `json:"decline_code"`
	ProcessorReference string `json:"processor_reference"`
	RequestedAt        string `json:"requested_at"`
	RespondedAt        string `json:"responded_at"`
}

type chargeView struct {
	ID               string `json:"id"`
	PaymentIntentID  string `json:"payment_intent_id"`
	PaymentAttemptID string `json:"payment_attempt_id"`
	Amount           int64  `json:"amount"`
	CapturedAmount   int64  `json:"captured_amount"`
	RefundedAmount   int64  `json:"refunded_amount"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type refundView struct {
	ID              string `json:"id"`
	ChargeID        string `json:"charge_id"`
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          int64  `json:"amount"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func newHarness(t *testing.T) *contractHarness {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(openapiYAMLPath(t))
	if err != nil {
		t.Fatalf("load openapi yaml: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate openapi doc: %v", err)
	}
	return &contractHarness{doc: doc}
}

func callContract(t *testing.T, h *contractHarness, server *httptest.Server, method, path string, payload any, idem string, headers map[string]string, out any) {
	t.Helper()

	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal request payload: %v", err)
		}
	}
	requestPath := strings.TrimPrefix(path, "/v1")
	if requestPath == "" {
		requestPath = "/"
	}
	contractPath := resolveOpenAPIPath(h.doc, requestPath)

	validationReq, err := http.NewRequest(method, server.URL+requestPath, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build validation request: %v", err)
	}
	if payload != nil {
		validationReq.Header.Set("Content-Type", "application/json")
	}
	if idem != "" {
		validationReq.Header.Set("Idempotency-Key", idem)
	}
	for key, value := range headers {
		validationReq.Header.Set(key, value)
	}

	route, pathParams := routeForContract(t, h.doc, method, contractPath, requestPath)
	requestInput := &openapi3filter.RequestValidationInput{Request: validationReq, PathParams: pathParams, Route: route}
	if err := openapi3filter.ValidateRequest(context.Background(), requestInput); err != nil {
		t.Fatalf("validate request for %s %s: %v", method, path, err)
	}

	request, err := http.NewRequest(method, server.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idem != "" {
		request.Header.Set("Idempotency-Key", idem)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("call endpoint: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	responseInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestInput,
		Status:                 response.StatusCode,
		Header:                 response.Header,
		Body:                   io.NopCloser(bytes.NewReader(responseBody)),
	}
	if err := openapi3filter.ValidateResponse(context.Background(), responseInput); err != nil {
		t.Fatalf("validate response for %s %s: %v", method, path, err)
	}

	if out != nil {
		if err := json.Unmarshal(responseBody, out); err != nil {
			t.Fatalf("decode response for %s %s: %v", method, path, err)
		}
	}
}

func routeForContract(t *testing.T, doc *openapi3.T, method, contractPath, requestPath string) (*routers.Route, map[string]string) {
	t.Helper()

	pathItem := doc.Paths.Find(contractPath)
	if pathItem == nil {
		t.Fatalf("no openapi path item for %s", contractPath)
	}

	var op *openapi3.Operation
	switch method {
	case http.MethodGet:
		op = pathItem.Get
	case http.MethodPost:
		op = pathItem.Post
	case http.MethodPut:
		op = pathItem.Put
	case http.MethodPatch:
		op = pathItem.Patch
	case http.MethodDelete:
		op = pathItem.Delete
	default:
		t.Fatalf("unsupported method %s", method)
	}
	if op == nil {
		t.Fatalf("no openapi operation for %s %s", method, contractPath)
	}

	return &routers.Route{Spec: doc, Path: contractPath, PathItem: pathItem, Method: method, Operation: op}, extractPathParams(contractPath, requestPath)
}

func resolveOpenAPIPath(doc *openapi3.T, requestPath string) string {
	for _, path := range doc.Paths.InMatchingOrder() {
		if pathMatches(path, requestPath) {
			return path
		}
	}
	return requestPath
}

func pathMatches(templatePath, requestPath string) bool {
	templateParts := strings.Split(strings.Trim(templatePath, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(templateParts) != len(requestParts) {
		return false
	}
	for i := range templateParts {
		templatePart := templateParts[i]
		requestPart := requestParts[i]
		if strings.HasPrefix(templatePart, "{") && strings.HasSuffix(templatePart, "}") {
			continue
		}
		if templatePart != requestPart {
			return false
		}
	}
	return true
}

func extractPathParams(templatePath, requestPath string) map[string]string {
	parts := strings.Split(strings.Trim(templatePath, "/"), "/")
	params := make(map[string]string)
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")
	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") && i < len(requestParts) {
			name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			params[name] = requestParts[i]
		}
	}
	if len(params) == 0 {
		return nil
	}
	if _, ok := params["id"]; ok {
		params["id"] = "pi_contract_fixture"
	}
	return params
}

func openapiYAMLPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", "docs", "openapi", "payment-sandbox.v1.yaml"))
}
