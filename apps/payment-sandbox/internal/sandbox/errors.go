package sandbox

import "payment-sandbox/internal/domain"

type Error = domain.Error

var NewError = domain.NewError

func newError(statusCode int, code, message string) error {
	return NewError(statusCode, code, message)
}
