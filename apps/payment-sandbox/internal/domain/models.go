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
	Amount          Amount              `json:"amount"`
	Currency        Currency            `json:"currency"`
	CaptureMethod   string              `json:"capture_method"`
	Status          PaymentIntentStatus `json:"status"`
	Scenario        string              `json:"scenario,omitempty"`
	IdempotencyKey  string              `json:"idempotency_key,omitempty"`
	LatestAttemptID string              `json:"latest_attempt_id,omitempty"`
	ChargeID        string              `json:"charge_id,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

func (p PaymentIntent) CanConfirm() error {
	if p.Status != PaymentIntentRequiresPaymentMethod && p.Status != PaymentIntentRequiresConfirmation {
		return NewError(409, "invalid_intent_state", "payment intent cannot be confirmed in its current state")
	}
	return nil
}

func (p PaymentIntent) CanCapture() error {
	if p.Status != PaymentIntentRequiresCapture {
		return NewError(409, "invalid_intent_state", "payment intent cannot be captured in its current state")
	}
	return nil
}

func (p PaymentIntent) CanFinalizeProcessing() error {
	if p.Status != PaymentIntentProcessing {
		return NewError(409, "invalid_intent_state", "payment intent cannot be finalized in its current state")
	}
	if NormalizeScenarioName(p.Scenario) != ScenarioProcessingThenSucceeded {
		return NewError(409, "invalid_scenario", "payment intent is not in the processing scenario")
	}
	return nil
}

func (p *PaymentIntent) Confirm(cmd ConfirmPaymentIntentCommand) (ConfirmPaymentIntentResult, error) {
	if err := p.CanConfirm(); err != nil {
		return ConfirmPaymentIntentResult{}, err
	}
	attempt := cmd.Attempt
	if attempt == nil {
		return ConfirmPaymentIntentResult{}, NewError(409, "invalid_attempt_state", "payment attempt is required")
	}
	charge := cmd.Charge
	now := cmd.Now
	outcome := cmd.Outcome

	p.Scenario = string(outcome.Scenario)
	p.LatestAttemptID = attempt.ID
	p.UpdatedAt = now

	if outcome.FinalizesLater {
		p.Status = outcome.IntentStatus
		return ConfirmPaymentIntentResult{PaymentIntent: *p, PaymentAttempt: *attempt}, nil
	}

	switch outcome.Scenario {
	case ScenarioApprovedImmediate:
		if charge == nil {
			return ConfirmPaymentIntentResult{}, NewError(409, "invalid_charge_state", "charge is required")
		}
		charge.PaymentIntentID = p.ID
		charge.PaymentAttemptID = attempt.ID
		charge.Amount = p.Amount
		charge.CreatedAt = now
		charge.UpdatedAt = now
		charge.Status = outcome.ChargeStatus
		if p.CaptureMethod == "automatic" {
			charge.CapturedAmount = p.Amount
			charge.Status = ChargeCaptured
			p.Status = PaymentIntentSucceeded
		} else {
			p.Status = PaymentIntentRequiresCapture
		}
		p.ChargeID = charge.ID
	case ScenarioDeclinedInsufficientFunds:
		p.Status = outcome.IntentStatus
		attempt.DeclineCode = outcome.DeclineCode
	case ScenarioRequiresAction3DS:
		p.Status = outcome.IntentStatus
	default:
		return ConfirmPaymentIntentResult{}, NewError(422, "invalid_scenario", "scenario not supported")
	}

	var resultCharge *Charge
	if charge != nil {
		resultCharge = charge
	}
	return ConfirmPaymentIntentResult{PaymentIntent: *p, PaymentAttempt: *attempt, Charge: resultCharge}, nil
}

func (p *PaymentIntent) FinalizeProcessing(cmd FinalizeProcessingCommand) (FinalizeProcessingResult, error) {
	if err := p.CanFinalizeProcessing(); err != nil {
		return FinalizeProcessingResult{}, err
	}
	attempt := cmd.Attempt
	if attempt == nil {
		return FinalizeProcessingResult{}, NewError(409, "invalid_attempt_state", "processing payment intent cannot be finalized in its current attempt state")
	}
	if attempt.Status != PaymentAttemptSubmitted {
		return FinalizeProcessingResult{}, NewError(409, "invalid_attempt_state", "processing payment intent cannot be finalized in its current attempt state")
	}
	charge := cmd.Charge
	if charge == nil {
		return FinalizeProcessingResult{}, NewError(409, "invalid_charge_state", "charge is required")
	}
	now := cmd.Now

	charge.PaymentIntentID = p.ID
	charge.PaymentAttemptID = attempt.ID
	charge.Amount = p.Amount
	charge.CreatedAt = now
	charge.UpdatedAt = now
	if p.CaptureMethod == "automatic" {
		charge.CapturedAmount = p.Amount
		charge.Status = ChargeCaptured
		p.Status = PaymentIntentSucceeded
	} else {
		charge.Status = ChargeAuthorized
		p.Status = PaymentIntentRequiresCapture
	}
	p.ChargeID = charge.ID
	p.UpdatedAt = now
	attempt.Status = PaymentAttemptAuthorized
	attempt.RespondedAt = now

	return FinalizeProcessingResult{PaymentIntent: *p, PaymentAttempt: *attempt, Charge: *charge}, nil
}

func (p *PaymentIntent) Capture(cmd CapturePaymentIntentCommand) (CapturePaymentIntentResult, error) {
	if err := p.CanCapture(); err != nil {
		return CapturePaymentIntentResult{}, err
	}
	charge := cmd.Charge
	if charge == nil {
		return CapturePaymentIntentResult{}, NewError(409, "invalid_charge_state", "charge cannot be captured in its current state")
	}
	if charge.Status != ChargeAuthorized {
		return CapturePaymentIntentResult{}, NewError(409, "invalid_charge_state", "charge cannot be captured in its current state")
	}

	amount := cmd.Amount
	now := cmd.Now
	if amount == 0 {
		amount = charge.Amount
	}
	if amount <= 0 {
		return CapturePaymentIntentResult{}, NewError(400, "invalid_amount", "capture amount must be greater than zero")
	}
	if amount > charge.Amount {
		return CapturePaymentIntentResult{}, NewError(400, "invalid_amount", "capture amount cannot exceed charge amount")
	}

	charge.CapturedAmount = amount
	charge.Status = ChargeCaptured
	charge.UpdatedAt = now
	p.Status = PaymentIntentSucceeded
	p.UpdatedAt = now

	return CapturePaymentIntentResult{PaymentIntent: *p, Charge: *charge}, nil
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
	Amount           Amount       `json:"amount"`
	CapturedAmount   Amount       `json:"captured_amount"`
	RefundedAmount   Amount       `json:"refunded_amount"`
	Status           ChargeStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type Refund struct {
	ID              string       `json:"id"`
	ChargeID        string       `json:"charge_id"`
	PaymentIntentID string       `json:"payment_intent_id"`
	Amount          Amount       `json:"amount"`
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
