package core

import "time"

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success    bool        `json:"success"`              // Operation success status
	Data       interface{} `json:"data,omitempty"`       // Response payload
	Error      *APIError   `json:"error,omitempty"`      // Error details if any
	Meta       APIMeta     `json:"meta"`                 // Metadata about the response
	RequestID  string      `json:"request_id,omitempty"` // Request tracking ID
	ServerTime time.Time   `json:"server_time"`          // Server timestamp
}

// APIError represents standardized error information
type APIError struct {
	Code       string      `json:"code"`              // Error code
	Message    string      `json:"message"`           // User-friendly message
	Details    interface{} `json:"details,omitempty"` // Additional error context
	HTTPStatus int         `json:"-"`                 // HTTP status code
}

// APIMeta contains metadata about the response
type APIMeta struct {
	Page     int     `json:"page,omitempty"`     // Current page number
	PerPage  int     `json:"per_page,omitempty"` // Items per page
	Total    int     `json:"total,omitempty"`    // Total items
	Duration float64 `json:"duration_ms"`        // Request duration
	Version  string  `json:"version"`            // API version
}

// Standard error codes
const (
	ErrValidation   = "ERR_VALIDATION"
	ErrUnauthorized = "ERR_UNAUTHORIZED"
	ErrForbidden    = "ERR_FORBIDDEN"
	ErrNotFound     = "ERR_NOT_FOUND"
	ErrRateLimited  = "ERR_RATE_LIMITED"
	ErrInternal     = "ERR_INTERNAL"
	ErrBadRequest   = "ERR_BAD_REQUEST"
	ErrConflict     = "ERR_CONFLICT"
)

// ErrorToStatusCode maps error codes to HTTP status codes
var ErrorToStatusCode = map[string]int{
	ErrValidation:   400,
	ErrUnauthorized: 401,
	ErrForbidden:    403,
	ErrNotFound:     404,
	ErrRateLimited:  429,
	ErrInternal:     500,
	ErrBadRequest:   400,
	ErrConflict:     409,
}

// NewAPIResponse creates a new successful API response
func NewAPIResponse(data interface{}, meta APIMeta) *APIResponse {
	return &APIResponse{
		Success:    true,
		Data:       data,
		Meta:       meta,
		ServerTime: time.Now(),
	}
}

// NewErrorResponse creates a new error API response
func NewErrorResponse(err *APIError, meta APIMeta) *APIResponse {
	return &APIResponse{
		Success:    false,
		Error:      err,
		Meta:       meta,
		ServerTime: time.Now(),
	}
}

// NewAPIError creates a new API error with the given code and message
func NewAPIError(code string, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: ErrorToStatusCode[code],
	}
}
