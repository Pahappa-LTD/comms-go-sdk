// Package exceptions holds the typed errors returned by the SDK.
package exceptions

// CommsError is implemented by every error the SDK returns, so callers can match any of them with errors.As.
type CommsError interface {
	error
	commsError()
}

// CommsAuthenticationError means the account credentials were rejected by the server, or could not be verified.
type CommsAuthenticationError struct {
	Message string
	Cause   error
}

func (e *CommsAuthenticationError) Error() string { return e.Message }
func (e *CommsAuthenticationError) Unwrap() error { return e.Cause }
func (e *CommsAuthenticationError) commsError()   {}

// CommsValidationError means the SDK rejected the input before any request was sent.
type CommsValidationError struct {
	Message string
}

func (e *CommsValidationError) Error() string { return e.Message }
func (e *CommsValidationError) commsError()   {}

// CommsApiError means a request could not be completed, or the server's response was unusable.
type CommsApiError struct {
	Message string
	Cause   error
}

func (e *CommsApiError) Error() string { return e.Message }
func (e *CommsApiError) Unwrap() error { return e.Cause }
func (e *CommsApiError) commsError()   {}
