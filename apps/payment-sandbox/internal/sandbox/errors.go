package sandbox

import "fmt"

type Error struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(statusCode int, code, message string) *Error {
	return &Error{StatusCode: statusCode, Code: code, Message: message}
}

func newError(statusCode int, code, message string) error {
	return NewError(statusCode, code, message)
}
