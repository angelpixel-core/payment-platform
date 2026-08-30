package domain

import "time"

type PaymentIntentStatus string

const (
	PaymentIntentRequiresPaymentMethod PaymentIntentStatus = "requires_payment_method"
	PaymentIntentRequiresConfirmation  PaymentIntentStatus = "requires_confirmation"
	PaymentIntentRequiresAction        PaymentIntentStatus = "requires_action"
	PaymentIntentProcessing            PaymentIntentStatus = "processing"
	PaymentIntentRequiresCapture       PaymentIntentStatus = "requires_capture"
	PaymentIntentSucceeded             PaymentIntentStatus = "succeeded"
	PaymentIntentCancelled             PaymentIntentStatus = "cancelled"
	PaymentIntentFailed                PaymentIntentStatus = "failed"
)

type PaymentAttemptStatus string

const (
	PaymentAttemptCreated        PaymentAttemptStatus = "created"
	PaymentAttemptSubmitted      PaymentAttemptStatus = "submitted"
	PaymentAttemptRequiresAction PaymentAttemptStatus = "requires_action"
	PaymentAttemptAuthorized     PaymentAttemptStatus = "authorized"
	PaymentAttemptDeclined       PaymentAttemptStatus = "declined"
	PaymentAttemptTimedOut       PaymentAttemptStatus = "timed_out"
	PaymentAttemptErrored        PaymentAttemptStatus = "errored"
)

type ChargeStatus string

const (
	ChargeAuthorized        ChargeStatus = "authorized"
	ChargeCaptured          ChargeStatus = "captured"
	ChargePartiallyRefunded ChargeStatus = "partially_refunded"
	ChargeRefunded          ChargeStatus = "refunded"
	ChargeReversed          ChargeStatus = "reversed"
	ChargeDisputed          ChargeStatus = "disputed"
)

type RefundStatus string

const (
	RefundSucceeded RefundStatus = "succeeded"
	RefundFailed    RefundStatus = "failed"
)

type PaymentIntent struct {
	ID              string              `json:"id"`
	MerchantID      string              `json:"merchant_id,omitempty"`
	CustomerID      string              `json:"customer_id,omitempty"`
	Amount          int64               `json:"amount"`
	Currency        string              `json:"currency"`
	CaptureMethod   string              `json:"capture_method"`
	Status          PaymentIntentStatus `json:"status"`
	Scenario        string              `json:"scenario,omitempty"`
	IdempotencyKey  string              `json:"idempotency_key,omitempty"`
	LatestAttemptID string              `json:"latest_attempt_id,omitempty"`
	ChargeID        string              `json:"charge_id,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type PaymentAttempt struct {
	ID                 string               `json:"id"`
	PaymentIntentID    string               `json:"payment_intent_id"`
	PaymentMethodToken string               `json:"payment_method_token,omitempty"`
	Status             PaymentAttemptStatus `json:"status"`
	DeclineCode        string               `json:"decline_code,omitempty"`
	ProcessorReference string               `json:"processor_reference,omitempty"`
	RequestedAt        time.Time            `json:"requested_at"`
	RespondedAt        time.Time            `json:"responded_at"`
}

type Charge struct {
	ID               string       `json:"id"`
	PaymentIntentID  string       `json:"payment_intent_id"`
	PaymentAttemptID string       `json:"payment_attempt_id,omitempty"`
	Amount           int64        `json:"amount"`
	CapturedAmount   int64        `json:"captured_amount"`
	RefundedAmount   int64        `json:"refunded_amount"`
	Status           ChargeStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type Refund struct {
	ID              string       `json:"id"`
	ChargeID        string       `json:"charge_id"`
	PaymentIntentID string       `json:"payment_intent_id"`
	Amount          int64        `json:"amount"`
	Status          RefundStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type CreatePaymentIntentRequest struct {
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	MerchantID     string `json:"merchant_id,omitempty"`
	CustomerID     string `json:"customer_id,omitempty"`
	CaptureMethod  string `json:"capture_method,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type ConfirmPaymentIntentRequest struct {
	PaymentMethodToken string `json:"payment_method_token,omitempty"`
	IdempotencyKey     string `json:"idempotency_key,omitempty"`
}

type CapturePaymentIntentRequest struct {
	Amount int64 `json:"amount,omitempty"`
}

type RefundRequest struct {
	ChargeID       string `json:"charge_id"`
	Amount         int64  `json:"amount,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type CreatePaymentIntentResponse struct {
	PaymentIntent PaymentIntent `json:"payment_intent"`
}

type ConfirmPaymentIntentResponse struct {
	PaymentIntent  PaymentIntent  `json:"payment_intent"`
	PaymentAttempt PaymentAttempt `json:"payment_attempt"`
	Charge         *Charge        `json:"charge,omitempty"`
}

type CapturePaymentIntentResponse struct {
	PaymentIntent PaymentIntent `json:"payment_intent"`
	Charge        Charge        `json:"charge"`
}

type RefundResponse struct {
	Refund Refund `json:"refund"`
	Charge Charge `json:"charge"`
}
