package domain

import "time"

type ConfirmPaymentIntentCommand struct {
	Outcome ScenarioOutcome
	Attempt *PaymentAttempt
	Charge  *Charge
	Now     time.Time
}

type FinalizeProcessingCommand struct {
	Attempt *PaymentAttempt
	Charge  *Charge
	Now     time.Time
}

type CapturePaymentIntentCommand struct {
	Charge *Charge
	Amount int64
	Now    time.Time
}

type ConfirmPaymentIntentResult struct {
	PaymentIntent  PaymentIntent
	PaymentAttempt PaymentAttempt
	Charge         *Charge
}

type FinalizeProcessingResult struct {
	PaymentIntent  PaymentIntent
	PaymentAttempt PaymentAttempt
	Charge         Charge
}

type CapturePaymentIntentResult struct {
	PaymentIntent PaymentIntent
	Charge        Charge
}
