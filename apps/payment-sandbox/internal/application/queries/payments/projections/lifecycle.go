package projections

import (
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

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

type PaymentLifecycleView struct {
	PaymentIntent    PaymentIntentView   `json:"payment_intent"`
	LatestAttempt    *PaymentAttemptView  `json:"latest_attempt,omitempty"`
	Charge           *ChargeView         `json:"charge,omitempty"`
	Status           string              `json:"status"`
	RefundableAmount int64               `json:"refundable_amount"`
	IsRefundable     bool                `json:"is_refundable"`
}

func BuildPaymentLifecycle(store ports.Store, id string) (PaymentLifecycleView, error) {
	intent, err := store.GetPaymentIntent(id)
	if err != nil {
		return PaymentLifecycleView{}, err
	}

	view := PaymentLifecycleView{
		PaymentIntent: PaymentIntentView{
			ID:              intent.ID,
			Status:          string(intent.Status),
			Amount:          intent.Amount.Int64(),
			Currency:        intent.Currency.String(),
			ChargeID:        intent.ChargeID,
			LatestAttemptID: intent.LatestAttemptID,
			CreatedAt:       intent.CreatedAt,
			UpdatedAt:       intent.UpdatedAt,
		},
	}

	if intent.LatestAttemptID != "" {
		attempt, err := store.GetPaymentAttempt(intent.LatestAttemptID)
		if err != nil {
			return PaymentLifecycleView{}, err
		}
		attemptView := PaymentAttemptView{
			ID:                 attempt.ID,
			PaymentIntentID:    attempt.PaymentIntentID,
			PaymentMethodToken: attempt.PaymentMethodToken,
			Status:             string(attempt.Status),
			DeclineCode:        attempt.DeclineCode,
			ProcessorReference: attempt.ProcessorReference,
			RequestedAt:        attempt.RequestedAt,
			RespondedAt:        attempt.RespondedAt,
		}
		view.LatestAttempt = &attemptView
	}

	if intent.ChargeID != "" {
		charge, err := store.GetCharge(intent.ChargeID)
		if err != nil {
			return PaymentLifecycleView{}, err
		}
		chargeView := ChargeView{
			ID:               charge.ID,
			PaymentIntentID:  charge.PaymentIntentID,
			PaymentAttemptID: charge.PaymentAttemptID,
			Amount:           charge.Amount.Int64(),
			CapturedAmount:   charge.CapturedAmount.Int64(),
			RefundedAmount:   charge.RefundedAmount.Int64(),
			Status:           string(charge.Status),
			CreatedAt:        charge.CreatedAt,
			UpdatedAt:        charge.UpdatedAt,
		}
		view.Charge = &chargeView
		remaining := charge.CapturedAmount - charge.RefundedAmount
		if remaining > 0 {
			view.IsRefundable = true
			view.RefundableAmount = remaining.Int64()
		}
	}

	view.Status = derivePaymentLifecycleStatus(intent.Status, view.Charge)
	return view, nil
}

func derivePaymentLifecycleStatus(intentStatus domain.PaymentIntentStatus, charge *ChargeView) string {
	switch intentStatus {
	case domain.PaymentIntentRequiresPaymentMethod, domain.PaymentIntentRequiresConfirmation, domain.PaymentIntentRequiresAction, domain.PaymentIntentProcessing, domain.PaymentIntentRequiresCapture, domain.PaymentIntentFailed:
		return string(intentStatus)
	case domain.PaymentIntentSucceeded:
		if charge == nil {
			return string(intentStatus)
		}
		switch charge.Status {
		case string(domain.ChargeRefunded):
			return "refunded"
		case string(domain.ChargePartiallyRefunded):
			return "partially_refunded"
		default:
			return string(intentStatus)
		}
	default:
		return string(intentStatus)
	}
}
