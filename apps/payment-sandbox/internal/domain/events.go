package domain

type Event interface {
	EventName() string
	EventVersion() int
}

const EventVersionV1 = 1

type PaymentIntentCreatedEvent struct {
	PaymentIntent PaymentIntent
}

func (PaymentIntentCreatedEvent) EventName() string { return "payment_intent.created" }
func (PaymentIntentCreatedEvent) EventVersion() int { return EventVersionV1 }

type PaymentIntentConfirmedEvent struct {
	PaymentIntent  PaymentIntent
	PaymentAttempt PaymentAttempt
	Charge         *Charge
}

func (PaymentIntentConfirmedEvent) EventName() string { return "payment_intent.confirmed" }
func (PaymentIntentConfirmedEvent) EventVersion() int { return EventVersionV1 }

type PaymentIntentFinalizedEvent struct {
	PaymentIntent PaymentIntent
}

func (PaymentIntentFinalizedEvent) EventName() string { return "payment_intent.finalized" }
func (PaymentIntentFinalizedEvent) EventVersion() int { return EventVersionV1 }

type PaymentIntentCapturedEvent struct {
	PaymentIntent PaymentIntent
	Charge        Charge
}

func (PaymentIntentCapturedEvent) EventName() string { return "payment_intent.captured" }
func (PaymentIntentCapturedEvent) EventVersion() int { return EventVersionV1 }

type RefundCreatedEvent struct {
	Refund Refund
	Charge Charge
}

func (RefundCreatedEvent) EventName() string { return "refund.created" }
func (RefundCreatedEvent) EventVersion() int { return EventVersionV1 }
