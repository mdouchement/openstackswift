package weberror

import (
	"fmt"
	"net/http"
)

type (
	// HTTPCoder interface is implemented by application errors.
	HTTPCoder interface {
		// HTTPCode return the HTTP status code for the given error.
		HTTPCode() int
	}

	// Error is the payload rendered in case of error.
	Error struct {
		Code    int    `json:"-"`
		Message string `json:"message"`
	}
)

// StatusCode the know HHTP status for the given err. If unknown, it returns 500.
func StatusCode(err error) int {
	if hc, ok := err.(HTTPCoder); ok {
		return hc.HTTPCode()
	}
	return http.StatusInternalServerError
}

// New returns a new Error.
func New(code int, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Error stringifies the error.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// HTTPCode returns the HTTP status code.
func (e *Error) HTTPCode() int {
	return e.Code
}
