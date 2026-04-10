package common

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

const (
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeServiceUnavail = "SERVICE_UNAVAILABLE"
)

func NewNotFoundError(resource string) *DomainError {
	return &DomainError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

func NewBadRequestError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeBadRequest,
		Message: message,
	}
}

func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
	}
}
