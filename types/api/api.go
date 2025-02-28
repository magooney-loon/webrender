package types

import (
	"net/http"
)

// Response represents a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Handler defines the interface for API handlers
type Handler interface {
	// ServeHTTP handles the HTTP request
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// Validate validates the request
	Validate(r *http.Request) error
}

// Responder defines methods for sending API responses
type Responder interface {
	// JSON sends a JSON response
	JSON(w http.ResponseWriter, status int, data interface{})
	// Error sends an error response
	Error(w http.ResponseWriter, err *Error)
	// Stream sends a streaming response
	Stream(w http.ResponseWriter, stream chan interface{})
}

// Validator defines methods for request validation
type Validator interface {
	// Validate validates the request data
	Validate() error
}

// RequestDecoder defines methods for decoding requests
type RequestDecoder interface {
	// DecodeJSON decodes JSON request body
	DecodeJSON(r *http.Request, v interface{}) error
	// DecodeQuery decodes query parameters
	DecodeQuery(r *http.Request, v interface{}) error
}

// ResponseEncoder defines methods for encoding responses
type ResponseEncoder interface {
	// EncodeJSON encodes response as JSON
	EncodeJSON(w http.ResponseWriter, v interface{}) error
	// EncodeError encodes an error response
	EncodeError(w http.ResponseWriter, err *Error) error
}
