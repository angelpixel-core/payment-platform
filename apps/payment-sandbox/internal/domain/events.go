package domain

type Event interface {
	EventName() string
}

type PaymentIntentCreatedEvent struct {
	PaymentIntent PaymentIntent
}

func (PaymentIntentCreatedEvent) EventName() string { return "payment_intent.created" }

type PaymentIntentConfirmedEvent struct {
	PaymentIntent  PaymentIntent
	PaymentAttempt PaymentAttempt
	Charge         *Charge
}

func (PaymentIntentConfirmedEvent) EventName() string { return "payment_intent.confirmed" }

type PaymentIntentFinalizedEvent struct {
	PaymentIntent PaymentIntent
}

func (PaymentIntentFinalizedEvent) EventName() string { return "payment_intent.finalized" }

type PaymentIntentCapturedEvent struct {
	PaymentIntent PaymentIntent
	Charge        Charge
}

func (PaymentIntentCapturedEvent) EventName() string { return "payment_intent.captured" }

type RefundCreatedEvent struct {
	Refund Refund
	Charge Charge
}

func (RefundCreatedEvent) EventName() string { return "refund.created" }
