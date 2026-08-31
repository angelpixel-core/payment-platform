package queries

import "time"

type PaymentIntentView struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	ChargeID        string    `json:"charge_id,omitempty"`
	LatestAttemptID string    `json:"latest_attempt_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PaymentAttemptView struct {
	ID                 string    `json:"id"`
	PaymentIntentID    string    `json:"payment_intent_id"`
	PaymentMethodToken string    `json:"payment_method_token,omitempty"`
	Status             string    `json:"status"`
	DeclineCode        string    `json:"decline_code,omitempty"`
	ProcessorReference string    `json:"processor_reference,omitempty"`
	RequestedAt        time.Time `json:"requested_at"`
	RespondedAt        time.Time `json:"responded_at"`
}

type ChargeView struct {
	ID               string    `json:"id"`
	PaymentIntentID  string    `json:"payment_intent_id"`
	PaymentAttemptID string    `json:"payment_attempt_id,omitempty"`
	Amount           int64     `json:"amount"`
	CapturedAmount   int64     `json:"captured_amount"`
	RefundedAmount   int64     `json:"refunded_amount"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RefundView struct {
	ID              string    `json:"id"`
	ChargeID        string    `json:"charge_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	Amount          int64     `json:"amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
