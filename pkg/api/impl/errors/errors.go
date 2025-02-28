package errors

import (
	"fmt"

	types "github.com/magooney-loon/webserver/types/api"
)

// Common error codes
const (
	CodeInvalidRequest     = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeMethodNotAllowed   = 405
	CodeConflict           = 409
	CodeInternalError      = 500
	CodeServiceUnavailable = 503
)

// New creates a new API error
func New(code int, message string) *types.Error {
	return &types.Error{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to an error
func WithDetails(err *types.Error, details string) *types.Error {
	err.Details = details
	return err
}

// Wrap wraps an error with an API error
func Wrap(err error, code int, message string) *types.Error {
	return &types.Error{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}

// Common error constructors
func InvalidRequest(format string, args ...interface{}) *types.Error {
	return New(CodeInvalidRequest, fmt.Sprintf(format, args...))
}

func Unauthorized(format string, args ...interface{}) *types.Error {
	return New(CodeUnauthorized, fmt.Sprintf(format, args...))
}

func Forbidden(format string, args ...interface{}) *types.Error {
	return New(CodeForbidden, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...interface{}) *types.Error {
	return New(CodeNotFound, fmt.Sprintf(format, args...))
}

func MethodNotAllowed(format string, args ...interface{}) *types.Error {
	return New(CodeMethodNotAllowed, fmt.Sprintf(format, args...))
}

func Conflict(format string, args ...interface{}) *types.Error {
	return New(CodeConflict, fmt.Sprintf(format, args...))
}

func InternalError(format string, args ...interface{}) *types.Error {
	return New(CodeInternalError, fmt.Sprintf(format, args...))
}

func ServiceUnavailable(format string, args ...interface{}) *types.Error {
	return New(CodeServiceUnavailable, fmt.Sprintf(format, args...))
}
