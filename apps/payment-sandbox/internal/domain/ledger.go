package domain

import "time"

type LedgerEntry struct {
	ID               string    `json:"id"`
	EventName        string    `json:"event_name"`
	EntityType       string    `json:"entity_type"`
	EntityID         string    `json:"entity_id"`
	PaymentIntentID  string    `json:"payment_intent_id,omitempty"`
	PaymentAttemptID string    `json:"payment_attempt_id,omitempty"`
	ChargeID         string    `json:"charge_id,omitempty"`
	RefundID         string    `json:"refund_id,omitempty"`
	MerchantID       string    `json:"merchant_id,omitempty"`
	Currency         Currency  `json:"currency,omitempty"`
	BalanceBucket    string    `json:"balance_bucket,omitempty"`
	BalanceDelta     int64     `json:"balance_delta,omitempty"`
	Amount           Amount    `json:"amount"`
	CreatedAt        time.Time `json:"created_at"`
}
