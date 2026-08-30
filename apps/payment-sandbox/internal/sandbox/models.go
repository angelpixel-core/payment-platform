package sandbox

import "payment-sandbox/internal/domain"

type PaymentIntentStatus = domain.PaymentIntentStatus
type PaymentAttemptStatus = domain.PaymentAttemptStatus
type ChargeStatus = domain.ChargeStatus
type RefundStatus = domain.RefundStatus

const (
	PaymentIntentRequiresPaymentMethod PaymentIntentStatus = domain.PaymentIntentRequiresPaymentMethod
	PaymentIntentRequiresConfirmation  PaymentIntentStatus = domain.PaymentIntentRequiresConfirmation
	PaymentIntentRequiresAction        PaymentIntentStatus = domain.PaymentIntentRequiresAction
	PaymentIntentProcessing            PaymentIntentStatus = domain.PaymentIntentProcessing
	PaymentIntentRequiresCapture       PaymentIntentStatus = domain.PaymentIntentRequiresCapture
	PaymentIntentSucceeded             PaymentIntentStatus = domain.PaymentIntentSucceeded
	PaymentIntentCancelled             PaymentIntentStatus = domain.PaymentIntentCancelled
	PaymentIntentFailed                PaymentIntentStatus = domain.PaymentIntentFailed

	PaymentAttemptCreated        PaymentAttemptStatus = domain.PaymentAttemptCreated
	PaymentAttemptSubmitted      PaymentAttemptStatus = domain.PaymentAttemptSubmitted
	PaymentAttemptRequiresAction PaymentAttemptStatus = domain.PaymentAttemptRequiresAction
	PaymentAttemptAuthorized     PaymentAttemptStatus = domain.PaymentAttemptAuthorized
	PaymentAttemptDeclined       PaymentAttemptStatus = domain.PaymentAttemptDeclined
	PaymentAttemptTimedOut       PaymentAttemptStatus = domain.PaymentAttemptTimedOut
	PaymentAttemptErrored        PaymentAttemptStatus = domain.PaymentAttemptErrored

	ChargeAuthorized        ChargeStatus = domain.ChargeAuthorized
	ChargeCaptured          ChargeStatus = domain.ChargeCaptured
	ChargePartiallyRefunded ChargeStatus = domain.ChargePartiallyRefunded
	ChargeRefunded          ChargeStatus = domain.ChargeRefunded
	ChargeReversed          ChargeStatus = domain.ChargeReversed
	ChargeDisputed          ChargeStatus = domain.ChargeDisputed

	RefundSucceeded RefundStatus = domain.RefundSucceeded
	RefundFailed    RefundStatus = domain.RefundFailed
)

type PaymentIntent = domain.PaymentIntent
type PaymentAttempt = domain.PaymentAttempt
type Charge = domain.Charge
type Refund = domain.Refund

type CreatePaymentIntentRequest = domain.CreatePaymentIntentRequest
type ConfirmPaymentIntentRequest = domain.ConfirmPaymentIntentRequest
type CapturePaymentIntentRequest = domain.CapturePaymentIntentRequest
type RefundRequest = domain.RefundRequest

type CreatePaymentIntentResponse = domain.CreatePaymentIntentResponse
type ConfirmPaymentIntentResponse = domain.ConfirmPaymentIntentResponse
type CapturePaymentIntentResponse = domain.CapturePaymentIntentResponse
type RefundResponse = domain.RefundResponse
